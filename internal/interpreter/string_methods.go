package interpreter

import (
	"strings"
	"unicode"
)

// StringMethod 字符串方法类型
type StringMethod func(s *String, args ...Object) Object

// stringMethods 存储所有字符串方法
var stringMethods = map[string]StringMethod{
	// ========== 基本信息 ==========
	"length":   stringLength,
	"isEmpty":  stringIsEmpty,
	"charAt":   stringCharAt,

	// ========== 查找 ==========
	"indexOf":     stringIndexOf,
	"lastIndexOf": stringLastIndexOf,
	"contains":    stringContains,
	"startsWith":  stringStartsWith,
	"endsWith":    stringEndsWith,

	// ========== 比较 ==========
	"equals":           stringEquals,
	"equalsIgnoreCase": stringEqualsIgnoreCase,

	// ========== 连接和子串 ==========
	"concat":    stringConcat,
	"substring": stringSubstring,
	"repeat":    stringRepeat,

	// ========== 去除空白和字符 ==========
	"trim":  stringTrim,
	"ltrim": stringLtrim,
	"rtrim": stringRtrim,

	// ========== 大小写转换 ==========
	"upper":   stringUpper,
	"lower":   stringLower,
	"ucfirst": stringUcfirst,
	"title":   stringTitle,

	// ========== 格式转换 ==========
	"camel":  stringCamel,
	"studly": stringStudly,
	"snake":  stringSnake,
	"kebab":  stringKebab,

	// ========== 替换 ==========
	"replace":    stringReplace,
	"replaceAll": stringReplaceAll,

	// ========== 填充 ==========
	"padLeft":  stringPadLeft,
	"padRight": stringPadRight,

	// ========== 其他 ==========
	"reverse": stringReverse,
}

// GetStringMethod 获取字符串方法
func GetStringMethod(name string) (StringMethod, bool) {
	method, ok := stringMethods[name]
	return method, ok
}

// ========== 基本信息方法 ==========

// stringLength 获取字符串长度
// "hello".length() => 5
func stringLength(s *String, args ...Object) Object {
	return &Integer{Value: int64(len([]rune(s.Value)))}
}

// stringIsEmpty 判断是否为空字符串
// "".isEmpty() => true
func stringIsEmpty(s *String, args ...Object) Object {
	return &Boolean{Value: len(s.Value) == 0}
}

// stringCharAt 获取指定位置的字符
// "hello".charAt(0) => "h"
func stringCharAt(s *String, args ...Object) Object {
	if len(args) != 1 {
		return NewError("charAt 需要1个参数")
	}
	index, ok := args[0].(*Integer)
	if !ok {
		return NewError("charAt 参数必须是整数")
	}
	runes := []rune(s.Value)
	if index.Value < 0 || index.Value >= int64(len(runes)) {
		return NewError("索引越界: %d", index.Value)
	}
	return &String{Value: string(runes[index.Value])}
}

// ========== 查找方法 ==========

// stringIndexOf 返回子串第一次出现的索引
// "hello".indexOf("l") => 2
func stringIndexOf(s *String, args ...Object) Object {
	if len(args) != 1 {
		return NewError("indexOf 需要1个参数")
	}
	substr, ok := args[0].(*String)
	if !ok {
		return NewError("indexOf 参数必须是字符串")
	}
	index := strings.Index(s.Value, substr.Value)
	return &Integer{Value: int64(index)}
}

// stringLastIndexOf 返回子串最后一次出现的索引
// "hello".lastIndexOf("l") => 3
func stringLastIndexOf(s *String, args ...Object) Object {
	if len(args) != 1 {
		return NewError("lastIndexOf 需要1个参数")
	}
	substr, ok := args[0].(*String)
	if !ok {
		return NewError("lastIndexOf 参数必须是字符串")
	}
	index := strings.LastIndex(s.Value, substr.Value)
	return &Integer{Value: int64(index)}
}

// stringContains 判断是否包含子串
// "hello".contains("ell") => true
func stringContains(s *String, args ...Object) Object {
	if len(args) != 1 {
		return NewError("contains 需要1个参数")
	}
	substr, ok := args[0].(*String)
	if !ok {
		return NewError("contains 参数必须是字符串")
	}
	return &Boolean{Value: strings.Contains(s.Value, substr.Value)}
}

