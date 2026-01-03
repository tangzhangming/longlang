package compiler

import (
	"fmt"
	"strings"

	"github.com/tangzhangming/longlang/internal/parser"
)

// ClassConverter 类转换器
type ClassConverter struct {
	stmtConverter *StatementConverter
	exprConverter *ExpressionConverter
	symbolTable   *SymbolTable
	typeMapper    *TypeMapper
}

// NewClassConverter 创建新的类转换器
func NewClassConverter(stmtConverter *StatementConverter, exprConverter *ExpressionConverter, symbolTable *SymbolTable, typeMapper *TypeMapper) *ClassConverter {
	return &ClassConverter{
		stmtConverter: stmtConverter,
		exprConverter: exprConverter,
		symbolTable:   symbolTable,
		typeMapper:    typeMapper,
	}
}

// ConvertClass 转换类
func (cc *ClassConverter) ConvertClass(cs *parser.ClassStatement) (string, error) {
	var result strings.Builder

	className := cs.Name.Value
	goTypeName := toPascalCase(className)

	// 获取父类名
	parentClass := ""
	if cs.Parent != nil {
		parentClass = cs.Parent.Value
	}
	goParentName := toPascalCase(parentClass)

	// 生成 struct 定义
	result.WriteString(fmt.Sprintf("type %s struct {\n", goTypeName))

	// 处理继承（嵌入父类）
	if cs.Parent != nil {
		result.WriteString(fmt.Sprintf("    %s\n", goParentName))
	}

	// 收集静态变量，用于生成包级变量
	var staticVars []string

	// 转换字段
	for _, member := range cs.Members {
		switch m := member.(type) {
		case *parser.ClassVariable:
			if m.IsStatic {
				// 静态变量需要生成包级变量
				staticVarStr, err := cc.convertStaticVariable(goTypeName, m)
				if err != nil {
					return "", err
				}
				staticVars = append(staticVars, staticVarStr)
			} else {
				fieldStr, err := cc.convertClassVariable(m)
				if err != nil {
					return "", err
				}
				result.WriteString(fieldStr)
			}
		case *parser.ClassConstant:
			// 常量转换为包级常量
			// 这里暂时跳过，在代码生成器中处理
		}
	}

	result.WriteString("}\n\n")

	// 输出静态变量（包级变量）
	for _, sv := range staticVars {
		result.WriteString(sv + "\n")
	}

	// 转换方法
	for _, member := range cs.Members {
		switch m := member.(type) {
		case *parser.ClassMethod:
			methodStr, err := cc.convertClassMethod(goTypeName, parentClass, m)
			if err != nil {
				return "", err
			}
			result.WriteString(methodStr + "\n")
		}
	}

	return result.String(), nil
}

// convertClassVariable 转换类变量
func (cc *ClassConverter) convertClassVariable(cv *parser.ClassVariable) (string, error) {
	fieldName := toPascalCase(cv.Name.Value)
	goType, err := cc.typeMapper.MapType(cv.Type)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("    %s %s", fieldName, goType))

	if cv.Value != nil {
		value, err := cc.exprConverter.Convert(cv.Value)
		if err != nil {
			return "", err
		}
		result.WriteString(fmt.Sprintf(" = %s", value))
	}

	result.WriteString("\n")
	return result.String(), nil
}

// convertStaticVariable 转换静态变量为包级变量
func (cc *ClassConverter) convertStaticVariable(className string, cv *parser.ClassVariable) (string, error) {
	// 静态变量名格式：ClassName + FieldName
	fieldName := cv.Name.Value
	// 移除开头的下划线（如果有）用于生成Go变量名
	goFieldName := toPascalCase(strings.TrimPrefix(fieldName, "_"))
	varName := className + goFieldName

	goType, err := cc.typeMapper.MapType(cv.Type)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("var %s %s", varName, goType))

	if cv.Value != nil {
		value, err := cc.exprConverter.Convert(cv.Value)
		if err != nil {
			return "", err
		}
		result.WriteString(fmt.Sprintf(" = %s", value))
	}

	result.WriteString("\n")
	return result.String(), nil
}

