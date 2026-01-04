package vm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tangzhangming/longlang/internal/config"
	"github.com/tangzhangming/longlang/internal/interpreter"
	"github.com/tangzhangming/longlang/internal/lexer"
	"github.com/tangzhangming/longlang/internal/parser"
)


// ========== 虚拟机常量 ==========

const (
	StackSize     = 2048  // 操作数栈大小
	FrameSize     = 1024  // 调用栈大小
	GlobalsSize   = 65536 // 全局变量数量上限
	MaxTryDepth   = 64    // try 块嵌套深度上限
)

// ========== 虚拟机 ==========

// VM 虚拟机
type VM struct {
	// 栈相关
	stack    []interpreter.Object // 操作数栈
	sp       int                  // 栈指针（指向下一个空闲位置）

	// 调用栈
	frames      []*Frame // 调用栈帧
	frameCount  int      // 当前帧数量

	// 全局变量
	globals map[string]interpreter.Object

	// 内置函数和对象
	builtins *interpreter.Environment

	// 当前执行的字节码
	bytecode *Bytecode

	// 异常处理
	tryStack   []*TryState // try 块栈
	tryCount   int         // 当前 try 块数量
	exception  interpreter.Object // 当前异常

	// 开放的 upvalue 链表
	openUpvalues *Upvalue

	// 项目配置
	projectRoot   string
	projectConfig *config.ProjectConfig

	// 命名空间管理
	namespaceMgr      *interpreter.NamespaceManager
	currentNamespace  *interpreter.Namespace
	loadedNamespaces  map[string]bool // 已加载的命名空间文件缓存
	loadingNamespaces map[string]bool // 正在加载中的命名空间（用于循环依赖检测）

	// 标准库路径
	stdlibPath string

	// 调试信息
	debug bool
}

// NewVM 创建新的虚拟机
func NewVM() *VM {
	vm := &VM{
		stack:             make([]interpreter.Object, StackSize),
		sp:                0,
		frames:            make([]*Frame, FrameSize),
		frameCount:        0,
		globals:           make(map[string]interpreter.Object),
		builtins:          interpreter.NewEnvironment(),
		tryStack:          make([]*TryState, MaxTryDepth),
		tryCount:          0,
		namespaceMgr:      interpreter.NewNamespaceManager(),
		loadedNamespaces:  make(map[string]bool),
		loadingNamespaces: make(map[string]bool),
		stdlibPath:        "stdlib",
		debug:             false,
	}

	// 初始化默认命名空间
	vm.currentNamespace = vm.namespaceMgr.GetNamespace("")

	// 初始化内置函数
	vm.initBuiltins()

	return vm
}

// initBuiltins 初始化内置函数
func (vm *VM) initBuiltins() {
	// 使用解释器的内置函数注册系统
	builtins := interpreter.GetAllBuiltins()
	
	// 将所有内置函数复制到 VM 的全局变量表
	for name, value := range builtins {
		vm.globals[name] = value
	}

	// 注册 VM 专有的反射内置函数，覆盖解释器的版本（如果存在）
	vm.registerVMReflectionBuiltins()
}

