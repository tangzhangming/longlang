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
	FLOAT  TokenType = "FLOAT"   // 浮点数字面量（如 3.14, 2.0）
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
	
	// 复合赋值运算符
	PLUS_ASSIGN     TokenType = "+="  // += 加法赋值
	MINUS_ASSIGN    TokenType = "-="  // -= 减法赋值
	ASTERISK_ASSIGN TokenType = "*="  // *= 乘法赋值
	SLASH_ASSIGN    TokenType = "/="  // /= 除法赋值
	MOD_ASSIGN      TokenType = "%="  // %= 取模赋值
	
	// 位运算符
	LSHIFT         TokenType = "<<"   // << 左移
	RSHIFT         TokenType = ">>"   // >> 右移
	LSHIFT_ASSIGN  TokenType = "<<="  // <<= 左移赋值
	RSHIFT_ASSIGN  TokenType = ">>="  // >>= 右移赋值
	BITAND         TokenType = "&"    // & 位与
	BITOR          TokenType = "|"    // | 位或
	BITXOR         TokenType = "^"    // ^ 位异或
	BITNOT         TokenType = "~"    // ~ 位非
	BITAND_ASSIGN  TokenType = "&="   // &= 位与赋值
	BITOR_ASSIGN   TokenType = "|="   // |= 位或赋值
	BITXOR_ASSIGN  TokenType = "^="   // ^= 位异或赋值
	BIT_AND        TokenType = "&"    // & 位与（别名）
	BIT_OR         TokenType = "|"    // | 位或（别名）
	BIT_XOR        TokenType = "^"    // ^ 位异或（别名）
	BIT_NOT        TokenType = "~"    // ~ 位非（别名）
	BIT_AND_ASSIGN TokenType = "&="   // &= 位与赋值（别名）
	BIT_OR_ASSIGN  TokenType = "|="   // |= 位或赋值（别名）
	BIT_XOR_ASSIGN TokenType = "^="   // ^= 位异或赋值（别名）
	
	// 自增自减运算符
	INCREMENT TokenType = "++" // ++ 自增
	DECREMENT TokenType = "--" // -- 自减
	
	// 箭头运算符
	ARROW TokenType = "=>" // => 箭头（match 表达式）
	
	// 其他运算符
	ELLIPSIS      TokenType = "..."  // ... 展开运算符
	AT            TokenType = "@"    // @ 注解符号
	INTERP_STRING TokenType = "INTERP_STRING" // 插值字符串
	
	// 类型转换运算符
	AS      TokenType = "AS"      // as 类型转换
	AS_SAFE TokenType = "AS_SAFE" // as? 安全类型转换

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
	FLOAT_TYPE  TokenType = "FLOAT_TYPE"  // float - 浮点类型
	BYTE_TYPE   TokenType = "BYTE_TYPE"   // byte - 字节类型
	UINT_TYPE   TokenType = "UINT_TYPE"   // uint - 无符号整数类型
	I8_TYPE     TokenType = "I8_TYPE"     // i8 - 8位有符号整数
	I16_TYPE    TokenType = "I16_TYPE"    // i16 - 16位有符号整数
	I32_TYPE    TokenType = "I32_TYPE"    // i32 - 32位有符号整数
	I64_TYPE    TokenType = "I64_TYPE"    // i64 - 64位有符号整数
	U8_TYPE     TokenType = "U8_TYPE"     // u8 - 8位无符号整数
	U16_TYPE    TokenType = "U16_TYPE"    // u16 - 16位无符号整数
	U32_TYPE    TokenType = "U32_TYPE"    // u32 - 32位无符号整数
	U64_TYPE    TokenType = "U64_TYPE"    // u64 - 64位无符号整数
	F32_TYPE    TokenType = "F32_TYPE"    // f32 - 32位浮点数
	F64_TYPE    TokenType = "F64_TYPE"    // f64 - 64位浮点数
	
	// ========== 其他关键字 ==========
	CONST TokenType = "CONST" // const - 常量关键字
	
	// ========== 包和导入关键字 ==========
	PACKAGE TokenType = "PACKAGE" // package - 包声明关键字
	IMPORT  TokenType = "IMPORT"  // import - 导入关键字
	
	// ========== 类相关关键字 ==========
	CLASS      TokenType = "CLASS"      // class - 类定义关键字
	PUBLIC     TokenType = "PUBLIC"     // public - 公开访问修饰符
	PRIVATE    TokenType = "PRIVATE"    // private - 私有访问修饰符
	PROTECTED  TokenType = "PROTECTED"  // protected - 受保护访问修饰符
	STATIC     TokenType = "STATIC"     // static - 静态关键字
	THIS       TokenType = "THIS"       // this - 当前对象关键字
	NEW        TokenType = "NEW"        // new - 创建对象关键字
	SUPER      TokenType = "SUPER"      // super - 父类关键字
	EXTENDS    TokenType = "EXTENDS"    // extends - 继承关键字
	IMPLEMENTS TokenType = "IMPLEMENTS" // implements - 实现接口关键字
	ABSTRACT   TokenType = "ABSTRACT"   // abstract - 抽象关键字
	
	// ========== 控制流关键字 ==========
	MAP      TokenType = "MAP"      // map - 映射类型
	MATCH    TokenType = "MATCH"    // match - 模式匹配
	FOR      TokenType = "FOR"      // for - 循环
	BREAK    TokenType = "BREAK"    // break - 跳出循环
	CONTINUE TokenType = "CONTINUE" // continue - 继续循环
	TRY      TokenType = "TRY"      // try - 异常处理
	CATCH    TokenType = "CATCH"    // catch - 捕获异常
	FINALLY  TokenType = "FINALLY"  // finally - 最终执行
	THROW    TokenType = "THROW"    // throw - 抛出异常
	GO       TokenType = "GO"       // go - 协程
	SWITCH   TokenType = "SWITCH"   // switch - 分支语句
	CASE     TokenType = "CASE"     // case - 分支条件
	DEFAULT  TokenType = "DEFAULT"  // default - 默认分支
	RANGE    TokenType = "RANGE"    // range - 范围迭代
	
	// ========== 命名空间和模块关键字 ==========
	NAMESPACE  TokenType = "NAMESPACE"  // namespace - 命名空间
	USE        TokenType = "USE"        // use - 导入
	ANNOTATION TokenType = "ANNOTATION" // annotation - 注解
	INTERNAL   TokenType = "INTERNAL"   // internal - 内部访问修饰符
	INTERFACE  TokenType = "INTERFACE"  // interface - 接口
	ENUM       TokenType = "ENUM"       // enum - 枚举
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
	"fn":         FUNCTION,
	"function":   FUNCTION, // function 也是函数关键字（用于类方法）
	"var":        VAR,
	"if":         IF,
	"else":       ELSE,
	"return":     RETURN,
	"any":        ANY,
	"void":       VOID,
	"true":       TRUE,
	"false":      FALSE,
	"null":       NULL,
	"string":     STRING_TYPE,
	"int":        INT_TYPE,
	"bool":       BOOL_TYPE,
	"float":      FLOAT_TYPE,
	"byte":       BYTE_TYPE,
	"uint":       UINT_TYPE,
	"i8":         I8_TYPE,
	"i16":        I16_TYPE,
	"i32":        I32_TYPE,
	"i64":        I64_TYPE,
	"u8":         U8_TYPE,
	"u16":        U16_TYPE,
	"u32":        U32_TYPE,
	"u64":        U64_TYPE,
	"f32":        F32_TYPE,
	"f64":        F64_TYPE,
	"package":    PACKAGE,
	"import":     IMPORT,
	"class":      CLASS,
	"public":     PUBLIC,
	"private":    PRIVATE,
	"protected":  PROTECTED,
	"static":     STATIC,
	"this":       THIS,
	"new":        NEW,
	"super":      SUPER,
	"extends":    EXTENDS,
	"implements": IMPLEMENTS,
	"abstract":   ABSTRACT,
	"map":        MAP,
	"match":      MATCH,
	"as":         AS,
	"const":      CONST,
	"for":        FOR,
	"break":      BREAK,
	"continue":   CONTINUE,
	"try":        TRY,
	"catch":      CATCH,
	"finally":    FINALLY,
	"throw":      THROW,
	"go":         GO,
	"switch":     SWITCH,
	"case":       CASE,
	"default":    DEFAULT,
	"range":      RANGE,
	"namespace":  NAMESPACE,
	"use":        USE,
	"annotation": ANNOTATION,
	"internal":   INTERNAL,
	"interface":  INTERFACE,
	"enum":       ENUM,
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
