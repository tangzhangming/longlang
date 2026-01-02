package interpreter

import (
	"github.com/tangzhangming/longlang/internal/parser"
)

// evalEnumStatement 执行枚举定义语句
func (i *Interpreter) evalEnumStatement(node *parser.EnumStatement) Object {
	enum := &Enum{
		Name:       node.Name.Value,
		Members:    make(map[string]*EnumValue),
		MemberList: []*EnumValue{},
		Methods:    make(map[string]*ClassMethod),
		Variables:  make(map[string]*ClassVariable),
		Env:        i.env,
	}

	// 设置底层类型
	if node.BackingType != nil {
		enum.BackingType = node.BackingType.Value
	}

	// 处理接口实现
	if len(node.Interfaces) > 0 {
		enum.Interfaces = make([]*Interface, 0, len(node.Interfaces))
		for _, ifaceName := range node.Interfaces {
			ifaceObj, ok := i.env.Get(ifaceName.Value)
			if !ok {
				return newError("未定义的接口: %s", ifaceName.Value)
			}
			iface, ok := ifaceObj.(*Interface)
			if !ok {
				return newError("%s 不是一个接口", ifaceName.Value)
			}
			enum.Interfaces = append(enum.Interfaces, iface)
		}
	}

	// 处理字段
	for _, variable := range node.Variables {
		var defaultValue Object
		if variable.Value != nil {
			defaultValue = i.Eval(variable.Value)
		}
		enum.Variables[variable.Name.Value] = &ClassVariable{
			Name:           variable.Name.Value,
			Type:           variable.Type.Value,
			AccessModifier: variable.AccessModifier,
			DefaultValue:   defaultValue,
		}
	}

	// 处理方法
	for _, method := range node.Methods {
		returnTypes := []string{}
		for _, rt := range method.ReturnType {
			returnTypes = append(returnTypes, rt.Value)
		}
		enum.Methods[method.Name.Value] = &ClassMethod{
			Name:           method.Name.Value,
			AccessModifier: method.AccessModifier,
			IsStatic:       method.IsStatic,
			Parameters:     toInterfaceSlice(method.Parameters),
			ReturnType:     returnTypes,
			Body:           method.Body,
			Env:            i.env,
		}
	}

	// 处理枚举成员
	for ordinal, member := range node.Members {
		enumValue := &EnumValue{
			Enum:    enum,
			Name:    member.Name.Value,
			Ordinal: ordinal,
			Fields:  make(map[string]Object),
		}

		// 处理成员值
		if member.Value != nil {
			val := i.Eval(member.Value)
			if isError(val) {
				return val
			}
			enumValue.Value = val
		} else if enum.BackingType == "" {
			// 简单枚举，没有值
			enumValue.Value = nil
		}

		// 处理构造参数（复杂枚举）
		if len(member.Arguments) > 0 {
			// 复杂枚举需要字段和构造函数
			// 这里简化处理：直接将参数存储到字段中
			for idx, arg := range member.Arguments {
				val := i.Eval(arg)
				if isError(val) {
					return val
				}
				// 按顺序匹配字段
				fieldIdx := 0
				for name := range enum.Variables {
					if fieldIdx == idx {
						enumValue.Fields[name] = val
						break
					}
					fieldIdx++
				}
			}
		}

		enum.Members[member.Name.Value] = enumValue
		enum.MemberList = append(enum.MemberList, enumValue)
	}

	// 注册到环境
	if i.currentNamespace != nil {
		i.currentNamespace.SetClass(node.Name.Value, &Class{Name: node.Name.Value})
	}
	i.env.Set(node.Name.Value, enum)

	return enum
}

// evalEnumAccess 评估枚举成员访问 (EnumName::MemberName)
func (i *Interpreter) evalEnumAccess(enumName string, memberName string) Object {
	// 查找枚举
	enumObj, ok := i.env.Get(enumName)
	if !ok {
		return newError("未定义的枚举: %s", enumName)
	}

	enum, ok := enumObj.(*Enum)
	if !ok {
		return newError("%s 不是一个枚举", enumName)
	}

	// 查找成员
	member, ok := enum.GetMember(memberName)
	if !ok {
		// 可能是静态方法调用
		method, hasMethod := enum.GetMethod(memberName)
		if hasMethod && method.IsStatic {
			return &BoundEnumMethod{
				EnumValue: nil,
				Method:    method,
				Enum:      enum,
			}
		}
		return newError("枚举 %s 没有成员 %s", enumName, memberName)
	}

	return member
}

