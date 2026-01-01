package parser

import (
	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== 运算符优先级常量 ==========
// 优先级从低到高，数值越大优先级越高
const (
	_           int = iota
	LOWEST          // 最低优先级
	ASSIGNMENT      // 赋值运算符：:=, =
	CONDITIONAL     // 三目运算符：? :
	OR              // 逻辑或：||
	AND             // 逻辑与：&&
	EQUALS          // 相等比较：==, !=
	LESSGREATER     // 大小比较：<, >, <=, >=
	SUM             // 加减法：+, -
	PRODUCT         // 乘除法：*, /, %
	PREFIX          // 前缀运算符：!, -
	CALL            // 函数调用：()
	INDEX           // 索引访问：[]
	TYPEASSERT      // 类型断言
)

// precedences 运算符优先级映射表
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
	lexer.LPAREN:       CALL,
	lexer.LBRACKET:     INDEX,
	lexer.QUESTION:     CONDITIONAL,
	lexer.ASSIGN:       ASSIGNMENT,
	lexer.DOT:          CALL,
	lexer.DOUBLE_COLON: CALL,
}

// ========== 语法分析器结构 ==========

// Parser 语法分析器
type Parser struct {
	l      *lexer.Lexer
	errors []string

	prevToken lexer.Token
	curToken  lexer.Token
	peekToken lexer.Token

	allowTernary bool

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

// prefixParseFn 前缀解析函数类型
type prefixParseFn func() Expression

// infixParseFn 中缀解析函数类型
type infixParseFn func(Expression) Expression

// New 创建新的语法分析器
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.allowTernary = true

	// 注册前缀解析函数
	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TRUE, p.parseBoolean)
	p.registerPrefix(lexer.FALSE, p.parseBoolean)
	p.registerPrefix(lexer.NULL, p.parseNull)
	p.registerPrefix(lexer.BANG, p.parsePrefixExpression)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(lexer.THIS, p.parseThisExpression)
	p.registerPrefix(lexer.SUPER, p.parseSuperExpression)
	p.registerPrefix(lexer.NEW, p.parseNewExpression)
	p.registerPrefix(lexer.LBRACE, p.parseArrayLiteral)
	p.registerPrefix(lexer.LBRACKET, p.parseArrayTypeOrLiteral)

	// 注册中缀解析函数
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
	p.registerInfix(lexer.DOT, p.parseMemberAccessExpression)
	p.registerInfix(lexer.ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(lexer.DOUBLE_COLON, p.parseStaticCallExpression)
	p.registerInfix(lexer.LBRACKET, p.parseIndexExpression)

	// 初始化 curToken 和 peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// ParseProgram 解析整个程序
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

// Errors 返回解析错误列表
func (p *Parser) Errors() []string {
	return p.errors
}
