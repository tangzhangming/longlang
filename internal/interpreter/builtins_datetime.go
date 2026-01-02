package interpreter

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// registerDateTimeBuiltins 注册日期时间相关的内置函数
func registerDateTimeBuiltins(env *Environment) {
	// __datetime_now() int - 返回当前时间戳（毫秒）
	env.Set("__datetime_now", &Builtin{Fn: func(args ...Object) Object {
		return &Integer{Value: time.Now().UnixMilli()}
	}})

	// __datetime_now_utc() int - 返回当前 UTC 时间戳（毫秒）
	env.Set("__datetime_now_utc", &Builtin{Fn: func(args ...Object) Object {
		return &Integer{Value: time.Now().UTC().UnixMilli()}
	}})

	// __datetime_from_millis(ms: int) map - 从毫秒时间戳创建日期时间组件
	env.Set("__datetime_from_millis", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__datetime_from_millis 需要1个参数")
		}
		ms, ok := args[0].(*Integer)
		if !ok {
			return newError("__datetime_from_millis 参数必须是整数")
		}

		t := time.UnixMilli(ms.Value)
		return timeToMap(t)
	}})

	// __datetime_from_millis_utc(ms: int) map - UTC 版本
	env.Set("__datetime_from_millis_utc", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__datetime_from_millis_utc 需要1个参数")
		}
		ms, ok := args[0].(*Integer)
		if !ok {
			return newError("__datetime_from_millis_utc 参数必须是整数")
		}

		t := time.UnixMilli(ms.Value).UTC()
		return timeToMap(t)
	}})

	// __datetime_create(year, month, day, hour, minute, second, ms) int - 创建时间戳
	env.Set("__datetime_create", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 7 {
			return newError("__datetime_create 需要7个参数")
		}

		year := getIntArg(args[0])
		month := getIntArg(args[1])
		day := getIntArg(args[2])
		hour := getIntArg(args[3])
		minute := getIntArg(args[4])
		second := getIntArg(args[5])
		ms := getIntArg(args[6])

		t := time.Date(year, time.Month(month), day, hour, minute, second, ms*1000000, time.Local)
		return &Integer{Value: t.UnixMilli()}
	}})

	// __datetime_parse(str: string) int - 解析日期字符串，返回毫秒时间戳
	env.Set("__datetime_parse", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__datetime_parse 需要1个参数")
		}
		str, ok := args[0].(*String)
		if !ok {
			return newError("__datetime_parse 参数必须是字符串")
		}

		t, err := parseDateTime(str.Value)
		if err != nil {
			return newError("无法解析日期时间: %s", str.Value)
		}
		return &Integer{Value: t.UnixMilli()}
	}})

	// __datetime_parse_exact(str: string, format: string) int - 精确格式解析
	env.Set("__datetime_parse_exact", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__datetime_parse_exact 需要2个参数")
		}
		str, ok := args[0].(*String)
		if !ok {
			return newError("__datetime_parse_exact 第一个参数必须是字符串")
		}
		format, ok := args[1].(*String)
		if !ok {
			return newError("__datetime_parse_exact 第二个参数必须是字符串")
		}

		goFormat := convertCSharpFormatToGo(format.Value)
		t, err := time.ParseInLocation(goFormat, str.Value, time.Local)
		if err != nil {
			return newError("无法按格式 '%s' 解析日期时间: %s", format.Value, str.Value)
		}
		return &Integer{Value: t.UnixMilli()}
	}})

	// __datetime_format(ms: int, format: string) string - 格式化日期时间
	env.Set("__datetime_format", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__datetime_format 需要2个参数")
		}
		ms, ok := args[0].(*Integer)
		if !ok {
			return newError("__datetime_format 第一个参数必须是整数")
		}
		format, ok := args[1].(*String)
		if !ok {
			return newError("__datetime_format 第二个参数必须是字符串")
		}

		t := time.UnixMilli(ms.Value)
		result := formatDateTime(t, format.Value)
		return &String{Value: result}
	}})

	// __datetime_add(ms: int, years, months, days, hours, minutes, seconds, milliseconds: int) int
	env.Set("__datetime_add", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 8 {
			return newError("__datetime_add 需要8个参数")
		}

		ms := getIntArg(args[0])
		years := getIntArg(args[1])
		months := getIntArg(args[2])
		days := getIntArg(args[3])
		hours := getIntArg(args[4])
		minutes := getIntArg(args[5])
		seconds := getIntArg(args[6])
		milliseconds := getIntArg(args[7])

		t := time.UnixMilli(int64(ms))
		t = t.AddDate(years, months, days)
		t = t.Add(time.Duration(hours)*time.Hour +
			time.Duration(minutes)*time.Minute +
			time.Duration(seconds)*time.Second +
			time.Duration(milliseconds)*time.Millisecond)

		return &Integer{Value: t.UnixMilli()}
	}})

	// __datetime_diff(ms1: int, ms2: int) map - 计算两个时间的差异
	env.Set("__datetime_diff", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__datetime_diff 需要2个参数")
		}
		ms1 := getIntArg(args[0])
		ms2 := getIntArg(args[1])

		diff := int64(ms2) - int64(ms1)
		return &Integer{Value: diff} // 返回毫秒差
	}})

	// __datetime_day_of_week(ms: int) int - 获取星期几 (0=Sunday)
	env.Set("__datetime_day_of_week", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__datetime_day_of_week 需要1个参数")
		}
		ms := getIntArg(args[0])
		t := time.UnixMilli(int64(ms))
		return &Integer{Value: int64(t.Weekday())}
	}})

	// __datetime_day_of_year(ms: int) int - 获取一年中的第几天
	env.Set("__datetime_day_of_year", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__datetime_day_of_year 需要1个参数")
		}
		ms := getIntArg(args[0])
		t := time.UnixMilli(int64(ms))
		return &Integer{Value: int64(t.YearDay())}
	}})

	// __datetime_week_of_year(ms: int) int - 获取一年中的第几周
	env.Set("__datetime_week_of_year", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__datetime_week_of_year 需要1个参数")
		}
		ms := getIntArg(args[0])
		t := time.UnixMilli(int64(ms))
		_, week := t.ISOWeek()
		return &Integer{Value: int64(week)}
	}})

	// __datetime_is_leap_year(year: int) bool - 判断是否是闰年
	env.Set("__datetime_is_leap_year", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__datetime_is_leap_year 需要1个参数")
		}
		year := getIntArg(args[0])
		isLeap := (year%4 == 0 && year%100 != 0) || (year%400 == 0)
		return &Boolean{Value: isLeap}
	}})

	// __datetime_days_in_month(year: int, month: int) int - 获取某月的天数
	env.Set("__datetime_days_in_month", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__datetime_days_in_month 需要2个参数")
		}
		year := getIntArg(args[0])
		month := getIntArg(args[1])
		// 获取下个月的第0天，即当月最后一天
		t := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.Local)
		return &Integer{Value: int64(t.Day())}
	}})

	// __datetime_set_timezone(ms: int, tz: string) int - 设置时区
	env.Set("__datetime_set_timezone", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__datetime_set_timezone 需要2个参数")
		}
		ms := getIntArg(args[0])
		tzStr, ok := args[1].(*String)
		if !ok {
			return newError("__datetime_set_timezone 第二个参数必须是字符串")
		}

		t := time.UnixMilli(int64(ms))
		
		var loc *time.Location
		var err error
		
		switch tzStr.Value {
		case "UTC":
			loc = time.UTC
		case "Local":
			loc = time.Local
		default:
			// 尝试解析时区偏移（如 "+08:00"）
			if strings.HasPrefix(tzStr.Value, "+") || strings.HasPrefix(tzStr.Value, "-") {
				loc, err = parseTimezoneOffset(tzStr.Value)
			} else {
				loc, err = time.LoadLocation(tzStr.Value)
			}
		}
		
		if err != nil {
			return newError("无效的时区: %s", tzStr.Value)
		}
		
		t = t.In(loc)
		return &Integer{Value: t.UnixMilli()}
	}})

	// __datetime_get_timezone(ms: int) string - 获取时区
	env.Set("__datetime_get_timezone", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__datetime_get_timezone 需要1个参数")
		}
		ms := getIntArg(args[0])
		t := time.UnixMilli(int64(ms))
		name, offset := t.Zone()
		if name == "" {
			// 返回偏移格式
			hours := offset / 3600
			minutes := (offset % 3600) / 60
			if offset >= 0 {
				return &String{Value: fmt.Sprintf("+%02d:%02d", hours, minutes)}
			}
			return &String{Value: fmt.Sprintf("-%02d:%02d", -hours, -minutes)}
		}
		return &String{Value: name}
	}})

	// __datetime_to_utc(ms: int) int - 转换为 UTC
	env.Set("__datetime_to_utc", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__datetime_to_utc 需要1个参数")
		}
		ms := getIntArg(args[0])
		t := time.UnixMilli(int64(ms)).UTC()
		return &Integer{Value: t.UnixMilli()}
	}})

	// __stopwatch_start() int - 开始计时，返回纳秒时间戳
	env.Set("__stopwatch_start", &Builtin{Fn: func(args ...Object) Object {
		return &Integer{Value: time.Now().UnixNano()}
	}})

	// __stopwatch_elapsed(startNano: int) int - 计算经过的毫秒数
	env.Set("__stopwatch_elapsed", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__stopwatch_elapsed 需要1个参数")
		}
		startNano := getIntArg(args[0])
		elapsed := time.Now().UnixNano() - int64(startNano)
		return &Integer{Value: elapsed / 1000000} // 转换为毫秒
	}})

	// __stopwatch_elapsed_nanos(startNano: int) int - 计算经过的纳秒数
	env.Set("__stopwatch_elapsed_nanos", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__stopwatch_elapsed_nanos 需要1个参数")
		}
		startNano := getIntArg(args[0])
		elapsed := time.Now().UnixNano() - int64(startNano)
		return &Integer{Value: elapsed}
	}})
}

