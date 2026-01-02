package parser

import (
	"fmt"
	"strconv"

	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== 表达式解析 ==========

// parseExpression 解析表达式（Pratt 解析器核心）
func (p *Parser) parseExpression(precedence int) Expression {
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

// ========== 字面量解析 ==========

// parseIdentifier 解析标识符
func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral 解析整数字面量
func (p *Parser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("无法将 %q 解析为整数 (行 %d, 列 %d)", p.curToken.Literal, p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// parseFloatLiteral 解析浮点数字面量
func (p *Parser) parseFloatLiteral() Expression {
	lit := &FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("无法将 %q 解析为浮点数 (行 %d, 列 %d)", p.curToken.Literal, p.curToken.Line, p.curToken.Column)
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

// parseInterpolatedStringLiteral 解析插值字符串
// 格式：$"Hello, {name}! Sum is {a + b}."
// 插值表达式不支持三元表达式
// 占位符处理：\x01 -> {，\x02 -> }
func (p *Parser) parseInterpolatedStringLiteral() Expression {
	lit := &InterpolatedStringLiteral{Token: p.curToken}
	literal := p.curToken.Literal
	
	// 解析字符串，找出 {} 中的表达式
	var parts []InterpolatedPart
	var currentText []byte
	i := 0
	
	for i < len(literal) {
		if literal[i] == 0x01 {
			// 占位符 \x01 -> 字面 {
			currentText = append(currentText, '{')
			i++
		} else if literal[i] == 0x02 {
			// 占位符 \x02 -> 字面 }
			currentText = append(currentText, '}')
			i++
		} else if literal[i] == '{' {
			// 保存之前的文本部分
			if len(currentText) > 0 {
				parts = append(parts, InterpolatedPart{
					IsExpr: false,
					Text:   string(currentText),
				})
				currentText = nil
			}
			
			// 提取 {} 中的表达式文本
			braceCount := 1
			exprStart := i + 1
			i++
			for i < len(literal) && braceCount > 0 {
				if literal[i] == '{' {
					braceCount++
				} else if literal[i] == '}' {
					braceCount--
				} else if literal[i] == '"' {
					// 跳过字符串中的内容
					i++
					for i < len(literal) && literal[i] != '"' {
						if literal[i] == '\\' {
							i++
						}
						i++
					}
				}
				if braceCount > 0 {
					i++
				}
			}
			
			exprText := literal[exprStart:i]
			i++ // 跳过 }
			
			// 解析表达式
			exprLexer := lexer.New(exprText)
			exprParser := New(exprLexer)
			expr := exprParser.ParseExpression()
			
			if len(exprParser.Errors()) > 0 {
				for _, err := range exprParser.Errors() {
					p.errors = append(p.errors, "插值表达式错误: "+err)
				}
				continue
			}
			
			if expr != nil {
				parts = append(parts, InterpolatedPart{
					IsExpr: true,
					Expr:   expr,
				})
			}
		} else {
			currentText = append(currentText, literal[i])
			i++
		}
	}
	
	// 保存最后的文本部分
	if len(currentText) > 0 {
		parts = append(parts, InterpolatedPart{
			IsExpr: false,
			Text:   string(currentText),
		})
	}
	
	lit.Parts = parts
	return lit
}

// parseBoolean 解析布尔字面量
func (p *Parser) parseBoolean() Expression {
	return &BooleanLiteral{Token: p.curToken, Value: p.curToken.Type == lexer.TRUE}
}

// parseNull 解析 null 字面量
func (p *Parser) parseNull() Expression {
	return &NullLiteral{Token: p.curToken}
}

// ========== 运算符表达式解析 ==========

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

// parseGroupedExpression 解析分组表达式（括号）
func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return exp
}

// parseTernaryExpression 解析三目运算符表达式
func (p *Parser) parseTernaryExpression(condition Expression) Expression {
	exp := &TernaryExpression{
		Token:     p.curToken,
		Condition: condition,
	}

	// 限制：三目运算符不能作为函数/方法参数
	if !p.allowTernary {
		p.errors = append(p.errors, fmt.Sprintf("三目运算符不能作为函数/方法参数使用 (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column))
	}

	// 格式检查
	condEndLine := p.prevToken.Line
	questionLine := p.curToken.Line

	p.nextToken()
	exp.TrueExpr = p.parseExpression(CONDITIONAL)

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	trueEndLine := p.prevToken.Line
	colonLine := p.curToken.Line

	// 单行检查
	if questionLine == condEndLine {
		if colonLine != questionLine {
			p.errors = append(p.errors, fmt.Sprintf("三目运算符单行写法要求 '?' 和 ':' 在同一行 (行 %d, 列 %d)", questionLine, p.curToken.Column))
		}
	} else {
		// 多行检查
		if colonLine == trueEndLine {
			p.errors = append(p.errors, fmt.Sprintf("三目运算符多行写法要求 ':' 单独换行，禁止 '? true : false' 这种混写 (行 %d, 列 %d)", colonLine, p.curToken.Column))
		}
	}

	p.nextToken()
	exp.FalseExpr = p.parseExpression(CONDITIONAL)

	return exp
}

// parseAssignmentExpression 解析赋值表达式
func (p *Parser) parseAssignmentExpression(left Expression) Expression {
	exp := &AssignmentExpression{
		Token: p.curToken,
		Left:  left,
	}

	p.nextToken()
	exp.Right = p.parseExpression(LOWEST)

	return exp
}






