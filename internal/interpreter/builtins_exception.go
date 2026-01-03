package interpreter

import (
	"fmt"
	"strings"
)

// registerExceptionClasses 注册异常相关的类
func registerExceptionClasses(env *Environment) {
	// 创建 Exception 基类
	exceptionClass := createExceptionClass()
	env.Set("Exception", exceptionClass)

	// 创建 RuntimeException 类
	runtimeExceptionClass := createRuntimeExceptionClass(exceptionClass)
	env.Set("RuntimeException", runtimeExceptionClass)

	// 创建其他标准异常类
	env.Set("InvalidArgumentException", createInvalidArgumentExceptionClass(exceptionClass))
	env.Set("OutOfBoundsException", createOutOfBoundsExceptionClass(runtimeExceptionClass))
	env.Set("NullPointerException", createNullPointerExceptionClass(runtimeExceptionClass))
	env.Set("IOException", createIOExceptionClass(exceptionClass))
	env.Set("TypeError", createTypeErrorClass(exceptionClass))
}

// createExceptionClass 创建 Exception 基类
func createExceptionClass() *Class {
	return &Class{
		Name: "Exception",
		Methods: map[string]*ClassMethod{
			"__construct": {
				Name:           "__construct",
				AccessModifier: "public",
				Parameters:     []interface{}{},
				ReturnType:     []string{},
				Body:           nil, // 内置方法
			},
			"getMessage": {
				Name:           "getMessage",
				AccessModifier: "public",
				Parameters:     []interface{}{},
				ReturnType:     []string{"string"},
				Body:           nil,
			},
			"getCode": {
				Name:           "getCode",
				AccessModifier: "public",
				Parameters:     []interface{}{},
				ReturnType:     []string{"int"},
				Body:           nil,
			},
			"getFile": {
				Name:           "getFile",
				AccessModifier: "public",
				Parameters:     []interface{}{},
				ReturnType:     []string{"string"},
				Body:           nil,
			},
			"getLine": {
				Name:           "getLine",
				AccessModifier: "public",
				Parameters:     []interface{}{},
				ReturnType:     []string{"int"},
				Body:           nil,
			},
			"getTrace": {
				Name:           "getTrace",
				AccessModifier: "public",
				Parameters:     []interface{}{},
				ReturnType:     []string{"[]string"},
				Body:           nil,
			},
			"getTraceAsString": {
				Name:           "getTraceAsString",
				AccessModifier: "public",
				Parameters:     []interface{}{},
				ReturnType:     []string{"string"},
				Body:           nil,
			},
			"getCause": {
				Name:           "getCause",
				AccessModifier: "public",
				Parameters:     []interface{}{},
				ReturnType:     []string{"Exception"},
				Body:           nil,
			},
			"toString": {
				Name:           "toString",
				AccessModifier: "public",
				Parameters:     []interface{}{},
				ReturnType:     []string{"string"},
				Body:           nil,
			},
			"printStackTrace": {
				Name:           "printStackTrace",
				AccessModifier: "public",
				Parameters:     []interface{}{},
				ReturnType:     []string{},
				Body:           nil,
			},
		},
		Variables: map[string]*ClassVariable{
			"message":    {Name: "message", Type: "string", AccessModifier: "protected"},
			"code":       {Name: "code", Type: "int", AccessModifier: "protected"},
			"file":       {Name: "file", Type: "string", AccessModifier: "protected"},
			"line":       {Name: "line", Type: "int", AccessModifier: "protected"},
			"stackTrace": {Name: "stackTrace", Type: "[]string", AccessModifier: "protected"},
			"cause":      {Name: "cause", Type: "Exception", AccessModifier: "protected"},
		},
	}
}

// createRuntimeExceptionClass 创建 RuntimeException 类
func createRuntimeExceptionClass(parent *Class) *Class {
	return &Class{
		Name:   "RuntimeException",
		Parent: parent,
		Methods: map[string]*ClassMethod{
			"__construct": {
				Name:           "__construct",
				AccessModifier: "public",
				Parameters:     []interface{}{},
				ReturnType:     []string{},
				Body:           nil,
			},
		},
		Variables: map[string]*ClassVariable{},
	}
}

// createInvalidArgumentExceptionClass 创建 InvalidArgumentException 类
func createInvalidArgumentExceptionClass(parent *Class) *Class {
	return &Class{
		Name:      "InvalidArgumentException",
		Parent:    parent,
		Methods:   map[string]*ClassMethod{},
		Variables: map[string]*ClassVariable{},
	}
}

// createOutOfBoundsExceptionClass 创建 OutOfBoundsException 类
func createOutOfBoundsExceptionClass(parent *Class) *Class {
	return &Class{
		Name:      "OutOfBoundsException",
		Parent:    parent,
		Methods:   map[string]*ClassMethod{},
		Variables: map[string]*ClassVariable{},
	}
}

// createNullPointerExceptionClass 创建 NullPointerException 类
func createNullPointerExceptionClass(parent *Class) *Class {
	return &Class{
		Name:      "NullPointerException",
		Parent:    parent,
		Methods:   map[string]*ClassMethod{},
		Variables: map[string]*ClassVariable{},
	}
}

