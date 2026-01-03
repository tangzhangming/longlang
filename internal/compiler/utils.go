package compiler

import "strings"

// toPascalCase 转换为 PascalCase
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(string(part[0])))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}
	return result.String()
}

// toCamelCase 转换为 camelCase
func toCamelCase(s string) string {
	if s == "" {
		return ""
	}
	pascal := toPascalCase(s)
	if len(pascal) > 0 {
		return strings.ToLower(string(pascal[0])) + pascal[1:]
	}
	return pascal
}

