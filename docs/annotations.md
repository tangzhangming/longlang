# 注解 (Annotations)

LongLang 支持 Java 风格的注解语法，用于为类、方法、字段添加元数据。

## 定义注解

使用 `annotation` 关键字定义注解：

```longlang
annotation Entity {
    table string = ""
}

annotation Column {
    name string = ""
    nullable bool = true
    length int = 255
}

annotation Deprecated {
    message string = ""
}
```

### 字段类型

注解字段支持以下类型：
- `string` - 字符串
- `int` - 整数
- `bool` - 布尔值
- `float` - 浮点数
- `any` - 任意类型

### 默认值

字段可以指定默认值，使用注解时可以省略有默认值的参数。

## 使用注解

### 类注解

```longlang
@Entity(table: "users")
public class User {
    // ...
}

// 多个注解
@Entity(table: "accounts")
@Serializable
public class Account {
    // ...
}
```

### 字段注解

```longlang
public class User {
    @Column(name: "user_id", nullable: false)
    private id int
    
    @Column(name: "username", length: 50)
    private username string
}
```

### 方法注解

```longlang
public class User {
    @Deprecated(message: "Use getUsername instead")
    public function getName() string {
        return this.name
    }
}
```

## 反射 API

使用 `System.Reflection.Reflection` 类在运行时获取注解信息：

```longlang
use System.Reflection.Reflection

// 检查类是否有注解
if Reflection::hasClassAnnotation("User", "Entity") {
    fmt.println("User 类有 @Entity 注解")
}

// 获取注解详情
entityAnn := Reflection::getClassAnnotation("User", "Entity")

// 获取注解参数
tableName := Reflection::getAnnotationParam(entityAnn, "table")
fmt.println("Table: " + tableName)

// 获取所有注解
annotations := Reflection::getClassAnnotations("User")
fmt.println("注解数量: " + toString(len(annotations)))
```

### Reflection 方法

| 方法 | 说明 |
|------|------|
| `getClassAnnotations(className)` | 获取类的所有注解 |
| `hasClassAnnotation(className, annName)` | 检查类是否有指定注解 |
| `getClassAnnotation(className, annName)` | 获取类上的指定注解 |
| `getAnnotationParam(ann, paramName)` | 获取注解的参数值 |

### 注解数据结构

注解以 map 形式返回：

```longlang
{
    "name": "Entity",
    "arguments": {
        "table": "users"
    }
}
```

## 内置注解

| 注解 | 说明 |
|------|------|
| `@Deprecated` | 标记废弃的类、方法或字段 |

## 元注解

元注解用于修饰注解定义本身（计划中）：

- `@Target` - 限制注解可应用的目标
- `@Retention` - 设置注解的保留策略

## 完整示例

```longlang
use System.Reflection.Reflection

// 定义注解
annotation Entity {
    table string = ""
}

annotation Column {
    name string = ""
    nullable bool = true
}

// 使用注解
@Entity(table: "users")
public class User {
    @Column(name: "id", nullable: false)
    private id int
    
    @Column(name: "name")
    private name string
    
    public function __construct(id: int, name: string) {
        this.id = id
        this.name = name
    }
    
    public function getId() int {
        return this.id
    }
    
    public function getName() string {
        return this.name
    }
}

// 使用反射获取注解
if Reflection::hasClassAnnotation("User", "Entity") {
    entity := Reflection::getClassAnnotation("User", "Entity")
    table := Reflection::getAnnotationParam(entity, "table")
    fmt.println("User 映射到表: " + table)
}
```

## 用例

### ORM 映射

```longlang
annotation Table {
    name string
}

annotation Column {
    name string = ""
    type string = "varchar"
    primaryKey bool = false
}

@Table(name: "products")
class Product {
    @Column(primaryKey: true)
    private id int
    
    @Column(name: "product_name", type: "varchar")
    private name string
}
```

### API 路由

```longlang
annotation Route {
    path string
    method string = "GET"
}

class UserController {
    @Route(path: "/users", method: "GET")
    public function list() {
        // ...
    }
    
    @Route(path: "/users", method: "POST")
    public function create() {
        // ...
    }
}
```

### 验证

```longlang
annotation NotNull {
    message string = "不能为空"
}

annotation Range {
    min int = 0
    max int = 100
}

class Form {
    @NotNull
    private name string
    
    @Range(min: 1, max: 150)
    private age int
}
```



