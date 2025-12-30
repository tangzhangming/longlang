package interpreter

import (
	"fmt"
	"github.com/tangzhangming/longlang/internal/parser"
)

// Interpreter 解释器，负责执行 AST 节点
type Interpreter struct {
	env *Environment // 当前作用域环境
}

// New 创建新解释器并初始化内置函数
func New() *Interpreter {
	env := NewEnvironment()
	// 注册内置函数（如 fmt.Println）
	registerBuiltins(env)
	return &Interpreter{env: env}
}

// Eval 执行 AST 节点，根据节点类型分发到相应的处理函数
func (i *Interpreter) Eval(node parser.Node) Object {
	switch node := node.(type) {
	case *parser.Program:
		return i.evalProgram(node)
	case *parser.PackageStatement:
		// 包声明目前只是声明性的，不执行任何操作
		return nil
	case *parser.ImportStatement:
		// 导入语句目前只是声明性的，不执行任何操作
		return nil
	case *parser.ClassStatement:
		// 类定义，注册到环境中
		return i.evalClassStatement(node)
	case *parser.ExpressionStatement:
		// 检查是否是函数定义
		if fl, ok := node.Expression.(*parser.FunctionLiteral); ok && fl.Name != nil {
			// 函数定义，存储到环境中
			fn := i.evalFunctionLiteral(fl)
			if isError(fn) {
				return fn
			}
			i.env.Set(fl.Name.Value, fn)
			return fn
		}
		return i.Eval(node.Expression)
	case *parser.LetStatement:
		val := i.Eval(node.Value)
		if isError(val) {
			return val
		}
		i.env.Set(node.Name.Value, val)
		return val
	case *parser.AssignStatement:
		val := i.Eval(node.Value)
		if isError(val) {
			return val
		}
		i.env.Set(node.Name.Value, val)
		return val
	case *parser.ReturnStatement:
		if node.ReturnValue == nil {
			return &ReturnValue{Value: &Null{}}
		}
		val := i.Eval(node.ReturnValue)
		if isError(val) {
			return val
		}
		return &ReturnValue{Value: val}
	case *parser.BlockStatement:
		return i.evalBlockStatement(node)
	case *parser.IfStatement:
		return i.evalIfStatement(node)
	case *parser.ForStatement:
		return i.evalForStatement(node)
	case *parser.BreakStatement:
		return &BreakSignal{}
	case *parser.ContinueStatement:
		return &ContinueSignal{}
	case *parser.IncrementStatement:
		return i.evalIncrementStatement(node)
	case *parser.IntegerLiteral:
		return &Integer{Value: node.Value}
	case *parser.FloatLiteral:
		return &Float{Value: node.Value}
	case *parser.StringLiteral:
		return &String{Value: node.Value}
	case *parser.BooleanLiteral:
		return &Boolean{Value: node.Value}
	case *parser.NullLiteral:
		return &Null{}
	case *parser.Identifier:
		return i.evalIdentifier(node)
	case *parser.PrefixExpression:
		right := i.Eval(node.Right)
		if isError(right) {
			return right
		}
		return i.evalPrefixExpression(node.Operator, right)
	case *parser.InfixExpression:
		left := i.Eval(node.Left)
		if isError(left) {
			return left
		}
		right := i.Eval(node.Right)
		if isError(right) {
			return right
		}
		return i.evalInfixExpression(node.Operator, left, right)
	case *parser.TernaryExpression:
		return i.evalTernaryExpression(node)
	case *parser.FunctionLiteral:
		return i.evalFunctionLiteral(node)
	case *parser.CallExpression:
		// 处理成员访问（如 fmt.Println）
		if ident, ok := node.Function.(*parser.Identifier); ok {
			parts := splitIdentifier(ident.Value)
			if len(parts) == 2 {
				// 命名空间访问，如 fmt.Println
				namespace, ok := i.env.Get(parts[0])
				if !ok {
					return newError("未定义的命名空间: %s", parts[0])
				}
				if builtinObj, ok := namespace.(*BuiltinObject); ok {
					member, ok := builtinObj.GetField(parts[1])
					if !ok {
						return newError("命名空间 %s 中没有成员 %s", parts[0], parts[1])
					}
					args := i.evalExpressions(node.Arguments)
					if len(args) == 1 && isError(args[0]) {
						return args[0]
					}
					return i.applyFunction(member, args, node.Arguments)
				}
			}
		}
		function := i.Eval(node.Function)
		if isError(function) {
			return function
		}
		args := i.evalExpressions(node.Arguments)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return i.applyFunction(function, args, node.Arguments)
	case *parser.NewExpression:
		return i.evalNewExpression(node)
	case *parser.MemberAccessExpression:
		return i.evalMemberAccessExpression(node)
	case *parser.AssignmentExpression:
		return i.evalAssignmentExpression(node)
	case *parser.ThisExpression:
		return i.evalThisExpression()
	case *parser.StaticCallExpression:
		return i.evalStaticCallExpression(node)
	}

	return newError("未知节点类型: %T", node)
}

