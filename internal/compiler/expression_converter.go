package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tangzhangming/longlang/internal/parser"
)

// ExpressionConverter 表达式转换器
type ExpressionConverter struct {
	symbolTable      *SymbolTable
	typeMapper       *TypeMapper
	namespaceMapper *NamespaceMapper
	currentNS        string
	currentReceiver  string // 当前方法的 receiver 名称
	currentClass     string // 当前类名
	currentParent    string // 当前类的父类名
}

// NewExpressionConverter 创建新的表达式转换器
func NewExpressionConverter(symbolTable *SymbolTable, typeMapper *TypeMapper, namespaceMapper *NamespaceMapper) *ExpressionConverter {
	return &ExpressionConverter{
		symbolTable:      symbolTable,
		typeMapper:       typeMapper,
		namespaceMapper:   namespaceMapper,
		currentNS:        symbolTable.GetCurrentScope(),
		currentReceiver:  "this",
	}
}

// SetCurrentReceiver 设置当前方法的 receiver
func (ec *ExpressionConverter) SetCurrentReceiver(receiver string) {
	ec.currentReceiver = receiver
}

// SetCurrentClass 设置当前类名和父类名
func (ec *ExpressionConverter) SetCurrentClass(className, parentClass string) {
	ec.currentClass = className
	ec.currentParent = parentClass
}

// Convert 转换表达式
func (ec *ExpressionConverter) Convert(expr parser.Expression) (string, error) {
	switch e := expr.(type) {
	case *parser.IntegerLiteral:
		return strconv.FormatInt(e.Value, 10), nil
	case *parser.FloatLiteral:
		return strconv.FormatFloat(e.Value, 'f', -1, 64), nil
	case *parser.StringLiteral:
		return strconv.Quote(e.Value), nil
	case *parser.BooleanLiteral:
		if e.Value {
			return "true", nil
		}
		return "false", nil
	case *parser.NullLiteral:
		return "nil", nil
	case *parser.Identifier:
		return ec.convertIdentifier(e)
	case *parser.PrefixExpression:
		return ec.convertPrefixExpression(e)
	case *parser.InfixExpression:
		return ec.convertInfixExpression(e)
	case *parser.TernaryExpression:
		return ec.convertTernaryExpression(e)
	case *parser.CallExpression:
		return ec.convertCallExpression(e)
	case *parser.MemberAccessExpression:
		return ec.convertMemberAccess(e)
	case *parser.IndexExpression:
		return ec.convertIndexExpression(e)
	case *parser.ArrayLiteral:
		return ec.convertArrayLiteral(e)
	case *parser.TypedArrayLiteral:
		return ec.convertTypedArrayLiteral(e)
	case *parser.MapLiteral:
		return ec.convertMapLiteral(e)
	case *parser.NewExpression:
		return ec.convertNewExpression(e)
	case *parser.ThisExpression:
		return ec.currentReceiver, nil
	case *parser.SuperExpression:
		return "super", nil
	case *parser.StaticCallExpression:
		return ec.convertStaticCall(e)
	case *parser.AssignmentExpression:
		return ec.convertAssignmentExpression(e)
	case *parser.CompoundAssignmentExpression:
		return ec.convertCompoundAssignment(e)
	case *parser.FunctionLiteral:
		return ec.convertFunctionLiteral(e)
	case *parser.InterpolatedStringLiteral:
		return ec.convertInterpolatedString(e)
	case *parser.SliceExpression:
		return ec.convertSliceExpression(e)
	case *parser.StaticAccessExpression:
		return ec.convertStaticAccess(e)
	default:
		return "", fmt.Errorf("未支持的表达式类型: %T", expr)
	}
}

// convertIdentifier 转换标识符
func (ec *ExpressionConverter) convertIdentifier(ident *parser.Identifier) (string, error) {
	name := ident.Value
	
	// 处理特殊标识符
	if specialName := ec.mapSpecialIdentifier(name); specialName != "" {
		return specialName, nil
	}
	
	// 检查是否是变量
	if symbol, ok := ec.symbolTable.GetSymbol(name, ec.currentNS); ok {
		return toCamelCase(symbol.Name), nil
	}
	// 检查是否是命名空间访问（如 fmt.Println）
	parts := strings.Split(name, ".")
	if len(parts) > 1 {
		return ec.convertNamespaceAccess(parts)
	}
	// 默认转换为 camelCase
	return toCamelCase(name), nil
}

