# LongLang

一个用 Go 语言实现的解释型编程语言。

## 特性

- ✅ 完整的类型系统（int, float, string, bool, any）
- ✅ 变量声明（var、短变量声明 :=）
- ✅ 控制流（if/else if/else、for 循环）
- ✅ 函数定义和调用（支持默认参数、命名参数）
- ✅ 面向对象（class、继承、接口、静态方法）
- ✅ 命名空间系统（namespace、use）
- ✅ 数组支持（固定长度、动态长度、多维数组）
- ✅ 异常处理（try-catch-finally、throw）
- ✅ 三目运算符
- ✅ 内置函数（fmt.println、fmt.print、fmt.printf、len）

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
| [数组](docs/array.md) | 固定长度数组、动态数组、多维数组 |
| [运算符](docs/operators.md) | 算术、比较、逻辑运算符 |
| [命名空间](docs/namespace.md) | namespace、use、项目结构 |
| [关键字](docs/keywords.md) | 语言保留关键字列表 |

### 面向对象

| 文档 | 说明 |
|------|------|
| [类基础](docs/class-basics.md) | 类定义、实例化、this、静态方法 |
| [类继承](docs/class-inheritance.md) | extends、方法重写、super |
| [类常量](docs/class-constants.md) | 常量定义、访问、类型声明 |
| [接口](docs/class-interface.md) | interface、implements、多接口 |

### 进阶文档

| 文档 | 说明 |
|------|------|
| [标准库](docs/stdlib.md) | System.IO、System.Exception 等标准库 |
| [三目运算符](docs/ternary.md) | 三目运算符使用规范 |

### 开发者文档

| 文档 | 说明 |
|------|------|
| [开发者指南](docs/developer-guide.md) | 架构、调试、添加新特性 |

## 快速入门

### Hello World

```longlang
namespace App

class Application {
    public static function main() {
        fmt.println("Hello, World!")
    }
}
```

### 变量和运算

```longlang
namespace App

class Application {
    public static function main() {
        // 变量声明
        name := "LongLang"
        version := 1.0
        
        // 算术运算
        a := 10
        b := 3
        fmt.println("a + b =", a + b)
        fmt.println("a * b =", a * b)
        
        // 字符串拼接
        greeting := "Hello, " + name
        fmt.println(greeting)
    }
}
```

### 控制流

```longlang
namespace App

class Application {
    public static function main() {
        score := 85
        
        // if-else
        if score >= 90 {
            fmt.println("优秀")
        } else if score >= 60 {
            fmt.println("及格")
        } else {
            fmt.println("不及格")
        }
        
        // for 循环
        for i := 0; i < 5; i++ {
            fmt.println("i =", i)
        }
    }
}
```

### 函数

```longlang
namespace App

class MathUtils {
    public static function add(a:int, b:int) int {
        return a + b
    }
}

class Application {
    public static function main() {
        result := MathUtils::add(10, 20)
        fmt.println("10 + 20 =", result)
    }
}
```

### 类和对象

```longlang
namespace App

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

class Application {
    public static function main() {
        person := new Person("Alice", 25)
        fmt.println(person.greet())
        fmt.println("Age:", person.age)
    }
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

- ⏳ Map 支持
- ⏳ 标准库扩展

## 许可证

Apache License 2.0
