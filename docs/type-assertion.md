# 类型断言

LongLang 提供类型断言机制，用于将 `any` 类型的值转换为具体类型。

## 语法

| 语法 | 说明 | 失败行为 |
|------|------|----------|
| `value as Type` | 强制类型断言 | 抛出 `TypeError` |
| `value as? Type` | 安全类型断言 | 返回 `null` |

## 强制断言 `as`

强制断言在类型不匹配时会抛出异常：

```longlang
// 基本类型
var data any = "Hello, World!"
str := data as string
fmt.println(str.upper())  // HELLO, WORLD!

var num any = 42
n := num as int
fmt.println(n * 2)  // 84

var pi any = 3.14159
f := pi as float
fmt.println(f)  // 3.14159

var flag any = true
b := flag as bool
fmt.println(b)  // true
```

### 类实例断言

```longlang
var obj any = new User("Alice", 25)

// 断言为具体类型
user := obj as User
fmt.println(user.getName())  // Alice

// 断言为父类类型
var adminObj any = new Admin("Bob", 30)
adminAsUser := adminObj as User  // Admin 继承自 User，可以断言
fmt.println(adminAsUser.getName())  // Bob
```

### 数组类型断言

```longlang
var arrAny any = []int{1, 2, 3, 4, 5}

intArr := arrAny as []int
fmt.println(intArr.length())  // 5
fmt.println(intArr[0])        // 1
```

### 失败时抛出异常

```longlang
var value any = "not a number"

try {
    x := value as int  // 类型不匹配，抛出异常
    fmt.println(x)
} catch (Exception e) {
    fmt.println(e.getMessage())  
    // 输出: 类型断言失败: 无法将 string 转换为 int
}
```

## 安全断言 `as?`

安全断言在类型不匹配时返回 `null`，不会抛出异常：

```longlang
var value any = 123

// 尝试转为字符串（会失败）
str := value as? string
if str == null {
    fmt.println("不是字符串")  // 输出这个
}

// 尝试转为整数（成功）
num := value as? int
if num != null {
    fmt.println($"整数: {num}")  // 整数: 123
}
```

### null 值处理

```longlang
var nullValue any = null

// 安全断言返回 null
result := nullValue as? string
if result == null {
    fmt.println("值为 null 或类型不匹配")
}

// 强制断言抛出异常
try {
    str := nullValue as string
} catch (Exception e) {
    fmt.println("null 无法转换为 string")
}
```

## 处理异构集合

类型断言在处理包含不同类型元素的集合时非常有用：

```longlang
items := []any{"hello", 123, 3.14, true, new User("Alice", 25)}

for idx, item := range items {
    // 尝试不同类型
    itemStr := item as? string
    itemInt := item as? int
    itemFloat := item as? float
    itemBool := item as? bool
    itemUser := item as? User
    
    if itemStr != null {
        fmt.println($"[{idx}] 字符串: {itemStr}")
    } else if itemInt != null {
        fmt.println($"[{idx}] 整数: {itemInt}")
    } else if itemFloat != null {
        fmt.println($"[{idx}] 浮点数: {itemFloat}")
    } else if itemBool != null {
        fmt.println($"[{idx}] 布尔: {itemBool}")
    } else if itemUser != null {
        fmt.println($"[{idx}] 用户: {itemUser.getName()}")
    } else {
        fmt.println($"[{idx}] 未知类型")
    }
}

// 输出:
// [0] 字符串: hello
// [1] 整数: 123
// [2] 浮点数: 3.14
// [3] 布尔: true
// [4] 用户: Alice
```

## 支持的类型

### 基本类型

| 类型 | 说明 |
|------|------|
| `int`, `i8`, `i16`, `i32`, `i64` | 有符号整数 |
| `uint`, `u8`, `u16`, `u32`, `u64` | 无符号整数 |
| `byte` | `u8` 的别名 |
| `float`, `f32`, `f64` | 浮点数 |
| `string` | 字符串 |
| `bool` | 布尔值 |
| `any` | 任意类型（始终成功） |

### 复合类型

| 类型 | 说明 | 示例 |
|------|------|------|
| 数组 | 切片或固定数组 | `[]int`, `[]string`, `[]User` |
| Map | 键值映射 | `map[string]int`, `map[string]User` |
| 类 | 类实例 | `User`, `Admin` |
| 接口 | 接口实例 | `Readable`, `Writable` |

## 继承关系

类型断言支持继承关系：

```longlang
class Animal {
    public function speak() string { return "..." }
}

class Dog extends Animal {
    public function speak() string { return "Woof!" }
    public function fetch() { fmt.println("Fetching...") }
}

var obj any = new Dog()

// Dog 可以断言为 Animal（父类）
animal := obj as Animal
fmt.println(animal.speak())  // Woof!

// Dog 也可以断言为 Dog（自身类型）
dog := obj as Dog
dog.fetch()  // Fetching...

// Animal 实例不能断言为 Dog
var animalObj any = new Animal()
dogFromAnimal := animalObj as? Dog
if dogFromAnimal == null {
    fmt.println("Animal 不能转换为 Dog")
}
```

## 运算符优先级

`as` / `as?` 的优先级低于成员访问（`.`），高于比较运算符：

```longlang
// 需要括号访问成员
name := (obj as User).getName()

// 在比较表达式中
if value as? int != null {
    // ...
}

// 在表达式中
length := (arr as []int).length()
```

## 最佳实践

1. **优先使用安全断言**：除非确定类型一定匹配，否则使用 `as?` 避免异常。

2. **检查 null**：使用 `as?` 后始终检查结果是否为 `null`。

3. **配合类型检查**：在不确定类型时，先用 `as?` 检查再使用。

4. **异常处理**：对强制断言使用 `try-catch` 处理可能的失败。

```longlang
// 推荐：安全断言 + null 检查
user := obj as? User
if user != null {
    fmt.println(user.getName())
}

// 或：强制断言 + 异常处理
try {
    user := obj as User
    fmt.println(user.getName())
} catch (Exception e) {
    fmt.println("类型转换失败")
}
```

## 与其他语言对比

| 语言 | 强制断言 | 安全断言 |
|------|----------|----------|
| LongLang | `value as Type` | `value as? Type` |
| Go | `value.(Type)` | `v, ok := value.(Type)` |
| TypeScript | `value as Type` | - |
| Kotlin | `value as Type` | `value as? Type` |
| C# | `(Type)value` | `value as Type` |