// mapSpecialIdentifier 映射特殊标识符
func (ec *ExpressionConverter) mapSpecialIdentifier(name string) string {
	switch name {
	case "__called_class_name":
		// late static binding - 需要运行时支持
		// 暂时返回当前类名作为占位
		if ec.currentClass != "" {
			return fmt.Sprintf("%q", ec.currentClass)
		}
		return `""`
	case "__create_instance":
		// 运行时创建实例的函数 - 需要生成运行时支持代码
		return "__createInstance"
	}
	return ""
}

// convertNamespaceAccess 转换命名空间访问
func (ec *ExpressionConverter) convertNamespaceAccess(parts []string) (string, error) {
	if len(parts) < 2 {
		return strings.Join(parts, "."), nil
	}
	// 例如：fmt.Println -> fmt.Println (Go 中直接使用)
	return strings.Join(parts, "."), nil
}

// convertPrefixExpression 转换前缀表达式
func (ec *ExpressionConverter) convertPrefixExpression(pe *parser.PrefixExpression) (string, error) {
	right, err := ec.Convert(pe.Right)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s%s)", pe.Operator, right), nil
}

// convertInfixExpression 转换中缀表达式
func (ec *ExpressionConverter) convertInfixExpression(ie *parser.InfixExpression) (string, error) {
	left, err := ec.Convert(ie.Left)
	if err != nil {
		return "", err
	}
	right, err := ec.Convert(ie.Right)
	if err != nil {
		return "", err
	}
	
	// 字符串拼接特殊处理
	if ie.Operator == "+" {
		// 检查左侧是否是字符串类型
		leftIsString := ec.isStringExpression(ie.Left)
		rightIsString := ec.isStringExpression(ie.Right)
		
		// 如果一侧是字符串，另一侧可能是 interface{}，需要转换
		if leftIsString || rightIsString {
			// 如果右侧使用了 __getIndex，替换为 __getIndexStr
			if strings.Contains(right, "__getIndex(") {
				right = strings.Replace(right, "__getIndex(", "__getIndexStr(", -1)
			}
			// 如果右侧使用了 __getMap，包装为 fmt.Sprint
			if strings.Contains(right, "__getMap(") && !strings.HasPrefix(right, "fmt.Sprint(") {
				right = fmt.Sprintf("fmt.Sprint(%s)", right)
			}
		}
	}
	
	return fmt.Sprintf("(%s %s %s)", left, ie.Operator, right), nil
}

// isStringExpression 检查表达式是否是字符串类型
func (ec *ExpressionConverter) isStringExpression(expr parser.Expression) bool {
	switch e := expr.(type) {
	case *parser.StringLiteral:
		return true
	case *parser.InterpolatedStringLiteral:
		return true
	case *parser.Identifier:
		// 检查常见的字符串变量名
		strVarNames := []string{"result", "part", "str", "s", "text", "name", "msg", "message"}
		for _, name := range strVarNames {
			if e.Value == name {
				return true
			}
		}
	case *parser.InfixExpression:
		// 如果是字符串拼接操作，则结果是字符串
		if e.Operator == "+" {
			return ec.isStringExpression(e.Left) || ec.isStringExpression(e.Right)
		}
	}
	return false
}

