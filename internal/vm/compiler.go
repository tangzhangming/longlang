package vm

import (
	"fmt"
	"os"

	"github.com/tangzhangming/longlang/internal/interpreter"
	"github.com/tangzhangming/longlang/internal/parser"
)

// ========== 编译器 ==========

// Compiler 字节码编译器
type Compiler struct {
	bytecode         *Bytecode           // 当前正在编译的字节码
	scopeStack       []*Scope            // 作用域栈
	currentScope     *Scope              // 当前作用域
	loopStack        []*LoopInfo         // 循环栈（用于 break/continue）
	classStack       []*ClassInfo        // 类编译栈
	globals          map[string]int      // 全局变量索引
	vm               *VM                 // 关联的虚拟机（用于运行时加载）
	currentNamespace string              // 当前命名空间
}

// Scope 作用域
type Scope struct {
	locals     []Local          // 局部变量
	maxLocals  int              // 最大局部变量数量
	upvalues   []UpvalueDesc    // upvalue 描述
	scopeDepth int              // 作用域深度
	parent     *Scope           // 父作用域
	function   *CompiledFunction // 当前编译的函数
}

// Local 局部变量
type Local struct {
	name      string // 变量名
	depth     int    // 作用域深度
	isCaptured bool  // 是否被闭包捕获
}

// LoopInfo 循环信息
type LoopInfo struct {
	start      int   // 循环开始位置
	breakJumps []int // break 跳转位置列表
	scopeDepth int   // 循环的作用域深度
}

// ClassInfo 类编译信息
type ClassInfo struct {
	name       string
	hasSuperclass bool
}

// NewCompiler 创建新的编译器
func NewCompiler() *Compiler {
	c := &Compiler{
		bytecode:         NewBytecode(),
		scopeStack:       make([]*Scope, 0),
		loopStack:        make([]*LoopInfo, 0),
		classStack:       make([]*ClassInfo, 0),
		globals:          make(map[string]int),
		currentNamespace: "",
	}

	// 创建全局作用域
	c.currentScope = &Scope{
		locals:     make([]Local, 0),
		upvalues:   make([]UpvalueDesc, 0),
		scopeDepth: 0,
		function:   nil,
	}

	return c
}

// SetVM 设置关联的虚拟机
func (c *Compiler) SetVM(vm *VM) {
	c.vm = vm
}

// Compile 编译 AST
func (c *Compiler) Compile(program *parser.Program) (*Bytecode, error) {
	for _, stmt := range program.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return nil, err
		}
	}

	// 添加 HALT 指令
	c.emit(OP_HALT, 0)

	return c.bytecode, nil
}

// CompileFunction 编译函数
func (c *Compiler) CompileFunction(fn *parser.FunctionLiteral) (*CompiledFunction, []UpvalueDesc, error) {
	return c.CompileFunctionWithContext(fn, false)
}

func (c *Compiler) CompileFunctionWithContext(fn *parser.FunctionLiteral, isInstanceMethod bool) (*CompiledFunction, []UpvalueDesc, error) {
	// 保存当前字节码
	prevBytecode := c.bytecode
	c.bytecode = NewBytecode()

	// 创建新函数作用域
	c.beginFunctionScope()
	
	// 对于实例方法，首先声明 this 变量（槽位 0）
	numParams := len(fn.Parameters)
	if isInstanceMethod {
		c.declareVariable("this")
		c.defineVariable("this")
		numParams++ // this 算作一个参数
	}
	
	// 处理参数
	for _, param := range fn.Parameters {
		c.declareVariable(param.Name.Value)
		c.defineVariable(param.Name.Value)
	}

	// 编译函数体 - 不再调用 beginScope，因为函数作用域已经是 scopeDepth=1
	for _, stmt := range fn.Body.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return nil, nil, err
		}
	}

	// 如果最后没有 return，添加隐式 return null
	c.emitReturn()

	// 记录局部变量数量
	numLocals := c.currentScope.maxLocals
	if c.vm != nil && c.vm.debug {
		name := ""
		if fn.Name != nil {
			name = fn.Name.Value
		}
		fmt.Fprintf(os.Stderr, "COMPLIED FUNCTION %s, NumLocals: %d\n", name, numLocals)
	}

	// 结束函数作用域 - 获取upvalues信息
	upvalues := c.endFunctionScope()

	// 处理默认参数值
	var defaultValues []interpreter.Object
	for _, param := range fn.Parameters {
		if param.DefaultValue != nil {
			// 计算默认值（必须是常量表达式）
			defaultVal := evalConstantExpression(param.DefaultValue)
			defaultValues = append(defaultValues, defaultVal)
		} else {
			defaultValues = append(defaultValues, nil)
		}
	}

	// 创建编译后的函数
	compiledFn := &CompiledFunction{
		Bytecode:      c.bytecode,
		NumLocals:     numLocals,
		NumParams:     numParams,
		UpvalueCount:  len(upvalues),
		Name:          "",
		DefaultValues: defaultValues,
	}

	if fn.Name != nil {
		compiledFn.Name = fn.Name.Value
	}

	// 恢复字节码
	c.bytecode = prevBytecode

	return compiledFn, upvalues, nil
}