// evalProgram 执行程序，遍历所有语句并执行，最后调用 main 函数
func (i *Interpreter) evalProgram(program *parser.Program) Object {
	var result Object

	// 首先执行所有语句（包括函数定义）
	for _, statement := range program.Statements {
		result = i.Eval(statement)

		// 如果遇到返回语句或错误，立即返回
		switch result := result.(type) {
		case *ReturnValue:
			return result.Value
		case *Error:
			return result
		}
	}

	// 查找并调用 main 函数
	mainFn, ok := i.env.Get("main")
	if !ok {
		return newError("未找到 main 函数")
	}

	// 调用 main 函数
	switch fn := mainFn.(type) {
	case *Function:
		body, ok := fn.Body.(*parser.BlockStatement)
		if !ok {
			return newError("main 函数体类型错误")
		}
		// main 函数使用当前环境
		evaluated := i.evalBlockStatement(body)
		return unwrapReturnValue(evaluated)
	default:
		return newError("main 不是函数")
	}
}

// evalBlockStatement 执行块语句
func (i *Interpreter) evalBlockStatement(block *parser.BlockStatement) Object {
	var result Object

	for _, statement := range block.Statements {
		result = i.Eval(statement)

		if result != nil {
			rt := result.Type()
			if rt == RETURN_VALUE_OBJ || rt == ERROR_OBJ || rt == BREAK_SIGNAL_OBJ || rt == CONTINUE_SIGNAL_OBJ {
				return result
			}
		}
	}

	return result
}

// evalIdentifier 执行标识符
func (i *Interpreter) evalIdentifier(node *parser.Identifier) Object {
	// 首先尝试直接查找
	val, ok := i.env.Get(node.Value)
	if ok {
		return val
	}

	// 如果标识符包含点号，尝试作为成员访问处理
	parts := splitIdentifier(node.Value)
	if len(parts) >= 2 {
		// 获取第一部分（对象）
		obj, ok := i.env.Get(parts[0])
		if !ok {
			return newError("未定义的标识符: " + parts[0])
		}

		// 逐层访问成员
		for idx := 1; idx < len(parts); idx++ {
			memberName := parts[idx]
			switch object := obj.(type) {
			case *Instance:
				if val, ok := object.Fields[memberName]; ok {
					obj = val
				} else if method, ok := object.Class.Methods[memberName]; ok {
					return &BoundMethod{Instance: object, Method: method}
				} else {
					return newError("实例没有成员: %s", memberName)
				}
			case *Class:
				if method, ok := object.StaticMethods[memberName]; ok {
					return &Function{
						Parameters: method.Parameters,
						Body:       method.Body,
						Env:        method.Env,
						ReturnType: method.ReturnType,
					}
				}
				return newError("类 %s 没有静态成员: %s", object.Name, memberName)
			case *BuiltinObject:
				if member, ok := object.GetField(memberName); ok {
					obj = member
				} else {
					return newError("命名空间 %s 中没有成员 %s", parts[0], memberName)
				}
			default:
				return newError("无法访问 %s 的成员 %s", obj.Type(), memberName)
			}
		}
		return obj
	}

	return newError("未定义的标识符: " + node.Value)
}

// evalPrefixExpression 执行前缀表达式
func (i *Interpreter) evalPrefixExpression(operator string, right Object) Object {
	switch operator {
	case "!":
		return i.evalBangOperatorExpression(right)
	case "-":
		return i.evalMinusPrefixOperatorExpression(right)
	default:
		return newError("未知运算符: %s%s", operator, right.Type())
	}
}

