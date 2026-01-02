package parser

import (
	"fmt"

	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== 枚举解析 ==========

// parseEnumStatement 解析枚举声明语句
// 语法: [public|internal] enum EnumName [: BackingType] [implements Interface1, Interface2] { Members }
// isPublic: 是否显式声明为 public
// isInternal: 是否显式声明为 internal
// 如果两者都为 false，则默认为 internal
func (p *Parser) parseEnumStatement(isPublic bool, isInternal bool) *EnumStatement {
	stmt := &EnumStatement{Token: p.curToken}

	// 设置可见性
	if isPublic {
		stmt.IsPublic = true
		stmt.IsInternal = false
	} else {
		// 默认或显式 internal
		stmt.IsPublic = false
		stmt.IsInternal = true
	}

	// 期望枚举名
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 检查是否有底层类型 (: int 或 : string)
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // 跳过枚举名，现在 curToken 是 :
		p.nextToken() // 跳过 :，现在 curToken 应该是类型
		
		if p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.STRING_TYPE) {
			stmt.BackingType = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		} else {
			p.errors = append(p.errors, fmt.Sprintf("枚举底层类型必须是 int 或 string，得到 %s (行 %d, 列 %d)", p.curToken.Literal, p.curToken.Line, p.curToken.Column))
			return nil
		}
	}

	// 检查是否实现接口
	if p.peekTokenIs(lexer.IMPLEMENTS) {
		p.nextToken() // 跳过 implements
		stmt.Interfaces = p.parseInterfaceList()
	}

	// 期望 {
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	// 解析枚举成员和方法
	p.parseEnumBody(stmt)

	return stmt
}

// parseEnumBody 解析枚举体（成员、字段、方法）
func (p *Parser) parseEnumBody(stmt *EnumStatement) {
	p.nextToken() // 跳过 {

	// 用于自动递增值
	nextAutoValue := int64(0)
	hasBackingType := stmt.BackingType != nil && stmt.BackingType.Value == "int"

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// 跳过分号和换行
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// 检查是否是访问修饰符（方法或字段定义的开始）
		if p.curTokenIs(lexer.PUBLIC) || p.curTokenIs(lexer.PRIVATE) || p.curTokenIs(lexer.PROTECTED) {
			accessModifier := p.curToken.Literal
			p.nextToken()

			// 检查是否是 static
			isStatic := false
			if p.curTokenIs(lexer.STATIC) {
				isStatic = true
				p.nextToken()
			}

			// 检查是否是方法
			if p.curTokenIs(lexer.FUNCTION) {
				method := p.parseClassMethod(accessModifier, isStatic, false)
				if method != nil {
					stmt.Methods = append(stmt.Methods, method)
				}
				// 方法解析完成后，curToken 应该是方法体的 }
				// 跳过方法体的 }，如果下一个不是枚举的 }
				if p.curTokenIs(lexer.RBRACE) && !p.peekTokenIs(lexer.EOF) {
					if p.peekTokenIs(lexer.PUBLIC) || p.peekTokenIs(lexer.PRIVATE) ||
						p.peekTokenIs(lexer.PROTECTED) || p.peekTokenIs(lexer.IDENT) {
						p.nextToken()
					}
				}
			} else if p.curTokenIs(lexer.IDENT) {
				// 字段定义
				variable := p.parseClassVariable(accessModifier)
				if variable != nil {
					stmt.Variables = append(stmt.Variables, variable)
				}
				if !p.curTokenIs(lexer.RBRACE) {
					p.nextToken()
				}
			}
			continue
		}

		// 枚举成员（大写开头的标识符）
		if p.curTokenIs(lexer.IDENT) {
			member := p.parseEnumMember(hasBackingType, &nextAutoValue)
			if member != nil {
				stmt.Members = append(stmt.Members, member)
			}

			// 检查是否有逗号分隔（可选）
			if p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // 移动到逗号
				p.nextToken() // 跳过逗号
				continue
			}
		}

		p.nextToken()
	}
}

// parseEnumMember 解析枚举成员
func (p *Parser) parseEnumMember(hasBackingType bool, nextAutoValue *int64) *EnumMember {
	member := &EnumMember{
		Token: p.curToken,
		Name:  &Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	// 检查是否有参数 (用于复杂枚举)
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // 跳过成员名
		p.nextToken() // 跳过 (
		member.Arguments = p.parseEnumMemberArguments()
		return member
	}

	// 检查是否有赋值
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // 跳过成员名
		p.nextToken() // 跳过 =
		member.Value = p.parseExpression(LOWEST)

		// 更新自动递增值
		if hasBackingType {
			if intLit, ok := member.Value.(*IntegerLiteral); ok {
				*nextAutoValue = intLit.Value + 1
			}
		}
	} else if hasBackingType {
		// 自动递增值
		member.Value = &IntegerLiteral{
			Token: lexer.Token{Type: lexer.INT, Literal: fmt.Sprintf("%d", *nextAutoValue)},
			Value: *nextAutoValue,
		}
		*nextAutoValue++
	}

	return member
}

// parseEnumMemberArguments 解析枚举成员的构造参数
func (p *Parser) parseEnumMemberArguments() []Expression {
	args := []Expression{}

	if p.curTokenIs(lexer.RPAREN) {
		return args
	}

	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // 跳过当前表达式
		p.nextToken() // 跳过逗号
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return args
}

