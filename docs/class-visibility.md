# 类可见性控制 (Class Visibility)

LongLang 提供了类、接口和枚举的可见性控制机制，用于管理跨命名空间的访问权限。这有助于实现更好的封装和模块化设计。

## 可见性修饰符

| 修饰符 | 作用 | 说明 |
|--------|------|------|
| `public` | 公开 | 任何命名空间都可以访问 |
| `internal` | 内部 | 仅当前命名空间及其子/父命名空间可访问 |
| （默认） | 内部 | 未指定时等同于 `internal` |

## 语法

### 类

```longlang
namespace MyApp.Models

// 公开类 - 可被任何命名空间访问
public class User {
    private name string
    
    public function __construct(name: string) {
        this.name = name
    }
    
    public function getName() string {
        return this.name
    }
}

// 内部类 - 仅 MyApp.Models 命名空间树内可访问
internal class UserValidator {
    public static function validate(user: User) bool {
        return user.getName().length() > 0
    }
}

// 默认为 internal
class UserHelper {
    // 只能在 MyApp.Models 命名空间树内使用
}
```

### 抽象类

```longlang
// 公开抽象类
public abstract class BaseModel {
    public abstract function save()
}

// 内部抽象类
internal abstract class InternalBase {
    // ...
}
```

### 接口

```longlang
// 公开接口
public interface Serializable {
    function serialize() string
}

// 内部接口
internal interface InternalLogger {
    function log(msg: string)
}
```

### 枚举

```longlang
// 公开枚举
public enum Color {
    Red
    Green
    Blue
}

// 内部枚举
internal enum InternalState {
    Loading
    Ready
}
```

## 命名空间树访问规则

命名空间 A 和 B 属于同一命名空间树，如果满足以下任一条件：

1. A 等于 B（同一命名空间）
2. A 是 B 的前缀（A 是 B 的父命名空间）
3. B 是 A 的前缀（B 是 A 的父命名空间）

### 命名空间树示例

```
MyApp                    // 根命名空间
├── MyApp.Models         // 子命名空间
│   └── MyApp.Models.Validators  // 孙命名空间
└── MyApp.Services       // 另一个子命名空间（兄弟关系）
```

在这个结构中：
- `MyApp` 和 `MyApp.Models` 互相可访问 internal 类型
- `MyApp.Models` 和 `MyApp.Models.Validators` 互相可访问 internal 类型
- `MyApp` 和 `MyApp.Models.Validators` 互相可访问 internal 类型
- `MyApp.Models` 和 `MyApp.Services` **不能**互相访问 internal 类型（兄弟关系）

## 访问规则表

| 声明 | 同命名空间 | 子命名空间 | 父命名空间 | 其他命名空间 |
|------|-----------|-----------|-----------|-------------|
| `public class` | ✓ | ✓ | ✓ | ✓ |
| `internal class` | ✓ | ✓ | ✓ | ✗ |
| `class`（默认） | ✓ | ✓ | ✓ | ✗ |

## 使用示例

### 示例 1: 跨命名空间访问 public 类

```longlang
namespace MyApp.Services

use MyApp.Models.User  // ✓ User 是 public，可以访问

class UserService {
    public function createUser(name: string) User {
        return new User(name)
    }
}
```

### 示例 2: 同命名空间访问 internal 类

```longlang
namespace MyApp.Models

use MyApp.Models.UserValidator  // ✓ 同命名空间，可以访问 internal 类

public class UserFactory {
    public static function create(name: string) User {
        user := new User(name)
        if UserValidator::validate(user) {
            return user
        }
        return null
    }
}
```

### 示例 3: 子命名空间访问父命名空间的 internal 类

```longlang
namespace MyApp.Models.Validators

use MyApp.Models.UserValidator  // ✓ 子命名空间可以访问父命名空间的 internal 类

public class EmailValidator {
    public static function validate(email: string) bool {
        // 可以使用父命名空间的 internal 类
        return true
    }
}
```

### 示例 4: 尝试从不相关的命名空间访问 internal 类（失败）

```longlang
namespace Other.Package

use MyApp.Models.UserValidator  // ✗ 编译错误: UserValidator 是 internal 类型

class MyClass {
    // ...
}
```

错误信息：
```
错误: 无法访问 'MyApp.Models.UserValidator': 'UserValidator' 是 internal 类型，只能在 'MyApp.Models' 命名空间树内访问
```

## 最佳实践

1. **默认使用 internal**：除非明确需要跨命名空间访问，否则不要添加 `public` 修饰符。
2. **API 设计**：只将公共 API 标记为 `public`，内部实现细节保持 `internal`。
3. **模块边界**：使用命名空间和可见性来定义清晰的模块边界。
4. **继承约束**：`public` 类只应继承 `public` 类，避免暴露 internal 实现。

## 与成员可见性的区别

类可见性（`public`/`internal`）控制的是**类本身**能否被其他命名空间访问。

成员可见性（`public`/`private`/`protected`）控制的是**类的成员**（属性和方法）能否被其他代码访问。

```longlang
namespace MyApp.Models

// 类可见性: public - 任何命名空间都可以访问这个类
public class User {
    // 成员可见性: private - 只有这个类内部可以访问
    private name string
    
    // 成员可见性: public - 任何能访问这个类的代码都可以调用
    public function getName() string {
        return this.name
    }
}
```





