# 标准库

LongLang 的标准库**使用 LongLang 语言自身编写**，存放在 `stdlib/` 目录下。

## 设计理念

- **自举**：标准库用 LongLang 编写，展示语言能力
- **内置函数**：仅 `fmt` 等底层 I/O 用 Go 实现（必须与系统交互）
- **命名空间**：标准库使用 `System` 等命名空间组织

## 导入标准库

使用 `use` 语句导入标准库中的类：

```longlang
use System.Str
use System.String
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

## System 命名空间

### System.Str - 字符串静态工具类

提供字符串操作的静态方法：

```longlang
use System.Str

fmt.println(Str::length("hello"))            // 5
fmt.println(Str::upper("hello"))             // HELLO
fmt.println(Str::contains("hello", "ell"))   // true
fmt.println(Str::replace("hello", "l", "L")) // heLlo
```

### System.String - 字符串对象类

提供面向对象的字符串封装：

```longlang
use System.String

name := new String("hello world")
fmt.println(name.length())                   // 11
fmt.println(name.upper().getValue())         // HELLO WORLD

// 链式调用
result := name.trim().upper().replace("WORLD", "LONGLANG")
fmt.println(result.getValue())
```

详细方法列表请参阅 [字符串文档](string.md)。

## 字符串语法糖

原始字符串可以直接调用方法，无需导入任何库：

```longlang
name := "hello world"
fmt.println(name.length())          // 11
fmt.println(name.upper())           // HELLO WORLD
fmt.println(name.contains("world")) // true

// 链式调用
result := name.trim().upper().replace("WORLD", "LONGLANG")
```

## 目录结构

```
longlang/
├── stdlib/
│   └── System/
│       ├── Exception.long           # 异常基类
│       ├── RuntimeException.long    # 运行时异常
│       ├── IOException.long         # IO 异常
│       ├── FileNotFoundException.long
│       ├── DirectoryNotFoundException.long
│       ├── PermissionException.long
│       ├── Str.long                 # 字符串静态工具类
│       ├── String.long              # 字符串对象类
│       └── IO/
│           ├── File.long            # 文件操作
│           ├── Directory.long       # 目录操作
│           ├── Path.long            # 路径操作
│           ├── FileStream.long      # 文件流
│           ├── FileInfo.long        # 文件信息
│           └── DirectoryInfo.long   # 目录信息
├── internal/
│   └── interpreter/
│       ├── builtins.go         # fmt 等内置函数（Go）
│       ├── builtins_io.go      # 文件操作内置函数（Go）
│       └── string_methods.go   # 字符串方法（Go，支持语法糖）
└── ...
```

## 添加自定义库

你可以在 `stdlib/` 目录下添加自己的命名空间和类：

```longlang
// stdlib/MyLib/Utils.long
namespace MyLib

class Utils {
    public static function hello() string {
        return "Hello from MyLib!"
    }
}
```

然后在代码中使用：

```longlang
namespace App

use MyLib.Utils

class Application {
    public static function main() {
        fmt.println(Utils::hello())
    }
}
```
