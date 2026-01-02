# 日期时间 (System.DateTime)

LongLang 提供完整的日期时间处理能力，融合 C# 的 DateTime API 和 PHP Carbon 的流畅设计。

## 命名空间

```longlang
use System.DateTime.DateTime
use System.DateTime.Duration
use System.DateTime.DateRange
use System.DateTime.Stopwatch
```

## DateTime 类

### 创建实例

```longlang
// 当前时间
now := DateTime::now()
utc := DateTime::utcNow()

// 特殊日期
today := DateTime::today()        // 今天 00:00:00
tomorrow := DateTime::tomorrow()  // 明天 00:00:00
yesterday := DateTime::yesterday() // 昨天 00:00:00

// 指定日期
birthday := DateTime::create(1990, 5, 15, 10, 30, 0)

// 解析字符串
dt := DateTime::parse("2026-01-02 15:04:05")
dt := DateTime::parseExact("02/01/2026", "dd/MM/yyyy")

// 时间戳
dt := DateTime::fromTimestamp(1735833845)
dt := DateTime::fromMillis(1735833845000)
```

### 获取属性

```longlang
now.getYear()         // 2026
now.getMonth()        // 1
now.getDay()          // 2
now.getHour()         // 15
now.getMinute()       // 4
now.getSecond()       // 5
now.getMillisecond()  // 0
now.getDayOfWeek()    // 4 (0=Sunday, 4=Thursday)
now.getDayOfYear()    // 2
now.getWeekOfYear()   // 1
now.getTimestamp()    // Unix 时间戳（秒）
now.getTimestampMillis() // 毫秒时间戳
```

### 加减操作（链式调用）

```longlang
// 向后加
future := now.addYears(1)
         .addMonths(2)
         .addDays(10)
         .addHours(3)
         .addMinutes(30)
         .addSeconds(15)

// 向前减
past := now.subYears(1)
        .subMonths(6)
        .subDays(5)
```

### 设置操作

```longlang
dt := now.setYear(2030)
      .setMonth(12)
      .setDay(25)
      .setHour(10)
```

### 边界操作（Carbon 风格）

```longlang
now.startOfDay()    // 今天 00:00:00
now.endOfDay()      // 今天 23:59:59.999
now.startOfMonth()  // 本月第一天 00:00:00
now.endOfMonth()    // 本月最后一天 23:59:59.999
now.startOfYear()   // 本年第一天
now.endOfYear()     // 本年最后一天
now.startOfWeek()   // 本周一 00:00:00
now.endOfWeek()     // 本周日 23:59:59.999
```

### 比较方法

```longlang
dt1.equals(dt2)           // 相等
dt1.isBefore(dt2)         // dt1 在 dt2 之前
dt1.isAfter(dt2)          // dt1 在 dt2 之后
dt1.isBeforeOrEqual(dt2)  // dt1 <= dt2
dt1.isAfterOrEqual(dt2)   // dt1 >= dt2
dt.isBetween(start, end)  // dt 在 start 和 end 之间

dt1.isSameDay(dt2)        // 同一天
dt1.isSameMonth(dt2)      // 同一月
dt1.isSameYear(dt2)       // 同一年
```

### 周期判断（Carbon 风格）

```longlang
now.isPast()       // 是否已过去
now.isFuture()     // 是否在未来
now.isToday()      // 是否是今天
now.isYesterday()  // 是否是昨天
now.isTomorrow()   // 是否是明天
now.isWeekend()    // 是否是周末
now.isWeekday()    // 是否是工作日
now.isMonday()     // 是否是周一
now.isTuesday()    // ... 到周日
now.isLeapYear()   // 是否是闰年
```

### 差异计算

```longlang
duration := dt1.diff(dt2)     // 返回 Duration 对象

dt1.diffInYears(dt2)    // 年数差
dt1.diffInMonths(dt2)   // 月数差
dt1.diffInDays(dt2)     // 天数差
dt1.diffInHours(dt2)    // 小时差
dt1.diffInMinutes(dt2)  // 分钟差
dt1.diffInSeconds(dt2)  // 秒数差
```

### 人性化输出（多语言）

```longlang
past := DateTime::now().subHours(2)
past.diffForHumans()                    // "2 hours ago"
past.diffForHumansLocale("zh")          // "2 小时前"

future := DateTime::now().addDays(3)
future.diffForHumans()                  // "in 3 days"
future.diffForHumansLocale("zh")        // "3 天后"
```

支持的语言：
- `en` - English（默认）
- `zh` - 中文

### C# 风格格式化

```longlang
// 自定义格式
now.format("yyyy-MM-dd")              // "2026-01-02"
now.format("yyyy/MM/dd HH:mm:ss")     // "2026/01/02 15:04:05"
now.format("dddd, MMMM d, yyyy")      // "Thursday, January 2, 2026"
now.format("hh:mm tt")                // "03:04 PM"

// 标准格式
now.format("s")  // ISO 8601: "2026-01-02T15:04:05"
now.format("D")  // 长日期: "Thursday, January 2, 2026"
now.format("d")  // 短日期: "1/2/2026"

// 快捷方法
now.toDateString()      // "2026-01-02"
now.toTimeString()      // "15:04:05"
now.toDateTimeString()  // "2026-01-02 15:04:05"
now.toISOString()       // "2026-01-02T15:04:05.000Z"
```

## 格式化符号参考

