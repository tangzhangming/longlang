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


