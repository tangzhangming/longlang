package parser

import (
	"fmt"
	"strconv"

	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== 数组类型和字面量解析 ==========

// parseArrayLiteral 解析数组字面量 {element1, element2, ...}
// 当 { 作为表达式开始时调用
func (p *Parser) parseArrayLiteral() Expression {
	lit := &ArrayLiteral{Token: p.curToken}
	lit.Elements = p.parseExpressionList(lexer.RBRACE)
	return lit
}

// parseArrayTypeOrLiteral 解析数组类型或带类型的数组字面量
// 当 [ 作为表达式开始时调用
// 可能是：
//   - [5]int{1,2,3} - 固定长度数组
//   - []int{1,2,3} - 切片
//   - [...]int{1,2,3} - 长度推导数组
func (p *Parser) parseArrayTypeOrLiteral() Expression {
	startToken := p.curToken

	// 解析数组类型
	arrayType := p.parseArrayType()
	if arrayType == nil {
		return nil
	}

	// 检查是否有字面量初始化 {
	if p.peekTokenIs(lexer.LBRACE) {
		p.nextToken() // 跳到 {
		lit := &TypedArrayLiteral{
			Token:    startToken,
			Type:     arrayType,
			Elements: p.parseExpressionList(lexer.RBRACE),
		}
		return lit
	}

	// 只有类型没有字面量（可能用于类型声明）
	return arrayType
}

// parseArrayType 解析数组类型 [size]elementType 或 []elementType 或 [...]elementType
func (p *Parser) parseArrayType() *ArrayType {
	arrayType := &ArrayType{Token: p.curToken}

	// 当前在 [，看下一个 token
	p.nextToken()

	if p.curTokenIs(lexer.RBRACKET) {
		// []type - 切片类型
		arrayType.Size = nil
		arrayType.IsInferred = false
	} else if p.curTokenIs(lexer.ELLIPSIS) {
		// [...]type - 长度推导数组
		arrayType.IsInferred = true
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}
	} else if p.curTokenIs(lexer.INT) {
		// [size]type - 固定长度数组
		size, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
		if err != nil {
			p.errors = append(p.errors, fmt.Sprintf("无法解析数组长度 %q (行 %d, 列 %d)", 
				p.curToken.Literal, p.curToken.Line, p.curToken.Column))
			return nil
		}
		arrayType.Size = &IntegerLiteral{Token: p.curToken, Value: size}
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}
	} else {
		p.errors = append(p.errors, fmt.Sprintf("数组类型格式错误，期望数字、']' 或 '...'，得到 %s (行 %d, 列 %d)", 
			p.curToken.Type, p.curToken.Line, p.curToken.Column))
		return nil
	}

	// 解析元素类型
	p.nextToken()
	arrayType.ElementType = p.parseElementType()

	return arrayType
}

// parseElementType 解析数组元素类型
// 可以是简单类型（int, string 等）或嵌套数组类型
func (p *Parser) parseElementType() Expression {
	// 检查是否是嵌套数组类型
	if p.curTokenIs(lexer.LBRACKET) {
		return p.parseArrayType()
	}

	// 检查是否是基本类型
	if p.isTypeToken(p.curToken.Type) {
		return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	// 检查是否是自定义类型（标识符）
	if p.curTokenIs(lexer.IDENT) {
		return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	p.errors = append(p.errors, fmt.Sprintf("期望元素类型，得到 %s (行 %d, 列 %d)", 
		p.curToken.Type, p.curToken.Line, p.curToken.Column))
	return nil
}

// parseIndexExpression 解析索引访问表达式 array[index] 或切片表达式 array[start:end]
// Go风格切片语法：
//   - array[start:end]  从 start 到 end-1
//   - array[start:]     从 start 到末尾
//   - array[:end]       从开头到 end-1
//   - array[:]          整个数组的副本
func (p *Parser) parseIndexExpression(left Expression) Expression {
	token := p.curToken // [
	p.nextToken()       // 跳过 [

	// 检查是否是 [:...] 形式（无起始索引）
	if p.curTokenIs(lexer.COLON) {
		// 切片表达式，无起始索引
		return p.parseSliceExpression(left, token, nil)
	}

	// 解析第一个表达式（可能是索引或切片的起始）
	firstExpr := p.parseExpression(LOWEST)

	// 检查下一个 token 是否是 :（切片语法）
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // 跳到 :
		return p.parseSliceExpression(left, token, firstExpr)
	}

	// 普通索引表达式
	exp := &IndexExpression{Token: token, Left: left, Index: firstExpr}

	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	return exp
}

// parseSliceExpression 解析切片表达式
// 调用时 curToken 应该是 :
func (p *Parser) parseSliceExpression(left Expression, token lexer.Token, start Expression) Expression {
	slice := &SliceExpression{
		Token: token,
		Left:  left,
		Start: start,
	}

	p.nextToken() // 跳过 :

	// 检查是否有结束索引
	if !p.curTokenIs(lexer.RBRACKET) {
		slice.End = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}
	}

	return slice
}

// parseExpressionList 解析表达式列表，用于数组字面量
func (p *Parser) parseExpressionList(end lexer.TokenType) []Expression {
	list := []Expression{}

	// 空列表
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // 跳过当前表达式
		p.nextToken() // 跳过逗号
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}


