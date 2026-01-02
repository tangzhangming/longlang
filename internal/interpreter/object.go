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
	INTEGER_OBJ           ObjectType = "INTEGER"           // 整数类型
	FLOAT_OBJ             ObjectType = "FLOAT"             // 浮点数类型
	STRING_OBJ            ObjectType = "STRING"            // 字符串类型
	BOOLEAN_OBJ           ObjectType = "BOOLEAN"           // 布尔类型
	NULL_OBJ              ObjectType = "NULL"              // null 类型
	ARRAY_OBJ             ObjectType = "ARRAY"             // 数组类型
	MAP_OBJ               ObjectType = "MAP"               // Map 类型
	RETURN_VALUE_OBJ      ObjectType = "RETURN_VALUE"      // 返回值类型（用于函数返回）
	ERROR_OBJ             ObjectType = "ERROR"             // 错误类型
	FUNCTION_OBJ          ObjectType = "FUNCTION"          // 函数类型
	BUILTIN_OBJ           ObjectType = "BUILTIN"           // 内置函数类型
	ANY_OBJ               ObjectType = "ANY"               // 任意类型（未完全实现）
	CLASS_OBJ             ObjectType = "CLASS"             // 类类型
	INTERFACE_OBJ         ObjectType = "INTERFACE"         // 接口类型
	INSTANCE_OBJ          ObjectType = "INSTANCE"          // 类实例类型
	PACKAGE_OBJ           ObjectType = "PACKAGE"           // 包类型
	BREAK_SIGNAL_OBJ      ObjectType = "BREAK_SIGNAL"      // break 信号
	CONTINUE_SIGNAL_OBJ   ObjectType = "CONTINUE_SIGNAL"   // continue 信号
	THROWN_EXCEPTION_OBJ  ObjectType = "THROWN_EXCEPTION"  // 抛出的异常信号
	ENUM_OBJ              ObjectType = "ENUM"              // 枚举类型
	ENUM_VALUE_OBJ        ObjectType = "ENUM_VALUE"        // 枚举值类型
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

// ========== 数组对象 ==========

// Array 数组对象
// 表示一个数组，可以是固定长度数组或动态数组（切片）
type Array struct {
	Elements    []Object // 数组元素
	ElementType string   // 元素类型（如 "int", "string" 等，空表示推导）
	IsFixed     bool     // 是否为固定长度数组（false 表示切片）
	Capacity    int64    // 数组容量（固定长度数组的大小）
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out strings.Builder
	elements := make([]string, len(a.Elements))
	for i, e := range a.Elements {
		elements[i] = e.Inspect()
	}
	out.WriteString("{")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("}")
	return out.String()
}

// ========== Map 对象 ==========

// Map Map 对象
// 表示一个键值对映射，键为字符串
type Map struct {
	Pairs     map[string]Object // 键值对（使用 Go 的 map）
	Keys      []string          // 有序的键列表（保持插入顺序）
	KeyType   string            // 键类型（目前仅支持 "string"）
	ValueType string            // 值类型（如 "int", "string" 等）
}

func (m *Map) Type() ObjectType { return MAP_OBJ }
func (m *Map) Inspect() string {
	var out strings.Builder
	out.WriteString("map[")
	out.WriteString(m.KeyType)
	out.WriteString("]")
	out.WriteString(m.ValueType)
	out.WriteString("{")
	for i, key := range m.Keys {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString("\"")
		out.WriteString(key)
		out.WriteString("\": ")
		out.WriteString(m.Pairs[key].Inspect())
	}
	out.WriteString("}")
	return out.String()
}

// Get 获取 Map 中的值
func (m *Map) Get(key string) (Object, bool) {
	val, ok := m.Pairs[key]
	return val, ok
}

// Set 设置 Map 中的值
func (m *Map) Set(key string, value Object) {
	if _, exists := m.Pairs[key]; !exists {
		m.Keys = append(m.Keys, key)
	}
	m.Pairs[key] = value
}

