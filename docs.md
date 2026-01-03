# LongLang 语法快速参考

> 本文档面向成熟的开发者和 AI，提供 LongLang 语法的极简概览。详细文档请参考 `docs/` 目录。

## 目录

- [基础语法](#基础语法)
- [控制流](#控制流)
- [数据结构](#数据结构)
- [面向对象](#面向对象)
- [命名空间](#命名空间)
- [异常处理](#异常处理)
- [并发编程](#并发编程)
- [高级特性](#高级特性)

---

## 基础语法

### 变量声明

```longlang
// 短变量声明（推荐）
name := "LongLang"
age := 25
price := 99.99
isActive := true

// 显式类型声明
var count int = 10
var message string = "Hello"
var uninitialized int  // 默认值为 null
```

### 基本类型

```longlang
// 整数类型
var a int = 42        // 64位整数（平台相关）
var b i32 = 100       // 32位整数
var c u8 = 255        // 8位无符号整数

// 浮点数
var pi float = 3.14159
var e f64 = 2.71828

// 布尔和字符串
var flag bool = true
var text string = "Hello"

// 任意类型
var value any = 42
value = "now it's a string"
```

### 函数定义

```longlang
// 无返回值
function greet(name: string) {
    Console::writeLine("Hello, " + name)
}

// 有返回值
function add(a: int, b: int) int {
    return a + b
}

// 多返回值
function divide(a: int, b: int) (int, int) {
    return a / b, a % b
}

// 默认参数
function greet(name: string = "World") {
    Console::writeLine("Hello, " + name)
}

// 命名参数调用
greet(name: "Alice")
```

### 注释

```longlang
// 单行注释

/*
 * 多行注释
 */

/**
 * 文档注释
 * @param name 用户名
 * @return 问候语
 */
```

---

## 控制流

### 条件语句

```longlang
// if-else
if score >= 90 {
    grade := "A"
} else if score >= 80 {
    grade := "B"
} else {
    grade := "C"
}

// 三目运算符
result := score >= 60 ? "Pass" : "Fail"
```

### 循环

```longlang
// for 循环
for i := 0; i < 10; i++ {
    Console::writeLine(i)
}

// while 循环（使用 for）
for count < 100 {
    count++
}

// for-range 遍历
arr := []int{1, 2, 3}
for index, value := range arr {
    Console::writeLine(index, value)
}

// 遍历 Map
m := map[string]int{"a": 1, "b": 2}
for key, value := range m {
    Console::writeLine(key, value)
}

// 遍历字符串
str := "Hello"
for index, char := range str {
    Console::writeLine(index, char)
}
```

### Switch/Match

```longlang
// switch 语句
switch day {
    case 1:
        Console::writeLine("Monday")
    case 2:
        Console::writeLine("Tuesday")
    default:
        Console::writeLine("Other")
}

// match 表达式（模式匹配）
result := match status {
    200 => "OK"
    404 => "Not Found"
    500 => "Server Error"
    _ => "Unknown"
}
```

---

## 数据结构

### 数组

```longlang
// 数组声明
var arr1 []int = []int{1, 2, 3}
arr2 := []int{1, 2, 3}           // 类型推导
arr3 := []int{0, 0, 0}           // 初始化
var arr4 [10]int                 // 固定长度数组

// 数组操作
arr := []int{1, 2, 3}
arr.push(4)                      // 添加元素
arr.pop()                         // 移除最后一个
arr.shift()                       // 移除第一个
length := arr.length()            // 长度
contains := arr.contains(2)       // 是否包含
index := arr.indexOf(2)           // 查找索引

// 切片
slice := arr[1:3]                // [2, 3]
first := arr[:2]                 // [1, 2]
last := arr[2:]                  // [3]
all := arr[:]                    // 完整数组
```

### Map

```longlang
// Map 声明
var m1 map[string]int = map[string]int{"a": 1, "b": 2}
m2 := map[string]int{"a": 1, "b": 2}

// Map 操作
m := map[string]int{"a": 1}
m["b"] = 2                       // 添加/更新
value := m["a"]                  // 读取
exists := isset(m, "a")          // 检查键是否存在
m.delete("a")                    // 删除
size := m.size()                 // 大小
keys := m.keys()                 // 获取所有键
values := m.values()             // 获取所有值
```

### 字符串

```longlang
// 字符串操作
str := "Hello, World"
length := str.length()            // 长度
upper := str.upper()             // 转大写
lower := str.lower()             // 转小写
trimmed := str.trim()            // 去除首尾空白
contains := str.contains("World") // 是否包含
index := str.indexOf("World")     // 查找位置
substr := str.substring(0, 5)    // 子字符串
replaced := str.replace("World", "LongLang") // 替换
parts := str.split(",")          // 分割

// 字符串插值
name := "Alice"
age := 25
message := $"Hello, {name}, you are {age} years old"
```

---

## 面向对象

### 类定义

```longlang
class Person {
    // 成员变量
    private name string
    private age int

    // 构造函数
    public function __construct(name: string, age: int) {
        this.name = name
        this.age = age
    }

    // 实例方法
    public function getName() string {
        return this.name
    }

    // 静态方法
    public static function create(name: string, age: int) Person {
        return new Person(name, age)
    }
}

// 使用
person := new Person("Alice", 25)
name := person.getName()
```

### 继承

```longlang
class Animal {
    protected name string

    public function __construct(name: string) {
        this.name = name
    }

    public function speak() string {
        return "Some sound"
    }
}

class Dog extends Animal {
    public function speak() string {
        return this.name + " says Woof!"
    }

    public function bark() {
        Console::writeLine("Woof! Woof!")
    }
}

// 使用
dog := new Dog("Buddy")
dog.speak()  // "Buddy says Woof!"
```

### 接口

```longlang
interface Runnable {
    function run() void
}

interface Flyable {
    function fly() void
}

class Bird implements Runnable, Flyable {
    public function run() void {
        Console::writeLine("Bird is running")
    }

    public function fly() void {
        Console::writeLine("Bird is flying")
    }
}
```

### 抽象类

```longlang
abstract class Shape {
    abstract function area() float
    abstract function perimeter() float

    public function describe() string {
        return $"Area: {this.area()}, Perimeter: {this.perimeter()}"
    }
}

class Circle extends Shape {
    private radius float

    public function __construct(radius: float) {
        this.radius = radius
    }

    public function area() float {
        return 3.14159 * this.radius * this.radius
    }

    public function perimeter() float {
        return 2 * 3.14159 * this.radius
    }
}
```

### 枚举

```longlang
enum Status {
    Pending
    Processing
    Completed
    Failed
}

// 使用
status := Status.Pending
if status == Status.Completed {
    Console::writeLine("Done!")
}
```

### 类常量和静态成员

```longlang
class Math {
    public const PI = 3.14159
    public const E = 2.71828

    public static function max(a: int, b: int) int {
        return a > b ? a : b
    }
}

// 使用
pi := Math::PI
maxValue := Math::max(10, 20)
```

### 访问修饰符

```longlang
class Example {
    public publicField string      // 公开
    private privateField string     // 私有
    protected protectedField string // 受保护
    internal internalField string   // 内部（同命名空间可见）
}
```

---

## 命名空间

```longlang
// 声明命名空间
namespace MyApp.Models

class User {
    // ...
}

// 导入命名空间
namespace MyApp.Controllers

use MyApp.Models.User
use System.Console

class UserController {
    public function create() {
        user := new User("Alice", 25)
        Console::writeLine("User created")
    }
}
```

---

## 异常处理

```longlang
try {
    result := 10 / 0
} catch (ArithmeticException e) {
    Console::writeLine("Division by zero: " + e.getMessage())
} catch (Exception e) {
    Console::writeLine("Error: " + e.getMessage())
} finally {
    Console::writeLine("Cleanup")
}

// 抛出异常
if value < 0 {
    throw new InvalidArgumentException("Value must be positive")
}
```

---

## 并发编程

### 协程

```longlang
// 启动协程
go {
    Console::writeLine("Running in goroutine")
}

// 带参数的协程
go task(42)

function task(value: int) {
    Console::writeLine("Task with value: " + toString(value))
}
```

### Channel

```longlang
// 创建 Channel
ch := new Channel(10)  // 缓冲大小为 10

// 发送数据
go {
    ch.send("Hello")
}

// 接收数据
message := ch.receive()
```

### WaitGroup

```longlang
wg := new WaitGroup()

for i := 0; i < 5; i++ {
    wg.add(1)
    go {
        defer wg.done()
        Console::writeLine("Task " + toString(i))
    }
}

wg.wait()  // 等待所有协程完成
```

---

## 高级特性

### 类型断言

```longlang
var value any = "Hello"

// 安全类型断言（返回 null 如果失败）
str := value as? string
if str != null {
    Console::writeLine(str)
}

// 强制类型断言（抛出异常如果失败）
str2 := value as string
```

### 闭包

```longlang
function makeCounter() {
    count := 0
    return fn() int {
        count++
        return count
    }
}

counter := makeCounter()
counter()  // 1
counter()  // 2
```

### 注解

```longlang
@Route("/api/users")
class UserController {
    @GET
    public function list() {
        // ...
    }

    @POST
    public function create() {
        // ...
    }
}
```

### 正则表达式

```longlang
use System.Regex

pattern := new Regex("\\d+")
if pattern.isMatch("123") {
    match := pattern.match("123")
    Console::writeLine(match.getValue())
}
```

### 日期时间

```longlang
use System.DateTime

now := DateTime::now()
today := DateTime::today()
parsed := DateTime::parse("2024-01-01 12:00:00")

formatted := now.format("YYYY-MM-DD HH:mm:ss")
future := now.addDays(7)
```

---

## 标准库速览

```longlang
use System.Console
use System.IO.File
use System.Net.TcpClient
use System.Http.HttpServer
use Database.Mysql.Client
use Database.Redis.Client

// Console
Console::writeLine("Hello")
input := Console::readLine()

// 文件操作
content := File::readAll("file.txt")
File::writeAll("output.txt", "Hello")

// 网络
client := new TcpClient("localhost", 8080)
data := client.read(1024)

// HTTP 服务器
server := new HttpServer(":8080")
server.onRequest(fn(req, res) {
    res.write("Hello, World!")
})
server.start()

// MySQL
mysql := Client::connectSimple("localhost", 3306, "user", "pass", "db")
result := mysql.query("SELECT * FROM users")

// Redis
redis := RedisClient::connect("localhost", 6379)
redis.set("key", "value")
value := redis.get("key")
```

---

## 语法特点总结

| 特性 | 说明 |
|------|------|
| **类型系统** | 静态类型，支持类型推导 |
| **变量声明** | `var` 或 `:=` 短变量声明 |
| **函数** | 支持默认参数、命名参数、多返回值 |
| **面向对象** | 类、继承、接口、抽象类、枚举 |
| **并发** | 协程（`go`）、Channel、WaitGroup |
| **异常** | try-catch-finally |
| **命名空间** | `namespace` 和 `use` 导入 |
| **字符串** | 插值 `$"..."`、丰富的方法 |
| **数组/Map** | 内置方法、切片语法 |
| **类型断言** | `as` 和 `as?` |

---

## 快速对比

### 与 Go 相似
- 协程和 Channel
- 短变量声明 `:=`
- 错误处理风格（但用异常）

### 与 Java/C# 相似
- 类和继承
- 命名空间系统
- 异常处理
- 接口和抽象类

### 与 PHP 相似
- `$` 字符串插值
- `__construct` 构造函数
- 动态特性

### 独特特性
- `match` 表达式
- 命名参数调用
- 默认参数
- 类型断言 `as`/`as?`

---

**更多信息**: 查看 `docs/` 目录下的详细文档。

