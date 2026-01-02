# 字符串系统

LongLang 提供了三种使用字符串的方式，满足不同场景的需求，并支持强大的字符串插值语法。

## 字符串插值

字符串插值使用 `$"..."` 语法，可以在字符串中嵌入表达式：

```longlang
name := "World"
greeting := $"Hello, {name}!"  // "Hello, World!"
```

### 支持的表达式

```longlang
// 变量
name := "LongLang"
fmt.println($"Welcome to {name}!")

// 算术表达式
a := 10
b := 20
fmt.println($"Sum: {a + b}")           // "Sum: 30"
fmt.println($"Result: {a * b + 5}")    // "Result: 205"

// 方法调用
text := "hello"
fmt.println($"Upper: {text.upper()}")  // "Upper: HELLO"
fmt.println($"Length: {text.length()}") // "Length: 5"

// 函数调用
fmt.println($"Len: {len(text)}")       // "Len: 5"

// 数组索引
arr := []int{10, 20, 30}
fmt.println($"First: {arr[0]}")        // "First: 10"

// match 表达式（不含字符串字面量时）
status := 200
fmt.println($"Code: {match status { 200 => 1, _ => 0 }}")
```

### 转义

```longlang
// {{ 转义为字面 {
// }} 转义为字面 }
fmt.println($"Syntax: {{expression}}")  // "Syntax: {expression}"

// \{ 和 \} 也可以用于转义
fmt.println($"Use \{braces\}")          // "Use {braces}"
```

### 自动类型转换

插值表达式的结果会自动转换为字符串：

| 类型 | 转换结果 |
|------|---------|
| `int/float` | 数字字符串 |
| `bool` | `"true"` / `"false"` |
| `null` | `"null"` |
| 数组/Map | `Inspect()` 格式 |

### 限制

- **不支持三元表达式**：`$"{a > b ? a : b}"` 不允许，请先计算后插值
- **不支持格式说明符**：如 `{price:F2}`（后续版本可能支持）

```longlang
// ❌ 不允许
result := $"Max: {a > b ? a : b}"

// ✅ 正确做法
max := a > b ? a : b
result := $"Max: {max}"
```

## 字符串类型

### 值类型（原始 string）

原始 `string` 类型是值类型，采用类似 Go 的实现：
- **不可变**：字符串一旦创建就不能修改
- **值语义**：赋值时复制值，而不是引用
- **内存高效**：底层使用指针 + 长度结构

```longlang
name := "hello"
other := name    // other 是 name 的副本
// 修改 other 不会影响 name
```

### 对象类型（System.String）

`System.String` 类封装原始字符串，提供面向对象的接口：
- **引用类型**：赋值时传递引用
- **方法链**：支持链式调用
- **面向对象**：适合需要对象封装的场景

```longlang
use System.String

name := new String("hello")
other := name    // other 和 name 指向同一个对象
```

## 三种使用方式

### 方式 1：语法糖（推荐）

原始字符串可以直接调用方法，语法简洁：

```longlang
name := "hello world"
fmt.println(name.length())          // 11
fmt.println(name.upper())           // HELLO WORLD
fmt.println(name.contains("world")) // true

// 链式调用
result := name.trim().upper().replace("WORLD", "LONGLANG")
fmt.println(result)  // HELLO LONGLANG
```

### 方式 2：静态工具类（System.Str）

使用静态方法操作字符串：

```longlang
use System.Str

fmt.println(Str::length("hello"))               // 5
fmt.println(Str::upper("hello"))                // HELLO
fmt.println(Str::contains("hello", "ell"))      // true
fmt.println(Str::replace("hello", "l", "L"))    // heLlo
```

### 方式 3：对象类（System.String）

使用对象封装字符串：

```longlang
use System.String

name := new String("hello world")
fmt.println(name.getValue())    // hello world
fmt.println(name.length())      // 11

// 链式调用（返回新的 String 对象）
result := name.upper().replace("WORLD", "LONGLANG")
fmt.println(result.getValue())  // HELLO LONGLANG
```

## 方法列表

### 基本信息

| 方法 | 说明 | 示例 | 结果 |
|------|------|------|------|
| `length()` | 获取字符串长度 | `"hello".length()` | `5` |
| `isEmpty()` | 判断是否为空 | `"".isEmpty()` | `true` |
| `charAt(index)` | 获取指定位置字符 | `"hello".charAt(0)` | `"h"` |

### 查找

