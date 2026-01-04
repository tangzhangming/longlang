package vm

import (
	"fmt"
	"strings"

	"github.com/tangzhangming/longlang/internal/interpreter"
)

// ========== 算术运算 ==========

// binaryAdd 加法运算
func (vm *VM) binaryAdd() error {
	b := vm.pop()
	a := vm.pop()

	if a == nil || b == nil {
		return fmt.Errorf("加法操作数为nil: a=%v, b=%v", a, b)
	}

	switch av := a.(type) {
	case *interpreter.Integer:
		switch bv := b.(type) {
		case *interpreter.Integer:
			vm.push(&interpreter.Integer{Value: av.Value + bv.Value})
			return nil
		case *interpreter.Float:
			vm.push(&interpreter.Float{Value: float64(av.Value) + bv.Value})
			return nil
		case *interpreter.String:
			vm.push(&interpreter.String{Value: fmt.Sprintf("%d%s", av.Value, bv.Value)})
			return nil
		}
	case *interpreter.Float:
		switch bv := b.(type) {
		case *interpreter.Integer:
			vm.push(&interpreter.Float{Value: av.Value + float64(bv.Value)})
			return nil
		case *interpreter.Float:
			vm.push(&interpreter.Float{Value: av.Value + bv.Value})
			return nil
		case *interpreter.String:
			vm.push(&interpreter.String{Value: fmt.Sprintf("%g%s", av.Value, bv.Value)})
			return nil
		}
	case *interpreter.String:
		// 字符串拼接
		if b == nil {
			return fmt.Errorf("加法操作数b为nil")
		}
		vm.push(&interpreter.String{Value: av.Value + vm.objectToString(b)})
		return nil
	}

	return fmt.Errorf("不支持的加法操作: %s + %s", a.Type(), b.Type())
}

// binaryOp 通用二元运算
func (vm *VM) binaryOp(intOp func(int64, int64) int64, floatOp func(float64, float64) float64) error {
	b := vm.pop()
	a := vm.pop()

	switch av := a.(type) {
	case *interpreter.Integer:
		switch bv := b.(type) {
		case *interpreter.Integer:
			vm.push(&interpreter.Integer{Value: intOp(av.Value, bv.Value)})
			return nil
		case *interpreter.Float:
			vm.push(&interpreter.Float{Value: floatOp(float64(av.Value), bv.Value)})
			return nil
		}
	case *interpreter.Float:
		switch bv := b.(type) {
		case *interpreter.Integer:
			vm.push(&interpreter.Float{Value: floatOp(av.Value, float64(bv.Value))})
			return nil
		case *interpreter.Float:
			vm.push(&interpreter.Float{Value: floatOp(av.Value, bv.Value)})
			return nil
		}
	}

	return fmt.Errorf("不支持的运算: %s 和 %s", a.Type(), b.Type())
}

// binaryDiv 除法运算
func (vm *VM) binaryDiv() error {
	b := vm.pop()
	a := vm.pop()

	switch av := a.(type) {
	case *interpreter.Integer:
		switch bv := b.(type) {
		case *interpreter.Integer:
			if bv.Value == 0 {
				return fmt.Errorf("除以零")
			}
			vm.push(&interpreter.Integer{Value: av.Value / bv.Value})
			return nil
		case *interpreter.Float:
			if bv.Value == 0 {
				return fmt.Errorf("除以零")
			}
			vm.push(&interpreter.Float{Value: float64(av.Value) / bv.Value})
			return nil
		}
	case *interpreter.Float:
		switch bv := b.(type) {
		case *interpreter.Integer:
			if bv.Value == 0 {
				return fmt.Errorf("除以零")
			}
			vm.push(&interpreter.Float{Value: av.Value / float64(bv.Value)})
			return nil
		case *interpreter.Float:
			if bv.Value == 0 {
				return fmt.Errorf("除以零")
			}
			vm.push(&interpreter.Float{Value: av.Value / bv.Value})
			return nil
		}
	}

	return fmt.Errorf("不支持的除法操作: %s / %s", a.Type(), b.Type())
}

