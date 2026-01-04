package vm

import (
	"fmt"

	"github.com/tangzhangming/longlang/internal/interpreter"
	"github.com/tangzhangming/longlang/internal/parser"
)

// ========== 表达式编译 ==========

// compileExpression 编译表达式
func (c *Compiler) compileExpression(expr parser.Expression) error {
	switch e := expr.(type) {
	case *parser.IntegerLiteral:
		return c.compileIntegerLiteral(e)
	case *parser.FloatLiteral:
		return c.compileFloatLiteral(e)
	case *parser.StringLiteral:
		return c.compileStringLiteral(e)
	case *parser.BooleanLiteral:
		return c.compileBooleanLiteral(e)
	case *parser.NullLiteral:
		return c.compileNullLiteral(e)
	case *parser.Identifier:
		return c.compileIdentifier(e)
	case *parser.PrefixExpression:
		return c.compilePrefixExpression(e)
	case *parser.InfixExpression:
		return c.compileInfixExpression(e)
	case *parser.CallExpression:
		return c.compileCallExpression(e)
	case *parser.FunctionLiteral:
		return c.compileFunctionLiteral(e)
	case *parser.ArrayLiteral:
		return c.compileArrayLiteral(e)
	case *parser.TypedArrayLiteral:
		return c.compileTypedArrayLiteral(e)
	case *parser.MapLiteral:
		return c.compileMapLiteral(e)
	case *parser.IndexExpression:
		return c.compileIndexExpression(e)
	case *parser.MemberAccessExpression:
		return c.compileMemberAccessExpression(e)
	case *parser.AssignmentExpression:
		return c.compileAssignmentExpression(e)
	case *parser.CompoundAssignmentExpression:
		return c.compileCompoundAssignmentExpression(e)
	case *parser.NewExpression:
		return c.compileNewExpression(e)
	case *parser.TernaryExpression:
		return c.compileTernaryExpression(e)
	case *parser.ThisExpression:
		return c.compileThisExpression(e)
	case *parser.SuperExpression:
		return c.compileSuperExpression(e)
	case *parser.StaticCallExpression:
		return c.compileStaticCallExpression(e)
	case *parser.StaticAccessExpression:
		return c.compileStaticAccessExpression(e)
	case *parser.ClassLiteralExpression:
		return c.compileClassLiteralExpression(e)
	case *parser.InterpolatedStringLiteral:
		return c.compileInterpolatedString(e)
	case *parser.SliceExpression:
		return c.compileSliceExpression(e)
	case *parser.TypeAssertionExpression:
		return c.compileTypeAssertionExpression(e)
	default:
		return fmt.Errorf("不支持的表达式类型: %T", expr)
	}
}

// compileIntegerLiteral 编译整数字面量
func (c *Compiler) compileIntegerLiteral(lit *parser.IntegerLiteral) error {
	index := c.addConstant(&interpreter.Integer{Value: lit.Value})
	c.emitWithOperand(OP_CONST, byte(index), lit.Token.Line)
	return nil
}

// compileFloatLiteral 编译浮点数字面量
func (c *Compiler) compileFloatLiteral(lit *parser.FloatLiteral) error {
	index := c.addConstant(&interpreter.Float{Value: lit.Value})
	c.emitWithOperand(OP_CONST, byte(index), lit.Token.Line)
	return nil
}

// compileStringLiteral 编译字符串字面量
func (c *Compiler) compileStringLiteral(lit *parser.StringLiteral) error {
	index := c.addConstant(&interpreter.String{Value: lit.Value})
	c.emitWithOperand(OP_CONST, byte(index), lit.Token.Line)
	return nil
}

// compileBooleanLiteral 编译布尔字面量
func (c *Compiler) compileBooleanLiteral(lit *parser.BooleanLiteral) error {
	if lit.Value {
		c.emit(OP_TRUE, lit.Token.Line)
	} else {
		c.emit(OP_FALSE, lit.Token.Line)
	}
	return nil
}

// compileNullLiteral 编译 null 字面量
func (c *Compiler) compileNullLiteral(lit *parser.NullLiteral) error {
	c.emit(OP_NULL, lit.Token.Line)
	return nil
}

