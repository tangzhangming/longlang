package parser

import "github.com/tangzhangming/longlang/internal/lexer"

// ========== AST 节点接口 ==========

// Node AST 节点接口
// AST (Abstract Syntax Tree) 是抽象语法树，表示源代码的语法结构
// 所有 AST 节点都必须实现这个接口
type Node interface {
	TokenLiteral() string // 返回节点对应的 token 字面值
	String() string      // 返回节点的字符串表示（用于调试）
}

// Statement 语句接口
// 语句是程序执行的基本单元，不产生值（或产生值但被忽略）
// 例如：变量声明、赋值、if 语句、函数定义等
type Statement interface {
	Node
	statementNode() // 标记方法，用于类型区分
}

// Expression 表达式接口
// 表达式会产生一个值
// 例如：字面量、变量引用、函数调用、运算符表达式等
type Expression interface {
	Node
	expressionNode() // 标记方法，用于类型区分
}

// ========== 程序结构 ==========

// Program 程序根节点
// 表示整个程序，包含所有顶层语句
type Program struct {
	Statements []Statement // 程序中的所有语句
}

// TokenLiteral 返回程序的第一个 token 字面值
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// String 返回程序的字符串表示
func (p *Program) String() string {
	var out string
	for _, s := range p.Statements {
		out += s.String()
	}
	return out
}

// ========== 语句节点 ==========

// LetStatement 变量声明语句
// 对应语法：var name type = value 或 var name = value
// 例如：var x int = 10 或 var y = 20
type LetStatement struct {
	Token lexer.Token  // var 关键字对应的 token
	Name  *Identifier  // 变量名
	Type  *Identifier  // 变量类型（可选，如果为 nil 则表示类型推导）
	Value Expression   // 变量的初始值（可选）
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out string
	out += ls.TokenLiteral() + " "
	out += ls.Name.String()
	if ls.Type != nil {
		out += " " + ls.Type.String()
	}
	if ls.Value != nil {
		out += " = " + ls.Value.String()
	}
	return out
}

// AssignStatement 赋值语句（短变量声明 :=）
// 对应语法：name := value
// 例如：x := 10
type AssignStatement struct {
	Token lexer.Token  // := 对应的 token
	Name  *Identifier  // 变量名
	Value Expression   // 要赋的值
}

func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignStatement) String() string {
	var out string
	out += as.Name.String() + " " + as.TokenLiteral() + " "
	if as.Value != nil {
		out += as.Value.String()
	}
	return out
}

// ReturnStatement 返回语句
// 对应语法：return value
// 例如：return 42
type ReturnStatement struct {
	Token       lexer.Token // return 关键字对应的 token
	ReturnValue Expression  // 返回值表达式
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out string
	out += rs.TokenLiteral() + " "
	if rs.ReturnValue != nil {
		out += rs.ReturnValue.String()
	}
	return out
}