// binaryMod 取模运算
func (vm *VM) binaryMod() error {
	b := vm.pop()
	a := vm.pop()

	if av, ok := a.(*interpreter.Integer); ok {
		if bv, ok := b.(*interpreter.Integer); ok {
			if bv.Value == 0 {
				return fmt.Errorf("模零")
			}
			vm.push(&interpreter.Integer{Value: av.Value % bv.Value})
			return nil
		}
	}

	return fmt.Errorf("取模运算只支持整数类型")
}

// unaryNeg 取负运算
func (vm *VM) unaryNeg() error {
	operand := vm.pop()

	switch v := operand.(type) {
	case *interpreter.Integer:
		vm.push(&interpreter.Integer{Value: -v.Value})
		return nil
	case *interpreter.Float:
		vm.push(&interpreter.Float{Value: -v.Value})
		return nil
	}

	return fmt.Errorf("取负运算只支持数字类型")
}

// ========== 比较运算 ==========

// compareOp 比较运算
func (vm *VM) compareOp(op string) error {
	b := vm.pop()
	a := vm.pop()

	var result bool

	switch av := a.(type) {
	case *interpreter.Integer:
		switch bv := b.(type) {
		case *interpreter.Integer:
			result = compareInts(av.Value, bv.Value, op)
		case *interpreter.Float:
			result = compareFloats(float64(av.Value), bv.Value, op)
		default:
			return fmt.Errorf("不能比较 %s 和 %s", a.Type(), b.Type())
		}
	case *interpreter.Float:
		switch bv := b.(type) {
		case *interpreter.Integer:
			result = compareFloats(av.Value, float64(bv.Value), op)
		case *interpreter.Float:
			result = compareFloats(av.Value, bv.Value, op)
		default:
			return fmt.Errorf("不能比较 %s 和 %s", a.Type(), b.Type())
		}
	case *interpreter.String:
		if bv, ok := b.(*interpreter.String); ok {
			result = compareStrings(av.Value, bv.Value, op)
		} else {
			return fmt.Errorf("不能比较 %s 和 %s", a.Type(), b.Type())
		}
	default:
		return fmt.Errorf("不支持的比较类型: %s", a.Type())
	}

	vm.push(&interpreter.Boolean{Value: result})
	return nil
}

func compareInts(a, b int64, op string) bool {
	switch op {
	case "<":
		return a < b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case ">=":
		return a >= b
	}
	return false
}

func compareFloats(a, b float64, op string) bool {
	switch op {
	case "<":
		return a < b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case ">=":
		return a >= b
	}
	return false
}

func compareStrings(a, b, op string) bool {
	switch op {
	case "<":
		return a < b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case ">=":
		return a >= b
	}
	return false
}

// isEqual 判断两个值是否相等
func (vm *VM) isEqual(a, b interpreter.Object) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch av := a.(type) {
	case *interpreter.Integer:
		switch bv := b.(type) {
		case *interpreter.Integer:
			return av.Value == bv.Value
		case *interpreter.Float:
			return float64(av.Value) == bv.Value
		}
	case *interpreter.Float:
		switch bv := b.(type) {
		case *interpreter.Integer:
			return av.Value == float64(bv.Value)
		case *interpreter.Float:
			return av.Value == bv.Value
		}
	case *interpreter.String:
		if bv, ok := b.(*interpreter.String); ok {
			return av.Value == bv.Value
		}
	case *interpreter.Boolean:
		if bv, ok := b.(*interpreter.Boolean); ok {
			return av.Value == bv.Value
		}
	case *interpreter.Null:
		_, ok := b.(*interpreter.Null)
		return ok
	}

	// 引用相等
	return a == b
}

// isTruthy 判断值是否为真
func (vm *VM) isTruthy(obj interpreter.Object) bool {
	if obj == nil {
		return false
	}

	switch v := obj.(type) {
	case *interpreter.Null:
		return false
	case *interpreter.Boolean:
		return v.Value
	case *interpreter.Integer:
		return v.Value != 0
	case *interpreter.Float:
		return v.Value != 0
	case *interpreter.String:
		return v.Value != ""
	default:
		return true
	}
}

// ========== 位运算 ==========

