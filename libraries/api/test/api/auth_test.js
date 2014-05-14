suite("api/auth", function() {
  require('superagent-hawk')(require('superagent'));
  var request         = require('superagent-promise');
  var assert          = require('assert');
  var Promise         = require('promise');
  var mockAuthServer  = require('../mockauthserver');

  // Reference to mock authentication server
  var _server = null;

  // Create a mock authentication server
  setup(function(){
    assert(_server === null, "_server must be null");
    return mockAuthServer({
      port:         23243
    }).then(function(server) {
      _server = server;
    });
  });

  // Close server
  teardown(function() {
    assert(_server, "_server doesn't exist");
    return new Promise(function(accept, reject) {
      _server.once('close', function() {
        _server = null;
        accept();
      });
      _server.close();
    });
  });

  // Test getCredentials from mockAuthServer
  test("getCredentials from mockAuthServer", function() {
    var url = 'http://localhost:23243/client/test-client/credentials';
    return request
      .get(url)
      .hawk({
        id:           'test-client',
        key:          'test-token',
        algorithm:    'sha256'
      })
      .end()
      .then(function(res) {
        assert(res.ok,                                "Request failed");
        assert(res.body.accessToken === 'test-token', "Got wrong token");
      });
  });
});