// ========== 语句编译 ==========

// compileStatement 编译语句
func (c *Compiler) compileStatement(stmt parser.Statement) error {
	switch s := stmt.(type) {
	case *parser.LetStatement:
		return c.compileLetStatement(s)
	case *parser.AssignStatement:
		return c.compileAssignStatement(s)
	case *parser.ExpressionStatement:
		return c.compileExpressionStatement(s)
	case *parser.ReturnStatement:
		return c.compileReturnStatement(s)
	case *parser.BlockStatement:
		return c.compileBlockStatement(s)
	case *parser.IfStatement:
		return c.compileIfStatement(s)
	case *parser.ForStatement:
		return c.compileForStatement(s)
	case *parser.ForRangeStatement:
		return c.compileForRangeStatement(s)
	case *parser.BreakStatement:
		return c.compileBreakStatement(s)
	case *parser.ContinueStatement:
		return c.compileContinueStatement(s)
	case *parser.SwitchStatement:
		return c.compileSwitchStatement(s)
	case *parser.TryStatement:
		return c.compileTryStatement(s)
	case *parser.ThrowStatement:
		return c.compileThrowStatement(s)
	case *parser.ClassStatement:
		return c.compileClassStatement(s)
	case *parser.IncrementStatement:
		return c.compileIncrementStatement(s)
	case *parser.NamespaceStatement:
		return c.compileNamespaceStatement(s)
	case *parser.UseStatement:
		return c.compileUseStatement(s)
	case *parser.EnumStatement:
		return c.compileEnumStatement(s)
	case *parser.InterfaceStatement:
		return c.compileInterfaceStatement(s)
	case *parser.GoStatement:
		return c.compileGoStatement(s)
	default:
		return fmt.Errorf("不支持的语句类型: %T", stmt)
	}
}

// compileLetStatement 编译变量声明
func (c *Compiler) compileLetStatement(stmt *parser.LetStatement) error {
	// 编译初始值
	if stmt.Value != nil {
		if err := c.compileExpression(stmt.Value); err != nil {
			return err
		}
	} else {
		c.emit(OP_NULL, stmt.Token.Line)
	}

	// 声明变量
	if c.currentScope.scopeDepth > 0 {
		// 局部变量
		c.declareVariable(stmt.Name.Value)
		// 存储到局部变量
		slot, ok := c.resolveLocal(stmt.Name.Value)
		if !ok {
			return fmt.Errorf("无法解析局部变量: %s", stmt.Name.Value)
		}
		c.emitWithOperand(OP_SET_LOCAL, byte(slot), stmt.Token.Line)
		c.defineVariable(stmt.Name.Value)
	} else {
		// 全局变量
		nameIndex := c.addConstant(&interpreter.String{Value: stmt.Name.Value})
		c.emitWithOperand(OP_DEFINE_GLOBAL, byte(nameIndex), stmt.Token.Line)
	}

	return nil
}

// compileAssignStatement 编译短变量声明
func (c *Compiler) compileAssignStatement(stmt *parser.AssignStatement) error {
	// 编译值
	if err := c.compileExpression(stmt.Value); err != nil {
		return err
	}

	// 声明变量
	if c.currentScope.scopeDepth > 0 {
		c.declareVariable(stmt.Name.Value)
		// 存储到局部变量
		slot, ok := c.resolveLocal(stmt.Name.Value)
		if !ok {
			return fmt.Errorf("无法解析局部变量: %s", stmt.Name.Value)
		}
		c.emitWithOperand(OP_SET_LOCAL, byte(slot), stmt.Token.Line)
		c.defineVariable(stmt.Name.Value)
	} else {
		// 全局变量
		nameIndex := c.addConstant(&interpreter.String{Value: stmt.Name.Value})
		c.emitWithOperand(OP_DEFINE_GLOBAL, byte(nameIndex), stmt.Token.Line)
	}

	return nil
}

// compileExpressionStatement 编译表达式语句
func (c *Compiler) compileExpressionStatement(stmt *parser.ExpressionStatement) error {
	if err := c.compileExpression(stmt.Expression); err != nil {
		return err
	}
	c.emit(OP_POP, 0)
	return nil
}

// compileReturnStatement 编译返回语句
func (c *Compiler) compileReturnStatement(stmt *parser.ReturnStatement) error {
	if stmt.ReturnValue != nil {
		if err := c.compileExpression(stmt.ReturnValue); err != nil {
			return err
		}
	} else {
		c.emit(OP_NULL, stmt.Token.Line)
	}
	c.emit(OP_RETURN, stmt.Token.Line)
	return nil
}

