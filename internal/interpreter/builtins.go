package interpreter

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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

	// 注册全局 parseInt 函数
	// parseInt(string) - 将字符串解析为整数
	env.Set("parseInt", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("parseInt 函数需要1个参数，得到 %d 个", len(args))
		}
		str, ok := args[0].(*String)
		if !ok {
			return newError("parseInt 参数必须是字符串，得到 %s", args[0].Type())
		}
		// 去除首尾空白
		s := strings.TrimSpace(str.Value)
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newError("无法将 '%s' 解析为整数", str.Value)
		}
		return &Integer{Value: val}
	}})

	// 注册全局 parseFloat 函数
	// parseFloat(string) - 将字符串解析为浮点数
	env.Set("parseFloat", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("parseFloat 函数需要1个参数，得到 %d 个", len(args))
		}
		str, ok := args[0].(*String)
		if !ok {
			return newError("parseFloat 参数必须是字符串，得到 %s", args[0].Type())
		}
		s := strings.TrimSpace(str.Value)
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newError("无法将 '%s' 解析为浮点数", str.Value)
		}
		return &Float{Value: val}
	}})

	// 注册全局 toString 函数
	// toString(value) - 将任意值转换为字符串
	env.Set("toString", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("toString 函数需要1个参数，得到 %d 个", len(args))
		}
		return &String{Value: args[0].Inspect()}
	}})

	// 注册全局 byteLen 函数
	// byteLen(string) - 获取字符串的字节长度（用于网络协议等场景）
	env.Set("byteLen", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("byteLen 函数需要1个参数，得到 %d 个", len(args))
		}
		str, ok := args[0].(*String)
		if !ok {
			return newError("byteLen 参数必须是字符串，得到 %s", args[0].Type())
		}
		return &Integer{Value: int64(len(str.Value))}
	}})

	// 注册全局 toBytes 函数
	// toBytes(string) - 将字符串转换为 []byte 数组
	env.Set("toBytes", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("toBytes 函数需要1个参数，得到 %d 个", len(args))
		}
		str, ok := args[0].(*String)
		if !ok {
			return newError("toBytes 参数必须是字符串，得到 %s", args[0].Type())
		}
		// 将字符串转换为字节数组
		bytes := []byte(str.Value)
		elements := make([]Object, len(bytes))
		for i, b := range bytes {
			elements[i] = &Integer{Value: int64(b)}
		}
		return &Array{
			Elements:    elements,
			ElementType: "byte",
			IsFixed:     false,
			Capacity:    int64(len(bytes)),
		}
	}})

	// 注册全局 bytesToString 函数
	// bytesToString([]byte) - 将 []byte 数组转换为字符串
	env.Set("bytesToString", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("bytesToString 函数需要1个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("bytesToString 参数必须是数组，得到 %s", args[0].Type())
		}
		// 将整数数组转换为字节数组
		bytes := make([]byte, len(arr.Elements))
		for i, elem := range arr.Elements {
			intVal, ok := elem.(*Integer)
			if !ok {
				return newError("bytesToString 数组元素必须是整数，得到 %s", elem.Type())
			}
			if intVal.Value < 0 || intVal.Value > 255 {
				return newError("bytesToString 数组元素必须在 [0, 255] 范围内，得到 %d", intVal.Value)
			}
			bytes[i] = byte(intVal.Value)
		}
		return &String{Value: string(bytes)}
	}})

	// 注册全局 chr 函数
	// chr(int) - 将 byte/int 转换为单字符字符串
	env.Set("chr", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("chr 函数需要1个参数，得到 %d 个", len(args))
		}
		intVal, ok := args[0].(*Integer)
		if !ok {
			return newError("chr 参数必须是整数，得到 %s", args[0].Type())
		}
		if intVal.Value < 0 || intVal.Value > 255 {
			return newError("chr 值超出范围 [0, 255]：得到 %d", intVal.Value)
		}
		return &String{Value: string(byte(intVal.Value))}
	}})

	// 注册全局 ord 函数
	// ord(string) - 获取字符串第一个字符的 byte 值
	env.Set("ord", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("ord 函数需要1个参数，得到 %d 个", len(args))
		}
		str, ok := args[0].(*String)
		if !ok {
			return newError("ord 参数必须是字符串，得到 %s", args[0].Type())
		}
		if len(str.Value) == 0 {
			return newError("ord 参数不能是空字符串")
		}
		return &Integer{Value: int64(str.Value[0])}
	}})

	// 注册全局 sleep 函数
	// sleep(ms) - 休眠指定毫秒数
	env.Set("sleep", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("sleep 函数需要1个参数，得到 %d 个", len(args))
		}
		ms, ok := args[0].(*Integer)
		if !ok {
			return newError("sleep 参数必须是整数（毫秒），得到 %s", args[0].Type())
		}
		time.Sleep(time.Duration(ms.Value) * time.Millisecond)
		return &Null{}
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