// registerVMReflectionBuiltins 注册 VM 专用的反射内置函数
func (vm *VM) registerVMReflectionBuiltins() {
	// __get_class_annotations(className)
	vm.globals["__get_class_annotations"] = &interpreter.Builtin{Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) != 1 {
			return &interpreter.Error{Message: "__get_class_annotations 需要1个参数"}
		}
		className, ok := args[0].(*interpreter.String)
		if !ok {
			return &interpreter.Error{Message: "__get_class_annotations 参数必须是字符串"}
		}
		class, ok := vm.getClassByName(className.Value)
		if !ok {
			return &interpreter.Array{Elements: []interpreter.Object{}}
		}
		return vm.annotationsToArray(class.Annotations)
	}}

	// __get_class_fields(className)
	vm.globals["__get_class_fields"] = &interpreter.Builtin{Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) != 1 {
			return &interpreter.Error{Message: "__get_class_fields 需要1个参数"}
		}
		className, ok := args[0].(*interpreter.String)
		if !ok {
			return &interpreter.Error{Message: "__get_class_fields 参数必须是字符串"}
		}
		class, ok := vm.getClassByName(className.Value)
		if !ok {
			return &interpreter.Map{Pairs: make(map[string]interpreter.Object), Keys: []string{}}
		}
		
		fields := make(map[string]interpreter.Object)
		keys := make([]string, 0)
		for name, variable := range class.Variables {
			fieldInfo := &interpreter.Map{
				Pairs: map[string]interpreter.Object{
					"name": &interpreter.String{Value: name},
					"type": &interpreter.String{Value: variable.Type},
				},
				Keys: []string{"name", "type"},
			}
			fields[name] = fieldInfo
			keys = append(keys, name)
		}
		return &interpreter.Map{Pairs: fields, Keys: keys}
	}}

	// __get_field_annotation(className, fieldName, annName)
	vm.globals["__get_field_annotation"] = &interpreter.Builtin{Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) != 3 {
			return &interpreter.Error{Message: "__get_field_annotation 需要3个参数"}
		}
		className, ok1 := args[0].(*interpreter.String)
		fieldName, ok2 := args[1].(*interpreter.String)
		annName, ok3 := args[2].(*interpreter.String)
		if !ok1 || !ok2 || !ok3 {
			return &interpreter.Error{Message: "__get_field_annotation 参数必须是字符串"}
		}
		
		class, ok := vm.getClassByName(className.Value)
		if !ok {
			return &interpreter.Null{}
		}
		
		variable, ok := class.Variables[fieldName.Value]
		if !ok {
			return &interpreter.Null{}
		}
		
		for _, ann := range variable.Annotations {
			if ann.Name == annName.Value {
				return vm.annotationToMap(ann)
			}
		}
		return &interpreter.Null{}
	}}

	// __has_field_annotation(className, fieldName, annName)
	vm.globals["__has_field_annotation"] = &interpreter.Builtin{Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) != 3 {
			return &interpreter.Error{Message: "__has_field_annotation 需要3个参数"}
		}
		className, ok1 := args[0].(*interpreter.String)
		fieldName, ok2 := args[1].(*interpreter.String)
		annName, ok3 := args[2].(*interpreter.String)
		if !ok1 || !ok2 || !ok3 {
			return &interpreter.Error{Message: "__has_field_annotation 参数必须是字符串"}
		}
		
		class, ok := vm.getClassByName(className.Value)
		if !ok {
			return &interpreter.Boolean{Value: false}
		}
		
		variable, ok := class.Variables[fieldName.Value]
		if !ok {
			return &interpreter.Boolean{Value: false}
		}
		
		for _, ann := range variable.Annotations {
			if ann.Name == annName.Value {
				return &interpreter.Boolean{Value: true}
			}
		}
		return &interpreter.Boolean{Value: false}
	}}

	// __get_field_value(obj, fieldName)
	vm.globals["__get_field_value"] = &interpreter.Builtin{Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) != 2 {
			return &interpreter.Error{Message: "__get_field_value 需要2个参数"}
		}
		instance, ok := args[0].(*interpreter.Instance)
		if !ok {
			return &interpreter.Error{Message: "__get_field_value 第一个参数必须是实例"}
		}
		fieldName, ok := args[1].(*interpreter.String)
		if !ok {
			return &interpreter.Error{Message: "__get_field_value 第二个参数必须是字符串"}
		}
		
		if val, ok := instance.Fields[fieldName.Value]; ok {
			return val
		}
		return &interpreter.Null{}
	}}

	// __set_field_value(obj, fieldName, value)
	vm.globals["__set_field_value"] = &interpreter.Builtin{Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) != 3 {
			return &interpreter.Error{Message: "__set_field_value 需要3个参数"}
		}
		instance, ok := args[0].(*interpreter.Instance)
		if !ok {
			return &interpreter.Error{Message: "__set_field_value 第一个参数必须是实例"}
		}
		fieldName, ok := args[1].(*interpreter.String)
		if !ok {
			return &interpreter.Error{Message: "__set_field_value 第二个参数必须是字符串"}
		}
		
		instance.Fields[fieldName.Value] = args[2]
		return &interpreter.Null{}
	}}

	// __get_class_name(obj)
	vm.globals["__get_class_name"] = &interpreter.Builtin{Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) != 1 {
			return &interpreter.Error{Message: "__get_class_name 需要1个参数"}
		}
		
		switch obj := args[0].(type) {
		case *interpreter.Instance:
			name := obj.Class.Name
			if obj.Class.Namespace != "" {
				name = obj.Class.Namespace + "." + name
			}
			return &interpreter.String{Value: name}
		case *interpreter.Class:
			name := obj.Name
			if obj.Namespace != "" {
				name = obj.Namespace + "." + name
			}
			return &interpreter.String{Value: name}
		}
		return &interpreter.String{Value: ""}
	}}

	// __create_instance(className, ...)
	vm.globals["__create_instance"] = &interpreter.Builtin{Fn: func(args ...interpreter.Object) interpreter.Object {
		if len(args) < 1 {
			return &interpreter.Error{Message: "__create_instance 需要至少1个参数"}
		}
		className, ok := args[0].(*interpreter.String)
		if !ok {
			return &interpreter.Error{Message: "__create_instance 第一个参数必须是字符串"}
		}
		
		class, ok := vm.getClassByName(className.Value)
		if !ok {
			return &interpreter.Error{Message: "未找到类: " + className.Value}
		}
		
		// 创建实例
		instance := &interpreter.Instance{
			Class:  class,
			Fields: make(map[string]interpreter.Object),
		}
		for name, variable := range class.Variables {
			if variable.DefaultValue != nil {
				instance.Fields[name] = variable.DefaultValue
			} else {
				instance.Fields[name] = &interpreter.Null{}
			}
		}
		
		// 如果有构造函数，调用它
		// 注意：这里的调用是同步的，对于 VM 来说有点复杂，因为构造函数可能是字节码
		// 但 __create_instance 通常在反射中使用
		// 这里我们简化处理，如果构造函数是字节码，目前不支持在 builtin 中直接调用并等待返回
		// 除非我们手动执行一个子虚拟机
		
		if constructor, ok := class.GetMethod("__construct"); ok {
			// TODO: 支持在内置函数中调用字节码构造函数
			_ = constructor
		}
		
		return instance
	}}
}

// getClassByName 根据完整名称查找类
func (vm *VM) getClassByName(name string) (*interpreter.Class, bool) {
	// 尝试直接查找（可能已经在全局命名空间）
	namespace, className, err := interpreter.ResolveClassName(name)
	if err != nil {
		return nil, false
	}
	
	ns := vm.namespaceMgr.GetNamespace(namespace)
	if ns == nil {
		return nil, false
	}
	
	return ns.GetClass(className)
}

// annotationsToArray 将注解列表转换为数组
func (vm *VM) annotationsToArray(annotations []*interpreter.AnnotationInstance) *interpreter.Array {
	elements := make([]interpreter.Object, len(annotations))
	for i, ann := range annotations {
		elements[i] = vm.annotationToMap(ann)
	}
	return &interpreter.Array{Elements: elements}
}

// annotationToMap 将单个注解转换为 map
func (vm *VM) annotationToMap(ann *interpreter.AnnotationInstance) *interpreter.Map {
	pairs := make(map[string]interpreter.Object)
	keys := make([]string, 0)
	
	pairs["name"] = &interpreter.String{Value: ann.Name}
	keys = append(keys, "name")
	
	args := make(map[string]interpreter.Object)
	argKeys := make([]string, 0)
	for k, v := range ann.Arguments {
		args[k] = v
		argKeys = append(argKeys, k)
	}
	pairs["arguments"] = &interpreter.Map{Pairs: args, Keys: argKeys}
	keys = append(keys, "arguments")
	
	return &interpreter.Map{Pairs: pairs, Keys: keys}
}

// SetDebug 设置调试模式
func (vm *VM) SetDebug(debug bool) {
	vm.debug = debug
}

// SetProjectConfig 设置项目配置
func (vm *VM) SetProjectConfig(projectRoot string, cfg *config.ProjectConfig) {
	vm.projectRoot = projectRoot
	vm.projectConfig = cfg
}

// SetStdlibPath 设置标准库路径
func (vm *VM) SetStdlibPath(path string) {
	vm.stdlibPath = path
}

// SetCurrentNamespace 设置当前命名空间
func (vm *VM) SetCurrentNamespace(namespace string) {
	vm.currentNamespace = vm.namespaceMgr.GetNamespace(namespace)
}