// bitwiseOp 位运算
func (vm *VM) bitwiseOp(op func(int64, int64) int64) error {
	b := vm.pop()
	a := vm.pop()

	if av, ok := a.(*interpreter.Integer); ok {
		if bv, ok := b.(*interpreter.Integer); ok {
			vm.push(&interpreter.Integer{Value: op(av.Value, bv.Value)})
			return nil
		}
	}

	return fmt.Errorf("位运算只支持整数类型")
}

// ========== 函数调用 ==========

// callValue 调用值
func (vm *VM) callValue(callee interpreter.Object, argCount int) error {
	switch fn := callee.(type) {
	case *Closure:
		return vm.callClosure(fn, argCount)
	case *BoundMethod:
		vm.stack[vm.sp-argCount-1] = fn.Receiver
		return vm.callClosure(fn.Method, argCount)
	case *interpreter.Builtin:
		return vm.callBuiltinFn(fn, argCount)
	case *interpreter.Class:
		return vm.callClass(fn, argCount)
	case *interpreter.BuiltinObject:
		// 命名空间对象不能直接调用
		return fmt.Errorf("不能直接调用命名空间对象")
	}

	return fmt.Errorf("不能调用 %s 类型", callee.Type())
}

// callClosure 调用闭包
func (vm *VM) callClosure(closure *Closure, argCount int) error {
	if argCount != closure.Fn.NumParams && !closure.Fn.IsVariadic {
		return fmt.Errorf("函数 %s 需要 %d 个参数，但传入了 %d 个",
			closure.Fn.Name, closure.Fn.NumParams, argCount)
	}

	// 创建新帧
	// basePointer 指向第一个参数在栈上的位置
	// 返回时需要额外弹出函数对象（在 basePointer - 1 的位置）
	frame := vm.pushFrame(closure, vm.sp-argCount)
	frame.isMethodCall = false
	return nil
}

// callMethod 调用方法（方法调用没有函数对象在栈上）
func (vm *VM) callMethod(closure *Closure, argCount int) error {
	// 允许参数数量小于等于 NumParams（支持默认参数）
	if argCount > closure.Fn.NumParams && !closure.Fn.IsVariadic {
		return fmt.Errorf("方法 %s 最多需要 %d 个参数，但传入了 %d 个",
			closure.Fn.Name, closure.Fn.NumParams, argCount)
	}

	// 创建新帧
	// 对于方法调用，basePointer 指向 receiver（this），不需要弹出额外的函数对象
	frame := vm.pushFrame(closure, vm.sp-argCount)
	frame.isMethodCall = true
	return nil
}

// callConstructor 调用构造函数
func (vm *VM) callConstructor(closure *Closure, argCount int) error {
	// 允许参数数量小于等于 NumParams（支持默认参数）
	if argCount > closure.Fn.NumParams && !closure.Fn.IsVariadic {
		return fmt.Errorf("构造函数 %s 最多需要 %d 个参数，但传入了 %d 个",
			closure.Fn.Name, closure.Fn.NumParams, argCount)
	}

	// 创建新帧
	frame := vm.pushFrame(closure, vm.sp-argCount)
	frame.isMethodCall = true
	frame.isConstructor = true
	return nil
}

// callBuiltinFn 调用内置函数
func (vm *VM) callBuiltinFn(builtin *interpreter.Builtin, argCount int) error {
	args := make([]interpreter.Object, argCount)
	for i := argCount - 1; i >= 0; i-- {
		args[i] = vm.pop()
	}
	vm.pop() // 弹出函数本身

	result := builtin.Fn(args...)
	if err, ok := result.(*interpreter.Error); ok {
		return fmt.Errorf("%s", err.Message)
	}

	vm.push(result)
	return nil
}

