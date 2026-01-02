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
	INT    TokenType = "INT"    // 整数字面量（如 123, 456）
	FLOAT  TokenType = "FLOAT"  // 浮点数字面量（如 3.14, 2.5）
	STRING TokenType = "STRING" // 字符串字面量（如 "hello", 'world', `raw`）
	TRUE   TokenType = "TRUE"   // 布尔值 true
	FALSE  TokenType = "FALSE"  // 布尔值 false
	NULL   TokenType = "NULL"   // null 值

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

	// 位运算符
	BIT_AND TokenType = "&"  // 按位与 &
	BIT_OR  TokenType = "|"  // 按位或 |
	BIT_XOR TokenType = "^"  // 按位异或 ^
	BIT_NOT TokenType = "~"  // 按位取反 ~
	LSHIFT  TokenType = "<<" // 左移 <<
	RSHIFT  TokenType = ">>" // 右移 >>

	// 复合赋值运算符
	PLUS_ASSIGN    TokenType = "+="  // 加法赋值 +=
	MINUS_ASSIGN   TokenType = "-="  // 减法赋值 -=
	ASTERISK_ASSIGN TokenType = "*=" // 乘法赋值 *=
	SLASH_ASSIGN   TokenType = "/="  // 除法赋值 /=
	MOD_ASSIGN     TokenType = "%="  // 取模赋值 %=
	BIT_AND_ASSIGN TokenType = "&="  // 按位与赋值 &=
	BIT_OR_ASSIGN  TokenType = "|="  // 按位或赋值 |=
	BIT_XOR_ASSIGN TokenType = "^="  // 按位异或赋值 ^=
	LSHIFT_ASSIGN  TokenType = "<<=" // 左移赋值 <<=
	RSHIFT_ASSIGN  TokenType = ">>=" // 右移赋值 >>=

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
	RBRACKET TokenType = "]" // 右方括号 []
	
	// 静态方法调用运算符
	DOUBLE_COLON TokenType = "::" // :: - 静态方法调用运算符
	DOT          TokenType = "."  // . - 成员访问运算符
	ELLIPSIS     TokenType = "..." // ... - 省略号（用于数组长度推导）

	// 自增自减运算符
	INCREMENT TokenType = "++" // 自增运算符 ++
	DECREMENT TokenType = "--" // 自减运算符 --

	// ========== 关键字 ==========
	FUNCTION TokenType = "FUNCTION" // fn - 函数定义关键字
	VAR      TokenType = "VAR"      // var - 变量声明关键字
	IF       TokenType = "IF"       // if - 条件语句关键字
	ELSE     TokenType = "ELSE"     // else - else 分支关键字
	FOR      TokenType = "FOR"      // for - 循环关键字
	BREAK    TokenType = "BREAK"    // break - 跳出循环关键字
	CONTINUE TokenType = "CONTINUE" // continue - 继续循环关键字
	RETURN   TokenType = "RETURN"   // return - 返回语句关键字
	ANY      TokenType = "ANY"      // any - 任意类型关键字
	VOID     TokenType = "VOID"     // void - 无返回值类型关键字

	// ========== 类型关键字 ==========
	STRING_TYPE TokenType = "STRING_TYPE" // string - 字符串类型
	BOOL_TYPE   TokenType = "BOOL_TYPE"   // bool - 布尔类型

	// 有符号整型
	INT_TYPE TokenType = "INT_TYPE" // int - 平台相关有符号整型
	I8_TYPE  TokenType = "I8_TYPE"  // i8 - 8位有符号整型
	I16_TYPE TokenType = "I16_TYPE" // i16 - 16位有符号整型
	I32_TYPE TokenType = "I32_TYPE" // i32 - 32位有符号整型
	I64_TYPE TokenType = "I64_TYPE" // i64 - 64位有符号整型

	// 无符号整型
	UINT_TYPE TokenType = "UINT_TYPE" // uint - 平台相关无符号整型
	U8_TYPE   TokenType = "U8_TYPE"   // u8 - 8位无符号整型
	BYTE_TYPE TokenType = "BYTE_TYPE" // byte - u8 的别名，8位无符号整型
	U16_TYPE  TokenType = "U16_TYPE"  // u16 - 16位无符号整型
	U32_TYPE  TokenType = "U32_TYPE"  // u32 - 32位无符号整型
	U64_TYPE  TokenType = "U64_TYPE"  // u64 - 64位无符号整型

	// 浮点数类型
	FLOAT_TYPE TokenType = "FLOAT_TYPE" // float - 平台相关浮点数
	F32_TYPE   TokenType = "F32_TYPE"   // f32 - 32位浮点数
	F64_TYPE   TokenType = "F64_TYPE"   // f64 - 64位浮点数
	
	// ========== 命名空间关键字 ==========
	NAMESPACE TokenType = "NAMESPACE" // namespace - 命名空间声明关键字
	USE       TokenType = "USE"       // use - 导入关键字
	
	// ========== 类相关关键字 ==========
	CLASS      TokenType = "CLASS"      // class - 类定义关键字
	ABSTRACT   TokenType = "ABSTRACT"   // abstract - 抽象类/方法关键字
	INTERFACE  TokenType = "INTERFACE"  // interface - 接口定义关键字
	EXTENDS    TokenType = "EXTENDS"    // extends - 继承关键字
	IMPLEMENTS TokenType = "IMPLEMENTS" // implements - 实现接口关键字
	PUBLIC     TokenType = "PUBLIC"     // public - 公开访问修饰符
	PRIVATE    TokenType = "PRIVATE"    // private - 私有访问修饰符
	PROTECTED  TokenType = "PROTECTED"  // protected - 受保护访问修饰符
	INTERNAL   TokenType = "INTERNAL"   // internal - 命名空间内部可见修饰符
	STATIC     TokenType = "STATIC"     // static - 静态关键字
	CONST      TokenType = "CONST"      // const - 常量关键字
	THIS       TokenType = "THIS"       // this - 当前对象关键字
	SUPER      TokenType = "SUPER"      // super - 父类关键字
	NEW        TokenType = "NEW"        // new - 创建对象关键字
	
	// ========== 异常处理关键字 ==========
	TRY     TokenType = "TRY"     // try - 异常捕获块
	CATCH   TokenType = "CATCH"   // catch - 异常处理块
	FINALLY TokenType = "FINALLY" // finally - 最终执行块
	THROW   TokenType = "THROW"   // throw - 抛出异常
	
	// ========== 复合类型关键字 ==========
	MAP  TokenType = "MAP"  // map - Map 类型关键字
	ENUM TokenType = "ENUM" // enum - 枚举类型关键字
	
	// ========== 并发关键字 ==========
	GO TokenType = "GO" // go - 启动协程关键字
	
	// ========== 控制流关键字 ==========
	SWITCH  TokenType = "SWITCH"  // switch - 分支语句关键字
	MATCH   TokenType = "MATCH"   // match - 模式匹配表达式关键字
	CASE    TokenType = "CASE"    // case - 分支条件关键字
	DEFAULT TokenType = "DEFAULT" // default - 默认分支关键字
	RANGE   TokenType = "RANGE"   // range - for-range 遍历关键字
	
	// ========== 类型断言关键字 ==========
	AS       TokenType = "AS"        // as - 类型断言关键字
	AS_SAFE  TokenType = "AS_SAFE"   // as? - 安全类型断言
	
	// ========== 特殊运算符 ==========
	ARROW TokenType = "=>" // => - 匹配箭头运算符
	
	// ========== 字符串插值 ==========
	INTERP_STRING TokenType = "INTERP_STRING" // $"..." - 插值字符串
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
	"fn":        FUNCTION,
	"function":  FUNCTION, // function 也是函数关键字（用于类方法）
	"var":       VAR,
	"if":        IF,
	"else":      ELSE,
	"for":       FOR,
	"break":     BREAK,
	"continue":  CONTINUE,
	"return":    RETURN,
	"any":       ANY,
	"void":      VOID,
	"true":      TRUE,
	"false":     FALSE,
	"null":      NULL,
	// 字符串和布尔类型
	"string": STRING_TYPE,
	"bool":   BOOL_TYPE,
	// 有符号整型
	"int": INT_TYPE,
	"i8":  I8_TYPE,
	"i16": I16_TYPE,
	"i32": I32_TYPE,
	"i64": I64_TYPE,
	// 无符号整型
	"uint": UINT_TYPE,
	"u8":   U8_TYPE,
	"byte": BYTE_TYPE, // byte 是 u8 的别名
	"u16":  U16_TYPE,
	"u32":  U32_TYPE,
	"u64":  U64_TYPE,
	// 浮点数类型
	"float": FLOAT_TYPE,
	"f32":   F32_TYPE,
	"f64":   F64_TYPE,
	// 命名空间和类相关
	"namespace":  NAMESPACE,
	"use":        USE,
	"class":      CLASS,
	"abstract":   ABSTRACT,
	"interface":  INTERFACE,
	"extends":    EXTENDS,
	"implements": IMPLEMENTS,
	"public":     PUBLIC,
	"private":    PRIVATE,
	"protected":  PROTECTED,
	"internal":   INTERNAL,
	"static":     STATIC,
	"const":      CONST,
	"this":       THIS,
	"super":      SUPER,
	"new":        NEW,
	// 异常处理
	"try":     TRY,
	"catch":   CATCH,
	"finally": FINALLY,
	"throw":   THROW,
	// 复合类型
	"map":  MAP,
	"enum": ENUM,
	// 并发
	"go": GO,
	// 控制流
	"switch":  SWITCH,
	"match":   MATCH,
	"case":    CASE,
	"default": DEFAULT,
	"range":   RANGE,
	// 类型断言
	"as": AS,
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
