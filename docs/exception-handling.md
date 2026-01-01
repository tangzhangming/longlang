# 异常处理

LongLang 提供了完整的异常处理机制，支持 `try-catch-finally` 结构和 `throw` 语句。

## 基本语法

### try-catch

```longlang
try {
    // 可能抛出异常的代码
    result := divide(a, b)
} catch (Exception e) {
    // 异常处理
    fmt.println("错误: " + e.getMessage())
}
```

### try-catch-finally

```longlang
try {
    file := openFile("data.txt")
    content := file.read()
} catch (Exception e) {
    fmt.println("文件读取错误: " + e.getMessage())
} finally {
    // 无论是否异常都会执行
    file.close()
}
```

### 多 catch 块

```longlang
try {
    process(data)
} catch (ArithmeticException e) {
    fmt.println("算术错误: " + e.getMessage())
} catch (RuntimeException e) {
    fmt.println("运行时错误: " + e.getMessage())
} catch (Exception e) {
    fmt.println("其他错误: " + e.getMessage())
}
```

### 无类型 catch

```longlang
try {
    process(data)
} catch (e) {
    // 捕获所有异常
    fmt.println("发生错误: " + e.getMessage())
}
```

**注意**: 无类型 catch 必须放在所有类型化 catch 之后。

## throw 语句

```longlang
// 抛出异常
throw new Exception("错误消息")

// 抛出具体类型的异常
throw new ArithmeticException("除数不能为零")

// 在方法中抛出
class Calculator {
    public static function divide(a: int, b: int) int {
        if b == 0 {
            throw new ArithmeticException("除数不能为零")
        }
        return a / b
    }
}
```

## 异常类层次结构

LongLang 提供了一组预定义的异常类，位于 `System` 命名空间：

| 异常类 | 说明 | 父类 |
|--------|------|------|
| `Exception` | 所有异常的基类 | - |
| `RuntimeException` | 运行时异常 | Exception |
| `ArithmeticException` | 算术异常（如除以零） | RuntimeException |
| `IndexOutOfBoundsException` | 索引越界异常 | RuntimeException |
| `NullPointerException` | 空指针异常 | RuntimeException |
| `TypeError` | 类型错误异常 | RuntimeException |
| `InvalidArgumentException` | 无效参数异常 | Exception |
| `FileNotFoundException` | 文件未找到异常 | Exception |
| `IOException` | IO 异常 | Exception |

## Exception 基类方法

| 方法 | 返回类型 | 说明 |
|------|----------|------|
| `getMessage()` | string | 获取错误消息 |
| `getCode()` | int | 获取错误代码 |
| `toString()` | string | 转换为字符串 |

## 使用示例

### 基本异常处理

```longlang
namespace App

use System.Exception

class MainApp {
    public static function main() {
        try {
            throw new Exception("测试异常")
        } catch (Exception e) {
            fmt.println("捕获: " + e.getMessage())
        }
    }
}
```

### 嵌套 try-catch

```longlang
try {
    fmt.println("外层 try")
    try {
        throw new Exception("内层异常")
    } catch (Exception e) {
        fmt.println("内层 catch: " + e.getMessage())
    }
    fmt.println("外层 try 继续")
} catch (Exception e) {
    fmt.println("外层 catch: " + e.getMessage())
}
```

### 在方法中使用异常

```longlang
namespace App

use System.ArithmeticException

class Calculator {
    public static function divide(a: int, b: int) int {
        if b == 0 {
            throw new ArithmeticException("除数不能为零")
        }
        return a / b
    }
    
    public static function safeDivide(a: int, b: int) int {
        try {
            return Calculator::divide(a, b)
        } catch (ArithmeticException e) {
            fmt.println("除法错误: " + e.getMessage())
            return 0
        }
    }
}
```

### finally 块保证执行

```longlang
class ResourceManager {
    private name string
    
    public function __construct(name: string) {
        this.name = name
    }
    
    public function open() {
        fmt.println("打开: " + this.name)
    }
    
    public function close() {
        fmt.println("关闭: " + this.name)
    }
}

class MainApp {
    public static function main() {
        resource := new ResourceManager("文件")
        try {
            resource.open()
            // 可能抛出异常的操作
            throw new Exception("操作失败")
        } catch (Exception e) {
            fmt.println("错误: " + e.getMessage())
        } finally {
            // 无论是否异常，都会关闭资源
            resource.close()
        }
    }
}
```

## 调用父类构造函数

在继承异常类时，使用 `super::__construct()` 调用父类构造函数：

```longlang
namespace System

use System.Exception

class CustomException extends Exception {
    public function __construct(message: string) {
        super::__construct(message, 0)
    }
}
```

## 运行时错误捕获

LongLang 的 `try-catch` 不仅能捕获通过 `throw` 抛出的异常，还能捕获运行时错误（如除以零、Map 键不存在等）。

### 除以零

```longlang
try {
    result := 100 / 0
} catch (Exception e) {
    fmt.println("错误: " + e.getMessage())
    // 输出: 错误: 除以零
}
```

### Map 键不存在

```longlang
scores := map[string]int{"Alice": 100}

try {
    x := scores["Unknown"]
} catch (Exception e) {
    fmt.println("错误: " + e.getMessage())
    // 输出: 错误: Map 键不存在: Unknown
}
```

### 数组索引越界

```longlang
arr := []int{1, 2, 3}

try {
    x := arr[10]
} catch (Exception e) {
    fmt.println("错误: " + e.getMessage())
    // 输出: 错误: 数组索引越界：索引 10 超出范围 [0, 2]
}
```

### 使用 isset 预防错误

更推荐的做法是在访问前检查，避免异常：

```longlang
// Map 安全访问
if isset(scores, "Alice") {
    fmt.println(scores["Alice"])
}

// 数组安全访问
if isset(arr, 5) {
    fmt.println(arr[5])
}
```

## 最佳实践

1. **使用具体的异常类型**: 尽可能使用具体的异常类型，而不是通用的 `Exception`。

2. **catch 块顺序**: 将更具体的异常类型放在前面，更通用的放在后面。

3. **使用 finally 清理资源**: 在 finally 块中进行资源清理（如关闭文件、释放连接）。

4. **不要捕获所有异常后忽略**: 如果捕获异常，应该进行适当的处理（记录日志、转换为用户友好的消息等）。

5. **异常消息要有意义**: 抛出异常时，提供清晰、有意义的错误消息。

6. **优先使用 isset 检查**: 对于 Map 和数组访问，优先使用 `isset()` 检查，而不是依赖异常捕获。

7. **finally 块不能访问 catch 变量**: `finally` 块无法访问 `catch` 块中声明的异常变量。

## 注意事项

### finally 块作用域

`finally` 块无法访问 `catch` 块中的异常变量：

```longlang
try {
    throw new Exception("错误")
} catch (Exception e) {
    fmt.println(e.getMessage())  // ✅ 正确
} finally {
    // fmt.println(e.getMessage())  // ❌ 错误: 未定义的标识符 e
    fmt.println("清理完成")
}
```

### catch 块顺序

更具体的异常类型必须放在前面：

```longlang
try {
    // ...
} catch (ArithmeticException e) {
    // 具体类型在前
} catch (RuntimeException e) {
    // 父类类型在后
} catch (Exception e) {
    // 最通用的在最后
}
```

