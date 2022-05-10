/* jshint node: true */

module.exports = function(environment) {
  var ENV = {
    modulePrefix: 'open-ethereum-pool',
    environment: environment,
    rootURL: '/',
    locationType: 'hash',
    EmberENV: {
      FEATURES: {
        // Here you can enable experimental features on an ember canary build
        // e.g. 'with-controller': true
      }
    },

    APP: {
      // API host and port
      ApiUrl: '//phat.pool2mine.net/',
      PoolName: 'ETH Pool',
      CompanyName: 'phat.pool2mine.net',
      // HTTP mining endpoint
      HttpHost: 'http://phat.pool2mine.net',
      HttpPort: 8882,

      // Stratum mining endpoint
      StratumHost: 'phat.pool2mine.net:',
      StratumPort: 8002,

      // Fee and payout details
      PoolFee: '0.1%',
      PayoutThreshold: '0.1',
      PayoutInterval: '3h',

      // For network hashrate (change for your favourite fork)
      BlockTime: 13.3,
      BlockReward: 2.0,
      Unit: 'ETH',

    }
  };

  if (environment === 'development') {
    /* Override ApiUrl just for development, while you are customizing
      frontend markup and css theme on your workstation.
    */
    ENV.APP.ApiUrl = 'http://localhost:8080/'
    // ENV.APP.LOG_RESOLVER = true;
    // ENV.APP.LOG_ACTIVE_GENERATION = true;
    // ENV.APP.LOG_TRANSITIONS = true;
    // ENV.APP.LOG_TRANSITIONS_INTERNAL = true;
    // ENV.APP.LOG_VIEW_LOOKUPS = true;
  }

  if (environment === 'test') {
    // Testem prefers this...
    ENV.locationType = 'none';

    // keep test console output quieter
    ENV.APP.LOG_ACTIVE_GENERATION = false;
    ENV.APP.LOG_VIEW_LOOKUPS = false;

    ENV.APP.rootElement = '#ember-testing';
  }

  if (environment === 'production') {

  }

  return ENV;
};