// convertTernaryExpression 转换三目表达式
func (ec *ExpressionConverter) convertTernaryExpression(te *parser.TernaryExpression) (string, error) {
	cond, err := ec.Convert(te.Condition)
	if err != nil {
		return "", err
	}
	trueExpr, err := ec.Convert(te.TrueExpr)
	if err != nil {
		return "", err
	}
	falseExpr, err := ec.Convert(te.FalseExpr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("func() interface{} { if %s { return %s } else { return %s } }()", cond, trueExpr, falseExpr), nil
}

// convertCallExpression 转换函数调用
func (ec *ExpressionConverter) convertCallExpression(ce *parser.CallExpression) (string, error) {
	// 首先检查是否是内置函数
	if ident, ok := ce.Function.(*parser.Identifier); ok {
		builtinCall := ec.mapBuiltinFunction(ident.Value, ce.Arguments)
		if builtinCall != "" {
			return builtinCall, nil
		}
	}

	// 检查是否是成员方法调用 (obj.method())
	if mae, ok := ce.Function.(*parser.MemberAccessExpression); ok {
		// 处理数组/切片方法
		methodName := mae.Member.Value
		obj, err := ec.Convert(mae.Object)
		if err != nil {
			return "", err
		}
		
		var args []string
		for _, arg := range ce.Arguments {
			argStr, err := ec.Convert(arg.Value)
			if err != nil {
				return "", err
			}
			args = append(args, argStr)
		}
		
		// 数组方法映射
		switch methodName {
		case "push", "Push":
			// array.push(item) -> array = append(array, item)
			if len(args) >= 1 {
				return fmt.Sprintf("func() { %s = append(%s, %s) }()", obj, obj, args[0]), nil
			}
		case "pop", "Pop":
			// array.pop() -> 移除并返回最后一个元素
			return fmt.Sprintf("func() interface{} { if len(%s) == 0 { return nil }; v := %s[len(%s)-1]; %s = %s[:len(%s)-1]; return v }()", obj, obj, obj, obj, obj, obj), nil
		case "shift", "Shift":
			// array.shift() -> 移除并返回第一个元素
			return fmt.Sprintf("func() interface{} { if len(%s) == 0 { return nil }; v := %s[0]; %s = %s[1:]; return v }()", obj, obj, obj, obj), nil
		case "length", "Length", "len", "Len":
			return fmt.Sprintf("len(%s)", obj), nil
		}
		
		// 字符串方法映射
		switch methodName {
		case "lower", "Lower", "toLowerCase":
			return fmt.Sprintf("strings.ToLower(%s)", obj), nil
		case "upper", "Upper", "toUpperCase":
			return fmt.Sprintf("strings.ToUpper(%s)", obj), nil
		case "startsWith", "Startswith":
			if len(args) >= 1 {
				return fmt.Sprintf("strings.HasPrefix(%s, %s)", obj, args[0]), nil
			}
		case "endsWith", "Endswith":
			if len(args) >= 1 {
				return fmt.Sprintf("strings.HasSuffix(%s, %s)", obj, args[0]), nil
			}
		case "contains", "Contains":
			if len(args) >= 1 {
				return fmt.Sprintf("strings.Contains(%s, %s)", obj, args[0]), nil
			}
		}
	}

	funcExpr, err := ec.Convert(ce.Function)
	if err != nil {
		return "", err
	}

	var args []string
	for _, arg := range ce.Arguments {
		argStr, err := ec.Convert(arg.Value)
		if err != nil {
			return "", err
		}
		args = append(args, argStr)
	}

	return fmt.Sprintf("%s(%s)", funcExpr, strings.Join(args, ", ")), nil
}

// mapBuiltinFunction 映射内置函数到Go实现
func (ec *ExpressionConverter) mapBuiltinFunction(funcName string, args []parser.CallArgument) string {
	var argStrs []string
	for _, arg := range args {
		argStr, err := ec.Convert(arg.Value)
		if err != nil {
			return ""
		}
		argStrs = append(argStrs, argStr)
	}

	switch funcName {
	case "parseInt":
		if len(argStrs) >= 1 {
			return fmt.Sprintf("__parseInt(%s)", argStrs[0])
		}
	case "parseFloat":
		if len(argStrs) >= 1 {
			return fmt.Sprintf("__parseFloat(%s)", argStrs[0])
		}
	case "toString":
		if len(argStrs) >= 1 {
			return fmt.Sprintf("fmt.Sprint(%s)", argStrs[0])
		}
	case "typeof":
		if len(argStrs) >= 1 {
			return fmt.Sprintf("__typeof(%s)", argStrs[0])
		}
	case "isset":
		if len(argStrs) >= 2 {
			return fmt.Sprintf("__isset(%s, %s)", argStrs[0], argStrs[1])
		}
	case "len":
		if len(argStrs) >= 1 {
			return fmt.Sprintf("__len(%s)", argStrs[0])
		}
	case "__create_instance":
		if len(argStrs) >= 1 {
			return fmt.Sprintf("__createInstance(%s)", argStrs[0])
		}
	}
	return ""
}

// convertMemberAccess 转换成员访问
func (ec *ExpressionConverter) convertMemberAccess(mae *parser.MemberAccessExpression) (string, error) {
	obj, err := ec.Convert(mae.Object)
	if err != nil {
		return "", err
	}
	member := toPascalCase(mae.Member.Value)
	return fmt.Sprintf("%s.%s", obj, member), nil
}

// convertIndexExpression 转换索引表达式
func (ec *ExpressionConverter) convertIndexExpression(ie *parser.IndexExpression) (string, error) {
	left, err := ec.Convert(ie.Left)
	if err != nil {
		return "", err
	}
	index, err := ec.Convert(ie.Index)
	if err != nil {
		return "", err
	}
	// 使用类型断言的安全索引访问
	// 对于 string 索引，使用 __getMap
	if _, ok := ie.Index.(*parser.StringLiteral); ok {
		return fmt.Sprintf("__getMap(%s, %s)", left, index), nil
	}
	// 对于 int 索引，使用 __getIndex  
	if _, ok := ie.Index.(*parser.IntegerLiteral); ok {
		return fmt.Sprintf("__getIndex(%s, int(%s))", left, index), nil
	}
	// 对于标识符索引，检查是否是整数类型的变量（如循环计数器 i, j, k 等）
	if ident, ok := ie.Index.(*parser.Identifier); ok {
		// 常见的整数循环变量名
		intVarNames := []string{"i", "j", "k", "n", "idx", "index", "count"}
		for _, name := range intVarNames {
			if ident.Value == name {
				return fmt.Sprintf("__getIndex(%s, %s)", left, index), nil
			}
		}
		// 其他标识符，假设是字符串键，使用 __getMap
		return fmt.Sprintf("__getMap(%s, %s)", left, index), nil
	}
	// 默认使用原始索引（可能会导致编译错误，但保留原有行为）
	return fmt.Sprintf("%s[%s]", left, index), nil
}

// convertArrayLiteral 转换数组字面量
func (ec *ExpressionConverter) convertArrayLiteral(al *parser.ArrayLiteral) (string, error) {
	var elements []string
	for _, elem := range al.Elements {
		elemStr, err := ec.Convert(elem)
		if err != nil {
			return "", err
		}
		elements = append(elements, elemStr)
	}
	return fmt.Sprintf("[]interface{}{%s}", strings.Join(elements, ", ")), nil
}

// convertMapLiteral 转换 Map 字面量
func (ec *ExpressionConverter) convertMapLiteral(ml *parser.MapLiteral) (string, error) {
	var pairs []string
	for i, key := range ml.Keys {
		keyStr, err := ec.Convert(key)
		if err != nil {
			return "", err
		}
		valueStr, err := ec.Convert(ml.Values[i])
		if err != nil {
			return "", err
		}
		pairs = append(pairs, fmt.Sprintf("%s: %s", keyStr, valueStr))
	}
	return fmt.Sprintf("map[string]interface{}{%s}", strings.Join(pairs, ", ")), nil
}

// convertNewExpression 转换 new 表达式
func (ec *ExpressionConverter) convertNewExpression(ne *parser.NewExpression) (string, error) {
	className := ne.ClassName.Value
	goTypeName := toPascalCase(className)

	var args []string
	for _, arg := range ne.Arguments {
		argStr, err := ec.Convert(arg.Value)
		if err != nil {
			return "", err
		}
		args = append(args, argStr)
	}

	// 如果有构造函数参数，使用 func() *Type { t := &Type{}; t.Init(...); return t }() 模式
	// 这样可以确保 Init 方法被调用
	if len(args) > 0 {
		return fmt.Sprintf("func() *%s { __t := &%s{}; __t.Init(%s); return __t }()", 
			goTypeName, goTypeName, strings.Join(args, ", ")), nil
	}
	// 无参构造，直接创建实例
	return fmt.Sprintf("&%s{}", goTypeName), nil
}

// convertStaticCall 转换静态调用
func (ec *ExpressionConverter) convertStaticCall(sce *parser.StaticCallExpression) (string, error) {
	className := sce.ClassName.Value
	methodName := sce.Method.Value

	var args []string
	for _, arg := range sce.Arguments {
		argStr, err := ec.Convert(arg.Value)
		if err != nil {
			return "", err
		}
		args = append(args, argStr)
	}

	// 处理 super:: 调用
	if className == "super" {
		if ec.currentParent == "" {
			return "", fmt.Errorf("super:: 调用但当前类没有父类")
		}
		goMethodName := methodName
		if methodName == "__construct" {
			goMethodName = "Init"
		} else {
			goMethodName = toPascalCase(methodName)
		}
		// super::method() -> receiver.ParentType.Method()
		return fmt.Sprintf("%s.%s.%s(%s)", ec.currentReceiver, toPascalCase(ec.currentParent), goMethodName, strings.Join(args, ", ")), nil
	}

	// 处理标准库映射
	goCall := ec.mapStaticCall(className, methodName)
	if goCall != "" {
		return fmt.Sprintf("%s(%s)", goCall, strings.Join(args, ", ")), nil
	}

	// 静态方法转换为包级函数或类型方法
	return fmt.Sprintf("%s%s(%s)", toPascalCase(className), toPascalCase(methodName), strings.Join(args, ", ")), nil
}

// mapStaticCall 映射静态调用到 Go 标准库
func (ec *ExpressionConverter) mapStaticCall(className, methodName string) string {
	// Console 类映射
	if className == "Console" {
		switch methodName {
		case "writeLine", "WriteLine":
			return "fmt.Println"
		case "write", "Write":
			return "fmt.Print"
		case "readLine", "ReadLine":
			return "fmt.Scanln"
		}
	}
	return ""
}

// convertAssignmentExpression 转换赋值表达式
func (ec *ExpressionConverter) convertAssignmentExpression(ae *parser.AssignmentExpression) (string, error) {
	right, err := ec.Convert(ae.Right)
	if err != nil {
		return "", err
	}
	
	// 检查左侧是否是索引表达式 (map/array assignment)
	if ie, ok := ae.Left.(*parser.IndexExpression); ok {
		leftObj, err := ec.Convert(ie.Left)
		if err != nil {
			return "", err
		}
		index, err := ec.Convert(ie.Index)
		if err != nil {
			return "", err
		}
		// 使用 __setMap 进行 map 赋值
		return fmt.Sprintf("__setMap(%s, %s, %s)", leftObj, index, right), nil
	}
	
	left, err := ec.Convert(ae.Left)
	if err != nil {
		return "", err
	}
	
	// 如果左侧是字符串变量，右侧可能是 interface{}，添加类型转换
	if ident, ok := ae.Left.(*parser.Identifier); ok {
		// 检查右侧是否是可能返回 interface{} 的表达式
		// （变量赋值且左侧是字符串类型变量）
		if ec.isStringVariableName(ident.Value) {
			// 检查右侧是否是标识符（可能是从 __getMap 赋值的变量）或者 __getMap 调用
			if _, isIdent := ae.Right.(*parser.Identifier); isIdent {
				right = fmt.Sprintf("__toString(%s)", right)
			} else if strings.Contains(right, "__getMap(") {
				right = fmt.Sprintf("__toString(%s)", right)
			}
		}
	}
	
	return fmt.Sprintf("%s = %s", left, right), nil
}

// isStringVariableName 检查变量名是否暗示是字符串类型
func (ec *ExpressionConverter) isStringVariableName(name string) bool {
	nameLower := strings.ToLower(name)
	stringIndicators := []string{"name", "text", "str", "string", "message", "msg", "title", "label", "description", "val", "value", "key", "table", "column", "field"}
	for _, indicator := range stringIndicators {
		if strings.Contains(nameLower, indicator) {
			return true
		}
	}
	return false
}

// convertCompoundAssignment 转换复合赋值表达式
func (ec *ExpressionConverter) convertCompoundAssignment(cae *parser.CompoundAssignmentExpression) (string, error) {
	left, err := ec.Convert(cae.Left)
	if err != nil {
		return "", err
	}
	right, err := ec.Convert(cae.Right)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s %s", left, cae.Operator, right), nil
}

// convertFunctionLiteral 转换函数字面量（匿名函数）
func (ec *ExpressionConverter) convertFunctionLiteral(fl *parser.FunctionLiteral) (string, error) {
	var params []string
	for _, param := range fl.Parameters {
		paramName := toCamelCase(param.Name.Value)
		paramType := "interface{}"
		if param.Type != nil {
			pt, err := ec.typeMapper.MapType(param.Type)
			if err == nil {
				paramType = pt
			}
		}
		params = append(params, fmt.Sprintf("%s %s", paramName, paramType))
	}

	returnType := ""
	if len(fl.ReturnType) > 0 {
		rt, err := ec.typeMapper.MapReturnTypes(fl.ReturnType)
		if err == nil {
			returnType = " " + rt
		}
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("func(%s)%s {\n", strings.Join(params, ", "), returnType))
	
	// 需要StatementConverter来处理函数体
	// 这里暂时返回简化版本
	result.WriteString("}")

	return result.String(), nil
}

// convertInterpolatedString 转换插值字符串
func (ec *ExpressionConverter) convertInterpolatedString(isl *parser.InterpolatedStringLiteral) (string, error) {
	var parts []string
	var fmtArgs []string
	fmtStr := ""

	for _, part := range isl.Parts {
		if part.IsExpr {
			fmtStr += "%v"
			exprStr, err := ec.Convert(part.Expr)
			if err != nil {
				return "", err
			}
			fmtArgs = append(fmtArgs, exprStr)
		} else {
			fmtStr += strings.ReplaceAll(part.Text, "%", "%%")
		}
	}

	if len(fmtArgs) > 0 {
		parts = append(parts, strconv.Quote(fmtStr))
		parts = append(parts, fmtArgs...)
		return fmt.Sprintf("fmt.Sprintf(%s)", strings.Join(parts, ", ")), nil
	}
	return strconv.Quote(fmtStr), nil
}

// convertSliceExpression 转换切片表达式
func (ec *ExpressionConverter) convertSliceExpression(se *parser.SliceExpression) (string, error) {
	left, err := ec.Convert(se.Left)
	if err != nil {
		return "", err
	}

	startStr := ""
	if se.Start != nil {
		startStr, err = ec.Convert(se.Start)
		if err != nil {
			return "", err
		}
	}

	endStr := ""
	if se.End != nil {
		endStr, err = ec.Convert(se.End)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%s[%s:%s]", left, startStr, endStr), nil
}

// convertTypedArrayLiteral 转换带类型的数组字面量
func (ec *ExpressionConverter) convertTypedArrayLiteral(tal *parser.TypedArrayLiteral) (string, error) {
	typeStr, err := ec.typeMapper.MapType(tal.Type)
	if err != nil {
		return "", err
	}

	var elements []string
	for _, elem := range tal.Elements {
		elemStr, err := ec.Convert(elem)
		if err != nil {
			return "", err
		}
		elements = append(elements, elemStr)
	}

	return fmt.Sprintf("%s{%s}", typeStr, strings.Join(elements, ", ")), nil
}

// convertStaticAccess 转换静态访问表达式（常量或静态变量访问）
// 例如：ClassName::CONST_NAME -> ClassNameCONST_NAME 或 package.ClassNameCONST_NAME
// 例如：Model::_connection -> ModelConnection
func (ec *ExpressionConverter) convertStaticAccess(sae *parser.StaticAccessExpression) (string, error) {
	className := sae.ClassName.Value
	memberName := sae.Name.Value
	
	// 如果是 self，使用当前类名
	if className == "self" {
		if ec.currentClass != "" {
			className = ec.currentClass
		} else {
			return "", fmt.Errorf("self:: 调用但不在类上下文中")
		}
	}
	
	// 转换为 Go 的访问格式
	// 在 Go 中，静态变量/常量是包级变量，格式为：ClassName + MemberName
	goClassName := toPascalCase(className)
	// 移除开头的下划线用于生成变量名
	goMemberName := toPascalCase(strings.TrimPrefix(memberName, "_"))
	
	return fmt.Sprintf("%s%s", goClassName, goMemberName), nil
}