// stringStartsWith 判断是否以指定前缀开始
// "hello".startsWith("he") => true
func stringStartsWith(s *String, args ...Object) Object {
	if len(args) != 1 {
		return NewError("startsWith 需要1个参数")
	}
	prefix, ok := args[0].(*String)
	if !ok {
		return NewError("startsWith 参数必须是字符串")
	}
	return &Boolean{Value: strings.HasPrefix(s.Value, prefix.Value)}
}

// stringEndsWith 判断是否以指定后缀结束
// "hello".endsWith("lo") => true
func stringEndsWith(s *String, args ...Object) Object {
	if len(args) != 1 {
		return NewError("endsWith 需要1个参数")
	}
	suffix, ok := args[0].(*String)
	if !ok {
		return NewError("endsWith 参数必须是字符串")
	}
	return &Boolean{Value: strings.HasSuffix(s.Value, suffix.Value)}
}

// ========== 比较方法 ==========

// stringEquals 比较两个字符串是否相等
// "hello".equals("hello") => true
func stringEquals(s *String, args ...Object) Object {
	if len(args) != 1 {
		return NewError("equals 需要1个参数")
	}
	other, ok := args[0].(*String)
	if !ok {
		return &Boolean{Value: false}
	}
	return &Boolean{Value: s.Value == other.Value}
}

// stringEqualsIgnoreCase 忽略大小写比较
// "Hello".equalsIgnoreCase("hello") => true
func stringEqualsIgnoreCase(s *String, args ...Object) Object {
	if len(args) != 1 {
		return NewError("equalsIgnoreCase 需要1个参数")
	}
	other, ok := args[0].(*String)
	if !ok {
		return &Boolean{Value: false}
	}
	return &Boolean{Value: strings.EqualFold(s.Value, other.Value)}
}

// ========== 连接和子串方法 ==========

// stringConcat 连接字符串
// "hello".concat(" world") => "hello world"
func stringConcat(s *String, args ...Object) Object {
	if len(args) != 1 {
		return NewError("concat 需要1个参数")
	}
	other, ok := args[0].(*String)
	if !ok {
		return NewError("concat 参数必须是字符串")
	}
	return &String{Value: s.Value + other.Value}
}

// stringSubstring 获取子串
// "hello".substring(0, 2) => "he"
// "hello".substring(2) => "llo"
func stringSubstring(s *String, args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return NewError("substring 需要1-2个参数")
	}

	start, ok := args[0].(*Integer)
	if !ok {
		return NewError("substring 起始参数必须是整数")
	}

	runes := []rune(s.Value)
	length := int64(len(runes))

	startIdx := start.Value
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx > length {
		startIdx = length
	}

	var endIdx int64 = length
	if len(args) == 2 {
		end, ok := args[1].(*Integer)
		if !ok {
			return NewError("substring 结束参数必须是整数")
		}
		endIdx = end.Value
		if endIdx < startIdx {
			endIdx = startIdx
		}
		if endIdx > length {
			endIdx = length
		}
	}

	return &String{Value: string(runes[startIdx:endIdx])}
}

// stringRepeat 重复字符串
// "ab".repeat(3) => "ababab"
func stringRepeat(s *String, args ...Object) Object {
	if len(args) != 1 {
		return NewError("repeat 需要1个参数")
	}
	count, ok := args[0].(*Integer)
	if !ok {
		return NewError("repeat 参数必须是整数")
	}
	if count.Value < 0 {
		return NewError("repeat 参数不能为负数")
	}
	return &String{Value: strings.Repeat(s.Value, int(count.Value))}
}

// ========== 去除空白和字符方法 ==========

