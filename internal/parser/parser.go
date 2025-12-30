package parser

import (
	"fmt"
	"strconv"

	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== 运算符优先级常量 ==========
// 优先级从低到高，数值越大优先级越高
// 用于确定表达式的解析顺序（如 1 + 2 * 3 应该解析为 1 + (2 * 3)）
const (
	_           int = iota
	LOWEST          // 最低优先级（用于初始化）
	ASSIGNMENT      // 赋值运算符优先级：:=, =
	CONDITIONAL     // 三目运算符优先级：? :
	OR              // 逻辑或优先级：||
	AND             // 逻辑与优先级：&&
	EQUALS          // 相等比较优先级：==, !=
	LESSGREATER     // 大小比较优先级：<, >, <=, >=
	SUM             // 加减法优先级：+, -
	PRODUCT         // 乘除法优先级：*, /, %
	PREFIX          // 前缀运算符优先级：!, -（负号）
	CALL            // 函数调用优先级：()
	INDEX           // 索引访问优先级：[]（未实现）
	TYPEASSERT      // 类型断言优先级：.(type)（未实现）
)

// precedences 运算符优先级映射表
// 将 token 类型映射到对应的优先级
// 用于在解析表达式时确定运算符的优先级
var precedences = map[lexer.TokenType]int{
	lexer.OR:           OR,
	lexer.AND:          AND,
	lexer.EQ:           EQUALS,
	lexer.NOT_EQ:       EQUALS,
	lexer.LT:           LESSGREATER,
	lexer.GT:           LESSGREATER,
	lexer.LE:           LESSGREATER,
	lexer.GE:           LESSGREATER,
	lexer.PLUS:         SUM,
	lexer.MINUS:        SUM,
	lexer.SLASH:        PRODUCT,
	lexer.ASTERISK:     PRODUCT,
	lexer.MOD:          PRODUCT,
	lexer.LPAREN:       CALL,        // 左括号用于函数调用
	lexer.QUESTION:     CONDITIONAL, // 问号用于三目运算符
	lexer.ASSIGN:       ASSIGNMENT,  // 赋值运算符
	lexer.DOT:          CALL,        // 点号用于成员访问（优先级与函数调用相同）
	lexer.DOUBLE_COLON: CALL,        // :: 用于静态方法调用（优先级与函数调用相同）
}

// ========== 语法分析器结构 ==========

// Parser 语法分析器
// 负责将 token 流转换为抽象语法树（AST）
// 使用递归下降解析算法
type Parser struct {
	l      *lexer.Lexer // 词法分析器，用于获取 token
	errors []string     // 解析过程中收集的错误信息

	// prevToken: 用于在需要时获取“当前 token 的前一个 token”（例如三目运算符格式校验）
	prevToken lexer.Token
	curToken  lexer.Token // 当前正在处理的 token
	peekToken lexer.Token // 下一个 token（用于前瞻）

	// allowTernary: 是否允许解析三目运算符（用于限制其出现在函数/方法参数中）
	allowTernary bool

	// 前缀解析函数映射表
	// 用于解析前缀表达式（如 !true, -5）
	prefixParseFns map[lexer.TokenType]prefixParseFn
	// 中缀解析函数映射表
	// 用于解析中缀表达式（如 1 + 2, a == b）
	infixParseFns map[lexer.TokenType]infixParseFn
}

// prefixParseFn 前缀解析函数类型
// 用于解析前缀表达式（运算符在操作数前面）
// 例如：!true, -5
type prefixParseFn func() Expression

// infixParseFn 中缀解析函数类型
// 用于解析中缀表达式（运算符在两个操作数之间）
// 例如：1 + 2, a == b
// 参数 left 是已经解析好的左操作数
type infixParseFn func(Expression) Expression

// New 创建新的语法分析器
// 参数:
//
//	l: 词法分析器实例
//
// 返回:
//
//	初始化好的 Parser 实例
//
// 功能:
//  1. 注册所有前缀和中缀解析函数
//  2. 初始化 curToken 和 peekToken
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.allowTernary = true

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TRUE, p.parseBoolean)
	p.registerPrefix(lexer.FALSE, p.parseBoolean)
	p.registerPrefix(lexer.NULL, p.parseNull)
	p.registerPrefix(lexer.BANG, p.parsePrefixExpression)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(lexer.THIS, p.parseThisExpression)
	p.registerPrefix(lexer.NEW, p.parseNewExpression)

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.ASTERISK, p.parseInfixExpression)
	p.registerInfix(lexer.MOD, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.GT, p.parseInfixExpression)
	p.registerInfix(lexer.LE, p.parseInfixExpression)
	p.registerInfix(lexer.GE, p.parseInfixExpression)
	p.registerInfix(lexer.AND, p.parseInfixExpression)
	p.registerInfix(lexer.OR, p.parseInfixExpression)
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.QUESTION, p.parseTernaryExpression)
	p.registerInfix(lexer.DOT, p.parseMemberAccessExpression)        // 成员访问 object.member
	p.registerInfix(lexer.ASSIGN, p.parseAssignmentExpression)       // 赋值 a = b
	p.registerInfix(lexer.DOUBLE_COLON, p.parseStaticCallExpression) // 静态调用 Class::method

	// 读取两个 token，初始化 curToken 和 peekToken
	// 这样我们就可以使用前瞻（lookahead）来辅助解析
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken 读取下一个 token
// 将 peekToken 移动到 curToken，然后从词法分析器读取新的 peekToken
// 这样实现了双 token 前瞻，可以处理需要查看下一个 token 的情况
func (p *Parser) nextToken() {
	p.prevToken = p.curToken
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseProgram 解析整个程序
// 这是语法分析的入口函数
// 返回:
//
//	解析后的程序 AST 根节点
//
// 功能:
//
//	遍历所有 token，解析为语句，直到遇到 EOF
func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = []Statement{}

	for p.curToken.Type != lexer.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// parseStatement 解析语句
// 根据当前 token 的类型，调用相应的语句解析函数
// 返回:
//
//	解析后的语句节点，如果解析失败返回 nil
func (p *Parser) parseStatement() Statement {
	// 跳过分号
	for p.curTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	switch p.curToken.Type {
	case lexer.PACKAGE:
		return p.parsePackageStatement()
	case lexer.IMPORT:
		return p.parseImportStatement()
	case lexer.CLASS:
		return p.parseClassStatement()
	case lexer.VAR:
		return p.parseLetStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.BREAK:
		return p.parseBreakStatement()
	case lexer.CONTINUE:
		return p.parseContinueStatement()
	case lexer.FUNCTION:
		return p.parseFunctionStatement()
	case lexer.RBRACE:
		// 如果在 parseStatement 中遇到 }，说明 block 可能解析有问题或者有多余的 }
		// 这里不报错，返回 nil，由调用者处理
		return nil
	case lexer.EOF:
		return nil
	case lexer.ILLEGAL:
		// 遇到非法字符，记录错误并跳过
		msg := fmt.Sprintf("非法字符: %s (行 %d, 列 %d)", p.curToken.Literal, p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	default:
		return p.parseExpressionStatement()
	}
}

// parseLetStatement 解析变量声明语句
func (p *Parser) parseLetStatement() Statement {
	stmt := &LetStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 可选类型声明
	if p.peekTokenIs(lexer.STRING_TYPE) || p.peekTokenIs(lexer.INT_TYPE) || p.peekTokenIs(lexer.BOOL_TYPE) || p.peekTokenIs(lexer.ANY) {
		p.nextToken()
		stmt.Type = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	// 赋值
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // 跳过 =
		p.nextToken() // 跳过赋值符号
		stmt.Value = p.parseExpression(LOWEST)
	}

	// 短变量声明 :=
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // 跳过 :
		if p.peekTokenIs(lexer.ASSIGN) {
			p.nextToken() // 跳过 =
			assignStmt := &AssignStatement{
				Token: lexer.Token{Type: lexer.ASSIGN, Literal: ":="},
				Name:  stmt.Name,
			}
			assignStmt.Value = p.parseExpression(LOWEST)
			return assignStmt
		}
	}

	return stmt
}

// parseReturnStatement 解析返回语句
func (p *Parser) parseReturnStatement() *ReturnStatement {
	stmt := &ReturnStatement{Token: p.curToken}

	// 如果后面紧跟的是 } 或 ;，说明没有返回值
	if p.peekTokenIs(lexer.RBRACE) || p.peekTokenIs(lexer.SEMICOLON) {
		return stmt
	}

	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)

	return stmt
}

