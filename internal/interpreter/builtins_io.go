package interpreter

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// registerIOBuiltins 注册文件操作内置函数
func registerIOBuiltins(env *Environment) {
	// ===== 文件读取函数 =====

	// __file_read_all(path) - 读取文件全部内容为字符串
	env.Set("__file_read_all", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__file_read_all 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__file_read_all 参数必须是字符串，得到 %s", args[0].Type())
		}
		content, err := os.ReadFile(pathStr.Value)
		if err != nil {
			if os.IsNotExist(err) {
				return newError("FileNotFoundException: %s", pathStr.Value)
			}
			if os.IsPermission(err) {
				return newError("PermissionException: read %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		return &String{Value: string(content)}
	}})

	// __file_read_lines(path) - 读取文件全部内容为行数组
	env.Set("__file_read_lines", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__file_read_lines 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__file_read_lines 参数必须是字符串，得到 %s", args[0].Type())
		}
		content, err := os.ReadFile(pathStr.Value)
		if err != nil {
			if os.IsNotExist(err) {
				return newError("FileNotFoundException: %s", pathStr.Value)
			}
			if os.IsPermission(err) {
				return newError("PermissionException: read %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		lines := strings.Split(string(content), "\n")
		// 移除最后的空行（如果文件以换行符结尾）
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		// 处理 Windows 换行符
		for i := range lines {
			lines[i] = strings.TrimSuffix(lines[i], "\r")
		}
		elements := make([]Object, len(lines))
		for i, line := range lines {
			elements[i] = &String{Value: line}
		}
		return &Array{Elements: elements}
	}})

	// ===== 文件写入函数 =====

	// __file_write_all(path, content) - 写入字符串到文件（覆盖）
	env.Set("__file_write_all", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__file_write_all 需要2个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__file_write_all 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		contentStr, ok := args[1].(*String)
		if !ok {
			return newError("__file_write_all 第二个参数必须是字符串，得到 %s", args[1].Type())
		}
		err := os.WriteFile(pathStr.Value, []byte(contentStr.Value), 0644)
		if err != nil {
			if os.IsPermission(err) {
				return newError("PermissionException: write %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __file_write_lines(path, lines) - 写入行数组到文件（覆盖）
	env.Set("__file_write_lines", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__file_write_lines 需要2个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__file_write_lines 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		linesArr, ok := args[1].(*Array)
		if !ok {
			return newError("__file_write_lines 第二个参数必须是数组，得到 %s", args[1].Type())
		}
		var lines []string
		for _, elem := range linesArr.Elements {
			lineStr, ok := elem.(*String)
			if !ok {
				return newError("__file_write_lines 数组元素必须是字符串，得到 %s", elem.Type())
			}
			lines = append(lines, lineStr.Value)
		}
		content := strings.Join(lines, "\n")
		if len(lines) > 0 {
			content += "\n"
		}
		err := os.WriteFile(pathStr.Value, []byte(content), 0644)
		if err != nil {
			if os.IsPermission(err) {
				return newError("PermissionException: write %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __file_append_all(path, content) - 追加字符串到文件末尾
	env.Set("__file_append_all", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__file_append_all 需要2个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__file_append_all 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		contentStr, ok := args[1].(*String)
		if !ok {
			return newError("__file_append_all 第二个参数必须是字符串，得到 %s", args[1].Type())
		}
		f, err := os.OpenFile(pathStr.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			if os.IsPermission(err) {
				return newError("PermissionException: append %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		defer f.Close()
		_, err = f.WriteString(contentStr.Value)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __file_append_lines(path, lines) - 追加行到文件末尾
	env.Set("__file_append_lines", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__file_append_lines 需要2个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__file_append_lines 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		linesArr, ok := args[1].(*Array)
		if !ok {
			return newError("__file_append_lines 第二个参数必须是数组，得到 %s", args[1].Type())
		}
		var lines []string
		for _, elem := range linesArr.Elements {
			lineStr, ok := elem.(*String)
			if !ok {
				return newError("__file_append_lines 数组元素必须是字符串，得到 %s", elem.Type())
			}
			lines = append(lines, lineStr.Value)
		}
		content := strings.Join(lines, "\n")
		if len(lines) > 0 {
			content += "\n"
		}
		f, err := os.OpenFile(pathStr.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			if os.IsPermission(err) {
				return newError("PermissionException: append %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		defer f.Close()
		_, err = f.WriteString(content)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// ===== 文件操作函数 =====

	// __file_exists(path) - 检查文件是否存在
	env.Set("__file_exists", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__file_exists 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__file_exists 参数必须是字符串，得到 %s", args[0].Type())
		}
		info, err := os.Stat(pathStr.Value)
		if err != nil {
			return &Boolean{Value: false}
		}
		return &Boolean{Value: !info.IsDir()}
	}})

	// __file_delete(path) - 删除文件
	env.Set("__file_delete", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__file_delete 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__file_delete 参数必须是字符串，得到 %s", args[0].Type())
		}
		err := os.Remove(pathStr.Value)
		if err != nil {
			if os.IsNotExist(err) {
				return newError("FileNotFoundException: %s", pathStr.Value)
			}
			if os.IsPermission(err) {
				return newError("PermissionException: delete %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __file_copy(source, dest, overwrite) - 复制文件
	env.Set("__file_copy", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("__file_copy 需要3个参数，得到 %d 个", len(args))
		}
		sourceStr, ok := args[0].(*String)
		if !ok {
			return newError("__file_copy 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		destStr, ok := args[1].(*String)
		if !ok {
			return newError("__file_copy 第二个参数必须是字符串，得到 %s", args[1].Type())
		}
		overwriteBool, ok := args[2].(*Boolean)
		if !ok {
			return newError("__file_copy 第三个参数必须是布尔值，得到 %s", args[2].Type())
		}

		// 检查源文件是否存在
		srcFile, err := os.Open(sourceStr.Value)
		if err != nil {
			if os.IsNotExist(err) {
				return newError("FileNotFoundException: %s", sourceStr.Value)
			}
			if os.IsPermission(err) {
				return newError("PermissionException: read %s", sourceStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		defer srcFile.Close()

		// 检查目标文件是否已存在
		if !overwriteBool.Value {
			if _, err := os.Stat(destStr.Value); err == nil {
				return newError("IOException: file already exists: %s", destStr.Value)
			}
		}

		// 创建目标文件
		dstFile, err := os.Create(destStr.Value)
		if err != nil {
			if os.IsPermission(err) {
				return newError("PermissionException: write %s", destStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		defer dstFile.Close()

		// 复制内容
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __file_move(source, dest) - 移动/重命名文件
	env.Set("__file_move", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__file_move 需要2个参数，得到 %d 个", len(args))
		}
		sourceStr, ok := args[0].(*String)
		if !ok {
			return newError("__file_move 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		destStr, ok := args[1].(*String)
		if !ok {
			return newError("__file_move 第二个参数必须是字符串，得到 %s", args[1].Type())
		}
		err := os.Rename(sourceStr.Value, destStr.Value)
		if err != nil {
			if os.IsNotExist(err) {
				return newError("FileNotFoundException: %s", sourceStr.Value)
			}
			if os.IsPermission(err) {
				return newError("PermissionException: move %s", sourceStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// ===== 文件信息函数 =====

	// __file_get_info(path) - 获取文件信息，返回 Map
	env.Set("__file_get_info", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__file_get_info 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__file_get_info 参数必须是字符串，得到 %s", args[0].Type())
		}
		info, err := os.Stat(pathStr.Value)
		if err != nil {
			if os.IsNotExist(err) {
				return newError("FileNotFoundException: %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}

		absPath, _ := filepath.Abs(pathStr.Value)
		result := &Map{
			Pairs:     make(map[string]Object),
			KeyType:   "string",
			ValueType: "any",
		}
		result.Pairs["path"] = &String{Value: absPath}
		result.Pairs["name"] = &String{Value: info.Name()}
		result.Pairs["size"] = &Integer{Value: info.Size()}
		result.Pairs["modTime"] = &Integer{Value: info.ModTime().Unix()}
		result.Pairs["isDir"] = &Boolean{Value: info.IsDir()}
		result.Pairs["isFile"] = &Boolean{Value: !info.IsDir()}

		// 检查权限
		mode := info.Mode()
		result.Pairs["isReadable"] = &Boolean{Value: mode&0444 != 0}
		result.Pairs["isWritable"] = &Boolean{Value: mode&0222 != 0}
		result.Pairs["isExecutable"] = &Boolean{Value: mode&0111 != 0}

		return result
	}})

	// ===== 目录操作函数 =====

	// __dir_create(path, recursive) - 创建目录
	env.Set("__dir_create", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__dir_create 需要2个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__dir_create 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		recursiveBool, ok := args[1].(*Boolean)
		if !ok {
			return newError("__dir_create 第二个参数必须是布尔值，得到 %s", args[1].Type())
		}
		var err error
		if recursiveBool.Value {
			err = os.MkdirAll(pathStr.Value, 0755)
		} else {
			err = os.Mkdir(pathStr.Value, 0755)
		}
		if err != nil {
			if os.IsPermission(err) {
				return newError("PermissionException: create %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __dir_exists(path) - 检查目录是否存在
	env.Set("__dir_exists", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__dir_exists 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__dir_exists 参数必须是字符串，得到 %s", args[0].Type())
		}
		info, err := os.Stat(pathStr.Value)
		if err != nil {
			return &Boolean{Value: false}
		}
		return &Boolean{Value: info.IsDir()}
	}})

	// __dir_delete(path, recursive) - 删除目录
	env.Set("__dir_delete", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__dir_delete 需要2个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__dir_delete 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		recursiveBool, ok := args[1].(*Boolean)
		if !ok {
			return newError("__dir_delete 第二个参数必须是布尔值，得到 %s", args[1].Type())
		}
		var err error
		if recursiveBool.Value {
			err = os.RemoveAll(pathStr.Value)
		} else {
			err = os.Remove(pathStr.Value)
		}
		if err != nil {
			if os.IsNotExist(err) {
				return newError("DirectoryNotFoundException: %s", pathStr.Value)
			}
			if os.IsPermission(err) {
				return newError("PermissionException: delete %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __dir_get_files(path, pattern) - 获取目录下的文件列表
	env.Set("__dir_get_files", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__dir_get_files 需要2个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__dir_get_files 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		patternStr, ok := args[1].(*String)
		if !ok {
			return newError("__dir_get_files 第二个参数必须是字符串，得到 %s", args[1].Type())
		}

		entries, err := os.ReadDir(pathStr.Value)
		if err != nil {
			if os.IsNotExist(err) {
				return newError("DirectoryNotFoundException: %s", pathStr.Value)
			}
			if os.IsPermission(err) {
				return newError("PermissionException: read %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}

		var files []Object
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			// 匹配模式
			matched, err := filepath.Match(patternStr.Value, entry.Name())
			if err != nil {
				return newError("IOException: invalid pattern: %s", patternStr.Value)
			}
			if matched {
				fullPath := filepath.Join(pathStr.Value, entry.Name())
				files = append(files, &String{Value: fullPath})
			}
		}
		return &Array{Elements: files}
	}})

	// __dir_get_directories(path) - 获取目录下的子目录列表
	env.Set("__dir_get_directories", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__dir_get_directories 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__dir_get_directories 参数必须是字符串，得到 %s", args[0].Type())
		}

		entries, err := os.ReadDir(pathStr.Value)
		if err != nil {
			if os.IsNotExist(err) {
				return newError("DirectoryNotFoundException: %s", pathStr.Value)
			}
			if os.IsPermission(err) {
				return newError("PermissionException: read %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}

		var dirs []Object
		for _, entry := range entries {
			if entry.IsDir() {
				fullPath := filepath.Join(pathStr.Value, entry.Name())
				dirs = append(dirs, &String{Value: fullPath})
			}
		}
		return &Array{Elements: dirs}
	}})

	// __dir_get_entries(path) - 获取目录下所有条目
	env.Set("__dir_get_entries", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__dir_get_entries 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__dir_get_entries 参数必须是字符串，得到 %s", args[0].Type())
		}

		entries, err := os.ReadDir(pathStr.Value)
		if err != nil {
			if os.IsNotExist(err) {
				return newError("DirectoryNotFoundException: %s", pathStr.Value)
			}
			if os.IsPermission(err) {
				return newError("PermissionException: read %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}

		var items []Object
		for _, entry := range entries {
			fullPath := filepath.Join(pathStr.Value, entry.Name())
			items = append(items, &String{Value: fullPath})
		}
		return &Array{Elements: items}
	}})

	// __dir_get_current() - 获取当前工作目录
	env.Set("__dir_get_current", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 0 {
			return newError("__dir_get_current 不需要参数，得到 %d 个", len(args))
		}
		cwd, err := os.Getwd()
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &String{Value: cwd}
	}})

	// __dir_set_current(path) - 设置当前工作目录
	env.Set("__dir_set_current", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__dir_set_current 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__dir_set_current 参数必须是字符串，得到 %s", args[0].Type())
		}
		err := os.Chdir(pathStr.Value)
		if err != nil {
			if os.IsNotExist(err) {
				return newError("DirectoryNotFoundException: %s", pathStr.Value)
			}
			if os.IsPermission(err) {
				return newError("PermissionException: chdir %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// ===== 路径操作函数 =====

	// __path_join(...paths) - 组合路径
	env.Set("__path_join", &Builtin{Fn: func(args ...Object) Object {
		if len(args) == 0 {
			return newError("__path_join 至少需要1个参数")
		}
		var parts []string
		for i, arg := range args {
			pathStr, ok := arg.(*String)
			if !ok {
				return newError("__path_join 第 %d 个参数必须是字符串，得到 %s", i+1, arg.Type())
			}
			parts = append(parts, pathStr.Value)
		}
		return &String{Value: filepath.Join(parts...)}
	}})

	// __path_get_filename(path) - 获取文件名（含扩展名）
	env.Set("__path_get_filename", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__path_get_filename 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__path_get_filename 参数必须是字符串，得到 %s", args[0].Type())
		}
		return &String{Value: filepath.Base(pathStr.Value)}
	}})

	// __path_get_filename_without_ext(path) - 获取文件名（不含扩展名）
	env.Set("__path_get_filename_without_ext", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__path_get_filename_without_ext 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__path_get_filename_without_ext 参数必须是字符串，得到 %s", args[0].Type())
		}
		base := filepath.Base(pathStr.Value)
		ext := filepath.Ext(base)
		return &String{Value: strings.TrimSuffix(base, ext)}
	}})

	// __path_get_extension(path) - 获取扩展名（含点号）
	env.Set("__path_get_extension", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__path_get_extension 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__path_get_extension 参数必须是字符串，得到 %s", args[0].Type())
		}
		return &String{Value: filepath.Ext(pathStr.Value)}
	}})

	// __path_get_directory(path) - 获取目录路径
	env.Set("__path_get_directory", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__path_get_directory 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__path_get_directory 参数必须是字符串，得到 %s", args[0].Type())
		}
		return &String{Value: filepath.Dir(pathStr.Value)}
	}})

	// __path_get_absolute(path) - 获取绝对路径
	env.Set("__path_get_absolute", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__path_get_absolute 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__path_get_absolute 参数必须是字符串，得到 %s", args[0].Type())
		}
		abs, err := filepath.Abs(pathStr.Value)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &String{Value: abs}
	}})

	// __path_normalize(path) - 规范化路径
	env.Set("__path_normalize", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__path_normalize 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__path_normalize 参数必须是字符串，得到 %s", args[0].Type())
		}
		return &String{Value: filepath.Clean(pathStr.Value)}
	}})

	// __path_is_absolute(path) - 检查是否是绝对路径
	env.Set("__path_is_absolute", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__path_is_absolute 需要1个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__path_is_absolute 参数必须是字符串，得到 %s", args[0].Type())
		}
		return &Boolean{Value: filepath.IsAbs(pathStr.Value)}
	}})

	// __path_get_relative(from, to) - 获取相对路径
	env.Set("__path_get_relative", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__path_get_relative 需要2个参数，得到 %d 个", len(args))
		}
		fromStr, ok := args[0].(*String)
		if !ok {
			return newError("__path_get_relative 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		toStr, ok := args[1].(*String)
		if !ok {
			return newError("__path_get_relative 第二个参数必须是字符串，得到 %s", args[1].Type())
		}
		rel, err := filepath.Rel(fromStr.Value, toStr.Value)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &String{Value: rel}
	}})

	// __path_change_extension(path, newExt) - 更改扩展名
	env.Set("__path_change_extension", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__path_change_extension 需要2个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__path_change_extension 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		newExtStr, ok := args[1].(*String)
		if !ok {
			return newError("__path_change_extension 第二个参数必须是字符串，得到 %s", args[1].Type())
		}
		ext := filepath.Ext(pathStr.Value)
		base := strings.TrimSuffix(pathStr.Value, ext)
		newExt := newExtStr.Value
		if !strings.HasPrefix(newExt, ".") && newExt != "" {
			newExt = "." + newExt
		}
		return &String{Value: base + newExt}
	}})

	// __path_get_separator() - 获取路径分隔符
	env.Set("__path_get_separator", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 0 {
			return newError("__path_get_separator 不需要参数，得到 %d 个", len(args))
		}
		return &String{Value: string(filepath.Separator)}
	}})

	// __path_get_temp_dir() - 获取临时目录路径
	env.Set("__path_get_temp_dir", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 0 {
			return newError("__path_get_temp_dir 不需要参数，得到 %d 个", len(args))
		}
		return &String{Value: os.TempDir()}
	}})

	// __path_get_temp_file() - 生成临时文件路径
	env.Set("__path_get_temp_file", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 0 {
			return newError("__path_get_temp_file 不需要参数，得到 %d 个", len(args))
		}
		f, err := os.CreateTemp("", "longlang_*")
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		name := f.Name()
		f.Close()
		return &String{Value: name}
	}})

	// ===== 文件流操作函数 =====

	// __stream_open(path, mode) - 打开文件流，返回句柄 ID
	env.Set("__stream_open", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__stream_open 需要2个参数，得到 %d 个", len(args))
		}
		pathStr, ok := args[0].(*String)
		if !ok {
			return newError("__stream_open 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		modeStr, ok := args[1].(*String)
		if !ok {
			return newError("__stream_open 第二个参数必须是字符串，得到 %s", args[1].Type())
		}

		var flag int
		switch modeStr.Value {
		case "r":
			flag = os.O_RDONLY
		case "w":
			flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		case "a":
			flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
		case "rw", "wr":
			flag = os.O_RDWR | os.O_CREATE
		default:
			return newError("IOException: invalid mode: %s", modeStr.Value)
		}

		f, err := os.OpenFile(pathStr.Value, flag, 0644)
		if err != nil {
			if os.IsNotExist(err) {
				return newError("FileNotFoundException: %s", pathStr.Value)
			}
			if os.IsPermission(err) {
				return newError("PermissionException: open %s", pathStr.Value)
			}
			return newError("IOException: %s", err.Error())
		}

		// 使用 FileHandle 对象包装文件句柄
		return &FileHandle{File: f, Path: pathStr.Value, Mode: modeStr.Value}
	}})

	// __stream_read(handle, count) - 读取指定字节数
	env.Set("__stream_read", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__stream_read 需要2个参数，得到 %d 个", len(args))
		}
		handle, ok := args[0].(*FileHandle)
		if !ok {
			return newError("__stream_read 第一个参数必须是文件句柄，得到 %s", args[0].Type())
		}
		if handle.Closed {
			return newError("IOException: file is closed")
		}
		countInt, ok := args[1].(*Integer)
		if !ok {
			return newError("__stream_read 第二个参数必须是整数，得到 %s", args[1].Type())
		}

		buf := make([]byte, countInt.Value)
		n, err := handle.File.Read(buf)
		if err != nil && err != io.EOF {
			return newError("IOException: %s", err.Error())
		}

		// 返回字节数组
		elements := make([]Object, n)
		for i := 0; i < n; i++ {
			elements[i] = &Integer{Value: int64(buf[i])}
		}
		return &Array{Elements: elements}
	}})

	// __stream_read_line(handle) - 读取一行
	env.Set("__stream_read_line", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__stream_read_line 需要1个参数，得到 %d 个", len(args))
		}
		handle, ok := args[0].(*FileHandle)
		if !ok {
			return newError("__stream_read_line 参数必须是文件句柄，得到 %s", args[0].Type())
		}
		if handle.Closed {
			return newError("IOException: file is closed")
		}

		// 逐字节读取直到换行符
		var line []byte
		buf := make([]byte, 1)
		for {
			n, err := handle.File.Read(buf)
			if n == 0 || err == io.EOF {
				break
			}
			if err != nil {
				return newError("IOException: %s", err.Error())
			}
			if buf[0] == '\n' {
				break
			}
			line = append(line, buf[0])
		}
		// 移除可能的 \r
		result := strings.TrimSuffix(string(line), "\r")
		return &String{Value: result}
	}})

	// __stream_write(handle, bytes) - 写入字节数组
	env.Set("__stream_write", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__stream_write 需要2个参数，得到 %d 个", len(args))
		}
		handle, ok := args[0].(*FileHandle)
		if !ok {
			return newError("__stream_write 第一个参数必须是文件句柄，得到 %s", args[0].Type())
		}
		if handle.Closed {
			return newError("IOException: file is closed")
		}
		bytesArr, ok := args[1].(*Array)
		if !ok {
			return newError("__stream_write 第二个参数必须是数组，得到 %s", args[1].Type())
		}

		buf := make([]byte, len(bytesArr.Elements))
		for i, elem := range bytesArr.Elements {
			byteInt, ok := elem.(*Integer)
			if !ok {
				return newError("__stream_write 数组元素必须是整数，得到 %s", elem.Type())
			}
			buf[i] = byte(byteInt.Value)
		}

		n, err := handle.File.Write(buf)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Integer{Value: int64(n)}
	}})

	// __stream_write_text(handle, text) - 写入字符串
	env.Set("__stream_write_text", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__stream_write_text 需要2个参数，得到 %d 个", len(args))
		}
		handle, ok := args[0].(*FileHandle)
		if !ok {
			return newError("__stream_write_text 第一个参数必须是文件句柄，得到 %s", args[0].Type())
		}
		if handle.Closed {
			return newError("IOException: file is closed")
		}
		textStr, ok := args[1].(*String)
		if !ok {
			return newError("__stream_write_text 第二个参数必须是字符串，得到 %s", args[1].Type())
		}

		_, err := handle.File.WriteString(textStr.Value)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __stream_write_line(handle, line) - 写入一行（自动添加换行符）
	env.Set("__stream_write_line", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__stream_write_line 需要2个参数，得到 %d 个", len(args))
		}
		handle, ok := args[0].(*FileHandle)
		if !ok {
			return newError("__stream_write_line 第一个参数必须是文件句柄，得到 %s", args[0].Type())
		}
		if handle.Closed {
			return newError("IOException: file is closed")
		}
		lineStr, ok := args[1].(*String)
		if !ok {
			return newError("__stream_write_line 第二个参数必须是字符串，得到 %s", args[1].Type())
		}

		_, err := handle.File.WriteString(lineStr.Value + "\n")
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __stream_seek(handle, offset, origin) - 移动文件指针
	env.Set("__stream_seek", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("__stream_seek 需要3个参数，得到 %d 个", len(args))
		}
		handle, ok := args[0].(*FileHandle)
		if !ok {
			return newError("__stream_seek 第一个参数必须是文件句柄，得到 %s", args[0].Type())
		}
		if handle.Closed {
			return newError("IOException: file is closed")
		}
		offsetInt, ok := args[1].(*Integer)
		if !ok {
			return newError("__stream_seek 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		originStr, ok := args[2].(*String)
		if !ok {
			return newError("__stream_seek 第三个参数必须是字符串，得到 %s", args[2].Type())
		}

		var whence int
		switch originStr.Value {
		case "begin":
			whence = io.SeekStart
		case "current":
			whence = io.SeekCurrent
		case "end":
			whence = io.SeekEnd
		default:
			return newError("IOException: invalid origin: %s", originStr.Value)
		}

		_, err := handle.File.Seek(offsetInt.Value, whence)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __stream_get_position(handle) - 获取当前位置
	env.Set("__stream_get_position", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__stream_get_position 需要1个参数，得到 %d 个", len(args))
		}
		handle, ok := args[0].(*FileHandle)
		if !ok {
			return newError("__stream_get_position 参数必须是文件句柄，得到 %s", args[0].Type())
		}
		if handle.Closed {
			return newError("IOException: file is closed")
		}

		pos, err := handle.File.Seek(0, io.SeekCurrent)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Integer{Value: pos}
	}})

	// __stream_get_length(handle) - 获取文件长度
	env.Set("__stream_get_length", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__stream_get_length 需要1个参数，得到 %d 个", len(args))
		}
		handle, ok := args[0].(*FileHandle)
		if !ok {
			return newError("__stream_get_length 参数必须是文件句柄，得到 %s", args[0].Type())
		}
		if handle.Closed {
			return newError("IOException: file is closed")
		}

		info, err := handle.File.Stat()
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Integer{Value: info.Size()}
	}})

	// __stream_flush(handle) - 刷新缓冲区到磁盘
	env.Set("__stream_flush", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__stream_flush 需要1个参数，得到 %d 个", len(args))
		}
		handle, ok := args[0].(*FileHandle)
		if !ok {
			return newError("__stream_flush 参数必须是文件句柄，得到 %s", args[0].Type())
		}
		if handle.Closed {
			return newError("IOException: file is closed")
		}

		err := handle.File.Sync()
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __stream_close(handle) - 关闭文件流
	env.Set("__stream_close", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__stream_close 需要1个参数，得到 %d 个", len(args))
		}
		handle, ok := args[0].(*FileHandle)
		if !ok {
			return newError("__stream_close 参数必须是文件句柄，得到 %s", args[0].Type())
		}
		if handle.Closed {
			return &Null{} // 已关闭，不报错
		}

		err := handle.File.Close()
		handle.Closed = true
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __stream_is_eof(handle) - 是否已到文件末尾
	env.Set("__stream_is_eof", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__stream_is_eof 需要1个参数，得到 %d 个", len(args))
		}
		handle, ok := args[0].(*FileHandle)
		if !ok {
			return newError("__stream_is_eof 参数必须是文件句柄，得到 %s", args[0].Type())
		}
		if handle.Closed {
			return &Boolean{Value: true}
		}

		// 获取当前位置
		currentPos, err := handle.File.Seek(0, io.SeekCurrent)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}

		// 获取文件长度
		info, err := handle.File.Stat()
		if err != nil {
			return newError("IOException: %s", err.Error())
		}

		return &Boolean{Value: currentPos >= info.Size()}
	}})

	// __stream_is_closed(handle) - 是否已关闭
	env.Set("__stream_is_closed", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__stream_is_closed 需要1个参数，得到 %d 个", len(args))
		}
		handle, ok := args[0].(*FileHandle)
		if !ok {
			return newError("__stream_is_closed 参数必须是文件句柄，得到 %s", args[0].Type())
		}
		return &Boolean{Value: handle.Closed}
	}})

	// ===== 系统信息函数 =====

	// __os_name() - 获取操作系统名称
	env.Set("__os_name", &Builtin{Fn: func(args ...Object) Object {
		return &String{Value: runtime.GOOS}
	}})

	// __time_now() - 获取当前时间戳（秒）
	env.Set("__time_now", &Builtin{Fn: func(args ...Object) Object {
		return &Integer{Value: time.Now().Unix()}
	}})

	// __time_now_ms() - 获取当前时间戳（毫秒）
	env.Set("__time_now_ms", &Builtin{Fn: func(args ...Object) Object {
		return &Integer{Value: time.Now().UnixMilli()}
	}})
}

// FileHandle 文件句柄对象
type FileHandle struct {
	File   *os.File
	Path   string
	Mode   string
	Closed bool
}

func (fh *FileHandle) Type() ObjectType { return "FILE_HANDLE" }
func (fh *FileHandle) Inspect() string  { return "FileHandle(" + fh.Path + ")" }

