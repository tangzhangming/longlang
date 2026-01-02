package interpreter

import (
	"regexp"
	"strings"
)

// 正则表达式选项常量（与 LongLang 的 RegexOptions 对应）
const (
	REGEX_NONE        = 0
	REGEX_IGNORE_CASE = 1
	REGEX_MULTILINE   = 2
	REGEX_SINGLELINE  = 4
	REGEX_UNICODE     = 8
	REGEX_UNGREEDY    = 16
)

// registerRegexBuiltins 注册正则表达式相关的内置函数
func registerRegexBuiltins(env *Environment) {
	// __regex_compile - 编译正则表达式（验证是否有效）
	env.Set("__regex_compile", &Builtin{Fn: func(args ...Object) Object {
		if len(args) < 1 || len(args) > 2 {
			return newError("__regex_compile 需要1-2个参数")
		}
		pattern, ok := args[0].(*String)
		if !ok {
			return newError("__regex_compile 第一个参数必须是字符串")
		}

		options := int64(0)
		if len(args) > 1 {
			if opt, ok := args[1].(*Integer); ok {
				options = opt.Value
			}
		}

		// 构建正则表达式模式
		finalPattern := buildRegexPattern(pattern.Value, options)
		_, err := regexp.Compile(finalPattern)
		if err != nil {
			return &Null{}
		}
		return &Boolean{Value: true}
	}})

	// __regex_is_match - 测试是否匹配
	env.Set("__regex_is_match", &Builtin{Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 3 {
			return newError("__regex_is_match 需要2-3个参数")
		}
		input, ok := args[0].(*String)
		if !ok {
			return newError("__regex_is_match 第一个参数必须是字符串")
		}
		pattern, ok := args[1].(*String)
		if !ok {
			return newError("__regex_is_match 第二个参数必须是字符串")
		}

		options := int64(0)
		if len(args) > 2 {
			if opt, ok := args[2].(*Integer); ok {
				options = opt.Value
			}
		}

		finalPattern := buildRegexPattern(pattern.Value, options)
		re, err := regexp.Compile(finalPattern)
		if err != nil {
			return &Boolean{Value: false}
		}
		return &Boolean{Value: re.MatchString(input.Value)}
	}})

	// __regex_match - 查找第一个匹配，返回 map
	env.Set("__regex_match", &Builtin{Fn: regexMatchBuiltin})

	// __regex_match_all - 查找所有匹配，返回 map 数组
	env.Set("__regex_match_all", &Builtin{Fn: regexMatchAllBuiltin})

	// __regex_replace - 替换
	env.Set("__regex_replace", &Builtin{Fn: func(args ...Object) Object {
		if len(args) < 3 || len(args) > 4 {
			return newError("__regex_replace 需要3-4个参数")
		}
		input, ok := args[0].(*String)
		if !ok {
			return newError("__regex_replace 第一个参数必须是字符串")
		}
		pattern, ok := args[1].(*String)
		if !ok {
			return newError("__regex_replace 第二个参数必须是字符串")
		}
		replacement, ok := args[2].(*String)
		if !ok {
			return newError("__regex_replace 第三个参数必须是字符串")
		}

		options := int64(0)
		if len(args) > 3 {
			if opt, ok := args[3].(*Integer); ok {
				options = opt.Value
			}
		}

		finalPattern := buildRegexPattern(pattern.Value, options)
		re, err := regexp.Compile(finalPattern)
		if err != nil {
			return &String{Value: input.Value}
		}

		// 将 $1, $2 等转换为 Go 的 ${1}, ${2}
		goReplacement := convertReplacement(replacement.Value)
		result := re.ReplaceAllString(input.Value, goReplacement)
		return &String{Value: result}
	}})

	// __regex_replace_callback - 使用回调函数替换
	env.Set("__regex_replace_callback", &Builtin{Fn: func(args ...Object) Object {
		// 这个函数需要特殊处理，因为需要调用 LongLang 函数
		// 暂时返回错误，后续通过 interpreter 实现
		return newError("__regex_replace_callback 需要通过 Interpreter 实现")
	}})

	// __regex_split - 分割字符串
	env.Set("__regex_split", &Builtin{Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 4 {
			return newError("__regex_split 需要2-4个参数")
		}
		input, ok := args[0].(*String)
		if !ok {
			return newError("__regex_split 第一个参数必须是字符串")
		}
		pattern, ok := args[1].(*String)
		if !ok {
			return newError("__regex_split 第二个参数必须是字符串")
		}

		limit := -1
		if len(args) > 2 {
			if l, ok := args[2].(*Integer); ok {
				limit = int(l.Value)
			}
		}

		options := int64(0)
		if len(args) > 3 {
			if opt, ok := args[3].(*Integer); ok {
				options = opt.Value
			}
		}

		finalPattern := buildRegexPattern(pattern.Value, options)
		re, err := regexp.Compile(finalPattern)
		if err != nil {
			return &Array{Elements: []Object{&String{Value: input.Value}}}
		}

		var parts []string
		if limit < 0 {
			parts = re.Split(input.Value, -1)
		} else {
			parts = re.Split(input.Value, limit)
		}

		elements := make([]Object, len(parts))
		for i, part := range parts {
			elements[i] = &String{Value: part}
		}
		return &Array{Elements: elements, ElementType: "string"}
	}})

	// __regex_escape - 转义正则特殊字符
	env.Set("__regex_escape", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__regex_escape 需要1个参数")
		}
		str, ok := args[0].(*String)
		if !ok {
			return newError("__regex_escape 参数必须是字符串")
		}
		return &String{Value: regexp.QuoteMeta(str.Value)}
	}})

	// __regex_find_index - 查找匹配位置
	env.Set("__regex_find_index", &Builtin{Fn: func(args ...Object) Object {
		if len(args) < 2 || len(args) > 3 {
			return newError("__regex_find_index 需要2-3个参数")
		}
		input, ok := args[0].(*String)
		if !ok {
			return newError("__regex_find_index 第一个参数必须是字符串")
		}
		pattern, ok := args[1].(*String)
		if !ok {
			return newError("__regex_find_index 第二个参数必须是字符串")
		}

		options := int64(0)
		if len(args) > 2 {
			if opt, ok := args[2].(*Integer); ok {
				options = opt.Value
			}
		}

		finalPattern := buildRegexPattern(pattern.Value, options)
		re, err := regexp.Compile(finalPattern)
		if err != nil {
			return &Integer{Value: -1}
		}

		loc := re.FindStringIndex(input.Value)
		if loc == nil {
			return &Integer{Value: -1}
		}
		return &Integer{Value: int64(loc[0])}
	}})
}