// Delete 删除 Map 中的键值对
func (m *Map) Delete(key string) bool {
	if _, exists := m.Pairs[key]; !exists {
		return false
	}
	delete(m.Pairs, key)
	// 从 Keys 中移除
	for i, k := range m.Keys {
		if k == key {
			m.Keys = append(m.Keys[:i], m.Keys[i+1:]...)
			break
		}
	}
	return true
}

// Size 返回 Map 的大小
func (m *Map) Size() int {
	return len(m.Pairs)
}

// IsEmpty 判断 Map 是否为空
func (m *Map) IsEmpty() bool {
	return len(m.Pairs) == 0
}

// Has 判断 Map 是否包含某个键
func (m *Map) Has(key string) bool {
	_, ok := m.Pairs[key]
	return ok
}

// Clear 清空 Map
func (m *Map) Clear() {
	m.Pairs = make(map[string]Object)
	m.Keys = []string{}
}

// GetKeys 获取所有键
func (m *Map) GetKeys() []string {
	return m.Keys
}

// GetValues 获取所有值
func (m *Map) GetValues() []Object {
	values := make([]Object, len(m.Keys))
	for i, key := range m.Keys {
		values[i] = m.Pairs[key]
	}
	return values
}

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

// ThrownException 抛出的异常信号对象
// 用于在解释器中传递被 throw 语句抛出的异常
// 当解释器遇到这个对象时，会沿着调用栈向上传递，直到被 catch 捕获
type ThrownException struct {
	Exception    *Instance // 被抛出的异常对象（必须是 Exception 类或其子类的实例）
	RuntimeError *Error    // 内置运行时错误（用于除零等运行时错误）
	StackTrace   []string  // 调用栈信息
}

func (te *ThrownException) Type() ObjectType { return THROWN_EXCEPTION_OBJ }
func (te *ThrownException) Inspect() string {
	if te.Exception != nil {
		// 尝试获取异常消息
		if msg, ok := te.Exception.Fields["message"]; ok {
			return "ThrownException: " + msg.Inspect()
		}
		return "ThrownException: " + te.Exception.Class.Name
	}
	if te.RuntimeError != nil {
		return "ThrownException: " + te.RuntimeError.Message
	}
	return "ThrownException"
}

// GetMessage 获取异常消息
func (te *ThrownException) GetMessage() string {
	if te.Exception != nil {
		if msg, ok := te.Exception.Fields["message"]; ok {
			if strMsg, ok := msg.(*String); ok {
				return strMsg.Value
			}
		}
	}
	if te.RuntimeError != nil {
		return te.RuntimeError.Message
	}
	return ""
}

// GetExceptionType 获取异常类型名称
func (te *ThrownException) GetExceptionType() string {
	if te.Exception != nil {
		return te.Exception.Class.Name
	}
	if te.RuntimeError != nil {
		return "RuntimeException"
	}
	return "Exception"
}

// IsInstanceOf 检查异常是否是指定类型或其子类
func (te *ThrownException) IsInstanceOf(className string) bool {
	if te.Exception != nil {
		class := te.Exception.Class
		for class != nil {
			if class.Name == className {
				return true
			}
			class = class.Parent
		}
		return false
	}
	// 内置运行时错误，匹配 Exception 或 RuntimeException
	if te.RuntimeError != nil {
		return className == "Exception" || className == "RuntimeException"
	}
	return false
}

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

// Interface 接口对象
// 表示一个接口定义，包含接口的方法签名
type Interface struct {
	Name    string                      // 接口名
	Methods map[string]*InterfaceMethod // 方法签名
}

func (i *Interface) Type() ObjectType { return INTERFACE_OBJ }
func (i *Interface) Inspect() string  { return "interface " + i.Name }

// InterfaceMethod 接口方法签名
type InterfaceMethod struct {
	Name       string   // 方法名
	Parameters []string // 参数类型列表
	ReturnType []string // 返回类型
}

// Class 类对象
// 表示一个类定义，包含类的方法和成员变量定义
type Class struct {
	Name          string                    // 类名
	Parent        *Class                    // 父类（用于继承）
	Interfaces    []*Interface              // 实现的接口列表
	Variables     map[string]*ClassVariable // 成员变量定义
	Constants     map[string]*ClassConstant // 常量定义
	Methods       map[string]*ClassMethod   // 实例方法
	StaticMethods map[string]*ClassMethod   // 静态方法
	Env           *Environment              // 类定义时的环境（用于闭包）
	IsExported    bool                      // 是否为导出类（与文件名相同）
	IsAbstract    bool                      // 是否是抽象类
}

