# 模块与包管理

LongLang 使用包（package）来组织代码，支持模块化开发。

## 包声明

每个 LongLang 源文件必须以包声明开始：

```longlang
package main

fn main() {
    fmt.Println("Hello, World!")
}
```

## 包命名规范

| 规范 | 说明 | 示例 |
|------|------|------|
| 小写字母 | 包名应使用小写字母 | `package utils` |
| 简短有意义 | 包名应简洁且描述其功能 | `package math` |
| 避免冲突 | 不要与标准库或常用包重名 | - |
| main 包 | 可执行程序的入口包 | `package main` |

## 导入包

使用 `import` 语句导入其他包：

```longlang
package main

import "fmt"
import "math"

fn main() {
    fmt.Println("Hello!")
}
```

### 多个导入

```longlang
package main

import "fmt"
import "strings"
import "math"

fn main() {
    // 使用导入的包
}
```

## 内置包

LongLang 提供了以下内置包：

| 包名 | 说明 | 常用函数 |
|------|------|----------|
| `fmt` | 格式化输入输出 | `Println`, `Print`, `Printf` |

### fmt 包

提供格式化输出功能：

```longlang
package main

fn main() {
    // 打印并换行
    fmt.Println("Hello, World!")
    fmt.Println("多个参数:", 1, 2, 3)
    
    // 打印不换行
    fmt.Print("Hello ")
    fmt.Print("World")
    fmt.Println("")  // 换行
    
    // 格式化打印
    name := "Alice"
    age := 25
    fmt.Printf("姓名: %s, 年龄: %d\n", name, age)
}
```

#### 格式化占位符

| 占位符 | 说明 | 示例 |
|--------|------|------|
| `%d` | 整数 | `fmt.Printf("%d", 42)` → `42` |
| `%f` | 浮点数 | `fmt.Printf("%f", 3.14)` → `3.140000` |
| `%s` | 字符串 | `fmt.Printf("%s", "hi")` → `hi` |
| `%t` | 布尔值 | `fmt.Printf("%t", true)` → `true` |
| `%v` | 任意值（默认格式）| `fmt.Printf("%v", x)` |
| `%%` | 百分号 | `fmt.Printf("%%")` → `%` |

## 项目结构

推荐的项目结构：

```
myproject/
├── main.long          # 主程序入口
├── utils/
│   ├── string.long    # 字符串工具
│   └── math.long      # 数学工具
├── models/
│   └── user.long      # 用户模型
└── README.md
```

## main 包

可执行程序必须包含 `main` 包和 `main` 函数：

```longlang
package main

fn main() {
    // 程序入口点
    fmt.Println("程序开始")
}
```

## 包的可见性

LongLang 使用访问修饰符控制成员的可见性：

| 修饰符 | 可见性 | 说明 |
|--------|--------|------|
| `public` | 公开 | 可以被其他包访问 |
| `private` | 私有 | 只能在当前包/类内访问 |
| `protected` | 受保护 | 可以被子类访问 |

```longlang
package mylib

class Calculator {
    public value int           // 公开属性
    private cache int          // 私有属性
    
    public function add(n:int) int {
        return this.value + n
    }
    
    private function helper() {
        // 内部辅助方法
    }
}
```

## 包使用示例

### 文件：main.long

```longlang
package main

fn main() {
    // 使用内置 fmt 包
    fmt.Println("=== 包管理示例 ===")
    
    // 定义和使用变量
    name := "LongLang"
    version := "1.0.0"
    
    fmt.Printf("语言: %s\n", name)
    fmt.Printf("版本: %s\n", version)
}
```

## 注意事项

1. **包声明必须在文件开头**：`package` 语句必须是文件的第一个非注释语句
2. **main 包是特殊的**：它定义了可执行程序的入口
3. **导入未使用的包**：目前不会报错，但建议只导入需要的包
4. **循环导入**：避免包之间的循环依赖

## 最佳实践

| 实践 | 说明 |
|------|------|
| 单一职责 | 每个包应该有明确的单一职责 |
| 命名清晰 | 包名应该清晰表达其功能 |
| 最小化导出 | 只公开必要的接口，隐藏实现细节 |
| 文档注释 | 为公开的函数和类型添加注释 |

