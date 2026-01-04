package vm

import (
	"github.com/tangzhangming/longlang/internal/interpreter"
)

// ========== 作用域管理 ==========

// beginScope 开始新作用域（用于块语句）
func (c *Compiler) beginScope() {
	c.currentScope.scopeDepth++
}

// beginFunctionScope 开始新函数作用域
func (c *Compiler) beginFunctionScope() {
	newScope := &Scope{
		locals:     make([]Local, 0),
		upvalues:   make([]UpvalueDesc, 0),
		scopeDepth: 1, // 函数体本身是一个作用域
		parent:     c.currentScope,
	}
	c.scopeStack = append(c.scopeStack, c.currentScope)
	c.currentScope = newScope
}

// endFunctionScope 结束函数作用域，返回upvalues信息
func (c *Compiler) endFunctionScope() []UpvalueDesc {
	upvalues := make([]UpvalueDesc, len(c.currentScope.upvalues))
	copy(upvalues, c.currentScope.upvalues)
	
	// 恢复父作用域
	if len(c.scopeStack) > 0 {
		c.currentScope = c.scopeStack[len(c.scopeStack)-1]
		c.scopeStack = c.scopeStack[:len(c.scopeStack)-1]
	}
	
	return upvalues
}

// endScope 结束作用域
func (c *Compiler) endScope() {
	c.currentScope.scopeDepth--

	// 弹出局部变量
	for len(c.currentScope.locals) > 0 {
		local := c.currentScope.locals[len(c.currentScope.locals)-1]
		if local.depth <= c.currentScope.scopeDepth {
			break
		}

		if local.isCaptured {
			// 如果被闭包捕获，发出 CLOSE_UPVALUE 指令
			c.emit(OP_CLOSE_UPVALUE, 0)
		} else {
			// 否则直接弹出
			c.emit(OP_POP, 0)
		}

		c.currentScope.locals = c.currentScope.locals[:len(c.currentScope.locals)-1]
	}
}

// declareVariable 声明变量
func (c *Compiler) declareVariable(name string) {
	if c.currentScope.scopeDepth == 0 {
		return // 全局变量不需要声明
	}

	// 检查是否在当前作用域已声明
	for i := len(c.currentScope.locals) - 1; i >= 0; i-- {
		local := c.currentScope.locals[i]
		if local.depth != -1 && local.depth < c.currentScope.scopeDepth {
			break
		}
		if local.name == name {
			// 允许重新声明（遮蔽）
			return
		}
	}

	// 添加局部变量
	c.currentScope.locals = append(c.currentScope.locals, Local{
		name:  name,
		depth: -1, // 未初始化
	})
	
	if len(c.currentScope.locals) > c.currentScope.maxLocals {
		c.currentScope.maxLocals = len(c.currentScope.locals)
	}
}

// defineVariable 定义变量
func (c *Compiler) defineVariable(name string) {
	if c.currentScope.scopeDepth == 0 {
		return // 全局变量
	}

	// 标记为已初始化
	c.currentScope.locals[len(c.currentScope.locals)-1].depth = c.currentScope.scopeDepth
}

// resolveLocal 解析局部变量
func (c *Compiler) resolveLocal(name string) (int, bool) {
	for i := len(c.currentScope.locals) - 1; i >= 0; i-- {
		if c.currentScope.locals[i].name == name {
			// 即使 depth == -1 也返回，因为变量可能正在初始化
			// 调用者需要检查是否在初始化器中引用自身
			return i, true
		}
	}
	return -1, false
}

// resolveUpvalue 解析 upvalue
func (c *Compiler) resolveUpvalue(name string) (int, bool) {
	if c.currentScope.parent == nil {
		return -1, false
	}

	// 在父作用域的局部变量中查找
	for i := len(c.currentScope.parent.locals) - 1; i >= 0; i-- {
		if c.currentScope.parent.locals[i].name == name {
			c.currentScope.parent.locals[i].isCaptured = true
			return c.addUpvalue(i, true), true
		}
	}

	// 在父作用域的 upvalue 中查找
	for i, upvalue := range c.currentScope.parent.upvalues {
		if c.currentScope.parent.locals[upvalue.Index].name == name {
			return c.addUpvalue(i, false), true
		}
	}

	return -1, false
}

// addUpvalue 添加 upvalue
func (c *Compiler) addUpvalue(index int, isLocal bool) int {
	// 检查是否已存在
	for i, upvalue := range c.currentScope.upvalues {
		if upvalue.Index == index && upvalue.IsLocal == isLocal {
			return i
		}
	}

	// 添加新的 upvalue
	c.currentScope.upvalues = append(c.currentScope.upvalues, UpvalueDesc{
		Index:   index,
		IsLocal: isLocal,
	})

	return len(c.currentScope.upvalues) - 1
}

// ========== 循环管理 ==========

// pushLoop 压入循环
func (c *Compiler) pushLoop(start int) {
	c.loopStack = append(c.loopStack, &LoopInfo{
		start:      start,
		breakJumps: make([]int, 0),
		scopeDepth: c.currentScope.scopeDepth,
	})
}

// popLoop 弹出循环
func (c *Compiler) popLoop() {
	c.loopStack = c.loopStack[:len(c.loopStack)-1]
}

// patchBreaks 修补 break 跳转
func (c *Compiler) patchBreaks() {
	if len(c.loopStack) == 0 {
		return
	}
	loop := c.loopStack[len(c.loopStack)-1]
	for _, jump := range loop.breakJumps {
		c.patchJump(jump)
	}
}

// ========== 字节码发出 ==========

// emit 发出指令
func (c *Compiler) emit(op Opcode, line int) int {
	return c.bytecode.Emit(op, line)
}

// emitWithOperand 发出带操作数的指令
func (c *Compiler) emitWithOperand(op Opcode, operand byte, line int) int {
	return c.bytecode.EmitWithOperand(op, operand, line)
}

// emitJump 发出跳转指令
func (c *Compiler) emitJump(op Opcode, line int) int {
	c.emit(op, line)
	// 预留两个字节给跳转偏移量
	c.bytecode.Instructions = append(c.bytecode.Instructions, 0xFF, 0xFF)
	c.bytecode.Lines = append(c.bytecode.Lines, line, line)
	return len(c.bytecode.Instructions) - 2
}

// patchJump 修补跳转
func (c *Compiler) patchJump(offset int) {
	c.bytecode.PatchJump(offset)
}

// emitLoop 发出循环跳转
func (c *Compiler) emitLoop(loopStart, line int) {
	c.emit(OP_LOOP, line)
	// 计算跳转距离
	offset := c.currentOffset() - loopStart + 2
	c.bytecode.Instructions = append(c.bytecode.Instructions, byte(offset>>8), byte(offset&0xFF))
	c.bytecode.Lines = append(c.bytecode.Lines, line, line)
}

// emitReturn 发出返回指令
func (c *Compiler) emitReturn() {
	c.emit(OP_NULL, 0)
	c.emit(OP_RETURN, 0)
}

// currentOffset 获取当前偏移量
func (c *Compiler) currentOffset() int {
	return c.bytecode.CurrentOffset()
}

// addConstant 添加常量
func (c *Compiler) addConstant(obj interpreter.Object) int {
	return c.bytecode.AddConstant(obj)
}

