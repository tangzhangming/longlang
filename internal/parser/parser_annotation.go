package parser

import (
	"fmt"

	"github.com/tangzhangming/longlang/internal/lexer"
)

// ========== 注解解析 ==========

// parseAnnotations 解析注解列表
// 注解以 @ 开头，可以有多个连续的注解
// 语法: @Annotation 或 @Annotation(param1: value1, param2: value2)
func (p *Parser) parseAnnotations() []*Annotation {
	annotations := []*Annotation{}

	for p.curTokenIs(lexer.AT) {
		ann := p.parseAnnotation()
		if ann != nil {
			annotations = append(annotations, ann)
		}
		// 跳过可能的换行或分号
		for p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}
	}

	return annotations
}

// parseAnnotation 解析单个注解
// 语法: @AnnotationName 或 @AnnotationName(param1: value1, param2: value2)
func (p *Parser) parseAnnotation() *Annotation {
	ann := &Annotation{
		Token:     p.curToken, // @
		Arguments: make(map[string]Expression),
		ArgOrder:  []string{},
	}

	// 期望注解名称
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	ann.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 检查是否有参数
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // 移动到 (
		p.parseAnnotationArguments(ann)
		// parseAnnotationArguments 结束后 curToken 应该在最后一个值或 ) 上
		// 如果是空参数，curToken 是 )；如果有参数，curToken 在最后一个值上
		if !p.curTokenIs(lexer.RPAREN) {
			if !p.expectPeek(lexer.RPAREN) {
				return nil
			}
		}
		p.nextToken() // 跳过 )，移动到下一个 token
	} else {
		p.nextToken() // 没有参数，移动到下一个 token
	}

	return ann
}

// parseAnnotationArguments 解析注解参数
// 语法: param1: value1, param2: value2
// 调用时 curToken 应该是 (
func (p *Parser) parseAnnotationArguments(ann *Annotation) {
	p.nextToken() // 跳过 (，移到第一个参数名或 )

	if p.curTokenIs(lexer.RPAREN) {
		// 空参数，回退一步让调用者处理 )
		return
	}

	for {
		// 解析参数名
		if !p.curTokenIs(lexer.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf("期望注解参数名，得到 %s (行 %d, 列 %d)", p.curToken.Type, p.curToken.Line, p.curToken.Column))
			return
		}
		paramName := p.curToken.Literal

		// 期望冒号
		if !p.expectPeek(lexer.COLON) {
			return
		}

		// 解析参数值
		p.nextToken()
		value := p.parseExpression(LOWEST)
		if value == nil {
			return
		}

		ann.Arguments[paramName] = value
		ann.ArgOrder = append(ann.ArgOrder, paramName)

		// 检查是否还有更多参数
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // 跳过逗号
			p.nextToken() // 移到下一个参数
		} else {
			break
		}
	}
}

// parseAnnotationDefinition 解析注解定义
// 语法: annotation AnnotationName { field1 type = default, ... }
func (p *Parser) parseAnnotationDefinition(annotations []*Annotation) *AnnotationDefinition {
	def := &AnnotationDefinition{
		Token:       p.curToken, // annotation
		Annotations: annotations,
	}

	// 期望注解名称
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	def.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 期望 {
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	// 解析字段
	def.Fields = p.parseAnnotationFields()

	return def
}

// parseAnnotationFields 解析注解字段列表
func (p *Parser) parseAnnotationFields() []*AnnotationField {
	fields := []*AnnotationField{}

	p.nextToken() // 跳过 {

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// 跳过分号和换行
		if p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		field := p.parseAnnotationField()
		if field != nil {
			fields = append(fields, field)
		}
		p.nextToken()
	}

	return fields
}

// parseAnnotationField 解析单个注解字段
// 语法: fieldName type = defaultValue
func (p *Parser) parseAnnotationField() *AnnotationField {
	if !p.curTokenIs(lexer.IDENT) {
		return nil
	}

	field := &AnnotationField{
		Token: p.curToken,
		Name:  &Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	// 解析类型
	p.nextToken()
	field.Type = p.parseAnnotationTypeExpression()

	// 检查默认值
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // 跳到 =
		p.nextToken() // 跳到值
		field.DefaultValue = p.parseExpression(LOWEST)
	}

	return field
}

// parseAnnotationTypeExpression 解析注解字段的类型表达式
func (p *Parser) parseAnnotationTypeExpression() Expression {
	switch p.curToken.Type {
	case lexer.STRING_TYPE:
		return &Identifier{Token: p.curToken, Value: "string"}
	case lexer.INT_TYPE:
		return &Identifier{Token: p.curToken, Value: "int"}
	case lexer.BOOL_TYPE:
		return &Identifier{Token: p.curToken, Value: "bool"}
	case lexer.FLOAT_TYPE:
		return &Identifier{Token: p.curToken, Value: "float"}
	case lexer.ANY:
		return &Identifier{Token: p.curToken, Value: "any"}
	case lexer.IDENT:
		return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	case lexer.LBRACKET:
		// 数组类型 []type
		return p.parseArrayType()
	default:
		p.errors = append(p.errors, fmt.Sprintf("无效的类型: %s (行 %d, 列 %d)", p.curToken.Literal, p.curToken.Line, p.curToken.Column))
		return nil
	}
}