// createIOExceptionClass 创建 IOException 类
func createIOExceptionClass(parent *Class) *Class {
	return &Class{
		Name:      "IOException",
		Parent:    parent,
		Methods:   map[string]*ClassMethod{},
		Variables: map[string]*ClassVariable{},
	}
}

// createTypeErrorClass 创建 TypeError 类
func createTypeErrorClass(parent *Class) *Class {
	return &Class{
		Name:      "TypeError",
		Parent:    parent,
		Methods:   map[string]*ClassMethod{},
		Variables: map[string]*ClassVariable{},
	}
}

// evalExceptionMethod 执行异常对象的方法
func (i *Interpreter) evalExceptionMethod(instance *Instance, methodName string, args []Object) Object {
	switch methodName {
	case "getMessage":
		if msg, ok := instance.Fields["message"]; ok {
			return msg
		}
		return &String{Value: ""}

	case "getCode":
		if code, ok := instance.Fields["code"]; ok {
			return code
		}
		return &Integer{Value: 0}

	case "getFile":
		if file, ok := instance.Fields["file"]; ok {
			return file
		}
		return &String{Value: ""}

	case "getLine":
		if line, ok := instance.Fields["line"]; ok {
			return line
		}
		return &Integer{Value: 0}

	case "getTrace":
		if trace, ok := instance.Fields["stackTrace"]; ok {
			return trace
		}
		return &Array{Elements: []Object{}, ElementType: "string"}

	case "getTraceAsString":
		if trace, ok := instance.Fields["stackTrace"]; ok {
			if arr, ok := trace.(*Array); ok {
				var sb strings.Builder
				for _, elem := range arr.Elements {
					if str, ok := elem.(*String); ok {
						sb.WriteString(str.Value)
						sb.WriteString("\n")
					}
				}
				return &String{Value: sb.String()}
			}
		}
		return &String{Value: ""}

	case "getCause":
		if cause, ok := instance.Fields["cause"]; ok {
			return cause
		}
		return &Null{}

	case "toString":
		return &String{Value: i.formatException(instance)}

	case "printStackTrace":
		fmt.Println(i.formatException(instance))
		return &Null{}

	default:
		return newError("异常没有方法: %s", methodName)
	}
}

// formatException 格式化异常为字符串
func (i *Interpreter) formatException(instance *Instance) string {
	var sb strings.Builder

	// 异常类型和消息
	sb.WriteString(instance.Class.Name)
	sb.WriteString(": ")
	if msg, ok := instance.Fields["message"]; ok {
		if str, ok := msg.(*String); ok {
			sb.WriteString(str.Value)
		}
	}
	sb.WriteString("\n")

	// 堆栈跟踪
	if trace, ok := instance.Fields["stackTrace"]; ok {
		if arr, ok := trace.(*Array); ok {
			for _, elem := range arr.Elements {
				if str, ok := elem.(*String); ok {
					sb.WriteString(str.Value)
					sb.WriteString("\n")
				}
			}
		}
	}

	// 异常链
	if cause, ok := instance.Fields["cause"]; ok {
		if causeInstance, ok := cause.(*Instance); ok {
			sb.WriteString("Caused by: ")
			sb.WriteString(i.formatException(causeInstance))
		}
	}

	return sb.String()
}

// createExceptionInstance 创建异常实例
// message: 异常消息
// code: 异常代码
// cause: 导致此异常的原始异常（可以为 nil）
func (i *Interpreter) createExceptionInstance(class *Class, message string, code int64, cause Object) *Instance {
	instance := &Instance{
		Class:  class,
		Fields: make(map[string]Object),
	}

	// 设置异常属性
	instance.Fields["message"] = &String{Value: message}
	instance.Fields["code"] = &Integer{Value: code}
	instance.Fields["file"] = &String{Value: i.currentFileName}
	instance.Fields["line"] = &Integer{Value: 0}

	// 捕获堆栈跟踪
	stackTrace := i.captureStackTrace()
	stackTraceArray := &Array{
		Elements:    make([]Object, len(stackTrace)),
		ElementType: "string",
	}
	for idx, trace := range stackTrace {
		stackTraceArray.Elements[idx] = &String{Value: trace}
	}
	instance.Fields["stackTrace"] = stackTraceArray

	// 设置异常链
	if cause != nil {
		instance.Fields["cause"] = cause
	} else {
		instance.Fields["cause"] = &Null{}
	}

	return instance
}

// createThrownException 创建 ThrownException 对象
func (i *Interpreter) createThrownException(class *Class, message string, code int64, cause *ThrownException) *ThrownException {
	var causeInstance Object = &Null{}
	if cause != nil && cause.Exception != nil {
		causeInstance = cause.Exception
	}

	instance := i.createExceptionInstance(class, message, code, causeInstance)

	return &ThrownException{
		Exception:  instance,
		StackTrace: i.captureStackTrace(),
		Cause:      cause,
	}
}