// callClass 调用类（创建实例）
func (vm *VM) callClass(class *interpreter.Class, argCount int) error {
	// 创建实例
	instance := &interpreter.Instance{
		Class:  class,
		Fields: make(map[string]interpreter.Object),
	}

	// 初始化字段
	for name, variable := range class.Variables {
		if variable.DefaultValue != nil {
			instance.Fields[name] = variable.DefaultValue
		} else {
			instance.Fields[name] = &interpreter.Null{}
		}
	}

	// 替换栈上的类为实例
	vm.stack[vm.sp-argCount-1] = instance

	// 调用构造函数（如果存在）
	if constructor, ok := class.GetMethod("__construct"); ok {
		if closure, ok := constructor.Body.(*Closure); ok {
			// 对于实例方法，argCount+1 包括 this（receiver）
			// 现在 NumParams 已经包括了 this，所以传递 argCount+1
			err := vm.callConstructor(closure, argCount+1)
			if err == nil {
				vm.currentFrame().constructorInstance = instance
			}
			return err
		}
	}

	// 没有构造函数，弹出参数
	vm.sp -= argCount

	return nil
}

// callBuiltin 调用内置函数（通过 ID）
func (vm *VM) callBuiltin(builtinID, argCount int) error {
	// 这里可以根据 builtinID 调用不同的内置函数
	// 暂时简化处理
	return nil
}

// invoke 调用方法
func (vm *VM) invoke(name string, argCount int) error {
	receiver := vm.peek(argCount)

	switch obj := receiver.(type) {
	case *interpreter.Instance:
		// 先查找字段（可能是一个闭包）
		if field, ok := obj.Fields[name]; ok {
			if closure, ok := field.(*Closure); ok {
				// 字段闭包不包括 this，直接调用
				return vm.callClosure(closure, argCount)
			}
		}
		// 查找方法
		if method, ok := obj.Class.GetMethod(name); ok {
			if closure, ok := method.Body.(*Closure); ok {
				// 实例方法调用：栈上有 [receiver, arg0, arg1, ...]
				// 为了与普通函数调用统一，在receiver前插入一个占位值
				// 但这会改变整个栈，太复杂了
				// 改用标记方法：在帧中标记这是方法调用
				return vm.callMethod(closure, argCount+1)
			}
		}
		return fmt.Errorf("实例没有方法: %s", name)

	case *interpreter.String:
		return vm.invokeStringMethod(obj, name, argCount)

	case *interpreter.Array:
		return vm.invokeArrayMethod(obj, name, argCount)

	case *interpreter.Map:
		return vm.invokeMapMethod(obj, name, argCount)

	case *interpreter.BuiltinObject:
		// 命名空间方法调用
		if field, ok := obj.GetField(name); ok {
			if builtin, ok := field.(*interpreter.Builtin); ok {
				return vm.callBuiltinFn(builtin, argCount)
			}
		}
		return fmt.Errorf("命名空间没有方法: %s", name)

	case *interpreter.Class:
		// 尝试调用静态方法
		if method, ok := obj.GetStaticMethod(name); ok {
			if closure, ok := method.Body.(*Closure); ok {
				err := vm.callClosure(closure, argCount)
				if err != nil {
					return err
				}
				return nil
			}
		}
		// 如果不是静态方法，返回错误
		return fmt.Errorf("不能在 CLASS 上调用实例方法: %s", name)
	}

	return fmt.Errorf("不能在 %s 上调用方法", receiver.Type())
}

// invokeStatic 调用静态方法
func (vm *VM) invokeStatic(name string, argCount int) error {
	receiver := vm.peek(argCount)

	switch obj := receiver.(type) {
	case *interpreter.Class:
		if method, ok := obj.GetStaticMethod(name); ok {
			if closure, ok := method.Body.(*Closure); ok {
				err := vm.callClosure(closure, argCount)
				if err == nil {
					// 设置被调用的类名，支持 Late Static Binding
					newFrame := vm.frames[vm.frameCount-1]
					newFrame.calledClassName = obj.Name
					// 如果有命名空间，使用完整类名
					if obj.Namespace != "" {
						newFrame.calledClassName = obj.Namespace + "." + obj.Name
					}
				}
				return err
			}
		}
		return fmt.Errorf("类 %s 没有静态方法: %s", obj.Name, name)

	case *interpreter.BuiltinObject:
		// 内置对象的"静态"方法调用（如 Console::writeLine）
		if field, ok := obj.GetField(name); ok {
			if builtin, ok := field.(*interpreter.Builtin); ok {
				return vm.callBuiltinStatic(builtin, argCount)
			}
		}
		return fmt.Errorf("内置对象 %s 没有方法: %s", obj.Name, name)

	case *interpreter.String:
		// 通过类名字符串调用静态方法（用于 self:: 和 static::）
		className := obj.Value
		class, ok := vm.getClassByName(className)
		if !ok {
			return fmt.Errorf("未找到类: %s", className)
		}
		// 替换栈上的字符串为类对象
		vm.stack[vm.sp-argCount-1] = class
		// 递归调用
		return vm.invokeStatic(name, argCount)
	}

	return fmt.Errorf("只能在类或内置对象上调用静态方法，得到: %s", receiver.Type())
}