// parseExpressionStatement 解析表达式语句
func (p *Parser) parseExpressionStatement() Statement {
	// 检查是否是短变量声明 :=
	if p.curToken.Type == lexer.IDENT && p.peekTokenIs(lexer.ASSIGN) && p.peekToken.Literal == ":=" {
		// 短变量声明：name := value
		name := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken() // 跳过 :=
		p.nextToken() // 跳过 ASSIGN token
		assignStmt := &AssignStatement{
			Token: lexer.Token{Type: lexer.ASSIGN, Literal: ":="},
			Name:  name,
		}
		assignStmt.Value = p.parseExpression(LOWEST)
		return assignStmt
	}

	// 检查是否是自增/自减语句 i++ 或 i--
	if p.curToken.Type == lexer.IDENT && (p.peekTokenIs(lexer.INCREMENT) || p.peekTokenIs(lexer.DECREMENT)) {
		name := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken() // 移动到 ++ 或 --
		return &IncrementStatement{
			Token:    p.curToken,
			Name:     name,
			Operator: p.curToken.Literal,
		}
	}

	// 普通表达式语句
	stmt := &ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	return stmt
}

// parseExpression 解析表达式
func (p *Parser) parseExpression(precedence int) Expression {
	// 如果遇到非法字符，直接返回 nil
	if p.curToken.Type == lexer.ILLEGAL {
		msg := fmt.Sprintf("非法字符: %s (行 %d, 列 %d)", p.curToken.Literal, p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	}

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// parseIdentifier 解析标识符
func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral 解析整数字面量
func (p *Parser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("无法将 %q 解析为整数", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// parseStringLiteral 解析字符串字面量
func (p *Parser) parseStringLiteral() Expression {
	return &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parseBoolean 解析布尔字面量
func (p *Parser) parseBoolean() Expression {
	return &BooleanLiteral{Token: p.curToken, Value: p.curToken.Type == lexer.TRUE}
}

// parseNull 解析 null 字面量
func (p *Parser) parseNull() Expression {
	return &NullLiteral{Token: p.curToken}
}

// parsePrefixExpression 解析前缀表达式
func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseInfixExpression 解析中缀表达式
func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseGroupedExpression 解析分组表达式
func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return exp
}

// parseIfStatement 解析 if 语句
func (p *Parser) parseIfStatement() *IfStatement {
	stmt := &IfStatement{Token: p.curToken}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken()
		if p.peekTokenIs(lexer.IF) {
			// else if
			p.nextToken()
			stmt.ElseIf = p.parseIfStatement()
		} else {
			// else
			if !p.expectPeek(lexer.LBRACE) {
				return nil
			}
			stmt.Alternative = p.parseBlockStatement()
		}
	}

	return stmt
}

// parseForStatement 解析 for 循环语句
// 支持三种形式：
// 1. for condition { ... }           - while 式循环
// 2. for { ... }                     - 无限循环
// 3. for init; condition; post { ... } - 传统 for 循环
func (p *Parser) parseForStatement() *ForStatement {
	stmt := &ForStatement{Token: p.curToken}

	p.nextToken() // 跳过 'for'

	// 检查是否是无限循环：for { ... }
	if p.curTokenIs(lexer.LBRACE) {
		stmt.Body = p.parseBlockStatement()
		return stmt
	}

	// 检查是否是传统 for 循环：for init; condition; post { ... }
	// 判断方法：看第一个标识符后是否是 :=（初始化语句）
	if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.ASSIGN) && p.peekToken.Literal == ":=" {
		// 传统 for 循环：for j := 0; j < 5; j++ { ... }
		stmt.Init = p.parseForInit()

		if !p.curTokenIs(lexer.SEMICOLON) {
			p.errors = append(p.errors, fmt.Sprintf("for 循环缺少第一个分号 (行 %d)", p.curToken.Line))
			return nil
		}
		p.nextToken() // 跳过第一个分号

		// 解析条件
		if !p.curTokenIs(lexer.SEMICOLON) {
			stmt.Condition = p.parseExpression(LOWEST)
			p.nextToken()
		}

		if !p.curTokenIs(lexer.SEMICOLON) {
			p.errors = append(p.errors, fmt.Sprintf("for 循环缺少第二个分号 (行 %d)", p.curToken.Line))
			return nil
		}
		p.nextToken() // 跳过第二个分号

		// 解析 post 语句（如 j++）
		if !p.curTokenIs(lexer.LBRACE) {
			stmt.Post = p.parseForPost()
		}
	} else {
		// while 式循环：for condition { ... }
		stmt.Condition = p.parseExpression(LOWEST)
		p.nextToken()
	}

	// 解析循环体
	if !p.curTokenIs(lexer.LBRACE) {
		p.errors = append(p.errors, fmt.Sprintf("期望 '{{' 但得到 %s (行 %d)", p.curToken.Literal, p.curToken.Line))
		return nil
	}
	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseForInit 解析 for 循环的初始化语句（如 j := 0）
func (p *Parser) parseForInit() Statement {
	name := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken() // 跳过变量名
	p.nextToken() // 跳过 :=
	assignStmt := &AssignStatement{
		Token: lexer.Token{Type: lexer.ASSIGN, Literal: ":="},
		Name:  name,
	}
	assignStmt.Value = p.parseExpression(LOWEST)
	p.nextToken() // 移动到分号
	return assignStmt
}

// parseForPost 解析 for 循环的 post 语句（如 j++）
func (p *Parser) parseForPost() Statement {
	// 检查是否是 i++ 或 i--
	if p.curTokenIs(lexer.IDENT) && (p.peekTokenIs(lexer.INCREMENT) || p.peekTokenIs(lexer.DECREMENT)) {
		name := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken() // 移动到 ++ 或 --
		stmt := &IncrementStatement{
			Token:    p.curToken,
			Name:     name,
			Operator: p.curToken.Literal,
		}
		p.nextToken() // 跳过 ++ 或 --，移动到 {
		return stmt
	}

	// 普通表达式
	stmt := &ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	p.nextToken()
	return stmt
}

// parseBreakStatement 解析 break 语句
func (p *Parser) parseBreakStatement() *BreakStatement {
	return &BreakStatement{Token: p.curToken}
}

// parseContinueStatement 解析 continue 语句
func (p *Parser) parseContinueStatement() *ContinueStatement {
	return &ContinueStatement{Token: p.curToken}
}

// parseBlockStatement 解析块语句
func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{Token: p.curToken}
	block.Statements = []Statement{}

	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		// 移动到下一个 token。
		// 不能在 curToken==RBRACE 时停住：因为嵌套块（如 if {...}）解析完后 curToken 也会是 RBRACE，
		// 如果这里不推进，会把“内部块的 }”误当成“外层块结束”，导致外层块提前结束。
		if !p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}
	}

	// 解析完成后，curToken 是 RBRACE，需要移动到 RBRACE 之后
	// 但这里不移动，由调用者决定是否移动

	return block
}

