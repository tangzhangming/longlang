package vm

import (
	"fmt"
	"strings"

	"github.com/tangzhangming/longlang/internal/interpreter"
)

// ========== 操作码定义 ==========

// Opcode 操作码类型
type Opcode byte

const (
	// 常量加载指令
	OP_CONST Opcode = iota // 加载常量（操作数：常量池索引）
	OP_NULL                // 加载 null
	OP_TRUE                // 加载 true
	OP_FALSE               // 加载 false

	// 变量操作指令
	OP_GET_LOCAL   // 获取局部变量（操作数：局部变量索引）
	OP_SET_LOCAL   // 设置局部变量（操作数：局部变量索引）
	OP_GET_GLOBAL  // 获取全局变量（操作数：常量池索引，变量名）
	OP_SET_GLOBAL  // 设置全局变量（操作数：常量池索引，变量名）
	OP_GET_UPVALUE // 获取闭包变量（操作数：upvalue 索引）
	OP_SET_UPVALUE // 设置闭包变量（操作数：upvalue 索引）
	OP_DEFINE_GLOBAL // 定义全局变量（操作数：常量池索引，变量名）

	// 算术运算指令
	OP_ADD // 加法
	OP_SUB // 减法
	OP_MUL // 乘法
	OP_DIV // 除法
	OP_MOD // 取模
	OP_NEG // 取负（一元）

	// 比较指令
	OP_EQ // 相等比较
	OP_NE // 不等比较
	OP_LT // 小于
	OP_LE // 小于等于
	OP_GT // 大于
	OP_GE // 大于等于

	// 逻辑运算指令
	OP_NOT // 逻辑非

	// 位运算指令
	OP_BIT_AND // 按位与
	OP_BIT_OR  // 按位或
	OP_BIT_XOR // 按位异或
	OP_BIT_NOT // 按位取反
	OP_LSHIFT  // 左移
	OP_RSHIFT  // 右移

	// 控制流指令
	OP_JUMP          // 无条件跳转（操作数：跳转偏移量，2字节）
	OP_JUMP_IF_FALSE // 条件跳转（如果栈顶为 false）
	OP_JUMP_IF_TRUE  // 条件跳转（如果栈顶为 true）
	OP_LOOP          // 循环跳转（向后跳转）

	// 函数相关指令
	OP_CALL       // 函数调用（操作数：参数个数）
	OP_RETURN     // 返回（弹出返回值）
	OP_CLOSURE    // 创建闭包（操作数：函数索引）
	OP_CLOSE_UPVALUE // 关闭 upvalue

	// 类和方法指令
	OP_CLASS        // 定义类（操作数：类名索引）
	OP_METHOD       // 定义方法（操作数：方法名索引）
	OP_STATIC_METHOD // 定义静态方法
	OP_CLASS_VAR    // 定义类变量（操作数：变量名索引）
	OP_STATIC_VAR   // 定义静态变量（操作数：变量名索引）
	OP_CLASS_CONST  // 定义类常量（操作数：常量名索引）
	OP_GET_PROPERTY // 获取属性（操作数：属性名索引）
	OP_SET_PROPERTY // 设置属性（操作数：属性名索引）
	OP_GET_STATIC_FIELD // 获取静态字段
	OP_SET_STATIC_FIELD // 设置静态字段
	OP_INVOKE       // 调用方法（操作数：方法名索引 + 参数个数）
	OP_INVOKE_STATIC // 调用静态方法
	OP_NEW          // 创建实例
	OP_INHERIT      // 继承
	OP_GET_SUPER    // 获取父类方法
	OP_SUPER_INVOKE // 调用父类方法

	// 数组和 Map 指令
	OP_ARRAY     // 创建数组（操作数：元素个数）
	OP_MAP       // 创建 Map（操作数：键值对个数）
	OP_INDEX     // 索引访问
	OP_INDEX_SET // 索引赋值
	OP_SLICE     // 切片操作
	OP_INSTANCE_OF // 检查实例是否属于类
	OP_TYPE_ASSERT // 类型断言（操作数：目标类型名索引 + 是否安全断言标志）

	// 异常处理指令
	OP_THROW        // 抛出异常
	OP_PUSH_TRY     // 开始 try 块（操作数：catch 块偏移量）
	OP_POP_TRY      // 结束 try 块
	OP_SETUP_FINALLY // 设置 finally 块

	// 协程指令
	OP_GO // 启动协程

	// 其他指令
	OP_POP  // 弹出栈顶值（丢弃）
	OP_DUP  // 复制栈顶值
	OP_SWAP // 交换栈顶两个值
	OP_PRINT // 打印（调试用）
	OP_HALT  // 停止执行

	// 内置函数调用
	OP_BUILTIN // 调用内置函数（操作数：内置函数ID + 参数个数）

	// 增量操作
	OP_INCREMENT // 自增
	OP_DECREMENT // 自减
)

