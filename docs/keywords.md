# 关键字

关键字是 LongLang 语言保留的标识符，具有特殊含义，不能用作变量名、函数名等。

## 关键字一览表

| 关键字 | 用途 | 示例 |
|--------|------|------|
| `fn` | 函数定义 | `fn add(a:int) int { }` |
| `function` | 函数定义（类方法中使用） | `public function getName() { }` |
| `var` | 变量声明 | `var x int = 10` |
| `if` | 条件判断 | `if condition { }` |
| `else` | 条件分支 | `else { }` |
| `for` | 循环 | `for i < 10 { }` |
| `break` | 跳出循环 | `break` |
| `continue` | 继续下一次循环 | `continue` |
| `return` | 返回值 | `return value` |
| `package` | 包声明 | `package main` |
| `import` | 导入包 | `import "fmt"` |
| `class` | 类定义 | `class Person { }` |
| `extends` | 类继承 | `class Dog extends Animal { }` |
| `public` | 公开访问修饰符 | `public name string` |
| `private` | 私有访问修饰符 | `private age int` |
| `protected` | 受保护访问修饰符 | `protected data string` |
| `static` | 静态成员 | `public static function create() { }` |
| `this` | 当前对象引用 | `this.name` |
| `super` | 父类引用 | `super::method()` |
| `new` | 创建对象实例 | `new Person("Alice")` |
| `true` | 布尔真值 | `isActive := true` |
| `false` | 布尔假值 | `isActive := false` |
| `null` | 空值 | `x := null` |
| `any` | 任意类型 | `fn print(x:any) { }` |
| `void` | 无返回值类型 | `fn doWork() void { }` |

## 类型关键字

| 关键字 | 用途 | 字节数 |
|--------|------|--------|
| `int` | 有符号整型 | 8 |
| `i8` | 8位有符号整型 | 1 |
| `i16` | 16位有符号整型 | 2 |
| `i32` | 32位有符号整型 | 4 |
| `i64` | 64位有符号整型 | 8 |
| `uint` | 无符号整型 | 8 |
| `u8` | 8位无符号整型 | 1 |
| `u16` | 16位无符号整型 | 2 |
| `u32` | 32位无符号整型 | 4 |
| `u64` | 64位无符号整型 | 8 |
| `float` | 浮点数 | 8 |
| `f32` | 32位浮点数 | 4 |
| `f64` | 64位浮点数 | 8 |
| `bool` | 布尔类型 | 1 |
| `string` | 字符串类型 | 可变 |

## 关键字详解

### 函数定义关键字

#### fn

用于定义普通函数：

```longlang
fn add(a:int, b:int) int {
    return a + b
}

fn main() {
    result := add(1, 2)
}
```

#### function

用于定义类的方法：

```longlang
class Calculator {
    public function add(a:int, b:int) int {
        return a + b
    }
}
```

### 变量声明关键字

#### var

显式声明变量：

```longlang
var name string = "Alice"
var age int = 25
var score = 95  // 类型推导
```

### 控制流关键字

#### if / else

条件判断：

```longlang
if score >= 60 {
    fmt.Println("及格")
} else {
    fmt.Println("不及格")
}
```

#### for / break / continue

循环控制：

```longlang
for i := 0; i < 10; i++ {
    if i == 5 {
        continue  // 跳过 5
    }
    if i == 8 {
        break     // 在 8 处退出
    }
    fmt.Println(i)
}
```

#### return

返回函数结果：

```longlang
fn max(a:int, b:int) int {
    if a > b {
        return a
    }
    return b
}
```

### 包管理关键字

#### package / import

```longlang
package main

import "fmt"

fn main() {
    fmt.Println("Hello!")
}
```

### 面向对象关键字

#### class / public / private / protected / static / this / new

```longlang
class Person {
    public name string
    private age int
    
    public function __construct(name:string, age:int) {
        this.name = name
        this.age = age
    }
    
    public static function create(name:string) Person {
        return new Person(name, 0)
    }
}

fn main() {
    p1 := new Person("Alice", 25)
    p2 := Person::create("Bob")
}
```

### 值关键字

#### true / false / null

```longlang
isActive := true
isDeleted := false
data := null
```

### 类型关键字

#### any / void

```longlang
// any 可以接受任意类型
fn printValue(x:any) {
    fmt.Println(x)
}

// void 表示无返回值
fn doSomething() void {
    fmt.Println("执行操作")
}
```

## 关键字分类总结

| 分类 | 关键字 |
|------|--------|
| 函数 | `fn`, `function`, `return` |
| 变量 | `var` |
| 控制流 | `if`, `else`, `for`, `break`, `continue` |
| 包管理 | `package`, `import` |
| 面向对象 | `class`, `extends`, `public`, `private`, `protected`, `static`, `this`, `super`, `new` |
| 值 | `true`, `false`, `null` |
| 类型 | `int`, `i8`, `i16`, `i32`, `i64`, `uint`, `u8`, `u16`, `u32`, `u64`, `float`, `f32`, `f64`, `bool`, `string`, `any`, `void` |

## 注意事项

1. **关键字不能作为标识符**：不能用关键字命名变量、函数、类等
2. **区分大小写**：`if` 是关键字，但 `If` 或 `IF` 不是
3. **类型关键字**：类型关键字用于声明变量和参数的类型

```longlang
// ❌ 错误：不能使用关键字作为变量名
var if = 10
var class = "test"

// ✅ 正确：使用合法的标识符
var condition = 10
var className = "test"
```

