package interpreter

// MapMethod Map 方法类型
type MapMethod func(m *Map, args ...Object) Object

// mapMethods 存储所有 Map 方法
var mapMethods = map[string]MapMethod{
	// ========== 基本信息 ==========
	"size":    mapSize,
	"isEmpty": mapIsEmpty,

	// ========== 操作 ==========
	"delete": mapDelete,
	"clear":  mapClear,

	// ========== 获取集合 ==========
	"keys":   mapKeys,
	"values": mapValues,
}

// GetMapMethod 获取 Map 方法
func GetMapMethod(name string) (MapMethod, bool) {
	method, ok := mapMethods[name]
	return method, ok
}

// ========== 基本信息方法 ==========

// mapSize 获取 Map 大小
// map.size() => int
func mapSize(m *Map, args ...Object) Object {
	return &Integer{Value: int64(m.Size())}
}

// mapIsEmpty 判断 Map 是否为空
// map.isEmpty() => bool
func mapIsEmpty(m *Map, args ...Object) Object {
	return &Boolean{Value: m.IsEmpty()}
}

// ========== 操作方法 ==========

// mapDelete 删除 Map 中的键值对
// map.delete("key") => bool
func mapDelete(m *Map, args ...Object) Object {
	if len(args) != 1 {
		return NewError("delete 方法需要1个参数，得到 %d 个", len(args))
	}
	keyStr, ok := args[0].(*String)
	if !ok {
		return NewError("delete 方法的参数必须是字符串，得到 %s", args[0].Type())
	}
	result := m.Delete(keyStr.Value)
	return &Boolean{Value: result}
}

// mapClear 清空 Map
// map.clear()
func mapClear(m *Map, args ...Object) Object {
	m.Clear()
	return &Null{}
}

// ========== 获取集合方法 ==========

// mapKeys 获取 Map 的所有键
// map.keys() => []string
func mapKeys(m *Map, args ...Object) Object {
	keys := m.GetKeys()
	elements := make([]Object, len(keys))
	for i, key := range keys {
		elements[i] = &String{Value: key}
	}
	return &Array{
		Elements:    elements,
		ElementType: "string",
		IsFixed:     false,
	}
}

// mapValues 获取 Map 的所有值
// map.values() => []ValueType
func mapValues(m *Map, args ...Object) Object {
	values := m.GetValues()
	return &Array{
		Elements:    values,
		ElementType: m.ValueType,
		IsFixed:     false,
	}
}