// loadNamespaceFile 根据命名空间路径加载对应的文件
// 例如：Mycompany.Myapp.Models + User -> src/Mycompany/Myapp/Models/User.long
func (vm *VM) loadNamespaceFile(namespace string, className string) error {
	// 构建完整的命名空间路径作为缓存 key
	fullKey := namespace + "." + className

	// 检查是否已经加载过
	if vm.loadedNamespaces[fullKey] {
		return nil
	}

	// 循环依赖检测
	if vm.loadingNamespaces[fullKey] {
		// 正在加载中，说明存在循环依赖，但我们允许这种情况
		// 因为类定义会在后续被正确注册
		return nil
	}

	// 标记为正在加载
	vm.loadingNamespaces[fullKey] = true
	defer func() {
		delete(vm.loadingNamespaces, fullKey)
	}()

	// 将命名空间转换为文件路径
	// 例如：Mycompany.Myapp.Models -> Mycompany/Myapp/Models
	namespacePath := strings.ReplaceAll(namespace, ".", string(filepath.Separator))

	// 构建可能的文件路径
	var filePaths []string

	// 如果有项目根目录，搜索项目目录
	if vm.projectRoot != "" {
		// 如果有 root_namespace，计算相对路径
		relativeNamespacePath := namespacePath
		if vm.projectConfig != nil && vm.projectConfig.RootNamespace != "" {
			rootNsPath := strings.ReplaceAll(vm.projectConfig.RootNamespace, ".", string(filepath.Separator))
			// 确保 rootNsPath 以路径分隔符结尾，以便正确匹配前缀
			rootNsPathWithSep := rootNsPath + string(filepath.Separator)
			if strings.HasPrefix(namespacePath, rootNsPathWithSep) {
				// 去掉 root_namespace 前缀
				relativeNamespacePath = strings.TrimPrefix(namespacePath, rootNsPathWithSep)
			} else if namespacePath == rootNsPath {
				// 命名空间正好是 root_namespace
				relativeNamespacePath = ""
			}
		}

		// 1. 使用相对路径在 src 目录下查找（优先）
		if relativeNamespacePath != "" {
			srcRelPath := filepath.Join(vm.projectRoot, "src", relativeNamespacePath, className+".long")
			filePaths = append(filePaths, srcRelPath)
		} else {
			// 如果相对路径为空，直接在 src 下查找
			srcRelPath := filepath.Join(vm.projectRoot, "src", className+".long")
			filePaths = append(filePaths, srcRelPath)
		}

		// 2. 使用完整路径在 src 目录下查找
		srcPath := filepath.Join(vm.projectRoot, "src", namespacePath, className+".long")
		filePaths = append(filePaths, srcPath)

		// 3. 在项目根目录下查找（相对路径）
		if relativeNamespacePath != "" {
			rootRelPath := filepath.Join(vm.projectRoot, relativeNamespacePath, className+".long")
			filePaths = append(filePaths, rootRelPath)
		}

		// 4. 在项目根目录下查找（完整路径）
		rootPath := filepath.Join(vm.projectRoot, namespacePath, className+".long")
		filePaths = append(filePaths, rootPath)

		// 5. 在 vendor 目录下查找
		vendorPath := filepath.Join(vm.projectRoot, "vendor", namespacePath, className+".long")
		filePaths = append(filePaths, vendorPath)
	}

	// 6. 在标准库目录下查找
	if vm.stdlibPath != "" {
		stdlibPath := filepath.Join(vm.stdlibPath, namespacePath, className+".long")
		filePaths = append(filePaths, stdlibPath)
	}

	// 尝试加载文件
	var loadedPath string
	var content string
	for _, path := range filePaths {
		if c, err := ioutil.ReadFile(path); err == nil {
			content = string(c)
			loadedPath = path
			break
		}
	}

	if loadedPath == "" {
		return fmt.Errorf("找不到文件: %s，尝试路径: %v", className+".long", filePaths)
	}

	// 标记为已加载
	vm.loadedNamespaces[fullKey] = true

	// 解析文件
	l := lexer.New(content)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Errorf("解析文件 %s 错误: %s", loadedPath, p.Errors()[0])
	}

	// 保存当前状态
	savedNamespace := vm.currentNamespace
	savedConfig := vm.projectConfig

	// 如果是从标准库加载的文件，临时禁用 projectConfig
	isStdlibFile := vm.stdlibPath != "" && strings.HasPrefix(loadedPath, vm.stdlibPath)
	if isStdlibFile {
		vm.projectConfig = nil
	}

	// 编译并执行文件
	compiler := NewCompiler()
	compiler.SetVM(vm)
	bytecode, err := compiler.Compile(program)
	if err != nil {
		vm.projectConfig = savedConfig
		vm.currentNamespace = savedNamespace
		return fmt.Errorf("编译文件 %s 错误: %s", loadedPath, err)
	}

	// 执行字节码
	_, err = vm.runBytecode(bytecode)
	if err != nil {
		vm.projectConfig = savedConfig
		vm.currentNamespace = savedNamespace
		return fmt.Errorf("执行文件 %s 错误: %s", loadedPath, err)
	}

	// 恢复状态
	vm.projectConfig = savedConfig
	vm.currentNamespace = savedNamespace

	return nil
}

// runBytecode 执行字节码（不调用入口点）
func (vm *VM) runBytecode(bytecode *Bytecode) (interpreter.Object, error) {
	// 保存当前状态
	savedBytecode := vm.bytecode
	savedFrameCount := vm.frameCount
	savedSP := vm.sp

	vm.bytecode = bytecode

	// 创建主函数
	mainFn := &CompiledFunction{
		Bytecode:  bytecode,
		NumLocals: 0,
		NumParams: 0,
		Name:      "<module>",
	}

	// 创建主闭包和帧
	mainClosure := NewClosure(mainFn)
	vm.pushFrame(mainClosure, vm.sp)

	// 执行指令
	result, err := vm.execute()

	// 恢复状态
	vm.bytecode = savedBytecode
	vm.frameCount = savedFrameCount
	vm.sp = savedSP

	return result, err
}

