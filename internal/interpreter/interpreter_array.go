package interpreter

import (
	"github.com/tangzhangming/longlang/internal/parser"
)

// ========== 数组相关执行函数 ==========

// evalArrayLiteral 执行数组字面量 {element1, element2, ...}
// 返回一个数组对象，类型根据第一个元素推导
func (i *Interpreter) evalArrayLiteral(node *parser.ArrayLiteral) Object {
	elements := []Object{}
	var elementType string

	for idx, elem := range node.Elements {
		evaluated := i.Eval(elem)
		if isError(evaluated) {
			return evaluated
		}

		// 推导元素类型
		if idx == 0 {
			elementType = string(evaluated.Type())
		} else {
			// 检查类型一致性
			if string(evaluated.Type()) != elementType {
				return newError("数组元素类型不一致：期望 %s，得到 %s", elementType, evaluated.Type())
			}
		}

		elements = append(elements, evaluated)
	}

	return &Array{
		Elements:    elements,
		ElementType: elementType,
		IsFixed:     false, // 类型推导的数组是动态数组
		Capacity:    int64(len(elements)),
	}
}

// evalTypedArrayLiteral 执行带类型的数组字面量 [size]type{elements}
func (i *Interpreter) evalTypedArrayLiteral(node *parser.TypedArrayLiteral) Object {
	elements := []Object{}

	// 获取数组类型信息
	arrayType := node.Type
	var capacity int64
	isFixed := false

	if arrayType.Size != nil {
		// 固定长度数组
		sizeObj := i.Eval(arrayType.Size)
		if isError(sizeObj) {
			return sizeObj
		}
		intSize, ok := sizeObj.(*Integer)
		if !ok {
			return newError("数组长度必须是整数")
		}
		capacity = intSize.Value
		isFixed = true
	} else if arrayType.IsInferred {
		// [...] 长度推导
		isFixed = true
	}

	// 获取元素类型
	elementType := ""
	if arrayType.ElementType != nil {
		if ident, ok := arrayType.ElementType.(*parser.Identifier); ok {
			elementType = ident.Value
		}
	}

	// 执行每个元素
	for _, elem := range node.Elements {
		evaluated := i.Eval(elem)
		if isError(evaluated) {
			return evaluated
		}

		// 类型检查
		if elementType != "" {
			if err := i.checkArrayElementType(evaluated, elementType); err != nil {
				return err
			}
		}

		elements = append(elements, evaluated)
	}

	// 长度推导
	if arrayType.IsInferred {
		capacity = int64(len(elements))
	}

	// 固定长度数组长度检查
	if isFixed && !arrayType.IsInferred {
		if int64(len(elements)) != capacity {
			return newError("数组长度不匹配：期望 %d 个元素，得到 %d 个", capacity, len(elements))
		}
	}

	return &Array{
		Elements:    elements,
		ElementType: elementType,
		IsFixed:     isFixed,
		Capacity:    capacity,
	}
}

// evalIndexExpression 执行索引访问表达式 array[index]
func (i *Interpreter) evalIndexExpression(node *parser.IndexExpression) Object {
	left := i.Eval(node.Left)
	if isError(left) {
		return left
	}

	index := i.Eval(node.Index)
	if isError(index) {
		return index
	}

	return i.evalIndexOperator(left, index)
}

// evalIndexOperator 执行索引操作
func (i *Interpreter) evalIndexOperator(left, index Object) Object {
	switch {
	case left.Type() == ARRAY_OBJ && index.Type() == INTEGER_OBJ:
		return i.evalArrayIndexExpression(left, index)
	case left.Type() == STRING_OBJ && index.Type() == INTEGER_OBJ:
		return i.evalStringIndexExpression(left, index)
	case left.Type() == MAP_OBJ:
		return i.evalMapIndexExpression(left.(*Map), index)
	default:
		return newError("索引操作不支持类型: %s", left.Type())
	}
}