// callBuiltinStatic 调用内置对象的静态方法
func (vm *VM) callBuiltinStatic(builtin *interpreter.Builtin, argCount int) error {
	args := make([]interpreter.Object, argCount)
	for i := argCount - 1; i >= 0; i-- {
		args[i] = vm.pop()
	}
	vm.pop() // 弹出接收者（BuiltinObject）

	result := builtin.Fn(args...)
	if err, ok := result.(*interpreter.Error); ok {
		return fmt.Errorf("%s", err.Message)
	}

	vm.push(result)
	return nil
}

// ========== 字符串方法 ==========

// invokeStringMethod 调用字符串方法
func (vm *VM) invokeStringMethod(str *interpreter.String, name string, argCount int) error {
	args := make([]interpreter.Object, argCount)
	for i := argCount - 1; i >= 0; i-- {
		args[i] = vm.pop()
	}
	vm.pop() // 弹出字符串本身

	var result interpreter.Object

	switch name {
	case "length":
		result = &interpreter.Integer{Value: int64(len([]rune(str.Value)))}
	case "toUpper":
		result = &interpreter.String{Value: strings.ToUpper(str.Value)}
	case "toLower":
		result = &interpreter.String{Value: strings.ToLower(str.Value)}
	case "trim":
		result = &interpreter.String{Value: strings.TrimSpace(str.Value)}
	case "contains":
		if len(args) > 0 {
			if s, ok := args[0].(*interpreter.String); ok {
				result = &interpreter.Boolean{Value: strings.Contains(str.Value, s.Value)}
			}
		}
	case "indexOf":
		if len(args) > 0 {
			if s, ok := args[0].(*interpreter.String); ok {
				result = &interpreter.Integer{Value: int64(strings.Index(str.Value, s.Value))}
			}
		}
	case "split":
		if len(args) > 0 {
			if s, ok := args[0].(*interpreter.String); ok {
				parts := strings.Split(str.Value, s.Value)
				elements := make([]interpreter.Object, len(parts))
				for i, p := range parts {
					elements[i] = &interpreter.String{Value: p}
				}
				result = &interpreter.Array{Elements: elements, ElementType: "string"}
			}
		}
	case "replace":
		if len(args) >= 2 {
			if old, ok := args[0].(*interpreter.String); ok {
				if newStr, ok := args[1].(*interpreter.String); ok {
					result = &interpreter.String{Value: strings.ReplaceAll(str.Value, old.Value, newStr.Value)}
				}
			}
		}
	case "substring":
		if len(args) >= 1 {
			if start, ok := args[0].(*interpreter.Integer); ok {
				runes := []rune(str.Value)
				startIdx := int(start.Value)
				if startIdx < 0 {
					startIdx = 0
				}
				if startIdx > len(runes) {
					startIdx = len(runes)
				}
				endIdx := len(runes)
				if len(args) >= 2 {
					if end, ok := args[1].(*interpreter.Integer); ok {
						endIdx = int(end.Value)
						if endIdx > len(runes) {
							endIdx = len(runes)
						}
					}
				}
				result = &interpreter.String{Value: string(runes[startIdx:endIdx])}
			}
		}
	default:
		// 尝试使用 interpreter 包中的字符串方法
		if method, ok := interpreter.GetStringMethod(name); ok {
			result = method(str, args...)
		} else {
			return fmt.Errorf("字符串没有方法: %s", name)
		}
	}

	if result == nil {
		result = &interpreter.Null{}
	}
	vm.push(result)
	return nil
}

