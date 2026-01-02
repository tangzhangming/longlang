package interpreter

import (
	"fmt"

	"github.com/tangzhangming/longlang/internal/parser"
)

// evalTypeAssertionExpression 执行类型断言表达式
// 对应语法：value as Type（强制断言）或 value as? Type（安全断言）
// 强制断言：类型不匹配时抛出 TypeError
// 安全断言：类型不匹配时返回 null
func (i *Interpreter) evalTypeAssertionExpression(node *parser.TypeAssertionExpression) Object {
	// 计算左侧表达式的值
	value := i.Eval(node.Left)
	if isError(value) {
		return value
	}

	// 获取目标类型名称
	targetTypeName := getTypeName(node.TargetType)
	if targetTypeName == "" {
		return newError("无效的类型断言目标类型")
	}

	// 执行类型断言
	result, ok := i.assertType(value, targetTypeName, node.TargetType)

	if !ok {
		if node.IsSafe {
			// 安全断言：返回 null
			return &Null{}
		}
		// 强制断言：返回类型错误（会被 try-catch 捕获）
		actualType := getActualTypeName(value)
		return &ThrownException{
			RuntimeError: newError("类型断言失败: 无法将 %s 转换为 %s", actualType, targetTypeName),
		}
	}

	return result
}

// getTypeName 从类型表达式中获取类型名称
func getTypeName(typeExpr parser.Expression) string {
	switch t := typeExpr.(type) {
	case *parser.Identifier:
		return t.Value
	case *parser.ArrayType:
		elemType := getTypeName(t.ElementType)
		if elemType == "" {
			return ""
		}
		return "[]" + elemType
	case *parser.MapType:
		keyType := ""
		if t.KeyType != nil {
			keyType = t.KeyType.Value
		}
		valueType := getTypeName(t.ValueType)
		if valueType == "" {
			return ""
		}
		return "map[" + keyType + "]" + valueType
	default:
		return ""
	}
}

// getActualTypeName 获取对象的实际类型名称
func getActualTypeName(obj Object) string {
	switch o := obj.(type) {
	case *Integer:
		return "int"
	case *Float:
		return "float"
	case *String:
		return "string"
	case *Boolean:
		return "bool"
	case *Null:
		return "null"
	case *Array:
		if o.ElementType != "" {
			return "[]" + o.ElementType
		}
		return "[]any"
	case *Map:
		keyType := o.KeyType
		if keyType == "" {
			keyType = "string"
		}
		valueType := o.ValueType
		if valueType == "" {
			valueType = "any"
		}
		return "map[" + keyType + "]" + valueType
	case *Instance:
		return o.Class.Name
	case *Function:
		return "function"
	case *BoundMethod:
		return "method"
	case *Enum:
		return "enum"
	case *EnumValue:
		return o.Enum.Name
	default:
		return fmt.Sprintf("%T", obj)
	}
}

// assertType 执行类型断言
// 返回转换后的值和是否成功
func (i *Interpreter) assertType(value Object, targetTypeName string, targetTypeExpr parser.Expression) (Object, bool) {
	// 处理 null 值
	if _, isNull := value.(*Null); isNull {
		return nil, false
	}

	// 处理 any 类型（始终成功）
	if targetTypeName == "any" {
		return value, true
	}

	switch targetTypeName {
	// 基本类型
	case "int", "i8", "i16", "i32", "i64":
		if intVal, ok := value.(*Integer); ok {
			return intVal, true
		}
		return nil, false

	case "uint", "u8", "u16", "u32", "u64", "byte":
		if intVal, ok := value.(*Integer); ok {
			// 无符号类型检查值是否为非负数
			if intVal.Value >= 0 {
				return intVal, true
			}
		}
		return nil, false

	case "float", "f32", "f64":
		if floatVal, ok := value.(*Float); ok {
			return floatVal, true
		}
		// int 可以转为 float
		if intVal, ok := value.(*Integer); ok {
			return &Float{Value: float64(intVal.Value)}, true
		}
		return nil, false

	case "string":
		if strVal, ok := value.(*String); ok {
			return strVal, true
		}
		return nil, false

	case "bool":
		if boolVal, ok := value.(*Boolean); ok {
			return boolVal, true
		}
		return nil, false

	default:
		// 检查是否是数组类型
		if len(targetTypeName) > 2 && targetTypeName[:2] == "[]" {
			return i.assertArrayType(value, targetTypeName, targetTypeExpr)
		}

		// 检查是否是 Map 类型
		if len(targetTypeName) > 4 && targetTypeName[:4] == "map[" {
			return i.assertMapType(value, targetTypeName, targetTypeExpr)
		}

		// 检查是否是类/接口类型
		return i.assertClassType(value, targetTypeName)
	}
}

// assertArrayType 断言数组类型
func (i *Interpreter) assertArrayType(value Object, targetTypeName string, targetTypeExpr parser.Expression) (Object, bool) {
	arr, ok := value.(*Array)
	if !ok {
		return nil, false
	}

	// 获取目标元素类型
	targetElemType := targetTypeName[2:] // 去掉 "[]"

	// 如果目标是 []any，总是成功
	if targetElemType == "any" {
		return arr, true
	}

	// 检查数组元素类型是否兼容
	if arr.ElementType == targetElemType || arr.ElementType == "" {
		// 空数组或类型匹配
		return arr, true
	}

	// 检查每个元素是否可以转换
	for _, elem := range arr.Elements {
		_, elemOk := i.assertType(elem, targetElemType, nil)
		if !elemOk {
			return nil, false
		}
	}

	return arr, true
}

// assertMapType 断言 Map 类型
func (i *Interpreter) assertMapType(value Object, targetTypeName string, targetTypeExpr parser.Expression) (Object, bool) {
	m, ok := value.(*Map)
	if !ok {
		return nil, false
	}

	// 简单检查：只要是 Map 就可以
	// 更严格的检查需要解析 targetTypeName 并验证键值类型
	return m, true
}

// assertClassType 断言类/接口类型
func (i *Interpreter) assertClassType(value Object, targetTypeName string) (Object, bool) {
	instance, ok := value.(*Instance)
	if !ok {
		return nil, false
	}

	// 检查是否是目标类型
	if instance.Class.Name == targetTypeName {
		return instance, true
	}

	// 检查继承链
	if i.isInstanceOf(instance, targetTypeName) {
		return instance, true
	}

	return nil, false
}

// isInstanceOf 检查实例是否是指定类型（包括父类和接口）
func (i *Interpreter) isInstanceOf(instance *Instance, typeName string) bool {
	// 检查直接类型
	if instance.Class.Name == typeName {
		return true
	}

	// 检查父类
	class := instance.Class
	for class.Parent != nil {
		if class.Parent.Name == typeName {
			return true
		}
		class = class.Parent
	}

	// 检查实现的接口
	for _, iface := range instance.Class.Interfaces {
		if iface.Name == typeName {
			return true
		}
	}

	return false
}