// compileBlockStatement 编译块语句
func (c *Compiler) compileBlockStatement(block *parser.BlockStatement) error {
	c.beginScope()
	for _, stmt := range block.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	c.endScope()
	return nil
}

// compileIfStatement 编译 if 语句
func (c *Compiler) compileIfStatement(stmt *parser.IfStatement) error {
	// 编译条件
	if err := c.compileExpression(stmt.Condition); err != nil {
		return err
	}

	// 条件跳转
	thenJump := c.emitJump(OP_JUMP_IF_FALSE, stmt.Token.Line)
	c.emit(OP_POP, stmt.Token.Line) // 弹出条件值

	// 编译 then 分支
	if err := c.compileStatement(stmt.Consequence); err != nil {
		return err
	}

	// 跳过 else 分支
	elseJump := c.emitJump(OP_JUMP, stmt.Token.Line)

	// 修补 then 跳转
	c.patchJump(thenJump)
	c.emit(OP_POP, stmt.Token.Line) // 弹出条件值

	// 编译 else 分支
	if stmt.Alternative != nil {
		if err := c.compileStatement(stmt.Alternative); err != nil {
			return err
		}
	} else if stmt.ElseIf != nil {
		if err := c.compileIfStatement(stmt.ElseIf); err != nil {
			return err
		}
	}

	// 修补 else 跳转
	c.patchJump(elseJump)

	return nil
}

// compileForStatement 编译 for 循环
func (c *Compiler) compileForStatement(stmt *parser.ForStatement) error {
	c.beginScope()

	// 初始化
	if stmt.Init != nil {
		if err := c.compileStatement(stmt.Init); err != nil {
			return err
		}
	}

	// 记录循环开始位置
	loopStart := c.currentOffset()
	c.pushLoop(loopStart)

	// 条件
	exitJump := -1
	if stmt.Condition != nil {
		if err := c.compileExpression(stmt.Condition); err != nil {
			return err
		}
		exitJump = c.emitJump(OP_JUMP_IF_FALSE, stmt.Token.Line)
		c.emit(OP_POP, stmt.Token.Line)
	}

	// 循环体
	if err := c.compileStatement(stmt.Body); err != nil {
		return err
	}

	// 增量
	if stmt.Post != nil {
		if err := c.compileStatement(stmt.Post); err != nil {
			return err
		}
	}

	// 跳回循环开始
	c.emitLoop(loopStart, stmt.Token.Line)

	// 修补退出跳转
	if exitJump != -1 {
		c.patchJump(exitJump)
		c.emit(OP_POP, stmt.Token.Line)
	}

	// 修补 break 跳转
	c.patchBreaks()
	c.popLoop()
	c.endScope()

	return nil
}