// ProcessUseStatement 处理 use 语句
func (vm *VM) ProcessUseStatement(fullPath string, alias string) error {
	// 解析完全限定名：Illuminate.Database.Eloquent.Model
	// 分解为：命名空间 Illuminate.Database.Eloquent，类名 Model
	namespace, symbolName, err := interpreter.ResolveClassName(fullPath)
	if err != nil {
		return fmt.Errorf("无效的 use 路径: %s", fullPath)
	}

	// 尝试加载命名空间文件
	loadErr := vm.loadNamespaceFile(namespace, symbolName)

	// 首先尝试在原始命名空间中查找
	targetNamespace := vm.namespaceMgr.GetNamespace(namespace)

	// 尝试查找类、枚举或接口
	var symbol interpreter.Object
	var found bool

	// 1. 尝试查找类
	if class, ok := targetNamespace.GetClass(symbolName); ok {
		symbol = class
		found = true
	}

	// 2. 尝试查找枚举
	if !found {
		if enum, ok := targetNamespace.GetEnum(symbolName); ok {
			symbol = enum
			found = true
		}
	}

	// 3. 尝试查找接口
	if !found {
		if iface, ok := targetNamespace.GetInterface(symbolName); ok {
			symbol = iface
			found = true
		}
	}

	// 如果找不到，且有项目配置，尝试使用 root_namespace 解析后的命名空间
	if !found && vm.projectConfig != nil {
		resolvedNamespace := vm.projectConfig.ResolveNamespace(namespace)
		if resolvedNamespace != namespace {
			targetNamespace = vm.namespaceMgr.GetNamespace(resolvedNamespace)

			if class, ok := targetNamespace.GetClass(symbolName); ok {
				symbol = class
				found = true
			}
			if !found {
				if enum, ok := targetNamespace.GetEnum(symbolName); ok {
					symbol = enum
					found = true
				}
			}
			if !found {
				if iface, ok := targetNamespace.GetInterface(symbolName); ok {
					symbol = iface
					found = true
				}
			}
		}
	}

	if !found {
		if loadErr != nil {
			return loadErr
		}
		return fmt.Errorf("命名空间 %s 中没有找到 %s", namespace, symbolName)
	}

	// 确定导入到当前环境的名称
	importName := symbolName
	if alias != "" {
		importName = alias
	}

	// 将符号注册到全局作用域
	vm.globals[importName] = symbol

	return nil
}

// Run 运行字节码
func (vm *VM) Run(bytecode *Bytecode) (interpreter.Object, error) {
	vm.bytecode = bytecode

	// 在栈底压入一个哨兵值，确保 basePointer 永远 >= 1
	// 这样 OP_RETURN 中的 `vm.sp = frame.basePointer - 1` 永远不会是负数
	vm.push(&interpreter.Null{})

	// 创建主函数
	mainFn := &CompiledFunction{
		Bytecode:  bytecode,
		NumLocals: 0,
		NumParams: 0,
		Name:      "<main>",
	}

	// 创建主闭包和帧
	mainClosure := NewClosure(mainFn)
	vm.pushFrame(mainClosure, 1) // basePointer 从 1 开始

	// 执行指令
	result, err := vm.execute()
	if err != nil {
		return nil, err
	}

	// 查找并调用入口点（Application::main 或其他包含 main 静态方法的类）
	entryResult, entryErr := vm.callEntryPoint()
	if entryErr != nil {
		return nil, entryErr
	}
	if entryResult != nil {
		return entryResult, nil
	}

	return result, nil
}

// callEntryPoint 查找并调用入口点
func (vm *VM) callEntryPoint() (interpreter.Object, error) {
	// 遍历全局变量，查找包含 main 静态方法的类
	for name, obj := range vm.globals {
		if class, ok := obj.(*interpreter.Class); ok {
			if method, ok := class.GetStaticMethod("main"); ok {
				// 找到入口点，调用它
				if vm.debug {
					fmt.Printf("=== 调用入口点 %s::main() ===\n", name)
				}
				
				// 获取方法闭包
				if closure, ok := method.Body.(*Closure); ok {
					// 压入闭包作为函数
					vm.push(closure)
					// 调用（无参数）
					if err := vm.callClosure(closure, 0); err != nil {
						return nil, err
					}
					// 执行
					return vm.execute()
				}
			}
		}
	}
	
	// 没有找到入口点，这是正常的（可能是库文件）
	return nil, nil
}

// execute 执行指令循环
func (vm *VM) execute() (interpreter.Object, error) {
	for vm.frameCount > 0 {
		frame := vm.currentFrame()
		
		// 检查是否到达指令末尾
		if frame.ip >= len(frame.Instructions()) {
			// 如果是主函数，返回栈顶值
			if vm.frameCount == 1 {
				if vm.sp > 0 {
					return vm.pop(), nil
				}
				return &interpreter.Null{}, nil
			}
			// 否则，隐式返回 null
			vm.popFrame()
			vm.push(&interpreter.Null{})
			continue
		}

		// 读取操作码
		op := Opcode(frame.ReadByte())

		// 调试输出
		if vm.debug {
			fmt.Printf("IP: %d, OP: %s, SP: %d\n", frame.ip-1, op.String(), vm.sp)
		}

		// 执行指令
		err := vm.executeInstruction(op, frame)
		if err != nil {
			// 检查是否有 try-catch
			if vm.tryCount > 0 {
				// 尝试处理异常
				if handled := vm.handleException(err); handled {
					continue
				}
			}
			return nil, err
		}
	}

	// 返回栈顶值
	if vm.sp > 0 {
		return vm.pop(), nil
	}
	return &interpreter.Null{}, nil
}

