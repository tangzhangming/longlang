package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tangzhangming/longlang/internal/compiler"
	"github.com/tangzhangming/longlang/internal/config"
	"github.com/tangzhangming/longlang/internal/interpreter"
	"github.com/tangzhangming/longlang/internal/lexer"
	"github.com/tangzhangming/longlang/internal/parser"
	"github.com/tangzhangming/longlang/internal/vm"
)

// 版本信息
const (
	Version   = "0.1.0"
	BuildDate = "2026-01-01"
)

// main 主函数入口
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "version", "-v", "--version":
		cmdVersion()
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "用法: %s run <文件路径>\n", os.Args[0])
			os.Exit(1)
		}
		cmdRun(os.Args[2])
	case "vm":
		// 使用虚拟机运行
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "用法: %s vm <文件路径> [--debug]\n", os.Args[0])
			os.Exit(1)
		}
		debug := len(os.Args) >= 4 && os.Args[3] == "--debug"
		cmdVMRun(os.Args[2], debug)
	case "build":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "用法: %s build <文件路径> [-o <输出目录>]\n", os.Args[0])
			os.Exit(1)
		}
		outputDir := ""
		if len(os.Args) >= 5 && os.Args[3] == "-o" {
			outputDir = os.Args[4]
		}
		cmdBuild(os.Args[2], outputDir)
	case "new":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "用法: %s new <项目名称>\n", os.Args[0])
			os.Exit(1)
		}
		cmdNew(os.Args[2])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("LongLang - 一个现代化的编程语言")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  longlang <命令> [参数]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  version     显示版本信息")
	fmt.Println("  run <file>  运行指定的 .long 文件（使用解释器）")
	fmt.Println("  vm <file>   运行指定的 .long 文件（使用字节码虚拟机）")
	fmt.Println("  build <file> [-o <dir>]  编译 .long 文件为 Go 程序")
	fmt.Println("  new <name>  创建一个新项目")
	fmt.Println("  help        显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  longlang version")
	fmt.Println("  longlang run main.long")
	fmt.Println("  longlang vm main.long")
	fmt.Println("  longlang vm main.long --debug")
	fmt.Println("  longlang new myproject")
}

// cmdVersion 显示版本信息
func cmdVersion() {
	fmt.Printf("LongLang version %s\n", Version)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Println()
	fmt.Println("一个现代化的编程语言，支持：")
	fmt.Println("  • 面向对象编程（类、继承、接口）")
	fmt.Println("  • 命名空间系统")
	fmt.Println("  • 静态类型")
	fmt.Println("  • 模块化设计")
	fmt.Println("  • 字节码虚拟机执行")
}

