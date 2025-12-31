# 运算符

LongLang 提供了丰富的运算符，包括算术运算符、比较运算符、逻辑运算符和赋值运算符。

## 算术运算符

| 运算符 | 名称 | 示例 | 结果 |
|--------|------|------|------|
| `+` | 加法 | `10 + 3` | `13` |
| `-` | 减法 | `10 - 3` | `7` |
| `*` | 乘法 | `10 * 3` | `30` |
| `/` | 除法 | `10 / 3` | `3`（整数除法）|
| `%` | 取模 | `10 % 3` | `1` |
| `++` | 自增 | `i++` | `i = i + 1` |
| `--` | 自减 | `i--` | `i = i - 1` |

### 算术运算示例

```longlang
fn main() {
    a := 10
    b := 3
    
    fmt.println("a + b =", a + b)   // 13
    fmt.println("a - b =", a - b)   // 7
    fmt.println("a * b =", a * b)   // 30
    fmt.println("a / b =", a / b)   // 3
    fmt.println("a % b =", a % b)   // 1
    
    // 自增自减
    a++
    fmt.println("a++ =", a)         // 11
    b--
    fmt.println("b-- =", b)         // 2
}
```

### 浮点数运算

```longlang
fn main() {
    x := 10.5
    y := 3.2
    
    fmt.println("x + y =", x + y)   // 13.7
    fmt.println("x - y =", x - y)   // 7.3
    fmt.println("x * y =", x * y)   // 33.6
    fmt.println("x / y =", x / y)   // 3.28125
}
```

### 混合运算

整数与浮点数混合运算，结果为浮点数：

```longlang
fn main() {
    intVal := 10
    floatVal := 3.5
    
    result := intVal + floatVal    // 13.5 (float)
    fmt.println("结果:", result)
}
```

## 比较运算符

| 运算符 | 名称 | 示例 | 结果 |
|--------|------|------|------|
| `==` | 等于 | `5 == 5` | `true` |
| `!=` | 不等于 | `5 != 3` | `true` |
| `<` | 小于 | `3 < 5` | `true` |
| `>` | 大于 | `5 > 3` | `true` |
| `<=` | 小于等于 | `5 <= 5` | `true` |
| `>=` | 大于等于 | `5 >= 3` | `true` |

### 比较运算示例

```longlang
fn main() {
    a := 10
    b := 20
    
    fmt.println("a == b:", a == b)   // false
    fmt.println("a != b:", a != b)   // true
    fmt.println("a < b:", a < b)     // true
    fmt.println("a > b:", a > b)     // false
    fmt.println("a <= b:", a <= b)   // true
    fmt.println("a >= b:", a >= b)   // false
}
```

### 字符串比较

```longlang
fn main() {
    s1 := "apple"
    s2 := "apple"
    s3 := "banana"
    
    fmt.println("s1 == s2:", s1 == s2)   // true
    fmt.println("s1 == s3:", s1 == s3)   // false
    fmt.println("s1 != s3:", s1 != s3)   // true
}
```

## 逻辑运算符

| 运算符 | 名称 | 示例 | 结果 |
|--------|------|------|------|
| `&&` | 逻辑与 | `true && false` | `false` |
| `\|\|` | 逻辑或 | `true \|\| false` | `true` |
| `!` | 逻辑非 | `!true` | `false` |

### 逻辑运算示例

```longlang
fn main() {
    a := true
    b := false
    
    fmt.println("a && b:", a && b)   // false
    fmt.println("a || b:", a || b)   // true
    fmt.println("!a:", !a)           // false
    fmt.println("!b:", !b)           // true
    
    // 复合条件
    x := 10
    y := 20
    
    if x > 5 && y > 15 {
        fmt.println("两个条件都满足")
    }
    
    if x > 15 || y > 15 {
        fmt.println("至少一个条件满足")
    }
}
```

### 短路求值

逻辑运算符支持短路求值：

```longlang
fn main() {
    // && 短路：如果左边为 false，右边不会执行
    result1 := false && someExpensiveFunction()
    
    // || 短路：如果左边为 true，右边不会执行
    result2 := true || someExpensiveFunction()
}
```

## 赋值运算符

| 运算符 | 名称 | 示例 | 等价于 |
|--------|------|------|--------|
| `=` | 赋值 | `a = 10` | - |
| `:=` | 短变量声明 | `a := 10` | `var a = 10` |

### 赋值运算示例

```longlang
fn main() {
    // 基本赋值
    var a int
    a = 10
    
    // 短变量声明
    b := 20
    
    // 重新赋值
    a = 30
    b = 40
    
    // 表达式赋值
    c := a + b
}
```

## 字符串运算符

| 运算符 | 名称 | 示例 | 结果 |
|--------|------|------|------|
| `+` | 拼接 | `"Hello" + " World"` | `"Hello World"` |

### 字符串拼接示例

```longlang
fn main() {
    firstName := "张"
    lastName := "三"
    
    fullName := firstName + lastName
    fmt.println("全名:", fullName)  // 张三
    
    greeting := "Hello, " + fullName + "!"
    fmt.println(greeting)           // Hello, 张三!
}
```

## 运算符优先级

从高到低排列：

| 优先级 | 运算符 | 说明 |
|--------|--------|------|
| 1 | `()` | 括号 |
| 2 | `!`, `-`(负号), `++`, `--` | 一元运算符 |
| 3 | `*`, `/`, `%` | 乘除取模 |
| 4 | `+`, `-` | 加减 |
| 5 | `<`, `<=`, `>`, `>=` | 比较 |
| 6 | `==`, `!=` | 相等判断 |
| 7 | `&&` | 逻辑与 |
| 8 | `\|\|` | 逻辑或 |
| 9 | `? :` | 三目运算符 |
| 10 | `=`, `:=` | 赋值 |

### 优先级示例

```longlang
fn main() {
    // 乘法优先于加法
    result1 := 2 + 3 * 4      // 14，不是 20
    
    // 使用括号改变优先级
    result2 := (2 + 3) * 4    // 20
    
    // 比较优先于逻辑运算
    result3 := 5 > 3 && 2 < 4 // true
    
    // 逻辑与优先于逻辑或
    result4 := true || false && false  // true (等价于 true || (false && false))
}
```

## 运算符总结表

| 分类 | 运算符 | 用途 |
|------|--------|------|
| 算术 | `+` `-` `*` `/` `%` | 数学计算 |
| 自增减 | `++` `--` | 变量自增减 1 |
| 比较 | `==` `!=` `<` `>` `<=` `>=` | 值比较 |
| 逻辑 | `&&` `\|\|` `!` | 布尔运算 |
| 赋值 | `=` `:=` | 变量赋值 |
| 字符串 | `+` | 字符串拼接 |
| 三目 | `? :` | 条件表达式 |
| 成员访问 | `.` | 访问对象成员 |
| 静态调用 | `::` | 调用静态方法 |

