package compiler

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ProjectGenerator 项目生成器
type ProjectGenerator struct {
}

// NewProjectGenerator 创建新的项目生成器
func NewProjectGenerator() *ProjectGenerator {
	return &ProjectGenerator{}
}

// Generate 生成项目结构
func (pg *ProjectGenerator) Generate(outputDir string, goCode *GoCode, symbolTable *SymbolTable, namespaceMapper *NamespaceMapper) error {
	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 生成 go.mod
	if err := pg.generateGoMod(outputDir); err != nil {
		return err
	}

	// 生成主文件
	mainFile := filepath.Join(outputDir, "main.go")
	if err := pg.generateMainFile(mainFile, goCode); err != nil {
		return err
	}

	return nil
}

// generateGoMod 生成 go.mod 文件
func (pg *ProjectGenerator) generateGoMod(outputDir string) error {
	goModContent := `module longlang-compiled

go 1.21
`
	goModPath := filepath.Join(outputDir, "go.mod")
	return ioutil.WriteFile(goModPath, []byte(goModContent), 0644)
}

// generateMainFile 生成主文件
func (pg *ProjectGenerator) generateMainFile(filePath string, goCode *GoCode) error {
	var result strings.Builder

	// 包声明
	result.WriteString(fmt.Sprintf("package %s\n\n", goCode.PackageName))

	// Imports
	if len(goCode.Imports) > 0 {
		result.WriteString("import (\n")
		for _, imp := range goCode.Imports {
			result.WriteString(fmt.Sprintf("    \"%s\"\n", imp))
		}
		result.WriteString(")\n\n")
	}

	// 运行时辅助函数
	if goCode.RuntimeHelpers != "" {
		result.WriteString(goCode.RuntimeHelpers)
		result.WriteString("\n")
	}

	// 类型定义
	for _, typ := range goCode.Types {
		result.WriteString(typ)
		result.WriteString("\n")
	}

	// 函数定义
	for _, fn := range goCode.Functions {
		result.WriteString(fn)
		result.WriteString("\n")
	}

	// Main 函数
	if goCode.MainCode != "" {
		result.WriteString(goCode.MainCode)
	}

	return ioutil.WriteFile(filePath, []byte(result.String()), 0644)
}