// stringTrim 去除首尾空白或指定字符串
// "  hello  ".trim() => "hello"
// "xxxhelloxxx".trim("xxx") => "hello"
func stringTrim(s *String, args ...Object) Object {
	if len(args) == 0 {
		return &String{Value: strings.TrimSpace(s.Value)}
	}
	if len(args) == 1 {
		cutset, ok := args[0].(*String)
		if !ok {
			return NewError("trim 参数必须是字符串")
		}
		result := s.Value
		// 去除前缀
		for strings.HasPrefix(result, cutset.Value) {
			result = strings.TrimPrefix(result, cutset.Value)
		}
		// 去除后缀
		for strings.HasSuffix(result, cutset.Value) {
			result = strings.TrimSuffix(result, cutset.Value)
		}
		return &String{Value: result}
	}
	return NewError("trim 最多接受1个参数")
}

// stringLtrim 去除左边空白或指定字符串
// "  hello".ltrim() => "hello"
// "http://example.com".ltrim("http://") => "example.com"
func stringLtrim(s *String, args ...Object) Object {
	if len(args) == 0 {
		return &String{Value: strings.TrimLeftFunc(s.Value, unicode.IsSpace)}
	}
	if len(args) == 1 {
		cutset, ok := args[0].(*String)
		if !ok {
			return NewError("ltrim 参数必须是字符串")
		}
		result := s.Value
		for strings.HasPrefix(result, cutset.Value) {
			result = strings.TrimPrefix(result, cutset.Value)
		}
		return &String{Value: result}
	}
	return NewError("ltrim 最多接受1个参数")
}

// stringRtrim 去除右边空白或指定字符串
// "hello  ".rtrim() => "hello"
// "example.com/".rtrim("/") => "example.com"
func stringRtrim(s *String, args ...Object) Object {
	if len(args) == 0 {
		return &String{Value: strings.TrimRightFunc(s.Value, unicode.IsSpace)}
	}
	if len(args) == 1 {
		cutset, ok := args[0].(*String)
		if !ok {
			return NewError("rtrim 参数必须是字符串")
		}
		result := s.Value
		for strings.HasSuffix(result, cutset.Value) {
			result = strings.TrimSuffix(result, cutset.Value)
		}
		return &String{Value: result}
	}
	return NewError("rtrim 最多接受1个参数")
}

// ========== 大小写转换方法 ==========

// stringUpper 转换为大写
// "hello".upper() => "HELLO"
func stringUpper(s *String, args ...Object) Object {
	return &String{Value: strings.ToUpper(s.Value)}
}

// stringLower 转换为小写
// "HELLO".lower() => "hello"
func stringLower(s *String, args ...Object) Object {
	return &String{Value: strings.ToLower(s.Value)}
}

// stringUcfirst 首字母大写
// "hello".ucfirst() => "Hello"
func stringUcfirst(s *String, args ...Object) Object {
	if len(s.Value) == 0 {
		return &String{Value: ""}
	}
	runes := []rune(s.Value)
	runes[0] = unicode.ToUpper(runes[0])
	return &String{Value: string(runes)}
}

// stringTitle 每个单词首字母大写
// "hello world".title() => "Hello World"
func stringTitle(s *String, args ...Object) Object {
	return &String{Value: strings.Title(s.Value)}
}

// ========== 格式转换方法 ==========

// stringCamel 转换为小驼峰（camelCase）
// "foo_bar".camel() => "fooBar"
// "foo-bar".camel() => "fooBar"
func stringCamel(s *String, args ...Object) Object {
	result := toCamelCase(s.Value, false)
	return &String{Value: result}
}

// stringStudly 转换为大驼峰（PascalCase）
// "foo_bar".studly() => "FooBar"
// "foo-bar".studly() => "FooBar"
func stringStudly(s *String, args ...Object) Object {
	result := toCamelCase(s.Value, true)
	return &String{Value: result}
}

// stringSnake 转换为蛇形（snake_case）
// "fooBar".snake() => "foo_bar"
// "fooBar".snake("-") => "foo-bar"
func stringSnake(s *String, args ...Object) Object {
	delimiter := "_"
	if len(args) == 1 {
		if d, ok := args[0].(*String); ok {
			delimiter = d.Value
		} else {
			return NewError("snake 参数必须是字符串")
		}
	}
	result := toSnakeCase(s.Value, delimiter)
	return &String{Value: result}
}

