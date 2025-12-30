package interpreter

import (
	"fmt"
	"strings"
)

// ========== 对象类型定义 ==========

// ObjectType 对象类型枚举
// 表示解释器中支持的所有对象类型
type ObjectType string

const (
	INTEGER_OBJ      ObjectType = "INTEGER"      // 整数类型
	STRING_OBJ       ObjectType = "STRING"       // 字符串类型
	BOOLEAN_OBJ      ObjectType = "BOOLEAN"      // 布尔类型
	NULL_OBJ         ObjectType = "NULL"         // null 类型
	RETURN_VALUE_OBJ ObjectType = "RETURN_VALUE" // 返回值类型（用于函数返回）
	ERROR_OBJ        ObjectType = "ERROR"        // 错误类型
	FUNCTION_OBJ     ObjectType = "FUNCTION"     // 函数类型
	BUILTIN_OBJ      ObjectType = "BUILTIN"      // 内置函数类型
	ANY_OBJ          ObjectType = "ANY"          // 任意类型（未完全实现）
)

// ========== 对象接口 ==========

// Object 对象接口
// 所有解释器中的值都实现了这个接口
// 这是解释器的核心抽象，所有值（整数、字符串、函数等）都是 Object
type Object interface {
	Type() ObjectType // 返回对象的类型
	Inspect() string  // 返回对象的字符串表示（用于调试和打印）
}

// ========== 基本类型对象 ==========

// Integer 整数对象
// 表示一个整数值
type Integer struct {
	Value int64 // 整数值
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

// String 字符串对象
// 表示一个字符串值
type String struct {
	Value string // 字符串值
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

// Boolean 布尔对象
// 表示一个布尔值（true 或 false）
type Boolean struct {
	Value bool // 布尔值
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

// Null null 对象
// 表示空值，类似于其他语言中的 null 或 nil
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// ========== 特殊对象 ==========

// ReturnValue 返回值对象
// 用于包装函数的返回值
// 当函数执行 return 语句时，会创建这个对象
// 解释器会检查这个对象并停止执行当前函数
type ReturnValue struct {
	Value Object // 实际返回的值
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

// Error 错误对象
// 表示运行时错误
// 当解释器遇到错误时（如未定义的变量、类型不匹配等），会创建这个对象
type Error struct {
	Message string // 错误消息
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "错误: " + e.Message }

// ========== 函数对象 ==========

// Function 函数对象
// 表示用户定义的函数
// 包含函数的参数、函数体和捕获的环境（实现闭包）
type Function struct {
	Parameters []interface{} // 函数参数列表（*parser.FunctionParameter）
	Body       interface{}   // 函数体（*parser.BlockStatement）
	Env        *Environment  // 函数定义时的环境（用于闭包）
	ReturnType []string      // 返回类型列表（当前未使用）
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out strings.Builder
	out.WriteString("fn(...)")
	return out.String()
}

// BuiltinFunction 内置函数类型
// 内置函数是用 Go 语言实现的函数，可以直接调用
// 参数:
//   args: 函数参数列表
// 返回:
//   函数执行结果
type BuiltinFunction func(args ...Object) Object

// Builtin 内置函数对象
// 表示一个内置函数（如 fmt.Println）
type Builtin struct {
	Fn BuiltinFunction // 内置函数的实现
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "内置函数" }

// ========== 其他对象 ==========

// Any 任意类型对象
// 用于表示 any 类型的值（当前版本未完全实现）
type Any struct {
	Value Object // 实际的值
}

func (a *Any) Type() ObjectType { return ANY_OBJ }
func (a *Any) Inspect() string  { return a.Value.Inspect() }
