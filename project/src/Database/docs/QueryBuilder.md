# 查询构造器 (Query Builder)

LongLang 的查询构造器提供了便捷、流畅的接口来创建和执行数据库查询。语法设计与 Laravel 保持一致，方便 Laravel 开发者快速上手。

## 目录

- [基本用法](#基本用法)
- [方法一览表](#方法一览表)
- [WHERE 条件](#where-条件)
- [闭包分组](#闭包分组)
- [Raw 原生查询](#raw-原生查询)
- [排序与分页](#排序与分页)
- [聚合函数](#聚合函数)
- [写操作](#写操作)

---

## 基本用法

```longlang
use App.Database.DatabaseManager
use App.Database.Config.DatabaseConfig
use App.Database.ORM.Model
use App.Models.User

// 初始化数据库连接
config := new DatabaseConfig()
config.setHost("127.0.0.1")
      .setPort(3306)
      .setUsername("root")
      .setPassword("password")
      .setDatabase("mydb")

manager := new DatabaseManager(config)
Model::setConnection(manager)

// 查询
users := User::query().where("status", "active").get()
user := User::query().find(1)
```

---

## 方法一览表

### WHERE 条件方法

| 方法 | 说明 | 示例 |
|------|------|------|
| `where(column, value)` | 等于条件 | `where("name", "test")` |
| `where(column, op, value)` | 运算符条件 | `where("age", ">=", 18)` |
| `orWhere(column, value)` | OR 等于条件 | `orWhere("status", "pending")` |
| `orWhere(column, op, value)` | OR 运算符条件 | `orWhere("age", "<", 10)` |
| `whereNot(column, value)` | 不等于条件 | `whereNot("status", "deleted")` |
| `whereNot(column, op, value)` | NOT 运算符条件 | `whereNot("age", ">", 60)` |
| `orWhereNot(column, value)` | OR NOT 条件 | `orWhereNot("type", "admin")` |
| `whereIn(column, values)` | IN 条件 | `whereIn("id", {1, 2, 3})` |
| `whereNotIn(column, values)` | NOT IN 条件 | `whereNotIn("status", {"banned", "deleted"})` |
| `orWhereIn(column, values)` | OR IN 条件 | `orWhereIn("type", {"a", "b"})` |
| `orWhereNotIn(column, values)` | OR NOT IN 条件 | `orWhereNotIn("id", {1, 2})` |
| `whereNull(column)` | IS NULL 条件 | `whereNull("deleted_at")` |
| `whereNotNull(column)` | IS NOT NULL 条件 | `whereNotNull("email")` |
| `orWhereNull(column)` | OR IS NULL 条件 | `orWhereNull("phone")` |
| `orWhereNotNull(column)` | OR IS NOT NULL 条件 | `orWhereNotNull("address")` |
| `whereBetween(column, values)` | BETWEEN 条件 | `whereBetween("age", {18, 30})` |
| `whereNotBetween(column, values)` | NOT BETWEEN 条件 | `whereNotBetween("price", {100, 200})` |
| `orWhereBetween(column, values)` | OR BETWEEN 条件 | `orWhereBetween("score", {80, 100})` |
| `orWhereNotBetween(column, values)` | OR NOT BETWEEN | `orWhereNotBetween("qty", {0, 10})` |
| `whereLike(column, value)` | LIKE 条件 | `whereLike("name", "%test%")` |
| `whereNotLike(column, value)` | NOT LIKE 条件 | `whereNotLike("email", "%spam%")` |
| `orWhereLike(column, value)` | OR LIKE 条件 | `orWhereLike("title", "%hot%")` |
| `orWhereNotLike(column, value)` | OR NOT LIKE | `orWhereNotLike("desc", "%ad%")` |
| `whereColumn(first, second)` | 列比较（等于） | `whereColumn("created_at", "updated_at")` |
| `whereColumn(first, op, second)` | 列比较（运算符） | `whereColumn("qty", ">", "min_qty")` |
| `whereAny(columns, op, value)` | 任一列匹配 | `whereAny({"name", "email"}, "like", "%test%")` |
| `whereAll(columns, op, value)` | 所有列匹配 | `whereAll({"name", "nick"}, "like", "%john%")` |
| `whereNone(columns, op, value)` | 无列匹配 | `whereNone({"name", "email"}, "like", "%spam%")` |
| `whereRaw(sql, bindings)` | 原生 WHERE | `whereRaw("price > ?", {100})` |
| `orWhereRaw(sql, bindings)` | 原生 OR WHERE | `orWhereRaw("score > ?", {90})` |

### SELECT 方法

| 方法 | 说明 | 示例 |
|------|------|------|
| `select(columns)` | 指定查询字段 | `select({"id", "name", "email"})` |
| `selectRaw(expr, bindings)` | 原生 SELECT | `selectRaw("price * ? as total", {1.1})` |
| `distinct()` | 去重 | `distinct()` |

### 排序方法

| 方法 | 说明 | 示例 |
|------|------|------|
| `orderBy(column, dir)` | 排序（默认 asc） | `orderBy("created_at", "desc")` |
| `orderByDesc(column)` | 降序排序 | `orderByDesc("id")` |
| `orderByRaw(sql, bindings)` | 原生排序 | `orderByRaw("FIELD(status, 'active', 'pending')")` |
| `latest(column)` | 按时间降序（默认 created_at） | `latest()` |
| `oldest(column)` | 按时间升序（默认 created_at） | `oldest("updated_at")` |

### 分页方法

| 方法 | 说明 | 示例 |
|------|------|------|
| `limit(n)` | 限制数量 | `limit(10)` |
| `take(n)` | 限制数量（别名） | `take(10)` |
| `offset(n)` | 偏移量 | `offset(20)` |
| `skip(n)` | 偏移量（别名） | `skip(20)` |
| `forPage(page, perPage)` | 分页（默认每页15条） | `forPage(2, 20)` |

### 分组方法

| 方法 | 说明 | 示例 |
|------|------|------|
| `groupBy(columns)` | 分组 | `groupBy("status")` |
| `having(column, op, value)` | HAVING 条件 | `having("count", ">", 5)` |
| `havingRaw(sql, bindings)` | 原生 HAVING | `havingRaw("SUM(price) > ?", {1000})` |

### 查询执行方法

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `get()` | 获取所有结果 | `[]Model` |
| `first()` | 获取第一条 | `Model` 或 `null` |
| `firstOrFail()` | 获取第一条，不存在则异常 | `Model` |
| `find(id)` | 按主键查找 | `Model` 或 `null` |
| `findOrFail(id)` | 按主键查找，不存在则异常 | `Model` |
| `value(column)` | 获取第一行某列值 | `any` |
| `pluck(column)` | 获取单列数组 | `[]any` |
| `pluck(column, key)` | 获取键值对 | `map[string]any` |
| `exists()` | 检查是否存在 | `bool` |
| `count(column)` | 计数（默认 *） | `int` |
| `max(column)` | 最大值 | `any` |
| `min(column)` | 最小值 | `any` |
| `avg(column)` | 平均值 | `any` |
| `sum(column)` | 总和 | `any` |

### 写操作方法

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `insert(data)` | 插入数据 | `bool` |
| `insertGetId(data)` | 插入并返回 ID | `int` |
| `update(data)` | 更新数据 | `int` (影响行数) |
| `delete()` | 删除数据 | `int` (影响行数) |
| `increment(column, amount, extra)` | 自增 | `int` |
| `decrement(column, amount, extra)` | 自减 | `int` |

### 条件执行

| 方法 | 说明 | 示例 |
|------|------|------|
| `when(condition, callback)` | 条件为真时执行回调 | `when(needFilter, fn(q){ q.where(...) })` |

---

## WHERE 条件

### 基本条件

```longlang
// 简单等于
User::query().where("status", "active").get()

// 使用运算符
User::query().where("age", ">=", 18).get()

// 多个 AND 条件
User::query()
    .where("status", "active")
    .where("age", ">=", 18)
    .get()

// OR 条件
User::query()
    .where("status", "active")
    .orWhere("role", "admin")
    .get()
// SQL: WHERE status = 'active' OR role = 'admin'
```

### whereIn / whereNotIn

```longlang
// IN 条件
User::query().whereIn("id", {1, 2, 3}).get()
// SQL: WHERE id IN (1, 2, 3)

// NOT IN 条件
User::query().whereNotIn("status", {"banned", "deleted"}).get()
// SQL: WHERE status NOT IN ('banned', 'deleted')
```

### whereNull / whereNotNull

```longlang
// 查找未删除的记录
User::query().whereNull("deleted_at").get()
// SQL: WHERE deleted_at IS NULL

// 查找有邮箱的用户
User::query().whereNotNull("email").get()
// SQL: WHERE email IS NOT NULL
```

### whereBetween

```longlang
// 年龄在 18-30 之间
User::query().whereBetween("age", {18, 30}).get()
// SQL: WHERE age BETWEEN 18 AND 30

// 不在范围内
User::query().whereNotBetween("price", {100, 200}).get()
// SQL: WHERE price NOT BETWEEN 100 AND 200
```

### whereLike

```longlang
// 模糊搜索
User::query().whereLike("name", "%john%").get()
// SQL: WHERE name LIKE '%john%'
```

### whereAny / whereAll / whereNone

```longlang
// 任一列匹配 (OR)
User::query().whereAny({"name", "email"}, "like", "%test%").get()
// SQL: WHERE (name LIKE '%test%' OR email LIKE '%test%')

// 所有列匹配 (AND)
User::query().whereAll({"name", "nickname"}, "like", "%john%").get()
// SQL: WHERE (name LIKE '%john%' AND nickname LIKE '%john%')

// 无列匹配 (NOT OR)
User::query().whereNone({"name", "email"}, "like", "%spam%").get()
// SQL: WHERE NOT (name LIKE '%spam%' OR email LIKE '%spam%')
```

---

## 闭包分组

使用闭包可以创建复杂的分组条件：

```longlang
// SQL: WHERE votes > 100 OR (name = 'Abigail' AND votes > 50)
User::query()
    .where("votes", ">", 100)
    .orWhere(fn(query){
        query.where("name", "Abigail")
             .where("votes", ">", 50)
    })
    .get()
```

```longlang
// SQL: WHERE status = 'active' AND (role = 'admin' OR role = 'moderator')
User::query()
    .where("status", "active")
    .where(fn(query){
        query.where("role", "admin")
             .orWhere("role", "moderator")
    })
    .get()
```

---

## Raw 原生查询

当需要使用原生 SQL 时，可以使用 Raw 系列方法：

### whereRaw

```longlang
// 原生 WHERE 条件
User::query().whereRaw("age > ? AND status = ?", {18, "active"}).get()

// 与其他条件组合
User::query()
    .where("type", "premium")
    .whereRaw("YEAR(created_at) = ?", {2024})
    .get()
```

### selectRaw

```longlang
// 原生 SELECT
User::query()
    .selectRaw("id, name, price * ? as total_price", {1.1})
    .get()
```

### orderByRaw

```longlang
// 原生排序
User::query()
    .orderByRaw("FIELD(status, 'urgent', 'normal', 'low')")
    .get()
```

### havingRaw

```longlang
// 原生 HAVING
User::query()
    .select({"department_id"})
    .groupBy("department_id")
    .havingRaw("COUNT(*) > ?", {5})
    .get()
```

---

## 排序与分页

### 排序

```longlang
// 单字段排序
User::query().orderBy("created_at", "desc").get()

// 多字段排序
User::query()
    .orderBy("status", "asc")
    .orderBy("created_at", "desc")
    .get()

// 便捷方法
User::query().latest().get()      // ORDER BY created_at DESC
User::query().oldest().get()       // ORDER BY created_at ASC
User::query().latest("updated_at").get()  // ORDER BY updated_at DESC
```

### 分页

```longlang
// 获取前 10 条
User::query().limit(10).get()

// 跳过前 20 条，获取 10 条
User::query().offset(20).limit(10).get()

// 分页（第 2 页，每页 15 条）
User::query().forPage(2, 15).get()
```

---

## 聚合函数

```longlang
// 计数
total := User::query().count()
activeCount := User::query().where("status", "active").count()

// 最大/最小/平均/总和
maxAge := User::query().max("age")
minPrice := Product::query().min("price")
avgScore := User::query().avg("score")
totalAmount := Order::query().sum("amount")

// 检查存在
hasAdmin := User::query().where("role", "admin").exists()

// 获取单值
userName := User::query().where("id", 1).value("name")

// 获取单列
names := User::query().pluck("name")
// 结果: {"Alice", "Bob", "Charlie"}

// 获取键值对
nameById := User::query().pluck("name", "id")
// 结果: {"1": "Alice", "2": "Bob", "3": "Charlie"}
```

---

## 写操作

### 创建记录

```longlang
// 使用 Model::create()（推荐）
user := User::create(map[string]any{
    "name": "Alice",
    "email": "alice@example.com"
})

// 使用实例 save()
user := new User()
user.name = "Bob"
user.email = "bob@example.com"
user.save()
```

### 更新记录

```longlang
// 更新单个模型
user := User::query().find(1)
user.name = "New Name"
user.save()

// 批量更新
User::query()
    .where("status", "pending")
    .update(map[string]any{"status": "active"})
```

### 删除记录

```longlang
// 删除单个模型
user := User::query().find(1)
user.delete()

// 批量删除
User::query().where("status", "banned").delete()
```

### 自增/自减

```longlang
// 自增
User::query().where("id", 1).increment("views")
User::query().where("id", 1).increment("views", 5)

// 自减
User::query().where("id", 1).decrement("stock")
User::query().where("id", 1).decrement("stock", 3)

// 自增同时更新其他字段
User::query().where("id", 1).increment("views", 1, map[string]any{
    "last_viewed_at": "2024-01-01 12:00:00"
})
```

---

## 条件执行

使用 `when` 方法可以根据条件动态添加查询：

```longlang
status := "active"  // 可能为空
role := ""          // 空值

users := User::query()
    .when(status != "", fn(query){
        query.where("status", status)
    })
    .when(role != "", fn(query){
        query.where("role", role)
    })
    .get()
```

---

## 注意事项

1. **链式调用**：所有方法都返回查询构建器实例，支持链式调用
2. **闭包参数**：闭包方法接收的参数是 `QueryBuilder` 实例
3. **绑定参数**：Raw 方法的 `bindings` 参数使用 `?` 占位符
4. **类型安全**：字段值会自动转义，防止 SQL 注入


