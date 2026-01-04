package compiler

import (
	"github.com/tangzhangming/longlang/internal/parser"
)

// SymbolType 符号类型
type SymbolType int

const (
	SymbolTypeVariable SymbolType = iota
	SymbolTypeFunction
	SymbolTypeClass
	SymbolTypeInterface
	SymbolTypeEnum
	SymbolTypeNamespace
	SymbolTypeType
)

// Symbol 符号信息
type Symbol struct {
	Name       string      // 符号名称
	Type       SymbolType  // 符号类型
	GoType     string      // Go 类型名称
	Namespace  string      // 所属命名空间
	IsExported bool        // 是否导出（public）
	IsStatic   bool        // 是否是静态的（类方法/字段）
	Node       parser.Node // 对应的 AST 节点
}

// SymbolTable 符号表
type SymbolTable struct {
	symbols      map[string]*Symbol // 符号映射：完全限定名 -> 符号
	namespaces   map[string]bool     // 命名空间集合
	classes      map[string]*Symbol  // 类符号：完全限定名 -> 符号
	interfaces   map[string]*Symbol  // 接口符号
	enums        map[string]*Symbol  // 枚举符号
	functions    map[string]*Symbol  // 函数符号
	variables    map[string]*Symbol  // 变量符号
	currentScope string              // 当前作用域（命名空间）
}

// NewSymbolTable 创建新的符号表
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols:    make(map[string]*Symbol),
		namespaces: make(map[string]bool),
		classes:    make(map[string]*Symbol),
		interfaces: make(map[string]*Symbol),
		enums:      make(map[string]*Symbol),
		functions:  make(map[string]*Symbol),
		variables:  make(map[string]*Symbol),
	}
}

// AddSymbol 添加符号
func (st *SymbolTable) AddSymbol(symbol *Symbol) {
	fullName := st.getFullName(symbol.Name, symbol.Namespace)
	st.symbols[fullName] = symbol

	switch symbol.Type {
	case SymbolTypeClass:
		st.classes[fullName] = symbol
	case SymbolTypeInterface:
		st.interfaces[fullName] = symbol
	case SymbolTypeEnum:
		st.enums[fullName] = symbol
	case SymbolTypeFunction:
		st.functions[fullName] = symbol
	case SymbolTypeVariable:
		st.variables[fullName] = symbol
	}
}

// GetSymbol 获取符号
func (st *SymbolTable) GetSymbol(name, namespace string) (*Symbol, bool) {
	// 先尝试完全限定名
	fullName := st.getFullName(name, namespace)
	if symbol, ok := st.symbols[fullName]; ok {
		return symbol, true
	}

	// 尝试在当前命名空间中查找
	if namespace != "" {
		fullName = st.getFullName(name, namespace)
		if symbol, ok := st.symbols[fullName]; ok {
			return symbol, true
		}
	}

	// 尝试全局查找（无命名空间）
	if symbol, ok := st.symbols[name]; ok {
		return symbol, true
	}

	return nil, false
}

// AddNamespace 添加命名空间
func (st *SymbolTable) AddNamespace(namespace string) {
	st.namespaces[namespace] = true
}

// HasNamespace 检查命名空间是否存在
func (st *SymbolTable) HasNamespace(namespace string) bool {
	return st.namespaces[namespace]
}

// GetClass 获取类符号
func (st *SymbolTable) GetClass(name, namespace string) (*Symbol, bool) {
	fullName := st.getFullName(name, namespace)
	symbol, ok := st.classes[fullName]
	return symbol, ok
}

// GetInterface 获取接口符号
func (st *SymbolTable) GetInterface(name, namespace string) (*Symbol, bool) {
	fullName := st.getFullName(name, namespace)
	symbol, ok := st.interfaces[fullName]
	return symbol, ok
}

// GetEnum 获取枚举符号
func (st *SymbolTable) GetEnum(name, namespace string) (*Symbol, bool) {
	fullName := st.getFullName(name, namespace)
	symbol, ok := st.enums[fullName]
	return symbol, ok
}

// GetFunction 获取函数符号
func (st *SymbolTable) GetFunction(name, namespace string) (*Symbol, bool) {
	fullName := st.getFullName(name, namespace)
	symbol, ok := st.functions[fullName]
	return symbol, ok
}

// SetCurrentScope 设置当前作用域
func (st *SymbolTable) SetCurrentScope(namespace string) {
	st.currentScope = namespace
}

// GetCurrentScope 获取当前作用域
func (st *SymbolTable) GetCurrentScope() string {
	return st.currentScope
}

// getFullName 获取完全限定名
func (st *SymbolTable) getFullName(name, namespace string) string {
	if namespace == "" {
		return name
	}
	return namespace + "." + name
}

// GetAllClasses 获取所有类
func (st *SymbolTable) GetAllClasses() map[string]*Symbol {
	return st.classes
}

// GetAllInterfaces 获取所有接口
func (st *SymbolTable) GetAllInterfaces() map[string]*Symbol {
	return st.interfaces
}

// GetAllEnums 获取所有枚举
func (st *SymbolTable) GetAllEnums() map[string]*Symbol {
	return st.enums
}

// GetAllFunctions 获取所有函数
func (st *SymbolTable) GetAllFunctions() map[string]*Symbol {
	return st.functions
}



