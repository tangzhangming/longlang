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
			
			// 解析表达式（继承当前 Lexer 的 isStdlib 状态）
			exprLexer := lexer.NewWithOptions(exprText, p.l.IsStdlib())
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

// parseCompoundAssignmentExpression 解析复合赋值表达式
// 对应语法：a += b, a -= b, a *= b, a /= b, a %= b, a &= b, a |= b, a ^= b, a <<= b, a >>= b
func (p *Parser) parseCompoundAssignmentExpression(left Expression) Expression {
	exp := &CompoundAssignmentExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	p.nextToken()
	exp.Right = p.parseExpression(LOWEST)

	return exp
}

// parseTypeAssertionExpression 解析类型断言表达式
// 对应语法：value as Type（强制断言）或 value as? Type（安全断言）
// 例如：x as string, obj as User, value as? int
// 支持的目标类型：
//   - 基本类型：int, float, string, bool, byte 等
//   - 数组类型：[]int, []string 等
//   - Map 类型：map[string]int 等
//   - 类/接口：User, Readable 等
func (p *Parser) parseTypeAssertionExpression(left Expression) Expression {
	exp := &TypeAssertionExpression{
		Token:  p.curToken,
		Left:   left,
		IsSafe: p.curToken.Type == lexer.AS_SAFE,
	}

	// 移动到类型位置
	p.nextToken()

	// 解析目标类型
	exp.TargetType = p.parseTypeExpressionForAssertion()
	if exp.TargetType == nil {
		msg := fmt.Sprintf("类型断言缺少目标类型 (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	}

	return exp
}

// parseTypeExpressionForAssertion 解析用于类型断言的类型表达式
// 支持：基本类型、数组类型 []T、Map 类型 map[K]V、类名/接口名
func (p *Parser) parseTypeExpressionForAssertion() Expression {
	switch p.curToken.Type {
	case lexer.LBRACKET:
		// 数组类型：[]int, []string 等
		return p.parseArrayTypeForAssertion()
	case lexer.MAP:
		// Map 类型：map[string]int 等
		return p.parseMapTypeForAssertion()
	case lexer.IDENT:
		return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	case lexer.INT_TYPE:
		return &Identifier{Token: p.curToken, Value: "int"}
	case lexer.I8_TYPE:
		return &Identifier{Token: p.curToken, Value: "i8"}
	case lexer.I16_TYPE:
		return &Identifier{Token: p.curToken, Value: "i16"}
	case lexer.I32_TYPE:
		return &Identifier{Token: p.curToken, Value: "i32"}
	case lexer.I64_TYPE:
		return &Identifier{Token: p.curToken, Value: "i64"}
	case lexer.UINT_TYPE:
		return &Identifier{Token: p.curToken, Value: "uint"}
	case lexer.U8_TYPE:
		return &Identifier{Token: p.curToken, Value: "u8"}
	case lexer.U16_TYPE:
		return &Identifier{Token: p.curToken, Value: "u16"}
	case lexer.U32_TYPE:
		return &Identifier{Token: p.curToken, Value: "u32"}
	case lexer.U64_TYPE:
		return &Identifier{Token: p.curToken, Value: "u64"}
	case lexer.BYTE_TYPE:
		return &Identifier{Token: p.curToken, Value: "byte"}
	case lexer.FLOAT_TYPE:
		return &Identifier{Token: p.curToken, Value: "float"}
	case lexer.F32_TYPE:
		return &Identifier{Token: p.curToken, Value: "f32"}
	case lexer.F64_TYPE:
		return &Identifier{Token: p.curToken, Value: "f64"}
	case lexer.STRING_TYPE:
		return &Identifier{Token: p.curToken, Value: "string"}
	case lexer.BOOL_TYPE:
		return &Identifier{Token: p.curToken, Value: "bool"}
	case lexer.ANY:
		return &Identifier{Token: p.curToken, Value: "any"}
	default:
		return nil
	}
}

// parseArrayTypeForAssertion 解析数组类型（用于类型断言）
// 例如：[]int, []string, []User
func (p *Parser) parseArrayTypeForAssertion() Expression {
	arrayType := &ArrayType{Token: p.curToken}

	// 跳过 [
	p.nextToken()

	// 检查是否有大小（固定数组）或直接是 ]（切片）
	if p.curToken.Type == lexer.INT {
		// 固定大小数组 [5]int
		size, _ := strconv.ParseInt(p.curToken.Literal, 0, 64)
		arrayType.Size = &IntegerLiteral{Token: p.curToken, Value: size}
		p.nextToken()
	} else if p.curToken.Type == lexer.ELLIPSIS {
		// 推导大小 [...]int
		arrayType.IsInferred = true
		p.nextToken()
	}
	// 否则是切片 []int

	// 期望 ]
	if !p.curTokenIs(lexer.RBRACKET) {
		msg := fmt.Sprintf("数组类型期望 ']'，得到 %s (行 %d, 列 %d)", p.curToken.Literal, p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	}

	// 跳过 ]
	p.nextToken()

	// 解析元素类型
	arrayType.ElementType = p.parseTypeExpressionForAssertion()

	return arrayType
}

// parseMapTypeForAssertion 解析 Map 类型（用于类型断言）
// 例如：map[string]int, map[string]User
func (p *Parser) parseMapTypeForAssertion() Expression {
	mapType := &MapType{Token: p.curToken}

	// 跳过 map
	p.nextToken()

	// 期望 [
	if !p.curTokenIs(lexer.LBRACKET) {
		msg := fmt.Sprintf("Map 类型期望 '['，得到 %s (行 %d, 列 %d)", p.curToken.Literal, p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()

	// 解析键类型
	if p.curToken.Type == lexer.IDENT || p.curToken.Type == lexer.STRING_TYPE {
		mapType.KeyType = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	} else {
		msg := fmt.Sprintf("Map 键类型无效 (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()

	// 期望 ]
	if !p.curTokenIs(lexer.RBRACKET) {
		msg := fmt.Sprintf("Map 类型期望 ']'，得到 %s (行 %d, 列 %d)", p.curToken.Literal, p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()

	// 解析值类型
	mapType.ValueType = p.parseTypeExpressionForAssertion()

	return mapType
}






