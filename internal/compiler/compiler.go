package compiler

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/tangzhangming/longlang/internal/config"
	"github.com/tangzhangming/longlang/internal/lexer"
	"github.com/tangzhangming/longlang/internal/parser"
)

// Compiler 编译器
type Compiler struct {
	symbolTable      *SymbolTable
	analyzer         *Analyzer
	typeMapper       *TypeMapper
	namespaceMapper *NamespaceMapper
	codegen          *CodeGen
	projectGen       *ProjectGenerator
	projectRoot      string
	projectConfig    *config.ProjectConfig
	outputDir        string
}

// NewCompiler 创建新编译器
func NewCompiler() *Compiler {
	symbolTable := NewSymbolTable()
	analyzer := NewAnalyzer(symbolTable)
	typeMapper := NewTypeMapper(symbolTable)
	namespaceMapper := NewNamespaceMapper()
	codegen := NewCodeGen(symbolTable, typeMapper, namespaceMapper)
	projectGen := NewProjectGenerator()

	return &Compiler{
		symbolTable:      symbolTable,
		analyzer:         analyzer,
		typeMapper:       typeMapper,
		namespaceMapper:   namespaceMapper,
		codegen:          codegen,
		projectGen:        projectGen,
	}
}

// SetProjectConfig 设置项目配置
func (c *Compiler) SetProjectConfig(projectRoot string, cfg *config.ProjectConfig) {
	c.projectRoot = projectRoot
	c.projectConfig = cfg
}

// SetOutputDir 设置输出目录
func (c *Compiler) SetOutputDir(outputDir string) {
	c.outputDir = outputDir
}

// Compile 编译程序
func (c *Compiler) Compile(program *parser.Program) error {
	// 1. 分析 AST
	if err := c.analyzer.Analyze(program); err != nil {
		return fmt.Errorf("分析 AST 失败: %w", err)
	}

	// 2. 生成代码
	goCode, err := c.codegen.Generate(program)
	if err != nil {
		return fmt.Errorf("生成代码失败: %w", err)
	}

	// 3. 生成项目结构
	if err := c.projectGen.Generate(c.outputDir, goCode, c.symbolTable, c.namespaceMapper); err != nil {
		return fmt.Errorf("生成项目失败: %w", err)
	}

	return nil
}

// CompileFile 编译文件
func (c *Compiler) CompileFile(filePath string) error {
	// 读取文件
	content, err := readFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 解析（使用文件路径判断是否为标准库）
	lexer := newLexerFromFile(content, filePath)
	parser := newParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		return fmt.Errorf("解析错误: %v", parser.Errors())
	}

	// 编译
	return c.Compile(program)
}

// CompileProject 编译整个项目
func (c *Compiler) CompileProject(entryFile string) error {
	// 获取项目根目录
	absPath, _ := filepath.Abs(entryFile)
	projectRoot := findProjectRoot(filepath.Dir(absPath))

	// 加载项目配置
	cfg, err := config.LoadProjectConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("加载项目配置失败: %w", err)
	}

	c.SetProjectConfig(projectRoot, cfg)

	// 设置输出目录（默认为项目根目录下的 build 目录）
	if c.outputDir == "" {
		c.outputDir = filepath.Join(projectRoot, "build")
	}

	// 解析所有依赖
	resolver := NewDependencyResolver(projectRoot, cfg)
	programs, err := resolver.ResolveDependencies(entryFile)
	if err != nil {
		return fmt.Errorf("解析依赖失败: %w", err)
	}

	// 合并所有程序
	mergedProgram := mergePrograms(programs)

	// 编译合并后的程序
	return c.Compile(mergedProgram)
}

// mergePrograms 合并多个程序
func mergePrograms(programs []*parser.Program) *parser.Program {
	merged := &parser.Program{
		Statements: []parser.Statement{},
	}

	for _, program := range programs {
		if program != nil {
			merged.Statements = append(merged.Statements, program.Statements...)
		}
	}

	return merged
}

// 辅助函数（从 interpreter 复制，避免依赖）
func readFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func newLexer(input string) *lexer.Lexer {
	return lexer.New(input)
}

func newLexerFromFile(input string, filePath string) *lexer.Lexer {
	return lexer.NewFromFile(input, filePath)
}

func newParser(l *lexer.Lexer) *parser.Parser {
	return parser.New(l)
}

func findProjectRoot(startDir string) string {
	dir := startDir
	for {
		projectFile := filepath.Join(dir, "project.toml")
		if _, err := os.Stat(projectFile); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return startDir
		}
		dir = parent
	}
}

