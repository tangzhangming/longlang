# 标准库

LongLang 的标准库**使用 LongLang 语言自身编写**，存放在 `stdlib/` 目录下。

## 设计理念

- **自举**：标准库用 LongLang 编写，展示语言能力
- **内置函数**：仅 `fmt` 等底层 I/O 用 Go 实现（必须与系统交互）
- **可扩展**：用户可以添加自己的 `.long` 库文件

## 导入标准库

```longlang
import "testing"
import "math"
```

## fmt - 格式化输出（内置）

`fmt` 是内置库，无需导入即可使用。由 Go 实现，因为需要与系统 I/O 交互。

```longlang
fmt.println("Hello, World!")
fmt.print("不换行")
fmt.printf("格式化: %s", "值")
```

| 函数 | 说明 | 示例 |
|------|------|------|
| `println(args...)` | 打印并换行 | `fmt.println("hello", 123)` |
| `print(args...)` | 打印不换行 | `fmt.print("hello")` |
| `printf(format, args...)` | 格式化打印 | `fmt.printf("num: %d", 10)` |

## math - 数学库

**用 LongLang 编写**，位于 `stdlib/math.long`。

提供数学计算函数和常量。

```longlang
import "math"

fn main() {
    fmt.println("PI = " + math.PI)
    fmt.println("2^10 = " + math.pow(2, 10))
    fmt.println("5! = " + math.factorial(5))
}
```

### 常量

| 常量 | 值 | 说明 |
|------|-----|------|
| `math.PI` | 3.14159... | 圆周率 π |
| `math.E` | 2.71828... | 自然常数 e |

### 函数列表

| 函数 | 说明 | 示例 |
|------|------|------|
| `abs(x)` | 绝对值 | `math.abs(-5)` → `5` |
| `max(a, b)` | 最大值 | `math.max(3, 5)` → `5` |
| `min(a, b)` | 最小值 | `math.min(3, 5)` → `3` |
| `pow(base, exp)` | 幂运算 | `math.pow(2, 3)` → `8` |
| `square(x)` | 平方 | `math.square(5)` → `25` |
| `cube(x)` | 立方 | `math.cube(3)` → `27` |
| `isEven(x)` | 是否偶数 | `math.isEven(4)` → `true` |
| `isOdd(x)` | 是否奇数 | `math.isOdd(5)` → `true` |
| `clamp(v, min, max)` | 限制范围 | `math.clamp(15, 0, 10)` → `10` |
| `sign(x)` | 符号 | `math.sign(-5)` → `-1` |
| `factorial(n)` | 阶乘 | `math.factorial(5)` → `120` |
| `gcd(a, b)` | 最大公约数 | `math.gcd(12, 8)` → `4` |
| `lcm(a, b)` | 最小公倍数 | `math.lcm(3, 4)` → `12` |

## 完整示例

```longlang
package main

import "testing"
import "math"

fn main() {
    // 数学计算
    testing.assertEqual(math.pow(2, 10), 1024, "2^10 = 1024")
    testing.assertEqual(math.factorial(5), 120, "5! = 120")
    testing.assertEqual(math.gcd(24, 36), 12, "gcd(24, 36) = 12")
    
    testing.assertTrue(math.isEven(100), "100 是偶数")
    testing.assertTrue(math.isOdd(99), "99 是奇数")
    
    testing.summary()
}
```

## 添加自定义库

你可以在 `stdlib/` 目录下添加自己的 `.long` 文件：

```longlang
// stdlib/mylib.long
package mylib

fn hello():string {
    return "Hello from mylib!"
}
```

然后在代码中使用：

```longlang
import "mylib"

fn main() {
    fmt.println(mylib.hello())
}
```

## 目录结构

```
longlang/
├── stdlib/
│   └── math.long       # 数学库（LongLang）
├── internal/
│   └── interpreter/
│       └── builtins.go # fmt 等内置函数（Go）
└── ...
```

