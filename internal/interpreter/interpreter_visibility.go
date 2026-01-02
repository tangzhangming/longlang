package interpreter

import (
	"strings"
)

// isNamespaceTreeRelated 检查两个命名空间是否在同一命名空间树中
// 命名空间 A 和 B 属于同一命名空间树，如果：
// - A 是 B 的前缀（父命名空间）
// - B 是 A 的前缀（子命名空间）
// - A 等于 B（同一命名空间）
func isNamespaceTreeRelated(ns1, ns2 string) bool {
	if ns1 == "" || ns2 == "" {
		return true // 空命名空间可以访问任何类
	}
	if ns1 == ns2 {
		return true
	}
	// 检查是否是父子关系
	return strings.HasPrefix(ns1+".", ns2+".") || strings.HasPrefix(ns2+".", ns1+".")
}

// checkClassVisibility 检查类的可见性
// 返回 nil 如果可访问，否则返回错误
func (i *Interpreter) checkClassVisibility(class *Class) Object {
	if class.IsPublic {
		return nil // public 类始终可访问
	}
	
	// internal 类只能在同一命名空间树内访问
	currentNS := ""
	if i.currentNamespace != nil {
		currentNS = i.currentNamespace.FullName
	}
	
	if !isNamespaceTreeRelated(currentNS, class.Namespace) {
		return newError("无法访问 '%s.%s': '%s' 是 internal 类型，只能在 '%s' 命名空间树内访问",
			class.Namespace, class.Name, class.Name, class.Namespace)
	}
	
	return nil
}

// checkInterfaceVisibility 检查接口的可见性
// 返回 nil 如果可访问，否则返回错误
func (i *Interpreter) checkInterfaceVisibility(iface *Interface) Object {
	if iface.IsPublic {
		return nil // public 接口始终可访问
	}
	
	// internal 接口只能在同一命名空间树内访问
	currentNS := ""
	if i.currentNamespace != nil {
		currentNS = i.currentNamespace.FullName
	}
	
	if !isNamespaceTreeRelated(currentNS, iface.Namespace) {
		return newError("无法访问 '%s.%s': '%s' 是 internal 接口，只能在 '%s' 命名空间树内访问",
			iface.Namespace, iface.Name, iface.Name, iface.Namespace)
	}
	
	return nil
}

// checkEnumVisibility 检查枚举的可见性
// 返回 nil 如果可访问，否则返回错误
func (i *Interpreter) checkEnumVisibility(enum *Enum) Object {
	if enum.IsPublic {
		return nil // public 枚举始终可访问
	}
	
	// internal 枚举只能在同一命名空间树内访问
	currentNS := ""
	if i.currentNamespace != nil {
		currentNS = i.currentNamespace.FullName
	}
	
	if !isNamespaceTreeRelated(currentNS, enum.Namespace) {
		return newError("无法访问 '%s.%s': '%s' 是 internal 枚举，只能在 '%s' 命名空间树内访问",
			enum.Namespace, enum.Name, enum.Name, enum.Namespace)
	}
	
	return nil
}

