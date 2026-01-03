package compiler

import (
	"fmt"
	"strings"

	"github.com/tangzhangming/longlang/internal/parser"
)

// StatementConverter 语句转换器
type StatementConverter struct {
	exprConverter *ExpressionConverter
	symbolTable   *SymbolTable
	typeMapper    *TypeMapper
	currentNS     string
}

// NewStatementConverter 创建新的语句转换器
func NewStatementConverter(exprConverter *ExpressionConverter, symbolTable *SymbolTable, typeMapper *TypeMapper) *StatementConverter {
	return &StatementConverter{
		exprConverter: exprConverter,
		symbolTable:   symbolTable,
		typeMapper:    typeMapper,
		currentNS:     symbolTable.GetCurrentScope(),
	}
}

// Convert 转换语句
func (sc *StatementConverter) Convert(stmt parser.Statement) (string, error) {
	switch s := stmt.(type) {
	case *parser.LetStatement:
		return sc.convertLetStatement(s)
	case *parser.AssignStatement:
		return sc.convertAssignStatement(s)
	case *parser.ReturnStatement:
		return sc.convertReturnStatement(s)
	case *parser.ExpressionStatement:
		return sc.convertExpressionStatement(s)
	case *parser.BlockStatement:
		return sc.convertBlockStatement(s)
	case *parser.IfStatement:
		return sc.convertIfStatement(s)
	case *parser.ForStatement:
		return sc.convertForStatement(s)
	case *parser.ForRangeStatement:
		return sc.convertForRangeStatement(s)
	case *parser.BreakStatement:
		return "break", nil
	case *parser.ContinueStatement:
		return "continue", nil
	case *parser.IncrementStatement:
		return sc.convertIncrementStatement(s)
	case *parser.ThrowStatement:
		return sc.convertThrowStatement(s)
	default:
		return "", fmt.Errorf("未支持的语句类型: %T", stmt)
	}
}

// convertLetStatement 转换变量声明
func (sc *StatementConverter) convertLetStatement(ls *parser.LetStatement) (string, error) {
	varName := toCamelCase(ls.Name.Value)
	var result strings.Builder

	if ls.Type != nil {
		// 有类型声明
		goType, err := sc.typeMapper.MapType(ls.Type)
		if err != nil {
			return "", err
		}
		if ls.Value != nil {
			// 有初始值
			value, err := sc.exprConverter.Convert(ls.Value)
			if err != nil {
				return "", err
			}
			result.WriteString(fmt.Sprintf("var %s %s = %s", varName, goType, value))
		} else {
			// 无初始值
			result.WriteString(fmt.Sprintf("var %s %s", varName, goType))
		}
	} else {
		// 无类型声明，使用 :=
		if ls.Value == nil {
			return "", fmt.Errorf("变量 %s 没有类型和初始值", varName)
		}
		value, err := sc.exprConverter.Convert(ls.Value)
		if err != nil {
			return "", err
		}
		// 检查是否是 nil 赋值，需要声明类型
		if value == "nil" {
			result.WriteString(fmt.Sprintf("var %s interface{} = nil", varName))
		} else {
			result.WriteString(fmt.Sprintf("%s := %s", varName, value))
		}
	}

	return result.String(), nil
}

// convertAssignStatement 转换赋值语句
func (sc *StatementConverter) convertAssignStatement(as *parser.AssignStatement) (string, error) {
	varName := toCamelCase(as.Name.Value)
	value, err := sc.exprConverter.Convert(as.Value)
	if err != nil {
		return "", err
	}
	// 检查是否是 nil 赋值，需要声明类型
	if value == "nil" {
		return fmt.Sprintf("var %s interface{} = nil", varName), nil
	}
	return fmt.Sprintf("%s := %s", varName, value), nil
}

// convertReturnStatement 转换返回语句
func (sc *StatementConverter) convertReturnStatement(rs *parser.ReturnStatement) (string, error) {
	if rs.ReturnValue == nil {
		return "return", nil
	}
	value, err := sc.exprConverter.Convert(rs.ReturnValue)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("return %s", value), nil
}

// convertExpressionStatement 转换表达式语句
func (sc *StatementConverter) convertExpressionStatement(es *parser.ExpressionStatement) (string, error) {
	return sc.exprConverter.Convert(es.Expression)
}

// convertBlockStatement 转换块语句
func (sc *StatementConverter) convertBlockStatement(bs *parser.BlockStatement) (string, error) {
	var result strings.Builder
	result.WriteString("{\n")
	for _, stmt := range bs.Statements {
		stmtStr, err := sc.Convert(stmt)
		if err != nil {
			return "", err
		}
		result.WriteString("    " + stmtStr + "\n")
	}
	result.WriteString("}")
	return result.String(), nil
}

