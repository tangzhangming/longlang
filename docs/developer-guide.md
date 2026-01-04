# LongLang 开发者指南

本文档面向 LongLang 语言的开发者，介绍解释器的内部架构、语法检查执行流程，以及如何修复 BUG 和添加新特性。

## 目录

1. [架构概览](#架构概览)
2. [语法检查执行流程](#语法检查执行流程)
3. [如何修复语法 BUG](#如何修复语法-bug)
4. [如何添加新语法特性](#如何添加新语法特性)

---

## 架构概览

LongLang 解释器采用经典的三阶段架构：

```
源代码 (.long 文件)
    │
    ▼
┌─────────────────┐
│   词法分析器     │  internal/lexer/
│   (Lexer)       │  - token.go    (Token 定义)
│                 │  - lexer.go    (词法分析逻辑)
└────────┬────────┘
         │ Token 流
         ▼
┌─────────────────┐
│   语法分析器     │  internal/parser/
│   (Parser)      │  - ast.go      (AST 节点定义)
│                 │  - parser.go   (语法分析逻辑)
└────────┬────────┘
         │ AST (抽象语法树)
         ▼
┌─────────────────┐
│   解释器        │  internal/interpreter/
│   (Interpreter) │  - object.go   (运行时对象)
│                 │  - interpreter.go (执行逻辑)
│                 │  - environment.go (环境/作用域)
└────────┬────────┘
         │
         ▼
      执行结果
```

### 核心文件说明

| 文件 | 作用 |
|------|------|
| `internal/lexer/token.go` | 定义所有 Token 类型和关键字映射 |
| `internal/lexer/lexer.go` | 将源代码转换为 Token 流 |
| `internal/parser/ast.go` | 定义所有 AST 节点结构 |
| `internal/parser/parser.go` | 将 Token 流解析为 AST |
| `internal/interpreter/object.go` | 定义运行时对象（Integer, String, Function 等）|
| `internal/interpreter/interpreter.go` | 遍历 AST 并执行 |
| `internal/interpreter/environment.go` | 管理变量作用域 |

---

## 语法检查执行流程

以一段简单代码为例：

```longlang
x := 10 + 5
fmt.Println(x)
```

### 阶段 1：词法分析 (Lexer)

**文件**: `internal/lexer/lexer.go`

词法分析器将源代码字符串分解为 Token 序列：

```
源代码: "x := 10 + 5"
    │
    ▼
Token 序列:
  [IDENT:"x"] [ASSIGN:":="] [INT:"10"] [PLUS:"+"] [INT:"5"]
```

**关键函数**:
- `NextToken()` - 读取下一个 Token
- `readIdentifier()` - 读取标识符
- `readNumber()` - 读取数字（整数/浮点数）
- `readString()` - 读取字符串
- `skipLineComment()` / `skipBlockComment()` - 跳过注释

**调试技巧**:
```go
// 在 lexer.go 的 NextToken() 中添加调试输出
fmt.Printf("Token: %s, Literal: %s, Line: %d, Col: %d\n", 
    tok.Type, tok.Literal, tok.Line, tok.Column)
```

### 阶段 2：语法分析 (Parser)

**文件**: `internal/parser/parser.go`

语法分析器将 Token 序列构建为 AST：

```
Token 序列: [IDENT:"x"] [ASSIGN:":="] [INT:"10"] [PLUS:"+"] [INT:"5"]
    │
    ▼
AST:
  AssignStatement
  ├── Name: Identifier("x")
  └── Value: InfixExpression
              ├── Left: IntegerLiteral(10)
              ├── Operator: "+"
              └── Right: IntegerLiteral(5)
```

**关键函数**:
- `ParseProgram()` - 解析整个程序
- `parseStatement()` - 解析语句（根据 Token 类型分发）
- `parseExpression(precedence)` - Pratt 解析器核心，解析表达式
- `parseXxxStatement()` - 解析特定语句（if, for, return 等）
- `parseXxxExpression()` - 解析特定表达式

**Pratt 解析器原理**:
```go
// 前缀解析函数 - 处理表达式开头（如: -5, !true, 标识符）
prefixParseFns map[lexer.TokenType]prefixParseFn

// 中缀解析函数 - 处理二元运算（如: a + b, a * b）
infixParseFns map[lexer.TokenType]infixParseFn

// 优先级表 - 控制运算符优先级
precedences map[lexer.TokenType]int
```

**调试技巧**:
```go
// 在 parseExpression() 中添加调试输出
fmt.Printf("Parsing expression, curToken: %s, peekToken: %s\n", 
    p.curToken.Type, p.peekToken.Type)
```

### 阶段 3：解释执行 (Interpreter)

**文件**: `internal/interpreter/interpreter.go`

解释器遍历 AST 并执行：

```
AST: AssignStatement(x = 10 + 5)
    │
    ▼
执行步骤:
  1. 计算 InfixExpression: 10 + 5 = 15
  2. 创建 Integer 对象: &Integer{Value: 15}
  3. 存入环境: env.Set("x", &Integer{Value: 15})
```

**关键函数**:
- `Eval(node)` - 核心分发函数，根据 AST 节点类型执行
- `evalXxxStatement()` - 执行特定语句
- `evalXxxExpression()` - 计算特定表达式
- `applyFunction()` - 执行函数调用

**调试技巧**:
```go
// 在 Eval() 中添加调试输出
fmt.Printf("Evaluating: %T\n", node)
```

---

## 如何修复语法 BUG

### 步骤 1：复现问题

创建最小测试用例：

```longlang
// test/bug_example.long
package main

fn main() {
    // 触发 BUG 的最小代码
}
```

运行并记录错误信息：
```bash
.\longlang.exe run .\test\bug_example.long
```

### 步骤 2：定位问题阶段

根据错误类型判断问题所在阶段：

| 错误类型 | 可能的阶段 | 相关文件 |
|----------|------------|----------|
| `非法字符` | 词法分析 | `lexer.go` |
| `没有找到 XXX 的前缀解析函数` | 语法分析 | `parser.go` |
| `期望下一个 token 是 XXX` | 语法分析 | `parser.go` |
| `未定义的标识符` | 解释执行 | `interpreter.go` |
| `类型不匹配` | 解释执行 | `interpreter.go` |
| `未知运算符` | 解释执行 | `interpreter.go` |

### 步骤 3：添加调试输出

**词法分析调试**:
```go
// lexer.go - NextToken()
func (l *Lexer) NextToken() Token {
    // ... 
    fmt.Printf("[Lexer] Token: %s, Literal: %q, Line: %d, Col: %d\n",
        tok.Type, tok.Literal, tok.Line, tok.Column)
    return tok
}
```

**语法分析调试**:
```go
// parser.go - parseStatement()
func (p *Parser) parseStatement() Statement {
    fmt.Printf("[Parser] curToken: %s, peekToken: %s, Line: %d\n",
        p.curToken.Type, p.peekToken.Type, p.curToken.Line)
    // ...
}
```

**解释器调试**:
```go
// interpreter.go - Eval()
func (i *Interpreter) Eval(node parser.Node) Object {
    fmt.Printf("[Eval] Node type: %T\n", node)
    // ...
}
```

### 步骤 4：修复并验证

1. 修改相关代码
2. 重新编译：`go build -o longlang.exe .`
3. 运行测试用例验证
4. 运行所有相关测试确保没有回归

### 常见 BUG 模式

**1. Token 未注册**
```go
// lexer/token.go - 添加新 Token 类型
const (
    NEW_TOKEN TokenType = "NEW_TOKEN"
)

// 添加到关键字映射
var keywords = map[string]TokenType{
    "newkeyword": NEW_TOKEN,
}
```

**2. 前缀/中缀解析函数未注册**
```go
// parser.go - New() 函数中注册
p.registerPrefix(lexer.NEW_TOKEN, p.parseNewExpression)
p.registerInfix(lexer.NEW_TOKEN, p.parseNewInfixExpression)
```

**3. AST 节点未处理**
```go
// interpreter.go - Eval() 中添加 case
case *parser.NewStatement:
    return i.evalNewStatement(node)
```

---

## 如何添加新语法特性

以添加 `while` 循环为例（假设还没有实现）：

### 步骤 1：定义 Token（如需要）

**文件**: `internal/lexer/token.go`

```go
// 1. 添加 Token 类型常量
const (
    // ...
    WHILE TokenType = "WHILE"
)

// 2. 添加到关键字映射
var keywords = map[string]TokenType{
    // ...
    "while": WHILE,
}
```

### 步骤 2：定义 AST 节点

**文件**: `internal/parser/ast.go`

```go
// WhileStatement 表示 while 循环
type WhileStatement struct {
    Token     lexer.Token     // 'while' token
    Condition Expression      // 循环条件
    Body      *BlockStatement // 循环体
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
    return "while " + ws.Condition.String() + " " + ws.Body.String()
}
```

### 步骤 3：实现解析逻辑

**文件**: `internal/parser/parser.go`

```go
// 1. 在 parseStatement() 中添加 case
func (p *Parser) parseStatement() Statement {
    switch p.curToken.Type {
    // ...
    case lexer.WHILE:
        return p.parseWhileStatement()
    // ...
    }
}

// 2. 实现解析函数
func (p *Parser) parseWhileStatement() *WhileStatement {
    stmt := &WhileStatement{Token: p.curToken}
    
    p.nextToken() // 跳过 'while'
    
    // 解析条件
    stmt.Condition = p.parseExpression(LOWEST)
    
    // 期望 {
    if !p.expectPeek(lexer.LBRACE) {
        return nil
    }
    
    // 解析循环体
    stmt.Body = p.parseBlockStatement()
    
    return stmt
}
```

### 步骤 4：实现执行逻辑

**文件**: `internal/interpreter/interpreter.go`

```go
// 1. 在 Eval() 中添加 case
func (i *Interpreter) Eval(node parser.Node) Object {
    switch node := node.(type) {
    // ...
    case *parser.WhileStatement:
        return i.evalWhileStatement(node)
    // ...
    }
}

// 2. 实现执行函数
func (i *Interpreter) evalWhileStatement(node *parser.WhileStatement) Object {
    for {
        // 计算条件
        condition := i.Eval(node.Condition)
        if isError(condition) {
            return condition
        }
        
        // 条件为假则退出
        if !isTruthy(condition) {
            break
        }
        
        // 执行循环体
        result := i.Eval(node.Body)
        
        // 处理 break/continue/return
        if result != nil {
            switch result.Type() {
            case BREAK_SIGNAL_OBJ:
                return nil
            case CONTINUE_SIGNAL_OBJ:
                continue
            case RETURN_VALUE_OBJ, ERROR_OBJ:
                return result
            }
        }
    }
    return nil
}
```

### 步骤 5：编写测试

**文件**: `test/test_while.long`

```longlang
package main

fn main() {
    // 测试基本 while
    i := 0
    while i < 5 {
        fmt.Println("i =", i)
        i++
    }
    
    // 测试 break
    j := 0
    while true {
        if j >= 3 {
            break
        }
        fmt.Println("j =", j)
        j++
    }
}
```

### 步骤 6：更新文档

**文件**: `docs/control-structures.md`

添加新语法的说明和示例。

---

## 调试工具清单

### 编译和运行

```bash
# 编译
go build -o longlang.exe .

# 运行测试文件
.\longlang.exe run .\test\xxx.long

# 检查 Go 语法错误
go build ./...
```

### 常用 grep 搜索

```bash
# 搜索 Token 定义
grep -n "TokenType" internal/lexer/token.go

# 搜索解析函数
grep -n "func (p \*Parser) parse" internal/parser/parser.go

# 搜索 Eval case
grep -n "case \*parser\." internal/interpreter/interpreter.go

# 搜索错误信息
grep -rn "错误信息关键字" internal/
```

### 单元测试（如有）

```bash
go test ./internal/lexer/
go test ./internal/parser/
go test ./internal/interpreter/
```

---

## 代码规范

1. **错误信息**：包含行号和列号 `(行 %d, 列 %d)`
2. **注释**：为每个公开函数添加注释说明
3. **命名**：
   - Token 类型：全大写 `NEW_TOKEN`
   - AST 节点：大驼峰 `WhileStatement`
   - 解析函数：`parseXxxStatement` / `parseXxxExpression`
   - 执行函数：`evalXxxStatement` / `evalXxxExpression`

---

## 快速参考

### 添加新关键字

1. `token.go`: 添加 `TokenType` 常量
2. `token.go`: 添加到 `keywords` 映射
3. `parser.go`: 在 `parseStatement()` 或注册前缀/中缀函数
4. `interpreter.go`: 在 `Eval()` 添加执行逻辑

### 添加新运算符

1. `token.go`: 添加 `TokenType` 常量
2. `lexer.go`: 在 `NextToken()` 中识别运算符
3. `parser.go`: 注册中缀解析函数，设置优先级
4. `interpreter.go`: 在 `evalInfixExpression()` 处理

### 添加新类型

1. `token.go`: 添加类型关键字
2. `object.go`: 添加运行时对象类型
3. `parser.go`: 更新类型检查列表
4. `interpreter.go`: 处理类型相关运算












