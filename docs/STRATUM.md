# Stratum Mining Protocol

This is the description of stratum protocol used in this pool.

Stratum defines simple exception handling. Example of rejected share looks like:

```javascript
{ "id": 1, "jsonrpc": "2.0", "result": null, "error": { code: 23, message: "Invalid share" } }
```

Each response with exception is followed by disconnect.

## Authentication

Request looks like:

```javascript
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "eth_submitLogin",
  "params": ["0xb85150eb365e7df0941f0cf08235f987ba91506a"]
}
```

Request can include additional 2nd param (email for example):

```javascript
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "eth_submitLogin",
  "params": ["0xb85150eb365e7df0941f0cf08235f987ba91506a", "admin@example.net"]
}
```

Successful response:

```javascript
{ "id": 1, "jsonrpc": "2.0", "result": true }
```

Exceptions:

```javascript
{ "id": 1, "jsonrpc": "2.0", "result": null, "error": { code: -1, message: "Invalid login" } }
```

## Request For Job

Request looks like:

```javascript
{ "id": 1, "jsonrpc": "2.0", "method": "eth_getWork" }
```

Successful response:

```javascript
{
  "id": 1,
  "jsonrpc": "2.0",
  "result": [
      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
      "0x5eed00000000000000000000000000005eed0000000000000000000000000000",
      "0xd1ff1c01710000000000000000000000d1ff1c01710000000000000000000000"
    ]
}
```

Exceptions:

```javascript
{ "id": 10, "result": null, "error": { code: 0, message: "Work not ready" } }
```

## New Job Notification

Server sends job to peers if new job is available:

```javascript
{
  "jsonrpc": "2.0",
  "result": [
      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
      "0x5eed00000000000000000000000000005eed0000000000000000000000000000",
      "0xd1ff1c01710000000000000000000000d1ff1c01710000000000000000000000"
    ]
}
```

## Share Submission

Request looks like:

```javascript
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "eth_submitWork",
  "params": [
    "0xe05d1fd4002d962f",
    "0x6c872e2304cd1e64b553a65387d7383470f22331aff288cbce5748dc430f016a",
    "0x2b20a6c641ed155b893ee750ef90ec3be5d24736d16838b84759385b6724220d"
  ]
}
```

Request can include optional `worker` param:

```javascript
{ "id": 1, "worker": "rig-1" /* ... */ }
```

Response:

```javascript
{ "id": 1, "jsonrpc": "2.0", "result": true }
{ "id": 1, "jsonrpc": "2.0", "result": false }
```

Exceptions:

Pool MAY return exception on invalid share submission usually followed by temporal ban.

```javascript
{ "id": 1, "jsonrpc": "2.0", "result": null, "error": { code: 23, message: "Invalid share" } }
```

```javascript
{ "id": 1, "jsonrpc": "2.0", "result": null, "error": { code: 22, message: "Duplicate share" } }
{ "id": 1, "jsonrpc": "2.0", "result": null, "error": { code: -1, message: "High rate of invalid shares" } }
{ "id": 1, "jsonrpc": "2.0", "result": null, "error": { code: 25, message: "Not subscribed" } }
{ "id": 1, "jsonrpc": "2.0", "result": null, "error": { code: -1, message: "Malformed PoW result" } }
```

## Submit Hashrate

`eth_submitHashrate` is a nonsense method. Pool ignores it and the reply is always:

```javascript
{ "id": 1, "jsonrpc": "2.0", "result": true }
```
