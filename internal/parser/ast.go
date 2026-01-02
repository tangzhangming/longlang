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
// 例如：var x int = 10 或 var y = 20 或 var numbers [5]int = {1,2,3,4,5}
type LetStatement struct {
	Token lexer.Token // var 关键字对应的 token
	Name  *Identifier // 变量名
	Type  Expression  // 变量类型（可选，*Identifier 或 *ArrayType）
	Value Expression  // 变量的初始值（可选）
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out string
	out += ls.TokenLiteral() + " "
	out += ls.Name.String()
	if ls.Type != nil {
		switch t := ls.Type.(type) {
		case *Identifier:
			out += " " + t.String()
		case *ArrayType:
			out += " " + t.String()
		default:
			out += " " + ls.Type.String()
		}
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

// ========== 异常处理语句 ==========

// TryStatement try-catch-finally 语句
// 对应语法：try { ... } catch (ExceptionType e) { ... } finally { ... }
type TryStatement struct {
	Token       lexer.Token     // try 关键字对应的 token
	TryBlock    *BlockStatement // try 块
	CatchClauses []*CatchClause // catch 子句列表（可多个）
	FinallyBlock *BlockStatement // finally 块（可选）
}

func (ts *TryStatement) statementNode()       {}
func (ts *TryStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *TryStatement) String() string {
	var out string
	out += "try " + ts.TryBlock.String()
	for _, catch := range ts.CatchClauses {
		out += " " + catch.String()
	}
	if ts.FinallyBlock != nil {
		out += " finally " + ts.FinallyBlock.String()
	}
	return out
}

// CatchClause catch 子句
// 对应语法：catch (ExceptionType variableName) { ... } 或 catch (variableName) { ... }
type CatchClause struct {
	Token         lexer.Token     // catch 关键字对应的 token
	ExceptionType *Identifier     // 异常类型（可选，nil 表示无类型 catch）
	ExceptionVar  *Identifier     // 异常变量名
	Body          *BlockStatement // catch 块
}

func (cc *CatchClause) TokenLiteral() string { return cc.Token.Literal }
func (cc *CatchClause) String() string {
	var out string
	out += "catch ("
	if cc.ExceptionType != nil {
		out += cc.ExceptionType.String() + " "
	}
	out += cc.ExceptionVar.String() + ") " + cc.Body.String()
	return out
}

// ThrowStatement throw 语句
// 对应语法：throw expression
// 例如：throw new Exception("错误消息")
type ThrowStatement struct {
	Token lexer.Token // throw 关键字对应的 token
	Value Expression  // 要抛出的异常表达式
}

func (ts *ThrowStatement) statementNode()       {}
func (ts *ThrowStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *ThrowStatement) String() string {
	return "throw " + ts.Value.String()
}

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
// 对应语法：name:type 或 name:type = defaultValue 或 ...name:type (可变参数)
// 例如：x:int 或 y:int = 10 或 ...args:any
type FunctionParameter struct {
	Name         *Identifier // 参数名
	Type         *Identifier // 参数类型（可选）
	DefaultValue Expression  // 默认值（可选）
	IsVariadic   bool        // 是否是可变参数（...args）
}

func (fp *FunctionParameter) String() string {
	var out string
	if fp.IsVariadic {
		out += "..."
	}
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
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral 浮点数字面量
// 对应语法：3.14, 2.5 等
type FloatLiteral struct {
	Token lexer.Token // 浮点数对应的 token
	Value float64     // 浮点数值
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral 字符串字面量
// 对应语法："字符串", '字符串', `原始字符串`
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

// ========== 命名空间和导入 ==========

// NamespaceStatement 命名空间声明语句
// 对应语法：namespace Namespace.Name 或 namespace Name
// 例如：namespace Mycompany.Myapp.Models 或 namespace Models
type NamespaceStatement struct {
	Token lexer.Token // namespace 关键字对应的 token
	Name  *Identifier // 命名空间名称（支持点分隔，如 "Mycompany.Myapp.Models"）
}

func (ns *NamespaceStatement) statementNode()       {}
func (ns *NamespaceStatement) TokenLiteral() string { return ns.Token.Literal }
func (ns *NamespaceStatement) String() string {
	return "namespace " + ns.Name.String()
}

// UseStatement 导入语句（use）
// 对应语法：use Full.Qualified.ClassName 或 use Namespace.ClassName as Alias
// 例如：use Illuminate.Database.Eloquent.Model 或 use Cache.Redis as RedisClient
type UseStatement struct {
	Token lexer.Token // use 关键字对应的 token
	Path  *Identifier // 完全限定名（如 "Illuminate.Database.Eloquent.Model"）
	Alias *Identifier // 别名（可选）
}

func (us *UseStatement) statementNode()       {}
func (us *UseStatement) TokenLiteral() string { return us.Token.Literal }
func (us *UseStatement) String() string {
	out := "use " + us.Path.String()
	if us.Alias != nil {
		out += " as " + us.Alias.String()
	}
	return out
}

// ========== 接口相关 ==========

// InterfaceStatement 接口声明语句
// 对应语法：interface InterfaceName { ... }
type InterfaceStatement struct {
	Token   lexer.Token         // interface 关键字对应的 token
	Name    *Identifier         // 接口名
	Methods []*InterfaceMethod  // 接口方法签名
}

func (is *InterfaceStatement) statementNode()       {}
func (is *InterfaceStatement) TokenLiteral() string { return is.Token.Literal }
func (is *InterfaceStatement) String() string {
	var out string
	out += "interface " + is.Name.String() + " { "
	for _, method := range is.Methods {
		out += method.String() + " "
	}
	out += "}"
	return out
}

// InterfaceMethod 接口方法签名
type InterfaceMethod struct {
	Token      lexer.Token            // function 关键字对应的 token
	Name       *Identifier            // 方法名
	Parameters []*FunctionParameter   // 参数列表
	ReturnType []*Identifier          // 返回类型
}

func (im *InterfaceMethod) TokenLiteral() string { return im.Token.Literal }
func (im *InterfaceMethod) String() string {
	var out string
	out += "function " + im.Name.String() + "("
	for i, p := range im.Parameters {
		if i > 0 {
			out += ", "
		}
		out += p.String()
	}
	out += ")"
	if len(im.ReturnType) > 0 {
		out += ":"
		for i, rt := range im.ReturnType {
			if i > 0 {
				out += ", "
			}
			out += rt.String()
		}
	}
	return out
}

// ========== 类相关 ==========

// ClassStatement 类声明语句
// 对应语法：class ClassName extends Parent implements Interface1, Interface2 { ... }
type ClassStatement struct {
	Token      lexer.Token     // class 关键字对应的 token
	Name       *Identifier     // 类名
	Parent     *Identifier     // 父类名（可选，用于继承）
	Interfaces []*Identifier   // 实现的接口列表
	Members    []ClassMember   // 类成员（变量、方法）
	IsAbstract bool            // 是否是抽象类
}

func (cs *ClassStatement) statementNode()       {}
func (cs *ClassStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ClassStatement) String() string {
	var out string
	if cs.IsAbstract {
		out += "abstract "
	}
	out += "class " + cs.Name.String()
	if cs.Parent != nil {
		out += " extends " + cs.Parent.String()
	}
	if len(cs.Interfaces) > 0 {
		out += " implements "
		for i, iface := range cs.Interfaces {
			if i > 0 {
				out += ", "
			}
			out += iface.String()
		}
	}
	out += " { "
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

// ClassConstant 类常量
// 对应语法：访问修饰符 const 常量名 [类型] = 值
// 例如：public const PI = 3.14159 或 public const PORT i16 = 8080
type ClassConstant struct {
	Token          lexer.Token // const 关键字对应的 token
	AccessModifier string      // 访问修饰符：public, private, protected
	Name           *Identifier // 常量名
	Type           *Identifier // 类型（可选，nil 表示类型推导）
	Value          Expression  // 常量值（必须是字面量）
}

func (cc *ClassConstant) classMemberNode()      {}
func (cc *ClassConstant) TokenLiteral() string { return cc.Token.Literal }
func (cc *ClassConstant) String() string {
	var out string
	out += cc.AccessModifier + " const " + cc.Name.String()
	if cc.Type != nil {
		out += " " + cc.Type.String()
	}
	out += " = " + cc.Value.String()
	return out
}

// ClassMethod 类方法
// 对应语法：访问修饰符 [static] function 方法名(参数): 返回类型 { ... }
type ClassMethod struct {
	Token          lexer.Token          // 访问修饰符对应的 token
	AccessModifier string               // 访问修饰符：public, private, protected
	IsStatic       bool                 // 是否是静态方法
	IsAbstract     bool                 // 是否是抽象方法
	Name           *Identifier         // 方法名（__construct 表示构造方法）
	Parameters     []*FunctionParameter // 参数列表
	ReturnType     []*Identifier       // 返回类型列表
	Body           *BlockStatement     // 方法体（抽象方法时为 nil）
}

func (cm *ClassMethod) classMemberNode()      {}
func (cm *ClassMethod) TokenLiteral() string { return cm.Token.Literal }
func (cm *ClassMethod) String() string {
	var out string
	if cm.IsAbstract {
		out += "abstract "
	}
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

// SuperExpression super 表达式
// 用于在子类中访问父类的方法
type SuperExpression struct {
	Token lexer.Token // super 关键字对应的 token
}

func (se *SuperExpression) expressionNode()      {}
func (se *SuperExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SuperExpression) String() string       { return "super" }

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

// StaticAccessExpression 静态访问表达式（常量访问）
// 对应语法：ClassName::CONST_NAME 或 self::CONST_NAME
// 例如：MyClass::MAX_SIZE 或 self::PI
type StaticAccessExpression struct {
	Token     lexer.Token // :: 对应的 token
	ClassName *Identifier // 类名或 self
	Name      *Identifier // 常量名
}

func (sae *StaticAccessExpression) expressionNode()      {}
func (sae *StaticAccessExpression) TokenLiteral() string { return sae.Token.Literal }
func (sae *StaticAccessExpression) String() string {
	return sae.ClassName.String() + "::" + sae.Name.String()
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

// ========== 数组相关 ==========

// ArrayType 数组类型
// 对应语法：[size]elementType 或 [...]elementType
// 例如：[5]int, [...]string
type ArrayType struct {
	Token       lexer.Token // [ 对应的 token
	Size        Expression  // 数组大小（IntegerLiteral），nil 表示切片类型
	IsInferred  bool        // 是否是 [...] 形式（长度推导）
	ElementType Expression  // 元素类型（Identifier 或嵌套的 ArrayType/SliceType）
}

func (at *ArrayType) expressionNode()      {}
func (at *ArrayType) TokenLiteral() string { return at.Token.Literal }
func (at *ArrayType) String() string {
	var out string
	out += "["
	if at.IsInferred {
		out += "..."
	} else if at.Size != nil {
		out += at.Size.String()
	}
	out += "]"
	if at.ElementType != nil {
		out += at.ElementType.String()
	}
	return out
}

// ArrayLiteral 数组字面量
// 对应语法：{element1, element2, ...}
// 例如：{1, 2, 3}, {"a", "b"}
type ArrayLiteral struct {
	Token    lexer.Token  // { 对应的 token
	Elements []Expression // 元素列表
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out string
	out += "{"
	for i, elem := range al.Elements {
		if i > 0 {
			out += ", "
		}
		out += elem.String()
	}
	out += "}"
	return out
}

// TypedArrayLiteral 带类型的数组字面量
// 对应语法：[size]type{elements} 或 []type{elements}
// 例如：[5]int{1, 2, 3, 4, 5}, []string{"a", "b"}
type TypedArrayLiteral struct {
	Token       lexer.Token  // [ 对应的 token
	Type        *ArrayType   // 数组类型
	Elements    []Expression // 元素列表
}

func (tal *TypedArrayLiteral) expressionNode()      {}
func (tal *TypedArrayLiteral) TokenLiteral() string { return tal.Token.Literal }
func (tal *TypedArrayLiteral) String() string {
	var out string
	out += tal.Type.String()
	out += "{"
	for i, elem := range tal.Elements {
		if i > 0 {
			out += ", "
		}
		out += elem.String()
	}
	out += "}"
	return out
}

// IndexExpression 索引访问表达式
// 对应语法：array[index]
// 例如：numbers[0], matrix[1][2]
type IndexExpression struct {
	Token lexer.Token // [ 对应的 token
	Left  Expression  // 数组表达式
	Index Expression  // 索引表达式
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	return "(" + ie.Left.String() + "[" + ie.Index.String() + "])"
}

// ========== Map 类型 ==========

// MapType Map 类型声明
// 对应语法：map[KeyType]ValueType
// 例如：map[string]int, map[string]User
type MapType struct {
	Token     lexer.Token // map 关键字的 token
	KeyType   *Identifier // 键类型（目前仅支持 string）
	ValueType Expression  // 值类型
}

func (mt *MapType) expressionNode()      {}
func (mt *MapType) TokenLiteral() string { return mt.Token.Literal }
func (mt *MapType) String() string {
	var out string
	out += "map["
	if mt.KeyType != nil {
		out += mt.KeyType.String()
	}
	out += "]"
	if mt.ValueType != nil {
		out += mt.ValueType.String()
	}
	return out
}

// MapLiteral Map 字面量
// 对应语法：map[KeyType]ValueType{key1: value1, key2: value2, ...}
// 例如：map[string]int{"Alice": 100, "Bob": 90}
type MapLiteral struct {
	Token   lexer.Token           // map 关键字的 token
	Type    *MapType              // Map 类型
	Pairs   map[Expression]Expression // 键值对（用于保持解析结果）
	Keys    []Expression          // 有序的键列表（保持插入顺序）
	Values  []Expression          // 对应的值列表
}

func (ml *MapLiteral) expressionNode()      {}
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Literal }
func (ml *MapLiteral) String() string {
	var out string
	out += ml.Type.String()
	out += "{"
	for i, key := range ml.Keys {
		if i > 0 {
			out += ", "
		}
		out += key.String() + ": " + ml.Values[i].String()
	}
	out += "}"
	return out
}

// ========== 枚举类型 ==========

// EnumStatement 枚举声明语句
// 对应语法：enum EnumName [: BackingType] [implements Interface1, ...] { Members }
// 例如：enum Color { Red, Green, Blue }
// 例如：enum Status: int { Pending = 0, Approved = 1 }
type EnumStatement struct {
	Token       lexer.Token      // enum 关键字对应的 token
	Name        *Identifier      // 枚举名
	BackingType *Identifier      // 底层类型（int 或 string，可选）
	Interfaces  []*Identifier    // 实现的接口列表
	Members     []*EnumMember    // 枚举成员
	Methods     []*ClassMethod   // 枚举方法
	Variables   []*ClassVariable // 枚举字段（用于复杂枚举）
}

func (es *EnumStatement) statementNode()       {}
func (es *EnumStatement) TokenLiteral() string { return es.Token.Literal }
func (es *EnumStatement) String() string {
	var out string
	out += "enum " + es.Name.String()
	if es.BackingType != nil {
		out += ": " + es.BackingType.String()
	}
	if len(es.Interfaces) > 0 {
		out += " implements "
		for i, iface := range es.Interfaces {
			if i > 0 {
				out += ", "
			}
			out += iface.String()
		}
	}
	out += " { "
	for _, member := range es.Members {
		out += member.String() + " "
	}
	out += "}"
	return out
}

// EnumMember 枚举成员
// 对应语法：MemberName [= value] 或 MemberName(args)
// 例如：Red, Pending = 0, Earth(mass, radius)
type EnumMember struct {
	Token     lexer.Token  // 成员名对应的 token
	Name      *Identifier  // 成员名
	Value     Expression   // 成员值（可选，用于带值枚举）
	Arguments []Expression // 构造参数（可选，用于复杂枚举）
}

func (em *EnumMember) TokenLiteral() string { return em.Token.Literal }
func (em *EnumMember) String() string {
	var out string
	out += em.Name.String()
	if len(em.Arguments) > 0 {
		out += "("
		for i, arg := range em.Arguments {
			if i > 0 {
				out += ", "
			}
			out += arg.String()
		}
		out += ")"
	} else if em.Value != nil {
		out += " = " + em.Value.String()
	}
	return out
}

// EnumAccessExpression 枚举成员访问表达式
// 对应语法：EnumName::MemberName
// 例如：Color::Red, Status::Pending
type EnumAccessExpression struct {
	Token    lexer.Token // :: 对应的 token
	EnumName *Identifier // 枚举名
	Member   *Identifier // 成员名
}

func (eae *EnumAccessExpression) expressionNode()      {}
func (eae *EnumAccessExpression) TokenLiteral() string { return eae.Token.Literal }
func (eae *EnumAccessExpression) String() string {
	return eae.EnumName.String() + "::" + eae.Member.String()
}

// ========== 并发相关 ==========

// GoStatement go 语句（启动协程）
// 对应语法：go expression
// 例如：go fn() { ... }
// 例如：go handler()
// 例如：go this.process()
// 例如：go Worker::run()
type GoStatement struct {
	Token lexer.Token // go 关键字对应的 token
	Call  Expression  // 要执行的表达式（通常是函数调用或闭包）
}

func (gs *GoStatement) statementNode()       {}
func (gs *GoStatement) TokenLiteral() string { return gs.Token.Literal }
func (gs *GoStatement) String() string {
	return "go " + gs.Call.String()
}