// parseFunctionLiteral 解析函数字面量
func (p *Parser) parseFunctionLiteral() Expression {
	lit := &FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	// 保存函数名
	lit.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	// 解析返回类型
	// 支持两种语法：
	// 1. fn name():type   - 使用冒号
	// 2. fn name() type   - 不使用冒号（直接跟类型）

	// 检查是否有冒号
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // 移动到 :
		p.nextToken() // 移动到返回类型
		if p.curTokenIs(lexer.LPAREN) {
			// 多返回值
			p.nextToken()
			lit.ReturnType = []*Identifier{}
			for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
				if p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.ANY) || p.curTokenIs(lexer.IDENT) {
					lit.ReturnType = append(lit.ReturnType, &Identifier{Token: p.curToken, Value: p.curToken.Literal})
				}
				p.nextToken()
				if p.curTokenIs(lexer.COMMA) {
					p.nextToken()
				}
			}
		} else {
			// 单返回值
			if p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.ANY) || p.curTokenIs(lexer.VOID) || p.curTokenIs(lexer.IDENT) {
				lit.ReturnType = []*Identifier{{Token: p.curToken, Value: p.curToken.Literal}}
			}
		}
	} else if p.peekTokenIs(lexer.STRING_TYPE) || p.peekTokenIs(lexer.INT_TYPE) || p.peekTokenIs(lexer.BOOL_TYPE) || p.peekTokenIs(lexer.ANY) || p.peekTokenIs(lexer.VOID) || p.peekTokenIs(lexer.IDENT) {
		// 不使用冒号的语法 fn name() type
		p.nextToken() // 移动到返回类型
		lit.ReturnType = []*Identifier{{Token: p.curToken, Value: p.curToken.Literal}}
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// parseFunctionStatement 解析函数声明语句
func (p *Parser) parseFunctionStatement() *ExpressionStatement {
	stmt := &ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseFunctionLiteral()
	return stmt
}

// parseFunctionParameters 解析函数参数
func (p *Parser) parseFunctionParameters() []*FunctionParameter {
	parameters := []*FunctionParameter{}

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return parameters
	}

	p.nextToken()

	param := &FunctionParameter{}
	param.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
		p.nextToken()
		param.Type = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken()
		p.nextToken()
		param.DefaultValue = p.parseExpression(LOWEST)
		// parseExpression 解析完成后，curToken 在表达式之后
		// 如果表达式后面是 , 或 )，parseExpression 已经停止在正确位置
	}

	parameters = append(parameters, param)

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()

		param := &FunctionParameter{}
		param.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if p.peekTokenIs(lexer.COLON) {
			p.nextToken()
			p.nextToken()
			param.Type = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		}

		if p.peekTokenIs(lexer.ASSIGN) {
			p.nextToken()
			p.nextToken()
			param.DefaultValue = p.parseExpression(LOWEST)
			// parseExpression 解析完成后，curToken 在表达式之后
			// 如果表达式后面是 , 或 )，parseExpression 已经停止在正确位置
		}

		parameters = append(parameters, param)
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return parameters
}

