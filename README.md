# longlang

一个用 Go 语言实现的解释型编程语言。

## 特性

- ✅ 基本表达式（算术、逻辑、比较）
- ✅ 变量声明（var、短变量声明 :=）
- ✅ 控制流（if/else if/else）
- ✅ 函数定义和调用
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

## 语法示例

### 变量声明

```long
// 带类型声明
var name string = "xiaoming"
var age int = 18

// 自动类型推导
var score = 95

// 短变量声明（只能在函数内使用）
count := 10

// 注意：未初始化的变量值为 null
var uninitialized int
```

### 基本运算

```long
var a = 10
var b = 3

var sum = a + b        // 加法
var diff = a - b      // 减法
var product = a * b   // 乘法
var quotient = a / b  // 除法
var remainder = a % b // 取模
```

### 比较和逻辑运算

```long
var a = 10
var b = 20

var isGreater = a > b      // 大于
var isLess = a < b         // 小于
var isEqual = a == b       // 等于
var isNotEqual = a != b    // 不等于
var and = a > 5 && b > 10  // 逻辑与
var or = a > 5 || b > 30  // 逻辑或
```

### 控制流

```long
var score = 85

if score >= 90 {
    fmt.Println("优秀")
} else if score >= 80 {
    fmt.Println("良好")
} else if score >= 60 {
    fmt.Println("及格")
} else {
    fmt.Println("不及格")
}
```

### 三目运算符

```long
var a = 10
var b = 20

var max = a > b ? a : b
var result = a > 5 ? "大于5" : "小于等于5"
```

### 函数定义

```long
// 无返回值
fn greet(name:string) {
    fmt.Println("Hello,", name)
}

// 有返回值（单返回值）
fn add(a:int, b:int): int {
    return a + b
}

// 多返回值
fn divide(a:int, b:int): (int, int) {
    var quotient = a / b
    var remainder = a % b
    return quotient, remainder
}

// 带默认参数
fn greet(name:string = "World") {
    fmt.Println("Hello,", name)
}
```

### 函数调用

```long
fn main() {
    greet("longlang")
    
    var sum = add(10, 20)
    fmt.Println("和:", sum)
    
    var q, r = divide(10, 3)
    fmt.Println("商:", q, "余数:", r)
}
```

### 内置函数

```long
fn main() {
    // 打印并换行
    fmt.Println("Hello", "World")
    
    // 打印不换行
    fmt.Print("Hello")
    fmt.Print("World")
    
    // 格式化打印
    fmt.Printf("数字: %d, 字符串: %s", 123, "test")
}
```

## 程序入口

所有程序必须有一个 `main` 函数作为入口点：

```long
fn main() {
    // 你的代码
}
```

## 测试用例

在 `test` 目录下提供了多个测试用例：

- `test1_basic.long` - 基本变量和打印
- `test2_arithmetic.long` - 算术运算
- `test3_if.long` - if 语句
- `test4_function.long` - 函数定义和调用
- `test5_ternary.long` - 三目运算符
- `test6_string.long` - 字符串操作
- `test7_short_declare.long` - 短变量声明
- `test8_complex.long` - 复杂示例

运行测试用例：

```bash
longlang.exe run test/test1_basic.long
```

## 注意事项

1. 所有程序必须有一个 `main` 函数
2. 短变量声明 `:=` 只能在函数内使用
3. 函数调用时参数数量必须匹配
4. 变量使用前必须先声明
5. 字符串使用双引号 `"` 包裹

## 开发状态

当前版本实现了基本的解释器功能，支持：
- ✅ 词法分析（Lexer）
- ✅ 语法分析（Parser）
- ✅ 解释执行（Interpreter）

未来计划：
- ⏳ 类型断言
- ⏳ 多返回值支持
- ⏳ 命名参数调用
- ⏳ 更多内置函数
- ⏳ 数组和对象支持

## 许可证

Apache License 2.0
