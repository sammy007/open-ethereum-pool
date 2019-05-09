

###     1.登录认证

矿机启动，首先以etrue_submitLogin方法向矿池连接登录。


Client:

```
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "etrue_submitLogin",
  "params": [
    "0xb85150eb365e7df0941f0cf08235f987ba91506a", 
    "admin@example.net"
  ],
  "worker":"test"
}
```

Server:

```
{ "id": 1, "jsonrpc": "2.0", "result": true }

Exceptions:

{ "id": 1, "jsonrpc": "2.0", "result": null, "error": { code: -1, message: "Invalid login" } }
```

其中：

用户奖励地址:0xb85150eb365e7df0941f0cf08235f987ba91506a；

admin@example.net:用户邮箱,可选。



### 	2.任务请求

矿机向矿池请求新挖矿任务，矿池分配新任务。

Client:

```
{id":2,"jsonrpc": "2.0","method":"etrue_getWork"}

```

Server:

```
{ "id":2,
  "jsonrpc": "2.0",
  "method":"etrue_getWork",
  "result": [
    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	"0x5eed00000000000000000000000000005eed0000000000000000000000000000",
	"0x123456eb365e7df0941f0cf08235f98b123456eb365e7df0941f0cf08235f98b"
  ]
}

Exceptions:

{ "id": 2, "result": null, "error": { code: 0, message: "Work not ready" } }
```
**headerhash**: 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef。

**seedhash**：0x5eed00000000000000000000000000005eed0000000000000000000000000000。

**任务难度target**：0x123456eb365e7df0941f0cf08235f98b123456eb365e7df0941f0cf08235f98b；前32字符表示
block难度，后32字符表示fruit难度。



### 	3.任务分配

矿池定期发给矿机。挖矿参数更新则发送ID为0的result消息给所有的矿机。


Server:

```
{
  "id": 0,
  "jsonrpc": "2.0",
  "method": "etrue_notify",
  "params": [
    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	"0x5eed00000000000000000000000000005eed0000000000000000000000000000",
	"0x123456eb365e7df0941f0cf08235f98b123456eb365e7df0941f0cf08235f98b"
  ]
}
```

**headerhash**: 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef。

**seedhash**：0x5eed00000000000000000000000000005eed0000000000000000000000000000。

**任务难度target**：0x123456eb365e7df0941f0cf08235f98b123456eb365e7df0941f0cf08235f98b；前32字符表示
block难度，后32字符表示fruit难度；


### 	4.结果提交

矿机找到合法share时，就以”etrue_submitWork“方法向矿池提交任务。矿池返回true即结果被接受。

```
Request :

{
  "id": 3,
  "jsonrpc": "2.0",
  "method": "etrue_submitWork",
  "params": [
    "0x1060",
    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	"0x2b20a6c641ed155b893ee750ef90ec3be5d24736d16838b84759385b6724220d"
  ],
  "worker":"test"
}

Response:

{ "id": 3, "jsonrpc": "2.0", "result": true }
{ "id": 3, "jsonrpc": "2.0", "result": false }

Exceptions:

Pool MAY return exception on invalid share submission usually followed by temporal ban.

{ "id": 3, "jsonrpc": "2.0", "result": null, "error": { code: 23, message: "Invalid share" } }
{ "id": 3, "jsonrpc": "2.0", "result": null, "error": { code: 22, message: "Duplicate share" } }
{ "id": 3, "jsonrpc": "2.0", "result": null, "error": { code: -1, message: "High rate of invalid shares" } }
{ "id": 3, "jsonrpc": "2.0", "result": null, "error": { code: 25, message: "Not subscribed" } }
{ "id": 3, "jsonrpc": "2.0", "result": null, "error": { code: -1, message: "Malformed PoW result" } }

```
**minernonce**: 0x1060。minernonce为无符号64位的整数。

**headerhash**: 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef。

**mixhash**: 0x2b20a6c641ed155b893ee750ef90ec3be5d24736d16838b84759385b6724220d。



### 	5.申请种子哈希

矿工使用seedhash识别DataSet，如果不匹配则向矿池申请种子哈希来生成DataSet。

矿池应该马上发送种子哈希给矿机(10240个)。


Client:

```
{
  "id": 4,
  "jsonrpc": "2.0",
  "method": "etrue_seedhash",
  "params": [
    "0x5eed00000000000000000000000000005eed0000000000000000000000000000"
  ]
}
```

**seedhash**：0x5eed00000000000000000000000000005eed0000000000000000000000000000。


Server:

```
  "id": 4,
  "jsonrpc": "2.0",
  "method": "etrue_seedhash",
  "result": [
    [
      "0x323cf20198c2f3861e947d4f67e3ab63",
      "0xb7b2e24dcc9095bd9123e7b33371f6cc",
      "0x6510010198c2f3861e947d4f67e3ab63",
      "0xb7b2e24dcc9095bd9123e7b33371f6cc",
      ...
    ],
	"0x5eed00000000000000000000000000005eed0000000000000000000000000000"
  ],
  "error": null
```

**result**: 10240个用于构建DataSet的种子哈希

**seedhash**: 0x5eed00000000000000000000000000005eed0000000000000000000000000000用于验证构建后的DataSet



###     6.获取cpuminer版本

矿池获取矿机cpuminer版本（暂无）。


Server:

```
{
  "id": 5,
  "method": "etrue_get_version"
}
```


Client:

```
  "id": 5,
  "result": "cpuminer/0.1.0",
  "error": null
```

###     6.获取hashrate

矿池获取矿机获取hashrate。


Server:

```
{
  "id": 6,
  "method": "etrue_get_hashrate"
}
```


Client:

```
  "id": 6,
  "method": "etrue_get_hashrate",
  "result": "600",
  "error": null
```
**result**: 600 hash/s,hex
