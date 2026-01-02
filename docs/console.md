# 控制台 (System.Console)

LongLang 提供 `System.Console` 类用于标准输入输出，设计参照 C# Console 类。

## 命名空间

```longlang
use System.Console
```

## 输出方法

### write

输出内容，不换行。

```longlang
Console.write("Hello, ")
Console.write("World!")
// 输出: Hello, World!
```

### writeLine

输出内容并换行。

```longlang
Console.writeLine("Hello, World!")
// 输出: Hello, World! (换行)

// 多参数（自动用空格分隔）
Console.writeLine("Name:", "LongLang", "Version:", "0.2.0")
// 输出: Name: LongLang Version: 0.2.0
```

### writeEmptyLine

输出空行。

```longlang
Console.writeLine("第一行")
Console.writeEmptyLine()
Console.writeLine("第三行")
```

## 输入方法

### readLine

读取一行输入（阻塞直到用户按回车）。

```longlang
Console.write("请输入您的名字: ")
name := Console.readLine()
Console.writeLine("Hello, " + name + "!")
```

### read

读取单个字符，返回 ASCII 码。如果到达输入末尾返回 -1。

```longlang
charCode := Console.read()
if charCode == -1 {
    Console.writeLine("输入结束")
} else {
    Console.writeLine("输入字符: " + toString(charCode))
}
```

## 控制台控制

### clear

清除控制台屏幕。

```longlang
Console.clear()
```

### setCursorPosition

设置光标位置。

```longlang
// 将光标移动到第 10 列，第 5 行（从 0 开始）
Console.setCursorPosition(10, 5)
Console.writeLine("从这里开始输出")
```

### getWindowWidth / getWindowHeight

获取控制台窗口尺寸。

```longlang
width := Console.getWindowWidth()
height := Console.getWindowHeight()
Console.writeLine("窗口大小: " + toString(width) + "x" + toString(height))
```

### setTitle

设置控制台窗口标题。

```longlang
Console.setTitle("我的应用程序")
```

### beep

发出蜂鸣声。

```longlang
Console.writeLine("操作完成!")
Console.beep()
```

## 完整示例

```longlang
use System.Console

class Main {
    public static function main() {
        Console.setTitle("欢迎程序")
        Console.writeLine("=== 欢迎使用 LongLang ===")
        Console.writeEmptyLine()
        
        Console.write("请输入您的姓名: ")
        name := Console.readLine()
        
        Console.write("请输入您的年龄: ")
        ageStr := Console.readLine()
        age := toInt(ageStr)
        
        Console.writeEmptyLine()
        Console.writeLine("个人信息:")
        Console.writeLine("  姓名: " + name)
        Console.writeLine("  年龄: " + toString(age))
        
        Console.writeEmptyLine()
        Console.writeLine("按任意键继续...")
        Console.read()
        
        Console.clear()
        Console.writeLine("程序结束")
        Console.beep()
    }
}
```

## 从 fmt 迁移

如果你之前使用 `fmt` 模块，需要改为 `Console`：

### 迁移对照表

| 旧代码 | 新代码 |
|--------|--------|
| `fmt.println("Hello")` | `Console.writeLine("Hello")` |
| `fmt.print("Hello")` | `Console.write("Hello")` |
| `fmt.printf("Name: %s", name)` | `Console.writeLine("Name:", name)` |

### 迁移步骤

1. 在文件开头添加：
   ```longlang
   use System.Console
   ```

2. 将所有的 `fmt.println` 替换为 `Console.writeLine`
3. 将所有的 `fmt.print` 替换为 `Console.write`
4. 移除所有 `fmt.printf` 调用，改用字符串拼接或 `Console.writeLine` 的多参数形式