// evalInfixExpression 执行中缀表达式
func (i *Interpreter) evalInfixExpression(operator string, left, right Object) Object {
	switch {
	case left.Type() == INTEGER_OBJ && right.Type() == INTEGER_OBJ:
		return i.evalIntegerInfixExpression(operator, left, right)
	case left.Type() == STRING_OBJ && right.Type() == STRING_OBJ:
		return i.evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return &Boolean{Value: left == right}
	case operator == "!=":
		return &Boolean{Value: left != right}
	case left.Type() == FLOAT_OBJ && right.Type() == FLOAT_OBJ:
		return i.evalFloatInfixExpression(operator, left, right)
	case left.Type() == FLOAT_OBJ && right.Type() == INTEGER_OBJ:
		// 浮点数 + 整数 -> 浮点数
		rightFloat := &Float{Value: float64(right.(*Integer).Value)}
		return i.evalFloatInfixExpression(operator, left, rightFloat)
	case left.Type() == INTEGER_OBJ && right.Type() == FLOAT_OBJ:
		// 整数 + 浮点数 -> 浮点数
		leftFloat := &Float{Value: float64(left.(*Integer).Value)}
		return i.evalFloatInfixExpression(operator, leftFloat, right)
	case left.Type() == BOOLEAN_OBJ && right.Type() == BOOLEAN_OBJ:
		return i.evalBooleanInfixExpression(operator, left, right)
	default:
		return newError("类型不匹配: %s %s %s", left.Type(), operator, right.Type())
	}
}

// evalIntegerInfixExpression 执行整数中缀表达式
func (i *Interpreter) evalIntegerInfixExpression(operator string, left, right Object) Object {
	leftVal := left.(*Integer).Value
	rightVal := right.(*Integer).Value

	switch operator {
	case "+":
		return &Integer{Value: leftVal + rightVal}
	case "-":
		return &Integer{Value: leftVal - rightVal}
	case "*":
		return &Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("除以零")
		}
		return &Integer{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return newError("模零")
		}
		return &Integer{Value: leftVal % rightVal}
	case "<":
		return &Boolean{Value: leftVal < rightVal}
	case ">":
		return &Boolean{Value: leftVal > rightVal}
	case "<=":
		return &Boolean{Value: leftVal <= rightVal}
	case ">=":
		return &Boolean{Value: leftVal >= rightVal}
	case "==":
		return &Boolean{Value: leftVal == rightVal}
	case "!=":
		return &Boolean{Value: leftVal != rightVal}
	default:
		return newError("未知运算符: %s %s %s", left.Type(), operator, right.Type())
	}
}

// evalFloatInfixExpression 执行浮点数中缀表达式
func (i *Interpreter) evalFloatInfixExpression(operator string, left, right Object) Object {
	leftVal := left.(*Float).Value
	rightVal := right.(*Float).Value

	switch operator {
	case "+":
		return &Float{Value: leftVal + rightVal}
	case "-":
		return &Float{Value: leftVal - rightVal}
	case "*":
		return &Float{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("除以零")
		}
		return &Float{Value: leftVal / rightVal}
	case "<":
		return &Boolean{Value: leftVal < rightVal}
	case ">":
		return &Boolean{Value: leftVal > rightVal}
	case "<=":
		return &Boolean{Value: leftVal <= rightVal}
	case ">=":
		return &Boolean{Value: leftVal >= rightVal}
	case "==":
		return &Boolean{Value: leftVal == rightVal}
	case "!=":
		return &Boolean{Value: leftVal != rightVal}
	default:
		return newError("未知运算符: %s %s %s", left.Type(), operator, right.Type())
	}
}

// evalStringInfixExpression 执行字符串中缀表达式
func (i *Interpreter) evalStringInfixExpression(operator string, left, right Object) Object {
	leftVal := left.(*String).Value
	rightVal := right.(*String).Value

	switch operator {
	case "+":
		return &String{Value: leftVal + rightVal}
	case "==":
		return &Boolean{Value: leftVal == rightVal}
	case "!=":
		return &Boolean{Value: leftVal != rightVal}
	default:
		return newError("未知运算符: STRING %s STRING", operator)
	}
}

