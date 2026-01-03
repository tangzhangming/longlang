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
//   - len: 获取数组或字符串的长度
//   - toString: 转换为字符串
//   - toInt/toFloat/toBool: 类型转换
//   - sleep: 延时
//   - exit: 退出程序
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

	// 注册全局 typeof 函数
	// typeof(value) - 获取值的类型名称
	env.Set("typeof", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("typeof 函数需要1个参数，得到 %d 个", len(args))
		}
		return &String{Value: string(args[0].Type())}
	}})

	// 注册全局 __get_field 函数
	// __get_field(object, fieldName) - 获取对象字段值（反射）
	env.Set("__get_field", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__get_field 函数需要2个参数，得到 %d 个", len(args))
		}

		instance, ok := args[0].(*Instance)
		if !ok {
			return newError("__get_field 第一个参数必须是类实例，得到 %s", args[0].Type())
		}

		fieldName, ok := args[1].(*String)
		if !ok {
			return newError("__get_field 第二个参数必须是字符串，得到 %s", args[1].Type())
		}

		// 尝试获取字段值
		if field, exists := instance.Fields[fieldName.Value]; exists {
			return field
		}

		return &Null{}
	}})

	// 注册全局 __set_field 函数
	// __set_field(object, fieldName, value) - 设置对象字段值（反射）
	env.Set("__set_field", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("__set_field 函数需要3个参数，得到 %d 个", len(args))
		}

		instance, ok := args[0].(*Instance)
		if !ok {
			return newError("__set_field 第一个参数必须是类实例，得到 %s", args[0].Type())
		}

		fieldName, ok := args[1].(*String)
		if !ok {
			return newError("__set_field 第二个参数必须是字符串，得到 %s", args[1].Type())
		}

		// 设置字段值
		instance.Fields[fieldName.Value] = args[2]

		return &Null{}
	}})

	// 注册全局 __get_class_name 函数
	// __get_class_name(object) - 获取对象的类名
	env.Set("__get_class_name", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__get_class_name 函数需要1个参数，得到 %d 个", len(args))
		}

		instance, ok := args[0].(*Instance)
		if !ok {
			return newError("__get_class_name 参数必须是类实例，得到 %s", args[0].Type())
		}

		if instance.Class != nil {
			return &String{Value: instance.Class.Name}
		}

		return &String{Value: ""}
	}})

}

// registerCalledClassBuiltin 注册 __called_class 内置函数
// 这个函数在运行时被调用，返回当前静态方法调用的类名
func registerCalledClassBuiltin(env *Environment) {
	// __called_class() - 获取当前静态方法调用时的类名（Late Static Binding）
	// 类似 PHP 的 static::class 或 get_called_class()
	env.Set("__called_class", &Builtin{Fn: func(args ...Object) Object {
		// 从环境中获取 __called_class_name 变量（由静态方法调用时设置）
		calledClass, found := env.Get("__called_class_name")
		if !found {
			return newError("__called_class 只能在静态方法内部使用")
		}
		if str, ok := calledClass.(*String); ok {
			return str
		}
		return &String{Value: ""}
	}})
}

// 全局变量存储
var globalVariables = make(map[string]Object)

// registerGlobalBuiltins 注册全局变量相关的内置函数
func registerGlobalBuiltins(env *Environment) {
	// __set_global(name, value) - 设置全局变量
	env.Set("__set_global", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__set_global 需要2个参数，得到 %d 个", len(args))
		}

		name, ok := args[0].(*String)
		if !ok {
			return newError("__set_global 第一个参数必须是字符串，得到 %s", args[0].Type())
		}

		globalVariables[name.Value] = args[1]
		return &Null{}
	}})

	// __get_global(name) - 获取全局变量
	env.Set("__get_global", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__get_global 需要1个参数，得到 %d 个", len(args))
		}

		name, ok := args[0].(*String)
		if !ok {
			return newError("__get_global 第一个参数必须是字符串，得到 %s", args[0].Type())
		}

		if val, exists := globalVariables[name.Value]; exists {
			return val
		}

		return &Null{}
	}})

	// __has_global(name) - 检查全局变量是否存在
	env.Set("__has_global", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__has_global 需要1个参数，得到 %d 个", len(args))
		}

		name, ok := args[0].(*String)
		if !ok {
			return newError("__has_global 第一个参数必须是字符串，得到 %s", args[0].Type())
		}

		_, exists := globalVariables[name.Value]
		return &Boolean{Value: exists}
	}})
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