// convertIfStatement 转换 if 语句
func (sc *StatementConverter) convertIfStatement(is *parser.IfStatement) (string, error) {
	var result strings.Builder

	cond, err := sc.exprConverter.Convert(is.Condition)
	if err != nil {
		return "", err
	}

	consequence, err := sc.convertBlockStatement(is.Consequence)
	if err != nil {
		return "", err
	}

	result.WriteString(fmt.Sprintf("if %s %s", cond, consequence))

	if is.Alternative != nil {
		alternative, err := sc.convertBlockStatement(is.Alternative)
		if err != nil {
			return "", err
		}
		result.WriteString(fmt.Sprintf(" else %s", alternative))
	}

	if is.ElseIf != nil {
		elseIf, err := sc.convertIfStatement(is.ElseIf)
		if err != nil {
			return "", err
		}
		result.WriteString(fmt.Sprintf(" else %s", elseIf))
	}

	return result.String(), nil
}

// convertForStatement 转换 for 语句
func (sc *StatementConverter) convertForStatement(fs *parser.ForStatement) (string, error) {
	var result strings.Builder
	result.WriteString("for ")

	if fs.Init != nil {
		init, err := sc.Convert(fs.Init)
		if err != nil {
			return "", err
		}
		result.WriteString(init + "; ")
	}

	if fs.Condition != nil {
		cond, err := sc.exprConverter.Convert(fs.Condition)
		if err != nil {
			return "", err
		}
		result.WriteString(cond + "; ")
	}

	if fs.Post != nil {
		post, err := sc.Convert(fs.Post)
		if err != nil {
			return "", err
		}
		result.WriteString(post)
	}

	body, err := sc.convertBlockStatement(fs.Body)
	if err != nil {
		return "", err
	}
	result.WriteString(" " + body)

	return result.String(), nil
}

// convertForRangeStatement 转换 for-range 语句
func (sc *StatementConverter) convertForRangeStatement(frs *parser.ForRangeStatement) (string, error) {
	var result strings.Builder
	result.WriteString("for ")

	iterable, err := sc.exprConverter.Convert(frs.Iterable)
	if err != nil {
		return "", err
	}

	// 先转换循环体，用于检测变量使用
	body, err := sc.convertBlockStatement(frs.Body)
	if err != nil {
		return "", err
	}

	// 根据键/值类型推断需要的转换
	// 如果有 key 和 value，通常是 map 迭代
	// 如果只有 key (或 value)，可能是数组迭代
	rangeExpr := iterable
	if frs.Key != nil && frs.Value != nil {
		// 假设是 map 迭代，使用 __toMap
		rangeExpr = fmt.Sprintf("__toMap(%s)", iterable)
	}

	if frs.Key != nil && frs.Value != nil {
		key := toCamelCase(frs.Key.Value)
		value := toCamelCase(frs.Value.Value)
		// 检查 value 变量是否在循环体中使用
		valueUsed := strings.Contains(body, value)
		if valueUsed {
			result.WriteString(fmt.Sprintf("%s, %s := range %s", key, value, rangeExpr))
		} else {
			// 值未使用，用 _ 代替
			result.WriteString(fmt.Sprintf("%s, _ := range %s", key, rangeExpr))
		}
	} else if frs.Key != nil {
		key := toCamelCase(frs.Key.Value)
		result.WriteString(fmt.Sprintf("%s := range %s", key, rangeExpr))
	} else {
		result.WriteString(fmt.Sprintf("range %s", rangeExpr))
	}

	result.WriteString(" " + body)

	return result.String(), nil
}

// convertIncrementStatement 转换自增/自减语句
func (sc *StatementConverter) convertIncrementStatement(inc *parser.IncrementStatement) (string, error) {
	varName := toCamelCase(inc.Name.Value)
	return fmt.Sprintf("%s%s", varName, inc.Operator), nil
}

// convertThrowStatement 转换 throw 语句
// 在 Go 中，throw 通常转换为 panic 或返回 error
func (sc *StatementConverter) convertThrowStatement(ts *parser.ThrowStatement) (string, error) {
	if ts.Value == nil {
		return "panic(nil)", nil
	}
	
	exception, err := sc.exprConverter.Convert(ts.Value)
	if err != nil {
		return "", err
	}
	
	// 转换为 panic
	return fmt.Sprintf("panic(%s)", exception), nil
}

