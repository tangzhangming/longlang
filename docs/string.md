# 字符串 (String)

LongLang 中的字符串是一个对象，支持丰富的内置方法。

## 字符串定义

```longlang
// 使用双引号
var name string = "hello"

// 使用单引号
var char string = 'world'

// 使用反引号（支持多行）
var text string = `line1
line2`

// 短变量声明
content := "this is a string"
```

## 字符串连接

```longlang
// 使用 + 运算符
result := "hello" + " " + "world"  // "hello world"

// 使用 concat 方法
result := "hello".concat(" world")  // "hello world"

// 字符串 + 其他类型会自动转换
result := "age: " + 25  // "age: 25"
```

---

## 字符串方法

### 基本信息

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `length()` | 获取字符串长度（字符数） | `"hello".length()` | `5` |
| `isEmpty()` | 判断是否为空字符串 | `"".isEmpty()` | `true` |
| `charAt(index)` | 获取指定位置的字符 | `"hello".charAt(0)` | `"h"` |

### 查找方法

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `indexOf(str)` | 返回子串第一次出现的索引，未找到返回 -1 | `"hello".indexOf("l")` | `2` |
| `lastIndexOf(str)` | 返回子串最后一次出现的索引 | `"hello".lastIndexOf("l")` | `3` |
| `contains(str)` | 判断是否包含子串 | `"hello".contains("ell")` | `true` |
| `startsWith(prefix)` | 判断是否以指定前缀开始 | `"hello".startsWith("he")` | `true` |
| `endsWith(suffix)` | 判断是否以指定后缀结束 | `"hello".endsWith("lo")` | `true` |

### 比较方法

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `equals(str)` | 比较两个字符串是否相等 | `"hello".equals("hello")` | `true` |
| `equalsIgnoreCase(str)` | 忽略大小写比较 | `"Hello".equalsIgnoreCase("hello")` | `true` |

### 连接和子串

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `concat(str)` | 连接字符串 | `"hello".concat(" world")` | `"hello world"` |
| `substring(start)` | 从指定位置截取到末尾 | `"hello".substring(2)` | `"llo"` |
| `substring(start, end)` | 截取指定范围的子串 | `"hello".substring(0, 2)` | `"he"` |
| `repeat(n)` | 重复字符串 n 次 | `"ab".repeat(3)` | `"ababab"` |

### 去除空白和字符

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `trim()` | 去除首尾空白 | `"  hello  ".trim()` | `"hello"` |
| `trim(str)` | 去除首尾指定字符串 | `"xxxhelloxxx".trim("xxx")` | `"hello"` |
| `ltrim()` | 去除左边空白 | `"  hello".ltrim()` | `"hello"` |
| `ltrim(str)` | 去除左边指定字符串 | `"http://example.com".ltrim("http://")` | `"example.com"` |
| `rtrim()` | 去除右边空白 | `"hello  ".rtrim()` | `"hello"` |
| `rtrim(str)` | 去除右边指定字符串 | `"example.com/".rtrim("/")` | `"example.com"` |

### 大小写转换

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `upper()` | 转换为大写 | `"hello".upper()` | `"HELLO"` |
| `lower()` | 转换为小写 | `"HELLO".lower()` | `"hello"` |
| `ucfirst()` | 首字母大写 | `"hello".ucfirst()` | `"Hello"` |
| `title()` | 每个单词首字母大写 | `"hello world".title()` | `"Hello World"` |

### 格式转换

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `camel()` | 转为小驼峰（camelCase） | `"foo_bar".camel()` | `"fooBar"` |
| `studly()` | 转为大驼峰（PascalCase） | `"foo_bar".studly()` | `"FooBar"` |
| `snake()` | 转为蛇形（snake_case） | `"fooBar".snake()` | `"foo_bar"` |
| `snake(delimiter)` | 转为蛇形（自定义分隔符） | `"fooBar".snake("-")` | `"foo-bar"` |
| `kebab()` | 转为烤串式（kebab-case） | `"fooBar".kebab()` | `"foo-bar"` |

### 替换方法

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `replace(old, new)` | 替换第一个匹配项 | `"hello".replace("l", "L")` | `"heLlo"` |
| `replaceAll(old, new)` | 替换所有匹配项 | `"hello".replaceAll("l", "L")` | `"heLLo"` |

### 填充方法

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `padLeft(length, pad)` | 左填充到指定长度 | `"5".padLeft(3, "0")` | `"005"` |
| `padRight(length, pad)` | 右填充到指定长度 | `"5".padRight(3, "0")` | `"500"` |

### 其他方法

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `reverse()` | 反转字符串 | `"hello".reverse()` | `"olleh"` |

---

## 完整示例

```longlang
package main

fn main() {
    // 基本信息
    name := "Hello World"
    fmt.println("长度:", name.length())           // 11
    fmt.println("是否为空:", name.isEmpty())       // false
    fmt.println("第一个字符:", name.charAt(0))     // H

    // 查找
    fmt.println("indexOf:", name.indexOf("o"))     // 4
    fmt.println("contains:", name.contains("World")) // true
    fmt.println("startsWith:", name.startsWith("Hello")) // true

    // 大小写
    fmt.println("upper:", name.upper())            // HELLO WORLD
    fmt.println("lower:", name.lower())            // hello world

    // 去除空白
    text := "  trim me  "
    fmt.println("trim:", text.trim())              // "trim me"

    // 格式转换
    varName := "user_name"
    fmt.println("camel:", varName.camel())         // userName
    fmt.println("studly:", varName.studly())       // UserName

    className := "MyClass"
    fmt.println("snake:", className.snake())       // my_class
    fmt.println("kebab:", className.kebab())       // my-class

    // 替换
    fmt.println("replace:", "hello".replaceAll("l", "L"))  // heLLo

    // 填充
    fmt.println("padLeft:", "5".padLeft(3, "0"))   // 005

    // 子串
    fmt.println("substring:", name.substring(0, 5)) // Hello
}
```

---

## 方法速查表

| 分类 | 方法列表 |
|------|----------|
| 基本信息 | `length()`, `isEmpty()`, `charAt()` |
| 查找 | `indexOf()`, `lastIndexOf()`, `contains()`, `startsWith()`, `endsWith()` |
| 比较 | `equals()`, `equalsIgnoreCase()` |
| 连接子串 | `concat()`, `substring()`, `repeat()` |
| 空白处理 | `trim()`, `ltrim()`, `rtrim()` |
| 大小写 | `upper()`, `lower()`, `ucfirst()`, `title()` |
| 格式转换 | `camel()`, `studly()`, `snake()`, `kebab()` |
| 替换 | `replace()`, `replaceAll()` |
| 填充 | `padLeft()`, `padRight()` |
| 其他 | `reverse()` |

---

## 暂未实现

以下方法因需要数组支持，暂未实现：

| 方法 | 说明 |
|------|------|
| `split(delimiter)` | 按分隔符分割字符串为数组 |
| `getBytes()` | 返回字符串的字节数组 |
| `toCharArray()` | 返回字符数组 |

这些方法将在数组功能实现后添加。

