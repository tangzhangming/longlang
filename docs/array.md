# 数组

LongLang 支持两种数组类型：**固定长度数组**和**动态数组（切片）**。

## 语法概览

| 类型 | 语法 | 说明 |
|------|------|------|
| 固定长度数组 | `[size]type{...}` | 长度固定，不可改变 |
| 长度推导数组 | `[...]type{...}` | 根据元素自动推导长度 |
| 动态数组（切片） | `[]type{...}` | 长度可变 |
| 类型推导数组 | `{...}` | 根据元素推导类型，创建切片 |

## 声明方式

### 1. 固定长度数组

```longlang
// var 声明
var numbers [5]int = {1, 2, 3, 4, 5}

// 短变量声明
nums := [5]int{10, 20, 30, 40, 50}
```

固定长度数组的元素个数必须与声明的长度一致，否则会报错。

### 2. 长度推导数组

```longlang
// var 声明 - 根据元素个数自动推导长度
var prices [...]float = {9.99, 19.99, 29.99}

// 短变量声明
codes := [...]int{100, 200, 300, 400}
```

### 3. 动态数组（切片）

```longlang
// var 声明
var ids []int = {1, 2, 3, 4, 5}

// 短变量声明
values := []float{1.1, 2.2, 3.3}
```

### 4. 类型推导数组

当不指定类型时，根据第一个元素的类型推导整个数组的类型：

```longlang
// var 声明
var names = {"xiaohong", "Bob", "Alice"}  // 推导为 []string

// 短变量声明
users := {"Alice", "Bob"}  // 推导为 []string

// 包含变量
var username = "David"
friends := {username, "Eva", "Frank"}  // 推导为 []string
```

**注意**：类型推导要求所有元素类型一致，否则会报错。

## 数组操作

### 索引访问

```longlang
var arr = {10, 20, 30, 40, 50}
fmt.println(arr[0])   // 输出: 10
fmt.println(arr[4])   // 输出: 50
fmt.println(arr[-1])  // 输出: 50 (负数索引从末尾开始)
fmt.println(arr[-2])  // 输出: 40
```

### 索引赋值

```longlang
var arr = {10, 20, 30}
arr[1] = 200
fmt.println(arr)  // 输出: {10, 200, 30}
```

### 获取长度

使用内置函数 `len()` 或方法 `length()` 获取数组长度：

```longlang
var arr = {1, 2, 3, 4, 5}
fmt.println(len(arr))       // 输出: 5
fmt.println(arr.length())   // 输出: 5

var str = "Hello"
fmt.println(len(str))  // 输出: 5 (字符串也支持 len)
```

## 数组方法

### 基本信息

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `length()` | 获取数组长度 | `arr.length()` | `int` |
| `isEmpty()` | 判断是否为空 | `arr.isEmpty()` | `bool` |

```longlang
arr := []int{1, 2, 3}
fmt.println(arr.length())   // 3
fmt.println(arr.isEmpty())  // false

empty := []int{}
fmt.println(empty.isEmpty())  // true
```

### 添加和删除

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `push(value)` | 在末尾添加元素 | `arr.push(5)` | - |
| `pop()` | 删除并返回最后一个元素 | `arr.pop()` | 元素值 |
| `shift()` | 删除并返回第一个元素 | `arr.shift()` | 元素值 |
| `clear()` | 清空数组 | `arr.clear()` | - |

```longlang
arr := []int{1, 2, 3}

arr.push(4)           // arr = {1, 2, 3, 4}
last := arr.pop()     // last = 4, arr = {1, 2, 3}
first := arr.shift()  // first = 1, arr = {2, 3}

arr.clear()           // arr = {}
```

### 查找

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `contains(value)` | 判断是否包含元素 | `arr.contains(2)` | `bool` |
| `indexOf(value)` | 返回元素第一次出现的索引 | `arr.indexOf(2)` | `int`（不存在返回 -1） |

```longlang
arr := []int{10, 20, 30, 20}

fmt.println(arr.contains(20))   // true
fmt.println(arr.contains(99))   // false
fmt.println(arr.indexOf(20))    // 1
fmt.println(arr.indexOf(99))    // -1
```