// timeToMap 将 time.Time 转换为 map
func timeToMap(t time.Time) *Map {
	m := &Map{
		Pairs:     make(map[string]Object),
		Keys:      []string{},
		KeyType:   "string",
		ValueType: "int",
	}
	m.Set("year", &Integer{Value: int64(t.Year())})
	m.Set("month", &Integer{Value: int64(t.Month())})
	m.Set("day", &Integer{Value: int64(t.Day())})
	m.Set("hour", &Integer{Value: int64(t.Hour())})
	m.Set("minute", &Integer{Value: int64(t.Minute())})
	m.Set("second", &Integer{Value: int64(t.Second())})
	m.Set("millisecond", &Integer{Value: int64(t.Nanosecond() / 1000000)})
	m.Set("timestamp", &Integer{Value: t.UnixMilli()})
	return m
}

// getIntArg 从 Object 获取 int 值
func getIntArg(obj Object) int {
	if i, ok := obj.(*Integer); ok {
		return int(i.Value)
	}
	return 0
}

// parseDateTime 尝试解析多种日期时间格式
func parseDateTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
		"01/02/2006",
		"01/02/2006 15:04:05",
		"02-01-2006",
		"02-01-2006 15:04:05",
		time.RFC3339,
		time.RFC1123,
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析: %s", s)
}

