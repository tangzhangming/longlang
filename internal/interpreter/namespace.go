package interpreter

import (
	"fmt"
	"strings"
)

// Namespace 命名空间对象
// 存储命名空间中的类、函数等符号
type Namespace struct {
	FullName  string            // 完全限定名，如 "Mycompany.Myapp.Models"
	Classes   map[string]*Class // 类定义
	Functions map[string]*Function // 函数定义
	Variables map[string]Object   // 变量（常量等）
}

// NewNamespace 创建新的命名空间
func NewNamespace(fullName string) *Namespace {
	return &Namespace{
		FullName:  fullName,
		Classes:   make(map[string]*Class),
		Functions: make(map[string]*Function),
		Variables: make(map[string]Object),
	}
}

// GetClass 获取类
func (ns *Namespace) GetClass(name string) (*Class, bool) {
	class, ok := ns.Classes[name]
	return class, ok
}

// SetClass 设置类
func (ns *Namespace) SetClass(name string, class *Class) {
	ns.Classes[name] = class
}

// GetFunction 获取函数
func (ns *Namespace) GetFunction(name string) (*Function, bool) {
	fn, ok := ns.Functions[name]
	return fn, ok
}

// SetFunction 设置函数
func (ns *Namespace) SetFunction(name string, fn *Function) {
	ns.Functions[name] = fn
}

// NamespaceManager 命名空间管理器
type NamespaceManager struct {
	namespaces map[string]*Namespace // 命名空间表，key为完全限定名
}

// NewNamespaceManager 创建命名空间管理器
func NewNamespaceManager() *NamespaceManager {
	return &NamespaceManager{
		namespaces: make(map[string]*Namespace),
	}
}

// GetNamespace 获取命名空间（如果不存在则创建）
func (nm *NamespaceManager) GetNamespace(fullName string) *Namespace {
	if ns, ok := nm.namespaces[fullName]; ok {
		return ns
	}
	ns := NewNamespace(fullName)
	nm.namespaces[fullName] = ns
	return ns
}

// FindNamespace 查找命名空间（不存在返回nil）
func (nm *NamespaceManager) FindNamespace(fullName string) (*Namespace, bool) {
	ns, ok := nm.namespaces[fullName]
	return ns, ok
}

// ResolveClassName 解析类名（完全限定名）
// 输入：Illuminate.Database.Eloquent.Model
// 返回：命名空间 "Illuminate.Database.Eloquent"，类名 "Model"
func ResolveClassName(fullPath string) (namespace string, className string, err error) {
	parts := strings.Split(fullPath, ".")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("无效的类路径: %s，必须包含命名空间和类名", fullPath)
	}

	className = parts[len(parts)-1]
	namespace = strings.Join(parts[:len(parts)-1], ".")
	return namespace, className, nil
}

// ResolveNamespacePath 解析命名空间路径为文件系统路径
// 输入：Mycompany.Myapp.Models
// 返回：Mycompany/Myapp/Models
func ResolveNamespacePath(namespace string) string {
	return strings.ReplaceAll(namespace, ".", "/")
}