// compileForRangeStatement 编译 for-range 循环
// 语法: for key, value := range iterable { ... } 或 for value := range iterable { ... }
func (c *Compiler) compileForRangeStatement(stmt *parser.ForRangeStatement) error {
	c.beginScope()

	// 编译可迭代对象并存储到局部变量
	if err := c.compileExpression(stmt.Iterable); err != nil {
		return err
	}
	c.declareVariable("__iterable__")
	iterableSlot, _ := c.resolveLocal("__iterable__")
	c.emitWithOperand(OP_SET_LOCAL, byte(iterableSlot), stmt.Token.Line)
	c.defineVariable("__iterable__")

	// 创建迭代索引变量并初始化为 0
	indexConst := c.addConstant(&interpreter.Integer{Value: 0})
	c.emitWithOperand(OP_CONST, byte(indexConst), stmt.Token.Line)
	c.declareVariable("__index__")
	indexSlot, _ := c.resolveLocal("__index__")
	c.emitWithOperand(OP_SET_LOCAL, byte(indexSlot), stmt.Token.Line)
	c.defineVariable("__index__")

	// 预先声明循环变量（在循环外部），初始化为 null
	var valueSlot int
	var keySlot int
	
	if stmt.Value != nil {
		// 双变量形式：value 变量
		c.emit(OP_NULL, stmt.Token.Line)
		c.declareVariable(stmt.Value.Value)
		valueSlot, _ = c.resolveLocal(stmt.Value.Value)
		c.emitWithOperand(OP_SET_LOCAL, byte(valueSlot), stmt.Token.Line)
		c.defineVariable(stmt.Value.Value)
		
		// key 变量
		c.emit(OP_NULL, stmt.Token.Line)
		c.declareVariable(stmt.Key.Value)
		keySlot, _ = c.resolveLocal(stmt.Key.Value)
		c.emitWithOperand(OP_SET_LOCAL, byte(keySlot), stmt.Token.Line)
		c.defineVariable(stmt.Key.Value)
	} else {
		// 单变量形式：Key 存储的是值变量名
		c.emit(OP_NULL, stmt.Token.Line)
		c.declareVariable(stmt.Key.Value)
		valueSlot, _ = c.resolveLocal(stmt.Key.Value)
		c.emitWithOperand(OP_SET_LOCAL, byte(valueSlot), stmt.Token.Line)
		c.defineVariable(stmt.Key.Value)
	}

	// 记录循环开始位置
	loopStart := c.currentOffset()
	c.pushLoop(loopStart)

	// === 检查是否结束：index < len(iterable) ===
	// 获取 __index__
	c.emitWithOperand(OP_GET_LOCAL, byte(indexSlot), stmt.Token.Line)

	// 获取 __iterable__ 的长度
	c.emitWithOperand(OP_GET_LOCAL, byte(iterableSlot), stmt.Token.Line)

	// 调用内置 len 函数
	lenNameIdx := c.addConstant(&interpreter.String{Value: "len"})
	c.emitWithOperand(OP_GET_GLOBAL, byte(lenNameIdx), stmt.Token.Line)
	// 交换顺序：现在栈上是 [index, iterable, len]，需要 [index, len, iterable]
	c.emit(OP_SWAP, stmt.Token.Line) // [index, len, iterable]
	c.emitWithOperand(OP_CALL, 1, stmt.Token.Line) // 调用 len(iterable)
	// 现在栈上是 [index, length]

	// index < length
	c.emit(OP_LT, stmt.Token.Line)

	// 如果 false，跳出循环
	exitJump := c.emitJump(OP_JUMP_IF_FALSE, stmt.Token.Line)
	c.emit(OP_POP, stmt.Token.Line) // 弹出比较结果

	// === 获取当前元素并更新循环变量 ===
	// 获取 __iterable__[__index__]
	c.emitWithOperand(OP_GET_LOCAL, byte(iterableSlot), stmt.Token.Line)
	c.emitWithOperand(OP_GET_LOCAL, byte(indexSlot), stmt.Token.Line)
	c.emit(OP_INDEX, stmt.Token.Line)
	// 存储到 value 变量
	c.emitWithOperand(OP_SET_LOCAL, byte(valueSlot), stmt.Token.Line)
	c.emit(OP_POP, stmt.Token.Line) // 弹出赋值结果

	if stmt.Value != nil {
		// 双变量形式：更新 key 为当前索引
		c.emitWithOperand(OP_GET_LOCAL, byte(indexSlot), stmt.Token.Line)
		c.emitWithOperand(OP_SET_LOCAL, byte(keySlot), stmt.Token.Line)
		c.emit(OP_POP, stmt.Token.Line) // 弹出赋值结果
	}

	// === 循环体 ===
	if err := c.compileStatement(stmt.Body); err != nil {
		return err
	}

	// === 增加索引：__index__++ ===
	c.emitWithOperand(OP_GET_LOCAL, byte(indexSlot), stmt.Token.Line)
	oneConst := c.addConstant(&interpreter.Integer{Value: 1})
	c.emitWithOperand(OP_CONST, byte(oneConst), stmt.Token.Line)
	c.emit(OP_ADD, stmt.Token.Line)
	c.emitWithOperand(OP_SET_LOCAL, byte(indexSlot), stmt.Token.Line)
	c.emit(OP_POP, stmt.Token.Line) // 弹出赋值结果

	// 跳回循环开始
	c.emitLoop(loopStart, stmt.Token.Line)

	// 修补退出跳转
	c.patchJump(exitJump)
	c.emit(OP_POP, stmt.Token.Line) // 弹出比较结果

	// 修补 break 跳转
	c.patchBreaks()
	c.popLoop()
	c.endScope()

	return nil
}

// compileBreakStatement 编译 break 语句
func (c *Compiler) compileBreakStatement(stmt *parser.BreakStatement) error {
	if len(c.loopStack) == 0 {
		return fmt.Errorf("break 只能在循环中使用")
	}

	// 发出跳转指令，稍后修补
	jump := c.emitJump(OP_JUMP, stmt.Token.Line)
	c.loopStack[len(c.loopStack)-1].breakJumps = append(
		c.loopStack[len(c.loopStack)-1].breakJumps, jump)

	return nil
}

// compileContinueStatement 编译 continue 语句
func (c *Compiler) compileContinueStatement(stmt *parser.ContinueStatement) error {
	if len(c.loopStack) == 0 {
		return fmt.Errorf("continue 只能在循环中使用")
	}

	// 跳回循环开始
	loopStart := c.loopStack[len(c.loopStack)-1].start
	c.emitLoop(loopStart, stmt.Token.Line)

	return nil
}

