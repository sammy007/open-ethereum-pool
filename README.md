# go-ethereum-pool
## High Performance Ethereum Mining Pool

![Miner's stats page](https://15254b2dcaab7f5478ab-24461f391e20b7336331d5789078af53.ssl.cf1.rackcdn.com/ethereum.vanillaforums.com/editor/pe/cf77cki0pjpt.png)

### Features

* Highly available mining endpoint module
* Payouts and block unlocking (maturity) module
* Configurable payouts period and balance threshold
* PROP payouts (miners are simply paid out when a block is found)
* Detailed block stats with luck percentage and full reward
* Failover geth instances: geth high availability built in
* Strict policy module (banning strategies using ipset/iptables)
* Designed for 100% distributed setup of all modules
* Modern beautiful Ember.js frontend
* Separate stats for workers: can highlight timed-out workers so miners can perform maintenance of rigs
* JSON-API for stats, miners can use for rigs maintenance automation (rig rebooting for example )
Also it's compatible with my *ether-proxy*  solo/pool proxy solution.

*Written in Go it's a rocket highly concurrent and low RAM consuming piece of code*

### Building on Linux

Dependencies:

  * go >= 1.4
  * geth
  * redis-server
  * nodejs
  * nginx

**I highly recommend to use Ubuntu 14.04 LTS.**

First of all you must install  [go-ethereum](https://github.com/ethereum/go-ethereum/wiki/Installation-Instructions-for-Ubuntu).

Export GOPATH:

    export GOPATH=$HOME/go

Install required packages:

    go get github.com/ethereum/ethash
    go get github.com/ethereum/go-ethereum/common
    go get github.com/gorilla/mux
    go get github.com/yvasiyarov/gorelic

Compile:

    go build -o ether-pool main.go

Install redis-server and all software you need too.
Install nodejs, I suggest to use >= 4.x LTS version from https://github.com/nodesource/distributions or from your Linux distribution.

### Building on Windows

It's a little bit crazy to run production pool on this platform, but you can follow
[geth building instructions](https://github.com/ethereum/go-ethereum/wiki/Installation-instructions-for-Windows) and compile pool this way.
Use some cloud Redis provider or give a try to https://github.com/MSOpenTech/redis/releases.

### Building Frontend

Frontend is a single-page Ember.js application. It polls API of the pool
to render pool stats to miners.

    cd www

Change <code>ApiUrl: '//example.net/'</code> to match your domain name.

    npm install -g ember-cli@2.4.3
    npm install -g bower
    npm install
    bower install
    ./build.sh

Configure nginx to serve API on <code>/api</code> subdirectory.
Configure your nginx instance to serve <code>www/dist</code> as static website.

#### Serving API using nginx

Create an upstream for API:

    upstream api {
        server 127.0.0.1:8080;
    }

and add this setting after <code>location /</code>:

    location /api {
        proxy_pass http://api;
    }

You can customize layout and other stuff using built-in web server with live reload:

    ember server --port 8082 --environment development

#### Customization

Check out <code>www/app/templates</code> directory and edit these templates
in order to add your own branding and contacts.

### Configuration

Configuration is actually simple, just read it twice and think twice before changing defaults.

**Don't copy config directly from this manual, use example config from package,
Otherwise you will get errors on start because JSON can't contain comments actually.**

```javascript
{
  // Set to a number of CPU threads of your server
  "threads": 2,
  // Prefix for keys in redis store
  "coin": "eth",
  // Give unique name to each instance
  "name": "main",

  "proxy": {
    "enabled": true,

    // Bind mining endpoint to this IP:PORT
    "listen": "0.0.0.0:8546",

    // Allow only this header and body size of HTTP request from miners
    "limitHeadersSize": 1024,
    "limitBodySize": 256,

    /* Use it if you are behind CloudFlare (bad idea) or behind http-reverse
      proxy to enable IP detection from X-Forwarded-For header.
      Advanced users only. It's tricky to make it right and secure.
    */
    "behindReverseProxy": false,

    // Try to get new job from geth in this interval
    "blockRefreshInterval": "120ms",
    "stateUpdateInterval": "3s",
    // Require this share difficulty from miners
    "difficulty": 2000000000,

    "hashrateExpiration": "30m",

    /* Reply error to miner instead of job if redis is unavailable.
    Should save electricity to miners if pool is sick and they didn't set up failovers.
    */
    "healthCheck": true,
    // Mark pool sick after this number of redis failures.
    "maxFails": 100,

    "policy": {
      "workers": 8,
      "resetInterval": "60m",
      "refreshInterval": "1m",

      "banning": {
        "enabled": false,
        /* Name of ipset for banning.
        Check http://ipset.netfilter.org/ documentation.
        */
        "ipset": "blacklist",
        // Remove ban after this amount of time
        "timeout": 1800,
        // Percent of invalid shares from all shares to ban miner
        "invalidPercent": 30,
        // Check after after miner submitted this number of shares
        "checkThreshold": 30,
        // Bad miner after this number of malformed requests
        "malformedLimit": 5
      }
    }
  },

  // Provides JSON data for frontend which is static website
  "api": {
    "enabled": true,

    /* If you are running API node in a distributed environment where this module
      is reading data from redis writeable slave, enable this option.
      Only redis writeable slave will work properly if you are distributing using redis slaves.
      Don't distribute!
    */
    "purgeOnly": false,
    "listen": "0.0.0.0:8080",
    // Collect miners stats (hashrate, ...) in this interval
    "statsCollectInterval": "5s",

    // Fast hashrate estimation window for each miner from it's shares
    "hashrateWindow": "30m",
    // Long and precise hashrate from shares, 3h is cool, keep it
    "hashrateLargeWindow": "3h",
    // Max number of payments to display in frontend
    "payments": 50,
    // Max numbers of blocks to display in frontend
    "blocks": 50
  },

  // Check health of each geth node in this interval
  "upstreamCheckInterval": "5s",

  /* List of geth nodes to poll for new jobs. Pool will try to get work from
    first alive one and check in background for failed to back up.
    Current block template of the pool is always cached in RAM indeed.
  */
  "upstream": [
    {
      "name": "main",
      "url": "http://127.0.0.1:8545",
      "timeout": "10s"
    },
    {
      "name": "backup",
      "url": "http://127.0.0.2:8545",
      "timeout": "10s"
    }
  ],

  // This is standard redis connection options
  "redis": {
    // Where your redis instance is listening for commands
    "endpoint": "127.0.0.1:6379",
    "poolSize": 8,
    "database": 0,
    /* Generate and specify very strong password for in redis
      configuration file and specify it here.
      This is done using the requirepass directive in the configuration file.
      */
    "password": "secret"
  },

  // This module periodically credits coins to miners
  "unlocker": {
    "enabled": false,
    // Pool fee percentage
    "poolFee": 1.0,
    // Unlock only if this number of blocks mined back
    "depth": 120,
    // Simply don't touch this option
    "immatureDepth": 20,
    // Run unlocker in this interval
    "interval": "10m",
    // Geth instance node rpc endpoint for unlocking blocks
    "daemon": "http://127.0.0.1:8545",
    // Rise error if can't reach geth in this amount of time
    "timeout": "10s"
  },

  // Paying out miners using this module
  "payouts": {
    "enabled": false,
    // Run payouts in this interval
    "interval": "12h",
    // Geth instance node rpc endpoint for payouts processing
    "daemon": "http://127.0.0.1:8545",
    // Rise error if can't reach geth in this amount of time
    "timeout": "10s",
    // Address with pool balance
    "address": "0x0",
    // Let geth to determine gas and gasPrice
    "autoGas": true,
    // Gas amount and price for payout tx
    "gas": "21000",
    "gasPrice": "50000000000",
    // Seend payment only if miner's balance is >= 0.5 Ether
    "threshold": 500000000
  },
}
```

If you are distributing your pool deployment to several servers or processes,
create several configs and disable unneeded modules on each server.
This is very advanced, better don't distribute to several servers until you really need it.

I recommend this deployment strategy:

* Mining instance - 1x (it depends, you can run one node for EU, one for US, one for Asia)
* Unlocker and payouts instance - 1x each (strict!)
* API instance - 1x

### Notes

Unlocking and payouts are sequential, 1st tx go, 2nd waiting for 1st to confirm and so on.
You can disable that in code. Also, keep in mind that unlocking and payouts will be stopped in case of any backend or geth failure.
You must restart module if you see such errors with *suspended* word, so I recommend to run unlocker and payouts in a separate processes.
Don't run payouts and unlocker as part of mining node.

### Credits

Made by sammy007.

### Donations

* **ETH**: [0xb85150eb365e7df0941f0cf08235f987ba91506a](https://etherchain.org/account/0xb85150eb365e7df0941f0cf08235f987ba91506a)

* **BTC**: [1PYqZATFuYAKS65dbzrGhkrvoN9au7WBj8](https://blockchain.info/address/1PYqZATFuYAKS65dbzrGhkrvoN9au7WBj8)