func (c *Class) Type() ObjectType { return CLASS_OBJ }
func (c *Class) Inspect() string {
	result := ""
	if c.IsAbstract {
		result += "abstract "
	}
	result += "class " + c.Name
	if c.Parent != nil {
		result += " extends " + c.Parent.Name
	}
	if len(c.Interfaces) > 0 {
		result += " implements "
		for i, iface := range c.Interfaces {
			if i > 0 {
				result += ", "
			}
			result += iface.Name
		}
	}
	return result
}

// Implements 检查类是否实现了指定接口
func (c *Class) Implements(iface *Interface) bool {
	for methodName, ifaceMethod := range iface.Methods {
		classMethod, ok := c.GetMethod(methodName)
		if !ok {
			return false
		}
		// 检查返回类型
		if len(classMethod.ReturnType) != len(ifaceMethod.ReturnType) {
			return false
		}
		for i, rt := range ifaceMethod.ReturnType {
			if classMethod.ReturnType[i] != rt {
				return false
			}
		}
	}
	return true
}

// GetMethod 获取方法（包括继承的方法）
func (c *Class) GetMethod(name string) (*ClassMethod, bool) {
	if method, ok := c.Methods[name]; ok {
		return method, true
	}
	// 从父类查找
	if c.Parent != nil {
		return c.Parent.GetMethod(name)
	}
	return nil, false
}

// GetStaticMethod 获取静态方法（包括继承的方法）
func (c *Class) GetStaticMethod(name string) (*ClassMethod, bool) {
	if method, ok := c.StaticMethods[name]; ok {
		return method, true
	}
	// 从父类查找
	if c.Parent != nil {
		return c.Parent.GetStaticMethod(name)
	}
	return nil, false
}

// GetVariable 获取成员变量定义（包括继承的变量）
func (c *Class) GetVariable(name string) (*ClassVariable, bool) {
	if variable, ok := c.Variables[name]; ok {
		return variable, true
	}
	// 从父类查找
	if c.Parent != nil {
		return c.Parent.GetVariable(name)
	}
	return nil, false
}

// GetConstant 获取常量（包括继承的常量，但子类同名常量会覆盖父类）
func (c *Class) GetConstant(name string) (*ClassConstant, bool) {
	if constant, ok := c.Constants[name]; ok {
		return constant, true
	}
	// 从父类查找
	if c.Parent != nil {
		return c.Parent.GetConstant(name)
	}
	return nil, false
}

// ClassVariable 类成员变量定义
type ClassVariable struct {
	Name           string      // 变量名
	Type           string      // 变量类型
	AccessModifier string      // 访问修饰符：public, private, protected
	DefaultValue   Object       // 默认值（可选）
}

// ClassConstant 类常量定义
type ClassConstant struct {
	Name           string      // 常量名
	Type           string      // 常量类型（空字符串表示类型推导）
	AccessModifier string      // 访问修饰符：public, private, protected
	Value          Object      // 常量值
}