// executeInstruction 执行单条指令
func (vm *VM) executeInstruction(op Opcode, frame *Frame) error {
	switch op {
	// 常量加载
	case OP_CONST:
		constant := frame.ReadConstant()
		vm.push(constant)

	case OP_NULL:
		vm.push(&interpreter.Null{})

	case OP_TRUE:
		vm.push(&interpreter.Boolean{Value: true})

	case OP_FALSE:
		vm.push(&interpreter.Boolean{Value: false})

	// 变量操作
	case OP_GET_LOCAL:
		slot := frame.ReadByte()
		value := vm.stack[frame.basePointer+int(slot)]
		if vm.debug {
			fmt.Fprintf(os.Stderr, "GET_LOCAL slot %d (idx %d) = %v (Type: %s)\n", 
				slot, frame.basePointer+int(slot), value, value.Type())
		}
		vm.push(value)

	case OP_SET_LOCAL:
		slot := frame.ReadByte()
		value := vm.peek(0)
		if vm.debug {
			fmt.Fprintf(os.Stderr, "SET_LOCAL slot %d (idx %d) = %v (Type: %s)\n", 
				slot, frame.basePointer+int(slot), value, value.Type())
		}
		vm.stack[frame.basePointer+int(slot)] = value

	case OP_GET_GLOBAL:
		name := frame.ReadConstant().(*interpreter.String).Value
		if name == "__called_class_name" {
			vm.push(&interpreter.String{Value: frame.calledClassName})
			return nil
		}
		if name == "super" {
			// 获取当前类
			className := frame.closure.Fn.ClassName
			if className != "" {
				class, ok := vm.getClassByName(className)
				if ok && class.Parent != nil {
					vm.push(class.Parent)
					return nil
				}
			}
			vm.push(&interpreter.Null{})
			return nil
		}
		value, ok := vm.globals[name]
		if !ok {
			return fmt.Errorf("未定义的变量: %s", name)
		}
		vm.push(value)

	case OP_SET_GLOBAL:
		name := frame.ReadConstant().(*interpreter.String).Value
		vm.globals[name] = vm.peek(0)

	case OP_DEFINE_GLOBAL:
		name := frame.ReadConstant().(*interpreter.String).Value
		value := vm.pop()
		vm.globals[name] = value

	case OP_GET_UPVALUE:
		slot := frame.ReadByte()
		vm.push(frame.closure.Upvalues[slot].Get())

	case OP_SET_UPVALUE:
		slot := frame.ReadByte()
		frame.closure.Upvalues[slot].Set(vm.peek(0))

	// 算术运算
	case OP_ADD:
		if err := vm.binaryAdd(); err != nil {
			return err
		}

	case OP_SUB:
		if err := vm.binaryOp(func(a, b int64) int64 { return a - b },
			func(a, b float64) float64 { return a - b }); err != nil {
			return err
		}

	case OP_MUL:
		if err := vm.binaryOp(func(a, b int64) int64 { return a * b },
			func(a, b float64) float64 { return a * b }); err != nil {
			return err
		}

	case OP_DIV:
		if err := vm.binaryDiv(); err != nil {
			return err
		}

	case OP_MOD:
		if err := vm.binaryMod(); err != nil {
			return err
		}

	case OP_NEG:
		if err := vm.unaryNeg(); err != nil {
			return err
		}

	// 比较运算
	case OP_EQ:
		b := vm.pop()
		a := vm.pop()
		vm.push(&interpreter.Boolean{Value: vm.isEqual(a, b)})

	case OP_NE:
		b := vm.pop()
		a := vm.pop()
		vm.push(&interpreter.Boolean{Value: !vm.isEqual(a, b)})

	case OP_LT:
		if err := vm.compareOp("<"); err != nil {
			return err
		}

	case OP_LE:
		if err := vm.compareOp("<="); err != nil {
			return err
		}

	case OP_GT:
		if err := vm.compareOp(">"); err != nil {
			return err
		}

	case OP_GE:
		if err := vm.compareOp(">="); err != nil {
			return err
		}

	// 逻辑运算
	case OP_NOT:
		value := vm.pop()
		vm.push(&interpreter.Boolean{Value: !vm.isTruthy(value)})

	// 位运算
	case OP_BIT_AND:
		if err := vm.bitwiseOp(func(a, b int64) int64 { return a & b }); err != nil {
			return err
		}

	case OP_BIT_OR:
		if err := vm.bitwiseOp(func(a, b int64) int64 { return a | b }); err != nil {
			return err
		}

	case OP_BIT_XOR:
		if err := vm.bitwiseOp(func(a, b int64) int64 { return a ^ b }); err != nil {
			return err
		}

	case OP_BIT_NOT:
		operand := vm.pop()
		if intVal, ok := operand.(*interpreter.Integer); ok {
			vm.push(&interpreter.Integer{Value: ^intVal.Value})
		} else {
			return fmt.Errorf("按位取反需要整数类型")
		}

	case OP_LSHIFT:
		if err := vm.bitwiseOp(func(a, b int64) int64 { return a << uint64(b) }); err != nil {
			return err
		}

	case OP_RSHIFT:
		if err := vm.bitwiseOp(func(a, b int64) int64 { return a >> uint64(b) }); err != nil {
			return err
		}

	// 控制流
	case OP_JUMP:
		offset := frame.ReadUint16()
		frame.ip += int(offset)

	case OP_JUMP_IF_FALSE:
		offset := frame.ReadUint16()
		if !vm.isTruthy(vm.peek(0)) {
			frame.ip += int(offset)
		}

	case OP_JUMP_IF_TRUE:
		offset := frame.ReadUint16()
		if vm.isTruthy(vm.peek(0)) {
			frame.ip += int(offset)
		}

	case OP_LOOP:
		offset := frame.ReadUint16()
		frame.ip -= int(offset)

	// 函数调用
	case OP_CALL:
		argCount := int(frame.ReadByte())
		if err := vm.callValue(vm.peek(argCount), argCount); err != nil {
			return err
		}

	case OP_RETURN:
		result := vm.pop()
		
		// 检查是否是构造函数返回
		isConstructor := frame.isConstructor
		isMethodCall := frame.isMethodCall
		basePointer := frame.basePointer
		constructorInstance := frame.constructorInstance
		
		// 关闭 upvalues
		vm.closeUpvalues(basePointer)
		
		// 弹出帧
		vm.popFrame()
		
		// 弹出函数对象和参数
		if isMethodCall {
			vm.sp = basePointer
		} else {
			vm.sp = basePointer - 1
		}
		
		// 如果是构造函数，返回实例而不是 null
		if isConstructor && constructorInstance != nil {
			result = constructorInstance
		}
		
		// 压入返回值
		vm.push(result)

	case OP_CLOSURE:
		fnIndex := frame.ReadByte()
		fn := frame.closure.Fn.Bytecode.Constants[fnIndex].(*CompiledFunction)
		closure := NewClosure(fn)

		// 读取 upvalue 信息
		for i := 0; i < fn.UpvalueCount; i++ {
			isLocal := frame.ReadByte()
			index := frame.ReadByte()
			if isLocal == 1 {
				// 从当前帧的局部变量创建 upvalue
				closure.Upvalues[i] = vm.captureUpvalue(frame.basePointer + int(index))
			} else {
				// 从外层闭包的 upvalue 获取
				closure.Upvalues[i] = frame.closure.Upvalues[index]
			}
		}

		vm.push(closure)

	case OP_CLOSE_UPVALUE:
		vm.closeUpvalues(vm.sp - 1)
		vm.pop()

	// 类和对象
	case OP_CLASS:
		name := frame.ReadConstant().(*interpreter.String).Value
		class := &interpreter.Class{
			Name:            name,
			Methods:         make(map[string]*interpreter.ClassMethod),
			StaticMethods:   make(map[string]*interpreter.ClassMethod),
			Variables:       make(map[string]*interpreter.ClassVariable),
			StaticVariables: make(map[string]*interpreter.ClassVariable),
			StaticFields:    make(map[string]interpreter.Object),
			Constants:       make(map[string]*interpreter.ClassConstant),
		}
		// 设置命名空间
		if vm.currentNamespace != nil {
			class.Namespace = vm.currentNamespace.FullName
			// 注册到命名空间
			vm.currentNamespace.SetClass(name, class)
		}
		vm.push(class)

	case OP_METHOD:
		name := frame.ReadConstant().(*interpreter.String).Value
		method := vm.pop().(*Closure)
		class := vm.peek(0).(*interpreter.Class)
		class.Methods[name] = &interpreter.ClassMethod{
			Name: name,
			Body: method, // 存储闭包
		}

	case OP_STATIC_METHOD:
		name := frame.ReadConstant().(*interpreter.String).Value
		method := vm.pop().(*Closure)
		class := vm.peek(0).(*interpreter.Class)
		class.StaticMethods[name] = &interpreter.ClassMethod{
			Name:     name,
			IsStatic: true,
			Body:     method,
		}

	case OP_CLASS_VAR:
		name := frame.ReadConstant().(*interpreter.String).Value
		defaultValue := vm.pop()
		class := vm.peek(0).(*interpreter.Class)
		class.Variables[name] = &interpreter.ClassVariable{
			Name:           name,
			Type:           "", // 类型信息在运行时不关键
			DefaultValue:   defaultValue,
			AccessModifier: "public", // 默认公开
		}

	case OP_STATIC_VAR:
		name := frame.ReadConstant().(*interpreter.String).Value
		defaultValue := vm.pop()
		class := vm.peek(0).(*interpreter.Class)
		class.StaticVariables[name] = &interpreter.ClassVariable{
			Name:           name,
			Type:           "",
			DefaultValue:   defaultValue,
			IsStatic:       true,
			AccessModifier: "public",
		}
		// 同时初始化静态字段
		class.StaticFields[name] = defaultValue

	case OP_CLASS_CONST:
		name := frame.ReadConstant().(*interpreter.String).Value
		value := vm.pop()
		class := vm.peek(0).(*interpreter.Class)
		class.Constants[name] = &interpreter.ClassConstant{
			Name:           name,
			Value:          value,
			AccessModifier: "public",
		}

	case OP_GET_PROPERTY:
		name := frame.ReadConstant().(*interpreter.String).Value
		obj := vm.pop()
		
		if instance, ok := obj.(*interpreter.Instance); ok {
			// 先查找字段
			if value, ok := instance.Fields[name]; ok {
				vm.push(value)
				return nil
			}
			// 再查找方法
			if method, ok := instance.Class.GetMethod(name); ok {
				// 创建绑定方法
				if closure, ok := method.Body.(*Closure); ok {
					vm.push(&BoundMethod{Receiver: instance, Method: closure})
				} else {
					return fmt.Errorf("方法类型错误")
				}
				return nil
			}
			return fmt.Errorf("实例没有属性: %s", name)
		}
		
		return fmt.Errorf("只能访问实例的属性")

	case OP_SET_PROPERTY:
		name := frame.ReadConstant().(*interpreter.String).Value
		
		// Stack has [..., value, object] with object on top
		// Pop in correct order: object first, then value
		obj := vm.pop()
		value := vm.pop()
		
		if instance, ok := obj.(*interpreter.Instance); ok {
			instance.Fields[name] = value
			vm.push(value)
			return nil
		}
		
		return fmt.Errorf("只能设置实例的属性，当前对象类型: %T", obj)

	case OP_GET_STATIC_FIELD:
		name := frame.ReadConstant().(*interpreter.String).Value
		obj := vm.pop()
		
		if class, ok := obj.(*interpreter.Class); ok {
			if value, ok := class.StaticFields[name]; ok {
				vm.push(value)
				return nil
			}
			return fmt.Errorf("类 %s 没有静态字段: %s", class.Name, name)
		}
		
		return fmt.Errorf("只能访问类的静态字段")

	case OP_SET_STATIC_FIELD:
		name := frame.ReadConstant().(*interpreter.String).Value
		obj := vm.pop()
		value := vm.pop()
		
		if class, ok := obj.(*interpreter.Class); ok {
			class.StaticFields[name] = value
			vm.push(value)
			return nil
		}
		
		return fmt.Errorf("只能设置类的静态字段")

	case OP_INVOKE:
		method := frame.ReadConstant().(*interpreter.String).Value
		argCount := int(frame.ReadByte())
		if vm.debug {
			receiver := vm.peek(argCount)
			fmt.Fprintf(os.Stderr, "INVOKE %s on %v (Type: %s)\n", method, receiver, receiver.Type())
		}
		if err := vm.invoke(method, argCount); err != nil {
			return err
		}

	case OP_INVOKE_STATIC:
		method := frame.ReadConstant().(*interpreter.String).Value
		argCount := int(frame.ReadByte())
		if err := vm.invokeStatic(method, argCount); err != nil {
			return err
		}

	case OP_NEW:
		argCount := int(frame.ReadByte())
		classObj := vm.peek(argCount)
		class, ok := classObj.(*interpreter.Class)
		if !ok {
			return fmt.Errorf("OP_NEW: 期望 CLASS 类型，但得到 %s", classObj.Type())
		}
		
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
		replaceIdx := vm.sp - argCount - 1
		if replaceIdx < 0 || replaceIdx >= len(vm.stack) {
			return fmt.Errorf("OP_NEW: 无效的栈索引 %d (sp=%d, argCount=%d)", replaceIdx, vm.sp, argCount)
		}
		vm.stack[replaceIdx] = instance
		
		// 调用构造函数（如果存在）
		if constructor, ok := class.GetMethod("__construct"); ok {
			if closure, ok := constructor.Body.(*Closure); ok {
				// 使用 callConstructor 因为构造函数也是方法调用
				if err := vm.callConstructor(closure, argCount+1); err != nil {
					// 构造函数调用失败，但仍然返回实例
					// 弹出参数（如果有）
					if argCount > 0 {
						vm.sp -= argCount
					}
					// 设置当前帧的实例
					frame := vm.currentFrame()
					if frame != nil {
						frame.constructorInstance = instance
					}
				} else {
					// 构造函数调用成功
					// 设置当前帧的实例
					vm.currentFrame().constructorInstance = instance
				}
			} else {
				// 构造函数不是闭包，可能是内置函数或其他
				// 设置当前帧的实例
				frame := vm.currentFrame()
				if frame != nil {
					frame.constructorInstance = instance
				}
				// 没有构造函数调用，弹出参数
				if argCount > 0 {
					vm.sp -= argCount
				}
			}
		} else {
			// 没有构造函数，弹出参数，只保留实例
			if argCount > 0 {
				vm.sp -= argCount
			}
			// 设置当前帧的实例（虽然没有构造函数调用）
			frame := vm.currentFrame()
			if frame != nil {
				frame.constructorInstance = instance
			}
		}

	case OP_INHERIT:
		superclass := vm.peek(1)
		subclass := vm.peek(0).(*interpreter.Class)
		
		if parent, ok := superclass.(*interpreter.Class); ok {
			subclass.Parent = parent
			vm.pop() // 弹出子类
		} else {
			return fmt.Errorf("父类必须是一个类")
		}

	case OP_INSTANCE_OF:
		class := vm.pop()
		obj := vm.pop()
		
		targetClass, ok := class.(*interpreter.Class)
		if !ok {
			vm.push(&interpreter.Boolean{Value: false})
			return nil
		}
		
		instance, ok := obj.(*interpreter.Instance)
		if !ok {
			vm.push(&interpreter.Boolean{Value: false})
			return nil
		}
		
		// 检查继承链
		curr := instance.Class
		found := false
		for curr != nil {
			if curr == targetClass {
				found = true
				break
			}
			curr = curr.Parent
		}
		vm.push(&interpreter.Boolean{Value: found})
		return nil

	case OP_GET_SUPER:
		name := frame.ReadConstant().(*interpreter.String).Value
		superclass := vm.pop().(*interpreter.Class)
		
		if method, ok := superclass.GetMethod(name); ok {
			if closure, ok := method.Body.(*Closure); ok {
				instance := vm.pop().(*interpreter.Instance)
				vm.push(&BoundMethod{Receiver: instance, Method: closure})
			}
		} else {
			return fmt.Errorf("父类没有方法: %s", name)
		}

	case OP_SUPER_INVOKE:
		method := frame.ReadConstant().(*interpreter.String).Value
		argCount := int(frame.ReadByte())
		superclassObj := vm.pop()
		
		superclass, ok := superclassObj.(*interpreter.Class)
		if !ok {
			return fmt.Errorf("super 只能在类方法内部使用")
		}
		
		if m, ok := superclass.GetMethod(method); ok {
			if closure, ok := m.Body.(*Closure); ok {
				return vm.callMethod(closure, argCount+1)
			}
		}
		return fmt.Errorf("父类没有方法: %s", method)

	// 数组和 Map
	case OP_ARRAY:
		count := int(frame.ReadByte())
		elements := make([]interpreter.Object, count)
		for i := count - 1; i >= 0; i-- {
			elements[i] = vm.pop()
		}
		vm.push(&interpreter.Array{Elements: elements})

	case OP_MAP:
		count := int(frame.ReadByte())
		pairs := make(map[string]interpreter.Object)
		keys := make([]string, 0, count)
		for i := 0; i < count; i++ {
			value := vm.pop()
			key := vm.pop().(*interpreter.String).Value
			pairs[key] = value
			keys = append([]string{key}, keys...)
		}
		vm.push(&interpreter.Map{Pairs: pairs, Keys: keys})

	case OP_INDEX:
		index := vm.pop()
		obj := vm.pop()
		result, err := vm.indexGet(obj, index)
		if err != nil {
			return err
		}
		vm.push(result)

	case OP_INDEX_SET:
		// 栈顺序: value(底), obj, index(顶)
		index := vm.pop()
		obj := vm.pop()
		value := vm.peek(0) // 保留值在栈上作为表达式结果
		if err := vm.indexSet(obj, index, value); err != nil {
			return err
		}

	case OP_SLICE:
		endIdx := vm.pop()
		startIdx := vm.pop()
		obj := vm.pop()
		result, err := vm.sliceGet(obj, startIdx, endIdx)
		if err != nil {
			return err
		}
		vm.push(result)

	// 异常处理
	case OP_THROW:
		exception := vm.pop()
		vm.exception = exception
		return &VMError{Message: "异常抛出", Exception: exception}

	case OP_PUSH_TRY:
		catchOffset := frame.ReadUint16()
		vm.pushTry(frame.ip+int(catchOffset), vm.sp, vm.frameCount)

	case OP_POP_TRY:
		vm.popTry()

	// 协程
	case OP_GO:
		argCount := int(frame.ReadByte())
		fn := vm.peek(argCount)
		if closure, ok := fn.(*Closure); ok {
			// 创建新的虚拟机实例执行协程
			go vm.runGoroutine(closure, argCount)
			// 弹出函数和参数
			vm.sp -= argCount + 1
			vm.push(&interpreter.Null{})
		} else {
			return fmt.Errorf("go 只能用于函数")
		}

	// 其他
	case OP_POP:
		vm.pop()

	case OP_DUP:
		vm.push(vm.peek(0))

	case OP_SWAP:
		a := vm.pop()
		b := vm.pop()
		vm.push(a)
		vm.push(b)

	case OP_PRINT:
		value := vm.pop()
		fmt.Println(value.Inspect())

	case OP_HALT:
		return nil

	case OP_INCREMENT:
		slot := frame.ReadByte()
		if intVal, ok := vm.stack[frame.basePointer+int(slot)].(*interpreter.Integer); ok {
			vm.stack[frame.basePointer+int(slot)] = &interpreter.Integer{Value: intVal.Value + 1}
		}

	case OP_DECREMENT:
		slot := frame.ReadByte()
		if intVal, ok := vm.stack[frame.basePointer+int(slot)].(*interpreter.Integer); ok {
			vm.stack[frame.basePointer+int(slot)] = &interpreter.Integer{Value: intVal.Value - 1}
		}

	case OP_BUILTIN:
		builtinID := frame.ReadByte()
		argCount := int(frame.ReadByte())
		if err := vm.callBuiltin(int(builtinID), argCount); err != nil {
			return err
		}

	default:
		return fmt.Errorf("未知的操作码: %d", op)
	}

	return nil
}