// parseTimezoneOffset 解析时区偏移字符串（如 "+08:00"）
func parseTimezoneOffset(s string) (*time.Location, error) {
	sign := 1
	if s[0] == '-' {
		sign = -1
	}
	s = s[1:]
	
	parts := strings.Split(s, ":")
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}
	
	minutes := 0
	if len(parts) > 1 {
		minutes, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
	}
	
	offset := sign * (hours*3600 + minutes*60)
	return time.FixedZone(s, offset), nil
}

// convertCSharpFormatToGo 将 C# 格式字符串转换为 Go 格式
func convertCSharpFormatToGo(format string) string {
	// C# 格式到 Go 格式的映射
	// 使用唯一占位符避免替换冲突（使用 § 作为分隔符，因为它不会出现在格式字符串中）
	result := format
	
	// 所有替换规则（从长到短排序很重要）
	replacements := []struct {
		csharp      string
		placeholder string
		golang      string
	}{
		// 年份
		{"yyyy", "§01§", "2006"},
		{"yy", "§02§", "06"},
		// 月份 - 名称
		{"MMMM", "§03§", "January"},
		{"MMM", "§04§", "Jan"},
		// 月份 - 数字
		{"MM", "§05§", "01"},
		{"M", "§06§", "1"},
		// 星期
		{"dddd", "§07§", "Monday"},
		{"ddd", "§08§", "Mon"},
		// 日期
		{"dd", "§09§", "02"},
		{"d", "§10§", "2"},
		// 小时 24小时制
		{"HH", "§11§", "15"},
		{"H", "§12§", "15"},
		// 小时 12小时制
		{"hh", "§13§", "03"},
		{"h", "§14§", "3"},
		// 分钟
		{"mm", "§15§", "04"},
		{"m", "§16§", "4"},
		// 秒
		{"ss", "§17§", "05"},
		{"s", "§18§", "5"},
		// 毫秒
		{"fff", "§19§", "000"},
		{"ff", "§20§", "00"},
		{"f", "§21§", "0"},
		// AM/PM
		{"tt", "§22§", "PM"},
		{"t", "§23§", "PM"},
		// 时区
		{"zzz", "§24§", "-07:00"},
		{"zz", "§25§", "-07"},
		{"z", "§26§", "-7"},
	}
	
	// 第一步：将所有 C# 格式符替换为占位符
	for _, r := range replacements {
		result = strings.ReplaceAll(result, r.csharp, r.placeholder)
	}
	
	// 第二步：将所有占位符替换为 Go 格式
	for _, r := range replacements {
		result = strings.ReplaceAll(result, r.placeholder, r.golang)
	}
	
	return result
}

// formatDateTime 格式化日期时间
func formatDateTime(t time.Time, format string) string {
	// 处理标准格式字符串
	switch format {
	case "d": // 短日期
		return t.Format("1/2/2006")
	case "D": // 长日期
		return t.Format("Monday, January 2, 2006")
	case "t": // 短时间
		return t.Format("3:04 PM")
	case "T": // 长时间
		return t.Format("3:04:05 PM")
	case "f": // 完整日期/短时间
		return t.Format("Monday, January 2, 2006 3:04 PM")
	case "F": // 完整日期/长时间
		return t.Format("Monday, January 2, 2006 3:04:05 PM")
	case "g": // 常规日期/短时间
		return t.Format("1/2/2006 3:04 PM")
	case "G": // 常规日期/长时间
		return t.Format("1/2/2006 3:04:05 PM")
	case "s": // 可排序 (ISO 8601)
		return t.Format("2006-01-02T15:04:05")
	case "o", "O": // 往返格式
		return t.Format("2006-01-02T15:04:05.0000000-07:00")
	case "r", "R": // RFC1123
		return t.Format(time.RFC1123)
	case "u": // 通用可排序
		return t.UTC().Format("2006-01-02 15:04:05Z")
	}

	// 自定义格式
	goFormat := convertCSharpFormatToGo(format)
	return t.Format(goFormat)
}
