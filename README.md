# LongLang

一个用 Go 语言实现的解释型编程语言。

## 特性

- ✅ 完整的类型系统（int, float, string, bool, any）
- ✅ 变量声明（var、短变量声明 :=）
- ✅ 控制流（if/else if/else、for 循环）
- ✅ 函数定义和调用（支持默认参数、命名参数）
- ✅ 面向对象（class、继承、静态方法）
- ✅ 三目运算符
- ✅ 内置函数（fmt.Println、fmt.Print、fmt.Printf）

## 安装

```bash
go build -o longlang.exe .
```

## 使用方法

```bash
longlang.exe run <文件路径>
```

例如：
```bash
longlang.exe run main.long
longlang.exe run test/test1_basic.long
```

## 📖 文档

详细文档请参阅 `docs/` 目录：

| 文档 | 说明 |
|------|------|
| [变量](docs/variables.md) | 变量声明、作用域、赋值 |
| [控制结构](docs/control-structures.md) | if/else、for 循环、break/continue |
| [函数](docs/functions.md) | 函数定义、参数、返回值 |
| [注释](docs/comments.md) | 单行注释用法 |
| [类型系统](docs/types.md) | 整数、浮点数、字符串、布尔类型 |
| [运算符](docs/operators.md) | 算术、比较、逻辑运算符 |
| [模块与包](docs/packages.md) | package、import、包管理 |
| [关键字](docs/keywords.md) | 语言保留关键字列表 |

### 进阶文档

| 文档 | 说明 |
|------|------|
| [类系统](docs/class-system-design.md) | 类定义、成员、方法 |
| [三目运算符](docs/ternary.md) | 三目运算符使用规范 |

### 开发者文档

| 文档 | 说明 |
|------|------|
| [开发者指南](docs/developer-guide.md) | 架构、调试、添加新特性 |

## 快速入门

### Hello World

```longlang
package main

fn main() {
    fmt.Println("Hello, World!")
}
```

### 变量和运算

```longlang
package main

fn main() {
    // 变量声明
    name := "LongLang"
    version := 1.0
    
    // 算术运算
    a := 10
    b := 3
    fmt.Println("a + b =", a + b)
    fmt.Println("a * b =", a * b)
    
    // 字符串拼接
    greeting := "Hello, " + name
    fmt.Println(greeting)
}
```

### 控制流

```longlang
package main

fn main() {
    score := 85
    
    // if-else
    if score >= 90 {
        fmt.Println("优秀")
    } else if score >= 60 {
        fmt.Println("及格")
    } else {
        fmt.Println("不及格")
    }
    
    // for 循环
    for i := 0; i < 5; i++ {
        fmt.Println("i =", i)
    }
}
```

### 函数

```longlang
package main

fn add(a:int, b:int) int {
    return a + b
}

fn greet(name:string = "World") {
    fmt.Println("Hello,", name)
}

fn main() {
    result := add(10, 20)
    fmt.Println("10 + 20 =", result)
    
    greet()
    greet("Alice")
}
```

### 类和对象

```longlang
package main

class Person {
    public name string
    public age int
    
    public function __construct(name:string, age:int) {
        this.name = name
        this.age = age
    }
    
    public function greet() string {
        return "Hello, I am " + this.name
    }
}

fn main() {
    person := new Person("Alice", 25)
    fmt.Println(person.greet())
    fmt.Println("Age:", person.age)
}
```

## 测试用例

在 `test` 目录下提供了多个测试用例：

```bash
# 运行基础测试
longlang.exe run test/test1_basic.long

# 运行类型测试
longlang.exe run test/test_types_integer.long
longlang.exe run test/test_types_float.long
longlang.exe run test/test_types_string.long

# 运行类测试
longlang.exe run test/test_class_basic.long
```

## 开发状态

### 已实现

- ✅ 词法分析（Lexer）
- ✅ 语法分析（Parser）
- ✅ 解释执行（Interpreter）
- ✅ 类型系统
- ✅ 面向对象
- ✅ 控制流语句
- ✅ 函数定义和调用

### 计划中

- ⏳ 数组和 Map 支持
- ⏳ 错误处理（try/catch）
- ⏳ 模块导入系统
- ⏳ 标准库扩展

## 许可证

Apache License 2.0