### 自定义格式说明符

| 符号 | 说明 | 示例 |
|------|------|------|
| `yyyy` | 4位年份 | 2026 |
| `yy` | 2位年份 | 26 |
| `MMMM` | 月份全名 | January |
| `MMM` | 月份缩写 | Jan |
| `MM` | 2位月份 | 01 |
| `M` | 月份 | 1 |
| `dddd` | 星期全名 | Thursday |
| `ddd` | 星期缩写 | Thu |
| `dd` | 2位日期 | 02 |
| `d` | 日期 | 2 |
| `HH` | 24小时制 (00-23) | 15 |
| `H` | 24小时制 (0-23) | 15 |
| `hh` | 12小时制 (01-12) | 03 |
| `h` | 12小时制 (1-12) | 3 |
| `mm` | 分钟 (00-59) | 04 |
| `m` | 分钟 (0-59) | 4 |
| `ss` | 秒 (00-59) | 05 |
| `s` | 秒 (0-59) | 5 |
| `fff` | 毫秒 | 123 |
| `tt` | AM/PM | PM |
| `zzz` | 时区偏移 | +08:00 |

### 标准格式字符串

| 符号 | 说明 | 示例 |
|------|------|------|
| `d` | 短日期 | 1/2/2026 |
| `D` | 长日期 | Thursday, January 2, 2026 |
| `t` | 短时间 | 3:04 PM |
| `T` | 长时间 | 3:04:05 PM |
| `f` | 完整日期/短时间 | Thursday, January 2, 2026 3:04 PM |
| `F` | 完整日期/长时间 | Thursday, January 2, 2026 3:04:05 PM |
| `g` | 常规日期/短时间 | 1/2/2026 3:04 PM |
| `G` | 常规日期/长时间 | 1/2/2026 3:04:05 PM |
| `s` | 可排序 (ISO 8601) | 2026-01-02T15:04:05 |
| `o` | 往返格式 | 2026-01-02T15:04:05.0000000+08:00 |
| `r` | RFC1123 | Thu, 02 Jan 2026 15:04:05 GMT |
| `u` | 通用可排序 | 2026-01-02 15:04:05Z |

## Duration 类

表示时间间隔。

```longlang
use System.DateTime.Duration

// 创建
d := Duration::zero()
d := Duration::fromDays(2)
d := Duration::fromHours(48)
d := Duration::fromMinutes(120)
d := Duration::fromSeconds(3600)
d := Duration::between(dt1, dt2)

// 属性
d.getDays()           // 整天数
d.getHours()          // 小时 (0-23)
d.getMinutes()        // 分钟 (0-59)
d.getSeconds()        // 秒 (0-59)

d.getTotalDays()      // 总天数 (float)
d.getTotalHours()     // 总小时数 (float)
d.getTotalMinutes()   // 总分钟数 (float)
d.getTotalSeconds()   // 总秒数 (float)

// 运算
d1.add(d2)           // 相加
d1.sub(d2)           // 相减
d.multiply(3)        // 乘以倍数
d.negate()           // 取反
d.abs()              // 绝对值

// 比较
d.isZero()           // 是否为零
d.isNegative()       // 是否为负
d.isPositive()       // 是否为正

// 格式化
d.toString()                  // "2.03:04:05"
d.toHumanString()             // "2 days 3 hours"
d.toHumanString("zh")         // "2 天 3 小时"
```

## DateRange 类

表示日期范围。

```longlang
use System.DateTime.DateRange

range := new DateRange(start, end)

range.getStart()      // 开始时间
range.getEnd()        // 结束时间
range.getDuration()   // 持续时间

range.contains(dt)           // 是否包含某时间点
range.overlaps(otherRange)   // 是否与另一范围重叠
range.includes(otherRange)   // 是否完全包含另一范围

// 遍历
for _, day := range range.eachDay() {
    fmt.println(day.toDateString())
}
```

## Stopwatch 类

高精度计时器。

```longlang
use System.DateTime.Stopwatch

sw := Stopwatch::startNew()

// 执行一些操作...

sw.stop()
fmt.println("耗时: " + toString(sw.getElapsedMilliseconds()) + " ms")

// 方法
sw.start()     // 开始/继续计时
sw.stop()      // 停止计时
sw.reset()     // 重置
sw.restart()   // 重置并开始

sw.isRunning()              // 是否正在运行
sw.getElapsed()             // 获取 Duration
sw.getElapsedMilliseconds() // 获取毫秒数
```

## 完整示例

```longlang
use System.DateTime.DateTime
use System.DateTime.Duration
use System.DateTime.Stopwatch

// 生日计算
birthday := DateTime::create(1990, 5, 15)
now := DateTime::now()
age := birthday.diffInYears(now)
fmt.println("年龄: " + toString(age) + " 岁")

// 下个周末
nextSaturday := now.startOfWeek().addDays(5)
fmt.println("下个周六: " + nextSaturday.format("yyyy-MM-dd"))

// 倒计时
newYear := DateTime::create(2027, 1, 1)
remaining := now.diff(newYear)
fmt.println("距离新年还有: " + remaining.toHumanString())

// 性能测量
sw := Stopwatch::startNew()
for i := 0; i < 1000000; i++ {
    // 计算
}
sw.stop()
fmt.println("执行耗时: " + toString(sw.getElapsedMilliseconds()) + " ms")
```
