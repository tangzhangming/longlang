package interpreter

import (
	"fmt"
	"os"
)

// registerBuiltins 注册内置函数
// 在解释器初始化时调用，将内置函数注册到全局环境中
// 参数:
//   env: 全局环境，用于注册内置函数
// 当前支持的内置函数:
//   - fmt.Println: 打印并换行
//   - fmt.Print: 打印不换行
//   - fmt.Printf: 格式化打印
func registerBuiltins(env *Environment) {
	// 注册 fmt 命名空间对象
	env.Set("fmt", &BuiltinObject{
		Fields: map[string]Object{
			// fmt.Println: 打印参数并换行
			// 参数可以是任意数量和类型
			// 例如：fmt.Println("Hello", 123, true)
			"Println": &Builtin{Fn: func(args ...Object) Object {
				values := []interface{}{}
				for _, arg := range args {
					// 将每个参数转换为字符串
					values = append(values, arg.Inspect())
				}
				fmt.Println(values...)
				return &Null{} // 返回 null 表示无返回值
			}},
			// fmt.Print: 打印参数不换行
			// 参数可以是任意数量和类型
			// 例如：fmt.Print("Hello", "World")
			"Print": &Builtin{Fn: func(args ...Object) Object {
				for _, arg := range args {
					fmt.Print(arg.Inspect())
				}
				return &Null{}
			}},
			// fmt.Printf: 格式化打印
			// 第一个参数是格式字符串，后续参数是要格式化的值
			// 例如：fmt.Printf("数字: %d, 字符串: %s", 123, "test")
			"Printf": &Builtin{Fn: func(args ...Object) Object {
				if len(args) == 0 {
					return newError("Printf 至少需要一个参数")
				}
				format := args[0].Inspect()
				values := []interface{}{}
				for i := 1; i < len(args); i++ {
					values = append(values, args[i].Inspect())
				}
				fmt.Printf(format, values...)
				return &Null{}
			}},
		},
	})
}

// BuiltinObject 内置对象（用于命名空间，如 fmt）
// 用于组织相关的内置函数，实现命名空间功能
// 例如：fmt.Println 中的 fmt 就是一个 BuiltinObject
type BuiltinObject struct {
	Fields map[string]Object // 命名空间中的字段（函数或其他对象）
}

func (bo *BuiltinObject) Type() ObjectType { return BUILTIN_OBJ }
func (bo *BuiltinObject) Inspect() string  { return "内置对象" }

// GetField 获取字段
// 用于访问命名空间中的成员（如 fmt.Println）
// 参数:
//   name: 字段名
// 返回:
//   obj: 字段值
//   ok: 是否找到字段
func (bo *BuiltinObject) GetField(name string) (Object, bool) {
	obj, ok := bo.Fields[name]
	return obj, ok
}

// exitError 辅助函数：用于退出程序
// 在发生严重错误时调用，打印错误信息并退出程序
// 参数:
//   code: 退出码
//   msg: 错误消息
func exitError(code int, msg string) {
	fmt.Fprintf(os.Stderr, "错误: %s\n", msg)
	os.Exit(code)
}