// stringKebab 转换为烤串式（kebab-case）
// "fooBar".kebab() => "foo-bar"
func stringKebab(s *String, args ...Object) Object {
	result := toSnakeCase(s.Value, "-")
	return &String{Value: result}
}

// ========== 替换方法 ==========

// stringReplace 替换第一个匹配项
// "hello".replace("l", "L") => "heLlo"
func stringReplace(s *String, args ...Object) Object {
	if len(args) != 2 {
		return NewError("replace 需要2个参数")
	}
	old, ok1 := args[0].(*String)
	new, ok2 := args[1].(*String)
	if !ok1 || !ok2 {
		return NewError("replace 参数必须是字符串")
	}
	return &String{Value: strings.Replace(s.Value, old.Value, new.Value, 1)}
}

// stringReplaceAll 替换所有匹配项
// "hello".replaceAll("l", "L") => "heLLo"
func stringReplaceAll(s *String, args ...Object) Object {
	if len(args) != 2 {
		return NewError("replaceAll 需要2个参数")
	}
	old, ok1 := args[0].(*String)
	new, ok2 := args[1].(*String)
	if !ok1 || !ok2 {
		return NewError("replaceAll 参数必须是字符串")
	}
	return &String{Value: strings.ReplaceAll(s.Value, old.Value, new.Value)}
}

// ========== 填充方法 ==========

// stringPadLeft 左填充
// "5".padLeft(3, "0") => "005"
func stringPadLeft(s *String, args ...Object) Object {
	if len(args) != 2 {
		return NewError("padLeft 需要2个参数")
	}
	length, ok1 := args[0].(*Integer)
	pad, ok2 := args[1].(*String)
	if !ok1 || !ok2 {
		return NewError("padLeft 参数类型错误")
	}

	runes := []rune(s.Value)
	targetLen := int(length.Value)
	if len(runes) >= targetLen || len(pad.Value) == 0 {
		return &String{Value: s.Value}
	}

	padRunes := []rune(pad.Value)
	result := make([]rune, 0, targetLen)
	for len(result)+len(runes) < targetLen {
		result = append(result, padRunes...)
	}
	// 截取到正确长度
	needLen := targetLen - len(runes)
	if len(result) > needLen {
		result = result[:needLen]
	}
	result = append(result, runes...)
	return &String{Value: string(result)}
}

// stringPadRight 右填充
// "5".padRight(3, "0") => "500"
func stringPadRight(s *String, args ...Object) Object {
	if len(args) != 2 {
		return NewError("padRight 需要2个参数")
	}
	length, ok1 := args[0].(*Integer)
	pad, ok2 := args[1].(*String)
	if !ok1 || !ok2 {
		return NewError("padRight 参数类型错误")
	}

	runes := []rune(s.Value)
	targetLen := int(length.Value)
	if len(runes) >= targetLen || len(pad.Value) == 0 {
		return &String{Value: s.Value}
	}

	padRunes := []rune(pad.Value)
	result := make([]rune, len(runes), targetLen)
	copy(result, runes)
	for len(result) < targetLen {
		result = append(result, padRunes...)
	}
	// 截取到正确长度
	if len(result) > targetLen {
		result = result[:targetLen]
	}
	return &String{Value: string(result)}
}

// ========== 其他方法 ==========

// stringReverse 反转字符串
// "hello".reverse() => "olleh"
func stringReverse(s *String, args ...Object) Object {
	runes := []rune(s.Value)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return &String{Value: string(runes)}
}

// ========== 辅助函数 ==========

// toCamelCase 将字符串转换为驼峰格式
func toCamelCase(s string, upperFirst bool) string {
	var result strings.Builder
	capitalize := upperFirst

	for i, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			capitalize = true
			continue
		}
		if capitalize {
			result.WriteRune(unicode.ToUpper(r))
			capitalize = false
		} else if i == 0 {
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// toSnakeCase 将字符串转换为蛇形格式
func toSnakeCase(s string, delimiter string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteString(delimiter)
			}
			result.WriteRune(unicode.ToLower(r))
		} else if r == '_' || r == '-' || r == ' ' {
			result.WriteString(delimiter)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}


