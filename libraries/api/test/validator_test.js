suite("validator", function() {
  var assert  = require('assert');
  var path    = require('path');
  var aws     = require('aws-sdk-promise');
  var base    = require('../');


  test("load from folder", function() {
    return base.validator({
      publish:      false,
      folder:       path.join(__dirname, 'schemas'),
      constants:    {"my-constant": 42}
    }).then(function(validator) {
      var errors = validator.check({
        value: 42
      }, 'http://localhost:1203/test-schema.json');
      assert(errors === null, "Got errors");
    });
  });

  test("find errors", function() {
    return base.validator({
      publish:      false,
      folder:       path.join(__dirname, 'schemas'),
      constants:    {"my-constant": 42}
    }).then(function(validator) {
      var errors = validator.check({
        value: 43
      }, 'http://localhost:1203/test-schema.json');
      assert(errors !== null, "Got no errors");
    });
  });

  test("test publish", function() {
    var cfg = base.config({
      defaults: {
        aws:                  {},
        schemaTestBucket:     'test-bucket-for-any-garbage'
      },
      filename:               'taskcluster-base-test'
    });

    if (cfg.get('aws') && cfg.get('schemaTestBucket')) {
      return base.validator({
        publish:      true,
        schemaPrefix: 'base/test/',
        schemaBucket: cfg.get('schemaTestBucket'),
        aws:          cfg.get('aws'),
        folder:       path.join(__dirname, 'schemas'),
        constants:    {"my-constant": 42}
      }).then(function(validator) {
        var errors = validator.check({
          value: 42
        }, 'http://localhost:1203/test-schema.json');
        assert(errors === null, "Got errors");

        // Get the file... we don't bother checking the contents this is good
        // enough
        var s3 = new aws.S3(cfg.get('aws'));
        return s3.getObject({
          Bucket:     cfg.get('schemaTestBucket'),
          Key:        'base/test/test-schema.json'
        }).promise();
      });
    } else {
      console.log("Skipping 'publish', missing config file: " +
                  "taskcluster-base-test.conf.json");
    }
  });
});
