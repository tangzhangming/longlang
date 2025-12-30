package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/tangzhangming/longlang/internal/interpreter"
	"github.com/tangzhangming/longlang/internal/lexer"
	"github.com/tangzhangming/longlang/internal/parser"
)

// main 主函数入口
// 运行方式: longlang.exe run main.long
func main() {
	// 检查命令行参数
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "用法: %s run <文件路径>\n", os.Args[0])
		os.Exit(1)
	}

	// 解析命令
	command := os.Args[1]
	if command != "run" {
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", command)
		fmt.Fprintf(os.Stderr, "用法: %s run <文件路径>\n", os.Args[0])
		os.Exit(1)
	}

	// 获取文件路径
	filename := os.Args[2]
	if filename == "" {
		fmt.Fprintf(os.Stderr, "请指定文件路径\n")
		os.Exit(1)
	}

	// 读取源文件
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取文件错误: %s\n", err)
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
	result := interp.Eval(program)

	// 检查运行时错误
	if result != nil && result.Type() == interpreter.ERROR_OBJ {
		fmt.Fprintf(os.Stderr, "%s\n", result.Inspect())
		os.Exit(1)
	}
}