// parseCallExpression 解析函数调用表达式
func (p *Parser) parseCallExpression(function Expression) Expression {
	exp := &CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

// parseCallArguments 解析函数调用参数
func (p *Parser) parseCallArguments() []CallArgument {
	args := []CallArgument{}

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return args
	}

	// 限制：调用参数内不允许三目运算符（包含命名参数的 value）
	oldAllowTernary := p.allowTernary
	p.allowTernary = false
	defer func() { p.allowTernary = oldAllowTernary }()

	p.nextToken()

	arg := CallArgument{}
	// 检查是否是命名参数
	if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.COLON) {
		arg.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken() // 跳过 :
		p.nextToken()
	}
	arg.Value = p.parseExpression(LOWEST)
	args = append(args, arg)

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()

		arg := CallArgument{}
		if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.COLON) {
			arg.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
			p.nextToken() // 跳过 :
			p.nextToken()
		}
		arg.Value = p.parseExpression(LOWEST)
		args = append(args, arg)
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return args
}

// parseTernaryExpression 解析三目运算符表达式
func (p *Parser) parseTernaryExpression(condition Expression) Expression {
	exp := &TernaryExpression{
		Token:     p.curToken,
		Condition: condition,
	}

	// 限制：三目运算符不能作为函数/方法/构造调用的参数
	if !p.allowTernary {
		p.errors = append(p.errors, fmt.Sprintf("三目运算符不能作为函数/方法参数使用 (行 %d)", p.curToken.Line))
	}

	// 格式限制：
	// - 允许单行：cond ? a : b（? 与 : 与各分支都在同一行）
	// - 允许多行：cond \n ? a \n : b（? 与 : 必须各自提到新的一行）
	// - 禁止：cond \n ? a : b（? 提行但 : 没提行）
	condEndLine := p.prevToken.Line   // ? 之前的 token 所在行（近似 condition 结束行）
	questionLine := p.curToken.Line   // ? 所在行

	p.nextToken()
	exp.TrueExpr = p.parseExpression(CONDITIONAL)

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	trueEndLine := p.prevToken.Line   // ':' 之前的 token 所在行（近似 trueExpr 结束行）
	colonLine := p.curToken.Line      // ':' 所在行

	// 单行：? 与 condition 同行，则 : 也必须同行
	if questionLine == condEndLine {
		if colonLine != questionLine {
			p.errors = append(p.errors, fmt.Sprintf("三目运算符单行写法要求 '?' 和 ':' 在同一行 (行 %d)", questionLine))
		}
	} else {
		// 多行：? 必须提行，且 : 也必须提行（不能与 trueExpr 同行）
		if questionLine == trueEndLine {
			// 允许 ? true（trueExpr 与 ? 同行）
		}
		if colonLine == trueEndLine {
			p.errors = append(p.errors, fmt.Sprintf("三目运算符多行写法要求 ':' 单独换行，禁止 '? true : false' 这种混写 (行 %d)", colonLine))
		}
	}

	p.nextToken()
	exp.FalseExpr = p.parseExpression(CONDITIONAL)

	return exp
}

