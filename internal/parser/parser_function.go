package parser

import (
	"fmt"

	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== 函数解析 ==========

// parseFunctionLiteral 解析函数字面量
func (p *Parser) parseFunctionLiteral() Expression {
	lit := &FunctionLiteral{Token: p.curToken}

	// 检查是否是匿名函数
	if p.peekTokenIs(lexer.LPAREN) {
		lit.Name = nil
		p.nextToken()
	} else if p.peekTokenIs(lexer.IDENT) {
		p.nextToken()
		lit.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.expectPeek(lexer.LPAREN) {
			return nil
		}
	} else {
		p.errors = append(p.errors, fmt.Sprintf("期望函数名或 '(' (行 %d, 列 %d)", p.peekToken.Line, p.peekToken.Column))
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	// 解析返回类型
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
		p.nextToken()
		if p.curTokenIs(lexer.LPAREN) {
			// 多返回值
			p.nextToken()
			lit.ReturnType = []*Identifier{}
			for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
				if p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.FLOAT_TYPE) || p.curTokenIs(lexer.ANY) || p.curTokenIs(lexer.IDENT) {
					lit.ReturnType = append(lit.ReturnType, &Identifier{Token: p.curToken, Value: p.curToken.Literal})
				}
				p.nextToken()
				if p.curTokenIs(lexer.COMMA) {
					p.nextToken()
				}
			}
		} else {
			// 单返回值
			if p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.FLOAT_TYPE) || p.curTokenIs(lexer.ANY) || p.curTokenIs(lexer.VOID) || p.curTokenIs(lexer.IDENT) {
				lit.ReturnType = []*Identifier{{Token: p.curToken, Value: p.curToken.Literal}}
			}
		}
	} else if p.peekTokenIs(lexer.STRING_TYPE) || p.peekTokenIs(lexer.INT_TYPE) || p.peekTokenIs(lexer.BOOL_TYPE) || p.peekTokenIs(lexer.FLOAT_TYPE) || p.peekTokenIs(lexer.ANY) || p.peekTokenIs(lexer.VOID) || p.peekTokenIs(lexer.IDENT) {
		// 不使用冒号的语法
		p.nextToken()
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

	// 限制：调用参数内不允许三目运算符
	oldAllowTernary := p.allowTernary
	p.allowTernary = false
	defer func() { p.allowTernary = oldAllowTernary }()

	p.nextToken()

	arg := CallArgument{}
	if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.COLON) {
		arg.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
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
			p.nextToken()
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



