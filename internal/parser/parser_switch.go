package parser

import (
	"fmt"

	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== Switch 语句解析 ==========

// parseSwitchStatement 解析 switch 语句
// 支持语法：
//   - switch expr { case ... }
//   - switch (expr) { case ... }
//   - switch { case condition: ... }  (条件 switch)
//   - switch init; expr { case ... }  (带初始化)
func (p *Parser) parseSwitchStatement() *SwitchStatement {
	stmt := &SwitchStatement{Token: p.curToken}
	p.nextToken() // 跳过 'switch'

	// 检查是否有括号
	hasParens := p.curTokenIs(lexer.LPAREN)
	if hasParens {
		p.nextToken() // 跳过 '('
	}

	// 检查是否是无表达式的条件 switch: switch { ... }
	alreadyAtBrace := false
	if p.curTokenIs(lexer.LBRACE) {
		// 无表达式的条件 switch: switch { ... }
		stmt.Value = nil
		alreadyAtBrace = true
	} else if hasParens && p.curTokenIs(lexer.RPAREN) {
		// switch () { ... } 形式
		p.nextToken() // 跳过 ')'
		stmt.Value = nil
	} else {
		// 解析表达式或初始化语句
		// 首先尝试解析为表达式
		expr := p.parseExpression(LOWEST)

		// 检查是否是初始化语句（后面跟着分号）
		if p.peekTokenIs(lexer.SEMICOLON) {
			// 是初始化语句
			stmt.Init = &ExpressionStatement{
				Token:      p.curToken,
				Expression: expr,
			}
			p.nextToken() // 移动到分号
			p.nextToken() // 跳过分号

			// 检查是否是条件 switch
			if p.curTokenIs(lexer.LBRACE) {
				stmt.Value = nil
				alreadyAtBrace = true
			} else {
				// 解析实际的 switch 表达式
				stmt.Value = p.parseExpression(LOWEST)
			}
		} else {
			// 不是初始化语句，expr 就是 switch 表达式
			stmt.Value = expr
		}

		if hasParens {
			if !p.expectPeek(lexer.RPAREN) {
				return nil
			}
		}
	}

	// 期望 { (如果还没有在 LBRACE 上)
	if !alreadyAtBrace {
		if !p.expectPeek(lexer.LBRACE) {
			return nil
		}
	}
	p.nextToken() // 跳过 '{'

	// 解析 case 分支
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.CASE) {
			caseClause := p.parseCaseClause(stmt.Value == nil)
			if caseClause != nil {
				stmt.Cases = append(stmt.Cases, caseClause)
			}
		} else if p.curTokenIs(lexer.DEFAULT) {
			p.nextToken() // 跳过 'default'，现在应该在 ':'
			if !p.curTokenIs(lexer.COLON) {
				p.errors = append(p.errors, fmt.Sprintf("switch default 期望 ':' (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column))
				return nil
			}
			p.nextToken() // 跳过 ':'
			stmt.Default = p.parseCaseBody()
		} else {
			p.nextToken()
		}
	}

	return stmt
}

// parseCaseClause 解析 case 分支
// isConditionSwitch: 是否是条件 switch（无表达式）
func (p *Parser) parseCaseClause(isConditionSwitch bool) *CaseClause {
	clause := &CaseClause{Token: p.curToken}
	p.nextToken() // 跳过 'case'

	if isConditionSwitch {
		// 条件 switch: case condition:
		clause.IsCondition = true
		clause.Condition = p.parseExpression(LOWEST)
	} else {
		// 值 switch: case value1, value2, ...:
		clause.Values = []Expression{}
		clause.Values = append(clause.Values, p.parseExpression(LOWEST))

		// 解析多个值
		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // 跳过 ','
			p.nextToken()
			clause.Values = append(clause.Values, p.parseExpression(LOWEST))
		}
	}

	// 期望 :
	if !p.expectPeek(lexer.COLON) {
		return nil
	}
	p.nextToken() // 跳过 ':'

	// 解析 case 体
	clause.Body = p.parseCaseBody()

	return clause
}

