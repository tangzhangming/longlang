package lexer

import (
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
}

// New 创建新的词法分析器
// 参数:
//   input: 要分析的源代码字符串
// 返回:
//   初始化好的 Lexer 实例
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1, // 行号从1开始
		column: 1, // 列号从1开始
	}
	l.readChar() // 读取第一个字符
	return l
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
		// 可能是 = 或 ==
		if l.peekChar() == '=' {
			// 是 ==
			ch := l.ch
			l.readChar()
			tok = Token{Type: EQ, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			// 是 =
			tok = newToken(ASSIGN, l.ch, l.line, l.column)
		}
	case '+':
		// 可能是 + 或 ++
		if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: INCREMENT, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			tok = newToken(PLUS, l.ch, l.line, l.column)
		}
	case '-':
		// 可能是 - 或 --
		if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: DECREMENT, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
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
		tok = newToken(ASTERISK, l.ch, l.line, l.column)
	case '/':
		// 可能是 / 或注释 //
		if l.peekChar() == '/' {
			// 是注释，跳过注释内容
			l.skipComment()
			// 递归调用，返回注释后的下一个 token
			return l.NextToken()
		}
		// 是除法运算符
		tok = newToken(SLASH, l.ch, l.line, l.column)
	case '%':
		tok = newToken(MOD, l.ch, l.line, l.column)
	case '<':
		// 可能是 < 或 <=
		if l.peekChar() == '=' {
			// 是 <=
			ch := l.ch
			l.readChar()
			tok = Token{Type: LE, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			// 是 <
			tok = newToken(LT, l.ch, l.line, l.column)
		}
	case '>':
		// 可能是 > 或 >=
		if l.peekChar() == '=' {
			// 是 >=
			ch := l.ch
			l.readChar()
			tok = Token{Type: GE, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			// 是 >
			tok = newToken(GT, l.ch, l.line, l.column)
		}
	case '&':
		// 可能是 & 或 &&
		if l.peekChar() == '&' {
			// 是 &&
			ch := l.ch
			l.readChar()
			tok = Token{Type: AND, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			// 单独的 & 是非法字符
			tok = newToken(ILLEGAL, l.ch, l.line, l.column)
		}
	case '|':
		// 可能是 | 或 ||
		if l.peekChar() == '|' {
			// 是 ||
			ch := l.ch
			l.readChar()
			tok = Token{Type: OR, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			// 单独的 | 是非法字符
			tok = newToken(ILLEGAL, l.ch, l.line, l.column)
		}
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
		// 点号用于成员访问（如 object.member）
		// 单独的点号，用于成员访问运算符
		tok = newToken(DOT, l.ch, l.line, l.column)
	case '"':
		// 字符串字面量
		tok.Type = STRING
		tok.Literal = l.readString()
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
			tok.Type = LookupIdent(tok.Literal)
			tok.Line = l.line
			tok.Column = l.column
			return tok
		} else if isDigit(l.ch) {
			// 是数字，读取整数
			tok.Type = INT
			tok.Literal = l.readNumber()
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
//   标识符字符串
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber 读取数字
// 目前只支持整数，不支持小数
// 返回:
//   数字字符串
func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readString 读取字符串字面量
// 从 " 开始，到 " 结束
// 返回:
//   字符串内容（不包含引号）
func (l *Lexer) readString() string {
	position := l.position + 1 // 跳过开始的引号
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			// 遇到结束引号或文件结束
			break
		}
	}
	if l.ch == '"' {
		// 跳过结束引号
		l.readChar()
	}
	return l.input[position : l.position-1]
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
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	if l.ch == '\n' {
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