| 方法 | 说明 | 示例 | 结果 |
|------|------|------|------|
| `indexOf(substr)` | 首次出现索引 | `"hello".indexOf("l")` | `2` |
| `lastIndexOf(substr)` | 最后出现索引 | `"hello".lastIndexOf("l")` | `3` |
| `contains(substr)` | 是否包含子串 | `"hello".contains("ell")` | `true` |
| `startsWith(prefix)` | 是否以前缀开始 | `"hello".startsWith("he")` | `true` |
| `endsWith(suffix)` | 是否以后缀结束 | `"hello".endsWith("lo")` | `true` |

### 比较

| 方法 | 说明 | 示例 | 结果 |
|------|------|------|------|
| `equals(other)` | 比较相等 | `"hello".equals("hello")` | `true` |
| `equalsIgnoreCase(other)` | 忽略大小写比较 | `"Hello".equalsIgnoreCase("hello")` | `true` |

### 连接和子串

| 方法 | 说明 | 示例 | 结果 |
|------|------|------|------|
| `concat(other)` | 连接字符串 | `"hello".concat(" world")` | `"hello world"` |
| `substring(start, end)` | 获取子串 | `"hello".substring(0, 2)` | `"he"` |
| `repeat(count)` | 重复字符串 | `"ab".repeat(3)` | `"ababab"` |

### 去除空白

| 方法 | 说明 | 示例 | 结果 |
|------|------|------|------|
| `trim()` | 去除首尾空白 | `"  hello  ".trim()` | `"hello"` |
| `ltrim()` | 去除左边空白 | `"  hello".ltrim()` | `"hello"` |
| `rtrim()` | 去除右边空白 | `"hello  ".rtrim()` | `"hello"` |

### 大小写转换

| 方法 | 说明 | 示例 | 结果 |
|------|------|------|------|
| `upper()` | 转大写 | `"hello".upper()` | `"HELLO"` |
| `lower()` | 转小写 | `"HELLO".lower()` | `"hello"` |
| `ucfirst()` | 首字母大写 | `"hello".ucfirst()` | `"Hello"` |
| `title()` | 每词首字母大写 | `"hello world".title()` | `"Hello World"` |

### 格式转换

| 方法 | 说明 | 示例 | 结果 |
|------|------|------|------|
| `camel()` | 小驼峰 | `"foo_bar".camel()` | `"fooBar"` |
| `studly()` | 大驼峰 | `"foo_bar".studly()` | `"FooBar"` |
| `snake()` | 蛇形 | `"fooBar".snake()` | `"foo_bar"` |
| `kebab()` | 烤串式 | `"fooBar".kebab()` | `"foo-bar"` |

### 替换

| 方法 | 说明 | 示例 | 结果 |
|------|------|------|------|
| `replace(search, replacement)` | 替换首个 | `"hello".replace("l", "L")` | `"heLlo"` |
| `replaceAll(search, replacement)` | 替换全部 | `"hello".replaceAll("l", "L")` | `"heLLo"` |

### 填充

| 方法 | 说明 | 示例 | 结果 |
|------|------|------|------|
| `padLeft(length, pad)` | 左填充 | `"5".padLeft(3, "0")` | `"005"` |
| `padRight(length, pad)` | 右填充 | `"5".padRight(3, "0")` | `"500"` |

### 其他

| 方法 | 说明 | 示例 | 结果 |
|------|------|------|------|
| `reverse()` | 反转字符串 | `"hello".reverse()` | `"olleh"` |

## 值类型 vs 对象类型

| 特性 | 值类型 (string) | 对象类型 (String) |
|------|-----------------|-------------------|
| 赋值行为 | 复制值 | 复制引用 |
| 内存使用 | 更少 | 更多（有对象开销） |
| 方法返回 | 新的 string | 新的 String 对象 |
| 适用场景 | 大多数情况 | 需要对象封装时 |

```longlang
// 值类型
a := "hello"
b := a           // b 是副本
// a 和 b 是独立的

// 对象类型
use System.String
a := new String("hello")
b := a           // b 和 a 指向同一对象
```

## 最佳实践

1. **优先使用语法糖**：简洁、高效，满足大多数需求
2. **静态方法用于工具函数**：当不需要保存中间状态时
3. **对象类型用于复杂操作**：当需要传递字符串对象或需要显式的面向对象语义时

```longlang
// 推荐：语法糖
result := "hello world".upper().replace("WORLD", "LONGLANG")

// 工具函数场景
use System.Str
if Str::isEmpty(input) {
    fmt.println("输入为空")
}

// 对象场景
use System.String
name := new String(userInput)
// 传递 name 对象到其他函数
processString(name)
```

## 命名空间

字符串相关类位于 `System` 命名空间：

```longlang
use System.String    // 对象类
use System.Str       // 静态工具类
```