// convertClassMethod 转换类方法
func (cc *ClassConverter) convertClassMethod(className string, parentClass string, cm *parser.ClassMethod) (string, error) {
	methodName := cm.Name.Value
	if methodName == "__construct" {
		// 构造函数转换为 Init 方法
		methodName = "Init"
	} else {
		methodName = toPascalCase(methodName)
	}

	var result strings.Builder

	// 转换参数
	var params []string
	for _, param := range cm.Parameters {
		paramName := toCamelCase(param.Name.Value)
		paramType, err := cc.typeMapper.MapType(param.Type)
		if err != nil {
			return "", err
		}
		params = append(params, fmt.Sprintf("%s %s", paramName, paramType))
	}

	// 转换返回类型
	returnType := ""
	if len(cm.ReturnType) > 0 {
		rt, err := cc.typeMapper.MapReturnTypes(cm.ReturnType)
		if err != nil {
			return "", err
		}
		returnType = " " + rt
	}

	// 生成方法签名
	receiver := strings.ToLower(string(className[0]))
	if cm.IsStatic {
		// 静态方法转换为包级函数（如果方法名是 main，则保持为 Main）
		if methodName == "Main" {
			result.WriteString(fmt.Sprintf("func %s(%s)%s {\n", methodName, strings.Join(params, ", "), returnType))
		} else {
			result.WriteString(fmt.Sprintf("func %s%s(%s)%s {\n", className, methodName, strings.Join(params, ", "), returnType))
		}
	} else {
		// 实例方法
		result.WriteString(fmt.Sprintf("func (%s *%s) %s(%s)%s {\n", receiver, className, methodName, strings.Join(params, ", "), returnType))
	}

	// 设置 receiver 和当前类信息，用于转换 this 和 super
	cc.exprConverter.SetCurrentReceiver(receiver)
	cc.exprConverter.SetCurrentClass(className, parentClass)

	// 转换方法体
	if cm.Body != nil {
		body, err := cc.stmtConverter.Convert(cm.Body)
		if err != nil {
			return "", err
		}
		// 移除外层的大括号，因为方法签名已经包含
		// 处理各种可能的格式：{\n...\n}, { ... }, 或 {}
		body = strings.TrimSpace(body)
		if strings.HasPrefix(body, "{") {
			body = strings.TrimPrefix(body, "{")
			body = strings.TrimSpace(body)
		}
		if strings.HasSuffix(body, "}") {
			body = strings.TrimSuffix(body, "}")
			body = strings.TrimSpace(body)
		}
		// 移除外层的换行符
		body = strings.TrimPrefix(body, "\n")
		body = strings.TrimSuffix(body, "\n")
		if body != "" {
			result.WriteString(body + "\n")
		}
	}

	result.WriteString("}")

	return result.String(), nil
}

// convertInterface 转换接口
func (cc *ClassConverter) ConvertInterface(is *parser.InterfaceStatement) (string, error) {
	interfaceName := is.Name.Value
	goTypeName := toPascalCase(interfaceName)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("type %s interface {\n", goTypeName))

	for _, method := range is.Methods {
		methodStr, err := cc.convertInterfaceMethod(method)
		if err != nil {
			return "", err
		}
		result.WriteString("    " + methodStr + "\n")
	}

	result.WriteString("}\n")
	return result.String(), nil
}

// convertInterfaceMethod 转换接口方法
func (cc *ClassConverter) convertInterfaceMethod(im *parser.InterfaceMethod) (string, error) {
	methodName := toPascalCase(im.Name.Value)

	var params []string
	for _, param := range im.Parameters {
		paramName := toCamelCase(param.Name.Value)
		paramType, err := cc.typeMapper.MapType(param.Type)
		if err != nil {
			return "", err
		}
		params = append(params, fmt.Sprintf("%s %s", paramName, paramType))
	}

	returnType := ""
	if len(im.ReturnType) > 0 {
		rt, err := cc.typeMapper.MapReturnTypes(im.ReturnType)
		if err != nil {
			return "", err
		}
		returnType = " " + rt
	}

	return fmt.Sprintf("%s(%s)%s", methodName, strings.Join(params, ", "), returnType), nil
}

