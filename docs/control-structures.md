# 控制结构

LongLang 提供了完整的控制流语句，包括条件判断、分支选择和循环结构。

## 条件语句

### if 语句

基本的条件判断：

```longlang
if condition {
    // 条件为真时执行
}
```

### if-else 语句

```longlang
if condition {
    // 条件为真时执行
} else {
    // 条件为假时执行
}
```

### if-else if-else 语句

```longlang
score := 85

if score >= 90 {
    fmt.println("优秀")
} else if score >= 80 {
    fmt.println("良好")
} else if score >= 60 {
    fmt.println("及格")
} else {
    fmt.println("不及格")
}
```

### 嵌套 if 语句

```longlang
age := 25
hasLicense := true

if age >= 18 {
    if hasLicense {
        fmt.println("可以开车")
    } else {
        fmt.println("需要先考驾照")
    }
} else {
    fmt.println("年龄不够")
}
```

## 三目运算符

三目运算符（`?:`）提供简洁的条件表达式。为了提升可读性，LongLang 对三目运算符有特定的书写规范。

### 基本语法

```longlang
result := condition ? trueValue : falseValue

// 示例
max := a > b ? a : b
status := age >= 18 ? "成年" : "未成年"
```

### 允许的写法

#### 1. 单行写法

单行时要求 `?` 和 `:` 都在同一行：

```longlang
value := (x > y) ? true : false
max := a > b ? a : b
```

#### 2. 多行写法

多行时要求 `?` 和 `:` **各自单独换行**（`?` 提行、`:` 也提行）：

```longlang
value := (x > y)
    ? true
    : false

result := condition
    ? trueValue
    : falseValue
```

### 禁止的写法

#### 1. 混写格式（`?` 提行但 `:` 不提行）

下面这种写法会被判定为不可读，因此禁止：

```longlang
// ❌ 禁止
value := (x > y)
    ? true : false
```

#### 2. 作为函数/方法参数

禁止把三目运算符当作参数传入函数/方法/构造：

```longlang
// ❌ 禁止
fmt.println(a > b ? a : b)
get_xxx(age > 20 ? true : false)

// ✅ 正确：先赋值再使用
max := a > b ? a : b
fmt.println(max)

flag := (age > 20) ? true : false
get_xxx(flag)
```

## switch 语句

`switch` 语句用于多分支条件判断，是 `if-else if-else` 链的更清晰替代方案。

### 基本语法

```longlang
x := 2
switch x {
    case 1:
        fmt.println("one")
    case 2:
        fmt.println("two")
    case 3:
        fmt.println("three")
    default:
        fmt.println("other")
}
```

### 带括号形式

```longlang
switch (x) {
    case 1:
        fmt.println("one")
    default:
        fmt.println("other")
}
```

### 多值匹配

一个 `case` 可以匹配多个值：

```longlang
day := 6
switch day {
    case 1, 2, 3, 4, 5:
        fmt.println("工作日")
    case 6, 7:
        fmt.println("周末")
    default:
        fmt.println("无效日期")
}
```

### 条件 switch

省略 switch 表达式时，可以在 case 中使用条件表达式（类似 Go）：

```longlang
score := 85
switch {
    case score >= 90:
        fmt.println("A")
    case score >= 80:
        fmt.println("B")
    case score >= 70:
        fmt.println("C")
    case score >= 60:
        fmt.println("D")
    default:
        fmt.println("F")
}
```

### 重要特性

- **无 fallthrough**：LongLang 的 switch 不支持 fallthrough，每个 case 执行完后自动跳出
- **不支持类型匹配**：当前版本不支持类型匹配，后续版本将支持

## match 表达式

`match` 是一个**表达式**（有返回值），用于值映射，语法灵感来自 Rust 和 PHP 8。

### 基本语法

```longlang
result := match x {
    1 => "one"
    2 => "two"
    3 => "three"
    _ => "other"
}
```

### 带括号形式

```longlang
result := match (x) {
    1 => "one"
    _ => "other"
}
```

### 多值模式

```longlang
result := match statusCode {
    200, 201 => "success"
    400, 401, 403 => "client error"
    500, 502, 503 => "server error"
    _ => "unknown"
}
```

### 守卫条件

使用 `if` 添加守卫条件：

```longlang
grade := match score {
    s if s >= 90 => "A"
    s if s >= 80 => "B"
    s if s >= 70 => "C"
    s if s >= 60 => "D"
    _ => "F"
}
```

在守卫条件中，`s` 绑定了被匹配的值，可以在条件和结果中使用。

### 代码块结果

使用代码块作为结果（需要 return 返回值）：

```longlang
result := match value {
    42 => {
        fmt.println("The answer!")
        return "answer"
    }
    _ => {
        fmt.println("Not the answer")
        return "other"
    }
}
```

### 通配符 `_`