// opcodeNames 操作码名称映射
var opcodeNames = map[Opcode]string{
	OP_CONST:          "OP_CONST",
	OP_NULL:           "OP_NULL",
	OP_TRUE:           "OP_TRUE",
	OP_FALSE:          "OP_FALSE",
	OP_GET_LOCAL:      "OP_GET_LOCAL",
	OP_SET_LOCAL:      "OP_SET_LOCAL",
	OP_GET_GLOBAL:     "OP_GET_GLOBAL",
	OP_SET_GLOBAL:     "OP_SET_GLOBAL",
	OP_GET_UPVALUE:    "OP_GET_UPVALUE",
	OP_SET_UPVALUE:    "OP_SET_UPVALUE",
	OP_DEFINE_GLOBAL:  "OP_DEFINE_GLOBAL",
	OP_ADD:            "OP_ADD",
	OP_SUB:            "OP_SUB",
	OP_MUL:            "OP_MUL",
	OP_DIV:            "OP_DIV",
	OP_MOD:            "OP_MOD",
	OP_NEG:            "OP_NEG",
	OP_EQ:             "OP_EQ",
	OP_NE:             "OP_NE",
	OP_LT:             "OP_LT",
	OP_LE:             "OP_LE",
	OP_GT:             "OP_GT",
	OP_GE:             "OP_GE",
	OP_NOT:            "OP_NOT",
	OP_BIT_AND:        "OP_BIT_AND",
	OP_BIT_OR:         "OP_BIT_OR",
	OP_BIT_XOR:        "OP_BIT_XOR",
	OP_BIT_NOT:        "OP_BIT_NOT",
	OP_LSHIFT:         "OP_LSHIFT",
	OP_RSHIFT:         "OP_RSHIFT",
	OP_JUMP:           "OP_JUMP",
	OP_JUMP_IF_FALSE:  "OP_JUMP_IF_FALSE",
	OP_JUMP_IF_TRUE:   "OP_JUMP_IF_TRUE",
	OP_LOOP:           "OP_LOOP",
	OP_CALL:           "OP_CALL",
	OP_RETURN:         "OP_RETURN",
	OP_CLOSURE:        "OP_CLOSURE",
	OP_CLOSE_UPVALUE:  "OP_CLOSE_UPVALUE",
	OP_CLASS:            "OP_CLASS",
	OP_METHOD:           "OP_METHOD",
	OP_STATIC_METHOD:    "OP_STATIC_METHOD",
	OP_CLASS_VAR:        "OP_CLASS_VAR",
	OP_STATIC_VAR:       "OP_STATIC_VAR",
	OP_CLASS_CONST:      "OP_CLASS_CONST",
	OP_GET_PROPERTY:     "OP_GET_PROPERTY",
	OP_SET_PROPERTY:     "OP_SET_PROPERTY",
	OP_GET_STATIC_FIELD: "OP_GET_STATIC_FIELD",
	OP_SET_STATIC_FIELD: "OP_SET_STATIC_FIELD",
	OP_INVOKE:           "OP_INVOKE",
	OP_INVOKE_STATIC:    "OP_INVOKE_STATIC",
	OP_NEW:            "OP_NEW",
	OP_INHERIT:        "OP_INHERIT",
	OP_GET_SUPER:      "OP_GET_SUPER",
	OP_SUPER_INVOKE:   "OP_SUPER_INVOKE",
	OP_ARRAY:          "OP_ARRAY",
	OP_MAP:            "OP_MAP",
	OP_INDEX:          "OP_INDEX",
	OP_INDEX_SET:      "OP_INDEX_SET",
	OP_SLICE:          "OP_SLICE",
	OP_INSTANCE_OF:    "OP_INSTANCE_OF",
	OP_TYPE_ASSERT:    "OP_TYPE_ASSERT",
	OP_THROW:          "OP_THROW",
	OP_PUSH_TRY:       "OP_PUSH_TRY",
	OP_POP_TRY:        "OP_POP_TRY",
	OP_SETUP_FINALLY:  "OP_SETUP_FINALLY",
	OP_GO:             "OP_GO",
	OP_POP:            "OP_POP",
	OP_DUP:            "OP_DUP",
	OP_SWAP:           "OP_SWAP",
	OP_PRINT:          "OP_PRINT",
	OP_HALT:           "OP_HALT",
	OP_BUILTIN:        "OP_BUILTIN",
	OP_INCREMENT:      "OP_INCREMENT",
	OP_DECREMENT:      "OP_DECREMENT",
}

// String 返回操作码的字符串表示
func (op Opcode) String() string {
	if name, ok := opcodeNames[op]; ok {
		return name
	}
	return fmt.Sprintf("OP_UNKNOWN(%d)", op)
}

