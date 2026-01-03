package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ProjectConfig 项目配置
type ProjectConfig struct {
	Name           string // 项目名称
	Version        string // 版本号
	RootNamespace  string // 根命名空间
	SourcePath     string // 源代码路径（默认 "src"）
	VendorPath     string // 依赖路径（默认 "vendor"）
}

// LoadProjectConfig 加载 project.toml 配置文件
func LoadProjectConfig(projectRoot string) (*ProjectConfig, error) {
	configPath := filepath.Join(projectRoot, "project.toml")
	
	// 如果文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ProjectConfig{
			Name:          "",
			Version:       "1.0.0",
			RootNamespace: "",
			SourcePath:    "src",
			VendorPath:    "vendor",
		}, nil
	}

	// 读取文件
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取 project.toml 失败: %s", err)
	}

	// 解析 TOML（简化实现，只解析基本字段）
	config := &ProjectConfig{
		SourcePath:    "src",
		VendorPath:    "vendor",
	}

	lines := strings.Split(string(content), "\n")
	var inProjectSection bool
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 检测 [project] 节
		if line == "[project]" {
			inProjectSection = true
			continue
		}
		
		// 检测其他节（结束 project 节）
		if strings.HasPrefix(line, "[") {
			inProjectSection = false
			continue
		}

		if inProjectSection {
			// 解析键值对
			if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				// 去除引号
				value = strings.Trim(value, `"`)
				value = strings.Trim(value, `'`)

				switch key {
				case "name":
					config.Name = value
				case "version":
					config.Version = value
				case "root_namespace":
					config.RootNamespace = value
				}
			}
		}
	}

	return config, nil
}

// ResolveNamespace 解析命名空间（如果是相对命名空间，添加根命名空间前缀）
func (c *ProjectConfig) ResolveNamespace(namespace string) string {
	if c.RootNamespace == "" {
		return namespace
	}

	// 如果命名空间以点开头，表示相对于根命名空间
	if strings.HasPrefix(namespace, ".") {
		if c.RootNamespace == "" {
			return strings.TrimPrefix(namespace, ".")
		}
		return c.RootNamespace + namespace
	}

	// 如果命名空间是顶级（不包含点），且根命名空间存在，则添加前缀
	if !strings.Contains(namespace, ".") && c.RootNamespace != "" {
		return c.RootNamespace + "." + namespace
	}

	// 否则返回原命名空间（绝对命名空间）
	return namespace
}

// GetSourcePath 获取源代码路径
func (c *ProjectConfig) GetSourcePath(projectRoot string) string {
	return filepath.Join(projectRoot, c.SourcePath)
}

// GetVendorPath 获取依赖路径
func (c *ProjectConfig) GetVendorPath(projectRoot string) string {
	return filepath.Join(projectRoot, c.VendorPath)
}










