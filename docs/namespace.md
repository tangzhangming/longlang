# 命名空间系统

LongLang 使用命名空间（namespace）来组织代码，支持模块化开发。类似于 Java 和 C# 的命名空间机制。

## 基本概念

### 命名空间声明

每个 LongLang 源文件使用 `namespace` 声明所属的命名空间：

```longlang
namespace Models

class User {
    public name string
    
    public function __construct(name: string) {
        this.name = name
    }
}
```

### 完全限定名

命名空间支持点分隔的层级结构：

```longlang
namespace Mycompany.Myapp.Models

class User {
    // ...
}
```

## 命名空间语法

| 语法 | 说明 | 示例 |
|------|------|------|
| `namespace Name` | 简单命名空间 | `namespace Models` |
| `namespace A.B.C` | 层级命名空间 | `namespace Mycompany.Myapp.Models` |

## use 导入

使用 `use` 语句导入其他命名空间中的类：

```longlang
namespace App

use Mycompany.Myapp.Models.User

class Application {
    public static function main() {
        user := new User("Alice")
        fmt.println(user.name)
    }
}
```

### 命名空间别名

可以为导入的类指定别名，避免命名冲突或简化长类名：

```longlang
namespace App

// 使用 as 关键字定义别名
use Mycompany.Myapp.Models.User as UserModel
use System.Redis.RedisConfig as Config
use Vendor.Database.Connection as DbConnection

class Application {
    public static function main() {
        user := new UserModel("Alice")
        config := new Config()
        db := new DbConnection()
    }
}
```

### 同命名空间自动引用

**同一命名空间下的类无需 `use` 即可直接引用**，类似 PHP Composer 的命名空间规则：

```longlang
// File: System/Redis/RedisClient.long
namespace System.Redis

use System.Net.TcpClient  // 跨命名空间需要 use

// 同命名空间的 RedisConfig 和 RedisException 无需 use，自动可用
class RedisClient {
    private _config any
    
    public static function connectWithConfig(config: RedisConfig) RedisClient {
        // 直接使用同命名空间的 RedisConfig，无需 use
        client := new RedisClient()
        // ...
        return client
    }
    
    private function handleError(msg: string) {
        // 直接使用同命名空间的 RedisException，无需 use
        throw new RedisException(msg)
    }
}
```

#### 自动引用规则

| 情况 | 是否需要 use | 示例 |
|------|-------------|------|
| 同命名空间的类 | ❌ 不需要 | `System.Redis.RedisClient` 引用 `System.Redis.RedisConfig` |
| 不同命名空间的类 | ✅ 需要 | `System.Redis.RedisClient` 引用 `System.Net.TcpClient` |
| 子命名空间的类 | ✅ 需要 | `System` 引用 `System.IO.File` |
| 父命名空间的类 | ✅ 需要 | `System.IO` 引用 `System.Exception` |

### 循环依赖处理

LongLang 支持同命名空间内的循环引用：

```longlang
// File: Models/Author.long
namespace App.Models

class Author {
    public name string
    
    // Author 引用 Book（同命名空间）
    public function createBook(title: string) Book {
        return new Book(title, this)
    }
}
```

```longlang
// File: Models/Book.long
namespace App.Models

class Book {
    public title string
    public author Author  // Book 引用 Author（循环引用）
    
    public function __construct(title: string, author: Author) {
        this.title = title
        this.author = author
    }
}
```

> **注意**：跨命名空间的循环依赖应尽量避免，可能导致加载问题。

## project.toml 配置

LongLang 项目使用 `project.toml` 配置文件定义项目信息和根命名空间：

```toml
[project]
name = "myapp"
version = "1.0.0"
root_namespace = "Mycompany.Myapp"
```

### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `name` | 项目名称 | 必填 |
| `version` | 项目版本 | 必填 |
| `root_namespace` | 根命名空间 | 无 |
| `source_path` | 源代码目录 | `src` |
| `vendor_path` | 依赖目录 | `vendor` |

### 命名空间简化

当配置了 `root_namespace` 后，可以使用简化的命名空间声明：

```toml
# project.toml
[project]
root_namespace = "Mycompany.Myapp"
```

```longlang
// src/Models/User.long
namespace Models  // 等同于 namespace Mycompany.Myapp.Models

class User {
    // ...
}
```

## 项目结构

推荐的项目结构：

```
myproject/
├── project.toml           # 项目配置文件
├── src/                   # 源代码目录
│   ├── Application.long   # 入口类
│   ├── Models/
│   │   └── User.long      # 模型类
│   └── Services/
│       └── UserService.long
├── vendor/                # 第三方依赖（可选）
└── README.md
```

## 程序入口

LongLang 程序的入口是包含 `main` 静态方法的类：

```longlang
namespace App

class Application {
    public static function main() {
        fmt.println("Hello, World!")
    }
}
```

### 入口规则

1. 程序入口必须是一个类的静态 `main` 方法
2. 入口文件中只能有一个类包含 `main` 方法
3. 如果有多个类都有 `main` 方法，将报错

## 类的可见性

在同一个文件中可以定义多个类，但只有与文件名同名的类可以被外部访问：

```longlang
// User.long
namespace Models

// User 类可以被外部访问（与文件名同名）
class User {
    public name string
}

// Helper 类只能在本文件内部使用
class Helper {
    public static function format(s: string) string {
        return s
    }
}
```

## 示例

### 完整项目示例

**project.toml**
```toml
[project]
name = "myapp"
version = "1.0.0"
root_namespace = "App"
```

**src/Application.long**
```longlang
namespace App

use App.Models.User
use App.Services.UserService

class Application {
    public static function main() {
        service := new UserService()
        user := service.createUser("Alice", 25)
        fmt.println("Created user:", user.name)
    }
}
```

**src/Models/User.long**
```longlang
namespace Models

class User {
    public name string
    public age int
    
    public function __construct(name: string, age: int) {
        this.name = name
        this.age = age
    }
}
```

**src/Services/UserService.long**
```longlang
namespace Services

use App.Models.User

class UserService {
    public function createUser(name: string, age: int) User {
        return new User(name, age)
    }
}
```

### 运行程序

```bash
cd myproject
longlang run src/Application.long
```

## 注意事项

1. **命名空间必须在文件开头**：`namespace` 语句必须是文件的第一个语句（注释除外）
2. **use 必须在类定义之前**：`use` 语句必须在 `namespace` 之后、类定义之前
3. **同命名空间无需 use**：同一命名空间下的类、枚举、接口可以直接引用
4. **循环依赖**：同命名空间内的循环引用是允许的，跨命名空间应尽量避免
5. **命名规范**：命名空间使用 PascalCase 风格

## 最佳实践

| 实践 | 说明 |
|------|------|
| 一文件一主类 | 每个文件定义一个主要的公开类 |
| 目录结构匹配命名空间 | 目录结构应该与命名空间层级对应 |
| 使用 root_namespace | 配置根命名空间简化代码中的命名空间声明 |
| 合理分层 | 按功能分层：Models、Services、Controllers 等 |



