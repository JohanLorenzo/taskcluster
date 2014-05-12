var express     = require('express');
var debug       = require('debug')('base:api');
var Promise     = require('promise');
var uuid        = require('uuid');
var hawk        = require('hawk');
var aws         = require('aws-sdk-promise');
var assert      = require('assert');
var _           = require('lodash');
var taskcluster = require('taskcluster-client');

/**
 * Declare {input, output} schemas as options to validate
 *
 * options:
 * {
 *   input:    'http://schemas...input-schema.json',   // optional, null if no input
 *   output:   'http://schemas...output-schema.json',  // optional, null if no output
 *   skipInputValidation:    true,                     // defaults to false
 *   skipOutputValidation:   true,                     // defaults to false
 *   validator:              new base.validator()      // JSON schema validator
 * }
 *
 * This validates body against the schema given in `options.input` and returns
 * and a 400 error messages to request if there is a schema mismatch.
 * Handlers below this should output the reply JSON structure with `req.reply`.
 * this will validate it against `options.output` if provided.
 * Handlers may output errors using `req.json`, as `req.reply` will validate
 * against schema and always returns a 200 OK reply.
 */
var schema = function(options) {
  return function(req, res, next) {
    // If input schema is defined we need to validate the input
    if (options.input !== undefined) {
      var errors = options.validator.check(req.body, options.input);
      if (errors) {
        debug("Request payload for %s didn't follow schema %s",
              req.url, options.input);
        res.json(400, {
          'message':  "Request payload must follow the schema: " + options.input,
          'error':    errors
        });
        return;
      }
    }
    // Add a reply method sending JSON replies, this will always reply with HTTP
    // code 200... errors should be sent with res.json(code, json)
    res.reply = function(json) {
      // If we're supposed to validate outgoing messages and output schema is
      // defined, then we have to validate against it...
      if(options.output !== undefined) {
        var errors = options.validator.check(json, options.output);
        if (errors) {
          res.json(500, {
            'message':  "Internal Server Error",
          });
          debug("Reply for %s didn't match schema: %s got errors:\n%s",
                req.url, options.output, JSON.stringify(errors, null, 4));
          return;
        }
      }
      // If JSON was valid or validation was skipped then reply with 200 OK
      res.json(200, json);
    };

    // Call next piece of middleware, typically the handler...
    next();
  };
};

/** Abstraction of a client with some helper methods */
var Client = function(options) {
  assert(options.clientId,                "ClientId is required");
  assert(options.accessToken,             "AccessToken is required");
  assert(options.scopes instanceof Array, "Scopes must be an array");
  if (typeof(options.expires) == 'string') {
    this.expires = new Date(options.expires);
  } else {
    this.expires = options.expires;
  }
  assert(this.expires instanceof Date,    "Expires must be a date");
  assert(isFinite(this.expires),          "Expires must be a valid date");
  this.clientId     = options.clientId;
  this.accessToken  = options.accessToken;
  this.scopes       = options.scopes;
};

/** Check if the client satisfies any of the given scopes */
Client.satisfies = function(scopes) {
  if (typeof(scopes) == 'string') {
    scopes = [scopes];
  }
  assert(scopes instanceof Array, "Scopes must be a string or an array");
  return this.scopes.some(function(pattern) {
    var match = /^(.*)\*$/.exec(pattern);
    return scopes.some(function(scope) {
      if (scope === pattern) {
        return true;
      }
      if (match) {
        return scope.indexOf(match[1]) == 0;
      }
      return false;
    });
  });
};

/** Check if client credentials are expired */
Client.isExpired = function() {
  return this.expires < (new Date());
};

/**
 * Create function to find a client from clientId
 *
 * options:
 * {
 *   clientId:          '...',  // TaskCluster clientId
 *   accessToken:       '...'   // Access token for clientId
 * }
 *
 * The client identified by `clientId` must have the scope 'auth:credentials'.
 *
 * Return a function that gives promise for an instance  of `Client`, use the
 * return value with `API.router(...)` or `API.initialize(...)`.
 */
