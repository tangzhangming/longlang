package interpreter

import (
	"github.com/tangzhangming/longlang/internal/parser"
)

// AnnotationDef 注解定义
type AnnotationDef struct {
	Name        string                     // 注解名称
	Fields      map[string]*AnnotationFieldDef // 字段定义
	FieldOrder  []string                   // 字段顺序
	Annotations map[string]*AnnotationInstance // 元注解
}

func (ad *AnnotationDef) Type() ObjectType { return "ANNOTATION_DEF" }
func (ad *AnnotationDef) Inspect() string  { return "annotation " + ad.Name }

// AnnotationFieldDef 注解字段定义
type AnnotationFieldDef struct {
	Name         string // 字段名
	Type         string // 字段类型
	DefaultValue Object // 默认值
}

// AnnotationInstance 注解实例（应用到类/方法/字段上的注解）
type AnnotationInstance struct {
	Name      string            // 注解名称
	Arguments map[string]Object // 参数值
}

func (ai *AnnotationInstance) Type() ObjectType { return "ANNOTATION_INSTANCE" }
func (ai *AnnotationInstance) Inspect() string  { return "@" + ai.Name }

// evalAnnotationDefinition 执行注解定义
func (i *Interpreter) evalAnnotationDefinition(node *parser.AnnotationDefinition) Object {
	def := &AnnotationDef{
		Name:        node.Name.Value,
		Fields:      make(map[string]*AnnotationFieldDef),
		FieldOrder:  []string{},
		Annotations: make(map[string]*AnnotationInstance),
	}

	// 处理注解字段
	for _, field := range node.Fields {
		fieldDef := &AnnotationFieldDef{
			Name: field.Name.Value,
			Type: field.Type.String(),
		}

		// 处理默认值
		if field.DefaultValue != nil {
			val := i.Eval(field.DefaultValue)
			if !isError(val) && !isThrownException(val) {
				fieldDef.DefaultValue = val
			}
		}

		def.Fields[field.Name.Value] = fieldDef
		def.FieldOrder = append(def.FieldOrder, field.Name.Value)
	}

	// 处理元注解
	for _, ann := range node.Annotations {
		instance := i.evalAnnotationInstance(ann)
		if instance != nil {
			def.Annotations[ann.Name.Value] = instance
		}
	}

	// 注册注解定义
	i.env.Set(node.Name.Value, def)
	
	// 同时注册到全局注解定义表
	if i.annotationDefs == nil {
		i.annotationDefs = make(map[string]*AnnotationDef)
	}
	i.annotationDefs[node.Name.Value] = def

	return &Null{}
}

// evalAnnotationInstance 创建注解实例
func (i *Interpreter) evalAnnotationInstance(ann *parser.Annotation) *AnnotationInstance {
	instance := &AnnotationInstance{
		Name:      ann.Name.Value,
		Arguments: make(map[string]Object),
	}

	// 处理注解参数
	for key, expr := range ann.Arguments {
		val := i.Eval(expr)
		if !isError(val) && !isThrownException(val) {
			instance.Arguments[key] = val
		}
	}

	return instance
}

// convertAnnotationsToInstances 将 AST 注解转换为注解实例
func (i *Interpreter) convertAnnotationsToInstances(annotations []*parser.Annotation) []*AnnotationInstance {
	if annotations == nil {
		return nil
	}

	instances := make([]*AnnotationInstance, 0, len(annotations))
	for _, ann := range annotations {
		instance := i.evalAnnotationInstance(ann)
		if instance != nil {
			instances = append(instances, instance)
		}
	}
	return instances
}

