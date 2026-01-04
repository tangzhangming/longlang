package interpreter

import (
	"crypto/sha1"
	"crypto/sha256"
)

// registerCryptoBuiltins 注册加密相关内置函数
func registerCryptoBuiltins(env *Environment) {
	// __sha1(data) - 计算 SHA1 哈希
	// data 可以是字符串或字节数组
	env.Set("__sha1", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__sha1 需要1个参数，得到 %d 个", len(args))
		}

		var data []byte
		switch arg := args[0].(type) {
		case *String:
			data = []byte(arg.Value)
		case *Array:
			data = arrayToBytes(arg)
		default:
			return newError("__sha1 参数必须是字符串或字节数组，得到 %s", args[0].Type())
		}

		hash := sha1.Sum(data)
		return bytesToArray(hash[:])
	}})

	// __sha256(data) - 计算 SHA256 哈希
	env.Set("__sha256", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__sha256 需要1个参数，得到 %d 个", len(args))
		}

		var data []byte
		switch arg := args[0].(type) {
		case *String:
			data = []byte(arg.Value)
		case *Array:
			data = arrayToBytes(arg)
		default:
			return newError("__sha256 参数必须是字符串或字节数组，得到 %s", args[0].Type())
		}

		hash := sha256.Sum256(data)
		return bytesToArray(hash[:])
	}})

	// __bytes_xor(bytes1, bytes2) - 两个字节数组异或
	env.Set("__bytes_xor", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytes_xor 需要2个参数，得到 %d 个", len(args))
		}
		arr1, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_xor 第一个参数必须是数组，得到 %s", args[0].Type())
		}
		arr2, ok := args[1].(*Array)
		if !ok {
			return newError("__bytes_xor 第二个参数必须是数组，得到 %s", args[1].Type())
		}

		data1 := arrayToBytes(arr1)
		data2 := arrayToBytes(arr2)

		// 使用较短的长度
		length := len(data1)
		if len(data2) < length {
			length = len(data2)
		}

		result := make([]byte, length)
		for i := 0; i < length; i++ {
			result[i] = data1[i] ^ data2[i]
		}

		return bytesToArray(result)
	}})
}