// compileIdentifier 编译标识符
func (c *Compiler) compileIdentifier(ident *parser.Identifier) error {
	// 尝试作为局部变量
	if slot, ok := c.resolveLocal(ident.Value); ok {
		c.emitWithOperand(OP_GET_LOCAL, byte(slot), ident.Token.Line)
		return nil
	}

	// 尝试作为 upvalue
	if slot, ok := c.resolveUpvalue(ident.Value); ok {
		c.emitWithOperand(OP_GET_UPVALUE, byte(slot), ident.Token.Line)
		return nil
	}

	// 作为全局变量
	nameIndex := c.addConstant(&interpreter.String{Value: ident.Value})
	c.emitWithOperand(OP_GET_GLOBAL, byte(nameIndex), ident.Token.Line)
	return nil
}

// compilePrefixExpression 编译前缀表达式
func (c *Compiler) compilePrefixExpression(expr *parser.PrefixExpression) error {
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	switch expr.Operator {
	case "-":
		c.emit(OP_NEG, expr.Token.Line)
	case "!":
		c.emit(OP_NOT, expr.Token.Line)
	case "~":
		c.emit(OP_BIT_NOT, expr.Token.Line)
	default:
		return fmt.Errorf("未知的前缀运算符: %s", expr.Operator)
	}

	return nil
}

// compileInfixExpression 编译中缀表达式
func (c *Compiler) compileInfixExpression(expr *parser.InfixExpression) error {
	// 特殊处理短路求值
	if expr.Operator == "&&" {
		return c.compileAndExpression(expr)
	}
	if expr.Operator == "||" {
		return c.compileOrExpression(expr)
	}

	// 编译左操作数
	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}

	// 编译右操作数
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	// 发出运算指令
	switch expr.Operator {
	case "+":
		c.emit(OP_ADD, expr.Token.Line)
	case "-":
		c.emit(OP_SUB, expr.Token.Line)
	case "*":
		c.emit(OP_MUL, expr.Token.Line)
	case "/":
		c.emit(OP_DIV, expr.Token.Line)
	case "%":
		c.emit(OP_MOD, expr.Token.Line)
	case "==":
		c.emit(OP_EQ, expr.Token.Line)
	case "!=":
		c.emit(OP_NE, expr.Token.Line)
	case "<":
		c.emit(OP_LT, expr.Token.Line)
	case "<=":
		c.emit(OP_LE, expr.Token.Line)
	case ">":
		c.emit(OP_GT, expr.Token.Line)
	case ">=":
		c.emit(OP_GE, expr.Token.Line)
	case "&":
		c.emit(OP_BIT_AND, expr.Token.Line)
	case "|":
		c.emit(OP_BIT_OR, expr.Token.Line)
	case "^":
		c.emit(OP_BIT_XOR, expr.Token.Line)
	case "<<":
		c.emit(OP_LSHIFT, expr.Token.Line)
	case ">>":
		c.emit(OP_RSHIFT, expr.Token.Line)
	default:
		return fmt.Errorf("未知的中缀运算符: %s", expr.Operator)
	}

	return nil
}

// compileAndExpression 编译逻辑与表达式（短路求值）
func (c *Compiler) compileAndExpression(expr *parser.InfixExpression) error {
	// 编译左操作数
	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}

	// 如果左操作数为 false，跳过右操作数
	endJump := c.emitJump(OP_JUMP_IF_FALSE, expr.Token.Line)

	// 弹出左操作数
	c.emit(OP_POP, expr.Token.Line)

	// 编译右操作数
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	// 修补跳转
	c.patchJump(endJump)

	return nil
}

// compileOrExpression 编译逻辑或表达式（短路求值）
func (c *Compiler) compileOrExpression(expr *parser.InfixExpression) error {
	// 编译左操作数
	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}

	// 如果左操作数为 true，跳过右操作数
	endJump := c.emitJump(OP_JUMP_IF_TRUE, expr.Token.Line)

	// 弹出左操作数
	c.emit(OP_POP, expr.Token.Line)

	// 编译右操作数
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	// 修补跳转
	c.patchJump(endJump)

	return nil
}

