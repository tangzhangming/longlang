package parser

import (
	"fmt"

	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== 语句解析 ==========

// parseStatement 解析语句
func (p *Parser) parseStatement() Statement {
	// 跳过分号
	for p.curTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	switch p.curToken.Type {
	case lexer.PACKAGE:
		return p.parsePackageStatement()
	case lexer.IMPORT:
		return p.parseImportStatement()
	case lexer.CLASS:
		return p.parseClassStatement()
	case lexer.VAR:
		return p.parseLetStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.BREAK:
		return p.parseBreakStatement()
	case lexer.CONTINUE:
		return p.parseContinueStatement()
	case lexer.FUNCTION:
		return p.parseFunctionStatement()
	case lexer.RBRACE:
		return nil
	case lexer.EOF:
		return nil
	case lexer.ILLEGAL:
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
	if p.peekTokenIs(lexer.STRING_TYPE) || p.peekTokenIs(lexer.INT_TYPE) || p.peekTokenIs(lexer.BOOL_TYPE) || p.peekTokenIs(lexer.FLOAT_TYPE) || p.peekTokenIs(lexer.ANY) {
		p.nextToken()
		stmt.Type = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	// 赋值
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken()
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}

	// 短变量声明 :=
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
		if p.peekTokenIs(lexer.ASSIGN) {
			p.nextToken()
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

	// 如果后面紧跟的是 } 或 ;，说明没有返回值
	if p.peekTokenIs(lexer.RBRACE) || p.peekTokenIs(lexer.SEMICOLON) {
		return stmt
	}

	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)

	return stmt
}

// parseExpressionStatement 解析表达式语句
func (p *Parser) parseExpressionStatement() Statement {
	// 短变量声明 :=
	if p.curToken.Type == lexer.IDENT && p.peekTokenIs(lexer.ASSIGN) && p.peekToken.Literal == ":=" {
		name := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
		p.nextToken()
		assignStmt := &AssignStatement{
			Token: lexer.Token{Type: lexer.ASSIGN, Literal: ":="},
			Name:  name,
		}
		assignStmt.Value = p.parseExpression(LOWEST)
		return assignStmt
	}

	// 自增/自减 i++ 或 i--
	if p.curToken.Type == lexer.IDENT && (p.peekTokenIs(lexer.INCREMENT) || p.peekTokenIs(lexer.DECREMENT)) {
		name := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
		return &IncrementStatement{
			Token:    p.curToken,
			Name:     name,
			Operator: p.curToken.Literal,
		}
	}

	// 普通表达式语句
	stmt := &ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	return stmt
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
			p.nextToken()
			stmt.ElseIf = p.parseIfStatement()
		} else {
			if !p.expectPeek(lexer.LBRACE) {
				return nil
			}
			stmt.Alternative = p.parseBlockStatement()
		}
	}

	return stmt
}

// parseForStatement 解析 for 循环语句
func (p *Parser) parseForStatement() *ForStatement {
	stmt := &ForStatement{Token: p.curToken}

	p.nextToken()

	// 无限循环：for { ... }
	if p.curTokenIs(lexer.LBRACE) {
		stmt.Body = p.parseBlockStatement()
		return stmt
	}

	// 传统 for 循环：for j := 0; j < 5; j++ { ... }
	if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.ASSIGN) && p.peekToken.Literal == ":=" {
		stmt.Init = p.parseForInit()

		if !p.curTokenIs(lexer.SEMICOLON) {
			p.errors = append(p.errors, fmt.Sprintf("for 循环缺少第一个分号 (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column))
			return nil
		}
		p.nextToken()

		if !p.curTokenIs(lexer.SEMICOLON) {
			stmt.Condition = p.parseExpression(LOWEST)
			p.nextToken()
		}

		if !p.curTokenIs(lexer.SEMICOLON) {
			p.errors = append(p.errors, fmt.Sprintf("for 循环缺少第二个分号 (行 %d, 列 %d)", p.curToken.Line, p.curToken.Column))
			return nil
		}
		p.nextToken()

		if !p.curTokenIs(lexer.LBRACE) {
			stmt.Post = p.parseForPost()
		}
	} else {
		// while 式循环：for condition { ... }
		stmt.Condition = p.parseExpression(LOWEST)
		p.nextToken()
	}

	if !p.curTokenIs(lexer.LBRACE) {
		p.errors = append(p.errors, fmt.Sprintf("期望 '{' 但得到 %s (行 %d, 列 %d)", p.curToken.Literal, p.curToken.Line, p.curToken.Column))
		return nil
	}
	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseForInit 解析 for 循环初始化语句
func (p *Parser) parseForInit() Statement {
	name := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()
	p.nextToken()
	assignStmt := &AssignStatement{
		Token: lexer.Token{Type: lexer.ASSIGN, Literal: ":="},
		Name:  name,
	}
	assignStmt.Value = p.parseExpression(LOWEST)
	p.nextToken()
	return assignStmt
}

// parseForPost 解析 for 循环 post 语句
func (p *Parser) parseForPost() Statement {
	if p.curTokenIs(lexer.IDENT) && (p.peekTokenIs(lexer.INCREMENT) || p.peekTokenIs(lexer.DECREMENT)) {
		name := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken()
		stmt := &IncrementStatement{
			Token:    p.curToken,
			Name:     name,
			Operator: p.curToken.Literal,
		}
		p.nextToken()
		return stmt
	}

	stmt := &ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	p.nextToken()
	return stmt
}

// parseBreakStatement 解析 break 语句
func (p *Parser) parseBreakStatement() *BreakStatement {
	return &BreakStatement{Token: p.curToken}
}

// parseContinueStatement 解析 continue 语句
func (p *Parser) parseContinueStatement() *ContinueStatement {
	return &ContinueStatement{Token: p.curToken}
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
		if !p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}
	}

	return block
}