// cmdVMRun 使用虚拟机运行指定的文件
func cmdVMRun(filename string, debug bool) {
	// 检查文件扩展名
	if !strings.HasSuffix(filename, ".long") {
		fmt.Fprintf(os.Stderr, "警告: 文件 %s 不是 .long 文件\n", filename)
	}

	// 读取源文件
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取文件错误: %s\n", err)
		os.Exit(1)
	}

	// 获取项目根目录
	absFilename, _ := filepath.Abs(filename)
	projectRoot := findProjectRoot(filepath.Dir(absFilename))

	// 加载项目配置
	projectConfig, err := config.LoadProjectConfig(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载项目配置错误: %s\n", err)
		os.Exit(1)
	}

	// 词法分析
	l := lexer.New(string(input))

	// 语法分析
	p := parser.New(l)
	program := p.ParseProgram()

	// 检查语法错误
	if len(p.Errors()) != 0 {
		fmt.Fprintf(os.Stderr, "语法错误:\n")
		for _, msg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "\t%s\n", msg)
		}
		os.Exit(1)
	}

	// 创建虚拟机
	fmt.Fprintf(os.Stderr, "[DEBUG] 创建虚拟机...\n")
	virtualMachine := vm.NewVM()
	fmt.Fprintf(os.Stderr, "[DEBUG] 虚拟机创建完成\n")
	virtualMachine.SetDebug(debug)
	virtualMachine.SetProjectConfig(projectRoot, projectConfig)

	// 设置标准库路径（相对于可执行文件）
	exePath, _ := os.Executable()
	stdlibPath := filepath.Join(filepath.Dir(exePath), "stdlib")
	// 如果不存在，尝试当前目录
	if _, statErr := os.Stat(stdlibPath); os.IsNotExist(statErr) {
		stdlibPath = "stdlib"
	}
	virtualMachine.SetStdlibPath(stdlibPath)

	// 创建编译器并关联虚拟机
	comp := vm.NewCompiler()
	comp.SetVM(virtualMachine)

	// 编译为字节码
	fmt.Fprintf(os.Stderr, "[DEBUG] 开始编译...\n")
	bytecode, err := comp.Compile(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "编译错误: %s\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] 编译完成\n")

	// 调试模式下输出字节码
	if debug {
		fmt.Println("=== 字节码 ===")
		fmt.Println(bytecode.Disassemble("main"))
		fmt.Println("=== 执行 ===")
	}

	// 运行
	fmt.Fprintf(os.Stderr, "[DEBUG] 开始运行虚拟机...\n")
	result, err := virtualMachine.Run(bytecode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "运行时错误: %s\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] 虚拟机运行完成\n")

	// 调试模式下输出结果
	if debug && result != nil {
		fmt.Printf("=== 结果 ===\n%s\n", result.Inspect())
	}
}

// cmdRun 运行指定的文件
func cmdRun(filename string) {
	// 检查文件扩展名
	if !strings.HasSuffix(filename, ".long") {
		fmt.Fprintf(os.Stderr, "警告: 文件 %s 不是 .long 文件\n", filename)
	}

	// 读取源文件
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取文件错误: %s\n", err)
		os.Exit(1)
	}

	// 获取项目根目录（向上查找 project.toml）
	absFilename, _ := filepath.Abs(filename)
	projectRoot := findProjectRoot(filepath.Dir(absFilename))

	// 加载项目配置（project.toml）
	projectConfig, err := config.LoadProjectConfig(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载项目配置错误: %s\n", err)
		os.Exit(1)
	}

	// 词法分析：将源代码转换为 token 流
	l := lexer.New(string(input))

	// 语法分析：将 token 流转换为 AST
	p := parser.New(l)
	program := p.ParseProgram()

	// 检查语法错误
	if len(p.Errors()) != 0 {
		fmt.Fprintf(os.Stderr, "语法错误:\n")
		for _, msg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "\t%s\n", msg)
		}
		os.Exit(1)
	}

	// 解释执行：执行 AST
	interp := interpreter.New()

	// 设置项目配置
	interp.SetProjectConfig(projectRoot, projectConfig)

	// 设置标准库路径（相对于可执行文件）
	exePath, _ := os.Executable()
	stdlibPath := filepath.Join(filepath.Dir(exePath), "stdlib")
	// 如果不存在，尝试当前目录
	if _, err := os.Stat(stdlibPath); os.IsNotExist(err) {
		stdlibPath = "stdlib"
	}
	interp.SetStdlibPath(stdlibPath)

	result := interp.Eval(program)

	// 检查运行时错误
	if result != nil && result.Type() == interpreter.ERROR_OBJ {
		fmt.Fprintf(os.Stderr, "%s\n", result.Inspect())
		os.Exit(1)
	}
}

