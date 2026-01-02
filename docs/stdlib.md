# 标准库

LongLang 的标准库**使用 LongLang 语言自身编写**，存放在 `stdlib/` 目录下。

## 设计理念

- **自举**：标准库用 LongLang 编写，展示语言能力
- **内置函数**：仅 `fmt` 等底层 I/O 用 Go 实现（必须与系统交互）
- **命名空间**：标准库使用 `System` 等命名空间组织

## 导入标准库

使用 `use` 语句导入标准库中的类：

```longlang
use System.Str
use System.String
```

## fmt - 格式化输出（内置）

`fmt` 是内置库，无需导入即可使用。由 Go 实现，因为需要与系统 I/O 交互。

```longlang
fmt.println("Hello, World!")
fmt.print("不换行")
fmt.printf("格式化: %s", "值")
```

| 函数 | 说明 | 示例 |
|------|------|------|
| `println(args...)` | 打印并换行 | `fmt.println("hello", 123)` |
| `print(args...)` | 打印不换行 | `fmt.print("hello")` |
| `printf(format, args...)` | 格式化打印 | `fmt.printf("num: %d", 10)` |

## System 命名空间

### System.Str - 字符串静态工具类

提供字符串操作的静态方法：

```longlang
use System.Str

fmt.println(Str::length("hello"))            // 5
fmt.println(Str::upper("hello"))             // HELLO
fmt.println(Str::contains("hello", "ell"))   // true
fmt.println(Str::replace("hello", "l", "L")) // heLlo
```

### System.String - 字符串对象类

提供面向对象的字符串封装：

```longlang
use System.String

name := new String("hello world")
fmt.println(name.length())                   // 11
fmt.println(name.upper().getValue())         // HELLO WORLD

// 链式调用
result := name.trim().upper().replace("WORLD", "LONGLANG")
fmt.println(result.getValue())
```

详细方法列表请参阅 [字符串文档](string.md)。

## 字符串语法糖

原始字符串可以直接调用方法，无需导入任何库：

```longlang
name := "hello world"
fmt.println(name.length())          // 11
fmt.println(name.upper())           // HELLO WORLD
fmt.println(name.contains("world")) // true

// 链式调用
result := name.trim().upper().replace("WORLD", "LONGLANG")
```

## System.Redis - Redis 客户端

提供生产级 Redis 客户端，支持完整的 RESP 协议。

### 基本使用

```longlang
use System.Redis.RedisClient

// 简单连接
client := RedisClient::connect("127.0.0.1", 6379)

// 基本操作
client.set("name", "LongLang")
name := client.get("name")
fmt.println(name)  // LongLang

client.close()
```

### 带认证连接

```longlang
use System.Redis.RedisClient

// 方式 1: 仅密码
client := RedisClient::connectWithAuth("127.0.0.1", 6379, "password")

// 方式 2: 用户名 + 密码 (Redis 6.0+ ACL)
client := RedisClient::connectWithUserAuth("127.0.0.1", 6379, "default", "password")
```

### 使用配置对象

```longlang
use System.Redis.RedisClient
use System.Redis.RedisConfig

config := new RedisConfig()
config.setHost("127.0.0.1")
      .setPort(6379)
      .setPassword("mypassword")
      .setUsername("default")        // 可选，Redis 6.0+ ACL
      .setDatabase(1)                // 默认数据库
      .setPrefix("myapp:")           // 键前缀
      .setConnectTimeout(5000)       // 连接超时
      .setMaxRetries(3)              // 最大重试次数

client := RedisClient::connectWithConfig(config)

// 使用前缀后，所有键自动添加前缀
client.set("user:1", "Alice")  // 实际键为 myapp:user:1
```

### 支持的数据类型

| 数据类型 | 方法 |
|----------|------|
| 字符串 | `set`, `get`, `setEx`, `setNx`, `incr`, `decr`, `append` |
| 哈希 | `hset`, `hget`, `hdel`, `hgetall`, `hkeys`, `hvals`, `hlen` |
| 列表 | `lpush`, `rpush`, `lpop`, `rpop`, `lrange`, `llen`, `lindex` |
| 集合 | `sadd`, `srem`, `sismember`, `smembers`, `scard` |
| 有序集合 | `zadd`, `zrange`, `zrevrange`, `zscore`, `zrank`, `zcard` |
| 键操作 | `del`, `exists`, `expire`, `ttl`, `keys`, `rename` |

### 服务器管理

```longlang
// 选择数据库
client.selectDb(1)

// 获取服务器信息
info := client.info()
dbSize := client.dbSize()

// 清空数据库
client.flushDb()      // 清空当前数据库
client.flushAll()     // 清空所有数据库（慎用）
```

## 目录结构

```
longlang/
├── stdlib/
│   └── System/
│       ├── Exception.long           # 异常基类
│       ├── RuntimeException.long    # 运行时异常
│       ├── IOException.long         # IO 异常
│       ├── FileNotFoundException.long
│       ├── DirectoryNotFoundException.long
│       ├── PermissionException.long
│       ├── Str.long                 # 字符串静态工具类
│       ├── String.long              # 字符串对象类
│       ├── IO/
│       │   ├── File.long            # 文件操作
│       │   ├── Directory.long       # 目录操作
│       │   ├── Path.long            # 路径操作
│       │   ├── FileStream.long      # 文件流
│       │   ├── FileInfo.long        # 文件信息
│       │   └── DirectoryInfo.long   # 目录信息
│       ├── Net/
│       │   ├── TcpListener.long     # TCP 服务器
│       │   ├── TcpClient.long       # TCP 客户端
│       │   ├── TcpConnection.long   # TCP 连接
│       │   └── SocketException.long # 网络异常
│       ├── Http/
│       │   ├── HttpServer.long      # HTTP 服务器
│       │   ├── HttpRequest.long     # HTTP 请求
│       │   ├── HttpResponse.long    # HTTP 响应
│       │   └── HttpStatus.long      # HTTP 状态码枚举
│       ├── Redis/
│       │   ├── RedisClient.long     # Redis 客户端
│       │   ├── RedisConfig.long     # Redis 配置
│       │   └── RedisException.long  # Redis 异常
│       └── Binary/
│           ├── Bytes.long           # 字节操作
│           ├── ByteBuffer.long      # 字节缓冲区
│           ├── BinaryReader.long    # 二进制读取
│           └── BinaryWriter.long    # 二进制写入
├── internal/
│   └── interpreter/
│       ├── builtins.go         # fmt 等内置函数（Go）
│       ├── builtins_io.go      # 文件操作内置函数（Go）
│       └── string_methods.go   # 字符串方法（Go，支持语法糖）
└── ...
```

## 添加自定义库

你可以在 `stdlib/` 目录下添加自己的命名空间和类：

```longlang
// stdlib/MyLib/Utils.long
namespace MyLib

class Utils {
    public static function hello() string {
        return "Hello from MyLib!"
    }
}
```

然后在代码中使用：

```longlang
namespace App

use MyLib.Utils

class Application {
    public static function main() {
        fmt.println(Utils::hello())
    }
}
```
