package interpreter

// ========== 环境（作用域）系统 ==========

// Environment 环境（作用域）
// 用于管理变量的存储和查找
// 支持嵌套作用域（闭包）：内部作用域可以访问外部作用域的变量
// 类似于其他语言中的作用域链（scope chain）
type Environment struct {
	store map[string]Object // 当前作用域中的变量存储
	outer *Environment      // 外部作用域（用于实现闭包和嵌套作用域）
}

// NewEnvironment 创建新的环境（全局作用域）
// 返回:
//   一个新的环境实例，outer 为 nil（表示这是最外层作用域）
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

// NewEnclosedEnvironment 创建封闭环境（用于函数作用域）
// 用于创建函数内部的作用域，可以访问外部作用域的变量
// 参数:
//   outer: 外部环境（父作用域）
// 返回:
//   一个新的环境实例，outer 指向外部环境
// 用途:
//   当函数被调用时，会创建一个新的封闭环境
//   函数内部可以访问外部环境的变量（闭包）
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get 获取变量
// 首先在当前作用域查找，如果找不到则在外部作用域查找
// 参数:
//   name: 变量名
// 返回:
//   obj: 变量的值
//   ok: 是否找到变量
// 查找顺序:
//   1. 在当前作用域的 store 中查找
//   2. 如果找不到且存在外部作用域，则在外部作用域中递归查找
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		// 在当前作用域找不到，尝试在外部作用域查找
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set 设置变量
// 在当前作用域中设置变量
// 参数:
//   name: 变量名
//   val: 变量的值
// 返回:
//   设置的值
// 注意:
//   如果变量已存在，会被覆盖
//   不会在外部作用域中设置变量（这是设计选择，避免意外修改外部变量）
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
