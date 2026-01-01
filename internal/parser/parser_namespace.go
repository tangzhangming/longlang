package parser

import (
	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== 命名空间解析 ==========

// parseNamespaceStatement 解析命名空间声明语句
// 语法：namespace Namespace.Name 或 namespace Name
func (p *Parser) parseNamespaceStatement() *NamespaceStatement {
	stmt := &NamespaceStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	// 解析命名空间名称（可能包含点分隔符）
	// 例如：Mycompany.Myapp.Models
	namespaceParts := []string{p.curToken.Literal}
	
	// 继续读取点分隔的部分
	for p.peekTokenIs(lexer.DOT) {
		p.nextToken() // 跳过 DOT
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		namespaceParts = append(namespaceParts, p.curToken.Literal)
	}

	// 合并为完整的命名空间名称
	fullName := namespaceParts[0]
	for i := 1; i < len(namespaceParts); i++ {
		fullName += "." + namespaceParts[i]
	}

	stmt.Name = &Identifier{
		Token: p.curToken,
		Value: fullName,
	}

	// 可选的分号
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseUseStatement 解析 use 导入语句
// 语法：use Full.Qualified.ClassName 或 use Namespace.ClassName as Alias
func (p *Parser) parseUseStatement() *UseStatement {
	stmt := &UseStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	// 解析完全限定名（可能包含点分隔符）
	// 例如：Illuminate.Database.Eloquent.Model
	pathParts := []string{p.curToken.Literal}
	
	// 继续读取点分隔的部分
	for p.peekTokenIs(lexer.DOT) {
		p.nextToken() // 跳过 DOT
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		pathParts = append(pathParts, p.curToken.Literal)
	}

	// 合并为完整的路径
	fullPath := pathParts[0]
	for i := 1; i < len(pathParts); i++ {
		fullPath += "." + pathParts[i]
	}

	stmt.Path = &Identifier{
		Token: p.curToken,
		Value: fullPath,
	}

	// 检查是否有别名（as Alias）
	if p.peekTokenIs(lexer.IDENT) && p.peekToken.Literal == "as" {
		p.nextToken() // 跳过 "as"
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		stmt.Alias = &Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
	}

	// 可选的分号
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

