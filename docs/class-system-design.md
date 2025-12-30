# 类系统设计文档

## 概述

longlang 的类系统设计参考了 Java、PHP 等面向对象语言，但有自己的特点：
- 严格的文件与类对应关系
- 类名必须与文件名一致才能被导出
- 支持访问修饰符（public、private、protected）
- 支持静态方法和实例方法
- 使用 `this` 关键字访问当前对象
- 使用 `::` 调用静态方法

## 1. 类声明

### 语法

```long
class ClassName {
    // 类成员
}
```

### 规则

1. **文件与类对应**：
   - 类名必须与文件名一致才能被导出
   - 一个文件可以有多个类，但只有与文件名相同的类可以被导出
   - 例如：`UserModel.long` 文件中，只有 `UserModel` 类可以被导出

2. **类名规则**：
   - 必须符合标识符规则
   - 通常使用 PascalCase 命名
   - 必须与文件名（不含扩展名）一致才能导出

## 2. 类成员变量

### 语法

```long
访问修饰符 变量名 变量类型
访问修饰符 变量名 变量类型 = 初始值
```

### 访问修饰符

- `public`：公开，可以在类外部访问
- `private`：私有，只能在类内部访问
- `protected`：受保护，可以在类内部和子类中访问（当前版本暂不支持继承）

### 规则

- **不允许类型推导**：必须显式声明类型
- **可以初始化**：可以在声明时赋初始值
- **必须声明可见性**：必须使用访问修饰符

### 示例

```long
class UserModel {
    public name string
    private age int
    protected email string = "default@example.com"
}
```

## 3. 构造方法

### 语法

```long
public function __construct(参数列表) {
    // 构造逻辑
}
```

### 规则

- 方法名固定为 `__construct`
- 必须是 `public`
- 可以有参数
- 可以有默认参数
- 可以有命名参数

### 示例

```long
class UserModel {
    private name string
    private age int
    
    public function __construct(name:string, age:int = 0) {
        this.name = name
        this.age = age
    }
}
```

## 4. 实例方法

### 语法

```long
访问修饰符 function 方法名(参数列表): 返回类型 {
    // 方法体
}
```

### 规则

- 必须声明访问修饰符
- 支持参数默认值
- 支持命名参数
- 支持方法重载（相同方法名，不同参数）
- 使用 `this.` 访问当前对象的成员

### 访问当前对象成员

- **必须使用 `this.`**：不能像 Java 一样省略
- `this` 是关键字，指向当前对象实例

### 示例

```long
class UserModel {
    private name string
    
    public function getName() string {
        return this.name  // 必须使用 this.
    }
    
    public function setName(name:string) {
        this.name = name
    }
    
    // 方法重载
    public function setAge(age:int) {
        this.age = age
    }
    
    public function setAge(age:string) {
        this.age = parseInt(age)
    }
}
```

## 5. 静态方法

### 语法

```long
访问修饰符 static function 方法名(参数列表): 返回类型 {
    // 方法体
}
```

### 规则

- `static` 关键字必须放在访问修饰符后面
- 不能访问实例成员（不能使用 `this`）
- 可以通过类名或 `self::` 调用
- 当前类中可以使用 `self::` 或类名访问

### 调用语法

```long
ClassName::methodName()  // 使用类名调用
self::methodName()       // 在类内部使用 self:: 调用
```

### 示例

```long
class UserModel {
    public static function getTableName() string {
        return "users"
    }
    
    public static function create() {
        table := self::getTableName()  // 使用 self::
        // 或
        table := UserModel::getTableName()  // 使用类名
    }
}
```

## 6. 创建对象

### 语法

```long
变量名 := new ClassName(参数列表)
```

### 规则

- 使用 `new` 关键字创建对象
- 调用类的构造方法
- 返回对象实例

### 示例

```long
import "app.models.UserModel"

fn main() {
    user := new UserModel("John", 25)
    name := user.getName()
}
```

## 7. 访问对象成员

### 访问实例成员

```long
对象.成员变量
对象.方法名(参数)
```

### 访问静态成员

```long
类名::静态方法名(参数)
```

### 示例

```long
import "app.models.UserModel"

fn main() {
    // 创建对象
    user := new UserModel("John")
    
    // 访问实例方法
    name := user.getName()
    user.setName("Jane")
    
    // 访问静态方法
    tableName := UserModel::getTableName()
}
```

## 8. 访问修饰符的作用域

### public

- 可以在类外部访问
- 可以在类内部访问
- 可以在子类中访问（未来支持继承后）

### private

- 只能在类内部访问
- 不能在类外部访问
- 不能在子类中访问

### protected

- 可以在类内部访问
- 可以在子类中访问（未来支持继承后）
- 不能在类外部访问

## 9. 实现要点

### 9.1 AST 扩展

在 `parser` 中添加：
- `ClassStatement`：类声明语句
- `ClassMember`：类成员（变量、方法）
- `MethodDeclaration`：方法声明
- `ThisExpression`：this 表达式
- `NewExpression`：new 表达式
- `StaticCallExpression`：静态方法调用表达式

### 9.2 词法分析扩展

在 `lexer` 中添加：
- `CLASS` 关键字
- `PUBLIC`、`PRIVATE`、`PROTECTED` 关键字
- `STATIC` 关键字
- `THIS` 关键字
- `NEW` 关键字

### 9.3 解释器扩展

在 `interpreter` 中添加：
- `Class` 对象类型
- `Instance` 对象类型（类实例）
- 类环境管理
- 方法查找和调用
- 静态方法调用
- 访问修饰符检查
- `this` 绑定

### 9.4 对象系统扩展

- 类对象：存储类定义（方法、成员变量）
- 实例对象：存储实例数据（成员变量值）
- 方法调用：支持实例方法和静态方法
- 成员访问：支持访问修饰符检查

## 10. 示例

### 完整的类定义

```long
// UserModel.long
package UserModel

class UserModel {
    // 成员变量
    public name string
    private age int
    private email string = "default@example.com"
    
    // 构造方法
    public function __construct(name:string, age:int = 0) {
        this.name = name
        this.age = age
    }
    
    // 实例方法
    public function getName() string {
        return this.name
    }
    
    private function getAge() int {
        return this.age
    }
    
    public function setEmail(email:string) {
        this.email = email
    }
    
    // 静态方法
    public static function getTableName() string {
        return "users"
    }
    
    public static function create(name:string) {
        table := self::getTableName()
        // 创建逻辑
    }
}
```

### 使用类

```long
// main.long
package main

import "app.models.UserModel"

fn main() {
    // 创建对象
    user := new UserModel("John", 25)
    
    // 访问实例方法
    name := user.getName()
    user.setEmail("john@example.com")
    
    // 访问静态方法
    tableName := UserModel::getTableName()
    UserModel::create("Jane")
}
```

## 11. 未来扩展

1. **继承**：支持类继承
2. **接口**：支持接口定义和实现
3. **多态**：支持方法重写和多态
4. **抽象类**：支持抽象类
5. **访问器**：支持 getter/setter 方法
6. **属性**：支持属性（property）

