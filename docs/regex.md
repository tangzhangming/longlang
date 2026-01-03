# 正则表达式 (System.Regex)

LongLang 提供完整的正则表达式支持，底层使用 RE2 引擎，保证线性时间复杂度，防止 ReDoS 攻击。

## 命名空间

```longlang
use System.Regex.Regex
use System.Regex.Match
use System.Regex.Group
use System.Regex.RegexOptions
use System.Regex.RegexException
```

## 快速开始

```longlang
use System.Regex.Regex

// 静态方法（一次性使用）
if Regex::isMatch("hello world", "world") {
    fmt.println("找到了!")
}

// 实例方法（重复使用，推荐）
regex := new Regex("\\d+")
matches := regex.findAll("a1b22c333")
for _, m := range matches {
    item := m as Match
    fmt.println(item.getValue())  // 输出: 1, 22, 333
}
```

## Regex 类

### 构造函数

```longlang
regex := new Regex(pattern: string, options: int = 0)
```

### 静态方法

| 方法 | 说明 |
|------|------|
| `Regex::isMatch(input, pattern)` | 快速测试是否匹配 |
| `Regex::replace(input, pattern, replacement)` | 快速替换 |
| `Regex::split(input, pattern)` | 快速分割 |
| `Regex::escape(str)` | 转义特殊字符 |

### 实例方法

| 方法 | 说明 |
|------|------|
| `test(input)` | 测试是否匹配 |
| `find(input)` | 查找第一个匹配 |
| `findAll(input)` | 查找所有匹配 |
| `replaceAll(input, replacement)` | 替换所有匹配 |
| `splitString(input, limit)` | 分割字符串 |
| `findIndex(input)` | 查找匹配位置 |
| `getPattern()` | 获取正则模式 |
| `getOptions()` | 获取选项 |

## Match 类

表示一次匹配的结果。

```longlang
regex := new Regex("(\\w+)@(\\w+)")
result := regex.find("test@example")

if result.isSuccess() {
    fmt.println(result.getValue())   // "test@example"
    fmt.println(result.getIndex())   // 0
    fmt.println(result.group(0))     // "test@example" (完整匹配)
    fmt.println(result.group(1))     // "test" (第一个捕获组)
    fmt.println(result.group(2))     // "example" (第二个捕获组)
}
```

### 方法

| 方法 | 说明 |
|------|------|
| `isSuccess()` | 是否匹配成功 |
| `getValue()` | 获取匹配的字符串 |
| `getIndex()` | 获取匹配的起始位置 |
| `getLength()` | 获取匹配的长度 |
| `group(index)` | 按索引获取捕获组 |
| `groupByName(name)` | 按名称获取命名捕获组 |
| `groupCount()` | 获取捕获组数量 |

## RegexOptions 选项

```longlang
use System.Regex.RegexOptions

// 忽略大小写
regex := new Regex("hello", RegexOptions::IGNORE_CASE)
regex.test("HELLO")  // true

// 组合选项
regex := new Regex("^line", RegexOptions::IGNORE_CASE | RegexOptions::MULTILINE)
```

| 选项 | 值 | 说明 |
|------|-----|------|
| `NONE` | 0 | 默认选项 |
| `IGNORE_CASE` | 1 | 忽略大小写 (?i) |
| `MULTILINE` | 2 | 多行模式 (?m) |
| `SINGLELINE` | 4 | 单行模式 (?s) |
| `UNICODE` | 8 | Unicode 模式 |
| `UNGREEDY` | 16 | 非贪婪模式 (?U) |

## 命名捕获组

```longlang
regex := new Regex("(?P<year>\\d{4})-(?P<month>\\d{2})-(?P<day>\\d{2})")
result := regex.find("2026-01-02")

if result.isSuccess() {
    fmt.println(result.groupByName("year"))   // "2026"
    fmt.println(result.groupByName("month"))  // "01"
    fmt.println(result.groupByName("day"))    // "02"
}
```

## 替换

```longlang
// 简单替换
result := Regex::replace("hello world", "world", "LongLang")
// "hello LongLang"

// 使用捕获组 ($1, $2, ...)
regex := new Regex("(\\w+)@(\\w+)")
result := regex.replaceAll("user@domain", "$1 at $2")
// "user at domain"
```

## 分割

```longlang
// 按正则分割
parts := Regex::split("a1b2c3", "\\d")
// ["a", "b", "c", ""]

// 按多种分隔符分割
regex := new Regex("[,;\\s]+")
parts := regex.splitString("apple, banana; cherry  date")
// ["apple", "banana", "cherry", "date"]
```

## 转义特殊字符

```longlang
escaped := Regex::escape("a.b*c?")
// "a\\.b\\*c\\?"
```

## 支持的正则语法

RE2 引擎支持的语法：

- **字符类**: `[abc]`, `[^abc]`, `[a-z]`, `\d`, `\D`, `\w`, `\W`, `\s`, `\S`
- **量词**: `*`, `+`, `?`, `{n}`, `{n,}`, `{n,m}`
- **锚点**: `^`, `$`, `\b`, `\B`
- **分组**: `(pattern)`, `(?:pattern)`, `(?P<name>pattern)`
- **选择**: `a|b`
- **转义**: `\.`, `\*`, `\+` 等

### 不支持的语法

RE2 为了保证线性时间复杂度，不支持：

- 反向引用 `\1`, `\2`
- 前向断言 `(?=...)`, `(?!...)`
- 后向断言 `(?<=...)`, `(?<!...)`
- 原子分组 `(?>...)`

## 异常处理

```longlang
use System.Regex.RegexException

try {
    regex := new Regex("[invalid")  // 无效的正则
} catch (RegexException e) {
    fmt.println(e.getMessage())
    fmt.println(e.getPattern())
}
```

## 性能说明

- RE2 引擎保证所有正则操作在 O(n) 时间内完成
- 重复使用相同正则时，建议创建实例而非使用静态方法
- 避免在循环中重复编译正则表达式


