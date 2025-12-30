# 包系统设计文档

## 概述

longlang 的包系统设计参考了 Go 语言的包系统，但有一些关键差异：
- 每个 `.long` 文件是一个独立的包
- 使用 `package` 关键字声明包名
- 使用 `import` 关键字导入其他包
- 包路径基于项目根目录的 `long.mod` 文件配置

## 1. long.mod 文件

### 格式

`long.mod` 文件定义项目的基础路径，类似于 Go 的 `go.mod`：

```
module github.com/example/myproject
```

### 位置

`long.mod` 文件必须放在项目根目录。

### 作用

- 定义项目的模块路径
- 用于解析相对导入路径
- 用于确定包的完整路径

## 2. 包声明

### 语法

```long
package packageName
```

### 规则

- 每个 `.long` 文件必须以 `package` 声明开始
- 包名必须符合标识符规则
- 包名通常与文件名相关，但不强制要求
- 一个文件中的所有代码都属于同一个包

### 示例

```long
// /util/string.long
package string

fn test() {
    fmt.Println("test")
}
```

## 3. 导入包

### 语法

```long
import "package/path"
```

### 导入路径规则

1. **绝对路径**：基于 `long.mod` 的 module 路径
   ```long
   import "github.com/example/myproject/util.string"
   ```

2. **相对路径**：相对于当前文件
   ```long
   import "../models.UserModel"
   ```

3. **项目内路径**：基于 `long.mod` 的 module 路径
   ```long
   import "app.models.UserModel"
   import "util.string"
   ```

### 路径到文件映射

导入路径 `"util.string"` 会映射到文件：
- 如果 `long.mod` 是 `module github.com/example/myproject`
- 则查找 `./util/string.long` 或 `./util/string/string.long`

### 导入规则

- 导入的包会被解析并执行一次
- 导入的包中的导出符号（函数、类）可以在当前文件中使用
- 使用 `包名.符号名` 的方式访问

## 4. 包的导出规则

### 导出符号

以下符号可以被其他包访问：

1. **函数**：所有顶层函数都可以被导出
   ```long
   package string
   
   fn test() {  // 可以被其他包通过 string.test() 访问
       // ...
   }
   ```

2. **类**：只有与文件名相同的类可以被导出
   ```long
   // /app/models/UserModel.long
   package UserModel
   
   class UserModel {  // 可以被导出
       // ...
   }
   
   class Helper {  // 不能被导出（类名与文件名不一致）
       // ...
   }
   ```

3. **变量**：顶层变量（当前版本暂不支持导出）

## 5. 使用导入的包

### 调用函数

```long
import "util.string"

fn main() {
    string.test()  // 调用 util.string 包中的 test 函数
}
```

### 使用类

```long
import "app.models.UserModel"

fn main() {
    // 创建对象
    user := new UserModel()
    
    // 调用实例方法
    name := user.getName()
    
    // 调用静态方法
    tableName := UserModel::getTableName()
}
```

## 6. 包解析流程

1. **读取 long.mod**：获取项目基础路径
2. **解析 import 语句**：提取导入路径
3. **路径解析**：将导入路径转换为文件路径
4. **文件查找**：在文件系统中查找对应的 `.long` 文件
5. **解析包**：对找到的文件进行词法分析和语法分析
6. **执行包**：执行包的顶层代码（函数定义、类定义等）
7. **注册符号**：将导出的符号注册到包命名空间中
8. **返回包对象**：返回包对象供当前文件使用

## 7. 包缓存

为了避免重复解析同一个包，实现包缓存机制：

- 每个包只解析一次
- 解析后的包对象缓存在内存中
- 以包的完整路径作为缓存键

## 8. 循环依赖处理

当前版本暂不支持循环依赖检测，如果出现循环依赖会导致无限递归。

未来版本需要实现：
- 循环依赖检测
- 错误提示
- 依赖图分析

## 9. 实现要点

### 9.1 包管理器

创建 `internal/pkg` 包，包含：
- `Package` 结构：表示一个包
- `PackageManager`：管理包的加载和缓存
- `ImportResolver`：解析导入路径

### 9.2 AST 扩展

在 `parser` 中添加：
- `PackageStatement`：包声明语句
- `ImportStatement`：导入语句

### 9.3 解释器扩展

在 `interpreter` 中添加：
- 包环境管理
- 包符号查找
- 跨包函数调用
- 跨包类实例化

### 9.4 文件系统

- 读取 `long.mod` 文件
- 根据导入路径查找文件
- 支持相对路径和绝对路径

## 10. 示例

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
    private name string
    
    public function __construct(name:string) {
        this.name = name
    }
    
    public function getName() string {
        return this.name
    }
    
    public static function getTableName() string {
        return "users"
    }
}
```

### main.long

```long
package main

import "util.string"
import "app.models.UserModel"

fn main() {
    string.test()
    
    user := new UserModel("John")
    fmt.Println(user.getName())
    fmt.Println(UserModel::getTableName())
}
```