// ========== 数组方法 ==========

// invokeArrayMethod 调用数组方法
func (vm *VM) invokeArrayMethod(arr *interpreter.Array, name string, argCount int) error {
	args := make([]interpreter.Object, argCount)
	for i := argCount - 1; i >= 0; i-- {
		args[i] = vm.pop()
	}
	vm.pop() // 弹出数组本身

	var result interpreter.Object

	switch name {
	case "length":
		result = &interpreter.Integer{Value: int64(len(arr.Elements))}
	case "push":
		if len(args) > 0 {
			arr.Elements = append(arr.Elements, args[0])
			result = &interpreter.Integer{Value: int64(len(arr.Elements))}
		}
	case "pop":
		if len(arr.Elements) > 0 {
			result = arr.Elements[len(arr.Elements)-1]
			arr.Elements = arr.Elements[:len(arr.Elements)-1]
		} else {
			result = &interpreter.Null{}
		}
	case "shift":
		if len(arr.Elements) > 0 {
			result = arr.Elements[0]
			arr.Elements = arr.Elements[1:]
		} else {
			result = &interpreter.Null{}
		}
	case "unshift":
		if len(args) > 0 {
			arr.Elements = append([]interpreter.Object{args[0]}, arr.Elements...)
			result = &interpreter.Integer{Value: int64(len(arr.Elements))}
		}
	case "join":
		sep := ""
		if len(args) > 0 {
			if s, ok := args[0].(*interpreter.String); ok {
				sep = s.Value
			}
		}
		parts := make([]string, len(arr.Elements))
		for i, elem := range arr.Elements {
			parts[i] = elem.Inspect()
		}
		result = &interpreter.String{Value: strings.Join(parts, sep)}
	case "indexOf":
		if len(args) > 0 {
			for i, elem := range arr.Elements {
				if vm.isEqual(elem, args[0]) {
					result = &interpreter.Integer{Value: int64(i)}
					break
				}
			}
			if result == nil {
				result = &interpreter.Integer{Value: -1}
			}
		}
	case "contains":
		if len(args) > 0 {
			found := false
			for _, elem := range arr.Elements {
				if vm.isEqual(elem, args[0]) {
					found = true
					break
				}
			}
			result = &interpreter.Boolean{Value: found}
		}
	case "reverse":
		reversed := make([]interpreter.Object, len(arr.Elements))
		for i, elem := range arr.Elements {
			reversed[len(arr.Elements)-1-i] = elem
		}
		result = &interpreter.Array{Elements: reversed, ElementType: arr.ElementType}
	case "slice":
		start := 0
		end := len(arr.Elements)
		if len(args) >= 1 {
			if s, ok := args[0].(*interpreter.Integer); ok {
				start = int(s.Value)
			}
		}
		if len(args) >= 2 {
			if e, ok := args[1].(*interpreter.Integer); ok {
				end = int(e.Value)
			}
		}
		if start < 0 {
			start = 0
		}
		if end > len(arr.Elements) {
			end = len(arr.Elements)
		}
		if start > end {
			start = end
		}
		result = &interpreter.Array{Elements: arr.Elements[start:end], ElementType: arr.ElementType}
	default:
		return fmt.Errorf("数组没有方法: %s", name)
	}

	if result == nil {
		result = &interpreter.Null{}
	}
	vm.push(result)
	return nil
}

// ========== Map 方法 ==========

