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
	case lexer.NAMESPACE:
		return p.parseNamespaceStatement()
	case lexer.USE:
		return p.parseUseStatement()
	case lexer.ABSTRACT:
		// abstract class ...
		return p.parseClassStatement()
	case lexer.CLASS:
		return p.parseClassStatement()
	case lexer.INTERFACE:
		return p.parseInterfaceStatement()
	case lexer.ENUM:
		return p.parseEnumStatement()
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
	case lexer.TRY:
		return p.parseTryStatement()
	case lexer.THROW:
		return p.parseThrowStatement()
	case lexer.GO:
		return p.parseGoStatement()
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
// 支持：
//   - var x int = 10
//   - var x = 10
//   - var numbers [5]int = {1, 2, 3, 4, 5}
//   - var ids []int = {1, 2, 3}
//   - var names = {"a", "b"}
func (p *Parser) parseLetStatement() Statement {
	stmt := &LetStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 检查是否是数组类型声明 [
	if p.peekTokenIs(lexer.LBRACKET) {
		p.nextToken() // 移动到 [
		arrayType := p.parseArrayType()
		if arrayType != nil {
			stmt.Type = arrayType
		}
	} else if p.peekTokenIs(lexer.STRING_TYPE) || p.peekTokenIs(lexer.INT_TYPE) || 
		p.peekTokenIs(lexer.BOOL_TYPE) || p.peekTokenIs(lexer.FLOAT_TYPE) || 
		p.peekTokenIs(lexer.ANY) || p.peekTokenIs(lexer.I8_TYPE) || 
		p.peekTokenIs(lexer.I16_TYPE) || p.peekTokenIs(lexer.I32_TYPE) ||
		p.peekTokenIs(lexer.I64_TYPE) || p.peekTokenIs(lexer.UINT_TYPE) || 
		p.peekTokenIs(lexer.U8_TYPE) || p.peekTokenIs(lexer.BYTE_TYPE) ||
		p.peekTokenIs(lexer.U16_TYPE) || 
		p.peekTokenIs(lexer.U32_TYPE) || p.peekTokenIs(lexer.U64_TYPE) ||
		p.peekTokenIs(lexer.F32_TYPE) || p.peekTokenIs(lexer.F64_TYPE) {
		// 简单类型声明
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

// parseTryStatement 解析 try-catch-finally 语句
// 语法：try { ... } catch (ExceptionType e) { ... } finally { ... }
func (p *Parser) parseTryStatement() *TryStatement {
	stmt := &TryStatement{Token: p.curToken}

	// 期望 try 后面是 {
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	// 解析 try 块
	stmt.TryBlock = p.parseBlockStatement()

	// 解析 catch 子句（可以有多个）
	stmt.CatchClauses = []*CatchClause{}
	for p.peekTokenIs(lexer.CATCH) {
		p.nextToken() // 移动到 catch
		catchClause := p.parseCatchClause()
		if catchClause != nil {
			stmt.CatchClauses = append(stmt.CatchClauses, catchClause)
		}
	}

	// 解析 finally 块（可选）
	if p.peekTokenIs(lexer.FINALLY) {
		p.nextToken() // 移动到 finally
		if !p.expectPeek(lexer.LBRACE) {
			return nil
		}
		stmt.FinallyBlock = p.parseBlockStatement()
	}

	// 验证：至少要有一个 catch 或 finally
	if len(stmt.CatchClauses) == 0 && stmt.FinallyBlock == nil {
		msg := fmt.Sprintf("try 语句必须至少有一个 catch 或 finally 块 (行 %d, 列 %d)", 
			stmt.Token.Line, stmt.Token.Column)
		p.errors = append(p.errors, msg)
		return nil
	}

	return stmt
}

// parseCatchClause 解析 catch 子句
// 语法：catch (ExceptionType variableName) { ... } 或 catch (variableName) { ... }
func (p *Parser) parseCatchClause() *CatchClause {
	clause := &CatchClause{Token: p.curToken}

	// 期望 catch 后面是 (
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	// 解析异常类型和变量名
	// catch (ExceptionType e) 或 catch (e)
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	firstIdent := &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 检查下一个 token
	if p.peekTokenIs(lexer.IDENT) {
		// 有类型：catch (ExceptionType e)
		clause.ExceptionType = firstIdent
		p.nextToken()
		clause.ExceptionVar = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	} else {
		// 无类型：catch (e)
		clause.ExceptionVar = firstIdent
	}

	// 期望 )
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	// 期望 {
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	// 解析 catch 块
	clause.Body = p.parseBlockStatement()

	return clause
}

// parseThrowStatement 解析 throw 语句
// 语法：throw expression
func (p *Parser) parseThrowStatement() *ThrowStatement {
	stmt := &ThrowStatement{Token: p.curToken}

	p.nextToken() // 跳过 throw

	// 解析要抛出的异常表达式
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

// parseGoStatement 解析 go 语句（启动协程）
// 语法：go expression
// 支持：go fn() { ... }
// 支持：go handler()
// 支持：go this.process()
// 支持：go Worker::run()
func (p *Parser) parseGoStatement() *GoStatement {
	stmt := &GoStatement{Token: p.curToken}

	p.nextToken() // 跳过 go

	// 解析要执行的表达式
	stmt.Call = p.parseExpression(LOWEST)

	return stmt
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

