package compiler

import (
	"fmt"

	"github.com/tangzhangming/longlang/internal/parser"
)

// Analyzer AST 分析器
type Analyzer struct {
	symbolTable *SymbolTable
	currentNS  string // 当前命名空间
}

// NewAnalyzer 创建新的分析器
func NewAnalyzer(symbolTable *SymbolTable) *Analyzer {
	return &Analyzer{
		symbolTable: symbolTable,
	}
}

// Analyze 分析 AST 程序
func (a *Analyzer) Analyze(program *parser.Program) error {
	// 第一遍：收集命名空间和类型定义
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *parser.NamespaceStatement:
			a.analyzeNamespace(s)
		case *parser.ClassStatement:
			a.analyzeClass(s)
		case *parser.InterfaceStatement:
			a.analyzeInterface(s)
		case *parser.EnumStatement:
			a.analyzeEnum(s)
		}
	}

	// 第二遍：收集函数和变量
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *parser.ExpressionStatement:
			if fl, ok := s.Expression.(*parser.FunctionLiteral); ok && fl.Name != nil {
				a.analyzeFunction(fl)
			}
		case *parser.LetStatement:
			a.analyzeVariable(s)
		}
	}

	return nil
}

// analyzeNamespace 分析命名空间
func (a *Analyzer) analyzeNamespace(ns *parser.NamespaceStatement) {
	namespace := ns.Name.Value
	a.currentNS = namespace
	a.symbolTable.AddNamespace(namespace)
	a.symbolTable.SetCurrentScope(namespace)

	symbol := &Symbol{
		Name:       namespace,
		Type:       SymbolTypeNamespace,
		Namespace:  "",
		IsExported: true,
		Node:       ns,
	}
	a.symbolTable.AddSymbol(symbol)
}

// analyzeClass 分析类
func (a *Analyzer) analyzeClass(cs *parser.ClassStatement) {
	className := cs.Name.Value
	goTypeName := toPascalCase(className)

	symbol := &Symbol{
		Name:       className,
		Type:       SymbolTypeClass,
		GoType:     goTypeName,
		Namespace:  a.currentNS,
		IsExported: cs.IsPublic,
		Node:       cs,
	}
	a.symbolTable.AddSymbol(symbol)
}

// analyzeInterface 分析接口
func (a *Analyzer) analyzeInterface(is *parser.InterfaceStatement) {
	interfaceName := is.Name.Value
	goTypeName := toPascalCase(interfaceName)

	symbol := &Symbol{
		Name:       interfaceName,
		Type:       SymbolTypeInterface,
		GoType:     goTypeName,
		Namespace:  a.currentNS,
		IsExported: is.IsPublic,
		Node:       is,
	}
	a.symbolTable.AddSymbol(symbol)
}

// analyzeEnum 分析枚举
func (a *Analyzer) analyzeEnum(es *parser.EnumStatement) {
	enumName := es.Name.Value
	goTypeName := toPascalCase(enumName)

	symbol := &Symbol{
		Name:       enumName,
		Type:       SymbolTypeEnum,
		GoType:     goTypeName,
		Namespace:  a.currentNS,
		IsExported: es.IsPublic,
		Node:       es,
	}
	a.symbolTable.AddSymbol(symbol)
}

// analyzeFunction 分析函数
func (a *Analyzer) analyzeFunction(fl *parser.FunctionLiteral) {
	if fl.Name == nil {
		return
	}

	funcName := fl.Name.Value
	goFuncName := toPascalCase(funcName)

	symbol := &Symbol{
		Name:       funcName,
		Type:       SymbolTypeFunction,
		GoType:     goFuncName,
		Namespace:  a.currentNS,
		IsExported: true, // 函数默认导出
		Node:       fl,
	}
	a.symbolTable.AddSymbol(symbol)
}

// analyzeVariable 分析变量
func (a *Analyzer) analyzeVariable(ls *parser.LetStatement) {
	varName := ls.Name.Value
	goVarName := toCamelCase(varName)

	symbol := &Symbol{
		Name:       varName,
		Type:       SymbolTypeVariable,
		GoType:     goVarName,
		Namespace:  a.currentNS,
		IsExported: false, // 变量默认不导出
		Node:       ls,
	}
	a.symbolTable.AddSymbol(symbol)
}

// GetTypeInfo 获取类型信息
func (a *Analyzer) GetTypeInfo(typeExpr parser.Expression) (string, error) {
	switch t := typeExpr.(type) {
	case *parser.Identifier:
		typeName := t.Value
		// 检查是否是基础类型
		if goType := mapBasicType(typeName); goType != "" {
			return goType, nil
		}
		// 检查是否是类/接口/枚举
		if symbol, ok := a.symbolTable.GetClass(typeName, a.currentNS); ok {
			return symbol.GoType, nil
		}
		if symbol, ok := a.symbolTable.GetInterface(typeName, a.currentNS); ok {
			return symbol.GoType, nil
		}
		if symbol, ok := a.symbolTable.GetEnum(typeName, a.currentNS); ok {
			return symbol.GoType, nil
		}
		return typeName, nil // 未知类型，返回原名称
	case *parser.ArrayType:
		elementType, err := a.GetTypeInfo(t.ElementType)
		if err != nil {
			return "", err
		}
		if t.Size != nil {
			// 固定长度数组
			if size, ok := t.Size.(*parser.IntegerLiteral); ok {
				return fmt.Sprintf("[%d]%s", size.Value, elementType), nil
			}
		}
		// 切片
		return fmt.Sprintf("[]%s", elementType), nil
	case *parser.MapType:
		keyType := "string" // 目前只支持 string 键
		valueType, err := a.GetTypeInfo(t.ValueType)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("map[%s]%s", keyType, valueType), nil
	default:
		return "", fmt.Errorf("未知类型表达式: %T", typeExpr)
	}
}

// mapBasicType 映射基础类型
func mapBasicType(longlangType string) string {
	switch longlangType {
	case "int":
		return "int64"
	case "float":
		return "float64"
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "any":
		return "interface{}"
	case "void":
		return ""
	default:
		return ""
	}
}


