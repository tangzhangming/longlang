# 控制结构

LongLang 提供了完整的控制流语句，包括条件判断和循环结构。

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

简洁的条件表达式：

```longlang
// 单行写法
result := condition ? trueValue : falseValue

// 示例
max := a > b ? a : b
status := age >= 18 ? "成年" : "未成年"
```

多行写法（`?` 和 `:` 必须各自换行）：

```longlang
result := condition
    ? trueValue
    : falseValue
```

> ⚠️ **注意**：三目运算符不能作为函数参数使用

```longlang
// ❌ 禁止
fmt.println(a > b ? a : b)

// ✅ 正确
max := a > b ? a : b
fmt.println(max)
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
| `for` | 循环 | `for cond { ... }` |
| `for` | 传统循环 | `for init; cond; post { ... }` |
| `for` | 无限循环 | `for { ... }` |
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