var clientLoader = function(options) {
  assert(options.clientId,    "ClientId is required");
  assert(options.accessToken, "AccessToken is required");
  var auth = new taskcluster.auth(options);
  return function(clientId) {
    return  auth.getCredentials(clientId).then(function(client) {
      return new Client(client);
    });
  };
};

/** Local nonce cache, using an over-approximation */
var nonceManager = function() {
  var nextnonce = 0;
  var N = 500;
  var noncedb = new Array(500);
  for(var i = 0; i < 500; i++) {
    noncedb[i] = {nonce: null, ts: null};
  }
  return function(nonce, ts, cb) {
    for(var i = 0; i < 500; i++) {
      if (noncedb[i].nonce === nonce && noncedb[i].ts === ts) {
        debug("CRITICAL: Replay attack detected!");
        return cb(new Error("Signature already used"));
      }
    }
    noncedb[nextnonce++].nonce  = nonce;
    noncedb[nextnonce++].ts     = ts;
    cb();
  };
};

/**
 * Authenticate client and validate that he has one of the scopes required.
 * Skips validation if `options.scopes` is `undefined`.
 *
 * options:
 * {
 *   scopes:  ['admin', 'superuser', 'service:method:action']
 * }
 *
 * Checks that the authenticated client has a scope pattern that matches one
 * of the declared `scopes`. If the client has pattern "service:*" this will
 * match any scope that starts with "service:" as is the case in the example
 * above.
 *
 * Note: Both API end-points and clients has a list of scopes. To authorize
 * a request we just need an intersection between these two lists.
 *
 * Reports 401 if authentication fails.
 */
var authenticate = function(nonceManager, clientLoader, options) {
  var getCredentials = function(clientId, cb) {
    clientLoader(clientId).then(function(client) {
      cb(null, {
        clientToken:    client.clientId,
        key:            client.accessToken,
        algorithm:      'sha256',
        client:         client
      });
    }).catch(function(err) {
      cb(err);
    });
  }
  return function(req, res, next) {
    // Restore originalUrl as needed by hawk for authentication
    req.url = req.originalUrl;
    if (options.scopes == undefined) {
      next();
    } else {
      hawk.server.authenticate(req, getCredentials, {
        // Not sure if JSON stringify is not deterministic by specification.
        // I suspect not, so we'll postpone this till we're sure we want to do
        // payload validation and how we want to do it.
        //payload:      JSON.stringify(req.body),
        nonceFunc:    nonceManager
      }, function(err, credentials, artifacts) {
        // Handle error fetching credentials or validating signature
        if (err) {
          var incidentId = uuid.v4();
          var message = "Ask administrator to lookup incidentId in log-file";
          if (err.output && err.output.payload && err.output.payload.error) {
            message = err.output.payload.error;
          }
          debug(
            "Error occurred authenticating, err: %s, %j, incidentId: %s",
            err, err, incidentId, err.stack
          );
          return res.json(401, {
            message:        "Internal Server Error",
            error: {
              info:         message,
              incidentId:   incidentId
            }
          });
        }
        // Check that the request is authorized
        if (!credentials.client.isExpired() &&
            credentials.client.satisfies(options.scopes)) {
          // Set client on request and call next
          req.client = credentials.client;
          next();
        } else {
          res.json(401, {message: "Authenticated by not authorized!"});
        }
      });
    }
  };
};

// Export authentication utilities
authenticate.Client       = Client;
authenticate.clientLoader = clientLoader;
authenticate.nonceManager = nonceManager;

/**
 * Handle API end-point request
 *
 * This invokes the handler with `context` as `this` and then catches
 * exceptions and failures of returned promises handler.
 */
