package compiler

import (
	"fmt"
	"strings"

	"github.com/tangzhangming/longlang/internal/parser"
)

// TypeMapper 类型映射器
type TypeMapper struct {
	symbolTable *SymbolTable
}

// NewTypeMapper 创建新的类型映射器
func NewTypeMapper(symbolTable *SymbolTable) *TypeMapper {
	return &TypeMapper{
		symbolTable: symbolTable,
	}
}

// MapType 将 longlang 类型映射到 Go 类型
func (tm *TypeMapper) MapType(typeExpr parser.Expression) (string, error) {
	if typeExpr == nil {
		return "interface{}", nil
	}
	switch t := typeExpr.(type) {
	case *parser.Identifier:
		return tm.mapBasicType(t.Value)
	case *parser.ArrayType:
		return tm.mapArrayType(t)
	case *parser.MapType:
		return tm.mapMapType(t)
	default:
		return "interface{}", nil
	}
}

// mapBasicType 映射基础类型
func (tm *TypeMapper) mapBasicType(longlangType string) (string, error) {
	// 基础类型映射
	switch longlangType {
	case "int":
		return "int64", nil
	case "i8":
		return "int8", nil
	case "i16":
		return "int16", nil
	case "i32":
		return "int32", nil
	case "i64":
		return "int64", nil
	case "uint":
		return "uint64", nil
	case "u8":
		return "uint8", nil
	case "u16":
		return "uint16", nil
	case "u32":
		return "uint32", nil
	case "u64":
		return "uint64", nil
	case "float":
		return "float64", nil
	case "f32":
		return "float32", nil
	case "f64":
		return "float64", nil
	case "string":
		return "string", nil
	case "bool":
		return "bool", nil
	case "any":
		return "interface{}", nil
	case "void":
		return "", nil
	default:
		// 检查是否是类/接口/枚举
		if symbol, ok := tm.symbolTable.GetClass(longlangType, tm.symbolTable.GetCurrentScope()); ok {
			return "*" + symbol.GoType, nil
		}
		if symbol, ok := tm.symbolTable.GetInterface(longlangType, tm.symbolTable.GetCurrentScope()); ok {
			return symbol.GoType, nil
		}
		if symbol, ok := tm.symbolTable.GetEnum(longlangType, tm.symbolTable.GetCurrentScope()); ok {
			return symbol.GoType, nil
		}
		// 未知类型，尝试使用原名称（可能是跨命名空间的类型）
		return "*" + toPascalCase(longlangType), nil
	}
}

// mapArrayType 映射数组类型
func (tm *TypeMapper) mapArrayType(at *parser.ArrayType) (string, error) {
	elementType, err := tm.MapType(at.ElementType)
	if err != nil {
		return "", err
	}

	if at.Size != nil {
		// 固定长度数组
		if size, ok := at.Size.(*parser.IntegerLiteral); ok {
			return fmt.Sprintf("[%d]%s", size.Value, elementType), nil
		}
	}
	// 切片
	return fmt.Sprintf("[]%s", elementType), nil
}

// mapMapType 映射 Map 类型
func (tm *TypeMapper) mapMapType(mt *parser.MapType) (string, error) {
	keyType := "string" // 目前只支持 string 键
	if mt.KeyType != nil {
		kt, err := tm.mapBasicType(mt.KeyType.Value)
		if err == nil {
			keyType = kt
		}
	}

	valueType, err := tm.MapType(mt.ValueType)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("map[%s]%s", keyType, valueType), nil
}

// MapReturnTypes 映射返回类型列表
func (tm *TypeMapper) MapReturnTypes(returnTypes []*parser.Identifier) (string, error) {
	if len(returnTypes) == 0 {
		return "", nil
	}

	if len(returnTypes) == 1 {
		return tm.mapBasicType(returnTypes[0].Value)
	}

	// 多返回值
	var types []string
	for _, rt := range returnTypes {
		t, err := tm.mapBasicType(rt.Value)
		if err != nil {
			return "", err
		}
		types = append(types, t)
	}
	return "(" + strings.Join(types, ", ") + ")", nil
}