// compileSwitchStatement 编译 switch 语句
func (c *Compiler) compileSwitchStatement(stmt *parser.SwitchStatement) error {
	// 编译 switch 值
	if stmt.Value != nil {
		if err := c.compileExpression(stmt.Value); err != nil {
			return err
		}
	}

	endJumps := make([]int, 0)

	// 编译每个 case
	for _, caseClause := range stmt.Cases {
		if stmt.Value != nil {
			// 复制 switch 值用于比较
			c.emit(OP_DUP, stmt.Token.Line)

			// 编译 case 值（可能有多个）
			if len(caseClause.Values) > 0 {
				// 编译第一个值
				if err := c.compileExpression(caseClause.Values[0]); err != nil {
					return err
				}
				c.emit(OP_EQ, stmt.Token.Line)

				// 编译其他值（用 OR 连接）
				for i := 1; i < len(caseClause.Values); i++ {
					c.emit(OP_DUP, stmt.Token.Line) // 复制 switch 值
					if err := c.compileExpression(caseClause.Values[i]); err != nil {
						return err
					}
					c.emit(OP_EQ, stmt.Token.Line)
					// OR 操作（简化处理）
				}
			}
		} else if caseClause.IsCondition && caseClause.Condition != nil {
			// 条件 switch
			if err := c.compileExpression(caseClause.Condition); err != nil {
				return err
			}
		}

		// 如果不匹配，跳到下一个 case
		nextCase := c.emitJump(OP_JUMP_IF_FALSE, stmt.Token.Line)
		c.emit(OP_POP, stmt.Token.Line) // 弹出比较结果
		if stmt.Value != nil {
			c.emit(OP_POP, stmt.Token.Line) // 弹出 switch 值副本
		}

		// 编译 case 体
		if caseClause.Body != nil {
			for _, bodyStmt := range caseClause.Body.Statements {
				if err := c.compileStatement(bodyStmt); err != nil {
					return err
				}
			}
		}

		// 跳到 switch 结束
		endJumps = append(endJumps, c.emitJump(OP_JUMP, stmt.Token.Line))

		// 修补 nextCase 跳转
		c.patchJump(nextCase)
		c.emit(OP_POP, stmt.Token.Line) // 弹出比较结果
	}

	// 编译 default
	if stmt.Value != nil {
		c.emit(OP_POP, stmt.Token.Line) // 弹出 switch 值
	}
	if stmt.Default != nil {
		for _, bodyStmt := range stmt.Default.Statements {
			if err := c.compileStatement(bodyStmt); err != nil {
				return err
			}
		}
	}

	// 修补所有结束跳转
	for _, jump := range endJumps {
		c.patchJump(jump)
	}

	return nil
}

// compileTryStatement 编译 try 语句
func (c *Compiler) compileTryStatement(stmt *parser.TryStatement) error {
	// 发出 PUSH_TRY 指令
	tryJump := c.emitJump(OP_PUSH_TRY, stmt.Token.Line)

	// 编译 try 块
	if err := c.compileStatement(stmt.TryBlock); err != nil {
		return err
	}

	// 发出 POP_TRY 指令
	c.emit(OP_POP_TRY, stmt.Token.Line)

	// 跳过 catch 块
	var endJumps []int
	endJumps = append(endJumps, c.emitJump(OP_JUMP, stmt.Token.Line))

	// 修补 try 跳转（指向 catch 块开始）
	c.patchJump(tryJump)

	// 此时栈顶是异常对象
	// 编译 catch 块
	var nextCatchJump int = -1

	for i, catchClause := range stmt.CatchClauses {
		if nextCatchJump != -1 {
			c.patchJump(nextCatchJump)
			c.emit(OP_POP, catchClause.Token.Line) // 弹出上一个 catch 检查留下的 false
			nextCatchJump = -1
		}

		// 如果有类型限制
		if catchClause.ExceptionType != nil {
			c.emit(OP_DUP, catchClause.Token.Line) // 复制异常对象进行检查
			// 加载异常类
			if err := c.compileExpression(catchClause.ExceptionType); err != nil {
				return err
			}
			c.emit(OP_INSTANCE_OF, catchClause.Token.Line)
			nextCatchJump = c.emitJump(OP_JUMP_IF_FALSE, catchClause.Token.Line)
			c.emit(OP_POP, catchClause.Token.Line) // 弹出 true
		}

		c.beginScope()
		// 此时栈顶是异常对象
		// 将异常对象存储到变量并弹出
		c.declareVariable(catchClause.ExceptionVar.Value)
		c.defineVariable(catchClause.ExceptionVar.Value)
		
		slot, _ := c.resolveLocal(catchClause.ExceptionVar.Value)
		c.emitWithOperand(OP_SET_LOCAL, byte(slot), catchClause.Token.Line)
		c.emit(OP_POP, catchClause.Token.Line) // 弹出异常对象

		// 编译 catch 体
		if err := c.compileStatement(catchClause.Body); err != nil {
			return err
		}

		c.endScope()

		// 跳到 try-catch 结束
		endJumps = append(endJumps, c.emitJump(OP_JUMP, catchClause.Token.Line))
		
		if i == len(stmt.CatchClauses)-1 {
			// 最后一个 catch 的 nextCatchJump 需要修补到 re-throw
			if nextCatchJump != -1 {
				c.patchJump(nextCatchJump)
				c.emit(OP_POP, stmt.Token.Line) // 弹出 false
				nextCatchJump = -1
				c.emit(OP_THROW, stmt.Token.Line) // 没有匹配的 catch，重新抛出
			}
		}
	}

	// 如果没有任何 catch 块（虽然语法上不常见），直接 re-throw
	if len(stmt.CatchClauses) == 0 {
		c.emit(OP_THROW, stmt.Token.Line)
	}

	// 修补所有结束跳转
	for _, jump := range endJumps {
		c.patchJump(jump)
	}

	// 编译 finally 块
	if stmt.FinallyBlock != nil {
		if err := c.compileStatement(stmt.FinallyBlock); err != nil {
			return err
		}
	}

	return nil
}