// ========== 栈操作 ==========

// push 压入栈
func (vm *VM) push(obj interpreter.Object) {
	if vm.sp >= StackSize {
		panic("栈溢出")
	}
	vm.stack[vm.sp] = obj
	vm.sp++
}

// pop 弹出栈
func (vm *VM) pop() interpreter.Object {
	vm.sp--
	return vm.stack[vm.sp]
}

// peek 查看栈顶（不弹出）
func (vm *VM) peek(distance int) interpreter.Object {
	return vm.stack[vm.sp-1-distance]
}

// ========== 帧操作 ==========

// currentFrame 获取当前帧
func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameCount-1]
}

// pushFrame 压入帧并返回帧指针
func (vm *VM) pushFrame(closure *Closure, basePointer int) *Frame {
	if vm.frameCount >= FrameSize {
		panic("调用栈溢出")
	}
	frame := NewFrame(closure, basePointer)
	vm.frames[vm.frameCount] = frame
	vm.frameCount++
	
	if vm.debug {
		fmt.Fprintf(os.Stderr, "PUSH FRAME %s, basePointer: %d, NumLocals: %d, current SP: %d\n", 
			closure.Fn.Name, basePointer, closure.Fn.NumLocals, vm.sp)
	}

	// 确保 SP 在局部变量之上，防止操作数覆盖局部变量
	if basePointer+closure.Fn.NumLocals > vm.sp {
		vm.sp = basePointer + closure.Fn.NumLocals
	}
	
	return frame
}

