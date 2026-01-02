package interpreter

import (
	"github.com/tangzhangming/longlang/internal/parser"
)

// ========== Switch 语句执行 ==========

// evalSwitchStatement 执行 switch 语句
func (i *Interpreter) evalSwitchStatement(node *parser.SwitchStatement) Object {
	// 执行初始化语句（如果有）
	if node.Init != nil {
		result := i.Eval(node.Init)
		if isError(result) {
			return result
		}
	}

	// 计算要匹配的值（如果有）
	var switchValue Object
	isConditionSwitch := node.Value == nil
	if !isConditionSwitch {
		switchValue = i.Eval(node.Value)
		if isError(switchValue) {
			return switchValue
		}
	}

	// 遍历 case 分支
	for _, caseClause := range node.Cases {
		matched := false

		if caseClause.IsCondition {
			// 条件 switch: case condition:
			condResult := i.Eval(caseClause.Condition)
			if isError(condResult) {
				return condResult
			}
			matched = isTruthy(condResult)
		} else {
			// 值 switch: case value1, value2, ...:
			for _, caseValue := range caseClause.Values {
				valueResult := i.Eval(caseValue)
				if isError(valueResult) {
					return valueResult
				}

				// 比较值是否相等
				if isEqual(switchValue, valueResult) {
					matched = true
					break
				}
			}
		}

		if matched {
			// 执行 case 体
			return i.evalBlockStatement(caseClause.Body)
		}
	}

	// 执行 default 分支（如果有）
	if node.Default != nil {
		return i.evalBlockStatement(node.Default)
	}

	return &Null{}
}

// ========== Match 表达式执行 ==========

// evalMatchExpression 执行 match 表达式
func (i *Interpreter) evalMatchExpression(node *parser.MatchExpression) Object {
	// 计算要匹配的值
	matchValue := i.Eval(node.Value)
	if isError(matchValue) {
		return matchValue
	}

	// 遍历匹配分支
	for _, arm := range node.Arms {
		matched := false

		if arm.IsWildcard {
			// 通配符 _ 总是匹配
			matched = true
		} else if arm.Binding != nil {
			// 带守卫的绑定变量: identifier if guard => ...
			// 创建新环境，绑定变量
			newEnv := NewEnclosedEnvironment(i.env)
			newEnv.Set(arm.Binding.Value, matchValue)

			// 在新环境中执行守卫条件
			oldEnv := i.env
			i.env = newEnv
			guardResult := i.Eval(arm.Guard)
			i.env = oldEnv

			if isError(guardResult) {
				return guardResult
			}
			matched = isTruthy(guardResult)

			if matched {
				// 在新环境中执行结果
				i.env = newEnv
				defer func() { i.env = oldEnv }()
				return i.evalMatchArmResult(arm)
			}
		} else {
			// 普通模式匹配: pattern1, pattern2, ... => ...
			for _, pattern := range arm.Patterns {
				patternValue := i.Eval(pattern)
				if isError(patternValue) {
					return patternValue
				}

				if isEqual(matchValue, patternValue) {
					matched = true
					break
				}
			}
		}

		if matched {
			return i.evalMatchArmResult(arm)
		}
	}

	// 没有匹配到任何分支，这是运行时错误
	return newError("match 表达式未匹配到任何分支，值: %s", matchValue.Inspect())
}

// evalMatchArmResult 执行 match 分支的结果
func (i *Interpreter) evalMatchArmResult(arm *parser.MatchArm) Object {
	if arm.Body != nil {
		// 代码块形式
		result := i.evalBlockStatement(arm.Body)
		return unwrapReturnValue(result)
	} else if arm.Result != nil {
		// 表达式形式
		return i.Eval(arm.Result)
	}
	return &Null{}
}

// ========== 辅助函数 ==========

// isEqual 检查两个对象是否相等
func isEqual(a, b Object) bool {
	if a == nil || b == nil {
		return a == b
	}

	// 类型不同，不相等（除非有特殊处理）
	if a.Type() != b.Type() {
		// 特殊情况：Integer 和 Float 可以比较
		if (a.Type() == INTEGER_OBJ && b.Type() == FLOAT_OBJ) ||
			(a.Type() == FLOAT_OBJ && b.Type() == INTEGER_OBJ) {
			aVal := getFloatValue(a)
			bVal := getFloatValue(b)
			return aVal == bVal
		}
		return false
	}

	switch av := a.(type) {
	case *Integer:
		return av.Value == b.(*Integer).Value
	case *Float:
		return av.Value == b.(*Float).Value
	case *String:
		return av.Value == b.(*String).Value
	case *Boolean:
		return av.Value == b.(*Boolean).Value
	case *Null:
		return true // null == null
	case *EnumValue:
		bv := b.(*EnumValue)
		return av.Enum.Name == bv.Enum.Name && av.Name == bv.Name
	default:
		// 对于其他类型（如对象引用），比较指针
		return a == b
	}
}

// getFloatValue 获取浮点值
func getFloatValue(obj Object) float64 {
	switch v := obj.(type) {
	case *Integer:
		return float64(v.Value)
	case *Float:
		return v.Value
	default:
		return 0
	}
}