// compileThrowStatement 编译 throw 语句
func (c *Compiler) compileThrowStatement(stmt *parser.ThrowStatement) error {
	if err := c.compileExpression(stmt.Value); err != nil {
		return err
	}
	c.emit(OP_THROW, stmt.Token.Line)
	return nil
}

// compileClassStatement 编译类定义
func (c *Compiler) compileClassStatement(stmt *parser.ClassStatement) error {
	// 类名常量
	nameIndex := c.addConstant(&interpreter.String{Value: stmt.Name.Value})

	// 发出 CLASS 指令
	c.emitWithOperand(OP_CLASS, byte(nameIndex), stmt.Token.Line)

	// 声明类变量
	if c.currentScope.scopeDepth > 0 {
		c.declareVariable(stmt.Name.Value)
	} else {
		c.emitWithOperand(OP_DEFINE_GLOBAL, byte(nameIndex), stmt.Token.Line)
	}

	// 压入类编译上下文
	c.classStack = append(c.classStack, &ClassInfo{name: stmt.Name.Value})

	// 处理继承
	if stmt.Parent != nil {
		// 加载父类
		if err := c.compileExpression(stmt.Parent); err != nil {
			return err
		}
		// 获取当前类
		c.emitWithOperand(OP_GET_GLOBAL, byte(nameIndex), stmt.Token.Line)
		// 继承
		c.emit(OP_INHERIT, stmt.Token.Line)
		c.classStack[len(c.classStack)-1].hasSuperclass = true
	}

	// 获取类（用于添加方法）
	c.emitWithOperand(OP_GET_GLOBAL, byte(nameIndex), stmt.Token.Line)

	// 编译成员
	for _, member := range stmt.Members {
		switch m := member.(type) {
		case *parser.ClassMethod:
			if err := c.compileClassMethod(m); err != nil {
				return err
			}
		case *parser.ClassVariable:
			if err := c.compileClassVariable(m); err != nil {
				return err
			}
		case *parser.ClassConstant:
			if err := c.compileClassConstant(m); err != nil {
				return err
			}
		}
	}

	// 弹出类
	c.emit(OP_POP, stmt.Token.Line)

	// 弹出类编译上下文
	c.classStack = c.classStack[:len(c.classStack)-1]

	return nil
}

// compileClassVariable 编译类变量
func (c *Compiler) compileClassVariable(variable *parser.ClassVariable) error {
	// 变量名常量
	nameIndex := c.addConstant(&interpreter.String{Value: variable.Name.Value})
	
	// 编译默认值（如果有）
	if variable.Value != nil {
		if err := c.compileExpression(variable.Value); err != nil {
			return err
		}
	} else {
		// 没有默认值，使用 null
		c.emit(OP_NULL, variable.Token.Line)
	}
	
	// 发出变量定义指令
	if variable.IsStatic {
		c.emitWithOperand(OP_STATIC_VAR, byte(nameIndex), variable.Token.Line)
	} else {
		c.emitWithOperand(OP_CLASS_VAR, byte(nameIndex), variable.Token.Line)
	}
	
	return nil
}

// compileClassConstant 编译类常量
func (c *Compiler) compileClassConstant(constant *parser.ClassConstant) error {
	// 常量名
	nameIndex := c.addConstant(&interpreter.String{Value: constant.Name.Value})
	
	// 编译常量值
	if err := c.compileExpression(constant.Value); err != nil {
		return err
	}
	
	// 发出常量定义指令
	c.emitWithOperand(OP_CLASS_CONST, byte(nameIndex), constant.Token.Line)
	
	return nil
}

