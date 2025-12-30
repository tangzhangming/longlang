package lexer

// TokenType 表示 token 的类型
// token 是词法分析的基本单元，代表源代码中的一个有意义的片段
type TokenType string

const (
	// ========== 特殊 token ==========
	ILLEGAL TokenType = "ILLEGAL" // 非法字符，无法识别
	EOF     TokenType = "EOF"     // 文件结束标记

	// ========== 标识符和字面量 ==========
	IDENT  TokenType = "IDENT"  // 标识符：变量名、函数名等（如 name, add）
	INT    TokenType = "INT"     // 整数字面量（如 123, 456）
	STRING TokenType = "STRING"   // 字符串字面量（如 "hello"）
	TRUE   TokenType = "TRUE"    // 布尔值 true
	FALSE  TokenType = "FALSE"    // 布尔值 false
	NULL   TokenType = "NULL"     // null 值

	// ========== 运算符 ==========
	ASSIGN   TokenType = "="  // 赋值运算符 =
	PLUS     TokenType = "+"  // 加法运算符 +
	MINUS    TokenType = "-"  // 减法运算符 -（也用作负号）
	BANG     TokenType = "!"  // 逻辑非运算符 !
	ASTERISK TokenType = "*"  // 乘法运算符 *
	SLASH    TokenType = "/"  // 除法运算符 /（也用作注释开始）
	MOD      TokenType = "%"  // 取模运算符 %

	// 比较运算符
	LT     TokenType = "<"   // 小于 <
	GT     TokenType = ">"   // 大于 >
	EQ     TokenType = "=="  // 等于 ==
	NOT_EQ TokenType = "!="  // 不等于 !=
	LE     TokenType = "<="  // 小于等于 <=
	GE     TokenType = ">="  // 大于等于 >=

	// 逻辑运算符
	AND TokenType = "&&" // 逻辑与 &&
	OR  TokenType = "||" // 逻辑或 ||

	// ========== 分隔符 ==========
	COMMA     TokenType = "," // 逗号，用于分隔参数、元素等
	SEMICOLON TokenType = ";" // 分号，语句结束符（可选）
	COLON     TokenType = ":" // 冒号，用于类型声明、命名参数等
	QUESTION  TokenType = "?" // 问号，用于三目运算符

	// 括号
	LPAREN   TokenType = "(" // 左圆括号 (
	RPAREN   TokenType = ")" // 右圆括号 )
	LBRACE   TokenType = "{" // 左花括号 {
	RBRACE   TokenType = "}" // 右花括号 }
	LBRACKET TokenType = "[" // 左方括号 [
	RBRACKET TokenType = "]" // 右方括号 ]

	// ========== 关键字 ==========
	FUNCTION TokenType = "FUNCTION" // fn - 函数定义关键字
	VAR      TokenType = "VAR"      // var - 变量声明关键字
	IF       TokenType = "IF"       // if - 条件语句关键字
	ELSE     TokenType = "ELSE"     // else - else 分支关键字
	RETURN   TokenType = "RETURN"   // return - 返回语句关键字
	ANY      TokenType = "ANY"      // any - 任意类型关键字
	VOID     TokenType = "VOID"     // void - 无返回值类型关键字

	// ========== 类型关键字 ==========
	STRING_TYPE TokenType = "STRING_TYPE" // string - 字符串类型
	INT_TYPE    TokenType = "INT_TYPE"    // int - 整数类型
	BOOL_TYPE   TokenType = "BOOL_TYPE"   // bool - 布尔类型
)

// Token 表示一个词法单元
// 包含类型、字面值、位置信息等
type Token struct {
	Type    TokenType // token 类型
	Literal string    // token 的字面值（源代码中的原始字符串）
	Line    int       // token 所在的行号（从1开始）
	Column  int       // token 所在的列号（从1开始）
}

// keywords 关键字映射表
// 将字符串关键字映射到对应的 TokenType
var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"var":    VAR,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"any":    ANY,
	"void":   VOID,
	"true":   TRUE,
	"false":  FALSE,
	"null":   NULL,
	"string": STRING_TYPE,
	"int":    INT_TYPE,
	"bool":   BOOL_TYPE,
}

// LookupIdent 检查标识符是否是关键字
// 如果标识符在关键字表中，返回对应的 TokenType
// 否则返回 IDENT，表示这是一个普通的标识符
// 参数:
//   ident: 要检查的标识符字符串
// 返回:
//   如果是关键字，返回对应的 TokenType；否则返回 IDENT
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
