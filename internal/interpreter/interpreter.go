package interpreter

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/tangzhangming/longlang/internal/config"
	"github.com/tangzhangming/longlang/internal/lexer"
	"github.com/tangzhangming/longlang/internal/parser"
)

// readFile 读取文件内容
func readFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// newLexer 创建词法分析器
func newLexer(input string) *lexer.Lexer {
	return lexer.New(input)
}

// newParser 创建语法分析器
func newParser(l *lexer.Lexer) *parser.Parser {
	return parser.New(l)
}

// Interpreter 解释器，负责执行 AST 节点
type Interpreter struct {
	env              *Environment          // 当前作用域环境
	stdlibPath       string                // 标准库目录路径
	loadedModules    map[string]*Module    // 已加载的模块缓存（已废弃，保留以兼容）
	currentFileName  string                // 当前正在处理的文件名（不含扩展名），用于判断导出类
	namespaceMgr     *NamespaceManager     // 命名空间管理器
	currentNamespace *Namespace            // 当前命名空间
	projectRoot      string                // 项目根目录
	projectConfig    *config.ProjectConfig // 项目配置
	loadedNamespaces map[string]bool       // 已加载的命名空间文件缓存
}

// Module 模块对象
type Module struct {
	Name    string            // 模块名
	Env     *Environment      // 模块环境
	Exports map[string]Object // 导出的符号
}

// New 创建新解释器并初始化内置函数
func New() *Interpreter {
	env := NewEnvironment()
	// 注册内置函数（如 fmt.Println）
	registerBuiltins(env)
	return &Interpreter{
		env:              env,
		stdlibPath:       "stdlib", // 默认标准库路径
		loadedModules:    make(map[string]*Module),
		namespaceMgr:     NewNamespaceManager(),
		projectRoot:      "",
		projectConfig:    nil,
		loadedNamespaces: make(map[string]bool),
	}
}

// SetStdlibPath 设置标准库目录路径
func (i *Interpreter) SetStdlibPath(path string) {
	i.stdlibPath = path
}

// SetProjectConfig 设置项目配置
func (i *Interpreter) SetProjectConfig(projectRoot string, cfg *config.ProjectConfig) {
	i.projectRoot = projectRoot
	i.projectConfig = cfg
}

// GetEnv 获取当前环境
func (i *Interpreter) GetEnv() *Environment {
	return i.env
}

// Eval 执行 AST 节点，根据节点类型分发到相应的处理函数
func (i *Interpreter) Eval(node parser.Node) Object {
	switch node := node.(type) {
	case *parser.Program:
		return i.evalProgram(node)
	case *parser.NamespaceStatement:
		// 处理命名空间声明
		return i.evalNamespaceStatement(node)
	case *parser.UseStatement:
		// 处理 use 导入语句
		return i.evalUseStatement(node)
	case *parser.ClassStatement:
		// 类定义，注册到环境中
		return i.evalClassStatement(node)
	case *parser.InterfaceStatement:
		// 接口定义，注册到环境中
		return i.evalInterfaceStatement(node)
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
	case *parser.SuperExpression:
		return i.evalSuperExpression()
	case *parser.StaticCallExpression:
		return i.evalStaticCallExpression(node)
	case *parser.StaticAccessExpression:
		return i.evalStaticAccessExpression(node)
	}

	return newError("未知节点类型: %T", node)
}

