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
		val := i.Eval(node.ReturnValue)
		if isError(val) {
			return val
		}
		return &ReturnValue{Value: val}
	case *parser.BlockStatement:
		return i.evalBlockStatement(node)
	case *parser.IfStatement:
		return i.evalIfStatement(node)
	case *parser.IntegerLiteral:
		return &Integer{Value: node.Value}
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
			if rt == RETURN_VALUE_OBJ || rt == ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

// evalIdentifier 执行标识符
func (i *Interpreter) evalIdentifier(node *parser.Identifier) Object {
	val, ok := i.env.Get(node.Value)
	if !ok {
		return newError("未定义的标识符: " + node.Value)
	}
	return val
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
	switch obj {
	case &Null{}:
		return false
	case &Boolean{Value: true}:
		return true
	case &Boolean{Value: false}:
		return false
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