var handle = function(handler, context) {
  assert(handler, "No handler is provided");
  return function(req, res) {
    Promise.from().then(function() {
      return handler.call(context, req, res);
    }).catch(function(err) {
      var incidentId = uuid.v4();
      debug(
        "Error occurred handling: %s, err: %s, as JSON: %j, incidentId: %s",
        req.url, err, err, incidentId, err.stack
      );
      res.json(500, {
        message:        "Internal Server Error",
        error: {
          info:         "Ask administrator to lookup incidentId in log-file",
          incidentId:   incidentId
        }
      });
    });
  };
};

/**
 * Create an API builder
 *
 * options:
 * {
 *   title: "API Title",
 *   desc:  "API description in markdown"
 * }
 *
 * The API object will only modified by declarations, when `mount` or `publish`
 * is called it'll use the currently defined entries to mount or publish the
 * API.
 */
var API = function(options) {
  ['title', 'desc'].forEach(function(key) {
    assert(options[key], "Option '" + key + "' must be provided");
  });
  this._options = options;
  this._entries = [];
};

/**
 * Declare an API end-point entry, where options is on the following form:
 *
 * {
 *   method:   'post|head|put|get|delete',
 *   route:    '/object/:id/action/:parameter',        // Only on illustrated form
 *   name:     'identifierForLibraries',               // identifier used by client libraries
 *   scopes:   ['admin', 'superuser'],                 // Scopes of which user must have one
 *   input:    'http://schemas...input-schema.json',   // optional, null if no input
 *   output:   'http://schemas...output-schema.json',  // optional, null if no output
 *   skipInputValidation:    true,                     // defaults to false
 *   skipOutputValidation:   true,                     // defaults to false
 *   title:    "My API Method",
 *   desc:     "Description of method in markdown, enjoy"
 * }
 *
 * The handler parameter is a normal connect/express request handler, it should
 * return JSON replies with `request.reply(json)` and errors with
 * `request.json(code, json)`, as `request.reply` may be validated against the
 * declared output schema.
 *
 * **Note** the handler may return a promise, if this promise fails we will
 * log the error and return an error message. If the promise is successful,
 * nothing happens.
 */
API.prototype.declare = function(options, handler) {
  ['method', 'route', 'title', 'desc'].forEach(function(key) {
    assert(options[key], "Option '" + key + "' must be provided");
  });
  options.handler = handler;
  this._entries.push(options);
};

/**
 * Construct a router that can be mounted on an express application
 *
 * options:
 * {
 *   inputLimit:          '10mb'  // Max input JSON size
 *   allowedCORSOrigin:   '*'     // Allowed CORS origin, null to disable CORS
 *   context:             {}      // Object to be provided as `this` in handlers
 *   validator:           new base.validator()      // JSON schema validator
 *   nonceManager:        function(nonce, ts, cb) { // Check for replay attack
 *   clientLoader:        function(clientId) {      // Return promise for client
 *   credentials: {
 *     clientId:          '...',  // TaskCluster clientId
 *     accessToken:       '...'   // Access token for clientId
 *     // Client must have the 'auth:credentials' scope.
 *   }
 * }
 *
 * The option `validator` must provided, and either `credentials` or
 * `clientLoader` must be provided.
 *
 * Return an `express.Router` instance.
 */
API.prototype.router = function(options) {
  assert(options.validator, "API.router() needs a validator");

  // Provide default options
  _.defaults(options, {
    inputLimit:           '10mb',
    allowedCORSOrigin:    '*',
    context:              {},
    nonceManager:         nonceManager()
  });

  // Create clientLoader, if not provided
  if (!options.clientLoader) {
    assert(options.credentials, "Either credentials or clientLoader is required");
    options.clientLoader = clientLoader(options.credentials);
  }

  // Create router
  var router = express.Router();

  // Use JSON middleware
  router.use(express.json({
    limit:                options.inputLimit
  }));

  // Allow CORS requests to the API
  if (options.allowedCORSOrigin) {
    router.use(mountpoint, function(req, res, next) {
      res.header('Access-Control-Allow-Origin',   options.allowedCORSOrigin);
      res.header('Access-Control-Allow-Headers',  'X-Requested-With,Content-Type');
      next();
    });
  }

  // Add entries to router
  this._entries.forEach(function(entry) {
    router[entry.method](
      // Route pattern
      entry.route,
      // Middleware
      authenticate(options.nonceManager, options.clientLoader, entry),
      schema(entry),
      handle(entry.handler, options.context)
    );
  });

  // Return router
  return router;
};