// compileCallExpression 编译函数调用
func (c *Compiler) compileCallExpression(expr *parser.CallExpression) error {
	// 检查是否是 super 方法调用 super.method() 或 super::method()
	if memberAccess, ok := expr.Function.(*parser.MemberAccessExpression); ok {
		if _, ok := memberAccess.Object.(*parser.SuperExpression); ok {
			// 1. 加载 this
			c.emitWithOperand(OP_GET_LOCAL, 0, expr.Token.Line)

			// 2. 加载参数
			for _, arg := range expr.Arguments {
				if err := c.compileExpression(arg.Value); err != nil {
					return err
				}
			}

			// 3. 加载 super class
			nameIndex := c.addConstant(&interpreter.String{Value: "super"})
			c.emitWithOperand(OP_GET_GLOBAL, byte(nameIndex), expr.Token.Line)

			// 4. 发出 SUPER_INVOKE 指令
			methodNameIndex := c.addConstant(&interpreter.String{Value: memberAccess.Member.Value})
			c.emitWithOperand(OP_SUPER_INVOKE, byte(methodNameIndex), expr.Token.Line)
			c.bytecode.Instructions = append(c.bytecode.Instructions, byte(len(expr.Arguments)))
			c.bytecode.Lines = append(c.bytecode.Lines, expr.Token.Line)
			return nil
		}

		// 普通方法调用
		// 编译对象
		if err := c.compileExpression(memberAccess.Object); err != nil {
			return err
		}

		// 编译参数
		for _, arg := range expr.Arguments {
			if err := c.compileExpression(arg.Value); err != nil {
				return err
			}
		}

		// 方法名常量
		nameIndex := c.addConstant(&interpreter.String{Value: memberAccess.Member.Value})

		// 发出 INVOKE 指令
		c.emitWithOperand(OP_INVOKE, byte(nameIndex), expr.Token.Line)
		c.bytecode.Instructions = append(c.bytecode.Instructions, byte(len(expr.Arguments)))
		c.bytecode.Lines = append(c.bytecode.Lines, expr.Token.Line)

		return nil
	}

	// 普通函数调用
	if err := c.compileExpression(expr.Function); err != nil {
		return err
	}

	// 编译参数
	for _, arg := range expr.Arguments {
		if err := c.compileExpression(arg.Value); err != nil {
			return err
		}
	}

	// 发出 CALL 指令
	c.emitWithOperand(OP_CALL, byte(len(expr.Arguments)), expr.Token.Line)

	return nil
}

// compileFunctionLiteral 编译函数字面量
func (c *Compiler) compileFunctionLiteral(fn *parser.FunctionLiteral) error {
	compiledFn, upvalues, err := c.CompileFunction(fn)
	if err != nil {
		return err
	}

	// 添加到常量池
	fnIndex := c.addConstant(compiledFn)
	c.emitWithOperand(OP_CLOSURE, byte(fnIndex), fn.Token.Line)

	// 发出 upvalue 信息 - 使用编译函数返回的upvalues
	for _, upvalue := range upvalues {
		if upvalue.IsLocal {
			c.bytecode.Instructions = append(c.bytecode.Instructions, 1)
		} else {
			c.bytecode.Instructions = append(c.bytecode.Instructions, 0)
		}
		c.bytecode.Instructions = append(c.bytecode.Instructions, byte(upvalue.Index))
		c.bytecode.Lines = append(c.bytecode.Lines, fn.Token.Line)
		c.bytecode.Lines = append(c.bytecode.Lines, fn.Token.Line)
	}

	// 如果是命名函数（顶级函数声明），自动定义全局变量
	if fn.Name != nil && fn.Name.Value != "" {
		nameIndex := c.addConstant(&interpreter.String{Value: fn.Name.Value})
		c.emitWithOperand(OP_DEFINE_GLOBAL, byte(nameIndex), fn.Token.Line)
		// 不需要弹出，因为 DEFINE_GLOBAL 会弹出值
		// 但是 ExpressionStatement 会再发出一个 POP，所以需要压入一个占位值
		c.emit(OP_NULL, fn.Token.Line)
	}

	return nil
}