// invokeMapMethod 调用 Map 方法
func (vm *VM) invokeMapMethod(m *interpreter.Map, name string, argCount int) error {
	args := make([]interpreter.Object, argCount)
	for i := argCount - 1; i >= 0; i-- {
		args[i] = vm.pop()
	}
	vm.pop() // 弹出 Map 本身

	var result interpreter.Object

	switch name {
	case "length", "size":
		result = &interpreter.Integer{Value: int64(len(m.Pairs))}
	case "keys":
		elements := make([]interpreter.Object, len(m.Keys))
		for i, k := range m.Keys {
			elements[i] = &interpreter.String{Value: k}
		}
		result = &interpreter.Array{Elements: elements, ElementType: "string"}
	case "values":
		elements := make([]interpreter.Object, len(m.Keys))
		for i, k := range m.Keys {
			elements[i] = m.Pairs[k]
		}
		result = &interpreter.Array{Elements: elements}
	case "has", "containsKey":
		if len(args) > 0 {
			if key, ok := args[0].(*interpreter.String); ok {
				_, exists := m.Pairs[key.Value]
				result = &interpreter.Boolean{Value: exists}
			}
		}
	case "get":
		if len(args) > 0 {
			if key, ok := args[0].(*interpreter.String); ok {
				if value, exists := m.Pairs[key.Value]; exists {
					result = value
				} else if len(args) > 1 {
					result = args[1] // 默认值
				} else {
					result = &interpreter.Null{}
				}
			}
		}
	case "set":
		if len(args) >= 2 {
			if key, ok := args[0].(*interpreter.String); ok {
				if _, exists := m.Pairs[key.Value]; !exists {
					m.Keys = append(m.Keys, key.Value)
				}
				m.Pairs[key.Value] = args[1]
				result = m
			}
		}
	case "delete", "remove":
		if len(args) > 0 {
			if key, ok := args[0].(*interpreter.String); ok {
				if _, exists := m.Pairs[key.Value]; exists {
					delete(m.Pairs, key.Value)
					// 从 keys 中删除
					for i, k := range m.Keys {
						if k == key.Value {
							m.Keys = append(m.Keys[:i], m.Keys[i+1:]...)
							break
						}
					}
					result = &interpreter.Boolean{Value: true}
				} else {
					result = &interpreter.Boolean{Value: false}
				}
			}
		}
	case "clear":
		m.Pairs = make(map[string]interpreter.Object)
		m.Keys = []string{}
		result = &interpreter.Null{}
	default:
		return fmt.Errorf("Map 没有方法: %s", name)
	}

	if result == nil {
		result = &interpreter.Null{}
	}
	vm.push(result)
	return nil
}

// ========== 索引操作 ==========

// indexGet 获取索引值
func (vm *VM) indexGet(obj, index interpreter.Object) (interpreter.Object, error) {
	switch o := obj.(type) {
	case *interpreter.Array:
		if idx, ok := index.(*interpreter.Integer); ok {
			i := int(idx.Value)
			if i < 0 {
				i = len(o.Elements) + i
			}
			if i < 0 || i >= len(o.Elements) {
				return nil, fmt.Errorf("数组索引越界: %d", idx.Value)
			}
			return o.Elements[i], nil
		}
		return nil, fmt.Errorf("数组索引必须是整数")

	case *interpreter.Map:
		if key, ok := index.(*interpreter.String); ok {
			if value, exists := o.Pairs[key.Value]; exists {
				return value, nil
			}
			return &interpreter.Null{}, nil
		}
		return nil, fmt.Errorf("Map 键必须是字符串")

	case *interpreter.String:
		if idx, ok := index.(*interpreter.Integer); ok {
			runes := []rune(o.Value)
			i := int(idx.Value)
			if i < 0 {
				i = len(runes) + i
			}
			if i < 0 || i >= len(runes) {
				return nil, fmt.Errorf("字符串索引越界: %d", idx.Value)
			}
			return &interpreter.String{Value: string(runes[i])}, nil
		}
		return nil, fmt.Errorf("字符串索引必须是整数")
	}

	return nil, fmt.Errorf("不支持索引访问的类型: %s", obj.Type())
}

