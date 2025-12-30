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
	INTEGER_OBJ         ObjectType = "INTEGER"         // 整数类型
	FLOAT_OBJ           ObjectType = "FLOAT"           // 浮点数类型
	STRING_OBJ          ObjectType = "STRING"          // 字符串类型
	BOOLEAN_OBJ         ObjectType = "BOOLEAN"         // 布尔类型
	NULL_OBJ            ObjectType = "NULL"            // null 类型
	RETURN_VALUE_OBJ    ObjectType = "RETURN_VALUE"    // 返回值类型（用于函数返回）
	ERROR_OBJ           ObjectType = "ERROR"           // 错误类型
	FUNCTION_OBJ        ObjectType = "FUNCTION"        // 函数类型
	BUILTIN_OBJ         ObjectType = "BUILTIN"         // 内置函数类型
	ANY_OBJ             ObjectType = "ANY"             // 任意类型（未完全实现）
	CLASS_OBJ           ObjectType = "CLASS"           // 类类型
	INSTANCE_OBJ        ObjectType = "INSTANCE"        // 类实例类型
	PACKAGE_OBJ         ObjectType = "PACKAGE"         // 包类型
	BREAK_SIGNAL_OBJ    ObjectType = "BREAK_SIGNAL"    // break 信号
	CONTINUE_SIGNAL_OBJ ObjectType = "CONTINUE_SIGNAL" // continue 信号
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

// Float 浮点数对象
// 表示一个浮点数值
type Float struct {
	Value float64 // 浮点数值
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return fmt.Sprintf("%g", f.Value) }

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

// BreakSignal break 信号对象
// 用于在 for 循环中处理 break 语句
type BreakSignal struct{}

func (bs *BreakSignal) Type() ObjectType { return BREAK_SIGNAL_OBJ }
func (bs *BreakSignal) Inspect() string  { return "break" }

// ContinueSignal continue 信号对象
// 用于在 for 循环中处理 continue 语句
type ContinueSignal struct{}

func (cs *ContinueSignal) Type() ObjectType { return CONTINUE_SIGNAL_OBJ }
func (cs *ContinueSignal) Inspect() string  { return "continue" }

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

// ========== 类和实例对象 ==========

// Class 类对象
// 表示一个类定义，包含类的方法和成员变量定义
type Class struct {
	Name          string                    // 类名
	Variables     map[string]*ClassVariable // 成员变量定义
	Methods       map[string]*ClassMethod   // 实例方法
	StaticMethods map[string]*ClassMethod   // 静态方法
	Env           *Environment              // 类定义时的环境（用于闭包）
}

func (c *Class) Type() ObjectType { return CLASS_OBJ }
func (c *Class) Inspect() string  { return "class " + c.Name }

// ClassVariable 类成员变量定义
type ClassVariable struct {
	Name           string      // 变量名
	Type           string      // 变量类型
	AccessModifier string      // 访问修饰符：public, private, protected
	DefaultValue   Object       // 默认值（可选）
}

// ClassMethod 类方法定义
type ClassMethod struct {
	Name           string                    // 方法名
	AccessModifier string                    // 访问修饰符
	IsStatic       bool                      // 是否是静态方法
	Parameters     []interface{}             // 参数列表（*parser.FunctionParameter）
	ReturnType     []string                  // 返回类型
	Body           interface{}               // 方法体（*parser.BlockStatement）
	Env            *Environment              // 方法定义时的环境
}

// Instance 类实例对象
// 表示一个类的实例，包含实例的成员变量值
type Instance struct {
	Class  *Class              // 所属的类
	Fields map[string]Object   // 实例的成员变量值
}

func (i *Instance) Type() ObjectType { return INSTANCE_OBJ }
func (i *Instance) Inspect() string  { return "instance of " + i.Class.Name }

// GetField 获取实例字段
func (i *Instance) GetField(name string) (Object, bool) {
	obj, ok := i.Fields[name]
	return obj, ok
}

// SetField 设置实例字段
func (i *Instance) SetField(name string, val Object) Object {
	i.Fields[name] = val
	return val
}

// Package 包对象
// 表示一个已加载的包
type Package struct {
	Name    string            // 包名
	Path    string            // 包路径
	Exports map[string]Object // 导出的符号
}

func (p *Package) Type() ObjectType { return PACKAGE_OBJ }
func (p *Package) Inspect() string  { return "package " + p.Name }

// BoundMethod 绑定方法对象
// 表示一个绑定了实例的方法（用于 this 访问）
type BoundMethod struct {
	Instance *Instance    // 绑定的实例
	Method   *ClassMethod // 方法定义
}

func (bm *BoundMethod) Type() ObjectType { return FUNCTION_OBJ }
func (bm *BoundMethod) Inspect() string  { return "bound method " + bm.Method.Name }