// compileArrayLiteral 编译数组字面量
func (c *Compiler) compileArrayLiteral(arr *parser.ArrayLiteral) error {
	// 编译元素
	for _, elem := range arr.Elements {
		if err := c.compileExpression(elem); err != nil {
			return err
		}
	}

	// 发出 ARRAY 指令
	c.emitWithOperand(OP_ARRAY, byte(len(arr.Elements)), arr.Token.Line)

	return nil
}

// compileTypedArrayLiteral 编译类型化数组字面量
func (c *Compiler) compileTypedArrayLiteral(arr *parser.TypedArrayLiteral) error {
	// 编译元素
	for _, elem := range arr.Elements {
		if err := c.compileExpression(elem); err != nil {
			return err
		}
	}

	// 发出 ARRAY 指令
	c.emitWithOperand(OP_ARRAY, byte(len(arr.Elements)), arr.Token.Line)

	return nil
}

// compileMapLiteral 编译 Map 字面量
func (c *Compiler) compileMapLiteral(m *parser.MapLiteral) error {
	// 编译键值对
	for _, key := range m.Keys {
		// 编译键
		if err := c.compileExpression(key); err != nil {
			return err
		}
		// 编译值
		if err := c.compileExpression(m.Pairs[key]); err != nil {
			return err
		}
	}

	// 发出 MAP 指令
	c.emitWithOperand(OP_MAP, byte(len(m.Keys)), m.Token.Line)

	return nil
}

// compileIndexExpression 编译索引表达式
func (c *Compiler) compileIndexExpression(expr *parser.IndexExpression) error {
	// 编译对象
	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}

	// 编译索引
	if err := c.compileExpression(expr.Index); err != nil {
		return err
	}

	// 发出 INDEX 指令
	c.emit(OP_INDEX, expr.Token.Line)

	return nil
}

// compileMemberAccessExpression 编译成员访问表达式
func (c *Compiler) compileMemberAccessExpression(expr *parser.MemberAccessExpression) error {
	// 编译对象
	if err := c.compileExpression(expr.Object); err != nil {
		return err
	}

	// 属性名常量
	nameIndex := c.addConstant(&interpreter.String{Value: expr.Member.Value})

	// 发出 GET_PROPERTY 指令
	c.emitWithOperand(OP_GET_PROPERTY, byte(nameIndex), expr.Token.Line)

	return nil
}

// compileAssignmentExpression 编译赋值表达式
func (c *Compiler) compileAssignmentExpression(expr *parser.AssignmentExpression) error {
	// 编译值
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	// 根据左值类型发出不同指令
	switch left := expr.Left.(type) {
	case *parser.Identifier:
		// 变量赋值
		if slot, ok := c.resolveLocal(left.Value); ok {
			c.emitWithOperand(OP_SET_LOCAL, byte(slot), expr.Token.Line)
		} else if slot, ok := c.resolveUpvalue(left.Value); ok {
			c.emitWithOperand(OP_SET_UPVALUE, byte(slot), expr.Token.Line)
		} else {
			nameIndex := c.addConstant(&interpreter.String{Value: left.Value})
			c.emitWithOperand(OP_SET_GLOBAL, byte(nameIndex), expr.Token.Line)
		}

	case *parser.MemberAccessExpression:
		// 属性赋值
		if err := c.compileExpression(left.Object); err != nil {
			return err
		}
		nameIndex := c.addConstant(&interpreter.String{Value: left.Member.Value})
		c.emitWithOperand(OP_SET_PROPERTY, byte(nameIndex), expr.Token.Line)

	case *parser.IndexExpression:
		// 索引赋值
		if err := c.compileExpression(left.Left); err != nil {
			return err
		}
		if err := c.compileExpression(left.Index); err != nil {
			return err
		}
		c.emit(OP_INDEX_SET, expr.Token.Line)

	case *parser.StaticAccessExpression:
		// 静态字段赋值
		if err := c.compileExpression(left.ClassName); err != nil {
			return err
		}
		nameIndex := c.addConstant(&interpreter.String{Value: left.Name.Value})
		c.emitWithOperand(OP_SET_STATIC_FIELD, byte(nameIndex), expr.Token.Line)

	default:
		return fmt.Errorf("不支持的赋值目标: %T", left)
	}

	return nil
}

