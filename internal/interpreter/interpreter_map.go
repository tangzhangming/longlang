package interpreter

import (
	"github.com/tangzhangming/longlang/internal/parser"
)

// evalMapLiteral 求值 Map 字面量
// 语法：map[KeyType]ValueType{key1: value1, key2: value2, ...}
func (i *Interpreter) evalMapLiteral(node *parser.MapLiteral) Object {
	// 获取类型信息
	keyType := "string"
	if node.Type.KeyType != nil {
		keyType = node.Type.KeyType.Value
	}

	// 目前只支持 string 键
	if keyType != "string" {
		return newError("Map 的键类型只支持 string，得到 %s (行 %d, 列 %d)",
			keyType, node.Token.Line, node.Token.Column)
	}

	// 获取值类型
	valueType := "any"
	if node.Type.ValueType != nil {
		if ident, ok := node.Type.ValueType.(*parser.Identifier); ok {
			valueType = ident.Value
		}
	}

	// 创建 Map 对象
	mapObj := &Map{
		Pairs:     make(map[string]Object),
		Keys:      []string{},
		KeyType:   keyType,
		ValueType: valueType,
	}

	// 求值所有键值对
	for idx, keyExpr := range node.Keys {
		// 求值键（必须是字符串）
		keyObj := i.Eval(keyExpr)
		if isError(keyObj) {
			return keyObj
		}
		keyStr, ok := keyObj.(*String)
		if !ok {
			return newError("Map 的键必须是字符串，得到 %s (行 %d, 列 %d)",
				keyObj.Type(), node.Token.Line, node.Token.Column)
		}

		// 求值值
		valueObj := i.Eval(node.Values[idx])
		if isError(valueObj) {
			return valueObj
		}

		// 类型检查（如果不是 any）
		if valueType != "any" && !i.checkMapValueType(valueObj, valueType) {
			return newError("Map 值类型不匹配：期望 %s，得到 %s (行 %d, 列 %d)",
				valueType, valueObj.Type(), node.Token.Line, node.Token.Column)
		}

		// 添加到 Map
		mapObj.Set(keyStr.Value, valueObj)
	}

	return mapObj
}

// checkMapValueType 检查 Map 值类型是否匹配
func (i *Interpreter) checkMapValueType(value Object, expectedType string) bool {
	switch expectedType {
	case "int":
		_, ok := value.(*Integer)
		return ok
	case "string":
		_, ok := value.(*String)
		return ok
	case "bool":
		_, ok := value.(*Boolean)
		return ok
	case "float":
		_, ok := value.(*Float)
		return ok
	case "any":
		return true
	default:
		// 自定义类型（类实例）
		if instance, ok := value.(*Instance); ok {
			return instance.Class.Name == expectedType
		}
		return false
	}
}

// evalMapIndexExpression 求值 Map 索引表达式
// 语法：map[key]
func (i *Interpreter) evalMapIndexExpression(mapObj *Map, index Object) Object {
	// 索引必须是字符串
	keyStr, ok := index.(*String)
	if !ok {
		return newError("Map 的键必须是字符串，得到 %s", index.Type())
	}

	// 获取值（不存在则抛出异常）
	value, exists := mapObj.Get(keyStr.Value)
	if !exists {
		return newError("Map 键不存在: %s", keyStr.Value)
	}

	return value
}

// evalMapAssignment 求值 Map 赋值表达式
// 语法：map[key] = value
func (i *Interpreter) evalMapAssignment(mapObj *Map, key Object, value Object) Object {
	// 键必须是字符串
	keyStr, ok := key.(*String)
	if !ok {
		return newError("Map 的键必须是字符串，得到 %s", key.Type())
	}

	// 类型检查
	if mapObj.ValueType != "any" && !i.checkMapValueType(value, mapObj.ValueType) {
		return newError("Map 值类型不匹配：期望 %s，得到 %s",
			mapObj.ValueType, value.Type())
	}

	// 设置值
	mapObj.Set(keyStr.Value, value)
	return value
}








