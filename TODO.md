# LongLang 待完成和已完成的功能

## ✅ 已完成的功能

### 命名空间系统
- ✅ 添加 `NAMESPACE` 和 `USE` 关键字到 lexer
- ✅ 添加 `NamespaceStatement` 和 `UseStatement` AST 节点
- ✅ 实现 parser 解析 `namespace` 和 `use` 语句
- ✅ 实现命名空间存储和查找机制
- ✅ 实现类入口（查找 `main` 静态方法）
- ✅ 修改 `evalClassStatement` 将类注册到命名空间
- ✅ 修复静态方法调用从命名空间查找类
- ✅ 修复 `new` 表达式从命名空间查找类
- ✅ 修复多个 `main` 方法的检测逻辑

### 项目配置
- ✅ 实现 `project.toml` 解析
- ✅ 集成 `project.toml` 到主程序
- ✅ 实现命名空间简化功能（使用 `root_namespace`）

### 模块导入
- ✅ 实现 `use` 语句的文件加载功能
  - 支持从 `src/` 目录加载
  - 支持从项目根目录加载
  - 支持从 `vendor/` 目录加载
  - 自动应用 `root_namespace` 解析

---

## 📋 测试用例

### 命名空间测试
- `test/test_namespace_basic.long` - 基本命名空间测试
- `test/test_namespace_multiple_classes.long` - 多类测试
- `test/test_namespace_simple_error.long` - 多个 main 方法检测
- `test/test_namespace_no_main.long` - 无 main 方法检测
- `test/test_use_import.long` - use 导入测试

### 文件结构
```
test/
├── project.toml              # 项目配置文件
├── src/
│   └── Utils/
│       └── StringHelper.long # 工具类
├── test_namespace_basic.long
├── test_namespace_multiple_classes.long
├── test_use_import.long
└── ...
```

---

## 🔧 使用说明

### project.toml 配置
```toml
[project]
name = "my-project"
version = "1.0.0"
root_namespace = "MyApp"
```

### 命名空间声明
```longlang
namespace Models  // 解析为 MyApp.Models（如果设置了 root_namespace）
```

### 导入类
```longlang
use Utils.StringHelper  // 从 src/Utils/StringHelper.long 加载
use Utils.StringHelper as Helper  // 使用别名
```

### 程序入口
程序入口必须是包含 `static main()` 方法的类：
```longlang
namespace App

class Program {
    public static function main() {
        fmt.println("Hello, World!")
    }
}
```
