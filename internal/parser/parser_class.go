package parser

import (
	"fmt"

	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== 类解析 ==========

// parseClassStatement 解析类声明语句
// 语法: class ClassName extends ParentClass implements Interface1, Interface2 { ... }
func (p *Parser) parseClassStatement() *ClassStatement {
	stmt := &ClassStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 解析继承 extends
	if p.peekTokenIs(lexer.EXTENDS) {
		p.nextToken() // 跳过 extends
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		stmt.Parent = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	// 解析实现 implements
	if p.peekTokenIs(lexer.IMPLEMENTS) {
		p.nextToken() // 跳过 implements
		stmt.Interfaces = p.parseInterfaceList()
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Members = p.parseClassMembers()

	return stmt
}

// parseInterfaceList 解析接口列表（用于 implements）
func (p *Parser) parseInterfaceList() []*Identifier {
	interfaces := []*Identifier{}

	if !p.expectPeek(lexer.IDENT) {
		return interfaces
	}
	interfaces = append(interfaces, &Identifier{Token: p.curToken, Value: p.curToken.Literal})

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // 跳过逗号
		if !p.expectPeek(lexer.IDENT) {
			return interfaces
		}
		interfaces = append(interfaces, &Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	return interfaces
}

// parseInterfaceStatement 解析接口声明语句
// 语法: interface InterfaceName { function method1(); function method2():type; }
func (p *Parser) parseInterfaceStatement() *InterfaceStatement {
	stmt := &InterfaceStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Methods = p.parseInterfaceMethods()

	return stmt
}

// parseInterfaceMethods 解析接口方法列表
func (p *Parser) parseInterfaceMethods() []*InterfaceMethod {
	methods := []*InterfaceMethod{}

	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		if p.curTokenIs(lexer.FUNCTION) {
			method := p.parseInterfaceMethod()
			if method != nil {
				methods = append(methods, method)
			}
		}
		p.nextToken()
	}

	return methods
}

// parseInterfaceMethod 解析接口方法签名
func (p *Parser) parseInterfaceMethod() *InterfaceMethod {
	method := &InterfaceMethod{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	method.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	method.Parameters = p.parseFunctionParameters()

	// 解析返回类型
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // 跳过 :
		p.nextToken() // 移动到返回类型
		if p.curTokenIs(lexer.LPAREN) {
			// 多返回值
			p.nextToken()
			method.ReturnType = []*Identifier{}
			for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
				if p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.FLOAT_TYPE) || p.curTokenIs(lexer.ANY) || p.curTokenIs(lexer.IDENT) {
					method.ReturnType = append(method.ReturnType, &Identifier{Token: p.curToken, Value: p.curToken.Literal})
				}
				p.nextToken()
				if p.curTokenIs(lexer.COMMA) {
					p.nextToken()
				}
			}
		} else {
			// 单返回值
			if p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.FLOAT_TYPE) || p.curTokenIs(lexer.ANY) || p.curTokenIs(lexer.VOID) || p.curTokenIs(lexer.IDENT) {
				method.ReturnType = []*Identifier{{Token: p.curToken, Value: p.curToken.Literal}}
			}
		}
	} else if p.peekTokenIs(lexer.STRING_TYPE) || p.peekTokenIs(lexer.INT_TYPE) || p.peekTokenIs(lexer.BOOL_TYPE) || p.peekTokenIs(lexer.FLOAT_TYPE) || p.peekTokenIs(lexer.ANY) || p.peekTokenIs(lexer.VOID) || p.peekTokenIs(lexer.IDENT) {
		p.nextToken()
		method.ReturnType = []*Identifier{{Token: p.curToken, Value: p.curToken.Literal}}
	}

	return method
}