// registerAnnotationBuiltins 注册注解相关的内置函数
func registerAnnotationBuiltins(env *Environment) {
	// __get_class_annotations - 获取类的注解
	env.Set("__get_class_annotations", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__get_class_annotations 需要1个参数")
		}
		
		className, ok := args[0].(*String)
		if !ok {
			return newError("__get_class_annotations 参数必须是字符串（类名）")
		}
		
		// 从当前环境获取类
		classObj, ok := globalEnv.Get(className.Value)
		if !ok {
			return &Array{Elements: []Object{}}
		}
		
		class, ok := classObj.(*Class)
		if !ok {
			return &Array{Elements: []Object{}}
		}
		
		// 返回注解数组
		return annotationsToArray(class.Annotations)
	}})

	// __has_annotation - 检查是否有指定注解
	env.Set("__has_annotation", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__has_annotation 需要2个参数")
		}
		
		// 第一个参数是注解数组
		annArray, ok := args[0].(*Array)
		if !ok {
			return &Boolean{Value: false}
		}
		
		// 第二个参数是注解名称
		annName, ok := args[1].(*String)
		if !ok {
			return newError("__has_annotation 第二个参数必须是字符串（注解名）")
		}
		
		// 查找注解
		for _, elem := range annArray.Elements {
			if annMap, ok := elem.(*Map); ok {
				if nameObj, ok := annMap.Get("name"); ok {
					if nameStr, ok := nameObj.(*String); ok {
						if nameStr.Value == annName.Value {
							return &Boolean{Value: true}
						}
					}
				}
			}
		}
		
		return &Boolean{Value: false}
	}})

	// __get_annotation - 获取指定名称的注解
	env.Set("__get_annotation", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__get_annotation 需要2个参数")
		}
		
		// 第一个参数是注解数组
		annArray, ok := args[0].(*Array)
		if !ok {
			return &Null{}
		}
		
		// 第二个参数是注解名称
		annName, ok := args[1].(*String)
		if !ok {
			return newError("__get_annotation 第二个参数必须是字符串（注解名）")
		}
		
		// 查找注解
		for _, elem := range annArray.Elements {
			if annMap, ok := elem.(*Map); ok {
				if nameObj, ok := annMap.Get("name"); ok {
					if nameStr, ok := nameObj.(*String); ok {
						if nameStr.Value == annName.Value {
							return annMap
						}
					}
				}
			}
		}
		
		return &Null{}
	}})

	// __get_class_fields - 获取类的字段列表
	env.Set("__get_class_fields", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__get_class_fields 需要1个参数")
		}

		className, ok := args[0].(*String)
		if !ok {
			return newError("__get_class_fields 参数必须是字符串（类名）")
		}

		// 从全局环境获取类
		classObj, ok := globalEnv.Get(className.Value)
		if !ok {
			return &Map{Pairs: make(map[string]Object), Keys: []string{}, KeyType: "string", ValueType: "any"}
		}

		class, ok := classObj.(*Class)
		if !ok {
			return &Map{Pairs: make(map[string]Object), Keys: []string{}, KeyType: "string", ValueType: "any"}
		}

		// 返回字段信息
		fieldsMap := &Map{
			Pairs:     make(map[string]Object),
			Keys:      []string{},
			KeyType:   "string",
			ValueType: "any",
		}

		for name, field := range class.Variables {
			fieldInfo := &Map{
				Pairs:     make(map[string]Object),
				Keys:      []string{},
				KeyType:   "string",
				ValueType: "any",
			}
			fieldInfo.Set("type", &String{Value: field.Type})
			fieldInfo.Set("access", &String{Value: field.AccessModifier})
			fieldsMap.Set(name, fieldInfo)
		}

		return fieldsMap
	}})

	// __get_field_annotations - 获取字段的注解列表
	env.Set("__get_field_annotations", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__get_field_annotations 需要2个参数")
		}

		className, ok := args[0].(*String)
		if !ok {
			return newError("__get_field_annotations 第一个参数必须是字符串（类名）")
		}

		fieldName, ok := args[1].(*String)
		if !ok {
			return newError("__get_field_annotations 第二个参数必须是字符串（字段名）")
		}

		// 从全局环境获取类
		classObj, ok := globalEnv.Get(className.Value)
		if !ok {
			return &Array{Elements: []Object{}}
		}

		class, ok := classObj.(*Class)
		if !ok {
			return &Array{Elements: []Object{}}
		}

		// 获取字段
		field, ok := class.Variables[fieldName.Value]
		if !ok {
			return &Array{Elements: []Object{}}
		}

		// 返回字段注解
		return annotationsToArray(field.Annotations)
	}})

	// __has_field_annotation - 检查字段是否有指定注解
	env.Set("__has_field_annotation", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("__has_field_annotation 需要3个参数")
		}

		className, ok := args[0].(*String)
		if !ok {
			return &Boolean{Value: false}
		}

		fieldName, ok := args[1].(*String)
		if !ok {
			return &Boolean{Value: false}
		}

		annName, ok := args[2].(*String)
		if !ok {
			return &Boolean{Value: false}
		}

		// 从全局环境获取类
		classObj, ok := globalEnv.Get(className.Value)
		if !ok {
			return &Boolean{Value: false}
		}

		class, ok := classObj.(*Class)
		if !ok {
			return &Boolean{Value: false}
		}

		// 获取字段
		field, ok := class.Variables[fieldName.Value]
		if !ok {
			return &Boolean{Value: false}
		}

		// 检查注解
		for _, ann := range field.Annotations {
			if ann.Name == annName.Value {
				return &Boolean{Value: true}
			}
		}

		return &Boolean{Value: false}
	}})

	// __get_field_annotation - 获取字段的指定注解
	env.Set("__get_field_annotation", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("__get_field_annotation 需要3个参数")
		}

		className, ok := args[0].(*String)
		if !ok {
			return &Null{}
		}

		fieldName, ok := args[1].(*String)
		if !ok {
			return &Null{}
		}

		annName, ok := args[2].(*String)
		if !ok {
			return &Null{}
		}

		// 从全局环境获取类
		classObj, ok := globalEnv.Get(className.Value)
		if !ok {
			return &Null{}
		}

		class, ok := classObj.(*Class)
		if !ok {
			return &Null{}
		}

		// 获取字段
		field, ok := class.Variables[fieldName.Value]
		if !ok {
			return &Null{}
		}

		// 查找注解
		for _, ann := range field.Annotations {
			if ann.Name == annName.Value {
				annMap := &Map{
					Pairs:     make(map[string]Object),
					Keys:      []string{},
					KeyType:   "string",
					ValueType: "any",
				}
				annMap.Set("name", &String{Value: ann.Name})

				argsMap := &Map{
					Pairs:     make(map[string]Object),
					Keys:      []string{},
					KeyType:   "string",
					ValueType: "any",
				}
				for key, val := range ann.Arguments {
					argsMap.Set(key, val)
				}
				annMap.Set("arguments", argsMap)

				return annMap
			}
		}

		return &Null{}
	}})

	// __new_instance - 创建类的新实例（不调用构造函数）
	env.Set("__new_instance", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__new_instance 需要1个参数")
		}

		className, ok := args[0].(*String)
		if !ok {
			return newError("__new_instance 参数必须是字符串（类名）")
		}

		// 从全局环境获取类
		classObj, ok := globalEnv.Get(className.Value)
		if !ok {
			return newError("未定义的类: %s", className.Value)
		}

		class, ok := classObj.(*Class)
		if !ok {
			return newError("%s 不是一个类", className.Value)
		}

		// 创建实例
		instance := &Instance{
			Class:  class,
			Fields: make(map[string]Object),
		}

		// 初始化字段默认值
		for name, field := range class.Variables {
			if field.DefaultValue != nil {
				instance.Fields[name] = field.DefaultValue
			} else {
				instance.Fields[name] = &Null{}
			}
		}

		return instance
	}})

	// __get_field_value - 获取实例字段的值
	env.Set("__get_field_value", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__get_field_value 需要2个参数")
		}

		instance, ok := args[0].(*Instance)
		if !ok {
			return newError("__get_field_value 第一个参数必须是类实例")
		}

		fieldName, ok := args[1].(*String)
		if !ok {
			return newError("__get_field_value 第二个参数必须是字符串（字段名）")
		}

		if val, ok := instance.Fields[fieldName.Value]; ok {
			return val
		}

		return &Null{}
	}})

	// __set_field_value - 设置实例字段的值
	env.Set("__set_field_value", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("__set_field_value 需要3个参数")
		}

		instance, ok := args[0].(*Instance)
		if !ok {
			return newError("__set_field_value 第一个参数必须是类实例")
		}

		fieldName, ok := args[1].(*String)
		if !ok {
			return newError("__set_field_value 第二个参数必须是字符串（字段名）")
		}

		instance.Fields[fieldName.Value] = args[2]
		return &Null{}
	}})

	// __create_instance - 创建类实例并调用构造函数
	env.Set("__create_instance", &Builtin{Fn: func(args ...Object) Object {
		if len(args) < 1 {
			return newError("__create_instance 需要至少1个参数")
		}

		className, ok := args[0].(*String)
		if !ok {
			return newError("__create_instance 第一个参数必须是字符串（类名）")
		}

		if globalInterpreter == nil {
			return newError("解释器未初始化")
		}

		// 使用解释器创建实例
		return globalInterpreter.CreateInstance(className.Value, args[1:])
	}})
}

// annotationsToArray 将注解实例列表转换为数组
func annotationsToArray(annotations []*AnnotationInstance) *Array {
	elements := make([]Object, 0, len(annotations))
	
	for _, ann := range annotations {
		annMap := &Map{
			Pairs:     make(map[string]Object),
			Keys:      []string{},
			KeyType:   "string",
			ValueType: "any",
		}
		
		// 设置注解名称
		annMap.Set("name", &String{Value: ann.Name})
		
		// 设置注解参数
		argsMap := &Map{
			Pairs:     make(map[string]Object),
			Keys:      []string{},
			KeyType:   "string",
			ValueType: "any",
		}
		for key, val := range ann.Arguments {
			argsMap.Set(key, val)
		}
		annMap.Set("arguments", argsMap)
		
		elements = append(elements, annMap)
	}
	
	return &Array{Elements: elements}
}

// 全局环境引用（用于内置函数访问）
var globalEnv *Environment

// 全局解释器引用（用于创建实例）
var globalInterpreter *Interpreter

// SetGlobalInterpreter 设置全局解释器引用
func SetGlobalInterpreter(i *Interpreter) {
	globalInterpreter = i
}