// 辅助函数

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("期望下一个 token 是 %s，但得到 %s", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("没有找到 %s 的前缀解析函数", t)
	p.errors = append(p.errors, msg)
}

// ========== 包和导入解析 ==========

// parsePackageStatement 解析包声明语句
// 对应语法：package packageName
func (p *Parser) parsePackageStatement() *PackageStatement {
	stmt := &PackageStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return stmt
}

// parseImportStatement 解析导入语句
// 对应语法：import "package/path"
func (p *Parser) parseImportStatement() *ImportStatement {
	stmt := &ImportStatement{Token: p.curToken}

	if !p.expectPeek(lexer.STRING) {
		return nil
	}

	stmt.Path = &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}

	return stmt
}

// ========== 类解析 ==========

// parseClassStatement 解析类声明语句
// 对应语法：class ClassName { ... }
func (p *Parser) parseClassStatement() *ClassStatement {
	stmt := &ClassStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	// 解析类成员
	stmt.Members = p.parseClassMembers()

	return stmt
}

// parseClassMembers 解析类成员
func (p *Parser) parseClassMembers() []ClassMember {
	members := []ClassMember{}

	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// 跳过空白字符、注释等
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// 解析访问修饰符
		if !p.curTokenIs(lexer.PUBLIC) && !p.curTokenIs(lexer.PRIVATE) && !p.curTokenIs(lexer.PROTECTED) {
			// 如果不是访问修饰符，可能是注释、空行等，跳过
			p.nextToken()
			continue
		}

		accessModifier := p.curToken.Literal
		p.nextToken()

		// 检查是否是静态方法
		isStatic := false
		if p.curTokenIs(lexer.STATIC) {
			isStatic = true
			p.nextToken()
		}

		// 检查是否是方法
		if p.curTokenIs(lexer.FUNCTION) {
			method := p.parseClassMethod(accessModifier, isStatic)
			if method != nil {
				members = append(members, method)
				// parseClassMethod 结束时 curToken 在方法体的 }
				if p.curTokenIs(lexer.RBRACE) {
					p.nextToken()
				}
			} else {
				// 解析失败，跳过直到下一个访问修饰符或类结束
				for !p.curTokenIs(lexer.EOF) &&
					!p.curTokenIs(lexer.RBRACE) &&
					!p.curTokenIs(lexer.PUBLIC) &&
					!p.curTokenIs(lexer.PRIVATE) &&
					!p.curTokenIs(lexer.PROTECTED) {
					p.nextToken()
				}
			}
		} else if p.curTokenIs(lexer.IDENT) {
			// 是成员变量（变量名）
			variable := p.parseClassVariable(accessModifier)
			if variable != nil {
				members = append(members, variable)
				// 变量解析完成后，移动到下一个 token
				// parseClassVariable 解析完成后，curToken 已经在变量声明之后
				// 如果变量有初始值，parseExpression 已经移动了 token
				// 如果没有初始值，curToken 已经在类型之后，需要移动到下一个 token
				// 但需要检查是否已经是 RBRACE 或下一个访问修饰符
				if !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.PUBLIC) && !p.curTokenIs(lexer.PRIVATE) && !p.curTokenIs(lexer.PROTECTED) && !p.curTokenIs(lexer.EOF) {
					p.nextToken()
				}
			} else {
				// parseClassVariable 返回 nil，说明解析失败
				// 不要继续移动 token，让循环继续处理下一个成员
				// 但需要跳过当前错误的 token，避免无限循环
				if !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.PUBLIC) && !p.curTokenIs(lexer.PRIVATE) && !p.curTokenIs(lexer.PROTECTED) && !p.curTokenIs(lexer.EOF) {
					p.nextToken()
				}
			}
		} else {
			// 既不是方法也不是变量，可能是语法错误
			p.errors = append(p.errors, fmt.Sprintf("类成员必须是方法或变量，得到 %s (行 %d)", p.curToken.Type, p.curToken.Line))
			p.nextToken()
		}
	}

	return members
}

