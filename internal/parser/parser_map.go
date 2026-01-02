package parser

import (
	"fmt"

	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== Map 类型和字面量解析 ==========

// parseMapLiteral 解析 Map 字面量
// 语法：map[KeyType]ValueType{key1: value1, key2: value2, ...}
// 例如：map[string]int{"Alice": 100, "Bob": 90}
func (p *Parser) parseMapLiteral() Expression {
	lit := &MapLiteral{Token: p.curToken}

	// 解析 Map 类型
	mapType := p.parseMapType()
	if mapType == nil {
		return nil
	}
	lit.Type = mapType

	// 期望 {
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	// 解析键值对
	lit.Keys = []Expression{}
	lit.Values = []Expression{}
	lit.Pairs = make(map[Expression]Expression)

	// 空 Map
	if p.peekTokenIs(lexer.RBRACE) {
		p.nextToken()
		return lit
	}

	// 解析第一个键值对
	p.nextToken()
	key, value := p.parseMapPair()
	if key == nil || value == nil {
		return nil
	}
	lit.Keys = append(lit.Keys, key)
	lit.Values = append(lit.Values, value)
	lit.Pairs[key] = value

	// 解析剩余键值对
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // 跳过 ,
		
		// 允许尾随逗号
		if p.peekTokenIs(lexer.RBRACE) {
			break
		}
		
		p.nextToken()
		key, value := p.parseMapPair()
		if key == nil || value == nil {
			return nil
		}
		lit.Keys = append(lit.Keys, key)
		lit.Values = append(lit.Values, value)
		lit.Pairs[key] = value
	}

	// 期望 }
	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}

	return lit
}

// parseMapType 解析 Map 类型声明
// 语法：map[KeyType]ValueType
// 当前 curToken 应该是 map
func (p *Parser) parseMapType() *MapType {
	mt := &MapType{Token: p.curToken}

	// 期望 [
	if !p.expectPeek(lexer.LBRACKET) {
		return nil
	}

	// 解析键类型（支持 IDENT 和类型关键字）
	p.nextToken()
	if !p.curTokenIsType() {
		p.errors = append(p.errors, fmt.Sprintf("期望键类型，得到 %s (行 %d, 列 %d)",
			p.curToken.Type, p.curToken.Line, p.curToken.Column))
		return nil
	}
	mt.KeyType = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 期望 ]
	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	// 解析值类型
	p.nextToken()
	mt.ValueType = p.parseTypeExpression()
	if mt.ValueType == nil {
		p.errors = append(p.errors, fmt.Sprintf("期望值类型，得到 %s (行 %d, 列 %d)",
			p.curToken.Type, p.curToken.Line, p.curToken.Column))
		return nil
	}

	return mt
}

// parseMapPair 解析单个键值对
// 语法：key: value
// 当前 curToken 应该是 key
func (p *Parser) parseMapPair() (Expression, Expression) {
	// 解析键（只允许字符串字面量）
	if !p.curTokenIs(lexer.STRING) {
		p.errors = append(p.errors, fmt.Sprintf("Map 的键必须是字符串，得到 %s (行 %d, 列 %d)",
			p.curToken.Type, p.curToken.Line, p.curToken.Column))
		return nil, nil
	}
	key := p.parseStringLiteral()

	// 期望 :
	if !p.expectPeek(lexer.COLON) {
		return nil, nil
	}

	// 解析值
	p.nextToken()
	value := p.parseExpression(LOWEST)
	if value == nil {
		return nil, nil
	}

	return key, value
}

// parseTypeExpression 解析类型表达式
// 支持：int, string, User, []int, map[string]int 等
func (p *Parser) parseTypeExpression() Expression {
	switch p.curToken.Type {
	case lexer.IDENT:
		return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	case lexer.INT_TYPE:
		return &Identifier{Token: p.curToken, Value: "int"}
	case lexer.STRING_TYPE:
		return &Identifier{Token: p.curToken, Value: "string"}
	case lexer.BYTE_TYPE:
		return &Identifier{Token: p.curToken, Value: "byte"}
	case lexer.U8_TYPE:
		return &Identifier{Token: p.curToken, Value: "u8"}
	case lexer.BOOL_TYPE:
		return &Identifier{Token: p.curToken, Value: "bool"}
	case lexer.FLOAT_TYPE:
		return &Identifier{Token: p.curToken, Value: "float"}
	case lexer.ANY:
		return &Identifier{Token: p.curToken, Value: "any"}
	case lexer.LBRACKET:
		// 数组类型 []Type
		return p.parseArrayTypeOrLiteral()
	case lexer.MAP:
		// 嵌套 Map 类型 map[K]V
		return p.parseMapType()
	default:
		return nil
	}
}