// popFrame 弹出帧
func (vm *VM) popFrame() *Frame {
	vm.frameCount--
	return vm.frames[vm.frameCount]
}

// ========== Upvalue 管理 ==========

// captureUpvalue 捕获 upvalue
func (vm *VM) captureUpvalue(stackIndex int) *Upvalue {
	// 检查栈索引是否有效
	if stackIndex < 0 || stackIndex >= len(vm.stack) {
		// 如果索引无效，创建一个默认的upvalue
		var defaultValue interpreter.Object = &interpreter.Integer{Value: 0}
		upvalue := NewUpvalueWithIndex(&defaultValue, stackIndex)
		upvalue.Closed = true
		upvalue.Value = defaultValue
		return upvalue
	}

	// 查找是否已存在
	prev := (*Upvalue)(nil)
	curr := vm.openUpvalues
	for curr != nil && curr.StackIndex > stackIndex {
		prev = curr
		curr = curr.Next
	}

	if curr != nil && curr.StackIndex == stackIndex {
		return curr
	}

	// 创建新的 upvalue
	upvalue := NewUpvalueWithIndex(&vm.stack[stackIndex], stackIndex)
	upvalue.Next = curr
	if prev == nil {
		vm.openUpvalues = upvalue
	} else {
		prev.Next = upvalue
	}

	return upvalue
}