// parseClassMembers 解析类成员
func (p *Parser) parseClassMembers() []ClassMember {
	members := []ClassMember{}

	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// 解析访问修饰符
		if !p.curTokenIs(lexer.PUBLIC) && !p.curTokenIs(lexer.PRIVATE) && !p.curTokenIs(lexer.PROTECTED) {
			p.nextToken()
			continue
		}

		accessModifier := p.curToken.Literal
		p.nextToken()

		// 检查是否是静态
		isStatic := false
		if p.curTokenIs(lexer.STATIC) {
			isStatic = true
			p.nextToken()
		}

		// 方法
		if p.curTokenIs(lexer.FUNCTION) {
			method := p.parseClassMethod(accessModifier, isStatic)
			if method != nil {
				members = append(members, method)
				// 方法解析完成后，curToken 是方法体的 }，需要跳过它
				// 但只有当下一个 token 不是类的 } 时才跳过（避免跳过类的 }）
				if p.curTokenIs(lexer.RBRACE) && !p.peekTokenIs(lexer.EOF) {
					// 检查是否还有更多成员
					if p.peekTokenIs(lexer.PUBLIC) || p.peekTokenIs(lexer.PRIVATE) || p.peekTokenIs(lexer.PROTECTED) {
						p.nextToken() // 跳过方法体的 }
					}
					// 如果 peekToken 是 }，说明到达类的末尾，不跳过
				}
			} else {
				// 错误恢复
				for !p.curTokenIs(lexer.EOF) &&
					!p.curTokenIs(lexer.RBRACE) &&
					!p.curTokenIs(lexer.PUBLIC) &&
					!p.curTokenIs(lexer.PRIVATE) &&
					!p.curTokenIs(lexer.PROTECTED) {
					p.nextToken()
				}
			}
		} else if p.curTokenIs(lexer.IDENT) {
			// 成员变量
			variable := p.parseClassVariable(accessModifier)
			if variable != nil {
				members = append(members, variable)
				if !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.PUBLIC) && !p.curTokenIs(lexer.PRIVATE) && !p.curTokenIs(lexer.PROTECTED) && !p.curTokenIs(lexer.EOF) {
					p.nextToken()
				}
			} else {
				if !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.PUBLIC) && !p.curTokenIs(lexer.PRIVATE) && !p.curTokenIs(lexer.PROTECTED) && !p.curTokenIs(lexer.EOF) {
					p.nextToken()
				}
			}
		} else {
			p.errors = append(p.errors, fmt.Sprintf("类成员必须是方法或变量，得到 %s (行 %d, 列 %d)", p.curToken.Type, p.curToken.Line, p.curToken.Column))
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
		p.errors = append(p.errors, fmt.Sprintf("期望变量名，得到 %s (行 %d, 列 %d)", p.curToken.Type, p.curToken.Line, p.curToken.Column))
		return nil
	}

	variable.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken()

	// 解析类型
	if !p.curTokenIs(lexer.STRING_TYPE) && !p.curTokenIs(lexer.INT_TYPE) && !p.curTokenIs(lexer.BOOL_TYPE) && !p.curTokenIs(lexer.FLOAT_TYPE) && !p.curTokenIs(lexer.ANY) {
		p.errors = append(p.errors, fmt.Sprintf("类成员变量必须声明类型，得到 %s (行 %d, 列 %d)", p.curToken.Type, p.curToken.Line, p.curToken.Column))
		return nil
	}

	variable.Type = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken()

	// 初始值
	if p.curTokenIs(lexer.ASSIGN) {
		p.nextToken()
		variable.Value = p.parseExpression(LOWEST)
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
		p.errors = append(p.errors, fmt.Sprintf("期望 function 关键字 (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column))
		return nil
	}

	p.nextToken()

	if !p.curTokenIs(lexer.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("期望方法名 (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column))
		return nil
	}

	method.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	method.Parameters = p.parseFunctionParameters()

	// 解析返回类型
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
		p.nextToken()

		if p.curTokenIs(lexer.LPAREN) {
			// 多返回值
			p.nextToken()
			method.ReturnType = []*Identifier{}
			for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
				if p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.FLOAT_TYPE) || p.curTokenIs(lexer.ANY) || p.curTokenIs(lexer.IDENT) {
					method.ReturnType = append(method.ReturnType, &Identifier{Token: p.curToken, Value: p.curToken.Literal})
				}
				p.nextToken()
				if p.curTokenIs(lexer.COMMA) {
					p.nextToken()
				}
			}
		} else {
			// 单返回值
			if p.curTokenIs(lexer.STRING_TYPE) || p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.FLOAT_TYPE) || p.curTokenIs(lexer.ANY) || p.curTokenIs(lexer.VOID) || p.curTokenIs(lexer.IDENT) {
				method.ReturnType = []*Identifier{{Token: p.curToken, Value: p.curToken.Literal}}
			}
		}
	} else if p.peekTokenIs(lexer.STRING_TYPE) || p.peekTokenIs(lexer.INT_TYPE) || p.peekTokenIs(lexer.BOOL_TYPE) || p.peekTokenIs(lexer.FLOAT_TYPE) || p.peekTokenIs(lexer.ANY) || p.peekTokenIs(lexer.VOID) || p.peekTokenIs(lexer.IDENT) {
		p.nextToken()
		method.ReturnType = []*Identifier{{Token: p.curToken, Value: p.curToken.Literal}}
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	method.Body = p.parseBlockStatement()

	return method
}

// parseThisExpression 解析 this 表达式
func (p *Parser) parseThisExpression() Expression {
	return &ThisExpression{Token: p.curToken}
}

// parseSuperExpression 解析 super 表达式
func (p *Parser) parseSuperExpression() Expression {
	return &SuperExpression{Token: p.curToken}
}

// parseNewExpression 解析 new 表达式
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

// parseStaticCallExpression 解析静态方法调用
func (p *Parser) parseStaticCallExpression(left Expression) Expression {
	className, ok := left.(*Identifier)
	if !ok {
		p.errors = append(p.errors, fmt.Sprintf("静态方法调用左侧必须是类名 (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column))
		return nil
	}

	exp := &StaticCallExpression{
		Token:     p.curToken,
		ClassName: className,
	}

	p.nextToken()

	if !p.curTokenIs(lexer.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("期望方法名 (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column))
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
func (p *Parser) parseMemberAccessExpression(left Expression) Expression {
	exp := &MemberAccessExpression{
		Token:  p.curToken,
		Object: left,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	if !p.curTokenIs(lexer.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("期望成员名 (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column))
		return nil
	}

	exp.Member = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 方法调用 object.method()
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken()
		call := &CallExpression{
			Token:    p.curToken,
			Function: exp,
		}
		call.Arguments = p.parseCallArguments()
		return call
	}

	// 链式访问 object.member.member2
	if p.peekTokenIs(lexer.DOT) {
		p.nextToken()
		return p.parseMemberAccessExpression(exp)
	}

	// 赋值 object.member = value
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken()
		return p.parseAssignmentExpression(exp)
	}

	_ = precedence
	return exp
}

