# longlang 包系统和类系统使用文档

## 概述

longlang 现在支持包系统和类系统，允许你组织代码并实现面向对象编程。

## 包系统

### 1. long.mod 文件

在项目根目录创建 `long.mod` 文件，定义项目的基础路径：

```
module github.com/example/myproject
```

### 2. 包声明

每个 `.long` 文件必须以 `package` 声明开始：

```long
package packageName
```

### 3. 导入包

使用 `import` 语句导入其他包：

```long
import "util.string"
import "app.models.UserModel"
```

导入路径规则：
- 基于 `long.mod` 的 module 路径
- 点号分隔的路径（如 `util.string`）会映射到 `util/string.long` 文件

### 4. 使用导入的包

#### 调用函数

```long
import "util.string"

fn main() {
    string.test()  // 调用 util.string 包中的 test 函数
}
```

#### 使用类

```long
import "app.models.UserModel"

fn main() {
    // 创建对象
    user := new UserModel("John")
    
    // 调用实例方法
    name := user.getName()
    
    // 调用静态方法
    tableName := UserModel::getTableName()
}
```

## 类系统

### 1. 类声明

```long
class ClassName {
    // 类成员
}
```

**重要规则**：
- 类名必须与文件名一致才能被导出
- 一个文件可以有多个类，但只有与文件名相同的类可以被导出

### 2. 成员变量

```long
class UserModel {
    public name string
    private age int
    protected email string = "default@example.com"
}
```

**规则**：
- 必须声明访问修饰符（`public`、`private`、`protected`）
- 必须声明类型（不允许类型推导）
- 可以初始化默认值

### 3. 构造方法

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

**规则**：
- 方法名固定为 `__construct`
- 必须是 `public`
- 支持参数默认值
- 使用 `this.` 访问当前对象的成员

### 4. 实例方法

```long
class UserModel {
    private name string
    
    public function getName() string {
        return this.name  // 必须使用 this.
    }
    
    public function setName(name:string) {
        this.name = name
    }
}
```

**规则**：
- 必须声明访问修饰符
- 支持参数默认值
- 支持命名参数
- 支持方法重载
- **必须使用 `this.` 访问当前对象成员**

### 5. 静态方法

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

**规则**：
- `static` 关键字必须放在访问修饰符后面
- 不能访问实例成员（不能使用 `this`）
- 使用 `ClassName::methodName()` 调用
- 在类内部可以使用 `self::` 或类名调用

### 6. 创建对象

```long
import "app.models.UserModel"

fn main() {
    user := new UserModel("John", 25)
    name := user.getName()
}
```

### 7. 访问对象成员

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

## 完整示例

### 项目结构

```
myproject/
├── long.mod
├── main.long
├── util/
│   └── string.long
└── app/
    └── models/
        └── UserModel.long
```

### long.mod

```
module github.com/example/myproject
```

### util/string.long

```long
package string

fn test() {
    fmt.Println("test from string package")
}
```

### app/models/UserModel.long

```long
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

### main.long

```long
package main

import "util.string"
import "app.models.UserModel"

fn main() {
    // 调用包函数
    string.test()
    
    // 创建对象
    user := new UserModel("John", 25)
    
    // 访问实例方法
    name := user.getName()
    fmt.Println(name)
    user.setEmail("john@example.com")
    
    // 访问静态方法
    tableName := UserModel::getTableName()
    fmt.Println(tableName)
}
```

## 注意事项

1. **文件与类对应**：只有与文件名相同的类可以被导出
2. **访问修饰符**：类成员必须声明访问修饰符
3. **this 关键字**：访问当前对象成员必须使用 `this.`
4. **静态方法调用**：使用 `ClassName::methodName()` 语法
5. **包路径**：导入路径基于 `long.mod` 的 module 路径

## 当前限制

1. 相对路径导入暂未实现
2. 循环依赖检测是基础版本
3. 访问修饰符检查需要进一步完善
4. 继承和多态暂未实现

## 未来计划

1. 支持相对路径导入
2. 完善循环依赖检测
3. 实现继承和多态
4. 支持接口
5. 支持抽象类