// sliceGet 切片操作
func (vm *VM) sliceGet(obj, startIdx, endIdx interpreter.Object) (interpreter.Object, error) {
	switch o := obj.(type) {
	case *interpreter.Array:
		start := 0
		end := len(o.Elements)

		// 处理起始索引
		if _, ok := startIdx.(*interpreter.Null); !ok {
			if idx, ok := startIdx.(*interpreter.Integer); ok {
				start = int(idx.Value)
				if start < 0 {
					start = len(o.Elements) + start
				}
			} else {
				return nil, fmt.Errorf("切片起始索引必须是整数或null")
			}
		}

		// 处理结束索引
		if _, ok := endIdx.(*interpreter.Null); !ok {
			if idx, ok := endIdx.(*interpreter.Integer); ok {
				end = int(idx.Value)
				if end < 0 {
					end = len(o.Elements) + end
				}
			} else {
				return nil, fmt.Errorf("切片结束索引必须是整数或null")
			}
		}

		// 边界检查
		if start < 0 {
			start = 0
		}
		if end > len(o.Elements) {
			end = len(o.Elements)
		}
		if start > end {
			start = end
		}

		// 创建新数组
		newElements := make([]interpreter.Object, end-start)
		copy(newElements, o.Elements[start:end])
		return &interpreter.Array{Elements: newElements, ElementType: o.ElementType}, nil

	case *interpreter.String:
		runes := []rune(o.Value)
		start := 0
		end := len(runes)

		// 处理起始索引
		if _, ok := startIdx.(*interpreter.Null); !ok {
			if idx, ok := startIdx.(*interpreter.Integer); ok {
				start = int(idx.Value)
				if start < 0 {
					start = len(runes) + start
				}
			} else {
				return nil, fmt.Errorf("切片起始索引必须是整数或null")
			}
		}

		// 处理结束索引
		if _, ok := endIdx.(*interpreter.Null); !ok {
			if idx, ok := endIdx.(*interpreter.Integer); ok {
				end = int(idx.Value)
				if end < 0 {
					end = len(runes) + end
				}
			} else {
				return nil, fmt.Errorf("切片结束索引必须是整数或null")
			}
		}

		// 边界检查
		if start < 0 {
			start = 0
		}
		if end > len(runes) {
			end = len(runes)
		}
		if start > end {
			start = end
		}

		return &interpreter.String{Value: string(runes[start:end])}, nil
	}

	return nil, fmt.Errorf("不支持切片操作的类型: %s", obj.Type())
}

// indexSet 设置索引值
func (vm *VM) indexSet(obj, index, value interpreter.Object) error {
	switch o := obj.(type) {
	case *interpreter.Array:
		if idx, ok := index.(*interpreter.Integer); ok {
			i := int(idx.Value)
			if i < 0 {
				i = len(o.Elements) + i
			}
			if i < 0 || i >= len(o.Elements) {
				return fmt.Errorf("数组索引越界: %d", idx.Value)
			}
			o.Elements[i] = value
			return nil
		}
		return fmt.Errorf("数组索引必须是整数")

	case *interpreter.Map:
		if key, ok := index.(*interpreter.String); ok {
			if _, exists := o.Pairs[key.Value]; !exists {
				o.Keys = append(o.Keys, key.Value)
			}
			o.Pairs[key.Value] = value
			return nil
		}
		return fmt.Errorf("Map 键必须是字符串")
	}

	return fmt.Errorf("不支持索引赋值的类型: %s", obj.Type())
}

// ========== 协程 ==========

// runGoroutine 运行协程
func (vm *VM) runGoroutine(closure *Closure, argCount int) {
	// 创建新的虚拟机实例
	newVM := NewVM()

	// 复制参数
	args := make([]interpreter.Object, argCount)
	for i := 0; i < argCount; i++ {
		args[i] = vm.stack[vm.sp-argCount+i]
	}

	// 设置参数
	for i, arg := range args {
		newVM.push(arg)
		_ = i
	}

	// 执行闭包
	newVM.pushFrame(closure, 0)
	newVM.execute()
}

// ========== 辅助函数 ==========

// objectToString 将对象转换为字符串
func (vm *VM) objectToString(obj interpreter.Object) string {
	switch v := obj.(type) {
	case *interpreter.String:
		return v.Value
	case *interpreter.Integer:
		return fmt.Sprintf("%d", v.Value)
	case *interpreter.Float:
		return fmt.Sprintf("%g", v.Value)
	case *interpreter.Boolean:
		if v.Value {
			return "true"
		}
		return "false"
	case *interpreter.Null:
		return "null"
	default:
		return obj.Inspect()
	}
}