// compileClassMethod 编译类方法
func (c *Compiler) compileClassMethod(method *parser.ClassMethod) error {
	// 编译方法体为闭包
	fn := &parser.FunctionLiteral{
		Token:      method.Token,
		Name:       method.Name,
		Parameters: method.Parameters,
		ReturnType: method.ReturnType,
		Body:       method.Body,
	}

	// 实例方法需要 this 参数，静态方法不需要
	isInstanceMethod := !method.IsStatic
	compiledFn, upvalues, err := c.CompileFunctionWithContext(fn, isInstanceMethod)
	if err != nil {
		return err
	}

	// 设置所属类名
	if len(c.classStack) > 0 {
		name := c.classStack[len(c.classStack)-1].name
		if c.currentNamespace != "" {
			name = c.currentNamespace + "." + name
		}
		compiledFn.ClassName = name
	}

	// 添加到常量池
	fnIndex := c.addConstant(compiledFn)
	
	// 发出闭包指令（支持大索引）
	if fnIndex > 255 {
		c.emitWithOperand16(OP_CLOSURE_WIDE, uint16(fnIndex), method.Token.Line)
	} else {
		c.emitWithOperand(OP_CLOSURE, byte(fnIndex), method.Token.Line)
	}
	
	// 发出 upvalue 信息
	for _, upvalue := range upvalues {
		if upvalue.IsLocal {
			c.bytecode.Instructions = append(c.bytecode.Instructions, 1)
		} else {
			c.bytecode.Instructions = append(c.bytecode.Instructions, 0)
		}
		c.bytecode.Instructions = append(c.bytecode.Instructions, byte(upvalue.Index))
		c.bytecode.Lines = append(c.bytecode.Lines, method.Token.Line, method.Token.Line)
	}

	// 方法名常量
	nameIndex := c.addConstant(&interpreter.String{Value: method.Name.Value})

	// 发出方法定义指令（支持大索引）
	if method.IsStatic {
		if nameIndex > 255 {
			c.emitWithOperand16(OP_STATIC_METHOD_WIDE, uint16(nameIndex), method.Token.Line)
		} else {
			c.emitWithOperand(OP_STATIC_METHOD, byte(nameIndex), method.Token.Line)
		}
	} else {
		if nameIndex > 255 {
			c.emitWithOperand16(OP_METHOD_WIDE, uint16(nameIndex), method.Token.Line)
		} else {
			c.emitWithOperand(OP_METHOD, byte(nameIndex), method.Token.Line)
		}
	}

	return nil
}

// compileIncrementStatement 编译自增/自减语句
func (c *Compiler) compileIncrementStatement(stmt *parser.IncrementStatement) error {
	// 查找变量
	if slot, ok := c.resolveLocal(stmt.Name.Value); ok {
		if stmt.Operator == "++" {
			c.emitWithOperand(OP_INCREMENT, byte(slot), stmt.Token.Line)
		} else {
			c.emitWithOperand(OP_DECREMENT, byte(slot), stmt.Token.Line)
		}
	} else {
		// 全局变量
		nameIndex := c.addConstant(&interpreter.String{Value: stmt.Name.Value})
		c.emitWithOperand(OP_GET_GLOBAL, byte(nameIndex), stmt.Token.Line)
		c.emit(OP_CONST, stmt.Token.Line)
		oneConst := c.addConstant(&interpreter.Integer{Value: 1})
		c.bytecode.Instructions = append(c.bytecode.Instructions, byte(oneConst))
		if stmt.Operator == "++" {
			c.emit(OP_ADD, stmt.Token.Line)
		} else {
			c.emit(OP_SUB, stmt.Token.Line)
		}
		c.emitWithOperand(OP_SET_GLOBAL, byte(nameIndex), stmt.Token.Line)
	}
	return nil
}

// compileGoStatement 编译 go 语句
func (c *Compiler) compileGoStatement(stmt *parser.GoStatement) error {
	// 编译函数调用表达式
	call, ok := stmt.Call.(*parser.CallExpression)
	if !ok {
		return fmt.Errorf("go 后面必须是函数调用")
	}

	// 编译函数
	if err := c.compileExpression(call.Function); err != nil {
		return err
	}

	// 编译参数
	for _, arg := range call.Arguments {
		if err := c.compileExpression(arg.Value); err != nil {
			return err
		}
	}

	// 发出 GO 指令
	c.emitWithOperand(OP_GO, byte(len(call.Arguments)), stmt.Token.Line)

	return nil
}

// compileNamespaceStatement 编译命名空间声明
func (c *Compiler) compileNamespaceStatement(stmt *parser.NamespaceStatement) error {
	// 记录当前命名空间
	c.currentNamespace = stmt.Name.Value

	// 如果有关联的 VM，设置当前命名空间
	if c.vm != nil {
		// 如果有项目配置，使用 root_namespace 解析命名空间
		namespaceName := c.currentNamespace
		if c.vm.projectConfig != nil {
			namespaceName = c.vm.projectConfig.ResolveNamespace(namespaceName)
		}
		c.vm.SetCurrentNamespace(namespaceName)
	}

	return nil
}

