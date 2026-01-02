# byte 类型

`byte` 是 LongLang 中的 8 位无符号整数类型，等价于 `u8`，值范围为 0-255。主要用于处理二进制数据、网络协议和文件 I/O。

## 基本用法

### 声明 byte 变量

```longlang
var b1 byte = 65          // 字母 'A' 的 ASCII 码
var b2 byte = 0           // 最小值
var b3 byte = 255         // 最大值
var b4 byte = 0xFF        // 十六进制
var b5 byte = 0b11111111  // 二进制
```

### 数字字面量

LongLang 支持三种数字字面量格式：

```longlang
// 十进制
value := 255

// 十六进制（0x 或 0X 前缀）
hex := 0xFF       // 255
hex2 := 0x41      // 65 ('A')

// 二进制（0b 或 0B 前缀）
bin := 0b11111111 // 255
bin2 := 0b01000001 // 65 ('A')
```

## byte 数组

### 声明和初始化

```longlang
// 直接初始化
data := []byte{72, 101, 108, 108, 111}  // "Hello"

// 十六进制初始化
hexData := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}  // "Hello"

// 二进制初始化
binData := []byte{0b01001000, 0b01100101}  // "He"
```

### 访问元素

```longlang
data := []byte{72, 101, 108}
first := data[0]      // 72
last := data[-1]      // 108 (负索引)
length := len(data)   // 3
```

## 类型转换

### string → []byte

```longlang
// 使用 toBytes() 函数
bytes := toBytes("Hello World")
fmt.println(len(bytes))  // 11
fmt.println(bytes[0])    // 72 ('H')
```

### []byte → string

```longlang
// 使用 bytesToString() 函数
data := []byte{72, 101, 108, 108, 111}
str := bytesToString(data)
fmt.println(str)  // "Hello"
```

### 字符与 byte 互转

```longlang
// chr(int) - byte/int 转单字符字符串
charA := chr(65)      // "A"
charZ := chr(90)      // "Z"
newline := chr(10)    // "\n"

// ord(string) - 获取字符串第一个字符的 byte 值
codeA := ord("A")     // 65
codeZ := ord("Z")     // 90
```

## 运算

byte 类型支持基本算术运算，结果为 int 类型（防止溢出）：

```longlang
var x byte = 100
var y byte = 50

sum := x + y      // 150 (int)
diff := x - y     // 50 (int)
prod := x * y     // 5000 (int)
quot := x / y     // 2 (int)
```

## 范围检查

当给 byte 类型或 `[]byte` 数组赋值时，会进行范围检查：

```longlang
// 正确
var b byte = 255

// 错误：超出范围
var b byte = 256   // 编译错误
var b byte = -1    // 编译错误

// 数组元素也会检查
data := []byte{256}  // 编译错误
```

## 与标准库配合

### System.Binary

```longlang
use System.Binary.Bytes

// 创建指定大小的字节数组
data := Bytes.create(1024)

// 从字符串创建
bytes := Bytes.fromString("Hello")

// 转回字符串
str := Bytes.toString(bytes)

// 十六进制转换
hex := Bytes.toHex(bytes)      // "48656c6c6f"
bytes2 := Bytes.fromHex(hex)   // 从十六进制字符串创建
```

### System.Net

```longlang
use System.Net.TcpClient

client := new TcpClient()
client.connect("127.0.0.1", 8080)
conn := client.getConnection()

// 发送 byte 数组
data := toBytes("GET / HTTP/1.1\r\n\r\n")
conn.writeBytes(data)

// 接收数据
response := conn.readBytes(1024)
text := bytesToString(response)
```

## 内置函数总结

| 函数 | 说明 | 示例 |
|------|------|------|
| `toBytes(string)` | 字符串转 []byte | `toBytes("Hi")` → `[72, 105]` |
| `bytesToString([]byte)` | []byte 转字符串 | `bytesToString([72, 105])` → `"Hi"` |
| `chr(int)` | 数字转单字符 | `chr(65)` → `"A"` |
| `ord(string)` | 字符转数字 | `ord("A")` → `65` |
| `byteLen(string)` | 字符串字节长度 | `byteLen("你好")` → `6` |

## 使用场景

1. **网络编程**：处理 TCP/UDP 数据包
2. **文件 I/O**：读写二进制文件
3. **协议实现**：如 Redis RESP、HTTP 等
4. **加密/编码**：处理二进制数据
5. **图像处理**：处理像素数据

## 示例：简单的二进制协议

```longlang
namespace Application

use System.Binary.Bytes

class Main {
    public static function main() {
        // 构建消息：[长度(2字节)] + [数据]
        message := "Hello, World!"
        msgBytes := toBytes(message)
        
        // 写入长度（大端序）
        lenBytes := Bytes.writeInt16BE(len(msgBytes))
        
        // 组合完整数据包
        packet := Bytes.concat(lenBytes, msgBytes)
        
        fmt.println("数据包长度: " + toString(len(packet)))
        fmt.println("数据包内容: " + Bytes.toHex(packet))
    }
}
```