`_` 是通配符，匹配所有未明确匹配的值（类似 default）：

```longlang
result := match x {
    1 => "one"
    2 => "two"
    _ => "unknown"  // 匹配所有其他值
}
```

### 穷尽性检查

> **当前实现**：运行时检查。如果没有匹配到任何分支且没有通配符，会产生运行时错误。
> 
> **未来改进**：编译器实现后，将支持编译时穷尽性检查（针对枚举类型等）。

### switch vs match

| 特性 | switch | match |
|------|--------|-------|
| 类型 | 语句 | 表达式 |
| 返回值 | 无 | 有 |
| 分支符号 | `case:` | `=>` |
| 默认分支 | `default:` | `_` |
| 条件分支 | 支持 | 通过守卫支持 |
| 用途 | 复杂控制流 | 值映射 |

### 实际应用示例

```longlang
// HTTP 状态码处理
fn getStatusText(code: int) string {
    return match code {
        200 => "OK"
        201 => "Created"
        400 => "Bad Request"
        401 => "Unauthorized"
        403 => "Forbidden"
        404 => "Not Found"
        500 => "Internal Server Error"
        _ => "Unknown Status"
    }
}

// 奇偶判断
fn isEven(n: int) bool {
    return match n % 2 {
        0 => true
        _ => false
    }
}

// 星期处理
fn getDayType(day: int) string {
    return match day {
        1, 2, 3, 4, 5 => "工作日"
        6, 7 => "周末"
        _ => "无效日期"
    }
}
```

## for 循环

LongLang 的 for 循环语法与 Go 语言一致，支持三种形式：

### 1. while 式循环

```longlang
i := 0
for i < 5 {
    fmt.println("i =", i)
    i++
}
```

### 2. 无限循环

```longlang
count := 0
for {
    fmt.println("count =", count)
    count++
    if count >= 3 {
        break
    }
}
```

### 3. 传统 for 循环

```longlang
for j := 0; j < 5; j++ {
    fmt.println("j =", j)
}
```

### 4. for-range 循环

遍历集合（Map、Array、String）：

```longlang
// 遍历 Map
myMap := map[string]int{"a": 1, "b": 2, "c": 3}
for k, v := range myMap {
    fmt.println("key:", k, "value:", v)
}

// 只取 key
for k := range myMap {
    fmt.println("key:", k)
}

// 忽略 key，只取 value
for _, v := range myMap {
    fmt.println("value:", v)
}
```

```longlang
// 遍历 Array
arr := []string{"apple", "banana", "cherry"}
for i, item := range arr {
    fmt.println("index:", i, "item:", item)
}

// 忽略 index
for _, item := range arr {
    fmt.println("item:", item)
}
```

```longlang
// 遍历 String（按字符）
str := "Hello"
for i, char := range str {
    fmt.println("index:", i, "char:", char)
}
```

## 循环控制

### break 语句

立即跳出当前循环：

```longlang
for i := 0; i < 10; i++ {
    if i == 5 {
        break  // 跳出循环
    }
    fmt.println(i)
}
// 输出: 0 1 2 3 4
```

### continue 语句

跳过当前迭代，继续下一次：

```longlang
for i := 0; i < 5; i++ {
    if i == 2 {
        continue  // 跳过 i=2
    }
    fmt.println(i)
}
// 输出: 0 1 3 4
```

## 控制结构一览表

| 结构 | 用途 | 语法 |
|------|------|------|
| `if` | 条件判断 | `if cond { ... }` |
| `if-else` | 二选一 | `if cond { ... } else { ... }` |
| `if-else if` | 多条件 | `if cond1 { } else if cond2 { }` |
| `? :` | 三目运算 | `cond ? a : b` |
| `switch` | 多分支语句 | `switch expr { case v: ... }` |
| `match` | 值映射表达式 | `match expr { v => result }` |
| `for` | 循环 | `for cond { ... }` |
| `for` | 传统循环 | `for init; cond; post { ... }` |
| `for` | 无限循环 | `for { ... }` |
| `for-range` | 遍历集合 | `for k, v := range collection { ... }` |
| `break` | 跳出循环 | `break` |
| `continue` | 继续下一次 | `continue` |

## 综合示例

```longlang
fn main() {
    // 查找第一个能被 7 整除的数
    for i := 1; i <= 100; i++ {
        if i % 7 == 0 {
            fmt.println("找到:", i)
            break
        }
    }
    
    // 打印 1-20 中的奇数
    fmt.println("1-20 的奇数:")
    for n := 1; n <= 20; n++ {
        if n % 2 == 0 {
            continue
        }
        fmt.println(n)
    }
    
    // 嵌套循环打印乘法表
    for i := 1; i <= 9; i++ {
        for j := 1; j <= i; j++ {
            fmt.Print(j, "*", i, "=", i*j, " ")
        }
        fmt.println("")
    }
}
```