// parseCaseBody 解析 case/default 体（直到下一个 case/default/}）
func (p *Parser) parseCaseBody() *BlockStatement {
	block := &BlockStatement{Token: p.curToken}
	block.Statements = []Statement{}

	for !p.curTokenIs(lexer.CASE) && !p.curTokenIs(lexer.DEFAULT) && 
		!p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		// 只有当 parseStatement 没有推进 token 时才手动推进
		// 检查是否已经到了边界
		if !p.curTokenIs(lexer.CASE) && !p.curTokenIs(lexer.DEFAULT) && 
			!p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}
	}

	return block
}

// ========== Match 表达式解析 ==========

// parseMatchExpression 解析 match 表达式
// 支持语法：
//   - match expr { pattern => result, ... }
//   - match (expr) { pattern => result, ... }
//   - match expr { n if guard => result, ... }
//   - match expr { _ => default_result }
func (p *Parser) parseMatchExpression() Expression {
	expr := &MatchExpression{Token: p.curToken}
	p.nextToken() // 跳过 'match'

	// 检查是否有括号
	hasParens := p.curTokenIs(lexer.LPAREN)
	if hasParens {
		p.nextToken() // 跳过 '('
	}

	// 解析要匹配的表达式
	expr.Value = p.parseExpression(LOWEST)

	if hasParens {
		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	}

	// 期望 {
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}
	p.nextToken() // 跳过 '{'

	// 解析匹配分支
	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		arm := p.parseMatchArm()
		if arm != nil {
			expr.Arms = append(expr.Arms, arm)
		} else {
			// 如果解析失败，跳过当前 token 避免无限循环
			p.nextToken()
		}

		// 跳过可选的逗号
		if p.curTokenIs(lexer.COMMA) {
			p.nextToken()
		}
	}

	return expr
}

// parseMatchArm 解析 match 分支
// 支持语法：
//   - pattern => result
//   - pattern1, pattern2 => result
//   - identifier if guard => result
//   - _ => default_result
//   - pattern => { block }
func (p *Parser) parseMatchArm() *MatchArm {
	arm := &MatchArm{Token: p.curToken}

	// 检查是否是通配符 _
	if p.curTokenIs(lexer.IDENT) && p.curToken.Literal == "_" {
		arm.IsWildcard = true
		// 不调用 nextToken，让 expectPeek 来推进
	} else {
		// 解析模式
		arm.Patterns = []Expression{}
		
		// 检查是否是带守卫的绑定变量
		// 格式：identifier if condition => ...
		firstExpr := p.parseExpression(LOWEST)
		
		if p.peekTokenIs(lexer.IF) && isSimpleIdentifier(firstExpr) {
			// 是带守卫的绑定变量
			arm.Binding = firstExpr.(*Identifier)
			p.nextToken() // 移动到 'if'
			p.nextToken() // 跳过 'if'，移动到守卫条件
			arm.Guard = p.parseExpression(LOWEST)
		} else {
			// 普通模式
			arm.Patterns = append(arm.Patterns, firstExpr)
			
			// 解析多个模式值
			for p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // 移动到 ','
				p.nextToken() // 跳过 ','
				
				// 检查是否遇到了 => (防止误解析)
				if p.curTokenIs(lexer.ARROW) {
					break
				}
				
				arm.Patterns = append(arm.Patterns, p.parseExpression(LOWEST))
			}
		}
	}

	// 期望 =>
	if !p.expectPeek(lexer.ARROW) {
		p.errors = append(p.errors, fmt.Sprintf("match 分支期望 '=>' (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column))
		return nil
	}
	p.nextToken() // 跳过 '=>'，移动到结果表达式

	// 解析结果：可以是表达式或代码块
	if p.curTokenIs(lexer.LBRACE) {
		arm.Body = p.parseBlockStatement()
		p.nextToken() // 跳过 '}'
	} else {
		arm.Result = p.parseExpression(LOWEST)
		// 解析完结果后，让主循环来决定是否需要推进
		// 检查下一个 token，如果是逗号或 RBRACE 就推进
		if p.peekTokenIs(lexer.COMMA) || p.peekTokenIs(lexer.RBRACE) {
			p.nextToken()
		} else if !p.peekTokenIs(lexer.EOF) {
			// 如果下一个是新的模式（如数字或标识符），也推进
			p.nextToken()
		}
	}

	return arm
}

// isSimpleIdentifier 检查表达式是否是简单标识符
func isSimpleIdentifier(expr Expression) bool {
	_, ok := expr.(*Identifier)
	return ok
}


