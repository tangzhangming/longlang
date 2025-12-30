package parser

import (
	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== 包和导入解析 ==========

// parsePackageStatement 解析包声明语句
func (p *Parser) parsePackageStatement() *PackageStatement {
	stmt := &PackageStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return stmt
}

// parseImportStatement 解析导入语句
func (p *Parser) parseImportStatement() *ImportStatement {
	stmt := &ImportStatement{Token: p.curToken}

	if !p.expectPeek(lexer.STRING) {
		return nil
	}

	stmt.Path = &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}

	return stmt
}

