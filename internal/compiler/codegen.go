package compiler

import (
	"fmt"
	"strings"

	"github.com/tangzhangming/longlang/internal/parser"
)

// GoCode 生成的 Go 代码结构
type GoCode struct {
	PackageName    string
	Imports        []string
	RuntimeHelpers string   // 运行时辅助函数
	Types          []string // 类型定义（struct、interface、enum）
	Functions      []string // 函数定义
	MainCode       string   // main 函数代码
}

// CodeGen 代码生成器
type CodeGen struct {
	symbolTable      *SymbolTable
	typeMapper       *TypeMapper
	namespaceMapper *NamespaceMapper
	exprConverter    *ExpressionConverter
	stmtConverter    *StatementConverter
	classConverter   *ClassConverter
}

// NewCodeGen 创建新的代码生成器
func NewCodeGen(symbolTable *SymbolTable, typeMapper *TypeMapper, namespaceMapper *NamespaceMapper) *CodeGen {
	exprConverter := NewExpressionConverter(symbolTable, typeMapper, namespaceMapper)
	stmtConverter := NewStatementConverter(exprConverter, symbolTable, typeMapper)
	classConverter := NewClassConverter(stmtConverter, exprConverter, symbolTable, typeMapper)

	return &CodeGen{
		symbolTable:      symbolTable,
		typeMapper:       typeMapper,
		namespaceMapper:   namespaceMapper,
		exprConverter:    exprConverter,
		stmtConverter:    stmtConverter,
		classConverter:   classConverter,
	}
}

// Generate 生成 Go 代码
func (cg *CodeGen) Generate(program *parser.Program) (*GoCode, error) {
	goCode := &GoCode{
		Imports:        []string{},
		RuntimeHelpers: "",
		Types:          []string{},
		Functions:      []string{},
	}

	// 确定包名
	currentNS := cg.symbolTable.GetCurrentScope()
	// 如果有包含 main 方法的类，使用 main 包
	hasMainMethod := false
	for _, stmt := range program.Statements {
		if cs, ok := stmt.(*parser.ClassStatement); ok {
			for _, member := range cs.Members {
				if cm, ok := member.(*parser.ClassMethod); ok {
					if cm.Name != nil && cm.Name.Value == "main" && cm.IsStatic {
						hasMainMethod = true
						break
					}
				}
			}
		}
	}
	if hasMainMethod {
		goCode.PackageName = "main"
	} else if cg.namespaceMapper.IsMainPackage(currentNS) {
		goCode.PackageName = "main"
	} else {
		goCode.PackageName = cg.namespaceMapper.GetGoPackageName(currentNS)
	}

	// 收集 imports
	importsMap := make(map[string]bool)

	// 遍历 AST，生成代码
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *parser.NamespaceStatement:
			// 命名空间声明，已处理
		case *parser.UseStatement:
			// use 语句，检查是否是标准库映射
			// 暂时跳过非标准库的 use（需要更复杂的处理）
			// 只处理基本情况
		case *parser.ClassStatement:
			classCode, err := cg.classConverter.ConvertClass(s)
			if err != nil {
				return nil, err
			}
			goCode.Types = append(goCode.Types, classCode)

			// 检查是否有 main 方法
			for _, member := range s.Members {
				if cm, ok := member.(*parser.ClassMethod); ok {
					if cm.Name != nil && cm.Name.Value == "main" && cm.IsStatic {
						mainCode, err := cg.generateMainFunction(s.Name.Value)
						if err != nil {
							return nil, err
						}
						goCode.MainCode = mainCode
					}
				}
			}
		case *parser.InterfaceStatement:
			interfaceCode, err := cg.classConverter.ConvertInterface(s)
			if err != nil {
				return nil, err
			}
			goCode.Types = append(goCode.Types, interfaceCode)
		case *parser.EnumStatement:
			// 枚举暂时跳过，后续实现
		case *parser.ExpressionStatement:
			// 检查是否是函数定义
			if fl, ok := s.Expression.(*parser.FunctionLiteral); ok && fl.Name != nil {
				funcCode, err := cg.generateFunction(fl)
				if err != nil {
					return nil, err
				}
				goCode.Functions = append(goCode.Functions, funcCode)
			}
		case *parser.LetStatement:
			// 顶层变量声明
			varCode, err := cg.stmtConverter.Convert(s)
			if err != nil {
				return nil, err
			}
			goCode.Functions = append(goCode.Functions, varCode)
		}
	}

	// 添加标准库 imports
	// 始终添加 fmt，因为大部分程序都会用到
	if !importsMap["fmt"] {
		goCode.Imports = append(goCode.Imports, "fmt")
	}
	// 添加 strings 包，因为字符串方法会用到
	if !importsMap["strings"] {
		goCode.Imports = append(goCode.Imports, "strings")
	}

	// 添加运行时辅助函数
	goCode.RuntimeHelpers = cg.generateRuntimeHelpers()

	return goCode, nil
}