// parseClassVariable 解析类成员变量
func (p *Parser) parseClassVariable(accessModifier string) *ClassVariable {
	variable := &ClassVariable{
		Token:          p.curToken,
		AccessModifier: accessModifier,
	}

	if !p.curTokenIs(lexer.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("期望变量名，得到 %s (行 %d)", p.curToken.Type, p.curToken.Line))
		return nil
	}

	variable.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken()

	// 解析类型
	if !p.curTokenIs(lexer.STRING_TYPE) && !p.curTokenIs(lexer.INT_TYPE) && !p.curTokenIs(lexer.BOOL_TYPE) && !p.curTokenIs(lexer.ANY) {
		p.errors = append(p.errors, fmt.Sprintf("类成员变量必须声明类型，得到 %s (行 %d)", p.curToken.Type, p.curToken.Line))
		// 返回 nil，curToken 在类型位置，由调用者处理
		return nil
	}

	variable.Type = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken()

	// 检查是否有初始值
	if p.curTokenIs(lexer.ASSIGN) {
		p.nextToken()
		variable.Value = p.parseExpression(LOWEST)
		// parseExpression 解析完成后，curToken 已经在表达式之后
	}

	return variable
}

// parseClassMethod 解析类方法
func (p *Parser) parseClassMethod(accessModifier string, isStatic bool) *ClassMethod {
	method := &ClassMethod{
		Token:          p.curToken,
		AccessModifier: accessModifier,
		IsStatic:       isStatic,
	}

	if !p.curTokenIs(lexer.FUNCTION) {
		p.errors = append(p.errors, fmt.Sprintf("期望 function 关键字 (行 %d)", p.curToken.Line))
		return nil
	}

	p.nextToken()

	if !p.curTokenIs(lexer.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("期望方法名 (行 %d)", p.curToken.Line))
		return nil
	}

	method.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	method.Parameters = p.parseFunctionParameters()
	// parseFunctionParameters 解析完成后，curToken 在 )

	// 解析返回类型
	// 支持两种语法：
	// 1. method():type   - 使用冒号
	// 2. method() type   - 不使用冒号（直接跟类型）

	// 解析返回类型
	// 支持两种语法：
	// 1. method():type   - 使用冒号
	// 2. method() type   - 不使用冒号（直接跟类型）

	// 检查是否有冒号
	if p.peekTokenIs(lexer.COLON) {
		// 使用冒号的语法 method():type
		p.nextToken() // 移动到 :
		p.nextToken() // 移动到返回类型

		if p.curTokenIs(lexer.LPAREN) {
			// 多返回值 (type1, type2)
			p.nextToken()
			method.ReturnType = []*Identifier{}
			for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
				if p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.ANY) || p.curTokenIs(lexer.IDENT) {
					method.ReturnType = append(method.ReturnType, &Identifier{Token: p.curToken, Value: p.curToken.Literal})
				}
				p.nextToken()
				if p.curTokenIs(lexer.COMMA) {
					p.nextToken()
				}
			}
		} else {
			// 单返回值
			if p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.ANY) || p.curTokenIs(lexer.VOID) || p.curTokenIs(lexer.IDENT) {
				method.ReturnType = []*Identifier{{Token: p.curToken, Value: p.curToken.Literal}}
			}
		}
	} else if p.peekTokenIs(lexer.STRING_TYPE) || p.peekTokenIs(lexer.INT_TYPE) || p.peekTokenIs(lexer.BOOL_TYPE) || p.peekTokenIs(lexer.ANY) || p.peekTokenIs(lexer.VOID) || p.peekTokenIs(lexer.IDENT) {
		// 不使用冒号的语法 method() type
		p.nextToken() // 移动到返回类型
		method.ReturnType = []*Identifier{{Token: p.curToken, Value: p.curToken.Literal}}
	}

	// 现在期望下一个 token 是 {
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	method.Body = p.parseBlockStatement()
	// parseBlockStatement 解析完成后，curToken 是 RBRACE
	// 不在这里移动 token，让调用者 parseClassMembers 来处理

	return method
}

