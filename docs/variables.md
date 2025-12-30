# 变量

LongLang 提供了灵活的变量声明方式，支持类型推导和显式类型声明。

## 变量声明方式

### 1. var 关键字声明

使用 `var` 关键字声明变量，可以指定类型或让编译器自动推导：

```longlang
// 带类型和初始值
var name string = "xiaoming"
var age int = 18
var pi float = 3.14159

// 自动类型推导（省略类型）
var score = 95
var message = "hello"

// 只声明不初始化（值为 null）
var uninitialized int
```

### 2. 短变量声明 :=

使用 `:=` 进行短变量声明，只能在函数内部使用：

```longlang
fn main() {
    // 短变量声明，自动推导类型
    count := 10
    name := "Alice"
    price := 99.9
    isActive := true
}
```

## 变量命名规则

| 规则 | 说明 | 示例 |
|------|------|------|
| 首字符 | 必须是字母或下划线 | `name`, `_count` |
| 后续字符 | 字母、数字或下划线 | `user123`, `max_value` |
| 大小写敏感 | `Name` 和 `name` 是不同变量 | - |
| 不能是关键字 | 不能使用保留关键字 | ❌ `var`, `if`, `fn` |

## 变量作用域

```longlang
var globalVar = "我是全局变量"

fn main() {
    var localVar = "我是局部变量"
    
    if true {
        var blockVar = "我是块级变量"
        fmt.Println(blockVar)    // ✅ 可访问
        fmt.Println(localVar)    // ✅ 可访问
        fmt.Println(globalVar)   // ✅ 可访问
    }
    
    // fmt.Println(blockVar)     // ❌ 不可访问，已超出作用域
}
```

## 变量赋值

```longlang
fn main() {
    // 声明后赋值
    var x int
    x = 10
    
    // 重新赋值
    x = 20
    
    // 表达式赋值
    x = x + 5
    
    // 自增自减
    x++
    x--
}
```

## 零值和 null

未初始化的变量值为 `null`：

```longlang
var x int       // x 的值是 null
var s string    // s 的值是 null

// 检查 null
if x == null {
    fmt.Println("x 未初始化")
}
```

## 常见用法示例

```longlang
fn main() {
    // 计数器
    counter := 0
    
    // 累加
    for counter < 10 {
        counter++
    }
    
    // 字符串拼接
    firstName := "张"
    lastName := "三"
    fullName := firstName + lastName
    
    // 条件赋值
    score := 85
    grade := score >= 60 ? "及格" : "不及格"
    
    fmt.Println("姓名:", fullName)
    fmt.Println("成绩:", grade)
}
```