// buildRegexPattern 根据选项构建最终的正则表达式模式
func buildRegexPattern(pattern string, options int64) string {
	var flags strings.Builder

	if options&REGEX_IGNORE_CASE != 0 {
		flags.WriteString("i")
	}
	if options&REGEX_MULTILINE != 0 {
		flags.WriteString("m")
	}
	if options&REGEX_SINGLELINE != 0 {
		flags.WriteString("s")
	}
	if options&REGEX_UNGREEDY != 0 {
		flags.WriteString("U")
	}

	if flags.Len() > 0 {
		return "(?" + flags.String() + ")" + pattern
	}
	return pattern
}

// convertReplacement 将 $1, $2 等转换为 Go 的 ${1}, ${2}
func convertReplacement(replacement string) string {
	// 简单替换 $n 为 ${n}
	result := replacement
	for i := 9; i >= 1; i-- {
		old := "$" + string(rune('0'+i))
		new := "${" + string(rune('0'+i)) + "}"
		result = strings.ReplaceAll(result, old, new)
	}
	return result
}

// regexMatchBuiltin 实现 __regex_match
func regexMatchBuiltin(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("__regex_match 需要2-3个参数")
	}
	input, ok := args[0].(*String)
	if !ok {
		return newError("__regex_match 第一个参数必须是字符串")
	}
	pattern, ok := args[1].(*String)
	if !ok {
		return newError("__regex_match 第二个参数必须是字符串")
	}

	options := int64(0)
	if len(args) > 2 {
		if opt, ok := args[2].(*Integer); ok {
			options = opt.Value
		}
	}

	finalPattern := buildRegexPattern(pattern.Value, options)
	re, err := regexp.Compile(finalPattern)
	if err != nil {
		// 返回失败的 Match
		return createMatchMap(false, "", -1, 0, nil, nil)
	}

	// 查找匹配
	match := re.FindStringSubmatchIndex(input.Value)
	if match == nil {
		return createMatchMap(false, "", -1, 0, nil, nil)
	}

	// 提取匹配信息
	fullMatch := input.Value[match[0]:match[1]]
	groups := [][]int{}
	for i := 0; i < len(match); i += 2 {
		if match[i] >= 0 && match[i+1] >= 0 {
			groups = append(groups, []int{match[i], match[i+1]})
		} else {
			groups = append(groups, []int{-1, -1})
		}
	}

	// 获取命名捕获组
	namedGroups := make(map[string]string)
	subexpNames := re.SubexpNames()
	for i, name := range subexpNames {
		if name != "" && i < len(groups) && groups[i][0] >= 0 {
			namedGroups[name] = input.Value[groups[i][0]:groups[i][1]]
		}
	}

	return createMatchMap(true, fullMatch, match[0], len(fullMatch), groups, namedGroups)
}