// parseThisExpression 解析 this 表达式
func (p *Parser) parseThisExpression() Expression {
	return &ThisExpression{Token: p.curToken}
}

// parseNewExpression 解析 new 表达式
// 对应语法：new ClassName(参数)
func (p *Parser) parseNewExpression() Expression {
	exp := &NewExpression{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	exp.ClassName = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	exp.Arguments = p.parseCallArguments()

	return exp
}

// parseStaticCallExpression 解析静态方法调用表达式
// 对应语法：ClassName::methodName(参数)
func (p *Parser) parseStaticCallExpression(left Expression) Expression {
	// left 应该是类名
	className, ok := left.(*Identifier)
	if !ok {
		p.errors = append(p.errors, fmt.Sprintf("静态方法调用左侧必须是类名 (行 %d)", p.curToken.Line))
		return nil
	}

	exp := &StaticCallExpression{
		Token:     p.curToken,
		ClassName: className,
	}

	p.nextToken()

	if !p.curTokenIs(lexer.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("期望方法名 (行 %d)", p.curToken.Line))
		return nil
	}

	exp.Method = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	exp.Arguments = p.parseCallArguments()

	return exp
}

// parseMemberAccessExpression 解析成员访问表达式
// 对应语法：object.member
func (p *Parser) parseMemberAccessExpression(left Expression) Expression {
	exp := &MemberAccessExpression{
		Token:  p.curToken,
		Object: left,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	if !p.curTokenIs(lexer.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("期望成员名 (行 %d)", p.curToken.Line))
		return nil
	}

	exp.Member = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 检查是否是方法调用 object.method()
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // 移动到 (
		call := &CallExpression{
			Token:    p.curToken,
			Function: exp,
		}
		call.Arguments = p.parseCallArguments()
		return call
	}

	// 检查是否是链式成员访问 object.member.member2
	if p.peekTokenIs(lexer.DOT) {
		p.nextToken() // 移动到 .
		return p.parseMemberAccessExpression(exp)
	}

	// 检查是否是赋值 object.member = value
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // 移动到 =
		return p.parseAssignmentExpression(exp)
	}

	_ = precedence // 保留变量以供将来使用
	return exp
}

// parseAssignmentExpression 解析赋值表达式
// 对应语法：left = right
func (p *Parser) parseAssignmentExpression(left Expression) Expression {
	exp := &AssignmentExpression{
		Token: p.curToken,
		Left:  left,
	}

	p.nextToken() // 移动到赋值运算符之后
	exp.Right = p.parseExpression(LOWEST)

	return exp
}
