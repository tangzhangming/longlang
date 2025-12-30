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
	_ int = iota
	LOWEST      // 最低优先级（用于初始化）
	ASSIGNMENT  // 赋值运算符优先级：:=, =
	CONDITIONAL // 三目运算符优先级：? :
	OR          // 逻辑或优先级：||
	AND         // 逻辑与优先级：&&
	EQUALS      // 相等比较优先级：==, !=
	LESSGREATER // 大小比较优先级：<, >, <=, >=
	SUM         // 加减法优先级：+, -
	PRODUCT     // 乘除法优先级：*, /, %
	PREFIX      // 前缀运算符优先级：!, -（负号）
	CALL        // 函数调用优先级：()
	INDEX       // 索引访问优先级：[]（未实现）
	TYPEASSERT  // 类型断言优先级：.(type)（未实现）
)

// precedences 运算符优先级映射表
// 将 token 类型映射到对应的优先级
// 用于在解析表达式时确定运算符的优先级
var precedences = map[lexer.TokenType]int{
	lexer.OR:          OR,
	lexer.AND:         AND,
	lexer.EQ:          EQUALS,
	lexer.NOT_EQ:      EQUALS,
	lexer.LT:          LESSGREATER,
	lexer.GT:          LESSGREATER,
	lexer.LE:          LESSGREATER,
	lexer.GE:          LESSGREATER,
	lexer.PLUS:        SUM,
	lexer.MINUS:       SUM,
	lexer.SLASH:       PRODUCT,
	lexer.ASTERISK:    PRODUCT,
	lexer.MOD:         PRODUCT,
	lexer.LPAREN:      CALL,      // 左括号用于函数调用
	lexer.QUESTION:    CONDITIONAL, // 问号用于三目运算符
	lexer.ASSIGN:      ASSIGNMENT,  // 赋值运算符
}

// ========== 语法分析器结构 ==========

// Parser 语法分析器
// 负责将 token 流转换为抽象语法树（AST）
// 使用递归下降解析算法
type Parser struct {
	l      *lexer.Lexer // 词法分析器，用于获取 token
	errors []string     // 解析过程中收集的错误信息

	curToken  lexer.Token // 当前正在处理的 token
	peekToken lexer.Token // 下一个 token（用于前瞻）

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
//   l: 词法分析器实例
// 返回:
//   初始化好的 Parser 实例
// 功能:
//   1. 注册所有前缀和中缀解析函数
//   2. 初始化 curToken 和 peekToken
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

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
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseProgram 解析整个程序
// 这是语法分析的入口函数
// 返回:
//   解析后的程序 AST 根节点
// 功能:
//   遍历所有 token，解析为语句，直到遇到 EOF
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
//   解析后的语句节点，如果解析失败返回 nil
func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case lexer.VAR:
		return p.parseLetStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.FUNCTION:
		return p.parseFunctionStatement()
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
		p.nextToken()
	}

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
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // 跳过 :
		p.nextToken()
		if p.curTokenIs(lexer.LPAREN) {
			// 多返回值
			p.nextToken()
			lit.ReturnType = []*Identifier{}
			for !p.curTokenIs(lexer.RPAREN) {
				if p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.ANY) {
					lit.ReturnType = append(lit.ReturnType, &Identifier{Token: p.curToken, Value: p.curToken.Literal})
				}
				p.nextToken()
				if p.curTokenIs(lexer.COMMA) {
					p.nextToken()
				}
			}
		} else {
			// 单返回值
			if p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.ANY) || p.curTokenIs(lexer.VOID) {
				lit.ReturnType = []*Identifier{{Token: p.curToken, Value: p.curToken.Literal}}
			}
		}
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

	p.nextToken()
	exp.TrueExpr = p.parseExpression(CONDITIONAL)

	if !p.expectPeek(lexer.COLON) {
		return nil
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