// regexMatchAllBuiltin 实现 __regex_match_all
func regexMatchAllBuiltin(args ...Object) Object {
	if len(args) < 2 || len(args) > 3 {
		return newError("__regex_match_all 需要2-3个参数")
	}
	input, ok := args[0].(*String)
	if !ok {
		return newError("__regex_match_all 第一个参数必须是字符串")
	}
	pattern, ok := args[1].(*String)
	if !ok {
		return newError("__regex_match_all 第二个参数必须是字符串")
	}

	options := int64(0)
	if len(args) > 2 {
		if opt, ok := args[2].(*Integer); ok {
			options = opt.Value
		}
	}

	finalPattern := buildRegexPattern(pattern.Value, options)
	re, err := regexp.Compile(finalPattern)
	if err != nil {
		return &Array{Elements: []Object{}}
	}

	// 查找所有匹配
	allMatches := re.FindAllStringSubmatchIndex(input.Value, -1)
	if allMatches == nil {
		return &Array{Elements: []Object{}}
	}

	// 获取命名捕获组名称
	subexpNames := re.SubexpNames()

	elements := []Object{}
	for _, match := range allMatches {
		fullMatch := input.Value[match[0]:match[1]]
		groups := [][]int{}
		for i := 0; i < len(match); i += 2 {
			if match[i] >= 0 && match[i+1] >= 0 {
				groups = append(groups, []int{match[i], match[i+1]})
			} else {
				groups = append(groups, []int{-1, -1})
			}
		}

		// 获取命名捕获组
		namedGroups := make(map[string]string)
		for i, name := range subexpNames {
			if name != "" && i < len(groups) && groups[i][0] >= 0 {
				namedGroups[name] = input.Value[groups[i][0]:groups[i][1]]
			}
		}

		elements = append(elements, createMatchMap(true, fullMatch, match[0], len(fullMatch), groups, namedGroups))
	}

	return &Array{Elements: elements}
}

// newRegexMap 创建新的 Map 对象
func newRegexMap() *Map {
	return &Map{
		Pairs:     make(map[string]Object),
		Keys:      []string{},
		KeyType:   "string",
		ValueType: "any",
	}
}

// createMatchMap 创建匹配结果的 Map
func createMatchMap(success bool, value string, index int, length int, groups [][]int, namedGroups map[string]string) *Map {
	m := newRegexMap()
	m.Set("success", &Boolean{Value: success})
	m.Set("value", &String{Value: value})
	m.Set("index", &Integer{Value: int64(index)})
	m.Set("length", &Integer{Value: int64(length)})

	// 转换 groups 为数组
	groupArray := []Object{}
	if groups != nil {
		for _, g := range groups {
			gMap := newRegexMap()
			gMap.Set("start", &Integer{Value: int64(g[0])})
			gMap.Set("end", &Integer{Value: int64(g[1])})
			groupArray = append(groupArray, gMap)
		}
	}
	m.Set("groups", &Array{Elements: groupArray})

	// 转换 namedGroups 为 Map
	ngMap := newRegexMap()
	if namedGroups != nil {
		for k, v := range namedGroups {
			ngMap.Set(k, &String{Value: v})
		}
	}
	m.Set("namedGroups", ngMap)

	return m
}