/**
 * Construct API reference as JSON
 *
 * options:
 * {
 *   baseUrl:    'https://example.com/v1' // URL under which routes are mounted
 * }
 */
API.prototype.reference = function(options) {
  assert(options.baseUrl, "A 'baseUrl' must be provided");
  return {
    version:            '0.2.0',
    title:              this._options.title,
    description:        this._options.desc,
    baseUrl:            options.baseUrl,
    entries: this._entries.map(function(entry) {
      // Find parameters for entry
      var params = [];
      var route = entry.route.replace(/\/:[^/]+/g, function(param) {
        param = param.substr(2);
        params.push(param);
        return '/<' + param + '>';
      });
      return {
        method:         entry.method,
        route:          route,
        args:           args,
        name:           entry.name,
        scopes:         entry.scopes,
        input:          entry.input,
        output:         entry.output,
        title:          entry.title,
        description:    entry.desc
      };
    })
  };
};

/**
 * Publish API reference to URL with given end-point
 *
 * options:
 * {
 *   baseUrl:         'https://example.com/v1' // URL under which routes are mounted
 *   referencePrefix: 'queue/v1/api.json'      // Prefix within S3 bucket
 *   referenceBucket: 'reference.taskcluster.net',
 *   aws: {             // AWS credentials and region
 *    accessKeyId:      '...',
 *    secretAccessKey:  '...',
 *    region:           'us-west-2'
 *   }
 * }
 *
 * Return a promise that reference was published.
 */
API.prototype.publish = function(options) {
  ['baseUrl', 'referencePrefix', 'referenceBucket'].forEach(function(key) {
    assert(options[key], "Option '" + key + "' must be provided");
  });
  var s3 = new aws.S3(options.aws);
  return s3.putObject({
    Bucket:           options.referenceBucket,
    Key:              referencePrefix,
    Body:             JSON.stringify(this.reference(options), undefined, 4),
    ContentType:      'application/json'
  }).promise();
};

/**
 * Setup API, by publishing reference and returning an `express.Router`.
 *
 * options:
 * {
 *   inputLimit:          '10mb'  // Max input JSON size
 *   allowedCORSOrigin:   '*'     // Allowed CORS origin, null to disable CORS
 *   context:             {}      // Object to be provided as `this` in handlers
 *   validator:           new base.validator()      // JSON schema validator
 *   nonceManager:        function(nonce, ts, cb) { // Check for replay attack
 *   clientLoader:        function(clientId) {      // Return promise for client
 *   credentials: {
 *     clientId:          '...',  // TaskCluster clientId
 *     accessToken:       '...'   // Access token for clientId
 *     // Client must have the 'auth:credentials' scope.
 *   },
 *   publish:             true,                    // Publish API reference
 *   baseUrl:             'https://example.com/v1' // URL under which routes are mounted
 *   referencePrefix:     'queue/v1/api.json'      // Prefix within S3 bucket
 *   referenceBucket:     'reference.taskcluster.net',
 *   aws: {               // AWS credentials and region
 *    accessKeyId:        '...',
 *    secretAccessKey:    '...',
 *    region:             'us-west-2'
 *   }
 * }
 *
 * The option `validator` must provided, and either `credentials` or
 * `clientLoader` must be provided.
 *
 * Return an `express.Router` instance.
 */
API.prototype.setup = function(options) {
  var that = this;
  return Promise.from().then(function() {
    if (options.publish) {
      return that.publish(options);
    }
  }).then(function() {
    return that.router(options);
  });
};

// Export API
module.exports = API;

// Export middleware utilities
API.authenticate  = authenticate;
API.schema        = schema;
API.handle        = handle;
