package compiler

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tangzhangming/longlang/internal/config"
	"github.com/tangzhangming/longlang/internal/lexer"
	"github.com/tangzhangming/longlang/internal/parser"
)

// DependencyResolver 依赖解析器
type DependencyResolver struct {
	projectRoot   string
	projectConfig *config.ProjectConfig
	sourcePath    string
	loadedFiles   map[string]bool // 已加载的文件路径
	programs      []*parser.Program // 所有解析的程序
}

// NewDependencyResolver 创建新的依赖解析器
func NewDependencyResolver(projectRoot string, projectConfig *config.ProjectConfig) *DependencyResolver {
	sourcePath := projectConfig.SourcePath
	var fullSourcePath string
	
	if sourcePath == "" || sourcePath == "." {
		// 如果 SourcePath 为空或 ".", 使用项目根目录作为源码路径
		fullSourcePath = projectRoot
	} else {
		fullSourcePath = filepath.Join(projectRoot, sourcePath)
		// 检查路径是否存在，如果不存在，则使用项目根目录
		if _, err := os.Stat(fullSourcePath); os.IsNotExist(err) {
			// 如果指定的源码路径不存在，尝试使用项目根目录
			fullSourcePath = projectRoot
		}
	}

	return &DependencyResolver{
		projectRoot:   projectRoot,
		projectConfig: projectConfig,
		sourcePath:    fullSourcePath,
		loadedFiles:   make(map[string]bool),
		programs:      []*parser.Program{},
	}
}

// ResolveDependencies 解析所有依赖
func (dr *DependencyResolver) ResolveDependencies(entryFile string) ([]*parser.Program, error) {
	// 首先加载入口文件
	entryProgram, err := dr.loadFile(entryFile)
	if err != nil {
		return nil, err
	}
	dr.programs = append(dr.programs, entryProgram)

	// 收集所有 use 语句
	useStatements := dr.collectUseStatements(entryProgram)

	// 递归加载所有依赖
	for _, usePath := range useStatements {
		if !strings.HasPrefix(usePath, "System.") {
			// 非标准库的依赖，尝试查找文件
			if err := dr.loadDependencyFile(usePath); err != nil {
				// 输出错误用于调试（生产环境可以移除或改为日志）
				// fmt.Fprintf(os.Stderr, "警告: 无法加载依赖 %s: %v\n", usePath, err)
			}
		}
	}

	// 递归处理所有已加载程序的依赖（最多递归 10 层，避免无限循环）
	maxDepth := 10
	for depth := 0; depth < maxDepth; depth++ {
		newDeps := false
		for i := 0; i < len(dr.programs); i++ {
			useStatements := dr.collectUseStatements(dr.programs[i])
			for _, usePath := range useStatements {
				if !strings.HasPrefix(usePath, "System.") {
					if dr.loadDependencyFile(usePath) == nil {
						newDeps = true
					}
				}
			}
		}
		if !newDeps {
			break
		}
	}

	return dr.programs, nil
}

// loadFile 加载文件
func (dr *DependencyResolver) loadFile(filePath string) (*parser.Program, error) {
	absPath, _ := filepath.Abs(filePath)
	if dr.loadedFiles[absPath] {
		return nil, nil // 已加载
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("解析错误: %v", p.Errors())
	}

	dr.loadedFiles[absPath] = true
	return program, nil
}

// collectUseStatements 收集 use 语句
func (dr *DependencyResolver) collectUseStatements(program *parser.Program) []string {
	var usePaths []string
	for _, stmt := range program.Statements {
		if useStmt, ok := stmt.(*parser.UseStatement); ok {
			usePaths = append(usePaths, useStmt.Path.Value)
		}
	}
	return usePaths
}

