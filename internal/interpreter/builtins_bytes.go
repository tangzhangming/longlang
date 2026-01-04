package interpreter

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"strings"
)

// registerBytesBuiltins 注册字节操作内置函数
func registerBytesBuiltins(env *Environment) {
	// ===== 字节数组创建和转换 =====

	// __bytes_new(size) - 创建指定大小的字节数组（初始化为 0）
	env.Set("__bytes_new", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytes_new 需要1个参数，得到 %d 个", len(args))
		}
		sizeInt, ok := args[0].(*Integer)
		if !ok {
			return newError("__bytes_new 参数必须是整数，得到 %s", args[0].Type())
		}
		if sizeInt.Value < 0 {
			return newError("__bytes_new 大小不能为负数")
		}

		elements := make([]Object, sizeInt.Value)
		for i := range elements {
			elements[i] = &Integer{Value: 0}
		}
		return &Array{Elements: elements}
	}})

	// __bytes_from_string(str) - 从字符串创建字节数组
	env.Set("__bytes_from_string", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytes_from_string 需要1个参数，得到 %d 个", len(args))
		}
		strObj, ok := args[0].(*String)
		if !ok {
			return newError("__bytes_from_string 参数必须是字符串，得到 %s", args[0].Type())
		}

		data := []byte(strObj.Value)
		elements := make([]Object, len(data))
		for i, b := range data {
			elements[i] = &Integer{Value: int64(b)}
		}
		return &Array{Elements: elements}
	}})

	// __bytes_to_string(bytes) - 字节数组转字符串
	env.Set("__bytes_to_string", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytes_to_string 需要1个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_to_string 参数必须是数组，得到 %s", args[0].Type())
		}

		data := make([]byte, len(arr.Elements))
		for i, elem := range arr.Elements {
			byteInt, ok := elem.(*Integer)
			if !ok {
				return newError("__bytes_to_string 数组元素必须是整数，得到 %s", elem.Type())
			}
			data[i] = byte(byteInt.Value)
		}
		return &String{Value: string(data)}
	}})

	// __bytes_to_hex(bytes) - 字节数组转十六进制字符串
	env.Set("__bytes_to_hex", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytes_to_hex 需要1个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_to_hex 参数必须是数组，得到 %s", args[0].Type())
		}

		var sb strings.Builder
		for _, elem := range arr.Elements {
			byteInt, ok := elem.(*Integer)
			if !ok {
				return newError("__bytes_to_hex 数组元素必须是整数，得到 %s", elem.Type())
			}
			sb.WriteString(strconv.FormatInt(byteInt.Value&0xFF, 16))
		}
		return &String{Value: sb.String()}
	}})

	// __bytes_from_hex(hexStr) - 十六进制字符串转字节数组
	env.Set("__bytes_from_hex", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytes_from_hex 需要1个参数，得到 %d 个", len(args))
		}
		hexStr, ok := args[0].(*String)
		if !ok {
			return newError("__bytes_from_hex 参数必须是字符串，得到 %s", args[0].Type())
		}

		hex := hexStr.Value
		if len(hex)%2 != 0 {
			hex = "0" + hex
		}

		elements := make([]Object, len(hex)/2)
		for i := 0; i < len(hex); i += 2 {
			b, err := strconv.ParseInt(hex[i:i+2], 16, 64)
			if err != nil {
				return newError("__bytes_from_hex 无效的十六进制字符: %s", hex[i:i+2])
			}
			elements[i/2] = &Integer{Value: b}
		}
		return &Array{Elements: elements}
	}})

	// ===== 字节数组操作 =====

	// __bytes_concat(bytes1, bytes2) - 连接两个字节数组
	env.Set("__bytes_concat", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytes_concat 需要2个参数，得到 %d 个", len(args))
		}
		arr1, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_concat 第一个参数必须是数组，得到 %s", args[0].Type())
		}
		arr2, ok := args[1].(*Array)
		if !ok {
			return newError("__bytes_concat 第二个参数必须是数组，得到 %s", args[1].Type())
		}

		elements := make([]Object, len(arr1.Elements)+len(arr2.Elements))
		copy(elements, arr1.Elements)
		copy(elements[len(arr1.Elements):], arr2.Elements)
		return &Array{Elements: elements}
	}})

	// __bytes_slice(bytes, start, end) - 切片字节数组
	env.Set("__bytes_slice", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("__bytes_slice 需要3个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_slice 第一个参数必须是数组，得到 %s", args[0].Type())
		}
		startInt, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytes_slice 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		endInt, ok := args[2].(*Integer)
		if !ok {
			return newError("__bytes_slice 第三个参数必须是整数，得到 %s", args[2].Type())
		}

		start := int(startInt.Value)
		end := int(endInt.Value)
		if start < 0 {
			start = 0
		}
		if end > len(arr.Elements) {
			end = len(arr.Elements)
		}
		if start > end {
			start = end
		}

		elements := make([]Object, end-start)
		copy(elements, arr.Elements[start:end])
		return &Array{Elements: elements}
	}})

	// __bytes_copy(src, srcOffset, dst, dstOffset, length) - 复制字节
	env.Set("__bytes_copy", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 5 {
			return newError("__bytes_copy 需要5个参数，得到 %d 个", len(args))
		}
		src, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_copy src 必须是数组，得到 %s", args[0].Type())
		}
		srcOffset, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytes_copy srcOffset 必须是整数，得到 %s", args[1].Type())
		}
		dst, ok := args[2].(*Array)
		if !ok {
			return newError("__bytes_copy dst 必须是数组，得到 %s", args[2].Type())
		}
		dstOffset, ok := args[3].(*Integer)
		if !ok {
			return newError("__bytes_copy dstOffset 必须是整数，得到 %s", args[3].Type())
		}
		length, ok := args[4].(*Integer)
		if !ok {
			return newError("__bytes_copy length 必须是整数，得到 %s", args[4].Type())
		}

		srcOff := int(srcOffset.Value)
		dstOff := int(dstOffset.Value)
		len_ := int(length.Value)

		for i := 0; i < len_; i++ {
			if srcOff+i < len(src.Elements) && dstOff+i < len(dst.Elements) {
				dst.Elements[dstOff+i] = src.Elements[srcOff+i]
			}
		}
		return &Integer{Value: int64(len_)}
	}})

	// __bytes_equals(bytes1, bytes2) - 比较两个字节数组是否相等
	env.Set("__bytes_equals", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytes_equals 需要2个参数，得到 %d 个", len(args))
		}
		arr1, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_equals 第一个参数必须是数组，得到 %s", args[0].Type())
		}
		arr2, ok := args[1].(*Array)
		if !ok {
			return newError("__bytes_equals 第二个参数必须是数组，得到 %s", args[1].Type())
		}

		if len(arr1.Elements) != len(arr2.Elements) {
			return &Boolean{Value: false}
		}

		for i := range arr1.Elements {
			b1, ok1 := arr1.Elements[i].(*Integer)
			b2, ok2 := arr2.Elements[i].(*Integer)
			if !ok1 || !ok2 || b1.Value != b2.Value {
				return &Boolean{Value: false}
			}
		}
		return &Boolean{Value: true}
	}})

	// __bytes_index_of(bytes, sub) - 查找子序列位置
	env.Set("__bytes_index_of", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytes_index_of 需要2个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_index_of 第一个参数必须是数组，得到 %s", args[0].Type())
		}
		sub, ok := args[1].(*Array)
		if !ok {
			return newError("__bytes_index_of 第二个参数必须是数组，得到 %s", args[1].Type())
		}

		data := arrayToBytes(arr)
		subData := arrayToBytes(sub)
		idx := bytes.Index(data, subData)
		return &Integer{Value: int64(idx)}
	}})

	// ===== 整数编码/解码（大端序）=====

	// __bytes_write_int8(value) - 写入 int8 为字节数组
	env.Set("__bytes_write_int8", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytes_write_int8 需要1个参数，得到 %d 个", len(args))
		}
		val, ok := args[0].(*Integer)
		if !ok {
			return newError("__bytes_write_int8 参数必须是整数，得到 %s", args[0].Type())
		}
		return &Array{Elements: []Object{&Integer{Value: val.Value & 0xFF}}}
	}})

	// __bytes_read_int8(bytes, offset) - 从字节数组读取 int8
	env.Set("__bytes_read_int8", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytes_read_int8 需要2个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_read_int8 第一个参数必须是数组，得到 %s", args[0].Type())
		}
		offset, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytes_read_int8 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		off := int(offset.Value)
		if off < 0 || off >= len(arr.Elements) {
			return newError("__bytes_read_int8 偏移量越界")
		}
		b, ok := arr.Elements[off].(*Integer)
		if !ok {
			return newError("__bytes_read_int8 元素必须是整数")
		}
		// 转换为有符号
		val := int8(b.Value)
		return &Integer{Value: int64(val)}
	}})

	// __bytes_write_int16_be(value) - 写入 int16（大端序）
	env.Set("__bytes_write_int16_be", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytes_write_int16_be 需要1个参数，得到 %d 个", len(args))
		}
		val, ok := args[0].(*Integer)
		if !ok {
			return newError("__bytes_write_int16_be 参数必须是整数，得到 %s", args[0].Type())
		}
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(val.Value))
		return bytesToArray(buf)
	}})

	// __bytes_read_int16_be(bytes, offset) - 读取 int16（大端序）
	env.Set("__bytes_read_int16_be", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytes_read_int16_be 需要2个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_read_int16_be 第一个参数必须是数组，得到 %s", args[0].Type())
		}
		offset, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytes_read_int16_be 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		off := int(offset.Value)
		if off < 0 || off+2 > len(arr.Elements) {
			return newError("__bytes_read_int16_be 偏移量越界")
		}
		data := arrayToBytes(arr)[off : off+2]
		val := int16(binary.BigEndian.Uint16(data))
		return &Integer{Value: int64(val)}
	}})

	// __bytes_write_int32_be(value) - 写入 int32（大端序）
	env.Set("__bytes_write_int32_be", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytes_write_int32_be 需要1个参数，得到 %d 个", len(args))
		}
		val, ok := args[0].(*Integer)
		if !ok {
			return newError("__bytes_write_int32_be 参数必须是整数，得到 %s", args[0].Type())
		}
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(val.Value))
		return bytesToArray(buf)
	}})

	// __bytes_read_int32_be(bytes, offset) - 读取 int32（大端序）
	env.Set("__bytes_read_int32_be", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytes_read_int32_be 需要2个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_read_int32_be 第一个参数必须是数组，得到 %s", args[0].Type())
		}
		offset, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytes_read_int32_be 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		off := int(offset.Value)
		if off < 0 || off+4 > len(arr.Elements) {
			return newError("__bytes_read_int32_be 偏移量越界")
		}
		data := arrayToBytes(arr)[off : off+4]
		val := int32(binary.BigEndian.Uint32(data))
		return &Integer{Value: int64(val)}
	}})

	// __bytes_write_int64_be(value) - 写入 int64（大端序）
	env.Set("__bytes_write_int64_be", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytes_write_int64_be 需要1个参数，得到 %d 个", len(args))
		}
		val, ok := args[0].(*Integer)
		if !ok {
			return newError("__bytes_write_int64_be 参数必须是整数，得到 %s", args[0].Type())
		}
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(val.Value))
		return bytesToArray(buf)
	}})

	// __bytes_read_int64_be(bytes, offset) - 读取 int64（大端序）
	env.Set("__bytes_read_int64_be", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytes_read_int64_be 需要2个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_read_int64_be 第一个参数必须是数组，得到 %s", args[0].Type())
		}
		offset, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytes_read_int64_be 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		off := int(offset.Value)
		if off < 0 || off+8 > len(arr.Elements) {
			return newError("__bytes_read_int64_be 偏移量越界")
		}
		data := arrayToBytes(arr)[off : off+8]
		val := int64(binary.BigEndian.Uint64(data))
		return &Integer{Value: val}
	}})

	// ===== 整数编码/解码（小端序）=====

	// __bytes_write_int16_le(value) - 写入 int16（小端序）
	env.Set("__bytes_write_int16_le", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytes_write_int16_le 需要1个参数，得到 %d 个", len(args))
		}
		val, ok := args[0].(*Integer)
		if !ok {
			return newError("__bytes_write_int16_le 参数必须是整数，得到 %s", args[0].Type())
		}
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(val.Value))
		return bytesToArray(buf)
	}})

	// __bytes_read_int16_le(bytes, offset) - 读取 int16（小端序）
	env.Set("__bytes_read_int16_le", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytes_read_int16_le 需要2个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_read_int16_le 第一个参数必须是数组，得到 %s", args[0].Type())
		}
		offset, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytes_read_int16_le 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		off := int(offset.Value)
		if off < 0 || off+2 > len(arr.Elements) {
			return newError("__bytes_read_int16_le 偏移量越界")
		}
		data := arrayToBytes(arr)[off : off+2]
		val := int16(binary.LittleEndian.Uint16(data))
		return &Integer{Value: int64(val)}
	}})

	// __bytes_write_int32_le(value) - 写入 int32（小端序）
	env.Set("__bytes_write_int32_le", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytes_write_int32_le 需要1个参数，得到 %d 个", len(args))
		}
		val, ok := args[0].(*Integer)
		if !ok {
			return newError("__bytes_write_int32_le 参数必须是整数，得到 %s", args[0].Type())
		}
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(val.Value))
		return bytesToArray(buf)
	}})

	// __bytes_read_int32_le(bytes, offset) - 读取 int32（小端序）
	env.Set("__bytes_read_int32_le", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytes_read_int32_le 需要2个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_read_int32_le 第一个参数必须是数组，得到 %s", args[0].Type())
		}
		offset, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytes_read_int32_le 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		off := int(offset.Value)
		if off < 0 || off+4 > len(arr.Elements) {
			return newError("__bytes_read_int32_le 偏移量越界")
		}
		data := arrayToBytes(arr)[off : off+4]
		val := int32(binary.LittleEndian.Uint32(data))
		return &Integer{Value: int64(val)}
	}})

	// __bytes_write_int64_le(value) - 写入 int64（小端序）
	env.Set("__bytes_write_int64_le", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytes_write_int64_le 需要1个参数，得到 %d 个", len(args))
		}
		val, ok := args[0].(*Integer)
		if !ok {
			return newError("__bytes_write_int64_le 参数必须是整数，得到 %s", args[0].Type())
		}
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(val.Value))
		return bytesToArray(buf)
	}})

	// __bytes_read_int64_le(bytes, offset) - 读取 int64（小端序）
	env.Set("__bytes_read_int64_le", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytes_read_int64_le 需要2个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("__bytes_read_int64_le 第一个参数必须是数组，得到 %s", args[0].Type())
		}
		offset, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytes_read_int64_le 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		off := int(offset.Value)
		if off < 0 || off+8 > len(arr.Elements) {
			return newError("__bytes_read_int64_le 偏移量越界")
		}
		data := arrayToBytes(arr)[off : off+8]
		val := int64(binary.LittleEndian.Uint64(data))
		return &Integer{Value: val}
	}})

	// ===== ByteBuffer 操作 =====

	// __bytebuffer_new(capacity) - 创建 ByteBuffer
	env.Set("__bytebuffer_new", &Builtin{Fn: func(args ...Object) Object {
		capacity := 64
		if len(args) >= 1 {
			if capInt, ok := args[0].(*Integer); ok {
				capacity = int(capInt.Value)
			}
		}
		return &ByteBuffer{
			Buffer:   make([]byte, 0, capacity),
			Position: 0,
		}
	}})

	// __bytebuffer_from_bytes(bytes) - 从字节数组创建 ByteBuffer
	env.Set("__bytebuffer_from_bytes", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_from_bytes 需要1个参数，得到 %d 个", len(args))
		}
		arr, ok := args[0].(*Array)
		if !ok {
			return newError("__bytebuffer_from_bytes 参数必须是数组，得到 %s", args[0].Type())
		}
		data := arrayToBytes(arr)
		return &ByteBuffer{
			Buffer:   data,
			Position: 0,
		}
	}})

	// __bytebuffer_write_byte(buf, byte) - 写入单个字节
	env.Set("__bytebuffer_write_byte", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytebuffer_write_byte 需要2个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_write_byte 第一个参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		byteInt, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytebuffer_write_byte 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		buf.Buffer = append(buf.Buffer, byte(byteInt.Value))
		return &Null{}
	}})

	// __bytebuffer_write_bytes(buf, bytes) - 写入字节数组
	env.Set("__bytebuffer_write_bytes", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytebuffer_write_bytes 需要2个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_write_bytes 第一个参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		arr, ok := args[1].(*Array)
		if !ok {
			return newError("__bytebuffer_write_bytes 第二个参数必须是数组，得到 %s", args[1].Type())
		}
		data := arrayToBytes(arr)
		buf.Buffer = append(buf.Buffer, data...)
		return &Null{}
	}})

	// __bytebuffer_write_string(buf, str) - 写入字符串
	env.Set("__bytebuffer_write_string", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytebuffer_write_string 需要2个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_write_string 第一个参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		str, ok := args[1].(*String)
		if !ok {
			return newError("__bytebuffer_write_string 第二个参数必须是字符串，得到 %s", args[1].Type())
		}
		buf.Buffer = append(buf.Buffer, []byte(str.Value)...)
		return &Null{}
	}})

	// __bytebuffer_write_int16_be(buf, value) - 写入 int16（大端序）
	env.Set("__bytebuffer_write_int16_be", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytebuffer_write_int16_be 需要2个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_write_int16_be 第一个参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		val, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytebuffer_write_int16_be 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(val.Value))
		buf.Buffer = append(buf.Buffer, b...)
		return &Null{}
	}})

	// __bytebuffer_write_int32_be(buf, value) - 写入 int32（大端序）
	env.Set("__bytebuffer_write_int32_be", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytebuffer_write_int32_be 需要2个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_write_int32_be 第一个参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		val, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytebuffer_write_int32_be 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(val.Value))
		buf.Buffer = append(buf.Buffer, b...)
		return &Null{}
	}})

	// __bytebuffer_write_int64_be(buf, value) - 写入 int64（大端序）
	env.Set("__bytebuffer_write_int64_be", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytebuffer_write_int64_be 需要2个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_write_int64_be 第一个参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		val, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytebuffer_write_int64_be 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(val.Value))
		buf.Buffer = append(buf.Buffer, b...)
		return &Null{}
	}})

	// __bytebuffer_read_byte(buf) - 读取单个字节
	env.Set("__bytebuffer_read_byte", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_read_byte 需要1个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_read_byte 参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		if buf.Position >= len(buf.Buffer) {
			return newError("__bytebuffer_read_byte 缓冲区已读完")
		}
		b := buf.Buffer[buf.Position]
		buf.Position++
		return &Integer{Value: int64(b)}
	}})

	// __bytebuffer_read_bytes(buf, count) - 读取指定数量字节
	env.Set("__bytebuffer_read_bytes", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytebuffer_read_bytes 需要2个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_read_bytes 第一个参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		count, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytebuffer_read_bytes 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		n := int(count.Value)
		if buf.Position+n > len(buf.Buffer) {
			n = len(buf.Buffer) - buf.Position
		}
		data := buf.Buffer[buf.Position : buf.Position+n]
		buf.Position += n
		return bytesToArray(data)
	}})

	// __bytebuffer_read_string(buf, count) - 读取指定长度字符串
	env.Set("__bytebuffer_read_string", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytebuffer_read_string 需要2个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_read_string 第一个参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		count, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytebuffer_read_string 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		n := int(count.Value)
		if buf.Position+n > len(buf.Buffer) {
			n = len(buf.Buffer) - buf.Position
		}
		data := buf.Buffer[buf.Position : buf.Position+n]
		buf.Position += n
		return &String{Value: string(data)}
	}})

	// __bytebuffer_read_line(buf) - 读取一行（到 \n 或 \r\n）
	env.Set("__bytebuffer_read_line", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_read_line 需要1个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_read_line 参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}

		start := buf.Position
		for i := buf.Position; i < len(buf.Buffer); i++ {
			if buf.Buffer[i] == '\n' {
				line := string(buf.Buffer[start:i])
				buf.Position = i + 1
				// 移除可能的 \r
				line = strings.TrimSuffix(line, "\r")
				return &String{Value: line}
			}
		}
		// 没有找到换行符，返回剩余内容
		line := string(buf.Buffer[start:])
		buf.Position = len(buf.Buffer)
		return &String{Value: line}
	}})

	// __bytebuffer_read_int16_be(buf) - 读取 int16（大端序）
	env.Set("__bytebuffer_read_int16_be", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_read_int16_be 需要1个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_read_int16_be 参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		if buf.Position+2 > len(buf.Buffer) {
			return newError("__bytebuffer_read_int16_be 缓冲区不足")
		}
		val := int16(binary.BigEndian.Uint16(buf.Buffer[buf.Position:]))
		buf.Position += 2
		return &Integer{Value: int64(val)}
	}})

	// __bytebuffer_read_int32_be(buf) - 读取 int32（大端序）
	env.Set("__bytebuffer_read_int32_be", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_read_int32_be 需要1个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_read_int32_be 参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		if buf.Position+4 > len(buf.Buffer) {
			return newError("__bytebuffer_read_int32_be 缓冲区不足")
		}
		val := int32(binary.BigEndian.Uint32(buf.Buffer[buf.Position:]))
		buf.Position += 4
		return &Integer{Value: int64(val)}
	}})

	// __bytebuffer_read_int64_be(buf) - 读取 int64（大端序）
	env.Set("__bytebuffer_read_int64_be", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_read_int64_be 需要1个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_read_int64_be 参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		if buf.Position+8 > len(buf.Buffer) {
			return newError("__bytebuffer_read_int64_be 缓冲区不足")
		}
		val := int64(binary.BigEndian.Uint64(buf.Buffer[buf.Position:]))
		buf.Position += 8
		return &Integer{Value: val}
	}})

	// __bytebuffer_get_position(buf) - 获取当前位置
	env.Set("__bytebuffer_get_position", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_get_position 需要1个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_get_position 参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		return &Integer{Value: int64(buf.Position)}
	}})

	// __bytebuffer_set_position(buf, pos) - 设置当前位置
	env.Set("__bytebuffer_set_position", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__bytebuffer_set_position 需要2个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_set_position 第一个参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		pos, ok := args[1].(*Integer)
		if !ok {
			return newError("__bytebuffer_set_position 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		buf.Position = int(pos.Value)
		if buf.Position < 0 {
			buf.Position = 0
		}
		if buf.Position > len(buf.Buffer) {
			buf.Position = len(buf.Buffer)
		}
		return &Null{}
	}})

	// __bytebuffer_reset(buf) - 重置位置到开始
	env.Set("__bytebuffer_reset", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_reset 需要1个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_reset 参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		buf.Position = 0
		return &Null{}
	}})

	// __bytebuffer_clear(buf) - 清空缓冲区
	env.Set("__bytebuffer_clear", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_clear 需要1个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_clear 参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		buf.Buffer = buf.Buffer[:0]
		buf.Position = 0
		return &Null{}
	}})

	// __bytebuffer_size(buf) - 获取缓冲区大小
	env.Set("__bytebuffer_size", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_size 需要1个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_size 参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		return &Integer{Value: int64(len(buf.Buffer))}
	}})

	// __bytebuffer_remaining(buf) - 获取剩余可读字节数
	env.Set("__bytebuffer_remaining", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_remaining 需要1个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_remaining 参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		return &Integer{Value: int64(len(buf.Buffer) - buf.Position)}
	}})

	// __bytebuffer_to_bytes(buf) - 转换为字节数组
	env.Set("__bytebuffer_to_bytes", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_to_bytes 需要1个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_to_bytes 参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		return bytesToArray(buf.Buffer)
	}})

	// __bytebuffer_to_string(buf) - 转换为字符串
	env.Set("__bytebuffer_to_string", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__bytebuffer_to_string 需要1个参数，得到 %d 个", len(args))
		}
		buf, ok := args[0].(*ByteBuffer)
		if !ok {
			return newError("__bytebuffer_to_string 参数必须是 ByteBuffer，得到 %s", args[0].Type())
		}
		return &String{Value: string(buf.Buffer)}
	}})
}

// ByteBuffer 字节缓冲区对象
type ByteBuffer struct {
	Buffer   []byte
	Position int
}

func (bb *ByteBuffer) Type() ObjectType { return "BYTE_BUFFER" }
func (bb *ByteBuffer) Inspect() string {
	return "ByteBuffer(size=" + strconv.Itoa(len(bb.Buffer)) + ", pos=" + strconv.Itoa(bb.Position) + ")"
}

// 辅助函数：字节数组转 []byte
func arrayToBytes(arr *Array) []byte {
	data := make([]byte, len(arr.Elements))
	for i, elem := range arr.Elements {
		if byteInt, ok := elem.(*Integer); ok {
			data[i] = byte(byteInt.Value)
		}
	}
	return data
}

// 辅助函数：[]byte 转字节数组
func bytesToArray(data []byte) *Array {
	elements := make([]Object, len(data))
	for i, b := range data {
		elements[i] = &Integer{Value: int64(b)}
	}
	return &Array{Elements: elements}
}











