# LongLang 枚举指南

本文档介绍 LongLang 中枚举（enum）的使用方法。

## 目录

1. [简单枚举](#简单枚举)
2. [带值枚举](#带值枚举)
3. [带方法的枚举](#带方法的枚举)
4. [内置方法](#内置方法)
5. [枚举比较](#枚举比较)
6. [最佳实践](#最佳实践)

---

## 简单枚举

简单枚举是一组命名常量的集合：

```longlang
enum Color {
    Red
    Green
    Blue
}
```

### 访问枚举成员

```longlang
color := Color::Red
fmt.println(color.name())     // "Red"
fmt.println(color.ordinal())  // 0
```

### 遍历所有成员

```longlang
colors := Color::cases()
for i := 0; i < len(colors); i++ {
    c := colors[i]
    fmt.println(c.name())
}
```

---

## 带值枚举

带值枚举的每个成员都关联一个值（int 或 string）：

### 整数值枚举

```longlang
enum Status: int {
    Pending = 0
    Approved = 1
    Rejected = 2
}

status := Status::Pending
fmt.println(status.value())  // 0
```

### 自动递增

整数枚举支持自动递增：

```longlang
enum Priority: int {
    Low = 1
    Medium      // 自动 = 2
    High        // 自动 = 3
    Critical    // 自动 = 4
}
```

### 字符串值枚举

```longlang
enum HttpMethod: string {
    Get = "GET"
    Post = "POST"
    Put = "PUT"
    Delete = "DELETE"
}

method := HttpMethod::Post
fmt.println(method.value())  // "POST"
```

### 从值创建枚举

```longlang
// from() - 无效值抛异常
status := Status::from(1)  // Status::Approved

// tryFrom() - 无效值返回 null
status := Status::tryFrom(99)  // null

// valueOf() - 从名称创建
status := Status::valueOf("Approved")  // Status::Approved
```

---

## 带方法的枚举

枚举可以定义方法：

```longlang
enum Direction {
    North
    South
    East
    West
    
    // 实例方法
    public function opposite() Direction {
        if this == Direction::North {
            return Direction::South
        }
        if this == Direction::South {
            return Direction::North
        }
        if this == Direction::East {
            return Direction::West
        }
        return Direction::East
    }
    
    public function description() string {
        if this == Direction::North {
            return "北方"
        }
        if this == Direction::South {
            return "南方"
        }
        // ...
        return "未知"
    }
}

// 使用
dir := Direction::North
fmt.println(dir.opposite().name())  // "South"
fmt.println(dir.description())       // "北方"
```

---

## 内置方法

### 实例方法

| 方法 | 返回类型 | 说明 |
|------|----------|------|
| `name()` | string | 返回成员名称 |
| `ordinal()` | int | 返回序号（从 0 开始）|
| `value()` | int/string | 返回成员值（仅带值枚举）|

### 静态方法

| 方法 | 返回类型 | 说明 |
|------|----------|------|
| `cases()` | []Enum | 返回所有成员数组 |
| `count()` | int | 返回成员数量 |
| `from(value)` | Enum | 从值创建（无效值抛异常）|
| `tryFrom(value)` | Enum/null | 从值创建（无效值返回 null）|
| `valueOf(name)` | Enum | 从名称创建（无效名称抛异常）|

### 使用示例

```longlang
enum Status: int {
    Pending = 0
    Approved = 1
    Rejected = 2
}

// 获取所有成员
cases := Status::cases()
for i := 0; i < len(cases); i++ {
    s := cases[i]
    fmt.println(s.name() + " = " + toString(s.value()))
}

// 成员数量
fmt.println(Status::count())  // 3

// 从值创建
s1 := Status::from(1)       // Status::Approved
s2 := Status::tryFrom(99)   // null

// 从名称创建
s3 := Status::valueOf("Rejected")  // Status::Rejected
```

---

## 枚举比较

### 同类型比较

```longlang
color1 := Color::Red
color2 := Color::Red
color3 := Color::Blue

fmt.println(color1 == color2)  // true
fmt.println(color1 == color3)  // false
fmt.println(color1 != color3)  // true
```

### 不同类型不能比较

```longlang
enum Color { Red }
enum Status: int { Pending = 0 }

// ❌ 运行时错误：不能比较不同枚举类型
// if Color::Red == Status::Pending { }
```

---

## 最佳实践

### 1. 使用枚举表示状态

```longlang
enum OrderState {
    Created
    Pending
    Paid
    Shipped
    Delivered
    Cancelled
}

class Order {
    private state OrderState
    
    public function __construct() {
        this.state = OrderState::Created
    }
    
    public function getState() OrderState {
        return this.state
    }
}
```

### 2. 使用带值枚举表示配置

```longlang
enum LogLevel: int {
    Debug = 0
    Info = 1
    Warning = 2
    Error = 3
    Fatal = 4
}

function log(level: LogLevel, message: string) {
    fmt.println("[" + level.name() + "] " + message)
}
```

### 3. 使用方法封装逻辑

```longlang
enum HttpStatus: int {
    OK = 200
    NotFound = 404
    InternalServerError = 500
    
    public function isSuccess() bool {
        v := this.value()
        return v >= 200 && v < 300
    }
    
    public function isError() bool {
        return this.value() >= 400
    }
}

status := HttpStatus::NotFound
if status.isError() {
    fmt.println("请求失败")
}
```

### 4. 使用枚举作为 Map 的键

```longlang
enum Day {
    Mon
    Tue
    Wed
    Thu
    Fri
    Sat
    Sun
}

// 枚举可以用 name() 作为 Map 键
schedule := map[string]string{
    "Mon": "工作",
    "Sat": "休息",
    "Sun": "休息"
}

day := Day::Mon
fmt.println(schedule[day.name()])  // "工作"
```

---

## 与其他语言的对比

| 特性 | LongLang | Java | PHP | C# |
|------|----------|------|-----|-----|
| 简单枚举 | ✅ | ✅ | ✅ | ✅ |
| 带值枚举 | ✅ | ✅ | ✅ | ✅ |
| 自定义方法 | ✅ | ✅ | ✅ | ❌ |
| 实现接口 | 🔜 | ✅ | ✅ | ✅ |
| 带数据成员 | 🔜 | ✅ | ✅ | ❌ |

✅ 已支持 | 🔜 计划中 | ❌ 不支持

---

## 相关文档

- [类和对象](./class.md)
- [接口](./interface.md)
- [抽象类](./abstract-class.md)