// evalBooleanInfixExpression 执行布尔中缀表达式
func (i *Interpreter) evalBooleanInfixExpression(operator string, left, right Object) Object {
	leftVal := left.(*Boolean).Value
	rightVal := right.(*Boolean).Value

	switch operator {
	case "&&":
		return &Boolean{Value: leftVal && rightVal}
	case "||":
		return &Boolean{Value: leftVal || rightVal}
	case "==":
		return &Boolean{Value: leftVal == rightVal}
	case "!=":
		return &Boolean{Value: leftVal != rightVal}
	default:
		return newError("未知运算符: BOOLEAN %s BOOLEAN", operator)
	}
}

// evalBangOperatorExpression 执行 ! 运算符
func (i *Interpreter) evalBangOperatorExpression(right Object) Object {
	switch right {
	case &Boolean{Value: true}:
		return &Boolean{Value: false}
	case &Boolean{Value: false}:
		return &Boolean{Value: true}
	case &Null{}:
		return &Boolean{Value: true}
	default:
		return &Boolean{Value: false}
	}
}

// evalMinusPrefixOperatorExpression 执行 - 前缀运算符
func (i *Interpreter) evalMinusPrefixOperatorExpression(right Object) Object {
	if right.Type() != INTEGER_OBJ {
		return newError("未知运算符: -%s", right.Type())
	}

	value := right.(*Integer).Value
	return &Integer{Value: -value}
}

// evalTernaryExpression 执行三目运算符
func (i *Interpreter) evalTernaryExpression(node *parser.TernaryExpression) Object {
	condition := i.Eval(node.Condition)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return i.Eval(node.TrueExpr)
	}
	return i.Eval(node.FalseExpr)
}

// evalIfStatement 执行 if 语句
func (i *Interpreter) evalIfStatement(ie *parser.IfStatement) Object {
	condition := i.Eval(ie.Condition)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return i.Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return i.Eval(ie.Alternative)
	} else if ie.ElseIf != nil {
		return i.Eval(ie.ElseIf)
	} else {
		return &Null{}
	}
}

// evalFunctionLiteral 执行函数字面量，创建函数对象并捕获当前环境
func (i *Interpreter) evalFunctionLiteral(node *parser.FunctionLiteral) Object {
	params := []interface{}{}
	for _, p := range node.Parameters {
		params = append(params, p)
	}

	// 创建函数对象，保存参数、函数体和当前环境（闭包）
	return &Function{
		Parameters: params,
		Body:       node.Body,
		Env:        i.env, // 捕获当前环境，实现闭包
		ReturnType: []string{},
	}
}

