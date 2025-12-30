# 函数

函数是 LongLang 中组织代码的基本单元，支持参数、返回值、默认参数等特性。

## 函数定义

### 基本语法

```longlang
fn 函数名(参数列表) 返回类型 {
    // 函数体
}
```

### 无返回值函数

```longlang
fn greet(name:string) {
    fmt.Println("Hello,", name)
}

fn sayHello() {
    fmt.Println("Hello, World!")
}
```

### 有返回值函数

```longlang
fn add(a:int, b:int) int {
    return a + b
}

fn multiply(x:float, y:float) float {
    return x * y
}

fn isPositive(n:int) bool {
    return n > 0
}
```

### 多返回值函数

```longlang
fn divide(a:int, b:int) (int, int) {
    quotient := a / b
    remainder := a % b
    return quotient, remainder
}
```

## 函数参数

### 必需参数

```longlang
fn greet(name:string, age:int) {
    fmt.Println(name, "今年", age, "岁")
}

// 调用时必须提供所有参数
greet("Alice", 25)
```

### 默认参数

```longlang
fn greet(name:string = "World") {
    fmt.Println("Hello,", name)
}

// 调用方式
greet()          // 使用默认值，输出: Hello, World
greet("Alice")   // 传入参数，输出: Hello, Alice
```

### 命名参数调用

```longlang
fn createUser(name:string, age:int = 0, city:string = "Beijing") {
    fmt.Println("姓名:", name, "年龄:", age, "城市:", city)
}

// 使用命名参数
createUser(name:"Alice", age:25)
createUser(name:"Bob", city:"Shanghai")
```

## return 语句

### 返回单个值

```longlang
fn square(n:int) int {
    return n * n
}
```

### 无返回值的 return

用于提前结束函数：

```longlang
fn processData(data:string) {
    if data == "" {
        fmt.Println("数据为空")
        return  // 提前退出
    }
    
    fmt.Println("处理数据:", data)
}
```

### 条件返回

```longlang
fn abs(x:int) int {
    if x < 0 {
        return -x
    }
    return x
}

fn max(a:int, b:int) int {
    if a > b {
        return a
    }
    return b
}
```

## 函数调用

```longlang
fn main() {
    // 基本调用
    greet("Alice")
    
    // 获取返回值
    result := add(10, 20)
    fmt.Println("结果:", result)
    
    // 直接在表达式中使用
    fmt.Println("5 + 3 =", add(5, 3))
    
    // 嵌套调用
    fmt.Println("max:", max(add(1, 2), add(3, 4)))
}
```

## 程序入口函数

每个 LongLang 程序必须有一个 `main` 函数作为入口点：

```longlang
fn main() {
    // 程序从这里开始执行
    fmt.Println("程序开始")
}
```

## 内置函数

LongLang 提供了一些内置函数：

| 函数 | 说明 | 示例 |
|------|------|------|
| `fmt.Println` | 打印并换行 | `fmt.Println("Hello")` |
| `fmt.Print` | 打印不换行 | `fmt.Print("Hello")` |
| `fmt.Printf` | 格式化打印 | `fmt.Printf("数字: %d", 42)` |

### 格式化占位符

| 占位符 | 说明 | 示例 |
|--------|------|------|
| `%d` | 整数 | `fmt.Printf("%d", 42)` |
| `%f` | 浮点数 | `fmt.Printf("%f", 3.14)` |
| `%s` | 字符串 | `fmt.Printf("%s", "hello")` |
| `%t` | 布尔值 | `fmt.Printf("%t", true)` |
| `%v` | 任意值 | `fmt.Printf("%v", value)` |

## 函数类型一览

| 类型 | 语法 | 示例 |
|------|------|------|
| 无参数无返回 | `fn name() { }` | `fn hello() { }` |
| 有参数无返回 | `fn name(p:type) { }` | `fn greet(s:string) { }` |
| 有返回值 | `fn name() type { }` | `fn rand() int { }` |
| 完整形式 | `fn name(p:type) type { }` | `fn add(a:int, b:int) int { }` |
| 多返回值 | `fn name() (t1, t2) { }` | `fn div(a, b:int) (int, int) { }` |
| 默认参数 | `fn name(p:type = val) { }` | `fn greet(s:string = "Hi") { }` |

## 匿名函数（闭包）

LongLang 支持匿名函数，可以将函数作为值赋给变量：

### 基本语法

```longlang
// 无参数无返回值
sayHello := fn() {
    fmt.Println("Hello!")
}
sayHello()

// 带参数和返回值
add := fn(a:int, b:int) int {
    return a + b
}
result := add(3, 5)  // 8
```

### 闭包捕获变量

匿名函数可以捕获外部作用域的变量：

```longlang
fn main() {
    x := 10
    addX := fn(n:int) int {
        return n + x  // 捕获外部变量 x
    }
    fmt.Println(addX(5))  // 15
}
```

### 立即调用的匿名函数 (IIFE)

```longlang
result := fn(a:int, b:int) int {
    return a * b
}(3, 4)  // 立即调用，result = 12
```

### 作为参数传递

```longlang
fn apply(f:any, x:int) int {
    return f(x)
}

double := fn(n:int) int {
    return n * 2
}

result := apply(double, 5)  // 10
```

## 综合示例

```longlang
// 计算阶乘
fn factorial(n:int) int {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}

// 斐波那契数列
fn fibonacci(n:int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

// 检查素数
fn isPrime(n:int) bool {
    if n < 2 {
        return false
    }
    for i := 2; i * i <= n; i++ {
        if n % i == 0 {
            return false
        }
    }
    return true
}

fn main() {
    fmt.Println("5! =", factorial(5))
    fmt.Println("fib(10) =", fibonacci(10))
    fmt.Println("17 是素数:", isPrime(17))
}
```