// ClassMethod 类方法定义
type ClassMethod struct {
	Name           string                    // 方法名
	AccessModifier string                    // 访问修饰符
	IsStatic       bool                      // 是否是静态方法
	IsAbstract     bool                      // 是否是抽象方法
	Parameters     []interface{}             // 参数列表（*parser.FunctionParameter）
	ReturnType     []string                  // 返回类型
	Body           interface{}               // 方法体（*parser.BlockStatement，抽象方法时为 nil）
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

// BoundStringMethod 绑定字符串方法对象
// 表示一个绑定了字符串实例的方法
type BoundStringMethod struct {
	String *String                              // 绑定的字符串
	Method func(s *String, args ...Object) Object // 方法实现
	Name   string                               // 方法名
}

func (bsm *BoundStringMethod) Type() ObjectType { return FUNCTION_OBJ }
func (bsm *BoundStringMethod) Inspect() string  { return "string method " + bsm.Name }

// BoundMapMethod 绑定 Map 方法对象
// 表示一个绑定了 Map 实例的方法
type BoundMapMethod struct {
	Map    *Map                                // 绑定的 Map
	Method func(m *Map, args ...Object) Object // 方法实现
	Name   string                              // 方法名
}

func (bmm *BoundMapMethod) Type() ObjectType { return FUNCTION_OBJ }
func (bmm *BoundMapMethod) Inspect() string  { return "map method " + bmm.Name }

// BoundArrayMethod 绑定数组方法对象
// 表示一个绑定了数组实例的方法
type BoundArrayMethod struct {
	Array  *Array                                // 绑定的数组
	Method func(a *Array, args ...Object) Object // 方法实现
	Name   string                                // 方法名
}

func (bam *BoundArrayMethod) Type() ObjectType { return FUNCTION_OBJ }
func (bam *BoundArrayMethod) Inspect() string  { return "array method " + bam.Name }

// ========== 枚举类型 ==========

// Enum 枚举类型定义
// 表示一个枚举类型，包含所有枚举成员
type Enum struct {
	Name        string                  // 枚举名
	BackingType string                  // 底层类型（"int"、"string" 或 ""）
	Members     map[string]*EnumValue   // 枚举成员（按名称）
	MemberList  []*EnumValue            // 枚举成员列表（保持顺序）
	Methods     map[string]*ClassMethod // 枚举方法
	Variables   map[string]*ClassVariable // 枚举字段（用于复杂枚举）
	Interfaces  []*Interface            // 实现的接口
	Env         *Environment            // 枚举定义时的环境
}

func (e *Enum) Type() ObjectType { return ENUM_OBJ }
func (e *Enum) Inspect() string {
	var out string
	out += "enum " + e.Name
	if e.BackingType != "" {
		out += ": " + e.BackingType
	}
	out += " { "
	for i, member := range e.MemberList {
		if i > 0 {
			out += ", "
		}
		out += member.Name
	}
	out += " }"
	return out
}

// GetMember 获取枚举成员
func (e *Enum) GetMember(name string) (*EnumValue, bool) {
	member, ok := e.Members[name]
	return member, ok
}

// GetMethod 获取枚举方法
func (e *Enum) GetMethod(name string) (*ClassMethod, bool) {
	method, ok := e.Methods[name]
	return method, ok
}

// EnumValue 枚举值对象
// 表示一个枚举成员的实例
type EnumValue struct {
	Enum    *Enum             // 所属枚举类型
	Name    string            // 成员名称
	Ordinal int               // 序号（从0开始）
	Value   Object            // 成员值（带值枚举使用）
	Fields  map[string]Object // 字段值（复杂枚举使用）
}

func (ev *EnumValue) Type() ObjectType { return ENUM_VALUE_OBJ }
func (ev *EnumValue) Inspect() string {
	return ev.Enum.Name + "::" + ev.Name
}

// GetName 获取成员名称
func (ev *EnumValue) GetName() string {
	return ev.Name
}

// GetValue 获取成员值
func (ev *EnumValue) GetValue() Object {
	return ev.Value
}

// GetOrdinal 获取序号
func (ev *EnumValue) GetOrdinal() int {
	return ev.Ordinal
}

// BoundEnumMethod 绑定枚举值方法对象
// 表示一个绑定了枚举值实例的方法
type BoundEnumMethod struct {
	EnumValue  *EnumValue   // 绑定的枚举值
	Method     *ClassMethod // 方法定义（内置方法时为 nil）
	Enum       *Enum        // 所属枚举类型
	MethodName string       // 方法名（用于内置方法）
}

func (bem *BoundEnumMethod) Type() ObjectType { return FUNCTION_OBJ }
func (bem *BoundEnumMethod) Inspect() string {
	if bem.Method != nil {
		return "enum method " + bem.Method.Name
	}
	return "enum method " + bem.MethodName
}

// ========== 辅助函数 ==========

// NewError 创建错误对象
func NewError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}