// evalEnumMethodCall 评估枚举方法调用
func (i *Interpreter) evalEnumMethodCall(enumValue *EnumValue, methodName string, args []Object) Object {
	enum := enumValue.Enum

	// 内置方法
	switch methodName {
	case "name":
		return &String{Value: enumValue.Name}
	case "ordinal":
		return &Integer{Value: int64(enumValue.Ordinal)}
	case "value":
		if enumValue.Value != nil {
			return enumValue.Value
		}
		return newError("简单枚举没有 value() 方法，请使用带值枚举")
	}

	// 自定义方法
	method, ok := enum.GetMethod(methodName)
	if !ok {
		return newError("枚举 %s 没有方法 %s", enum.Name, methodName)
	}

	// 创建方法执行环境
	methodEnv := NewEnclosedEnvironment(enum.Env)
	methodEnv.Set("this", enumValue)

	// 设置字段访问
	for name, value := range enumValue.Fields {
		methodEnv.Set(name, value)
	}

	// 绑定参数
	params := method.Parameters
	for idx, param := range params {
		p := param.(*parser.FunctionParameter)
		if idx < len(args) {
			methodEnv.Set(p.Name.Value, args[idx])
		} else if p.DefaultValue != nil {
			val := i.Eval(p.DefaultValue)
			methodEnv.Set(p.Name.Value, val)
		}
	}

	// 执行方法
	oldEnv := i.env
	i.env = methodEnv

	body, ok := method.Body.(*parser.BlockStatement)
	if !ok {
		i.env = oldEnv
		return newError("枚举方法体无效")
	}

	result := i.Eval(body)
	i.env = oldEnv

	// 处理返回值
	if returnValue, ok := result.(*ReturnValue); ok {
		return returnValue.Value
	}

	return result
}

// evalEnumStaticMethodCall 评估枚举静态方法调用
func (i *Interpreter) evalEnumStaticMethodCall(enum *Enum, methodName string, args []Object) Object {
	// 内置静态方法
	switch methodName {
	case "cases":
		// 返回所有成员的数组
		elements := make([]Object, len(enum.MemberList))
		for idx, member := range enum.MemberList {
			elements[idx] = member
		}
		return &Array{Elements: elements}

	case "count":
		return &Integer{Value: int64(len(enum.MemberList))}

	case "from":
		// 从值创建枚举（无效值抛异常）
		if len(args) != 1 {
			return newError("from() 需要1个参数")
		}
		for _, member := range enum.MemberList {
			if member.Value != nil && objectsEqual(member.Value, args[0]) {
				return member
			}
		}
		return newError("无效的枚举值: %s，枚举 %s 没有此值", args[0].Inspect(), enum.Name)

	case "tryFrom":
		// 从值创建枚举（无效值返回 null）
		if len(args) != 1 {
			return newError("tryFrom() 需要1个参数")
		}
		for _, member := range enum.MemberList {
			if member.Value != nil && objectsEqual(member.Value, args[0]) {
				return member
			}
		}
		return &Null{}

	case "valueOf":
		// 从名称创建枚举
		if len(args) != 1 {
			return newError("valueOf() 需要1个参数")
		}
		nameStr, ok := args[0].(*String)
		if !ok {
			return newError("valueOf() 参数必须是字符串")
		}
		member, found := enum.GetMember(nameStr.Value)
		if !found {
			return newError("无效的枚举名称: %s，枚举 %s 没有此成员", nameStr.Value, enum.Name)
		}
		return member
	}

	// 自定义静态方法
	method, ok := enum.GetMethod(methodName)
	if !ok || !method.IsStatic {
		return newError("枚举 %s 没有静态方法 %s", enum.Name, methodName)
	}

	// 执行静态方法
	methodEnv := NewEnclosedEnvironment(enum.Env)

	params := method.Parameters
	for idx, param := range params {
		p := param.(*parser.FunctionParameter)
		if idx < len(args) {
			methodEnv.Set(p.Name.Value, args[idx])
		} else if p.DefaultValue != nil {
			val := i.Eval(p.DefaultValue)
			methodEnv.Set(p.Name.Value, val)
		}
	}

	oldEnv := i.env
	i.env = methodEnv

	body, ok := method.Body.(*parser.BlockStatement)
	if !ok {
		i.env = oldEnv
		return newError("枚举静态方法体无效")
	}

	result := i.Eval(body)
	i.env = oldEnv

	if returnValue, ok := result.(*ReturnValue); ok {
		return returnValue.Value
	}

	return result
}