// ========== 字节码结构 ==========

// Bytecode 字节码结构
// 包含常量池和指令序列
type Bytecode struct {
	Constants    []interpreter.Object // 常量池
	Instructions []byte               // 指令序列
	Lines        []int                // 行号信息（用于错误报告）
}

// NewBytecode 创建新的字节码结构
func NewBytecode() *Bytecode {
	return &Bytecode{
		Constants:    make([]interpreter.Object, 0),
		Instructions: make([]byte, 0),
		Lines:        make([]int, 0),
	}
}

// AddConstant 添加常量到常量池，返回常量索引
func (b *Bytecode) AddConstant(obj interpreter.Object) int {
	// 检查是否已存在相同常量（优化）
	for i, c := range b.Constants {
		if constantsEqual(c, obj) {
			return i
		}
	}
	b.Constants = append(b.Constants, obj)
	return len(b.Constants) - 1
}

// constantsEqual 判断两个常量是否相等
func constantsEqual(a, b interpreter.Object) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *interpreter.Integer:
		if bv, ok := b.(*interpreter.Integer); ok {
			return av.Value == bv.Value
		}
	case *interpreter.Float:
		if bv, ok := b.(*interpreter.Float); ok {
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
	}
	return false
}

// Emit 发出一条指令
func (b *Bytecode) Emit(op Opcode, line int) int {
	b.Instructions = append(b.Instructions, byte(op))
	b.Lines = append(b.Lines, line)
	return len(b.Instructions) - 1
}

// EmitWithOperand 发出带一个字节操作数的指令
func (b *Bytecode) EmitWithOperand(op Opcode, operand byte, line int) int {
	pos := b.Emit(op, line)
	b.Instructions = append(b.Instructions, operand)
	b.Lines = append(b.Lines, line)
	return pos
}

// EmitWithOperand16 发出带两个字节操作数的指令（16位）
func (b *Bytecode) EmitWithOperand16(op Opcode, operand uint16, line int) int {
	pos := b.Emit(op, line)
	b.Instructions = append(b.Instructions, byte(operand>>8), byte(operand&0xFF))
	b.Lines = append(b.Lines, line)
	b.Lines = append(b.Lines, line)
	return pos
}

// PatchJump 修补跳转指令的目标地址
func (b *Bytecode) PatchJump(offset int) {
	// 计算跳转距离（从跳转指令后面开始计算）
	jump := len(b.Instructions) - offset - 2
	if jump > 65535 {
		panic("跳转距离过大")
	}
	b.Instructions[offset] = byte(jump >> 8)
	b.Instructions[offset+1] = byte(jump & 0xFF)
}

// CurrentOffset 返回当前指令偏移量
func (b *Bytecode) CurrentOffset() int {
	return len(b.Instructions)
}

// Disassemble 反汇编字节码（用于调试）
func (b *Bytecode) Disassemble(name string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== %s ===\n", name))

	offset := 0
	for offset < len(b.Instructions) {
		offset = b.disassembleInstruction(&sb, offset)
	}

	return sb.String()
}

// disassembleInstruction 反汇编单条指令
func (b *Bytecode) disassembleInstruction(sb *strings.Builder, offset int) int {
	sb.WriteString(fmt.Sprintf("%04d ", offset))

	// 显示行号
	if offset > 0 && b.Lines[offset] == b.Lines[offset-1] {
		sb.WriteString("   | ")
	} else {
		sb.WriteString(fmt.Sprintf("%4d ", b.Lines[offset]))
	}

	op := Opcode(b.Instructions[offset])
	switch op {
	case OP_CONST:
		return b.constantInstruction(sb, op.String(), offset)
	case OP_GET_LOCAL, OP_SET_LOCAL:
		return b.byteInstruction(sb, op.String(), offset)
	case OP_GET_GLOBAL, OP_SET_GLOBAL, OP_DEFINE_GLOBAL:
		return b.constantInstruction(sb, op.String(), offset)
	case OP_GET_UPVALUE, OP_SET_UPVALUE:
		return b.byteInstruction(sb, op.String(), offset)
	case OP_JUMP, OP_JUMP_IF_FALSE, OP_JUMP_IF_TRUE:
		return b.jumpInstruction(sb, op.String(), 1, offset)
	case OP_LOOP:
		return b.jumpInstruction(sb, op.String(), -1, offset)
	case OP_CALL:
		return b.byteInstruction(sb, op.String(), offset)
	case OP_CLOSURE:
		return b.closureInstruction(sb, offset)
	case OP_CLASS, OP_METHOD, OP_STATIC_METHOD, OP_GET_PROPERTY, OP_SET_PROPERTY:
		return b.constantInstruction(sb, op.String(), offset)
	case OP_CLASS_VAR, OP_STATIC_VAR, OP_CLASS_CONST:
		return b.constantInstruction(sb, op.String(), offset)
	case OP_GET_STATIC_FIELD, OP_SET_STATIC_FIELD:
		return b.constantInstruction(sb, op.String(), offset)
	case OP_INVOKE, OP_INVOKE_STATIC, OP_SUPER_INVOKE:
		return b.invokeInstruction(sb, op.String(), offset)
	case OP_ARRAY, OP_MAP, OP_NEW:
		return b.byteInstruction(sb, op.String(), offset)
	case OP_PUSH_TRY:
		return b.jumpInstruction(sb, op.String(), 1, offset)
	case OP_INCREMENT, OP_DECREMENT:
		return b.byteInstruction(sb, op.String(), offset)
	default:
		sb.WriteString(fmt.Sprintf("%s\n", op.String()))
		return offset + 1
	}
}

