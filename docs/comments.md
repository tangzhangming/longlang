# 注释

注释用于在代码中添加说明和解释，不会被执行。LongLang 支持单行注释和块注释。

## 单行注释

使用 `//` 开始单行注释，从 `//` 到行尾的内容都会被忽略：

```longlang
// 这是一个单行注释

fn main() {
    // 声明变量
    name := "Alice"  // 行尾注释
    
    // 打印问候语
    fmt.Println("Hello,", name)
}
```

## 块注释（多行注释）

使用 `/*` 和 `*/` 包围块注释，可以跨越多行：

```longlang
/*
 * 这是一个块注释
 * 可以跨越多行
 * 常用于函数或类的文档说明
 */

fn main() {
    /* 这是行内块注释 */
    name := "Alice"
    
    /*
    临时禁用的代码块：
    age := 25
    fmt.Println(age)
    */
    
    fmt.Println("Hello,", name)
}
```

## 注释用途

### 1. 代码说明

```longlang
// 计算圆的面积
// 参数: radius - 圆的半径
// 返回: 圆的面积
fn circleArea(radius:float) float {
    pi := 3.14159
    return pi * radius * radius
}
```

### 2. 临时禁用代码

```longlang
fn main() {
    a := 10
    // b := 20  // 暂时不需要这个变量
    fmt.Println(a)
}
```

### 3. TODO 标记

```longlang
fn processData(data:string) {
    // TODO: 添加数据验证
    // TODO: 处理空字符串情况
    fmt.Println(data)
}
```

### 4. 分隔代码块

```longlang
fn main() {
    // ========== 初始化 ==========
    name := "Alice"
    age := 25
    
    // ========== 处理逻辑 ==========
    if age >= 18 {
        fmt.Println(name, "是成年人")
    }
    
    // ========== 输出结果 ==========
    fmt.Println("处理完成")
}
```

## 注释规范建议

| 场景 | 建议 |
|------|------|
| 函数说明 | 在函数定义前添加功能、参数、返回值说明 |
| 复杂逻辑 | 解释算法思路或业务逻辑 |
| 重要决策 | 说明为什么选择某种实现方式 |
| 临时代码 | 标注 TODO 或 FIXME |
| 行尾注释 | 简短说明变量或表达式的用途 |

## 注释示例

```longlang
// Package main 是程序的入口包
package main

// Person 表示一个人的信息
class Person {
    public name string   // 姓名
    public age int       // 年龄
    
    // 构造函数
    // 参数: name - 姓名, age - 年龄
    public function __construct(name:string, age:int) {
        this.name = name
        this.age = age
    }
    
    // 获取问候语
    // 返回: 包含姓名的问候字符串
    public function greet() string {
        return "Hello, " + this.name
    }
}

fn main() {
    // 创建用户对象
    user := new Person("Alice", 25)
    
    // 输出问候语
    fmt.Println(user.greet())
}
```

## 注释类型对比

| 类型 | 语法 | 用途 |
|------|------|------|
| 单行注释 | `// 注释内容` | 简短说明、行尾注释 |
| 块注释 | `/* 注释内容 */` | 多行说明、文档注释、临时禁用代码 |

## 注意事项

1. **块注释不能嵌套**：`/* /* */ */` 是不允许的
2. **注释不要过度**：代码应该尽量自解释，注释用于补充说明
3. **保持同步**：修改代码时记得更新相关注释
4. **使用中文或英文**：保持注释语言的一致性
5. **文档注释**：推荐在函数和类定义前使用块注释说明功能