// loadDependency 加载依赖
func (dr *DependencyResolver) loadDependency(usePath string) error {
	// 解析命名空间和类名
	parts := strings.Split(usePath, ".")
	if len(parts) < 2 {
		return fmt.Errorf("无效的 use 路径: %s", usePath)
	}

	className := parts[len(parts)-1]
	namespace := strings.Join(parts[:len(parts)-1], ".")

	// 根据命名空间找到文件
	filePath := dr.findFileByNamespace(namespace, className)
	if filePath == "" {
		return fmt.Errorf("找不到文件: %s", usePath)
	}

	return dr.loadDependencyFile(usePath)
}

// loadDependencyFile 加载依赖文件
func (dr *DependencyResolver) loadDependencyFile(usePath string) error {
	// 解析命名空间和类名
	parts := strings.Split(usePath, ".")
	if len(parts) < 2 {
		return nil
	}

	className := parts[len(parts)-1]
	namespace := strings.Join(parts[:len(parts)-1], ".")

	// 根据命名空间找到文件
	filePath := dr.findFileByNamespace(namespace, className)
	if filePath == "" {
		// 找不到文件，可能是标准库，返回 nil 而不是错误
		// 但如果是非标准库，应该输出警告
		return nil
	}

	// 检查是否已加载
	absPath, _ := filepath.Abs(filePath)
	if dr.loadedFiles[absPath] {
		return nil // 已加载
	}

	program, err := dr.loadFile(filePath)
	if err != nil {
		return err
	}

	if program != nil {
		dr.programs = append(dr.programs, program)
		return nil // 成功加载
	}

	// program 为 nil 但 err 也为 nil，说明文件已加载过
	return nil
}

// findFileByNamespace 根据命名空间找到文件
func (dr *DependencyResolver) findFileByNamespace(namespace, className string) string {
	// 将命名空间转换为路径
	// App.Models -> Models (去掉 App 前缀，因为 App 是根命名空间)
	var namespacePath string
	if dr.projectConfig.RootNamespace != "" && strings.HasPrefix(namespace, dr.projectConfig.RootNamespace+".") {
		// 去掉根命名空间前缀
		relativeNS := strings.TrimPrefix(namespace, dr.projectConfig.RootNamespace+".")
		namespacePath = strings.ReplaceAll(relativeNS, ".", string(filepath.Separator))
	} else if dr.projectConfig.RootNamespace != "" && namespace == dr.projectConfig.RootNamespace {
		// 如果命名空间就是根命名空间，路径为空
		namespacePath = ""
	} else {
		// 直接使用命名空间路径
		namespacePath = strings.ReplaceAll(namespace, ".", string(filepath.Separator))
	}
	
	// 尝试多个可能的路径
	var possiblePaths []string
	if namespacePath == "" {
		possiblePaths = []string{
			filepath.Join(dr.sourcePath, className+".long"),
			filepath.Join(dr.projectRoot, "src", className+".long"),
		}
	} else {
		possiblePaths = []string{
			filepath.Join(dr.sourcePath, namespacePath, className+".long"),
			filepath.Join(dr.projectRoot, "src", namespacePath, className+".long"),
		}
	}

	// 查找文件
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 如果找不到，尝试在命名空间目录下查找所有 .long 文件，找到与类名匹配的
	if namespacePath != "" {
		namespaceDir := filepath.Join(dr.sourcePath, namespacePath)
		if info, err := os.Stat(namespaceDir); err == nil && info.IsDir() {
			files, _ := ioutil.ReadDir(namespaceDir)
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".long") {
					// 读取文件，检查是否包含该类
					filePath := filepath.Join(namespaceDir, file.Name())
					if dr.fileContainsClass(filePath, className) {
						return filePath
					}
				}
			}
		}
	}

	return ""
}

// fileContainsClass 检查文件是否包含指定的类
func (dr *DependencyResolver) fileContainsClass(filePath, className string) bool {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false
	}

	// 简单检查：查找 "class " + className
	contentStr := string(content)
	// 检查多种可能的类声明格式
	return strings.Contains(contentStr, "class "+className) ||
		strings.Contains(contentStr, "public class "+className) ||
		strings.Contains(contentStr, "private class "+className) ||
		strings.Contains(contentStr, "protected class "+className)
}