// evalExpressions 执行表达式列表（函数调用参数）
func (i *Interpreter) evalExpressions(args []parser.CallArgument) []Object {
	result := []Object{}

	for _, arg := range args {
		evaluated := i.Eval(arg.Value)
		if isError(evaluated) {
			return []Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

// applyFunction 应用函数
func (i *Interpreter) applyFunction(fn Object, args []Object, callArgs []parser.CallArgument) Object {
	switch fn := fn.(type) {
	case *Function:
		// 需要将 interface{} 转换为正确的类型
		body, ok := fn.Body.(*parser.BlockStatement)
		if !ok {
			return newError("函数体类型错误")
		}
		extendedEnv := i.extendFunctionEnv(fn, args, callArgs)
		evaluated := i.evalBlockStatementWithEnv(body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *Builtin:
		return fn.Fn(args...)
	case *BuiltinObject:
		// 处理命名空间访问（如 fmt.Println）
		return newError("不能直接调用命名空间对象")
	case *BoundMethod:
		// 处理绑定方法调用
		return i.applyBoundMethod(fn, args, callArgs)
	default:
		return newError("不是函数: %s", fn.Type())
	}
}

// extendFunctionEnv 扩展函数环境
func (i *Interpreter) extendFunctionEnv(fn *Function, args []Object, callArgs []parser.CallArgument) *Environment {
	env := NewEnclosedEnvironment(fn.Env)

	// 处理命名参数和位置参数
	paramMap := make(map[string]Object)
	for idx, arg := range callArgs {
		if arg.Name != nil {
			// 命名参数
			paramMap[arg.Name.Value] = args[idx]
		}
	}

	// 设置参数
	for idx, paramInterface := range fn.Parameters {
		param, ok := paramInterface.(*parser.FunctionParameter)
		if !ok {
			continue
		}
		var val Object
		if idx < len(callArgs) && callArgs[idx].Name != nil {
			// 使用命名参数
			val = paramMap[callArgs[idx].Name.Value]
		} else if idx < len(args) {
			// 使用位置参数
			val = args[idx]
		} else if param.DefaultValue != nil {
			// 使用默认值
			val = i.Eval(param.DefaultValue)
		} else {
			val = &Null{}
		}
		env.Set(param.Name.Value, val)
	}

	return env
}

// evalBlockStatementWithEnv 在指定环境中执行块语句
func (i *Interpreter) evalBlockStatementWithEnv(block *parser.BlockStatement, env *Environment) Object {
	previousEnv := i.env
	i.env = env
	defer func() { i.env = previousEnv }()

	return i.evalBlockStatement(block)
}

// ========== 辅助函数 ==========

// isTruthy 判断值是否为真（用于 if 语句和三目运算符）
func isTruthy(obj Object) bool {
	if obj == nil {
		return false
	}
	switch o := obj.(type) {
	case *Null:
		return false
	case *Boolean:
		return o.Value
	default:
		return true // 非 null 和非 false 的值都视为真
	}
}

// isError 判断对象是否是错误对象
func isError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}

// newError 创建错误对象
func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

// unwrapReturnValue 解包返回值对象，提取实际的值
func unwrapReturnValue(obj Object) Object {
	if returnValue, ok := obj.(*ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

// splitIdentifier 分割标识符（用于处理成员访问，如 fmt.Println -> ["fmt", "Println"]）
func splitIdentifier(ident string) []string {
	// 简单实现：按 "." 分割
	parts := []string{}
	current := ""
	for _, ch := range ident {
		if ch == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// evalClassStatement 执行类定义语句
func (i *Interpreter) evalClassStatement(node *parser.ClassStatement) Object {
	class := &Class{
		Name:          node.Name.Value,
		Variables:     make(map[string]*ClassVariable),
		Methods:       make(map[string]*ClassMethod),
		StaticMethods: make(map[string]*ClassMethod),
		Env:           i.env,
	}

	// 遍历类成员，分别处理变量和方法
	for _, member := range node.Members {
		switch m := member.(type) {
		case *parser.ClassVariable:
			// 处理成员变量
			var defaultValue Object
			if m.Value != nil {
				defaultValue = i.Eval(m.Value)
			}
			class.Variables[m.Name.Value] = &ClassVariable{
				Name:           m.Name.Value,
				Type:           m.Type.Value,
				AccessModifier: m.AccessModifier,
				DefaultValue:   defaultValue,
			}
		case *parser.ClassMethod:
			// 处理方法
			returnTypes := []string{}
			for _, rt := range m.ReturnType {
				returnTypes = append(returnTypes, rt.Value)
			}
			method := &ClassMethod{
				Name:           m.Name.Value,
				AccessModifier: m.AccessModifier,
				IsStatic:       m.IsStatic,
				Parameters:     toInterfaceSlice(m.Parameters),
				ReturnType:     returnTypes,
				Body:           m.Body,
				Env:            i.env,
			}
			if m.IsStatic {
				class.StaticMethods[m.Name.Value] = method
			} else {
				class.Methods[m.Name.Value] = method
			}
		}
	}

	i.env.Set(node.Name.Value, class)
	return class
}

// toInterfaceSlice 将 []*parser.FunctionParameter 转换为 []interface{}
func toInterfaceSlice(params []*parser.FunctionParameter) []interface{} {
	result := make([]interface{}, len(params))
	for i, p := range params {
		result[i] = p
	}
	return result
}

// evalNewExpression 执行 new 表达式，创建类实例
func (i *Interpreter) evalNewExpression(node *parser.NewExpression) Object {
	// 获取类定义
	className := node.ClassName.Value
	classObj, ok := i.env.Get(className)
	if !ok {
		return newError("未定义的类: %s", className)
	}

	class, ok := classObj.(*Class)
	if !ok {
		return newError("%s 不是一个类", className)
	}

	// 创建实例
	instance := &Instance{
		Class:  class,
		Fields: make(map[string]Object),
	}

	// 初始化成员变量（使用默认值）
	for name, variable := range class.Variables {
		if variable.DefaultValue != nil {
			instance.Fields[name] = variable.DefaultValue
		} else {
			instance.Fields[name] = &Null{}
		}
	}

	// 调用构造函数（如果存在）
	if constructor, ok := class.Methods["__construct"]; ok {
		// 创建新的环境，绑定 this
		constructorEnv := NewEnclosedEnvironment(class.Env)
		constructorEnv.Set("this", instance)
		// 在类方法中提供 self（指向当前类），便于 self::xxx 形式的调用
		constructorEnv.Set("self", class)

		// 绑定参数
		args := i.evalExpressions(node.Arguments)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		params := constructor.Parameters
		for idx, param := range params {
			if p, ok := param.(*parser.FunctionParameter); ok {
				if idx < len(args) {
					constructorEnv.Set(p.Name.Value, args[idx])
				} else if p.DefaultValue != nil {
					// 使用默认值
					defaultVal := i.Eval(p.DefaultValue)
					constructorEnv.Set(p.Name.Value, defaultVal)
				}
			}
		}

		// 执行构造函数体
		oldEnv := i.env
		i.env = constructorEnv
		if body, ok := constructor.Body.(*parser.BlockStatement); ok {
			i.evalBlockStatement(body)
		}
		i.env = oldEnv
	}

	return instance
}

// evalMemberAccessExpression 执行成员访问表达式
func (i *Interpreter) evalMemberAccessExpression(node *parser.MemberAccessExpression) Object {
	obj := i.Eval(node.Object)
	if isError(obj) {
		return obj
	}

	memberName := node.Member.Value

	switch object := obj.(type) {
	case *Instance:
		// 访问实例成员
		if val, ok := object.Fields[memberName]; ok {
			return val
		}
		// 检查是否是方法
		if method, ok := object.Class.Methods[memberName]; ok {
			// 返回绑定了 this 的方法
			return &BoundMethod{
				Instance: object,
				Method:   method,
			}
		}
		return newError("实例没有成员: %s", memberName)
	case *Class:
		// 访问静态成员
		if method, ok := object.StaticMethods[memberName]; ok {
			return &Function{
				Parameters: method.Parameters,
				Body:       method.Body,
				Env:        method.Env,
				ReturnType: method.ReturnType,
			}
		}
		return newError("类 %s 没有静态成员: %s", object.Name, memberName)
	case *BuiltinObject:
		// 访问内置对象成员（如 fmt.Println）
		if member, ok := object.GetField(memberName); ok {
			return member
		}
		return newError("命名空间中没有成员: %s", memberName)
	default:
		return newError("无法访问 %s 的成员", obj.Type())
	}
}

// evalAssignmentExpression 执行赋值表达式
func (i *Interpreter) evalAssignmentExpression(node *parser.AssignmentExpression) Object {
	val := i.Eval(node.Right)
	if isError(val) {
		return val
	}

	switch left := node.Left.(type) {
	case *parser.Identifier:
		// 检查标识符是否包含点号（如 this.name）
		parts := splitIdentifier(left.Value)
		if len(parts) >= 2 {
			// 这是成员访问赋值
			obj, ok := i.env.Get(parts[0])
			if !ok {
				return newError("未定义的标识符: %s", parts[0])
			}
			// 逐层访问到倒数第二层
			for idx := 1; idx < len(parts)-1; idx++ {
				if instance, ok := obj.(*Instance); ok {
					if field, ok := instance.Fields[parts[idx]]; ok {
						obj = field
					} else {
						return newError("实例没有成员: %s", parts[idx])
					}
				} else {
					return newError("无法访问 %s 的成员", obj.Type())
				}
			}
			// 设置最后一个成员
			lastMember := parts[len(parts)-1]
			if instance, ok := obj.(*Instance); ok {
				instance.Fields[lastMember] = val
				return val
			}
			return newError("无法给 %s 的成员赋值", obj.Type())
		}
		// 普通标识符赋值
		i.env.Set(left.Value, val)
		return val
	case *parser.MemberAccessExpression:
		obj := i.Eval(left.Object)
		if isError(obj) {
			return obj
		}
		if instance, ok := obj.(*Instance); ok {
			instance.Fields[left.Member.Value] = val
			return val
		}
		return newError("无法给 %s 的成员赋值", obj.Type())
	default:
		return newError("无效的赋值目标")
	}
}

// evalThisExpression 执行 this 表达式
func (i *Interpreter) evalThisExpression() Object {
	if this, ok := i.env.Get("this"); ok {
		return this
	}
	return newError("this 只能在类方法中使用")
}

// evalStaticCallExpression 执行静态方法调用
func (i *Interpreter) evalStaticCallExpression(node *parser.StaticCallExpression) Object {
	// 获取类定义
	className := node.ClassName.Value
	classObj, ok := i.env.Get(className)
	if !ok {
		return newError("未定义的类: %s", className)
	}

	class, ok := classObj.(*Class)
	if !ok {
		return newError("%s 不是一个类", className)
	}

	// 获取静态方法
	methodName := node.Method.Value
	method, ok := class.StaticMethods[methodName]
	if !ok {
		return newError("类 %s 没有静态方法: %s", className, methodName)
	}

	// 求值参数
	args := i.evalExpressions(node.Arguments)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	// 创建函数环境
	env := NewEnclosedEnvironment(method.Env)
	// 在静态方法中提供 self（指向当前类），支持 self::xxx
	env.Set("self", class)

	// 绑定参数
	for idx, paramInterface := range method.Parameters {
		param, ok := paramInterface.(*parser.FunctionParameter)
		if !ok {
			continue
		}
		var val Object
		if idx < len(args) {
			val = args[idx]
		} else if param.DefaultValue != nil {
			val = i.Eval(param.DefaultValue)
		} else {
			val = &Null{}
		}
		env.Set(param.Name.Value, val)
	}

	// 执行方法体
	body, ok := method.Body.(*parser.BlockStatement)
	if !ok {
		return newError("方法体类型错误")
	}

	evaluated := i.evalBlockStatementWithEnv(body, env)
	return unwrapReturnValue(evaluated)
}

// applyBoundMethod 执行绑定方法
func (i *Interpreter) applyBoundMethod(bm *BoundMethod, args []Object, callArgs []parser.CallArgument) Object {
	method := bm.Method
	body, ok := method.Body.(*parser.BlockStatement)
	if !ok {
		return newError("方法体类型错误")
	}

	// 创建新环境，绑定 this
	env := NewEnclosedEnvironment(method.Env)
	env.Set("this", bm.Instance)
	// 在实例方法中也提供 self（指向该实例所属类），便于 self::xxx
	env.Set("self", bm.Instance.Class)

	// 绑定参数
	for idx, paramInterface := range method.Parameters {
		param, ok := paramInterface.(*parser.FunctionParameter)
		if !ok {
			continue
		}
		var val Object
		if idx < len(args) {
			val = args[idx]
		} else if param.DefaultValue != nil {
			val = i.Eval(param.DefaultValue)
		} else {
			val = &Null{}
		}
		env.Set(param.Name.Value, val)
	}

	// 执行方法体
	evaluated := i.evalBlockStatementWithEnv(body, env)
	return unwrapReturnValue(evaluated)
}

// evalForStatement 执行 for 循环语句
func (i *Interpreter) evalForStatement(node *parser.ForStatement) Object {
	// 执行初始化语句
	if node.Init != nil {
		i.Eval(node.Init)
	}

	for {
		// 检查条件
		if node.Condition != nil {
			condition := i.Eval(node.Condition)
			if isError(condition) {
				return condition
			}
			if !isTruthy(condition) {
				break
			}
		}

		// 执行循环体
		result := i.Eval(node.Body)

		// 检查是否有控制流信号
		if result != nil {
			switch result.Type() {
			case BREAK_SIGNAL_OBJ:
				return nil // break 跳出循环
			case CONTINUE_SIGNAL_OBJ:
				// continue 跳过本次迭代，执行 post 语句后继续
			case RETURN_VALUE_OBJ:
				return result // return 直接返回
			case ERROR_OBJ:
				return result
			}
		}

		// 执行 post 语句（如 i++）
		if node.Post != nil {
			i.Eval(node.Post)
		}
	}

	return nil
}

// evalIncrementStatement 执行自增/自减语句
func (i *Interpreter) evalIncrementStatement(node *parser.IncrementStatement) Object {
	val, ok := i.env.Get(node.Name.Value)
	if !ok {
		return newError("未定义的变量: %s", node.Name.Value)
	}

	intVal, ok := val.(*Integer)
	if !ok {
		return newError("自增/自减只能用于整数类型")
	}

	var newVal int64
	if node.Operator == "++" {
		newVal = intVal.Value + 1
	} else {
		newVal = intVal.Value - 1
	}

	i.env.Set(node.Name.Value, &Integer{Value: newVal})
	return &Integer{Value: newVal}
}

