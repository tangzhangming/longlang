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
//   - fmt.println: 打印并换行
//   - fmt.print: 打印不换行
//   - fmt.printf: 格式化打印
//   - len: 获取数组或字符串的长度
func registerBuiltins(env *Environment) {
	// 注册全局 len 函数
	env.Set("len", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("len 函数需要1个参数，得到 %d 个", len(args))
		}
		switch arg := args[0].(type) {
		case *Array:
			return &Integer{Value: int64(len(arg.Elements))}
		case *String:
			return &Integer{Value: int64(len([]rune(arg.Value)))}
		case *Map:
			return &Integer{Value: int64(arg.Size())}
		default:
			return newError("len 函数不支持类型 %s", args[0].Type())
		}
	}})

	// 注册全局 isset 函数
	// isset(map, key) - 检查 Map 是否包含指定键
	// isset(array, index) - 检查数组索引是否有效
	env.Set("isset", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("isset 函数需要2个参数，得到 %d 个", len(args))
		}
		switch container := args[0].(type) {
		case *Map:
			// Map: isset(map, "key")
			keyStr, ok := args[1].(*String)
			if !ok {
				return newError("isset 检查 Map 时，第二个参数必须是字符串，得到 %s", args[1].Type())
			}
			return &Boolean{Value: container.Has(keyStr.Value)}
		case *Array:
			// Array: isset(array, index)
			indexInt, ok := args[1].(*Integer)
			if !ok {
				return newError("isset 检查数组时，第二个参数必须是整数，得到 %s", args[1].Type())
			}
			idx := indexInt.Value
			// 支持负数索引
			if idx < 0 {
				idx = int64(len(container.Elements)) + idx
			}
			return &Boolean{Value: idx >= 0 && idx < int64(len(container.Elements))}
		default:
			return newError("isset 函数的第一个参数必须是 Map 或数组，得到 %s", args[0].Type())
		}
	}})
	// 注册 fmt 命名空间对象
	env.Set("fmt", &BuiltinObject{
		Name: "fmt",
		Fields: map[string]Object{
			// fmt.println: 打印参数并换行
			// 参数可以是任意数量和类型
			// 例如：fmt.println("Hello", 123, true)
			"println": &Builtin{Fn: func(args ...Object) Object {
				values := []interface{}{}
				for _, arg := range args {
					// 将每个参数转换为字符串
					values = append(values, arg.Inspect())
				}
				fmt.Println(values...)
				return &Null{} // 返回 null 表示无返回值
			}},
			// fmt.print: 打印参数不换行
			// 参数可以是任意数量和类型
			// 例如：fmt.print("Hello", "World")
			"print": &Builtin{Fn: func(args ...Object) Object {
				for _, arg := range args {
					fmt.Print(arg.Inspect())
				}
				return &Null{}
			}},
			// fmt.printf: 格式化打印
			// 第一个参数是格式字符串，后续参数是要格式化的值
			// 例如：fmt.printf("数字: %d, 字符串: %s", 123, "test")
			"printf": &Builtin{Fn: func(args ...Object) Object {
				if len(args) == 0 {
					return newError("printf 至少需要一个参数")
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
	Name   string            // 命名空间名称
	Fields map[string]Object // 命名空间中的字段（函数或其他对象）
}

// NewBuiltinObject 创建内置对象
func NewBuiltinObject(name string) *BuiltinObject {
	return &BuiltinObject{
		Name:   name,
		Fields: make(map[string]Object),
	}
}

func (bo *BuiltinObject) Type() ObjectType { return BUILTIN_OBJ }
func (bo *BuiltinObject) Inspect() string  { return "builtin " + bo.Name }

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

// SetField 设置字段
func (bo *BuiltinObject) SetField(name string, obj Object) {
	bo.Fields[name] = obj
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