### 转换

| 方法 | 说明 | 示例 | 返回值 |
|------|------|------|--------|
| `join(separator)` | 用分隔符连接元素 | `arr.join(",")` | `string` |
| `reverse()` | 反转数组（返回新数组） | `arr.reverse()` | 新数组 |
| `slice(start, end)` | 截取数组片段 | `arr.slice(1, 3)` | 新数组 |

```longlang
arr := []int{1, 2, 3, 4, 5}

// join - 连接成字符串
fmt.println(arr.join(","))    // "1,2,3,4,5"
fmt.println(arr.join(" - "))  // "1 - 2 - 3 - 4 - 5"

// reverse - 反转
reversed := arr.reverse()     // {5, 4, 3, 2, 1}

// slice - 截取
part1 := arr.slice(1)         // {2, 3, 4, 5} (从索引1到末尾)
part2 := arr.slice(1, 3)      // {2, 3} (从索引1到3，不包含3)
```

## isset 函数

使用全局函数 `isset()` 检查数组索引是否有效：

```longlang
arr := []int{10, 20, 30, 40, 50}

fmt.println(isset(arr, 0))    // true
fmt.println(isset(arr, 4))    // true
fmt.println(isset(arr, 5))    // false (越界)
fmt.println(isset(arr, -1))   // true (负数索引有效)
fmt.println(isset(arr, -5))   // true
fmt.println(isset(arr, -6))   // false (越界)

// 安全访问
if isset(arr, 10) {
    fmt.println(arr[10])
} else {
    fmt.println("索引不存在")
}
```

## 多维数组

```longlang
// 二维数组
var matrix [2][3]int = {{1, 2, 3}, {4, 5, 6}}
fmt.println(matrix[0][0])  // 输出: 1
fmt.println(matrix[1][2])  // 输出: 6

// 修改元素
matrix[0][1] = 100
fmt.println(matrix[0])  // 输出: {1, 100, 3}

// 类型推导的嵌套数组
nested := {{10, 20}, {30, 40}}
fmt.println(nested[1][0])  // 输出: 30
```

## 字符串索引

字符串也支持索引访问：

```longlang
var str = "Hello"
fmt.println(str[0])   // 输出: H
fmt.println(str[-1])  // 输出: o (负数索引)
```

## 支持的元素类型

| 类型 | 示例 |
|------|------|
| `int` | `[3]int{1, 2, 3}` |
| `i8`, `i16`, `i32`, `i64` | `[2]i64{100, 200}` |
| `uint` | `[3]uint{1, 2, 3}` |
| `u8`, `u16`, `u32`, `u64` | `[2]u32{100, 200}` |
| `float` | `[3]float{1.1, 2.2, 3.3}` |
| `f32`, `f64` | `[2]f64{3.14, 2.71}` |
| `string` | `[2]string{"a", "b"}` |
| `bool` | `[2]bool{true, false}` |
| `any` | `[3]any{1, "two", true}` |
| 嵌套数组 | `[2][3]int{{1,2,3}, {4,5,6}}` |

## 错误处理

### 索引越界

```longlang
var arr = {1, 2, 3}
fmt.println(arr[10])  // 错误: 数组索引越界：索引 10 超出范围 [0, 2]
```

### 类型不一致

```longlang
var arr = {1, "two", 3}  // 错误: 数组元素类型不一致
```

### 长度不匹配

```longlang
var arr [3]int = {1, 2, 3, 4, 5}  // 错误: 数组长度不匹配：期望 3 个元素，得到 5 个
```

## 与其他语言对比

| 特性 | LongLang | Go | JavaScript |
|------|----------|-----|------------|
| 字面量语法 | `{1, 2, 3}` | `[]int{1, 2, 3}` | `[1, 2, 3]` |
| 固定长度 | `[5]int{...}` | `[5]int{...}` | 不支持 |
| 切片 | `[]int{...}` | `[]int{...}` | `[...]` |
| 负数索引 | ✅ 支持 | ❌ 不支持 | ❌ 不支持 |
| 类型推导 | ✅ 支持 | 部分支持 | ✅ 支持 |