// evalProgram 执行程序，遍历所有语句并执行，最后调用类的 main 静态方法
func (i *Interpreter) evalProgram(program *parser.Program) Object {
	var result Object

	// 首先执行所有语句（包括类定义）
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

	// 查找包含 main 静态方法的类
	// 先收集所有包含main方法的类
	type mainClassInfo struct {
		class  *Class
		method *ClassMethod
		name   string
	}
	var mainClasses []mainClassInfo

	// 从所有命名空间中查找并收集所有包含main方法的类
	for _, ns := range i.namespaceMgr.namespaces {
		for className, class := range ns.Classes {
			if method, ok := class.StaticMethods["main"]; ok {
				mainClasses = append(mainClasses, mainClassInfo{
					class:  class,
					method: method,
					name:   className,
				})
			}
		}
	}

	// 如果找到多个 main 方法，报错
	if len(mainClasses) > 1 {
		classList := ""
		for idx, info := range mainClasses {
			if idx > 0 {
				classList += ", "
			}
			classList += info.name
		}
		return newError("找到多个包含 main 方法的类: %s，必须指定启动入口类", classList)
	}

	// 如果只找到一个，使用它
	var mainClass *Class
	var mainMethod *ClassMethod
	if len(mainClasses) == 1 {
		mainClass = mainClasses[0].class
		mainMethod = mainClasses[0].method
	}

	// 注意：暂时只从命名空间查找，如果需要支持没有命名空间的类，
	// 需要扩展Environment提供遍历所有对象的方法

	if mainClass == nil || mainMethod == nil {
		return newError("未找到包含 main 静态方法的类")
	}

	// 调用 main 静态方法
	// 创建函数环境
	env := NewEnclosedEnvironment(mainMethod.Env)
	// 在静态方法中提供 self（指向当前类）
	env.Set("self", mainClass)

	// 执行方法体
	body, ok := mainMethod.Body.(*parser.BlockStatement)
	if !ok {
		return newError("main 方法体类型错误")
	}

	// 保存当前环境并切换
	oldEnv := i.env
	i.env = env
	defer func() { i.env = oldEnv }()

	evaluated := i.evalBlockStatement(body)
	return unwrapReturnValue(evaluated)
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
	// 字符串 + 其他类型 -> 字符串拼接（自动转换）
	case left.Type() == STRING_OBJ && operator == "+":
		return i.evalStringConcatExpression(left, right)
	case right.Type() == STRING_OBJ && operator == "+":
		return i.evalStringConcatExpression(left, right)
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

// evalStringConcatExpression 字符串拼接（自动将其他类型转换为字符串）
func (i *Interpreter) evalStringConcatExpression(left, right Object) Object {
	leftStr := objectToString(left)
	rightStr := objectToString(right)
	return &String{Value: leftStr + rightStr}
}

// objectToString 将对象转换为字符串
func objectToString(obj Object) string {
	switch o := obj.(type) {
	case *String:
		return o.Value
	case *Integer:
		return fmt.Sprintf("%d", o.Value)
	case *Float:
		return fmt.Sprintf("%g", o.Value)
	case *Boolean:
		if o.Value {
			return "true"
		}
		return "false"
	case *Null:
		return "null"
	default:
		return obj.Inspect()
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
	switch right.Type() {
	case INTEGER_OBJ:
		value := right.(*Integer).Value
		return &Integer{Value: -value}
	case FLOAT_OBJ:
		value := right.(*Float).Value
		return &Float{Value: -value}
	default:
		return newError("未知运算符: -%s", right.Type())
	}
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

	// 提取返回类型
	returnTypes := []string{}
	for _, rt := range node.ReturnType {
		returnTypes = append(returnTypes, rt.Value)
	}

	// 创建函数对象，保存参数、函数体和当前环境（闭包）
	return &Function{
		Parameters: params,
		Body:       node.Body,
		Env:        i.env, // 捕获当前环境，实现闭包
		ReturnType: returnTypes,
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
		result := unwrapReturnValue(evaluated)

		// 检查返回类型
		if len(fn.ReturnType) == 0 {
			// 函数没有声明返回类型，不应该返回非 null 值
			if result != nil && result.Type() != NULL_OBJ {
				return newError("函数未声明返回类型，但返回了值")
			}
		}

		return result
	case *Builtin:
		return fn.Fn(args...)
	case *BuiltinObject:
		// 处理命名空间访问（如 fmt.Println）
		return newError("不能直接调用命名空间对象")
	case *BoundMethod:
		// 处理绑定方法调用
		return i.applyBoundMethod(fn, args, callArgs)
	case *BoundStringMethod:
		// 处理字符串方法调用
		return fn.Method(fn.String, args...)
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

// loadModule 加载模块文件
func (i *Interpreter) loadModule(name string) (*Module, error) {
	// 构建文件路径
	filePath := i.stdlibPath + "/" + name + ".long"

	// 读取文件
	content, err := readFile(filePath)
	if err != nil {
		return nil, err
	}

	// 创建模块环境（继承全局环境中的内置函数）
	moduleEnv := NewEnclosedEnvironment(i.env)

	// 词法分析
	l := newLexer(content)

	// 语法分析
	p := newParser(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("模块 %s 语法错误: %v", name, p.Errors())
	}

	// 提取文件名（不含扩展名）用于判断导出类
	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	// 创建模块解释器，设置当前文件名
	moduleInterpreter := &Interpreter{
		env:             moduleEnv,
		stdlibPath:      i.stdlibPath,
		loadedModules:   i.loadedModules,
		currentFileName: fileName,
	}

	// 执行模块代码（不调用 main）
	for _, stmt := range program.Statements {
		result := moduleInterpreter.Eval(stmt)
		if isError(result) {
			return nil, fmt.Errorf("模块 %s 执行错误: %s", name, result.Inspect())
		}
	}

	// 收集导出的符号（只导出 IsExported=true 的类，其他符号全部导出）
	exports := make(map[string]Object)
	for key, val := range moduleEnv.store {
		// 如果是类，检查是否为导出类
		if class, ok := val.(*Class); ok {
			if class.IsExported {
				exports[key] = val
			}
		} else {
			// 函数、变量等其他符号全部导出
			exports[key] = val
		}
	}

	return &Module{
		Name:    name,
		Env:     moduleEnv,
		Exports: exports,
	}, nil
}

// registerModule 将模块注册到当前环境
func (i *Interpreter) registerModule(name string, module *Module) {
	// 创建模块命名空间对象
	moduleObj := NewBuiltinObject(name)

	// 将导出的符号添加到命名空间
	for key, val := range module.Exports {
		moduleObj.SetField(key, val)
	}

	// 注册到当前环境
	i.env.Set(name, moduleObj)
}

// evalNamespaceStatement 执行命名空间声明语句
func (i *Interpreter) evalNamespaceStatement(node *parser.NamespaceStatement) Object {
	namespaceName := node.Name.Value

	// 如果有项目配置，使用 root_namespace 解析命名空间
	if i.projectConfig != nil {
		namespaceName = i.projectConfig.ResolveNamespace(namespaceName)
	}

	// 获取或创建命名空间
	namespace := i.namespaceMgr.GetNamespace(namespaceName)
	i.currentNamespace = namespace

	// 命名空间声明不返回值，只是设置当前命名空间
	return nil
}

// evalUseStatement 执行 use 导入语句
func (i *Interpreter) evalUseStatement(node *parser.UseStatement) Object {
	fullPath := node.Path.Value

	// 解析完全限定名：Illuminate.Database.Eloquent.Model
	// 分解为：命名空间 Illuminate.Database.Eloquent，类名 Model
	namespace, className, err := ResolveClassName(fullPath)
	if err != nil {
		return newError("无效的 use 路径: %s", fullPath)
	}

	// 尝试加载命名空间文件（如果尚未加载）
	_ = i.loadNamespaceFile(namespace, className)

	// 首先尝试在原始命名空间中查找（支持标准库）
	targetNamespace := i.namespaceMgr.GetNamespace(namespace)
	class, ok := targetNamespace.GetClass(className)

	// 如果找不到，且有项目配置，尝试使用 root_namespace 解析后的命名空间
	if !ok && i.projectConfig != nil {
		resolvedNamespace := i.projectConfig.ResolveNamespace(namespace)
		if resolvedNamespace != namespace {
			targetNamespace = i.namespaceMgr.GetNamespace(resolvedNamespace)
			class, ok = targetNamespace.GetClass(className)
		}
	}

	if !ok {
		return newError("命名空间 %s 中没有找到类 %s", namespace, className)
	}

	// 确定导入到当前环境的名称
	importName := className
	if node.Alias != nil {
		importName = node.Alias.Value
	}

	// 将类导入到当前环境
	i.env.Set(importName, class)

	return nil
}

// loadNamespaceFile 根据命名空间路径加载对应的文件
// 例如：Mycompany.Myapp.Models + User -> src/Mycompany/Myapp/Models/User.long
func (i *Interpreter) loadNamespaceFile(namespace string, className string) error {
	// 构建完整的命名空间路径作为缓存 key
	fullKey := namespace + "." + className

	// 检查是否已经加载过
	if i.loadedNamespaces[fullKey] {
		return nil
	}

	// 将命名空间转换为文件路径
	// 例如：Mycompany.Myapp.Models -> Mycompany/Myapp/Models
	namespacePath := strings.ReplaceAll(namespace, ".", string(filepath.Separator))

	// 构建可能的文件路径
	var filePaths []string

	// 如果有项目根目录，搜索项目目录
	if i.projectRoot != "" {
		// 如果有 root_namespace，计算相对路径
		// 例如：namespace = "Usoppsoft.Account.Models", root_namespace = "Usoppsoft.Account"
		// 则相对路径为 "Models"
		relativeNamespacePath := namespacePath
		if i.projectConfig != nil && i.projectConfig.RootNamespace != "" {
			rootNsPath := strings.ReplaceAll(i.projectConfig.RootNamespace, ".", string(filepath.Separator))
			if strings.HasPrefix(namespacePath, rootNsPath) {
				// 去掉 root_namespace 前缀
				relativeNamespacePath = strings.TrimPrefix(namespacePath, rootNsPath)
				relativeNamespacePath = strings.TrimPrefix(relativeNamespacePath, string(filepath.Separator))
			}
		}

		// 1. 使用相对路径在 src 目录下查找（优先）
		if relativeNamespacePath != "" {
			srcRelPath := filepath.Join(i.projectRoot, "src", relativeNamespacePath, className+".long")
			filePaths = append(filePaths, srcRelPath)
		} else {
			// 如果相对路径为空，直接在 src 下查找
			srcRelPath := filepath.Join(i.projectRoot, "src", className+".long")
			filePaths = append(filePaths, srcRelPath)
		}

		// 2. 使用完整路径在 src 目录下查找
		srcPath := filepath.Join(i.projectRoot, "src", namespacePath, className+".long")
		filePaths = append(filePaths, srcPath)

		// 3. 在项目根目录下查找（相对路径）
		if relativeNamespacePath != "" {
			rootRelPath := filepath.Join(i.projectRoot, relativeNamespacePath, className+".long")
			filePaths = append(filePaths, rootRelPath)
		}

		// 4. 在项目根目录下查找（完整路径）
		rootPath := filepath.Join(i.projectRoot, namespacePath, className+".long")
		filePaths = append(filePaths, rootPath)

		// 5. 在 vendor 目录下查找
		vendorPath := filepath.Join(i.projectRoot, "vendor", namespacePath, className+".long")
		filePaths = append(filePaths, vendorPath)
	}

	// 6. 在标准库目录下查找（无论是否有项目根目录）
	if i.stdlibPath != "" {
		stdlibPath := filepath.Join(i.stdlibPath, namespacePath, className+".long")
		filePaths = append(filePaths, stdlibPath)
	}

	// 尝试加载文件
	var loadedPath string
	var content string
	for _, path := range filePaths {
		if c, err := readFile(path); err == nil {
			content = c
			loadedPath = path
			break
		}
	}

	if loadedPath == "" {
		return fmt.Errorf("找不到文件: %s，尝试路径: %v", className+".long", filePaths)
	}

	// 标记为已加载
	i.loadedNamespaces[fullKey] = true

	// 解析并执行文件
	l := newLexer(content)
	p := newParser(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Errorf("解析文件 %s 错误: %s", loadedPath, p.Errors()[0])
	}

	// 保存当前状态，执行完后恢复
	savedNamespace := i.currentNamespace
	savedConfig := i.projectConfig

	// 如果是从标准库加载的文件，临时禁用 projectConfig
	// 这样 namespace 声明不会被 root_namespace 前缀化
	isStdlibFile := i.stdlibPath != "" && strings.HasPrefix(loadedPath, i.stdlibPath)
	if isStdlibFile {
		i.projectConfig = nil
	}

	// 执行文件中的语句（但不执行 main）
	for _, stmt := range program.Statements {
		result := i.Eval(stmt)
		if isError(result) {
			i.currentNamespace = savedNamespace
			i.projectConfig = savedConfig
			return fmt.Errorf("执行文件 %s 错误: %s", loadedPath, result.Inspect())
		}
	}

	// 恢复状态
	i.currentNamespace = savedNamespace
	i.projectConfig = savedConfig

	return nil
}

// evalInterfaceStatement 执行接口定义语句
func (i *Interpreter) evalInterfaceStatement(node *parser.InterfaceStatement) Object {
	iface := &Interface{
		Name:    node.Name.Value,
		Methods: make(map[string]*InterfaceMethod),
	}

	// 处理接口方法
	for _, m := range node.Methods {
		returnTypes := []string{}
		for _, rt := range m.ReturnType {
			returnTypes = append(returnTypes, rt.Value)
		}
		paramTypes := []string{}
		for _, p := range m.Parameters {
			if p.Type != nil {
				paramTypes = append(paramTypes, p.Type.Value)
			}
		}
		iface.Methods[m.Name.Value] = &InterfaceMethod{
			Name:       m.Name.Value,
			Parameters: paramTypes,
			ReturnType: returnTypes,
		}
	}

	i.env.Set(node.Name.Value, iface)
	return iface
}

// evalClassStatement 执行类定义语句
func (i *Interpreter) evalClassStatement(node *parser.ClassStatement) Object {
	// 判断是否为导出类：类名与文件名相同
	isExported := false
	if i.currentFileName != "" && node.Name.Value == i.currentFileName {
		isExported = true
	}

	class := &Class{
		Name:          node.Name.Value,
		Variables:     make(map[string]*ClassVariable),
		Constants:     make(map[string]*ClassConstant),
		Methods:       make(map[string]*ClassMethod),
		StaticMethods: make(map[string]*ClassMethod),
		Env:           i.env,
		IsExported:    isExported,
	}

	// 处理继承
	if node.Parent != nil {
		parentObj, ok := i.env.Get(node.Parent.Value)
		if !ok {
			return newError("未定义的父类: %s", node.Parent.Value)
		}
		parentClass, ok := parentObj.(*Class)
		if !ok {
			return newError("%s 不是一个类", node.Parent.Value)
		}
		class.Parent = parentClass
	}

	// 处理接口实现
	if len(node.Interfaces) > 0 {
		class.Interfaces = make([]*Interface, 0, len(node.Interfaces))
		for _, ifaceName := range node.Interfaces {
			ifaceObj, ok := i.env.Get(ifaceName.Value)
			if !ok {
				return newError("未定义的接口: %s", ifaceName.Value)
			}
			iface, ok := ifaceObj.(*Interface)
			if !ok {
				return newError("%s 不是一个接口", ifaceName.Value)
			}
			class.Interfaces = append(class.Interfaces, iface)
		}
	}

	// 遍历类成员，分别处理常量、变量和方法
	for _, member := range node.Members {
		switch m := member.(type) {
		case *parser.ClassConstant:
			// 处理常量
			constValue := i.Eval(m.Value)
			if isError(constValue) {
				return constValue
			}
			constType := ""
			if m.Type != nil {
				constType = m.Type.Value
			}
			class.Constants[m.Name.Value] = &ClassConstant{
				Name:           m.Name.Value,
				Type:           constType,
				AccessModifier: m.AccessModifier,
				Value:          constValue,
			}
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

	// 检查是否实现了所有接口方法
	for _, iface := range class.Interfaces {
		if !class.Implements(iface) {
			return newError("类 %s 没有完全实现接口 %s", class.Name, iface.Name)
		}
	}

	// 如果有当前命名空间，将类注册到命名空间，否则注册到环境（兼容旧代码）
	if i.currentNamespace != nil {
		i.currentNamespace.SetClass(node.Name.Value, class)
	} else {
		// 如果没有命名空间，也注册到环境（向后兼容）
		i.env.Set(node.Name.Value, class)
	}

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

	// 首先从当前环境查找（向后兼容）
	classObj, ok := i.env.Get(className)

	// 如果当前环境找不到，且存在当前命名空间，从当前命名空间查找
	if !ok && i.currentNamespace != nil {
		if class, found := i.currentNamespace.GetClass(className); found {
			classObj = class
			ok = true
		}
	}

	// 如果还是找不到，从所有命名空间中查找
	if !ok {
		for _, ns := range i.namespaceMgr.namespaces {
			if class, found := ns.GetClass(className); found {
				classObj = class
				ok = true
				break
			}
		}
	}

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

	// 初始化成员变量（包括从父类继承的变量）
	i.initInstanceFields(instance, class)

	// 调用构造函数（如果存在，包括继承的构造函数）
	if constructor, ok := class.GetMethod("__construct"); ok {
		// 创建新的环境，绑定 this
		constructorEnv := NewEnclosedEnvironment(class.Env)
		constructorEnv.Set("this", instance)
		// 在类方法中提供 self（指向当前类），便于 self::xxx 形式的调用
		constructorEnv.Set("self", class)
		// 提供 super（指向父类）
		if class.Parent != nil {
			constructorEnv.Set("super", class.Parent)
		}

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

// initInstanceFields 初始化实例的所有字段（包括继承的字段）
func (i *Interpreter) initInstanceFields(instance *Instance, class *Class) {
	// 先初始化父类字段
	if class.Parent != nil {
		i.initInstanceFields(instance, class.Parent)
	}

	// 初始化当前类的字段
	for name, variable := range class.Variables {
		if variable.DefaultValue != nil {
			instance.Fields[name] = variable.DefaultValue
		} else {
			instance.Fields[name] = &Null{}
		}
	}
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
		// 检查是否是方法（包括继承的方法）
		if method, ok := object.Class.GetMethod(memberName); ok {
			// 返回绑定了 this 的方法
			return &BoundMethod{
				Instance: object,
				Method:   method,
			}
		}
		return newError("实例没有成员: %s", memberName)
	case *Class:
		// 访问静态成员（包括继承的静态方法）
		if method, ok := object.GetStaticMethod(memberName); ok {
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
	case *String:
		// 访问字符串方法
		if method, ok := GetStringMethod(memberName); ok {
			// 返回绑定的字符串方法
			return &BoundStringMethod{
				String: object,
				Method: method,
				Name:   memberName,
			}
		}
		return newError("字符串没有方法: %s", memberName)
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

// evalSuperExpression 执行 super 表达式
func (i *Interpreter) evalSuperExpression() Object {
	if super, ok := i.env.Get("super"); ok {
		return super
	}
	return newError("super 只能在子类方法中使用")
}

// evalStaticCallExpression 执行静态方法调用
func (i *Interpreter) evalStaticCallExpression(node *parser.StaticCallExpression) Object {
	// 获取类定义
	className := node.ClassName.Value

	// 首先从当前环境查找（向后兼容）
	classObj, ok := i.env.Get(className)

	// 如果当前环境找不到，且存在当前命名空间，从当前命名空间查找
	if !ok && i.currentNamespace != nil {
		if class, found := i.currentNamespace.GetClass(className); found {
			classObj = class
			ok = true
		}
	}

	// 如果还是找不到，从所有命名空间中查找
	if !ok {
		for _, ns := range i.namespaceMgr.namespaces {
			if class, found := ns.GetClass(className); found {
				classObj = class
				ok = true
				break
			}
		}
	}

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

// evalStaticAccessExpression 执行静态访问（常量访问）
func (i *Interpreter) evalStaticAccessExpression(node *parser.StaticAccessExpression) Object {
	className := node.ClassName.Value
	constName := node.Name.Value

	var class *Class
	var ok bool

	// 处理 self:: 的情况
	if className == "self" {
		// 从环境中获取 self（应该在类方法内部）
		selfObj, found := i.env.Get("self")
		if !found {
			return newError("self 只能在类方法内部使用")
		}
		class, ok = selfObj.(*Class)
		if !ok {
			return newError("self 必须指向一个类")
		}
	} else {
		// 获取类定义
		classObj, found := i.env.Get(className)
		if !found && i.currentNamespace != nil {
			if c, f := i.currentNamespace.GetClass(className); f {
				classObj = c
				found = true
			}
		}
		if !found {
			for _, ns := range i.namespaceMgr.namespaces {
				if c, f := ns.GetClass(className); f {
					classObj = c
					found = true
					break
				}
			}
		}
		if !found {
			return newError("未定义的类: %s", className)
		}
		class, ok = classObj.(*Class)
		if !ok {
			return newError("%s 不是一个类", className)
		}
	}

	// 获取常量
	constant, ok := class.GetConstant(constName)
	if !ok {
		return newError("类 %s 没有常量: %s", class.Name, constName)
	}

	// 检查访问权限（暂时所有常量都可以访问，后续可以添加访问控制）
	return constant.Value
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
	// 提供 super（指向父类）
	if bm.Instance.Class.Parent != nil {
		env.Set("super", bm.Instance.Class.Parent)
	}

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
