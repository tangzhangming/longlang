package interpreter

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// registerConsoleBuiltins 注册控制台相关的内置函数
func registerConsoleBuiltins(env *Environment) {
	// __console_write(args: []any) - 输出不换行
	env.Set("__console_write", &Builtin{Fn: func(args ...Object) Object {
		if len(args) == 0 {
			return &Null{}
		}
		
		// 参数是一个数组
		arr, ok := args[0].(*Array)
		if !ok {
			// 如果不是数组，直接输出
			fmt.Print(args[0].Inspect())
			return &Null{}
		}
		
		// 输出数组中的每个元素
		for i, elem := range arr.Elements {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(elem.Inspect())
		}
		return &Null{}
	}})

	// __console_write_line(args: []any) - 输出并换行
	env.Set("__console_write_line", &Builtin{Fn: func(args ...Object) Object {
		if len(args) == 0 {
			fmt.Println()
			return &Null{}
		}
		
		// 参数是一个数组
		arr, ok := args[0].(*Array)
		if !ok {
			// 如果不是数组，直接输出
			fmt.Println(args[0].Inspect())
			return &Null{}
		}
		
		// 如果数组为空，只输出换行
		if len(arr.Elements) == 0 {
			fmt.Println()
			return &Null{}
		}
		
		// 输出数组中的每个元素
		for i, elem := range arr.Elements {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(elem.Inspect())
		}
		fmt.Println()
		return &Null{}
	}})

	// __console_write_empty_line() - 输出空行
	env.Set("__console_write_empty_line", &Builtin{Fn: func(args ...Object) Object {
		fmt.Println()
		return &Null{}
	}})

	// __console_read_line() string - 读取一行输入
	env.Set("__console_read_line", &Builtin{Fn: func(args ...Object) Object {
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return &String{Value: ""}
		}
		// 移除换行符
		line = strings.TrimRight(line, "\r\n")
		return &String{Value: line}
	}})

	// __console_read() int - 读取单个字符
	env.Set("__console_read", &Builtin{Fn: func(args ...Object) Object {
		reader := bufio.NewReader(os.Stdin)
		char, err := reader.ReadByte()
		if err != nil {
			return &Integer{Value: -1}
		}
		return &Integer{Value: int64(char)}
	}})

	// __console_clear() - 清屏
	env.Set("__console_clear", &Builtin{Fn: func(args ...Object) Object {
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", "cls")
		} else {
			cmd = exec.Command("clear")
		}
		cmd.Stdout = os.Stdout
		cmd.Run()
		return &Null{}
	}})

	// __console_set_cursor_position(left: int, top: int) - 设置光标位置
	env.Set("__console_set_cursor_position", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__console_set_cursor_position 需要2个参数")
		}
		left := getIntArg(args[0])
		top := getIntArg(args[1])
		// ANSI 转义序列，光标位置从 1 开始
		fmt.Printf("\033[%d;%dH", top+1, left+1)
		return &Null{}
	}})

	// __console_get_window_width() int - 获取窗口宽度
	env.Set("__console_get_window_width", &Builtin{Fn: func(args ...Object) Object {
		width := getTerminalWidth()
		return &Integer{Value: int64(width)}
	}})

	// __console_get_window_height() int - 获取窗口高度
	env.Set("__console_get_window_height", &Builtin{Fn: func(args ...Object) Object {
		height := getTerminalHeight()
		return &Integer{Value: int64(height)}
	}})

	// __console_set_title(title: string) - 设置窗口标题
	env.Set("__console_set_title", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__console_set_title 需要1个参数")
		}
		title, ok := args[0].(*String)
		if !ok {
			return newError("__console_set_title 参数必须是字符串")
		}
		// ANSI 转义序列设置标题
		fmt.Printf("\033]0;%s\007", title.Value)
		return &Null{}
	}})

	// __console_beep() - 发出蜂鸣声
	env.Set("__console_beep", &Builtin{Fn: func(args ...Object) Object {
		fmt.Print("\007") // ASCII BEL 字符
		return &Null{}
	}})
}

// getTerminalWidth 获取终端宽度
func getTerminalWidth() int {
	// 默认宽度
	defaultWidth := 80
	
	if runtime.GOOS == "windows" {
		// Windows 下使用 mode con 命令
		cmd := exec.Command("cmd", "/c", "mode", "con")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Columns") || strings.Contains(line, "列") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						var width int
						fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &width)
						if width > 0 {
							return width
						}
					}
				}
			}
		}
	} else {
		// Unix 系统使用 tput
		cmd := exec.Command("tput", "cols")
		output, err := cmd.Output()
		if err == nil {
			var width int
			fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &width)
			if width > 0 {
				return width
			}
		}
	}
	
	return defaultWidth
}

// getTerminalHeight 获取终端高度
func getTerminalHeight() int {
	// 默认高度
	defaultHeight := 24
	
	if runtime.GOOS == "windows" {
		// Windows 下使用 mode con 命令
		cmd := exec.Command("cmd", "/c", "mode", "con")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Lines") || strings.Contains(line, "行") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						var height int
						fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &height)
						if height > 0 {
							return height
						}
					}
				}
			}
		}
	} else {
		// Unix 系统使用 tput
		cmd := exec.Command("tput", "lines")
		output, err := cmd.Output()
		if err == nil {
			var height int
			fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &height)
			if height > 0 {
				return height
			}
		}
	}
	
	return defaultHeight
}

