package parser

import (
	"fmt"

	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== Token 操作辅助函数 ==========

// nextToken 读取下一个 token
func (p *Parser) nextToken() {
	p.prevToken = p.curToken
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs 检查当前 token 类型
func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs 检查下一个 token 类型
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek 期望下一个 token 是指定类型，如果是则前进
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// ========== 优先级辅助函数 ==========

// peekPrecedence 获取下一个 token 的优先级
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence 获取当前 token 的优先级
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// ========== 注册函数 ==========

// registerPrefix 注册前缀解析函数
func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix 注册中缀解析函数
func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// ========== 错误处理 ==========

// peekError 添加期望 token 错误
func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("期望下一个 token 是 %s，但得到 %s (行 %d, 列 %d)",
		t, p.peekToken.Type, p.peekToken.Line, p.peekToken.Column)
	p.errors = append(p.errors, msg)
}

// noPrefixParseFnError 添加没有前缀解析函数错误
func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("没有找到 %s 的前缀解析函数 (行 %d, 列 %d)",
		t, p.curToken.Line, p.curToken.Column)
	p.errors = append(p.errors, msg)
}

// ========== 类型检查辅助函数 ==========

// isTypeToken 检查是否是类型 token
func (p *Parser) isTypeToken(t lexer.TokenType) bool {
	return t == lexer.STRING_TYPE ||
		t == lexer.INT_TYPE ||
		t == lexer.BOOL_TYPE ||
		t == lexer.FLOAT_TYPE ||
		t == lexer.ANY ||
		t == lexer.VOID ||
		t == lexer.IDENT
}

// curTokenIsType 检查当前 token 是否是类型
func (p *Parser) curTokenIsType() bool {
	return p.isTypeToken(p.curToken.Type)
}

// peekTokenIsType 检查下一个 token 是否是类型
func (p *Parser) peekTokenIsType() bool {
	return p.isTypeToken(p.peekToken.Type)
}


