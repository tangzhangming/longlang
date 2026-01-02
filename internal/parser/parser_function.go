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
// 支持可变参数语法：...args 或 ...args:type
func (p *Parser) parseFunctionParameters() []*FunctionParameter {
	parameters := []*FunctionParameter{}
	hasVariadic := false // 标记是否已有可变参数

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return parameters
	}

	p.nextToken()

	// 解析第一个参数
	param := p.parseSingleParameter(&hasVariadic)
	if param == nil {
		return nil
	}
	parameters = append(parameters, param)

	for p.peekTokenIs(lexer.COMMA) {
		// 如果已经有可变参数，不允许再有其他参数
		if hasVariadic {
			p.errors = append(p.errors, fmt.Sprintf("可变参数必须是最后一个参数 (行 %d, 列 %d)", p.peekToken.Line, p.peekToken.Column))
			return nil
		}

		p.nextToken()
		p.nextToken()

		param := p.parseSingleParameter(&hasVariadic)
		if param == nil {
			return nil
		}
		parameters = append(parameters, param)
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return parameters
}

// parseSingleParameter 解析单个函数参数
// 支持普通参数和可变参数 (...args)
func (p *Parser) parseSingleParameter(hasVariadic *bool) *FunctionParameter {
	param := &FunctionParameter{}

	// 检查是否是可变参数 (...)
	if p.curTokenIs(lexer.ELLIPSIS) {
		if *hasVariadic {
			p.errors = append(p.errors, fmt.Sprintf("函数只能有一个可变参数 (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column))
			return nil
		}
		param.IsVariadic = true
		*hasVariadic = true
		p.nextToken() // 跳过 ...
	}

	// 参数名
	if !p.curTokenIs(lexer.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("期望参数名，得到 %s (行 %d, 列 %d)", p.curToken.Type, p.curToken.Line, p.curToken.Column))
		return nil
	}
	param.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 参数类型
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
		p.nextToken()
		param.Type = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	// 默认值（可变参数不允许有默认值）
	if p.peekTokenIs(lexer.ASSIGN) {
		if param.IsVariadic {
			p.errors = append(p.errors, fmt.Sprintf("可变参数不能有默认值 (行 %d, 列 %d)", p.peekToken.Line, p.peekToken.Column))
			return nil
		}
		p.nextToken()
		p.nextToken()
		param.DefaultValue = p.parseExpression(LOWEST)
	}

	return param
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



