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
    fmt.println("Hello,", name)
}

fn sayHello() {
    fmt.println("Hello, World!")
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
    fmt.println(name, "今年", age, "岁")
}

// 调用时必须提供所有参数
greet("Alice", 25)
```

### 默认参数

```longlang
fn greet(name:string = "World") {
    fmt.println("Hello,", name)
}

// 调用方式
greet()          // 使用默认值，输出: Hello, World
greet("Alice")   // 传入参数，输出: Hello, Alice
```

### 命名参数调用

```longlang
fn createUser(name:string, age:int = 0, city:string = "Beijing") {
    fmt.println("姓名:", name, "年龄:", age, "城市:", city)
}

// 使用命名参数
createUser(name:"Alice", age:25)
createUser(name:"Bob", city:"Shanghai")
```

### 可变参数

可变参数允许函数接收任意数量的参数。使用 `...` 前缀声明可变参数，可变参数在函数内部表现为数组。

#### 基本语法

```longlang
fn sum(...numbers:int) int {
    total := 0
    for i := 0; i < numbers.length(); i++ {
        total = total + numbers[i]
    }
    return total
}

// 调用方式
sum(1, 2, 3)        // 返回 6
sum(1, 2, 3, 4, 5)  // 返回 15
sum()               // 返回 0
```

#### 规则

1. **可变参数必须是最后一个参数**
2. **函数只能有一个可变参数**
3. **可变参数不能有默认值**

#### 结合固定参数

```longlang
fn printf(format:string, ...args:any) {
    // format 是固定参数
    // args 是可变参数，收集其余所有参数
    for i := 0; i < args.length(); i++ {
        fmt.println("[" + toString(i) + "] " + toString(args[i]))
    }
}

printf("Hello", "World", 123, true)
// 输出:
// [0] World
// [1] 123
// [2] true
```

#### 结合默认参数

可变参数可以与默认参数一起使用，但可变参数必须在最后：

```longlang
fn greet(name:string, greeting:string = "Hi", ...extras:string) {
    msg := greeting + ", " + name
    for i := 0; i < extras.length(); i++ {
        msg = msg + " " + extras[i]
    }
    fmt.println(msg)
}

greet("World", "Hello", "from", "LongLang")  // Hello, World from LongLang
greet("User")                                  // Hi, User
```

#### 类方法中的可变参数

可变参数同样适用于类的实例方法和静态方法：

```longlang
class Logger {
    private prefix string

    public function __construct(prefix:string) {
        this.prefix = prefix
    }

    // 实例方法可变参数
    public function log(...messages:string) {
        for i := 0; i < messages.length(); i++ {
            fmt.println("[" + this.prefix + "] " + messages[i])
        }
    }
}

class MathUtil {
    // 静态方法可变参数
    public static function max(...nums:int) int {
        if nums.length() == 0 {
            return 0
        }
        maxVal := nums[0]
        for i := 1; i < nums.length(); i++ {
            if nums[i] > maxVal {
                maxVal = nums[i]
            }
        }
        return maxVal
    }
}

// 使用
logger := new Logger("App")
logger.log("Message1", "Message2", "Message3")

maxValue := MathUtil::max(5, 2, 8, 1, 9)  // 返回 9
```

#### 闭包中的可变参数

闭包（匿名函数）同样支持可变参数：

```longlang
multiply := fn(...nums:int) int {
    result := 1
    for i := 0; i < nums.length(); i++ {
        result = result * nums[i]
    }
    return result
}

multiply(2, 3, 4)  // 返回 24
multiply()         // 返回 1
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
        fmt.println("数据为空")
        return  // 提前退出
    }
    
    fmt.println("处理数据:", data)
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
    fmt.println("结果:", result)
    
    // 直接在表达式中使用
    fmt.println("5 + 3 =", add(5, 3))
    
    // 嵌套调用
    fmt.println("max:", max(add(1, 2), add(3, 4)))
}
```

## 程序入口函数

每个 LongLang 程序必须有一个 `main` 函数作为入口点：

```longlang
fn main() {
    // 程序从这里开始执行
    fmt.println("程序开始")
}
```

## 内置函数

LongLang 提供了一些内置函数：

| 函数 | 说明 | 示例 |
|------|------|------|
| `fmt.println` | 打印并换行 | `fmt.println("Hello")` |
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
| 可变参数 | `fn name(...args:type) { }` | `fn sum(...nums:int) int { }` |

## 匿名函数（闭包）

LongLang 支持匿名函数，可以将函数作为值赋给变量：

### 基本语法

```longlang
// 无参数无返回值
sayHello := fn() {
    fmt.println("Hello!")
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
    fmt.println(addX(5))  // 15
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
    fmt.println("5! =", factorial(5))
    fmt.println("fib(10) =", fibonacci(10))
    fmt.println("17 是素数:", isPrime(17))
}
```

