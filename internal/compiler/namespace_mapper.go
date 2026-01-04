package compiler

import (
	"path/filepath"
	"strings"
)

// NamespaceMapper 命名空间映射器
type NamespaceMapper struct {
	namespaceToPackage map[string]string // 命名空间 -> Go 包路径
	packageToNamespace map[string]string // Go 包路径 -> 命名空间
	useStatements      map[string]string // use 语句映射：别名 -> 完全限定名
}

// NewNamespaceMapper 创建新的命名空间映射器
func NewNamespaceMapper() *NamespaceMapper {
	return &NamespaceMapper{
		namespaceToPackage: make(map[string]string),
		packageToNamespace: make(map[string]string),
		useStatements:      make(map[string]string),
	}
}

// MapNamespaceToPackage 将命名空间映射到 Go 包路径
func (nm *NamespaceMapper) MapNamespaceToPackage(namespace string) string {
	if pkg, ok := nm.namespaceToPackage[namespace]; ok {
		return pkg
	}

	// 默认映射规则：App.Models -> app/models
	parts := strings.Split(namespace, ".")
	var pkgParts []string
	for _, part := range parts {
		pkgParts = append(pkgParts, strings.ToLower(part))
	}
	pkg := strings.Join(pkgParts, "/")

	nm.namespaceToPackage[namespace] = pkg
	nm.packageToNamespace[pkg] = namespace

	return pkg
}

// GetPackagePath 获取包的完整路径（相对于项目根目录）
func (nm *NamespaceMapper) GetPackagePath(namespace string) string {
	pkg := nm.MapNamespaceToPackage(namespace)
	return filepath.Join("src", pkg)
}

// AddUseStatement 添加 use 语句
func (nm *NamespaceMapper) AddUseStatement(path, alias string) {
	if alias != "" {
		nm.useStatements[alias] = path
	} else {
		// 如果没有别名，使用最后一部分作为别名
		parts := strings.Split(path, ".")
		if len(parts) > 0 {
			nm.useStatements[parts[len(parts)-1]] = path
		}
	}
}

// GetUseAlias 获取 use 语句的别名
func (nm *NamespaceMapper) GetUseAlias(fullPath string) string {
	for alias, path := range nm.useStatements {
		if path == fullPath {
			return alias
		}
	}
	// 如果没有找到，返回最后一部分
	parts := strings.Split(fullPath, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullPath
}

// GetImportPath 获取 Go import 路径
func (nm *NamespaceMapper) GetImportPath(namespace string) string {
	return nm.MapNamespaceToPackage(namespace)
}

// GetGoPackageName 获取 Go 包名（最后一个部分）
func (nm *NamespaceMapper) GetGoPackageName(namespace string) string {
	pkg := nm.MapNamespaceToPackage(namespace)
	parts := strings.Split(pkg, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return pkg
}

// IsMainPackage 检查是否是 main 包
func (nm *NamespaceMapper) IsMainPackage(namespace string) bool {
	// 如果命名空间为空或者是根命名空间，可能是 main 包
	return namespace == "" || namespace == "main"
}


