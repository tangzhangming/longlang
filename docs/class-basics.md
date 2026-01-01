# 类基础

LongLang 支持面向对象编程。本文档介绍类的基本用法。

## 类的定义

使用 `class` 关键字定义类：

```longlang
class Person {
    // 成员变量
    public name string
    public age int

    // 构造函数
    public function __construct(name:string, age:int) {
        this.name = name
        this.age = age
    }

    // 实例方法
    public function greet():string {
        return "你好，我是 " + this.name
    }
}
```

## 一个文件多个类

**一个文件可以包含多个类，但只有与文件名相同的类才是导出类（可以被外部调用）。**

### 文件名导出规则

**核心规则：只有与文件名（不含扩展名）相同的类可以被外部访问。**

```longlang
// 文件：Redis.long
package cache

// ✅ 导出类（类名 Redis 与文件名 Redis.long 相同）
class Redis {
    private connection Connection  // 使用私有类
    
    public function __construct(host:string, port:int) {
        this.connection = new Connection(host, port)
    }
}

// ❌ 私有类（只能在文件内使用）
class Connection {
    private host string
    private port int
    
    public function __construct(host:string, port:int) {
        this.host = host
        this.port = port
    }
}
```

### 使用示例

```longlang
// main.long
package main

import "cache"

fn main() {
    // ✅ 可以使用导出类 Redis
    redis := new cache.Redis("localhost", 6379)
    
    // ❌ 错误！Connection 是私有类，不能访问
    // conn := new cache.Connection("localhost", 6379)  // 编译错误
}
```

### 文件名规则

| 规则 | 说明 | 示例 |
|------|------|------|
| 类名匹配文件名 | 导出类名必须与文件名（不含扩展名）相同 | `Redis.long` → `class Redis` |
| 大小写敏感 | 类名和文件名必须完全匹配（包括大小写） | `Redis.long` ≠ `redis.long` |
| 一个文件一个导出类 | 每个文件只能有一个与文件名相同的类 | - |
| 其他类是私有类 | 文件内其他类只能在文件内使用 | - |

详细说明请参考 [类多文件设计](../design/class-multi-file.md)。

## 访问修饰符

| 修饰符 | 说明 |
|-------|------|
| `public` | 公开成员，可以从类外部访问 |
| `private` | 私有成员，只能在类内部访问 |
| `protected` | 受保护成员，只能在类及其子类中访问 |

## 实例化

使用 `new` 关键字创建类的实例：

```longlang
person := new Person("张三", 25)
fmt.println(person.name)    // 张三
fmt.println(person.greet()) // 你好，我是 张三
```

## this 关键字

在类方法中，使用 `this` 关键字引用当前实例：

```longlang
public function setName(name:string) {
    this.name = name
}
```

## 构造函数

构造函数使用特殊名称 `__construct`：

```longlang
class Person {
    public name string
    
    public function __construct(name:string) {
        this.name = name
    }
}
```

## 静态方法和 self

使用 `static` 关键字定义静态方法，使用 `self` 关键字调用：

```longlang
class MathUtil {
    public static function add(a:int, b:int):int {
        return a + b
    }
    
    public static function double(n:int):int {
        return self::add(n, n)
    }
}

// 调用静态方法
result := MathUtil::add(1, 2)
```

## 成员变量默认值

成员变量可以设置默认值：

```longlang
class Config {
    public timeout int = 30
    public debug bool = false
    public name string = "default"
}
```

## 完整示例

```longlang
package main

import "fmt"

class Counter {
    public value int

    public function __construct(initial:int = 0) {
        this.value = initial
    }

    public function increment() {
        this.value = this.value + 1
    }

    public function decrement() {
        this.value = this.value - 1
    }

    public function getValue():int {
        return this.value
    }
}

fn main() {
    counter := new Counter(10)
    fmt.println("初始值: " + counter.getValue())
    
    counter.increment()
    fmt.println("递增后: " + counter.getValue())
    
    counter.decrement()
    fmt.println("递减后: " + counter.getValue())
}
```

## 相关文档

- [类继承](class-inheritance.md)
- [接口](class-interface.md)

