# LongLang Database ORM

一个类似 Laravel Eloquent 的 ORM 框架，支持注解定义模型和链式查询。

## 目录结构

```
Database/
├── Config/                     # 配置类
│   ├── ConnectionConfig.long   # 单个数据库连接配置
│   ├── DatabaseConfig.long     # 数据库总配置
│   └── PoolConfig.long         # 连接池配置
├── Core/                       # 核心组件
│   ├── Connection.long         # 数据库连接基类
│   ├── ConnectionPool.long     # 连接池实现
│   └── QueryBuilder.long       # SQL 查询构建器
├── Drivers/                    # 数据库驱动
│   └── MysqlConnection.long    # MySQL 驱动实现
├── Exception/                  # 异常类
│   └── DatabaseException.long  # 数据库异常
├── ORM/                        # ORM 组件
│   ├── Annotations.long        # ORM 注解定义
│   ├── Model.long              # 模型基类
│   └── ModelBuilder.long       # 模型查询构建器
├── DatabaseManager.long        # 数据库管理器
└── docs/
    ├── README.md               # 本文档
    └── QueryBuilder.md         # 查询构建器详细文档
```

## 快速开始

### 1. 初始化数据库连接

```longlang
use App.Database.Config.DatabaseConfig
use App.Database.DatabaseManager
use App.Database.ORM.Model

// 创建配置
config := new DatabaseConfig()
config.setHost("127.0.0.1")
      .setPort(3306)
      .setUsername("root")
      .setPassword("password")
      .setDatabase("mydb")
      .setCharset("utf8mb4")

// 创建管理器并设置连接
manager := new DatabaseManager(config)
Model::setConnection(manager)
```

### 2. 定义模型（使用注解）

```longlang
namespace App.Models

use App.Database.ORM.Model

@Table(name: "users")
public class User extends Model {
    
    @Id
    public id int
    
    @Column
    @Fillable
    public name string
    
    @Column(name: "email_address")  // 自定义列名
    @Fillable
    public email string
    
    @Column
    @Fillable
    public age int
    
    @Column
    @Fillable
    public status string
    
    @Column
    @Hidden  // 序列化时隐藏
    public password string
    
    @Column(name: "created_at")
    @CreatedAt
    public createdAt string
    
    public function __construct() {
        super::__construct()
        this.status = "active"
    }
    
    // 自定义方法
    public function isAdult() bool {
        return parseInt(toString(this.age)) >= 18
    }
}
```

### 3. 可用注解

| 注解 | 说明 | 参数 |
|------|------|------|
| `@Table` | 定义表名 | `name: string` |
| `@Id` | 标记主键字段 | - |
| `@Column` | 标记数据库列 | `name: string` (可选，默认使用字段名) |
| `@Fillable` | 允许批量赋值 | - |
| `@Hidden` | 序列化时隐藏 | - |
| `@CreatedAt` | 自动设置创建时间 | - |
| `@UpdatedAt` | 自动更新修改时间 | - |

### 4. 使用模型

#### 查询

```longlang
// 按 ID 查找
user := User::query().find(1)

// 条件查询
adults := User::query().where("age", ">=", 18).get()

// 获取所有
allUsers := User::query().all()

// 链式查询
users := User::query()
    .where("status", "active")
    .orderBy("age", "desc")
    .limit(10)
    .get()

// 获取第一条
user := User::query().where("email", "test@example.com").first()

// 检查是否存在
exists := User::query().where("email", "test@example.com").exists()

// 聚合函数
count := User::query().where("status", "active").count()
maxAge := User::query().max("age")
avgAge := User::query().avg("age")
```

#### 创建

```longlang
// 使用 User::create() 静态方法
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
// 获取后修改保存
user := User::query().find(1)
user.name = "Updated Name"
user.save()

// 批量更新
User::query()
    .where("status", "inactive")
    .update(map[string]any{"status": "deleted"})

// 自增/自减
User::query().where("id", 1).increment("views")
User::query().where("id", 1).decrement("stock", 5)
```

#### 删除

```longlang
// 删除单个模型
user := User::query().find(1)
user.delete()

// 批量删除
User::query().where("status", "deleted").delete()
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
user := User::query().find(1)

// 转为 Map（会排除 @Hidden 字段）
data := user.toArray()

// 转为 JSON
json := user.toJson()
```

## 脏检查

```longlang
user := User::query().find(1)
Console::writeLine(user.isDirty())  // false

user.name = "New Name"
Console::writeLine(user.isDirty())  // true

dirty := user.getDirty()  // {"name": "New Name"}
```

## 查询方法速查

详细文档请参考 [QueryBuilder.md](./QueryBuilder.md)

### WHERE 条件

```longlang
.where("field", "value")                // 等于
.where("field", ">=", 18)               // 运算符
.orWhere("field", "value")              // OR 条件
.whereIn("id", {1, 2, 3})               // IN
.whereNotIn("status", {"a", "b"})       // NOT IN
.whereNull("deleted_at")                // IS NULL
.whereNotNull("email")                  // IS NOT NULL
.whereBetween("age", {18, 30})          // BETWEEN
.whereLike("name", "%test%")            // LIKE
.whereAny({"name", "email"}, "like", "%test%")  // 任一列匹配
```

### 排序与分页

```longlang
.orderBy("created_at", "desc")          // 排序
.orderByDesc("id")                      // 降序
.latest()                               // 按 created_at 降序
.limit(10)                              // 限制数量
.offset(20)                             // 偏移
.forPage(2, 15)                         // 分页
```

### 聚合

```longlang
.count()                                // 计数
.max("age")                             // 最大值
.min("price")                           // 最小值
.avg("score")                           // 平均值
.sum("amount")                          // 总和
.exists()                               // 是否存在
```

### 获取结果

```longlang
.get()                                  // 获取所有
.first()                                // 获取第一条
.find(1)                                // 按主键查找
.value("name")                          // 获取单个值
.pluck("name")                          // 获取单列数组
.pluck("name", "id")                    // 获取键值对
```

### Raw 原生查询

```longlang
.whereRaw("price > ?", {100})           // 原生 WHERE
.selectRaw("price * ? as total", {1.1}) // 原生 SELECT
.orderByRaw("FIELD(status, 'a', 'b')")  // 原生 ORDER BY
.havingRaw("SUM(price) > ?", {1000})    // 原生 HAVING
```

## 完整示例

```longlang
namespace App

use System.Console
use App.Database.Config.DatabaseConfig
use App.Database.DatabaseManager
use App.Database.ORM.Model
use App.Models.User

public class Main {
    public static function main() {
        // 初始化
        config := new DatabaseConfig()
        config.setHost("127.0.0.1")
              .setPort(3306)
              .setUsername("root")
              .setPassword("root")
              .setDatabase("mydb")
              .setCharset("utf8mb4")
        
        manager := new DatabaseManager(config)
        Model::setConnection(manager)
        
        // 创建用户
        user := User::create(map[string]any{
            "name": "Alice",
            "email": "alice@example.com",
            "age": 25
        })
        Console::writeLine("Created user: " + user.name)
        
        // 查询用户
        adults := User::query()
            .where("age", ">=", 18)
            .where("status", "active")
            .orderBy("name")
            .get()
        
        Console::writeLine("Adult active users: " + toString(len(adults)))
        
        for i := 0; i < len(adults); i++ {
            u := adults[i]
            Console::writeLine("  - " + u.name + " (age: " + toString(u.age) + ")")
        }
        
        // 统计
        count := User::query().count()
        avgAge := User::query().avg("age")
        Console::writeLine("Total: " + toString(count) + ", Avg age: " + toString(avgAge))
    }
}
```