// constantInstruction 反汇编常量指令
func (b *Bytecode) constantInstruction(sb *strings.Builder, name string, offset int) int {
	constant := b.Instructions[offset+1]
	sb.WriteString(fmt.Sprintf("%-16s %4d '", name, constant))
	if int(constant) < len(b.Constants) {
		sb.WriteString(b.Constants[constant].Inspect())
	}
	sb.WriteString("'\n")
	return offset + 2
}

// byteInstruction 反汇编字节操作数指令
func (b *Bytecode) byteInstruction(sb *strings.Builder, name string, offset int) int {
	slot := b.Instructions[offset+1]
	sb.WriteString(fmt.Sprintf("%-16s %4d\n", name, slot))
	return offset + 2
}

// jumpInstruction 反汇编跳转指令
func (b *Bytecode) jumpInstruction(sb *strings.Builder, name string, sign int, offset int) int {
	jump := int(b.Instructions[offset+1])<<8 | int(b.Instructions[offset+2])
	target := offset + 3 + sign*jump
	sb.WriteString(fmt.Sprintf("%-16s %4d -> %d\n", name, jump, target))
	return offset + 3
}

// closureInstruction 反汇编闭包指令
func (b *Bytecode) closureInstruction(sb *strings.Builder, offset int) int {
	constant := b.Instructions[offset+1]
	sb.WriteString(fmt.Sprintf("%-16s %4d ", "OP_CLOSURE", constant))
	if int(constant) < len(b.Constants) {
		sb.WriteString(b.Constants[constant].Inspect())
	}
	sb.WriteString("\n")

	// 读取 upvalue 信息
	offset += 2
	if fn, ok := b.Constants[constant].(*CompiledFunction); ok {
		for i := 0; i < fn.UpvalueCount; i++ {
			isLocal := b.Instructions[offset]
			index := b.Instructions[offset+1]
			localStr := "upvalue"
			if isLocal == 1 {
				localStr = "local"
			}
			sb.WriteString(fmt.Sprintf("%04d      |                     %s %d\n", offset, localStr, index))
			offset += 2
		}
	}
	return offset
}

// invokeInstruction 反汇编调用指令
func (b *Bytecode) invokeInstruction(sb *strings.Builder, name string, offset int) int {
	constant := b.Instructions[offset+1]
	argCount := b.Instructions[offset+2]
	sb.WriteString(fmt.Sprintf("%-16s (%d args) %4d '", name, argCount, constant))
	if int(constant) < len(b.Constants) {
		sb.WriteString(b.Constants[constant].Inspect())
	}
	sb.WriteString("'\n")
	return offset + 3
}

// ========== 编译后的函数对象 ==========

// CompiledFunction 编译后的函数
type CompiledFunction struct {
	Bytecode      *Bytecode // 函数的字节码
	NumLocals     int       // 局部变量数量
	NumParams     int       // 参数数量
	UpvalueCount  int       // upvalue 数量
	Name          string    // 函数名
	ClassName     string    // 所属类名（如果是方法）
	IsVariadic    bool      // 是否是可变参数函数
	IsConstructor bool      // 是否是构造函数
}

func (cf *CompiledFunction) Type() interpreter.ObjectType {
	return "COMPILED_FUNCTION"
}

func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("<fn %s>", cf.Name)
}

// ========== Upvalue 描述 ==========

// UpvalueDesc upvalue 描述（用于编译期间）
type UpvalueDesc struct {
	Index   int  // 索引
	IsLocal bool // 是否是局部变量（否则是外层 upvalue）
}

// ========== 异常处理信息 ==========

// TryBlock try 块信息
type TryBlock struct {
	TryStart     int // try 块开始位置
	TryEnd       int // try 块结束位置
	CatchStart   int // catch 块开始位置
	FinallyStart int // finally 块开始位置（-1 表示没有）
	StackDepth   int // 进入 try 时的栈深度
}