// compileUseStatement 编译 use 语句
func (c *Compiler) compileUseStatement(stmt *parser.UseStatement) error {
	fullPath := stmt.Path.Value
	var alias string
	if stmt.Alias != nil {
		alias = stmt.Alias.Value
	}

	// 如果有关联的 VM，在运行时加载
	if c.vm != nil {
		if err := c.vm.ProcessUseStatement(fullPath, alias); err != nil {
			return err
		}
	}

	return nil
}

// compileEnumStatement 编译枚举声明
func (c *Compiler) compileEnumStatement(stmt *parser.EnumStatement) error {
	enumName := stmt.Name.Value

	// 创建枚举对象
	enum := &interpreter.Enum{
		Name:       enumName,
		Members:    make(map[string]*interpreter.EnumValue),
		Methods:    make(map[string]*interpreter.ClassMethod),
		IsPublic:   stmt.IsPublic,
		IsInternal: stmt.IsInternal,
		Namespace:  c.currentNamespace,
	}

	// 编译枚举成员
	for i, member := range stmt.Members {
		var value interpreter.Object = &interpreter.Integer{Value: int64(i)}
		if member.Value != nil {
			// 如果有显式值，求值
			if intLit, ok := member.Value.(*parser.IntegerLiteral); ok {
				value = &interpreter.Integer{Value: intLit.Value}
			}
		}

		enumValue := &interpreter.EnumValue{
			Name:    member.Name.Value,
			Ordinal: i,
			Value:   value,
			Enum:    enum,
		}
		enum.Members[member.Name.Value] = enumValue
	}

	// 编译枚举方法
	for _, method := range stmt.Methods {
		// 转换参数类型
		params := make([]interface{}, len(method.Parameters))
		for i, p := range method.Parameters {
			params[i] = p
		}
		// 转换返回类型
		returnTypes := make([]string, len(method.ReturnType))
		for i, rt := range method.ReturnType {
			returnTypes[i] = rt.Value
		}

		classMethod := &interpreter.ClassMethod{
			Name:           method.Name.Value,
			Parameters:     params,
			Body:           method.Body,
			ReturnType:     returnTypes,
			AccessModifier: method.AccessModifier,
			IsStatic:       method.IsStatic,
		}
		enum.Methods[method.Name.Value] = classMethod
	}

	// 将枚举添加到常量池并定义为全局变量
	enumIndex := c.addConstant(enum)
	c.emitWithOperand(OP_CONST, byte(enumIndex), stmt.Token.Line)
	nameIndex := c.addConstant(&interpreter.String{Value: enumName})
	c.emitWithOperand(OP_DEFINE_GLOBAL, byte(nameIndex), stmt.Token.Line)

	// 如果有关联的 VM，注册到命名空间
	if c.vm != nil && c.vm.currentNamespace != nil {
		c.vm.currentNamespace.SetEnum(enumName, enum)
	}

	return nil
}

// compileInterfaceStatement 编译接口声明
func (c *Compiler) compileInterfaceStatement(stmt *parser.InterfaceStatement) error {
	interfaceName := stmt.Name.Value

	// 创建接口对象
	iface := &interpreter.Interface{
		Name:       interfaceName,
		Methods:    make(map[string]*interpreter.InterfaceMethod),
		IsPublic:   stmt.IsPublic,
		IsInternal: stmt.IsInternal,
		Namespace:  c.currentNamespace,
	}

	// 编译接口方法
	for _, method := range stmt.Methods {
		// 转换参数类型
		params := make([]string, len(method.Parameters))
		for i, p := range method.Parameters {
			if p.Type != nil {
				params[i] = p.Type.Value
			} else {
				params[i] = "any"
			}
		}
		// 转换返回类型
		returnTypes := make([]string, len(method.ReturnType))
		for i, rt := range method.ReturnType {
			returnTypes[i] = rt.Value
		}

		ifaceMethod := &interpreter.InterfaceMethod{
			Name:       method.Name.Value,
			Parameters: params,
			ReturnType: returnTypes,
		}
		iface.Methods[method.Name.Value] = ifaceMethod
	}

	// 将接口添加到常量池并定义为全局变量
	ifaceIndex := c.addConstant(iface)
	c.emitWithOperand(OP_CONST, byte(ifaceIndex), stmt.Token.Line)
	nameIndex := c.addConstant(&interpreter.String{Value: interfaceName})
	c.emitWithOperand(OP_DEFINE_GLOBAL, byte(nameIndex), stmt.Token.Line)

	// 如果有关联的 VM，注册到命名空间
	if c.vm != nil && c.vm.currentNamespace != nil {
		c.vm.currentNamespace.SetInterface(interfaceName, iface)
	}

	return nil
}

