package vm

import (
	"github.com/tangzhangming/longlang/internal/interpreter"
)

// ========== 调用栈帧 ==========

// Frame 调用栈帧
// 存储函数调用的上下文信息
type Frame struct {
	closure      *Closure              // 当前执行的闭包
	ip           int                   // 指令指针（Instruction Pointer）
	basePointer  int                   // 基址指针，指向栈中局部变量的起始位置
	locals       []interpreter.Object  // 局部变量（备用，主要使用栈）
	isMethodCall bool                  // 是否是方法调用（方法调用不需要弹出函数对象）
	isConstructor bool                 // 是否是构造函数（构造函数返回时需要返回实例）
	calledClassName string             // 被调用的类名（用于 Late Static Binding）
	constructorInstance *interpreter.Instance // 正在构造的实例
}

// NewFrame 创建新的调用栈帧
func NewFrame(closure *Closure, basePointer int) *Frame {
	return &Frame{
		closure:     closure,
		ip:          0,
		basePointer: basePointer,
	}
}

// Instructions 获取当前帧的指令序列
func (f *Frame) Instructions() []byte {
	return f.closure.Fn.Bytecode.Instructions
}

// ReadByte 读取一个字节并移动 IP
func (f *Frame) ReadByte() byte {
	b := f.closure.Fn.Bytecode.Instructions[f.ip]
	f.ip++
	return b
}

// ReadUint16 读取两个字节（大端序）并移动 IP
func (f *Frame) ReadUint16() uint16 {
	high := f.closure.Fn.Bytecode.Instructions[f.ip]
	low := f.closure.Fn.Bytecode.Instructions[f.ip+1]
	f.ip += 2
	return uint16(high)<<8 | uint16(low)
}

// ReadConstant 读取常量（从常量池）
func (f *Frame) ReadConstant() interpreter.Object {
	index := f.ReadByte()
	return f.closure.Fn.Bytecode.Constants[index]
}

// ReadConstant16 读取常量（16位索引）
func (f *Frame) ReadConstant16() interpreter.Object {
	index := f.ReadUint16()
	return f.closure.Fn.Bytecode.Constants[index]
}

// GetLine 获取当前行号（用于错误报告）
func (f *Frame) GetLine() int {
	if f.ip > 0 && f.ip <= len(f.closure.Fn.Bytecode.Lines) {
		return f.closure.Fn.Bytecode.Lines[f.ip-1]
	}
	return 0
}

// ========== 闭包 ==========

// Closure 闭包对象
// 包含编译后的函数和捕获的 upvalue
type Closure struct {
	Fn       *CompiledFunction // 编译后的函数
	Upvalues []*Upvalue        // 捕获的 upvalue
}

func (c *Closure) Type() interpreter.ObjectType {
	return "CLOSURE"
}

func (c *Closure) Inspect() string {
	return c.Fn.Inspect()
}

// NewClosure 创建新的闭包
func NewClosure(fn *CompiledFunction) *Closure {
	upvalues := make([]*Upvalue, fn.UpvalueCount)
	return &Closure{
		Fn:       fn,
		Upvalues: upvalues,
	}
}

// ========== Upvalue ==========

// Upvalue 捕获的变量
// 当闭包捕获外部变量时，使用 upvalue 存储
type Upvalue struct {
	Value      interpreter.Object  // 当变量关闭后，值存储在这里
	Location   *interpreter.Object // 指向栈上的位置（未关闭时使用）
	StackIndex int                 // 栈索引（用于比较和查找）
	Closed     bool                // 是否已关闭
	Next       *Upvalue            // 链表指针（用于管理开放的 upvalue）
}

// NewUpvalue 创建新的 upvalue
func NewUpvalue(location *interpreter.Object) *Upvalue {
	return &Upvalue{
		Location: location,
		Closed:   false,
	}
}

// NewUpvalueWithIndex 创建带索引的 upvalue
func NewUpvalueWithIndex(location *interpreter.Object, stackIndex int) *Upvalue {
	return &Upvalue{
		Location:   location,
		StackIndex: stackIndex,
		Closed:     false,
	}
}

// Get 获取 upvalue 的值
func (u *Upvalue) Get() interpreter.Object {
	if u.Closed {
		if u.Value == nil {
			return &interpreter.Integer{Value: 0}
		}
		return u.Value
	}
	if u.Location == nil || *u.Location == nil {
		return &interpreter.Integer{Value: 0}
	}
	return *u.Location
}

// Set 设置 upvalue 的值
func (u *Upvalue) Set(value interpreter.Object) {
	if u.Closed {
		u.Value = value
	} else {
		*u.Location = value
	}
}

// Close 关闭 upvalue（将值从栈复制到堆）
func (u *Upvalue) Close() {
	u.Value = *u.Location
	u.Closed = true
	u.Location = nil
}

// ========== 调用信息 ==========

// CallInfo 调用信息（用于错误堆栈）
type CallInfo struct {
	FunctionName string // 函数名
	ClassName    string // 类名（如果是方法）
	FileName     string // 文件名
	Line         int    // 行号
}

// ========== 异常状态 ==========

// TryState try 块状态
type TryState struct {
	CatchTarget   int                  // catch 块目标地址
	FinallyTarget int                  // finally 块目标地址（-1 表示没有）
	StackDepth    int                  // 进入 try 时的栈深度
	FrameIndex    int                  // 进入 try 时的帧索引
	ExceptionVar  string               // 异常变量名
	Exception     interpreter.Object   // 捕获的异常
}

// ========== 绑定方法 ==========

// BoundMethod 绑定方法（实例方法调用）
type BoundMethod struct {
	Receiver interpreter.Object // 接收者（实例）
	Method   *Closure           // 方法闭包
}

func (bm *BoundMethod) Type() interpreter.ObjectType {
	return "BOUND_METHOD"
}

func (bm *BoundMethod) Inspect() string {
	return "<bound method>"
}