// ExpressionStatement 表达式语句
// 将表达式作为语句使用（表达式的值被忽略）
// 例如：x + y; 或 add(1, 2);
type ExpressionStatement struct {
	Token      lexer.Token // 表达式对应的 token
	Expression Expression  // 表达式
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// BlockStatement 块语句
// 对应语法：{ statement1; statement2; ... }
// 用于函数体、if 语句体等
type BlockStatement struct {
	Token      lexer.Token  // { 对应的 token
	Statements []Statement  // 块内的所有语句
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out string
	for _, s := range bs.Statements {
		out += s.String()
	}
	return out
}

// IfStatement if 语句
// 对应语法：if condition { ... } else { ... } 或 if condition { ... } else if { ... }
type IfStatement struct {
	Token       lexer.Token    // if 关键字对应的 token
	Condition   Expression      // 条件表达式
	Consequence *BlockStatement // if 分支的语句块
	Alternative *BlockStatement // else 分支的语句块（如果存在）
	ElseIf      *IfStatement    // else if 链（如果存在）
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out string
	out += "if " + is.Condition.String() + " " + is.Consequence.String()
	if is.Alternative != nil {
		out += " else " + is.Alternative.String()
	}
	if is.ElseIf != nil {
		out += " else " + is.ElseIf.String()
	}
	return out
}

// ForStatement for 循环语句
// 支持三种形式：
// 1. for condition { ... }         - while 式循环
// 2. for { ... }                   - 无限循环
// 3. for init; condition; post { ... } - 传统 for 循环
type ForStatement struct {
	Token     lexer.Token     // for 关键字对应的 token
	Init      Statement       // 初始化语句（可选）
	Condition Expression      // 条件表达式（可选，nil 表示无限循环）
	Post      Statement       // 循环后执行的语句（可选，如 i++）
	Body      *BlockStatement // 循环体
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out string
	out += "for "
	if fs.Init != nil {
		out += fs.Init.String() + "; "
	}
	if fs.Condition != nil {
		out += fs.Condition.String()
	}
	if fs.Post != nil {
		out += "; " + fs.Post.String()
	}
	out += " " + fs.Body.String()
	return out
}

// BreakStatement break 语句
// 用于跳出 for 循环
type BreakStatement struct {
	Token lexer.Token // break 关键字对应的 token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return "break" }

// ContinueStatement continue 语句
// 用于跳过当前循环迭代，继续下一次迭代
type ContinueStatement struct {
	Token lexer.Token // continue 关键字对应的 token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return "continue" }

// IncrementStatement 自增/自减语句
// 对应语法：i++ 或 i--
type IncrementStatement struct {
	Token    lexer.Token // ++ 或 -- 对应的 token
	Name     *Identifier // 要自增/自减的变量
	Operator string      // "++" 或 "--"
}

func (inc *IncrementStatement) statementNode()       {}
func (inc *IncrementStatement) TokenLiteral() string { return inc.Token.Literal }
func (inc *IncrementStatement) String() string       { return inc.Name.String() + inc.Operator }

// ========== 表达式节点 ==========

// FunctionLiteral 函数字面量
// 对应语法：fn name(param1:type1, param2:type2): returnType { ... }
// 例如：fn add(a:int, b:int): int { return a + b }
type FunctionLiteral struct {
	Token      lexer.Token          // fn 关键字对应的 token
	Name       *Identifier          // 函数名
	Parameters []*FunctionParameter // 函数参数列表
	ReturnType []*Identifier       // 返回类型列表（支持多返回值）
	Body       *BlockStatement      // 函数体
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out string
	out += "fn "
	if fl.Name != nil {
		out += fl.Name.String() + " "
	}
	out += "("
	for i, p := range fl.Parameters {
		if i > 0 {
			out += ", "
		}
		out += p.String()
	}
	out += ")"
	if len(fl.ReturnType) > 0 {
		out += ": "
		if len(fl.ReturnType) == 1 {
			out += fl.ReturnType[0].String()
		} else {
			out += "("
			for i, rt := range fl.ReturnType {
				if i > 0 {
					out += ", "
				}
				out += rt.String()
			}
			out += ")"
		}
	}
	out += " " + fl.Body.String()
	return out
}

// FunctionParameter 函数参数
// 对应语法：name:type 或 name:type = defaultValue
// 例如：x:int 或 y:int = 10
type FunctionParameter struct {
	Name         *Identifier // 参数名
	Type         *Identifier // 参数类型（可选）
	DefaultValue Expression  // 默认值（可选）
}

func (fp *FunctionParameter) String() string {
	var out string
	out += fp.Name.String()
	if fp.Type != nil {
		out += ":" + fp.Type.String()
	}
	if fp.DefaultValue != nil {
		out += " = " + fp.DefaultValue.String()
	}
	return out
}

// CallExpression 函数调用表达式
// 对应语法：function(arg1, arg2, ...) 或 function(name1:arg1, name2:arg2, ...)
// 例如：add(1, 2) 或 greet(name:"world")
type CallExpression struct {
	Token     lexer.Token    // ( 对应的 token
	Function  Expression      // 要调用的函数（可以是标识符或表达式）
	Arguments []CallArgument // 参数列表
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out string
	out += ce.Function.String() + "("
	for i, arg := range ce.Arguments {
		if i > 0 {
			out += ", "
		}
		out += arg.String()
	}
	out += ")"
	return out
}

// CallArgument 函数调用参数（支持命名参数）
// 对应语法：value 或 name:value
// 例如：10 或 x:10
type CallArgument struct {
	Name  *Identifier // 参数名（可选，用于命名参数）
	Value Expression  // 参数值
}

func (ca *CallArgument) String() string {
	if ca.Name != nil {
		return ca.Name.String() + ": " + ca.Value.String()
	}
	return ca.Value.String()
}

// Identifier 标识符
// 对应语法：变量名、函数名等
// 例如：x, add, fmt.Println
type Identifier struct {
	Token lexer.Token // 标识符对应的 token
	Value string      // 标识符的值（字符串）
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral 整数字面量
// 对应语法：整数
// 例如：42, 100
type IntegerLiteral struct {
	Token lexer.Token // 整数对应的 token
	Value int64       // 整数值
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string      { return il.Token.Literal }

// StringLiteral 字符串字面量
// 对应语法："字符串"
// 例如："hello", "world"
type StringLiteral struct {
	Token lexer.Token // 字符串对应的 token
	Value string      // 字符串值
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

// BooleanLiteral 布尔字面量
// 对应语法：true 或 false
type BooleanLiteral struct {
	Token lexer.Token // true/false 对应的 token
	Value bool        // 布尔值
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string      { return bl.Token.Literal }

// NullLiteral null 字面量
// 对应语法：null
// 表示空值
type NullLiteral struct {
	Token lexer.Token // null 对应的 token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullLiteral) String() string       { return "null" }

// PrefixExpression 前缀表达式
// 对应语法：operator operand
// 例如：!true, -5
type PrefixExpression struct {
	Token    lexer.Token // 运算符对应的 token
	Operator string      // 运算符（如 !, -）
	Right    Expression  // 右操作数
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

// InfixExpression 中缀表达式
// 对应语法：left operator right
// 例如：1 + 2, a == b, x > 10
type InfixExpression struct {
	Token    lexer.Token // 运算符对应的 token
	Left     Expression  // 左操作数
	Operator string      // 运算符（如 +, -, ==, && 等）
	Right    Expression  // 右操作数
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

// TernaryExpression 三目运算符表达式
// 对应语法：condition ? trueExpr : falseExpr
// 例如：x > 0 ? 1 : -1
type TernaryExpression struct {
	Token     lexer.Token // ? 对应的 token
	Condition Expression  // 条件表达式
	TrueExpr  Expression  // 条件为真时的表达式
	FalseExpr Expression  // 条件为假时的表达式
}

func (te *TernaryExpression) expressionNode()      {}
func (te *TernaryExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TernaryExpression) String() string {
	return "(" + te.Condition.String() + " ? " + te.TrueExpr.String() + " : " + te.FalseExpr.String() + ")"
}

// TypeAssertionExpression 类型断言表达式
// 对应语法：value.(type)
// 例如：x.(string)
// 注意：当前版本尚未完全实现
type TypeAssertionExpression struct {
	Token lexer.Token // .( 对应的 token
	Left  Expression  // 要断言的值
	Type  *Identifier // 要断言的类型
}

func (tae *TypeAssertionExpression) expressionNode()      {}
func (tae *TypeAssertionExpression) TokenLiteral() string { return tae.Token.Literal }
func (tae *TypeAssertionExpression) String() string {
	return "(" + tae.Left.String() + ".( " + tae.Type.String() + "))"
}

// ========== 包和导入 ==========

// PackageStatement 包声明语句
// 对应语法：package packageName
// 例如：package main
type PackageStatement struct {
	Token lexer.Token // package 关键字对应的 token
	Name  *Identifier // 包名
}

func (ps *PackageStatement) statementNode()       {}
func (ps *PackageStatement) TokenLiteral() string { return ps.Token.Literal }
func (ps *PackageStatement) String() string {
	return "package " + ps.Name.String()
}

// ImportStatement 导入语句
// 对应语法：import "package/path"
// 例如：import "util.string"
type ImportStatement struct {
	Token lexer.Token // import 关键字对应的 token
	Path  *StringLiteral // 导入路径
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	return "import " + is.Path.String()
}

// ========== 类相关 ==========

// ClassStatement 类声明语句
// 对应语法：class ClassName { ... }
type ClassStatement struct {
	Token    lexer.Token      // class 关键字对应的 token
	Name     *Identifier      // 类名
	Members  []ClassMember    // 类成员（变量、方法）
}

func (cs *ClassStatement) statementNode()       {}
func (cs *ClassStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ClassStatement) String() string {
	var out string
	out += "class " + cs.Name.String() + " { "
	for _, member := range cs.Members {
		out += member.String() + " "
	}
	out += "}"
	return out
}

// ClassMember 类成员接口
type ClassMember interface {
	Node
	classMemberNode()
}

// ClassVariable 类成员变量
// 对应语法：访问修饰符 变量名 类型 或 访问修饰符 变量名 类型 = 值
type ClassVariable struct {
	Token         lexer.Token // 访问修饰符对应的 token
	AccessModifier string     // 访问修饰符：public, private, protected
	Name          *Identifier // 变量名
	Type          *Identifier // 变量类型
	Value         Expression  // 初始值（可选）
}

func (cv *ClassVariable) classMemberNode()      {}
func (cv *ClassVariable) TokenLiteral() string { return cv.Token.Literal }
func (cv *ClassVariable) String() string {
	var out string
	out += cv.AccessModifier + " " + cv.Name.String() + " " + cv.Type.String()
	if cv.Value != nil {
		out += " = " + cv.Value.String()
	}
	return out
}

// ClassMethod 类方法
// 对应语法：访问修饰符 [static] function 方法名(参数): 返回类型 { ... }
type ClassMethod struct {
	Token          lexer.Token          // 访问修饰符对应的 token
	AccessModifier string               // 访问修饰符：public, private, protected
	IsStatic       bool                 // 是否是静态方法
	Name           *Identifier         // 方法名（__construct 表示构造方法）
	Parameters     []*FunctionParameter // 参数列表
	ReturnType     []*Identifier       // 返回类型列表
	Body           *BlockStatement     // 方法体
}

func (cm *ClassMethod) classMemberNode()      {}
func (cm *ClassMethod) TokenLiteral() string { return cm.Token.Literal }
func (cm *ClassMethod) String() string {
	var out string
	out += cm.AccessModifier + " "
	if cm.IsStatic {
		out += "static "
	}
	out += "function " + cm.Name.String() + "("
	for i, p := range cm.Parameters {
		if i > 0 {
			out += ", "
		}
		out += p.String()
	}
	out += ")"
	if len(cm.ReturnType) > 0 {
		out += ": "
		if len(cm.ReturnType) == 1 {
			out += cm.ReturnType[0].String()
		} else {
			out += "("
			for i, rt := range cm.ReturnType {
				if i > 0 {
					out += ", "
				}
				out += rt.String()
			}
			out += ")"
		}
	}
	out += " " + cm.Body.String()
	return out
}

// ThisExpression this 表达式
// 对应语法：this
// 用于访问当前对象的成员
type ThisExpression struct {
	Token lexer.Token // this 关键字对应的 token
}

func (te *ThisExpression) expressionNode()      {}
func (te *ThisExpression) TokenLiteral() string { return te.Token.Literal }
func (te *ThisExpression) String() string       { return "this" }

// NewExpression new 表达式
// 对应语法：new ClassName(参数)
// 例如：new UserModel("John")
type NewExpression struct {
	Token     lexer.Token    // new 关键字对应的 token
	ClassName *Identifier    // 类名
	Arguments []CallArgument // 构造参数
}

func (ne *NewExpression) expressionNode()      {}
func (ne *NewExpression) TokenLiteral() string { return ne.Token.Literal }
func (ne *NewExpression) String() string {
	var out string
	out += "new " + ne.ClassName.String() + "("
	for i, arg := range ne.Arguments {
		if i > 0 {
			out += ", "
		}
		out += arg.String()
	}
	out += ")"
	return out
}

// StaticCallExpression 静态方法调用表达式
// 对应语法：ClassName::methodName(参数)
// 例如：UserModel::getTableName()
type StaticCallExpression struct {
	Token     lexer.Token    // :: 对应的 token
	ClassName *Identifier    // 类名
	Method    *Identifier    // 方法名
	Arguments []CallArgument // 参数列表
}

func (sce *StaticCallExpression) expressionNode()      {}
func (sce *StaticCallExpression) TokenLiteral() string { return sce.Token.Literal }
func (sce *StaticCallExpression) String() string {
	var out string
	out += sce.ClassName.String() + "::" + sce.Method.String() + "("
	for i, arg := range sce.Arguments {
		if i > 0 {
			out += ", "
		}
		out += arg.String()
	}
	out += ")"
	return out
}

// MemberAccessExpression 成员访问表达式
// 对应语法：object.member
// 例如：user.getName()
type MemberAccessExpression struct {
	Token  lexer.Token // . 对应的 token
	Object Expression  // 对象表达式
	Member *Identifier // 成员名
}

func (mae *MemberAccessExpression) expressionNode()      {}
func (mae *MemberAccessExpression) TokenLiteral() string { return mae.Token.Literal }
func (mae *MemberAccessExpression) String() string {
	return "(" + mae.Object.String() + "." + mae.Member.String() + ")"
}

// AssignmentExpression 赋值表达式
// 对应语法：left = right
// 例如：this.name = value, x = 10
type AssignmentExpression struct {
	Token lexer.Token // = 对应的 token
	Left  Expression  // 左边的表达式（通常是标识符或成员访问表达式）
	Right Expression  // 要赋的值
}

func (ae *AssignmentExpression) expressionNode()      {}
func (ae *AssignmentExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AssignmentExpression) String() string {
	return "(" + ae.Left.String() + " = " + ae.Right.String() + ")"
}