// closeUpvalues 关闭 upvalues
func (vm *VM) closeUpvalues(last int) {
	for vm.openUpvalues != nil {
		upvalue := vm.openUpvalues
		if upvalue.StackIndex < last {
			break
		}
		upvalue.Close()
		vm.openUpvalues = upvalue.Next
	}
}

// ========== Try-Catch 管理 ==========

// pushTry 压入 try 状态
func (vm *VM) pushTry(catchTarget, stackDepth, frameIndex int) {
	if vm.tryCount >= MaxTryDepth {
		panic("try 嵌套过深")
	}
	vm.tryStack[vm.tryCount] = &TryState{
		CatchTarget: catchTarget,
		StackDepth:  stackDepth,
		FrameIndex:  frameIndex,
	}
	vm.tryCount++
}

// popTry 弹出 try 状态
func (vm *VM) popTry() *TryState {
	vm.tryCount--
	return vm.tryStack[vm.tryCount]
}

// handleException 处理异常
func (vm *VM) handleException(err error) bool {
	if vm.tryCount == 0 {
		return false
	}

	// 获取最近的 try 块
	tryState := vm.tryStack[vm.tryCount-1]
	vm.tryCount--

	// 恢复栈和帧
	vm.sp = tryState.StackDepth
	for vm.frameCount > tryState.FrameIndex {
		vm.popFrame()
	}

	// 压入异常对象
	if vmErr, ok := err.(*VMError); ok && vmErr.Exception != nil {
		vm.push(vmErr.Exception)
	} else {
		// 对于非 VMError 类型的错误，创建一个 RuntimeException 实例
		exceptionObj := vm.createRuntimeException(err.Error())
		vm.push(exceptionObj)
	}

	// 跳转到 catch 块
	vm.currentFrame().ip = tryState.CatchTarget

	return true
}

// createRuntimeException 创建一个运行时异常实例
func (vm *VM) createRuntimeException(message string) interpreter.Object {
	// 尝试查找 RuntimeException 类
	runtimeExceptionClass, ok := vm.getClassByName("System.RuntimeException")
	if !ok {
		// 如果找不到 RuntimeException，尝试查找 Exception 类
		exceptionClass, ok := vm.getClassByName("System.Exception")
		if !ok {
			// 尝试从全局查找
			if cls, ok := vm.globals["Exception"].(*interpreter.Class); ok {
				exceptionClass = cls
				ok = true
			}
		}
		
		if !ok {
			// 如果都找不到，返回错误对象
			return &interpreter.Error{Message: message}
		}
		runtimeExceptionClass = exceptionClass
	}

	// 创建异常实例
	instance := &interpreter.Instance{
		Class:  runtimeExceptionClass,
		Fields: make(map[string]interpreter.Object),
	}
	
	// 设置 message 字段
	instance.Fields["message"] = &interpreter.String{Value: message}
	
	// 设置 code 字段（默认为0）
	instance.Fields["code"] = &interpreter.Integer{Value: 0}
	
	return instance
}

// ========== VMError ==========

// VMError 虚拟机错误
type VMError struct {
	Message   string
	Exception interpreter.Object
	Line      int
}

func (e *VMError) Error() string {
	return e.Message
}

