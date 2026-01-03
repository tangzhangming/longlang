# 类常量

类常量提供了一种在类中定义不可变值的方式，类似于其他语言的类常量。

## 基本语法

### 定义常量

```longlang
class Math {
    public const PI = 3.14159
    public const MAX_SIZE = 100
    public const APP_NAME = "MyApp"
    public const DEBUG = false
}
```

### 访问常量

```longlang
// 外部访问
fmt.println(Math::PI)         // 3.14159
fmt.println(Math::MAX_SIZE)   // 100

// 类内部访问
class Math {
    public const PI = 3.14159
    
    public static function test() {
        fmt.println(self::PI)  // 使用 self:: 访问
    }
}
```

## 类型声明

常量支持可选的类型声明：

```longlang
class Config {
    // 类型推导
    public const VERSION = "1.0.0"    // 推导为 string
    public const MAX_RETRY = 3        // 推导为 int
    public const PI = 3.14159         // 推导为 float
    
    // 显式类型
    public const PORT i16 = 8080      // 明确指定为 i16
    public const RATE f32 = 0.15      // 明确指定为 f32
}
```

## 常量值限制

常量值必须是**字面量**，不支持表达式：

```longlang
class Example {
    public const PI = 3.14159         // ✅ OK
    public const MAX = 100            // ✅ OK
    public const NAME = "App"         // ✅ OK
    
    public const SUM = 1 + 2          // ❌ 不允许表达式
    public const NOW = getTime()      // ❌ 不允许函数调用
}
```

## 访问修饰符

常量支持访问修饰符，控制可见性：

```longlang
class Example {
    public const PUB = 1      // 可在类外部访问
    private const PRI = 2     // 仅类内部可访问
    protected const PRO = 3   // 类及其子类可访问
}
```

## 常量和静态方法同名

常量可以与静态方法同名，通过**括号**区分：

```longlang
class Test {
    public const FOO = 100
    
    public static function FOO() int {
        return 200
    }
}

Test::FOO      // 访问常量 → 100
Test::FOO()    // 调用静态方法 → 200
```

| 语法 | 含义 | 结果 |
|------|------|------|
| `ClassName::CONST_NAME` | 常量访问（无括号） | 返回常量值 |
| `ClassName::methodName()` | 静态方法调用（有括号） | 调用方法 |

## 继承

子类可以覆盖父类的常量：

```longlang
class Parent {
    public const VALUE = 100
}

class Child extends Parent {
    public const VALUE = 200  // 覆盖父类常量
}

fmt.println(Parent::VALUE)  // 100
fmt.println(Child::VALUE)   // 200
```

## 完整示例

```longlang
namespace App

class Config {
    // 应用配置常量
    public const APP_NAME = "MyApp"
    public const VERSION = "1.0.0"
    public const MAX_RETRY = 3
    public const DEBUG = false
    
    // 网络配置（显式类型）
    public const PORT i16 = 8080
    public const TIMEOUT i64 = 30000
    
    // 业务常量
    public const PI = 3.14159
    public const DEFAULT_RATE = 0.15
    
    public static function getInfo() string {
        return self::APP_NAME + " v" + self::VERSION
    }
    
    public static function main() {
        fmt.println("应用信息: " + Config::getInfo())
        fmt.println("端口: " + Config::PORT)
        fmt.println("最大重试: " + Config::MAX_RETRY)
    }
}
```

## 命名规范

| 规范 | 示例 | 说明 |
|------|------|------|
| 全大写 | `MAX_SIZE` | 推荐使用全大写，多个单词用下划线分隔 |
| 驼峰命名 | `maxSize` | 也支持，但不推荐 |
| 有意义的名称 | `PI`, `DEFAULT_PORT` | 使用清晰、描述性的名称 |

## 注意事项

1. **常量值不可变**：常量一旦定义就不能修改
2. **必须是字面量**：不支持表达式或函数调用
3. **访问方式**：使用 `::` 运算符，无括号表示常量，有括号表示方法
4. **类型推导**：如果不指定类型，根据字面量自动推导

## 与变量的区别

| 特性 | 常量 | 变量 |
|------|------|------|
| 语法 | `public const NAME = value` | `public name type = value` |
| 值 | 必须初始化 | 可选初始化 |
| 类型 | 可选 | 必须 |
| 可变性 | 不可变 | 可变 |
| 访问 | `ClassName::CONST` | `instance.name` |
| 值限制 | 必须是字面量 | 可以是表达式 |