// compileCompoundAssignmentExpression 编译复合赋值表达式
func (c *Compiler) compileCompoundAssignmentExpression(expr *parser.CompoundAssignmentExpression) error {
	// 获取运算符
	op := expr.Operator[:len(expr.Operator)-1] // "+=" -> "+"

	// 根据左值类型处理
	switch left := expr.Left.(type) {
	case *parser.Identifier:
		// 获取当前值
		if slot, ok := c.resolveLocal(left.Value); ok {
			c.emitWithOperand(OP_GET_LOCAL, byte(slot), expr.Token.Line)
		} else if slot, ok := c.resolveUpvalue(left.Value); ok {
			c.emitWithOperand(OP_GET_UPVALUE, byte(slot), expr.Token.Line)
		} else {
			nameIndex := c.addConstant(&interpreter.String{Value: left.Value})
			c.emitWithOperand(OP_GET_GLOBAL, byte(nameIndex), expr.Token.Line)
		}

		// 编译右值
		if err := c.compileExpression(expr.Right); err != nil {
			return err
		}

		// 发出运算指令
		c.emitBinaryOp(op, expr.Token.Line)

		// 赋值
		if slot, ok := c.resolveLocal(left.Value); ok {
			c.emitWithOperand(OP_SET_LOCAL, byte(slot), expr.Token.Line)
		} else if slot, ok := c.resolveUpvalue(left.Value); ok {
			c.emitWithOperand(OP_SET_UPVALUE, byte(slot), expr.Token.Line)
		} else {
			nameIndex := c.addConstant(&interpreter.String{Value: left.Value})
			c.emitWithOperand(OP_SET_GLOBAL, byte(nameIndex), expr.Token.Line)
		}

	default:
		return fmt.Errorf("不支持的复合赋值目标: %T", left)
	}

	return nil
}

// compileNewExpression 编译 new 表达式
func (c *Compiler) compileNewExpression(expr *parser.NewExpression) error {
	// 编译类名
	if err := c.compileExpression(expr.ClassName); err != nil {
		return err
	}

	// 编译参数
	for _, arg := range expr.Arguments {
		if err := c.compileExpression(arg.Value); err != nil {
			return err
		}
	}

	// 发出 NEW 指令
	c.emitWithOperand(OP_NEW, byte(len(expr.Arguments)), expr.Token.Line)

	return nil
}

// compileTernaryExpression 编译三目运算符
func (c *Compiler) compileTernaryExpression(expr *parser.TernaryExpression) error {
	// 编译条件
	if err := c.compileExpression(expr.Condition); err != nil {
		return err
	}

	// 条件跳转
	elseJump := c.emitJump(OP_JUMP_IF_FALSE, expr.Token.Line)
	c.emit(OP_POP, expr.Token.Line)

	// 编译 true 分支
	if err := c.compileExpression(expr.TrueExpr); err != nil {
		return err
	}

	// 跳过 false 分支
	endJump := c.emitJump(OP_JUMP, expr.Token.Line)

	// 修补 else 跳转
	c.patchJump(elseJump)
	c.emit(OP_POP, expr.Token.Line)

	// 编译 false 分支
	if err := c.compileExpression(expr.FalseExpr); err != nil {
		return err
	}

	// 修补结束跳转
	c.patchJump(endJump)

	return nil
}

// compileThisExpression 编译 this 表达式
func (c *Compiler) compileThisExpression(expr *parser.ThisExpression) error {
	// 查找 this 的槽位
	slot, ok := c.resolveLocal("this")
	if !ok {
		return fmt.Errorf("'this' 只能在实例方法中使用")
	}
	c.emitWithOperand(OP_GET_LOCAL, byte(slot), expr.Token.Line)
	return nil
}

