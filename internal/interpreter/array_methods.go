package interpreter

// ArrayMethod 数组方法类型
type ArrayMethod func(a *Array, args ...Object) Object

// arrayMethods 存储所有数组方法
var arrayMethods = map[string]ArrayMethod{
	// ========== 基本信息 ==========
	"length":  arrayLength,
	"isEmpty": arrayIsEmpty,

	// ========== 添加和删除 ==========
	"push":  arrayPush,
	"pop":   arrayPop,
	"shift": arrayShift,

	// ========== 查找 ==========
	"contains": arrayContains,
	"indexOf":  arrayIndexOf,

	// ========== 转换 ==========
	"join":    arrayJoin,
	"reverse": arrayReverse,
	"slice":   arraySlice,

	// ========== 其他 ==========
	"clear": arrayClear,
}

// GetArrayMethod 获取数组方法
func GetArrayMethod(name string) (ArrayMethod, bool) {
	method, ok := arrayMethods[name]
	return method, ok
}

// ========== 基本信息方法 ==========

// arrayLength 获取数组长度
// arr.length() => int
func arrayLength(a *Array, args ...Object) Object {
	return &Integer{Value: int64(len(a.Elements))}
}

// arrayIsEmpty 判断数组是否为空
// arr.isEmpty() => bool
func arrayIsEmpty(a *Array, args ...Object) Object {
	return &Boolean{Value: len(a.Elements) == 0}
}

// ========== 添加和删除方法 ==========

// arrayPush 在数组末尾添加元素
// arr.push(value)
func arrayPush(a *Array, args ...Object) Object {
	if len(args) != 1 {
		return NewError("push 方法需要1个参数，得到 %d 个", len(args))
	}
	a.Elements = append(a.Elements, args[0])
	return &Null{}
}

// arrayPop 删除并返回数组最后一个元素
// arr.pop() => element
func arrayPop(a *Array, args ...Object) Object {
	if len(a.Elements) == 0 {
		return NewError("不能从空数组中 pop")
	}
	last := a.Elements[len(a.Elements)-1]
	a.Elements = a.Elements[:len(a.Elements)-1]
	return last
}

// arrayShift 删除并返回数组第一个元素
// arr.shift() => element
func arrayShift(a *Array, args ...Object) Object {
	if len(a.Elements) == 0 {
		return NewError("不能从空数组中 shift")
	}
	first := a.Elements[0]
	a.Elements = a.Elements[1:]
	return first
}

// ========== 查找方法 ==========

// arrayContains 判断数组是否包含某个元素
// arr.contains(value) => bool
func arrayContains(a *Array, args ...Object) Object {
	if len(args) != 1 {
		return NewError("contains 方法需要1个参数，得到 %d 个", len(args))
	}
	target := args[0]
	for _, elem := range a.Elements {
		if objectsEqual(elem, target) {
			return &Boolean{Value: true}
		}
	}
	return &Boolean{Value: false}
}

// arrayIndexOf 返回元素第一次出现的索引，不存在返回 -1
// arr.indexOf(value) => int
func arrayIndexOf(a *Array, args ...Object) Object {
	if len(args) != 1 {
		return NewError("indexOf 方法需要1个参数，得到 %d 个", len(args))
	}
	target := args[0]
	for i, elem := range a.Elements {
		if objectsEqual(elem, target) {
			return &Integer{Value: int64(i)}
		}
	}
	return &Integer{Value: -1}
}

// ========== 转换方法 ==========

// arrayJoin 用分隔符连接数组元素
// arr.join(",") => string
func arrayJoin(a *Array, args ...Object) Object {
	if len(args) != 1 {
		return NewError("join 方法需要1个参数，得到 %d 个", len(args))
	}
	sep, ok := args[0].(*String)
	if !ok {
		return NewError("join 方法的参数必须是字符串，得到 %s", args[0].Type())
	}
	
	result := ""
	for i, elem := range a.Elements {
		if i > 0 {
			result += sep.Value
		}
		result += elem.Inspect()
	}
	return &String{Value: result}
}

// arrayReverse 反转数组（返回新数组）
// arr.reverse() => []
func arrayReverse(a *Array, args ...Object) Object {
	newElements := make([]Object, len(a.Elements))
	for i, elem := range a.Elements {
		newElements[len(a.Elements)-1-i] = elem
	}
	return &Array{
		Elements:    newElements,
		ElementType: a.ElementType,
		IsFixed:     a.IsFixed,
	}
}

// arraySlice 截取数组片段
// arr.slice(start) 或 arr.slice(start, end)
func arraySlice(a *Array, args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return NewError("slice 方法需要1-2个参数，得到 %d 个", len(args))
	}
	
	start, ok := args[0].(*Integer)
	if !ok {
		return NewError("slice 的起始参数必须是整数，得到 %s", args[0].Type())
	}
	
	startIdx := int(start.Value)
	endIdx := len(a.Elements)
	
	if len(args) == 2 {
		end, ok := args[1].(*Integer)
		if !ok {
			return NewError("slice 的结束参数必须是整数，得到 %s", args[1].Type())
		}
		endIdx = int(end.Value)
	}
	
	// 边界检查
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(a.Elements) {
		endIdx = len(a.Elements)
	}
	if startIdx > endIdx {
		startIdx = endIdx
	}
	
	newElements := make([]Object, endIdx-startIdx)
	copy(newElements, a.Elements[startIdx:endIdx])
	
	return &Array{
		Elements:    newElements,
		ElementType: a.ElementType,
		IsFixed:     false,
	}
}

// ========== 其他方法 ==========

// arrayClear 清空数组
// arr.clear()
func arrayClear(a *Array, args ...Object) Object {
	a.Elements = []Object{}
	return &Null{}
}

// ========== 辅助函数 ==========

// objectsEqual 比较两个对象是否相等
func objectsEqual(a, b Object) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *Integer:
		bv := b.(*Integer)
		return av.Value == bv.Value
	case *Float:
		bv := b.(*Float)
		return av.Value == bv.Value
	case *String:
		bv := b.(*String)
		return av.Value == bv.Value
	case *Boolean:
		bv := b.(*Boolean)
		return av.Value == bv.Value
	case *Null:
		return true
	default:
		return a == b // 引用相等
	}
}