// evalArrayIndexExpression 执行数组索引访问
func (i *Interpreter) evalArrayIndexExpression(array, index Object) Object {
	arrayObject := array.(*Array)
	idx := index.(*Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	// 支持负数索引
	if idx < 0 {
		idx = int64(len(arrayObject.Elements)) + idx
	}

	if idx < 0 || idx > max {
		return newError("数组索引越界：索引 %d 超出范围 [0, %d]", index.(*Integer).Value, max)
	}

	return arrayObject.Elements[idx]
}

// evalStringIndexExpression 执行字符串索引访问
func (i *Interpreter) evalStringIndexExpression(str, index Object) Object {
	stringObject := str.(*String)
	idx := index.(*Integer).Value

	runes := []rune(stringObject.Value)
	max := int64(len(runes) - 1)

	// 支持负数索引
	if idx < 0 {
		idx = int64(len(runes)) + idx
	}

	if idx < 0 || idx > max {
		return newError("字符串索引越界：索引 %d 超出范围 [0, %d]", index.(*Integer).Value, max)
	}

	return &String{Value: string(runes[idx])}
}

// checkArrayElementType 检查数组元素类型
func (i *Interpreter) checkArrayElementType(element Object, expectedType string) *Error {
	actualType := element.Type()

	// 类型映射
	typeMatch := map[string]ObjectType{
		"int":    INTEGER_OBJ,
		"i8":     INTEGER_OBJ,
		"i16":    INTEGER_OBJ,
		"i32":    INTEGER_OBJ,
		"i64":    INTEGER_OBJ,
		"uint":   INTEGER_OBJ,
		"u8":     INTEGER_OBJ,
		"byte":   INTEGER_OBJ, // byte 是 u8 的别名
		"u16":    INTEGER_OBJ,
		"u32":    INTEGER_OBJ,
		"u64":    INTEGER_OBJ,
		"float":  FLOAT_OBJ,
		"f32":    FLOAT_OBJ,
		"f64":    FLOAT_OBJ,
		"string": STRING_OBJ,
		"bool":   BOOLEAN_OBJ,
		"any":    "", // any 类型接受任何类型
	}

	if expectedType == "any" {
		return nil
	}

	expected, ok := typeMatch[expectedType]
	if !ok {
		// 可能是自定义类型
		return nil
	}

	if actualType != expected {
		return newError("数组元素类型不匹配：期望 %s，得到 %s", expectedType, actualType)
	}

	// byte 类型的范围检查 (0-255)
	if expectedType == "byte" || expectedType == "u8" {
		if intVal, ok := element.(*Integer); ok {
			if intVal.Value < 0 || intVal.Value > 255 {
				return newError("byte 类型值超出范围 [0, 255]：得到 %d", intVal.Value)
			}
		}
	}

	return nil
}

// evalSliceExpression 执行切片表达式 array[start:end]
// Go风格切片语法，支持数组和字符串
func (i *Interpreter) evalSliceExpression(node *parser.SliceExpression) Object {
	left := i.Eval(node.Left)
	if isError(left) {
		return left
	}

	// 解析 start 索引
	var start int64 = 0
	if node.Start != nil {
		startObj := i.Eval(node.Start)
		if isError(startObj) {
			return startObj
		}
		startInt, ok := startObj.(*Integer)
		if !ok {
			return newError("切片起始索引必须是整数，得到 %s", startObj.Type())
		}
		start = startInt.Value
	}

	// 解析 end 索引
	var end int64 = -1 // -1 表示到末尾
	if node.End != nil {
		endObj := i.Eval(node.End)
		if isError(endObj) {
			return endObj
		}
		endInt, ok := endObj.(*Integer)
		if !ok {
			return newError("切片结束索引必须是整数，得到 %s", endObj.Type())
		}
		end = endInt.Value
	}

	switch obj := left.(type) {
	case *Array:
		return i.evalArraySlice(obj, start, end)
	case *String:
		return i.evalStringSlice(obj, start, end)
	default:
		return newError("切片操作不支持类型: %s", left.Type())
	}
}

// evalArraySlice 执行数组切片
// end == -1 表示切到末尾（来自 arr[start:] 语法）
func (i *Interpreter) evalArraySlice(array *Array, start, end int64) Object {
	length := int64(len(array.Elements))

	// end == -1 表示切到末尾
	if end == -1 {
		end = length
	}

	// 处理负数索引（用户显式使用负数，如 arr[-2:]）
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}

	// 边界检查
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start > end {
		return newError("切片范围无效：起始索引 %d 大于结束索引 %d", start, end)
	}

	// 创建新数组
	newElements := make([]Object, end-start)
	copy(newElements, array.Elements[start:end])

	return &Array{
		Elements:    newElements,
		ElementType: array.ElementType,
		IsFixed:     false,
		Capacity:    int64(len(newElements)),
	}
}

// evalStringSlice 执行字符串切片
// end == -1 表示切到末尾（来自 str[start:] 语法）
func (i *Interpreter) evalStringSlice(str *String, start, end int64) Object {
	runes := []rune(str.Value)
	length := int64(len(runes))

	// end == -1 表示切到末尾
	if end == -1 {
		end = length
	}

	// 处理负数索引
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}

	// 边界检查
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start > end {
		return newError("切片范围无效：起始索引 %d 大于结束索引 %d", start, end)
	}

	return &String{Value: string(runes[start:end])}
}

// evalArrayAssignment 执行数组或 Map 元素赋值 array[index] = value / map[key] = value
func (i *Interpreter) evalArrayAssignment(obj Object, index Object, value Object) Object {
	switch obj.Type() {
	case ARRAY_OBJ:
		return i.evalArrayElementAssignment(obj.(*Array), index, value)
	case MAP_OBJ:
		return i.evalMapAssignment(obj.(*Map), index, value)
	default:
		return newError("索引赋值只能用于数组或 Map 类型，得到 %s", obj.Type())
	}
}

// evalArrayElementAssignment 执行数组元素赋值
func (i *Interpreter) evalArrayElementAssignment(arrayObject *Array, index Object, value Object) Object {
	if index.Type() != INTEGER_OBJ {
		return newError("数组索引必须是整数，得到 %s", index.Type())
	}

	idx := index.(*Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	// 支持负数索引
	if idx < 0 {
		idx = int64(len(arrayObject.Elements)) + idx
	}

	if idx < 0 || idx > max {
		return newError("数组索引越界：索引 %d 超出范围 [0, %d]", index.(*Integer).Value, max)
	}

	// 类型检查
	if arrayObject.ElementType != "" {
		if err := i.checkArrayElementType(value, arrayObject.ElementType); err != nil {
			return err
		}
	}

	arrayObject.Elements[idx] = value
	return value
}