// compileSuperExpression 编译 super 表达式
func (c *Compiler) compileSuperExpression(expr *parser.SuperExpression) error {
	// 获取 this
	c.emitWithOperand(OP_GET_LOCAL, 0, expr.Token.Line)
	// 获取父类
	nameIndex := c.addConstant(&interpreter.String{Value: "super"})
	c.emitWithOperand(OP_GET_GLOBAL, byte(nameIndex), expr.Token.Line)
	return nil
}

// compileStaticCallExpression 编译静态方法调用
func (c *Compiler) compileStaticCallExpression(expr *parser.StaticCallExpression) error {
	// 检查是否是 super 调用 super::method()
	if expr.ClassName.Value == "super" {
		// 1. 加载 this
		c.emitWithOperand(OP_GET_LOCAL, 0, expr.Token.Line)

		// 2. 加载参数
		for _, arg := range expr.Arguments {
			if err := c.compileExpression(arg.Value); err != nil {
				return err
			}
		}

		// 3. 加载 super class
		nameIndex := c.addConstant(&interpreter.String{Value: "super"})
		c.emitWithOperand(OP_GET_GLOBAL, byte(nameIndex), expr.Token.Line)

		// 4. 发出 SUPER_INVOKE 指令
		methodNameIndex := c.addConstant(&interpreter.String{Value: expr.Method.Value})
		c.emitWithOperand(OP_SUPER_INVOKE, byte(methodNameIndex), expr.Token.Line)
		c.bytecode.Instructions = append(c.bytecode.Instructions, byte(len(expr.Arguments)))
		c.bytecode.Lines = append(c.bytecode.Lines, expr.Token.Line)
		return nil
	}

	// 编译类名
	if err := c.compileExpression(expr.ClassName); err != nil {
		return err
	}

	// 编译参数
	for _, arg := range expr.Arguments {
		if err := c.compileExpression(arg.Value); err != nil {
			return err
		}
	}

	// 方法名常量
	nameIndex := c.addConstant(&interpreter.String{Value: expr.Method.Value})

	// 发出 INVOKE_STATIC 指令
	c.emitWithOperand(OP_INVOKE_STATIC, byte(nameIndex), expr.Token.Line)
	c.bytecode.Instructions = append(c.bytecode.Instructions, byte(len(expr.Arguments)))
	c.bytecode.Lines = append(c.bytecode.Lines, expr.Token.Line)

	return nil
}

// compileStaticAccessExpression 编译静态成员访问
func (c *Compiler) compileStaticAccessExpression(expr *parser.StaticAccessExpression) error {
	// 编译类名
	if err := c.compileExpression(expr.ClassName); err != nil {
		return err
	}

	// 成员名常量
	nameIndex := c.addConstant(&interpreter.String{Value: expr.Name.Value})

	// 发出 GET_STATIC_FIELD 指令
	c.emitWithOperand(OP_GET_STATIC_FIELD, byte(nameIndex), expr.Token.Line)

	return nil
}

// compileClassLiteralExpression 编译类名字面量表达式
// ClassName::class 返回类名字符串
func (c *Compiler) compileClassLiteralExpression(expr *parser.ClassLiteralExpression) error {
	// 将类名作为字符串常量加载
	index := c.addConstant(&interpreter.String{Value: expr.ClassName.Value})
	c.emitWithOperand(OP_CONST, byte(index), expr.Token.Line)
	return nil
}