// cmdBuild 编译指定的文件
func cmdBuild(filename string, outputDir string) {
	// 检查文件扩展名
	if !strings.HasSuffix(filename, ".long") {
		fmt.Fprintf(os.Stderr, "警告: 文件 %s 不是 .long 文件\n", filename)
	}

	// 创建编译器
	comp := compiler.NewCompiler()

	// 设置输出目录
	if outputDir == "" {
		absFilename, _ := filepath.Abs(filename)
		projectRoot := findProjectRoot(filepath.Dir(absFilename))
		outputDir = filepath.Join(projectRoot, "build")
	}

	comp.SetOutputDir(outputDir)

	// 编译项目
	if err := comp.CompileProject(filename); err != nil {
		fmt.Fprintf(os.Stderr, "编译错误: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("编译成功！输出目录: %s\n", outputDir)
	fmt.Printf("运行以下命令编译为可执行文件:\n")
	fmt.Printf("  cd %s\n", outputDir)
	fmt.Printf("  go build -o app\n")
}

// cmdNew 创建新项目
func cmdNew(projectName string) {
	// 验证项目名称
	if projectName == "" {
		fmt.Fprintf(os.Stderr, "错误: 项目名称不能为空\n")
		os.Exit(1)
	}

	// 检查目录是否已存在
	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 目录 %s 已存在\n", projectName)
		os.Exit(1)
	}

	fmt.Printf("创建新项目: %s\n", projectName)

	// 创建项目目录结构
	dirs := []string{
		projectName,
		filepath.Join(projectName, "src"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "创建目录失败: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("  创建目录: %s/\n", dir)
	}

	// 生成命名空间名称（首字母大写的驼峰命名）
	namespace := toPascalCase(projectName)

	// 创建 project.toml
	projectToml := fmt.Sprintf(`[project]
name = "%s"
version = "1.0.0"
root_namespace = "%s"
`, projectName, namespace)

	projectTomlPath := filepath.Join(projectName, "project.toml")
	if err := ioutil.WriteFile(projectTomlPath, []byte(projectToml), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "创建 project.toml 失败: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("  创建文件: %s\n", projectTomlPath)

	// 创建入口类 Application.long
	applicationLong := fmt.Sprintf(`namespace App

class Application {
    public static function main() {
        fmt.println("Hello, %s!")
        fmt.println("欢迎使用 LongLang!")
    }
}
`, namespace)

	applicationPath := filepath.Join(projectName, "src", "Application.long")
	if err := ioutil.WriteFile(applicationPath, []byte(applicationLong), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "创建 Application.long 失败: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("  创建文件: %s\n", applicationPath)

	// 创建 README.md
	readme := fmt.Sprintf(`# %s

一个使用 LongLang 编写的项目。

## 运行

`+"```bash"+`
longlang run src/Application.long
`+"```"+`

## 项目结构

`+"```"+`
%s/
├── project.toml      # 项目配置文件
├── src/
│   └── Application.long  # 程序入口
└── README.md
`+"```"+`
`, projectName, projectName)

	readmePath := filepath.Join(projectName, "README.md")
	if err := ioutil.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "创建 README.md 失败: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("  创建文件: %s\n", readmePath)

	fmt.Println()
	fmt.Println("项目创建成功！")
	fmt.Println()
	fmt.Println("开始使用：")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  longlang run src/Application.long")
}

// findProjectRoot 向上查找包含 project.toml 的目录
func findProjectRoot(startDir string) string {
	dir := startDir
	for {
		// 检查当前目录是否有 project.toml
		projectFile := filepath.Join(dir, "project.toml")
		if _, err := os.Stat(projectFile); err == nil {
			return dir
		}

		// 获取父目录
		parent := filepath.Dir(dir)
		if parent == dir {
			// 到达根目录，没有找到 project.toml，返回起始目录
			return startDir
		}
		dir = parent
	}
}

// toPascalCase 将字符串转换为 PascalCase（首字母大写的驼峰命名）
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}

	// 处理分隔符（-、_、空格）
	words := strings.FieldsFunc(s, func(c rune) bool {
		return c == '-' || c == '_' || c == ' '
	})

	var result strings.Builder
	for _, word := range words {
		if len(word) > 0 {
			// 首字母大写
			result.WriteString(strings.ToUpper(string(word[0])))
			if len(word) > 1 {
				result.WriteString(strings.ToLower(word[1:]))
			}
		}
	}

	return result.String()
}
