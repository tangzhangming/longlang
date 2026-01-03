# LongLang Database ORM

一个类似 Laravel Eloquent 的 ORM 框架，支持简洁的静态方法语法。

## 目录结构

```
Database/
├── Config/              # 配置类
│   ├── ConnectionConfig.long    # 单个数据库连接配置
│   ├── DatabaseConfig.long      # 数据库总配置
│   └── PoolConfig.long          # 连接池配置
├── Core/                # 核心组件
│   ├── Connection.long          # 数据库连接基类
│   ├── ConnectionPool.long      # 连接池实现
│   └── QueryBuilder.long        # SQL 查询构建器
├── Drivers/             # 数据库驱动
│   └── MysqlConnection.long     # MySQL 驱动实现
├── Exception/           # 异常类
│   └── DatabaseException.long   # 数据库异常
└── ORM/                 # ORM 组件
    ├── DB.long                  # 数据库门面类
    ├── Model.long               # 模型基类
    └── ModelBuilder.long        # 模型查询构建器
```

## 快速开始

### 1. 初始化数据库

```longlang
use App.Database.ORM.DB
use App.Database.Config.DatabaseConfig
use App.Database.Config.ConnectionConfig
use App.Database.Config.PoolConfig

// 创建连接配置
poolConfig := new PoolConfig()
poolConfig.setMinSize(1).setMaxSize(10)

connConfig := new ConnectionConfig()
connConfig.setHost("127.0.0.1")
          .setPort(3306)
          .setUsername("root")
          .setPassword("password")
          .setDatabase("mydb")
          .setCharset("utf8mb4")
          .setPoolConfig(poolConfig)

// 创建数据库配置
dbConfig := new DatabaseConfig()
dbConfig.setDefault("mysql").addConnection("mysql", connConfig)

// 初始化 DB 门面
db := new DB()
db.init(dbConfig)
__set_global("__db", db)
```

### 2. 定义模型

```longlang
namespace App.Models

use App.Database.ORM.Model

public class User extends Model {
    public id int
    public name string
    public email string
    public age int
    public status string
    
    public function __construct() {
        super::__construct()
        this.status = "active"
    }
    
    // 自定义方法
    public function isAdult() bool {
        return parseInt(toString(this.age)) >= 18
    }
    
    // 模型元数据
    public static function getMeta() any {
        return map[string]any{
            "table": "users",
            "primaryKey": "id",
            "autoIncrement": true,
            "columns": map[string]any{
                "id": map[string]any{"column": "id", "type": "int"},
                "name": map[string]any{"column": "name", "type": "string"},
                "email": map[string]any{"column": "email", "type": "string"},
                "age": map[string]any{"column": "age", "type": "int"},
                "status": map[string]any{"column": "status", "type": "string"}
            },
            "fillable": {"name", "email", "age", "status"},
            "hidden": {},
            "factory": function() any {
                return new User()
            }
        }
    }
}
```

### 3. 注册模型

```longlang
db.registerModel("User", User::getMeta())
```

### 4. 使用模型

#### 查询

```longlang
// 按 ID 查找
user := User::find(1)

// 条件查询
adults := User::where("age", ">=", 18).get()

// 获取所有
allUsers := User::all()

// 链式查询
users := User::where("status", "active")
              .orderBy("age", "desc")
              .limit(10)
              .get()

// 检查是否存在
exists := User::where("email", "test@example.com").exists()

// 聚合函数
count := User::where("status", "active").count()
maxAge := User::where("status", "active").max("age")
```

#### 创建

```longlang
// 使用 create
user := User::create(map[string]any{
    "name": "Alice",
    "email": "alice@example.com",
    "age": 25
})

// 使用 new + save
user := new User()
user.name = "Bob"
user.email = "bob@example.com"
user.age = 30
user.save()
```

#### 更新

```longlang
user := User::find(1)
user.name = "Updated Name"
user.save()

// 或使用 update
User::where("status", "inactive").update(map[string]any{
    "status": "deleted"
})
```

#### 删除

```longlang
user := User::find(1)
user.delete()

// 或批量删除
User::where("status", "deleted").delete()
```

### 5. 获取类名

```longlang
// 使用 ::class 语法
className := User::class  // 返回 "User"
```

## Late Static Binding

LongLang 支持后期静态绑定，允许在继承链中正确解析调用的类：

```longlang
public class Animal {
    public static function getClassName() string {
        return __called_class_name  // 返回实际调用的类名
    }
}

public class Dog extends Animal {}

Animal::getClassName()  // 返回 "Animal"
Dog::getClassName()     // 返回 "Dog"
```

## 序列化

```longlang
user := User::find(1)

// 转为 Map
data := user.toArray()

// 转为 JSON
json := user.toJson()
```

## 脏检查

```longlang
user := User::find(1)
Console::writeLine(user.isDirty())  // false

user.name = "New Name"
Console::writeLine(user.isDirty())  // true

dirty := user.getDirty()  // {"name": "New Name"}
```