// compileInterpolatedString 编译插值字符串
func (c *Compiler) compileInterpolatedString(expr *parser.InterpolatedStringLiteral) error {
	if len(expr.Parts) == 0 {
		c.emit(OP_CONST, expr.Token.Line)
		index := c.addConstant(&interpreter.String{Value: ""})
		c.bytecode.Instructions = append(c.bytecode.Instructions, byte(index))
		return nil
	}

	// 编译第一部分
	first := expr.Parts[0]
	if first.IsExpr {
		if err := c.compileExpression(first.Expr); err != nil {
			return err
		}
	} else {
		index := c.addConstant(&interpreter.String{Value: first.Text})
		c.emitWithOperand(OP_CONST, byte(index), expr.Token.Line)
	}

	// 编译后续部分，每个部分与前一个拼接
	for i := 1; i < len(expr.Parts); i++ {
		part := expr.Parts[i]
		if part.IsExpr {
			if err := c.compileExpression(part.Expr); err != nil {
				return err
			}
		} else {
			index := c.addConstant(&interpreter.String{Value: part.Text})
			c.emitWithOperand(OP_CONST, byte(index), expr.Token.Line)
		}
		// 拼接
		c.emit(OP_ADD, expr.Token.Line)
	}

	return nil
}

// compileSliceExpression 编译切片表达式
func (c *Compiler) compileSliceExpression(expr *parser.SliceExpression) error {
	// 编译对象
	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}

	// 编译起始索引
	if expr.Start != nil {
		if err := c.compileExpression(expr.Start); err != nil {
			return err
		}
	} else {
		c.emit(OP_NULL, expr.Token.Line)
	}

	// 编译结束索引
	if expr.End != nil {
		if err := c.compileExpression(expr.End); err != nil {
			return err
		}
	} else {
		c.emit(OP_NULL, expr.Token.Line)
	}

	// 发出 SLICE 指令
	c.emit(OP_SLICE, expr.Token.Line)

	return nil
}

// compileTypeAssertionExpression 编译类型断言表达式
// 对应语法：value as Type（强制断言）或 value as? Type（安全断言）
func (c *Compiler) compileTypeAssertionExpression(expr *parser.TypeAssertionExpression) error {
	// 1. 编译左侧表达式（被断言的值）
	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}

	// 2. 获取目标类型名称字符串
	targetTypeName := c.getTypeNameFromExpr(expr.TargetType)
	if targetTypeName == "" {
		return fmt.Errorf("无效的类型断言目标类型")
	}

	// 3. 将类型名称添加到常量池
	typeNameIndex := c.addConstant(&interpreter.String{Value: targetTypeName})

	// 4. 发出 OP_TYPE_ASSERT 指令
	// 操作数：类型名索引 + 是否安全断言标志
	c.emitWithOperand(OP_TYPE_ASSERT, byte(typeNameIndex), expr.Token.Line)
	if expr.IsSafe {
		c.bytecode.Instructions = append(c.bytecode.Instructions, 1)
	} else {
		c.bytecode.Instructions = append(c.bytecode.Instructions, 0)
	}
	c.bytecode.Lines = append(c.bytecode.Lines, expr.Token.Line)

	return nil
}

// getTypeNameFromExpr 从类型表达式中获取类型名称
func (c *Compiler) getTypeNameFromExpr(typeExpr parser.Expression) string {
	switch t := typeExpr.(type) {
	case *parser.Identifier:
		return t.Value
	case *parser.ArrayType:
		elemType := c.getTypeNameFromExpr(t.ElementType)
		if elemType == "" {
			return ""
		}
		return "[]" + elemType
	case *parser.MapType:
		keyType := ""
		if t.KeyType != nil {
			keyType = t.KeyType.Value
		}
		valueType := c.getTypeNameFromExpr(t.ValueType)
		if valueType == "" {
			return ""
		}
		return "map[" + keyType + "]" + valueType
	default:
		return ""
	}
}

// emitBinaryOp 发出二元运算指令
func (c *Compiler) emitBinaryOp(op string, line int) {
	switch op {
	case "+":
		c.emit(OP_ADD, line)
	case "-":
		c.emit(OP_SUB, line)
	case "*":
		c.emit(OP_MUL, line)
	case "/":
		c.emit(OP_DIV, line)
	case "%":
		c.emit(OP_MOD, line)
	case "&":
		c.emit(OP_BIT_AND, line)
	case "|":
		c.emit(OP_BIT_OR, line)
	case "^":
		c.emit(OP_BIT_XOR, line)
	case "<<":
		c.emit(OP_LSHIFT, line)
	case ">>":
		c.emit(OP_RSHIFT, line)
	}
}
