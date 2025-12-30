package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/tangzhangming/longlang/internal/interpreter"
	"github.com/tangzhangming/longlang/internal/lexer"
	"github.com/tangzhangming/longlang/internal/parser"
)

// Package 表示一个包
// 包含包的名称、路径、导出的符号等
type Package struct {
	Name      string                    // 包名
	Path      string                    // 包的完整路径
	FilePath  string                    // 包文件路径
	Program   *parser.Program          // 包的 AST
	Exports   map[string]interface{}   // 导出的符号（函数、类等）
	Interpreter *interpreter.Interpreter // 包的解释器环境
}

// PackageManager 包管理器
// 负责包的加载、缓存、路径解析等
type PackageManager struct {
	modulePath string                    // 项目模块路径（从 long.mod 读取）
	rootDir    string                    // 项目根目录
	packages   map[string]*Package       // 包缓存（路径 -> 包）
	importing  map[string]bool           // 正在导入的包（用于检测循环依赖）
}

// NewPackageManager 创建新的包管理器
// 参数:
//   rootDir: 项目根目录
// 返回:
//   包管理器实例
func NewPackageManager(rootDir string) (*PackageManager, error) {
	pm := &PackageManager{
		rootDir:   rootDir,
		packages:  make(map[string]*Package),
		importing: make(map[string]bool),
	}
	
	// 读取 long.mod 文件
	modulePath, err := pm.readModuleFile()
	if err != nil {
		return nil, fmt.Errorf("读取 long.mod 失败: %v", err)
	}
	pm.modulePath = modulePath
	
	return pm, nil
}

// readModuleFile 读取 long.mod 文件
// 返回:
//   module 路径
func (pm *PackageManager) readModuleFile() (string, error) {
	modPath := filepath.Join(pm.rootDir, "long.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		// 如果没有 long.mod 文件，使用默认路径
		return "local", nil
	}
	
	// 解析 module 行
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			module := strings.TrimSpace(strings.TrimPrefix(line, "module"))
			return module, nil
		}
	}
	
	return "local", nil
}

// ResolveImportPath 解析导入路径
// 将导入路径（如 "util.string"）转换为文件路径
// 参数:
//   importPath: 导入路径
// 返回:
//   文件路径
func (pm *PackageManager) ResolveImportPath(importPath string) (string, error) {
	// 处理绝对路径（基于 module 路径）
	if strings.HasPrefix(importPath, pm.modulePath) {
		// 绝对路径：github.com/example/myproject/util.string
		relativePath := strings.TrimPrefix(importPath, pm.modulePath+"/")
		return pm.resolveRelativePath(relativePath)
	}
	
	// 处理相对路径
	if strings.HasPrefix(importPath, "../") || strings.HasPrefix(importPath, "./") {
		return "", fmt.Errorf("相对路径导入暂未实现: %s", importPath)
	}
	
	// 处理项目内路径（基于 module 路径）
	return pm.resolveRelativePath(importPath)
}

// resolveRelativePath 解析相对路径
// 将路径（如 "util.string"）转换为文件路径
func (pm *PackageManager) resolveRelativePath(path string) (string, error) {
	// 将点号替换为路径分隔符
	// util.string -> util/string
	parts := strings.Split(path, ".")
	filePath := filepath.Join(parts...)
	
	// 尝试两种路径格式：
	// 1. util/string.long
	// 2. util/string/string.long
	path1 := filepath.Join(pm.rootDir, filePath+".long")
	path2 := filepath.Join(pm.rootDir, filePath, filepath.Base(filePath)+".long")
	
	if _, err := os.Stat(path1); err == nil {
		return path1, nil
	}
	if _, err := os.Stat(path2); err == nil {
		return path2, nil
	}
	
	return "", fmt.Errorf("找不到包文件: %s (尝试了 %s 和 %s)", path, path1, path2)
}

// LoadPackage 加载包
// 参数:
//   importPath: 导入路径
// 返回:
//   包对象
func (pm *PackageManager) LoadPackage(importPath string) (*Package, error) {
	// 检查缓存
	if pkg, ok := pm.packages[importPath]; ok {
		return pkg, nil
	}
	
	// 检查循环依赖
	if pm.importing[importPath] {
		return nil, fmt.Errorf("检测到循环依赖: %s", importPath)
	}
	
	// 标记为正在导入
	pm.importing[importPath] = true
	defer delete(pm.importing, importPath)
	
	// 解析文件路径
	filePath, err := pm.ResolveImportPath(importPath)
	if err != nil {
		return nil, err
	}
	
	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取包文件失败: %v", err)
	}
	
	// 词法分析
	l := lexer.New(string(data))
	
	// 语法分析
	p := parser.New(l)
	program := p.ParseProgram()
	
	// 检查语法错误
	if len(p.Errors()) != 0 {
		return nil, fmt.Errorf("包语法错误: %v", p.Errors())
	}
	
	// 创建包对象
	pkg := &Package{
		Path:     importPath,
		FilePath: filePath,
		Program:  program,
		Exports:  make(map[string]interface{}),
		Interpreter: interpreter.New(),
	}
	
	// 执行包代码并收集导出符号
	err = pm.executePackage(pkg)
	if err != nil {
		return nil, err
	}
	
	// 缓存包
	pm.packages[importPath] = pkg
	
	return pkg, nil
}

// executePackage 执行包代码并收集导出符号
func (pm *PackageManager) executePackage(pkg *Package) error {
	// 执行包的顶层代码
	result := pkg.Interpreter.Eval(pkg.Program)
	if result != nil && result.Type() == interpreter.ERROR_OBJ {
		return fmt.Errorf("包执行错误: %s", result.Inspect())
	}
	
	// 从环境中收集导出符号
	// 这里需要扩展 Environment 来支持导出符号的收集
	// 暂时先返回成功
	return nil
}

// GetPackage 获取已加载的包
func (pm *PackageManager) GetPackage(importPath string) (*Package, error) {
	pkg, ok := pm.packages[importPath]
	if !ok {
		return pm.LoadPackage(importPath)
	}
	return pkg, nil
}

