package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer 词法分析器
// 负责将源代码字符串转换为 token 流
// 词法分析是编译/解释的第一步，将字符流转换为有意义的词法单元
type Lexer struct {
	input        string // 输入的源代码
	position     int    // 当前读取位置（当前字符的位置）
	readPosition int    // 下一个读取位置（用于预读）
	ch           byte   // 当前正在处理的字符
	line         int    // 当前行号（用于错误报告）
	column       int    // 当前列号（用于错误报告）
	isStdlib     bool   // 是否是标准库文件（允许使用 __ 前缀的内部函数）
}

// New 创建新的词法分析器
// 参数:
//   input: 要分析的源代码字符串
// 返回:
//   初始化好的 Lexer 实例
func New(input string) *Lexer {
	l := &Lexer{
		input:    input,
		line:     1, // 行号从1开始
		column:   1, // 列号从1开始
		isStdlib: false,
	}
	l.readChar() // 读取第一个字符
	return l
}

// NewWithOptions 创建新的词法分析器（带选项）
// 参数:
//   input: 要分析的源代码字符串
//   isStdlib: 是否是标准库文件
// 返回:
//   初始化好的 Lexer 实例
func NewWithOptions(input string, isStdlib bool) *Lexer {
	l := &Lexer{
		input:    input,
		line:     1,
		column:   1,
		isStdlib: isStdlib,
	}
	l.readChar()
	return l
}

// NewFromFile 根据文件路径创建词法分析器，自动判断是否为标准库
// 参数:
//   input: 要分析的源代码字符串
//   filePath: 文件路径
// 返回:
//   初始化好的 Lexer 实例
func NewFromFile(input string, filePath string) *Lexer {
	// 判断是否为标准库文件：路径包含 stdlib/ 或 stdlib\
	isStdlib := strings.Contains(filePath, "stdlib/") || strings.Contains(filePath, "stdlib\\")
	return NewWithOptions(input, isStdlib)
}

// IsStdlib 返回当前 Lexer 是否在标准库模式
func (l *Lexer) IsStdlib() bool {
	return l.isStdlib
}

// readChar 读取下一个字符并更新位置信息
// 将 readPosition 的字符读取到 ch，并更新 position
// 同时更新行号和列号信息
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		// 已到达文件末尾
		l.ch = 0
	} else {
		// 读取下一个字符
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	// 更新行号和列号
	if l.ch == '\n' {
		// 遇到换行符，行号加1，列号重置为1
		l.line++
		l.column = 1
	} else {
		// 列号加1
		l.column++
	}
}