// generateRuntimeHelpers 生成运行时辅助函数
func (cg *CodeGen) generateRuntimeHelpers() string {
	var sb strings.Builder
	sb.WriteString("// ========== Runtime Helpers ==========\n\n")
	
	// Exception 基础异常类型
	sb.WriteString("// Exception 基础异常类型\n")
	sb.WriteString("type Exception struct {\n")
	sb.WriteString("\tMessage string\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("func (e *Exception) Init(message string) {\n")
	sb.WriteString("\te.Message = message\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("func (e *Exception) Getmessage() string {\n")
	sb.WriteString("\treturn e.Message\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("func (e *Exception) Error() string {\n")
	sb.WriteString("\treturn e.Message\n")
	sb.WriteString("}\n\n")
	
	// __parseInt
	sb.WriteString("// __parseInt 将字符串转换为整数\n")
	sb.WriteString("func __parseInt(s interface{}) int64 {\n")
	sb.WriteString("\tswitch v := s.(type) {\n")
	sb.WriteString("\tcase string:\n")
	sb.WriteString("\t\tvar result int64\n")
	sb.WriteString("\t\tfmt.Sscanf(v, \"%d\", &result)\n")
	sb.WriteString("\t\treturn result\n")
	sb.WriteString("\tcase int64:\n")
	sb.WriteString("\t\treturn v\n")
	sb.WriteString("\tcase int:\n")
	sb.WriteString("\t\treturn int64(v)\n")
	sb.WriteString("\tcase float64:\n")
	sb.WriteString("\t\treturn int64(v)\n")
	sb.WriteString("\tdefault:\n")
	sb.WriteString("\t\treturn 0\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")
	
	// __parseFloat
	sb.WriteString("// __parseFloat 将字符串转换为浮点数\n")
	sb.WriteString("func __parseFloat(s interface{}) float64 {\n")
	sb.WriteString("\tswitch v := s.(type) {\n")
	sb.WriteString("\tcase string:\n")
	sb.WriteString("\t\tvar result float64\n")
	sb.WriteString("\t\tfmt.Sscanf(v, \"%f\", &result)\n")
	sb.WriteString("\t\treturn result\n")
	sb.WriteString("\tcase float64:\n")
	sb.WriteString("\t\treturn v\n")
	sb.WriteString("\tcase int64:\n")
	sb.WriteString("\t\treturn float64(v)\n")
	sb.WriteString("\tcase int:\n")
	sb.WriteString("\t\treturn float64(v)\n")
	sb.WriteString("\tdefault:\n")
	sb.WriteString("\t\treturn 0\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")
	
	// __typeof
	sb.WriteString("// __typeof 获取值的类型\n")
	sb.WriteString("func __typeof(v interface{}) string {\n")
	sb.WriteString("\tif v == nil {\n")
	sb.WriteString("\t\treturn \"NULL\"\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tswitch v.(type) {\n")
	sb.WriteString("\tcase string:\n")
	sb.WriteString("\t\treturn \"STRING\"\n")
	sb.WriteString("\tcase int, int64, int32:\n")
	sb.WriteString("\t\treturn \"INT\"\n")
	sb.WriteString("\tcase float64, float32:\n")
	sb.WriteString("\t\treturn \"FLOAT\"\n")
	sb.WriteString("\tcase bool:\n")
	sb.WriteString("\t\treturn \"BOOL\"\n")
	sb.WriteString("\tcase []interface{}:\n")
	sb.WriteString("\t\treturn \"ARRAY\"\n")
	sb.WriteString("\tcase map[string]interface{}:\n")
	sb.WriteString("\t\treturn \"MAP\"\n")
	sb.WriteString("\tdefault:\n")
	sb.WriteString("\t\treturn \"OBJECT\"\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")
	
	// __isset
	sb.WriteString("// __isset 检查 map 中是否存在键\n")
	sb.WriteString("func __isset(m interface{}, key interface{}) bool {\n")
	sb.WriteString("\tif m == nil {\n")
	sb.WriteString("\t\treturn false\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tswitch mv := m.(type) {\n")
	sb.WriteString("\tcase map[string]interface{}:\n")
	sb.WriteString("\t\tkeyStr := fmt.Sprint(key)\n")
	sb.WriteString("\t\t_, ok := mv[keyStr]\n")
	sb.WriteString("\t\treturn ok\n")
	sb.WriteString("\tdefault:\n")
	sb.WriteString("\t\treturn false\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")
	
	// __len
	sb.WriteString("// __len 获取集合的长度\n")
	sb.WriteString("func __len(v interface{}) int {\n")
	sb.WriteString("\tif v == nil {\n")
	sb.WriteString("\t\treturn 0\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tswitch vv := v.(type) {\n")
	sb.WriteString("\tcase string:\n")
	sb.WriteString("\t\treturn len(vv)\n")
	sb.WriteString("\tcase []interface{}:\n")
	sb.WriteString("\t\treturn len(vv)\n")
	sb.WriteString("\tcase map[string]interface{}:\n")
	sb.WriteString("\t\treturn len(vv)\n")
	sb.WriteString("\tdefault:\n")
	sb.WriteString("\t\treturn 0\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")
	
	// __createInstance
	sb.WriteString("// __createInstance 运行时创建实例\n")
	sb.WriteString("func __createInstance(className string) interface{} {\n")
	sb.WriteString("\t// TODO: 实现运行时实例创建\n")
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")
	
	// __getMap 安全获取 map 值
	sb.WriteString("// __getMap 安全获取 map[string]interface{} 值\n")
	sb.WriteString("func __getMap(m interface{}, key interface{}) interface{} {\n")
	sb.WriteString("\tif m == nil {\n")
	sb.WriteString("\t\treturn nil\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tkeyStr := fmt.Sprint(key)\n")
	sb.WriteString("\tif mv, ok := m.(map[string]interface{}); ok {\n")
	sb.WriteString("\t\treturn mv[keyStr]\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")
	
	// __setMap 安全设置 map 值
	sb.WriteString("// __setMap 安全设置 map[string]interface{} 值\n")
	sb.WriteString("func __setMap(m interface{}, key interface{}, value interface{}) {\n")
	sb.WriteString("\tkeyStr := fmt.Sprint(key)\n")
	sb.WriteString("\tif mv, ok := m.(map[string]interface{}); ok {\n")
	sb.WriteString("\t\tmv[keyStr] = value\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")
	
	// __getIndex 安全获取数组索引
	sb.WriteString("// __getIndex 安全获取数组索引\n")
	sb.WriteString("func __getIndex(arr interface{}, index int) interface{} {\n")
	sb.WriteString("\tif arr == nil {\n")
	sb.WriteString("\t\treturn nil\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tif av, ok := arr.([]interface{}); ok {\n")
	sb.WriteString("\t\tif index >= 0 && index < len(av) {\n")
	sb.WriteString("\t\t\treturn av[index]\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")
	
	// __getIndexStr 安全获取数组索引并转换为字符串
	sb.WriteString("// __getIndexStr 安全获取数组索引并转换为字符串\n")
	sb.WriteString("func __getIndexStr(arr interface{}, index int) string {\n")
	sb.WriteString("\tv := __getIndex(arr, index)\n")
	sb.WriteString("\tif v == nil {\n")
	sb.WriteString("\t\treturn \"\"\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn fmt.Sprint(v)\n")
	sb.WriteString("}\n\n")
	
	// Reflection helpers
	sb.WriteString("// ========== Reflection Helpers ==========\n\n")
	
	sb.WriteString("// ReflectionGetclassname 获取实例的类名\n")
	sb.WriteString("func ReflectionGetclassname(obj interface{}) string {\n")
	sb.WriteString("\tif obj == nil {\n")
	sb.WriteString("\t\treturn \"\"\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tt := fmt.Sprintf(\"%T\", obj)\n")
	sb.WriteString("\t// 移除指针前缀 *\n")
	sb.WriteString("\tif len(t) > 0 && t[0] == '*' {\n")
	sb.WriteString("\t\tt = t[1:]\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\t// 移除包名前缀 main.\n")
	sb.WriteString("\tif idx := len(\"main.\"); len(t) > idx && t[:idx] == \"main.\" {\n")
	sb.WriteString("\t\tt = t[idx:]\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn t\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("// ReflectionGetfieldvalue 获取字段值 (占位实现)\n")
	sb.WriteString("func ReflectionGetfieldvalue(obj interface{}, fieldName interface{}) interface{} {\n")
	sb.WriteString("\t// TODO: 使用反射实现\n")
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("// ReflectionSetfieldvalue 设置字段值 (占位实现)\n")
	sb.WriteString("func ReflectionSetfieldvalue(obj interface{}, fieldName interface{}, value interface{}) {\n")
	sb.WriteString("\t// TODO: 使用反射实现\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("// ReflectionGetclassannotation 获取类注解 (占位实现)\n")
	sb.WriteString("func ReflectionGetclassannotation(className string, annotationName string) interface{} {\n")
	sb.WriteString("\t// TODO: 实现注解读取\n")
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("// ReflectionGetclassfields 获取类字段 (占位实现)\n")
	sb.WriteString("func ReflectionGetclassfields(className string) map[string]interface{} {\n")
	sb.WriteString("\t// TODO: 实现字段读取\n")
	sb.WriteString("\treturn map[string]interface{}{}\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("// ReflectionHasfieldannotation 检查字段是否有注解 (占位实现)\n")
	sb.WriteString("func ReflectionHasfieldannotation(className string, fieldName string, annotationName string) bool {\n")
	sb.WriteString("\t// TODO: 实现注解检查\n")
	sb.WriteString("\treturn false\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("// ReflectionGetfieldannotation 获取字段注解 (占位实现)\n")
	sb.WriteString("func ReflectionGetfieldannotation(className string, fieldName string, annotationName string) interface{} {\n")
	sb.WriteString("\t// TODO: 实现注解读取\n")
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")
	
	// __toMap 将 interface{} 转换为 map[string]interface{}
	sb.WriteString("// __toMap 安全转换为 map[string]interface{}\n")
	sb.WriteString("func __toMap(v interface{}) map[string]interface{} {\n")
	sb.WriteString("\tif v == nil {\n")
	sb.WriteString("\t\treturn map[string]interface{}{}\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tif m, ok := v.(map[string]interface{}); ok {\n")
	sb.WriteString("\t\treturn m\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn map[string]interface{}{}\n")
	sb.WriteString("}\n\n")
	
	// __toSlice 将 interface{} 转换为 []interface{}
	sb.WriteString("// __toSlice 安全转换为 []interface{}\n")
	sb.WriteString("func __toSlice(v interface{}) []interface{} {\n")
	sb.WriteString("\tif v == nil {\n")
	sb.WriteString("\t\treturn []interface{}{}\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tif s, ok := v.([]interface{}); ok {\n")
	sb.WriteString("\t\treturn s\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn []interface{}{}\n")
	sb.WriteString("}\n\n")
	
	// __toString 安全转换为 string
	sb.WriteString("// __toString 安全转换为 string\n")
	sb.WriteString("func __toString(v interface{}) string {\n")
	sb.WriteString("\tif v == nil {\n")
	sb.WriteString("\t\treturn \"\"\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tif s, ok := v.(string); ok {\n")
	sb.WriteString("\t\treturn s\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn fmt.Sprint(v)\n")
	sb.WriteString("}\n\n")
	
	// __toInt 安全转换为 int64
	sb.WriteString("// __toInt 安全转换为 int64\n")
	sb.WriteString("func __toInt(v interface{}) int64 {\n")
	sb.WriteString("\tif v == nil {\n")
	sb.WriteString("\t\treturn 0\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tswitch n := v.(type) {\n")
	sb.WriteString("\tcase int64:\n")
	sb.WriteString("\t\treturn n\n")
	sb.WriteString("\tcase int:\n")
	sb.WriteString("\t\treturn int64(n)\n")
	sb.WriteString("\tcase float64:\n")
	sb.WriteString("\t\treturn int64(n)\n")
	sb.WriteString("\tcase string:\n")
	sb.WriteString("\t\tvar result int64\n")
	sb.WriteString("\t\tfmt.Sscanf(n, \"%d\", &result)\n")
	sb.WriteString("\t\treturn result\n")
	sb.WriteString("\tdefault:\n")
	sb.WriteString("\t\treturn 0\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")
	
	return sb.String()
}

// generateFunction 生成函数
func (cg *CodeGen) generateFunction(fl *parser.FunctionLiteral) (string, error) {
	funcName := toPascalCase(fl.Name.Value)

	var params []string
	for _, param := range fl.Parameters {
		paramName := toCamelCase(param.Name.Value)
		paramType, err := cg.typeMapper.MapType(param.Type)
		if err != nil {
			return "", err
		}
		params = append(params, fmt.Sprintf("%s %s", paramName, paramType))
	}

	returnType := ""
	if len(fl.ReturnType) > 0 {
		rt, err := cg.typeMapper.MapReturnTypes(fl.ReturnType)
		if err != nil {
			return "", err
		}
		returnType = " " + rt
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("func %s(%s)%s {\n", funcName, strings.Join(params, ", "), returnType))

	if fl.Body != nil {
		body, err := cg.stmtConverter.Convert(fl.Body)
		if err != nil {
			return "", err
		}
		// 移除外层大括号
		body = strings.TrimPrefix(body, "{\n")
		body = strings.TrimSuffix(body, "\n}")
		result.WriteString(body)
	}

	result.WriteString("\n}")

	return result.String(), nil
}

// generateMainFunction 生成 main 函数
func (cg *CodeGen) generateMainFunction(className string) (string, error) {
	var result strings.Builder
	result.WriteString("func main() {\n")
	// 静态方法 Main 直接调用
	result.WriteString("    Main()\n")
	result.WriteString("}\n")
	return result.String(), nil
}

// needsFmtImport 检查是否需要 fmt 包
func (cg *CodeGen) needsFmtImport(program *parser.Program) bool {
	// 简单检查：如果代码中有 fmt. 调用，则需要导入
	code := program.String()
	return strings.Contains(code, "fmt.")
}

