# Map（映射）

LongLang 支持 Map 类型，用于存储键值对数据。Map 的语法与 Go 语言一致。

## 语法概览

| 操作 | 语法 | 说明 |
|------|------|------|
| 类型声明 | `map[KeyType]ValueType` | 声明 Map 类型 |
| 空 Map | `map[string]int{}` | 创建空 Map |
| 带初始值 | `map[string]int{"a": 1, "b": 2}` | 创建带值的 Map |
| 访问值 | `m["key"]` | 获取键对应的值 |
| 设置值 | `m["key"] = value` | 设置或添加键值对 |

## 声明和初始化

### 基本声明

```longlang
// 显式声明类型
var scores map[string]int

// 初始化空 Map
scores := map[string]int{}

// 带初始值
users := map[string]int{"Alice": 100, "Bob": 90, "Charlie": 85}
```

### 多行写法

```longlang
config := map[string]any{
    "host": "localhost",
    "port": 8080,
    "debug": true
}
```

### 支持的值类型

| 值类型 | 示例 |
|--------|------|
| `int` | `map[string]int{"age": 25}` |
| `string` | `map[string]string{"name": "Alice"}` |
| `bool` | `map[string]bool{"active": true}` |
| `float` | `map[string]float{"price": 9.99}` |
| `any` | `map[string]any{"mixed": 123}` |
| 自定义类 | `map[string]User{"admin": user}` |

**注意**: 当前只支持 `string` 类型作为键。

## 访问和修改

### 读取值

```longlang
scores := map[string]int{"Alice": 100, "Bob": 90}

// 直接访问
fmt.println(scores["Alice"])  // 100

// 注意：访问不存在的键会抛出异常
fmt.println(scores["Unknown"])  // 错误: Map 键不存在: Unknown
```

### 安全访问

使用 `isset()` 函数在访问前检查键是否存在：

```longlang
scores := map[string]int{"Alice": 100, "Bob": 90}

if isset(scores, "Alice") {
    fmt.println(scores["Alice"])  // 100
}

if isset(scores, "Unknown") {
    fmt.println(scores["Unknown"])
} else {
    fmt.println("键不存在")
}
```

### 设置和添加值

```longlang
scores := map[string]int{"Alice": 100}

// 修改现有值
scores["Alice"] = 105

// 添加新键值对
scores["Bob"] = 90
scores["Charlie"] = 85
```

## Map 方法

### 基本信息

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `size()` | 获取键值对数量 | `int` |
| `isEmpty()` | 判断是否为空 | `bool` |

```longlang
scores := map[string]int{"Alice": 100, "Bob": 90}

fmt.println(scores.size())      // 2
fmt.println(scores.isEmpty())   // false

empty := map[string]int{}
fmt.println(empty.isEmpty())    // true
```

### 操作方法

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `delete(key)` | 删除键值对 | `bool`（是否成功） |
| `clear()` | 清空 Map | - |

```longlang
scores := map[string]int{"Alice": 100, "Bob": 90, "Charlie": 85}

// 删除键值对
deleted := scores.delete("Bob")
fmt.println(deleted)            // true
fmt.println(scores.size())      // 2

// 删除不存在的键
deleted = scores.delete("Unknown")
fmt.println(deleted)            // false

// 清空 Map
scores.clear()
fmt.println(scores.isEmpty())   // true
```

### 获取集合

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `keys()` | 获取所有键 | `[]string` |
| `values()` | 获取所有值 | `[]ValueType` |

```longlang
scores := map[string]int{"Alice": 100, "Bob": 90, "Charlie": 85}

// 获取所有键
keys := scores.keys()           // {"Alice", "Bob", "Charlie"}
fmt.println(keys.length())      // 3

// 获取所有值
values := scores.values()       // {100, 90, 85}
fmt.println(values.length())    // 3
```

## 遍历 Map

使用 `keys()` 方法遍历 Map：

```longlang
scores := map[string]int{"Alice": 100, "Bob": 90, "Charlie": 85}

// 获取键列表
keys := scores.keys()

// 遍历
for i := 0; i < keys.length(); i++ {
    key := keys[i]
    fmt.println(key + ": " + scores[key])
}

// 输出:
// Alice: 100
// Bob: 90
// Charlie: 85
```

## len 函数

Map 支持 `len()` 全局函数：

```longlang
scores := map[string]int{"Alice": 100, "Bob": 90}

fmt.println(len(scores))  // 2
```

## isset 函数

使用 `isset()` 检查键是否存在：

```longlang
scores := map[string]int{"Alice": 100, "Bob": 90}

fmt.println(isset(scores, "Alice"))    // true
fmt.println(isset(scores, "Unknown"))  // false
```

## 异常处理

访问不存在的键会抛出异常：

```longlang
scores := map[string]int{"Alice": 100}

try {
    x := scores["Unknown"]
} catch (Exception e) {
    fmt.println("错误: " + e.getMessage())
    // 输出: 错误: Map 键不存在: Unknown
}
```

## 完整示例

```longlang
namespace App

class MapDemo {
    public static function main() {
        // 创建 Map
        scores := map[string]int{
            "Alice": 100,
            "Bob": 90,
            "Charlie": 85
        }
        
        // 基本信息
        fmt.println("大小: " + scores.size())        // 3
        fmt.println("是否为空: " + scores.isEmpty())  // false
        
        // 安全访问
        if isset(scores, "Alice") {
            fmt.println("Alice: " + scores["Alice"])  // 100
        }
        
        // 添加/修改
        scores["David"] = 95
        scores["Alice"] = 105
        
        // 删除
        scores.delete("Bob")
        
        // 遍历
        keys := scores.keys()
        for i := 0; i < keys.length(); i++ {
            name := keys[i]
            fmt.println(name + ": " + scores[name])
        }
        
        // 清空
        scores.clear()
        fmt.println("清空后大小: " + scores.size())  // 0
    }
}
```

## 与其他语言对比

| 特性 | LongLang | Go | JavaScript | PHP |
|------|----------|-----|------------|-----|
| 字面量语法 | `map[K]V{...}` | `map[K]V{...}` | `{...}` | `["k" => v]` |
| 键类型 | 仅 string | 多种 | 任意 | 任意 |
| 访问不存在的键 | 抛异常 | 返回零值 | 返回 undefined | 返回 null |
| 检查键存在 | `isset(m, k)` | `_, ok := m[k]` | `"k" in obj` | `isset($m["k"])` |
| 删除键 | `m.delete(k)` | `delete(m, k)` | `delete obj.k` | `unset($m["k"])` |

## 限制

1. **键类型**: 当前只支持 `string` 类型作为键
2. **必须显式写类型**: 不支持类型推导（如 `{"a": 1}` 不合法，必须写 `map[string]int{"a": 1}`）
3. **无序性**: Map 的键顺序为插入顺序，但不保证在所有操作后保持