// peekChar 查看下一个字符但不移动位置
// 用于前瞻，判断下一个字符是什么，但不实际读取
// 返回:
//   下一个字符，如果已到文件末尾则返回 0
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken 读取并返回下一个 token
// 这是词法分析器的核心方法，根据当前字符生成对应的 token
// 返回:
//   下一个 token
func (l *Lexer) NextToken() Token {
	var tok Token

	// 跳过空白字符（空格、制表符、换行符等）
	l.skipWhitespace()

	// 设置 token 的位置信息
	tok.Line = l.line
	tok.Column = l.column

	// 根据当前字符生成对应的 token
	switch l.ch {
	case '=':
		// 可能是 = 或 == 或 =>
		if l.peekChar() == '=' {
			// 是 ==
			ch := l.ch
			l.readChar()
			tok = Token{Type: EQ, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '>' {
			// 是 => (match 箭头)
			ch := l.ch
			l.readChar()
			tok = Token{Type: ARROW, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			// 是 =
			tok = newToken(ASSIGN, l.ch, l.line, l.column)
		}
	case '+':
		// 可能是 + 或 ++ 或 +=
		if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: INCREMENT, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: PLUS_ASSIGN, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			tok = newToken(PLUS, l.ch, l.line, l.column)
		}
	case '-':
		// 可能是 - 或 -- 或 -=
		if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: DECREMENT, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: MINUS_ASSIGN, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			tok = newToken(MINUS, l.ch, l.line, l.column)
		}
	case '!':
		// 可能是 ! 或 !=
		if l.peekChar() == '=' {
			// 是 !=
			ch := l.ch
			l.readChar()
			tok = Token{Type: NOT_EQ, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			// 是 !
			tok = newToken(BANG, l.ch, l.line, l.column)
		}
	case '*':
		// 可能是 * 或 *=
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: ASTERISK_ASSIGN, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			tok = newToken(ASTERISK, l.ch, l.line, l.column)
		}
	case '/':
		// 可能是 / 或 /= 或单行注释 // 或块注释 /*
		if l.peekChar() == '/' {
			// 单行注释，跳过注释内容
			l.skipLineComment()
			// 递归调用，返回注释后的下一个 token
			return l.NextToken()
		} else if l.peekChar() == '*' {
			// 块注释 /* */
			l.skipBlockComment()
			// 递归调用，返回注释后的下一个 token
			return l.NextToken()
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: SLASH_ASSIGN, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			// 是除法运算符
			tok = newToken(SLASH, l.ch, l.line, l.column)
		}
	case '%':
		// 可能是 % 或 %=
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: MOD_ASSIGN, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			tok = newToken(MOD, l.ch, l.line, l.column)
		}
	case '<':
		// 可能是 < 或 <= 或 << 或 <<=
		if l.peekChar() == '=' {
			// 是 <=
			ch := l.ch
			l.readChar()
			tok = Token{Type: LE, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '<' {
			// 可能是 << 或 <<=
			ch := l.ch
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = Token{Type: LSHIFT_ASSIGN, Literal: "<<=", Line: l.line, Column: l.column - 2}
			} else {
				tok = Token{Type: LSHIFT, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
			}
		} else {
			// 是 <
			tok = newToken(LT, l.ch, l.line, l.column)
		}
	case '>':
		// 可能是 > 或 >= 或 >> 或 >>=
		if l.peekChar() == '=' {
			// 是 >=
			ch := l.ch
			l.readChar()
			tok = Token{Type: GE, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '>' {
			// 可能是 >> 或 >>=
			ch := l.ch
			l.readChar()
			if l.peekChar() == '=' {
				l.readChar()
				tok = Token{Type: RSHIFT_ASSIGN, Literal: ">>=", Line: l.line, Column: l.column - 2}
			} else {
				tok = Token{Type: RSHIFT, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
			}
		} else {
			// 是 >
			tok = newToken(GT, l.ch, l.line, l.column)
		}
	case '&':
		// 可能是 & 或 && 或 &=
		if l.peekChar() == '&' {
			// 是 && 逻辑与
			ch := l.ch
			l.readChar()
			tok = Token{Type: AND, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '=' {
			// 是 &= 按位与赋值
			ch := l.ch
			l.readChar()
			tok = Token{Type: BIT_AND_ASSIGN, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			// 是 & 按位与
			tok = newToken(BIT_AND, l.ch, l.line, l.column)
		}
	case '|':
		// 可能是 | 或 || 或 |=
		if l.peekChar() == '|' {
			// 是 || 逻辑或
			ch := l.ch
			l.readChar()
			tok = Token{Type: OR, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '=' {
			// 是 |= 按位或赋值
			ch := l.ch
			l.readChar()
			tok = Token{Type: BIT_OR_ASSIGN, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			// 是 | 按位或
			tok = newToken(BIT_OR, l.ch, l.line, l.column)
		}
	case '^':
		// 可能是 ^ 或 ^=
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: BIT_XOR_ASSIGN, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			tok = newToken(BIT_XOR, l.ch, l.line, l.column)
		}
	case '~':
		// 按位取反
		tok = newToken(BIT_NOT, l.ch, l.line, l.column)
	case '?':
		// 三目运算符的 ?
		tok = newToken(QUESTION, l.ch, l.line, l.column)
	case ':':
		// 可能是 : 或 := 或 ::
		if l.peekChar() == '=' {
			// 是短变量声明 :=
			l.readChar()
			tok = Token{Type: ASSIGN, Literal: ":=", Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == ':' {
			// 是静态方法调用 ::
			ch := l.ch
			l.readChar()
			tok = Token{Type: DOUBLE_COLON, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			// 是冒号 :
			tok = newToken(COLON, l.ch, l.line, l.column)
		}
	case ';':
		tok = newToken(SEMICOLON, l.ch, l.line, l.column)
	case ',':
		tok = newToken(COMMA, l.ch, l.line, l.column)
	case '(':
		tok = newToken(LPAREN, l.ch, l.line, l.column)
	case ')':
		tok = newToken(RPAREN, l.ch, l.line, l.column)
	case '{':
		tok = newToken(LBRACE, l.ch, l.line, l.column)
	case '}':
		tok = newToken(RBRACE, l.ch, l.line, l.column)
	case '[':
		tok = newToken(LBRACKET, l.ch, l.line, l.column)
	case ']':
		tok = newToken(RBRACKET, l.ch, l.line, l.column)
	case '.':
		// 可能是 . 或 ...
		if l.peekChar() == '.' {
			// 检查是否是 ...
			l.readChar() // 读取第二个 .
			if l.peekChar() == '.' {
				// 是 ...
				l.readChar() // 读取第三个 .
				tok = Token{Type: ELLIPSIS, Literal: "...", Line: l.line, Column: l.column - 2}
			} else {
				// 只有两个点 .. 是非法的
				tok = Token{Type: ILLEGAL, Literal: "..", Line: l.line, Column: l.column - 1}
			}
		} else {
			// 单独的点号，用于成员访问运算符
			tok = newToken(DOT, l.ch, l.line, l.column)
		}
	case '$':
		// 检查是否是插值字符串 $"..."
		if l.peekChar() == '"' {
			tok.Line = l.line
			tok.Column = l.column
			l.readChar() // 跳过 $
			tok.Type = INTERP_STRING
			tok.Literal = l.readInterpolatedString()
			return tok
		}
		// 单独的 $ 是非法字符
		tok = newToken(ILLEGAL, l.ch, l.line, l.column)
	case '@':
		// 注解标记符 @
		tok = newToken(AT, l.ch, l.line, l.column)
	case '"':
		// 双引号字符串字面量
		tok.Type = STRING
		tok.Literal = l.readString('"')
		tok.Line = l.line
		tok.Column = l.column
		return tok
	case '\'':
		// 单引号字符串字面量
		tok.Type = STRING
		tok.Literal = l.readString('\'')
		tok.Line = l.line
		tok.Column = l.column
		return tok
	case '`':
		// 反引号原始字符串字面量
		tok.Type = STRING
		tok.Literal = l.readRawString()
		tok.Line = l.line
		tok.Column = l.column
		return tok
	case 0:
		// 文件结束
		tok.Literal = ""
		tok.Type = EOF
		tok.Line = l.line
		tok.Column = l.column
	default:
		// 其他字符
		if isLetter(l.ch) || l.ch == '_' {
			// 是字母或下划线，可能是标识符或关键字
			// 支持以下划线开头的标识符（如 __construct）
			tok.Literal = l.readIdentifier()
			tok.Line = l.line
			tok.Column = l.column

			// 检查是否是非法的内部函数调用
			if strings.HasPrefix(tok.Literal, "__ILLEGAL_INTERNAL_FUNC__:") {
				// 提取原始函数名
				originalName := strings.TrimPrefix(tok.Literal, "__ILLEGAL_INTERNAL_FUNC__:")
				tok.Type = ILLEGAL
				tok.Literal = "禁止使用内部函数 '" + originalName + "'，内部函数只能在标准库中使用"
				return tok
			}

			tok.Type = LookupIdent(tok.Literal)

			// 特殊处理 as? (安全类型断言)
			if tok.Type == AS && l.ch == '?' {
				l.readChar() // 跳过 ?
				tok.Type = AS_SAFE
				tok.Literal = "as?"
			}

			return tok
		} else if isDigit(l.ch) {
			// 是数字，读取整数或浮点数
			literal, isFloat := l.readNumber()
			if isFloat {
				tok.Type = FLOAT
			} else {
				tok.Type = INT
			}
			tok.Literal = literal
			tok.Line = l.line
			tok.Column = l.column
			return tok
		} else {
			// 无法识别的字符
			tok = newToken(ILLEGAL, l.ch, l.line, l.column)
		}
	}

	// 读取下一个字符，为下次调用做准备
	l.readChar()
	return tok
}

// newToken 创建新 token 的辅助函数
// 参数:
//   tokenType: token 类型
//   ch: 字符
//   line: 行号
//   column: 列号
// 返回:
//   新创建的 token
func newToken(tokenType TokenType, ch byte, line, column int) Token {
	return Token{
		Type:    tokenType,
		Literal: string(ch),
		Line:    line,
		Column:  column,
	}
}

// readIdentifier 读取标识符
// 标识符可以包含字母、数字和下划线
// 点号不再包含在标识符中，而是作为单独的 DOT token 处理
// 返回:
//   标识符字符串，如果是非法的内部函数调用则返回空字符串
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	identifier := l.input[position:l.position]

	// 检查是否是内部函数（以 __ 开头）
	// 只有标准库文件才允许使用内部函数
	if strings.HasPrefix(identifier, "__") && !l.isStdlib {
		// 返回特殊标记，表示非法使用内部函数
		return "__ILLEGAL_INTERNAL_FUNC__:" + identifier
	}

	return identifier
}

// readNumber 读取数字（整数或浮点数）
// 支持十进制、十六进制(0x)、二进制(0b)
// 返回:
//   数字字符串和是否是浮点数
func (l *Lexer) readNumber() (string, bool) {
	position := l.position
	isFloat := false

	// 检查是否是特殊进制（0x 十六进制，0b 二进制）
	if l.ch == '0' {
		next := l.peekChar()
		if next == 'x' || next == 'X' {
			// 十六进制
			l.readChar() // 跳过 0
			l.readChar() // 跳过 x
			for isHexDigit(l.ch) {
				l.readChar()
			}
			return l.input[position:l.position], false
		} else if next == 'b' || next == 'B' {
			// 二进制
			l.readChar() // 跳过 0
			l.readChar() // 跳过 b
			for l.ch == '0' || l.ch == '1' {
				l.readChar()
			}
			return l.input[position:l.position], false
		}
	}

	// 读取十进制整数部分
	for isDigit(l.ch) {
		l.readChar()
	}

	// 检查是否有小数点
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // 跳过小数点

		// 读取小数部分
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position], isFloat
}

// isHexDigit 判断字符是否是十六进制数字
func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// readString 读取字符串字面量
// 支持双引号 " 和单引号 '
// 参数:
//   quote: 引号字符（" 或 '）
// 返回:
//   字符串内容（不包含引号，转义字符已处理）
func (l *Lexer) readString(quote byte) string {
	var result []byte
	l.readChar() // 跳过开始的引号
	for l.ch != quote && l.ch != 0 {
		if l.ch == '\\' {
			// 处理转义字符
			l.readChar()
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 'r':
				result = append(result, '\r')
			case 't':
				result = append(result, '\t')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			case '\'':
				result = append(result, '\'')
			case '0':
				result = append(result, 0)
			default:
				// 未知转义序列，保留原样
				result = append(result, '\\')
				result = append(result, l.ch)
			}
		} else {
			result = append(result, l.ch)
		}
		l.readChar()
	}
	if l.ch == quote {
		// 跳过结束引号
		l.readChar()
	}
	return string(result)
}

// readRawString 读取原始字符串字面量
// 使用反引号 ` 包围，不处理转义字符
// 返回:
//   字符串内容（不包含引号）
func (l *Lexer) readRawString() string {
	position := l.position + 1 // 跳过开始的反引号
	for {
		l.readChar()
		if l.ch == '`' || l.ch == 0 {
			// 遇到结束反引号或文件结束
			break
		}
	}
	result := l.input[position:l.position]
	if l.ch == '`' {
		// 跳过结束反引号
		l.readChar()
	}
	return result
}

// readInterpolatedString 读取插值字符串字面量
// 格式：$"...{expr}..."
// 返回整个字符串内容（包括 {} 标记），由解析器负责解析插值表达式
// 支持：
//   - {{ 转义为字面 {（使用 \x01 占位符）
//   - }} 转义为字面 }（使用 \x02 占位符）
//   - 嵌套的 {} 会被正确处理（用于表达式中的 map/对象字面量）
// 占位符 \x01 和 \x02 在解析器中会被还原为 { 和 }
func (l *Lexer) readInterpolatedString() string {
	var result []byte
	l.readChar() // 跳过开始的引号 "
	
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			// 处理转义字符
			l.readChar()
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 'r':
				result = append(result, '\r')
			case 't':
				result = append(result, '\t')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			case '{':
				// \{ 转义为字面 {，使用占位符
				result = append(result, 0x01)
			case '}':
				// \} 转义为字面 }，使用占位符
				result = append(result, 0x02)
			default:
				result = append(result, '\\')
				result = append(result, l.ch)
			}
		} else if l.ch == '{' {
			// 检查是否是 {{ 转义
			if l.peekChar() == '{' {
				// {{ 转义为字面 {，使用占位符 \x01
				result = append(result, 0x01)
				l.readChar() // 跳过第二个 {
			} else {
				// 插值开始标记，保留 { 让解析器处理
				result = append(result, l.ch)
				// 读取直到匹配的 }，处理嵌套
				braceCount := 1
				l.readChar()
				for braceCount > 0 && l.ch != 0 {
					if l.ch == '{' {
						braceCount++
					} else if l.ch == '}' {
						braceCount--
					} else if l.ch == '"' {
						// 处理表达式中的字符串
						result = append(result, l.ch)
						l.readChar()
						for l.ch != '"' && l.ch != 0 {
							if l.ch == '\\' {
								result = append(result, l.ch)
								l.readChar()
							}
							result = append(result, l.ch)
							l.readChar()
						}
					}
					result = append(result, l.ch)
					if braceCount > 0 {
						l.readChar()
					}
				}
			}
		} else if l.ch == '}' {
			// 检查是否是 }} 转义
			if l.peekChar() == '}' {
				// }} 转义为字面 }，使用占位符 \x02
				result = append(result, 0x02)
				l.readChar() // 跳过第二个 }
			} else {
				// 单独的 } 保留
				result = append(result, l.ch)
			}
		} else {
			result = append(result, l.ch)
		}
		l.readChar()
	}
	
	if l.ch == '"' {
		l.readChar() // 跳过结束引号
	}
	
	return string(result)
}

// skipWhitespace 跳过空白字符
// 空白字符包括：空格、制表符、换行符、回车符
// 这些字符在词法分析中通常被忽略
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// skipComment 跳过注释
// 支持单行注释，从 // 开始到行尾
// 注释内容会被完全忽略
// skipLineComment 跳过单行注释
// 从 // 开始到行尾
func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	if l.ch == '\n' {
		l.readChar()
	}
}

// skipBlockComment 跳过块注释
// 从 /* 开始到 */ 结束，支持多行
func (l *Lexer) skipBlockComment() {
	// 跳过 /*
	l.readChar() // 跳过 /
	l.readChar() // 跳过 *

	for {
		if l.ch == 0 {
			// 文件结束但没有找到 */，报错或忽略
			break
		}
		if l.ch == '*' && l.peekChar() == '/' {
			// 找到 */，跳过并退出
			l.readChar() // 跳过 *
			l.readChar() // 跳过 /
			break
		}
		l.readChar()
	}
}

// isLetter 判断字符是否是字母
// 支持 Unicode 字母
// 参数:
//   ch: 要检查的字符
// 返回:
//   如果是字母返回 true，否则返回 false
func isLetter(ch byte) bool {
	r, _ := utf8.DecodeRune([]byte{ch})
	return unicode.IsLetter(r)
}

// isDigit 判断字符是否是数字
// 参数:
//   ch: 要检查的字符
// 返回:
//   如果是数字返回 true，否则返回 false
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// contains 检查字符串是否包含指定字符
// 参数:
//   s: 要检查的字符串
//   ch: 要查找的字符
// 返回:
//   如果包含返回 true，否则返回 false
func contains(s string, ch byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == ch {
			return true
		}
	}
	return false
}

