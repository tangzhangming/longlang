package interpreter

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// registerNetBuiltins 注册网络操作内置函数
func registerNetBuiltins(env *Environment) {
	// ===== TCP 监听器函数 =====

	// __tcp_listen(host, port) - 创建 TCP 监听器
	env.Set("__tcp_listen", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__tcp_listen 需要2个参数，得到 %d 个", len(args))
		}
		hostStr, ok := args[0].(*String)
		if !ok {
			return newError("__tcp_listen 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		portInt, ok := args[1].(*Integer)
		if !ok {
			return newError("__tcp_listen 第二个参数必须是整数，得到 %s", args[1].Type())
		}

		addr := fmt.Sprintf("%s:%d", hostStr.Value, portInt.Value)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}

		return &TcpListener{Listener: listener, Address: addr}
	}})

	// __tcp_listener_accept(listener) - 接受连接
	env.Set("__tcp_listener_accept", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__tcp_listener_accept 需要1个参数，得到 %d 个", len(args))
		}
		listener, ok := args[0].(*TcpListener)
		if !ok {
			return newError("__tcp_listener_accept 参数必须是 TcpListener，得到 %s", args[0].Type())
		}
		if listener.Closed {
			return newError("IOException: listener is closed")
		}

		conn, err := listener.Listener.Accept()
		if err != nil {
			return newError("IOException: %s", err.Error())
		}

		return &TcpConnection{
			Conn:       conn,
			Reader:     bufio.NewReader(conn),
			Writer:     bufio.NewWriter(conn),
			LocalAddr:  conn.LocalAddr().String(),
			RemoteAddr: conn.RemoteAddr().String(),
		}
	}})

	// __tcp_listener_close(listener) - 关闭监听器
	env.Set("__tcp_listener_close", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__tcp_listener_close 需要1个参数，得到 %d 个", len(args))
		}
		listener, ok := args[0].(*TcpListener)
		if !ok {
			return newError("__tcp_listener_close 参数必须是 TcpListener，得到 %s", args[0].Type())
		}
		if listener.Closed {
			return &Null{}
		}

		err := listener.Listener.Close()
		listener.Closed = true
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __tcp_listener_get_address(listener) - 获取监听地址
	env.Set("__tcp_listener_get_address", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__tcp_listener_get_address 需要1个参数，得到 %d 个", len(args))
		}
		listener, ok := args[0].(*TcpListener)
		if !ok {
			return newError("__tcp_listener_get_address 参数必须是 TcpListener，得到 %s", args[0].Type())
		}
		return &String{Value: listener.Listener.Addr().String()}
	}})

	// ===== TCP 客户端函数 =====

	// __tcp_connect(host, port) - 连接到 TCP 服务器
	env.Set("__tcp_connect", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__tcp_connect 需要2个参数，得到 %d 个", len(args))
		}
		hostStr, ok := args[0].(*String)
		if !ok {
			return newError("__tcp_connect 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		portInt, ok := args[1].(*Integer)
		if !ok {
			return newError("__tcp_connect 第二个参数必须是整数，得到 %s", args[1].Type())
		}

		addr := fmt.Sprintf("%s:%d", hostStr.Value, portInt.Value)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}

		return &TcpConnection{
			Conn:       conn,
			Reader:     bufio.NewReader(conn),
			Writer:     bufio.NewWriter(conn),
			LocalAddr:  conn.LocalAddr().String(),
			RemoteAddr: conn.RemoteAddr().String(),
		}
	}})

	// __tcp_connect_timeout(host, port, timeout_ms) - 带超时连接到 TCP 服务器
	env.Set("__tcp_connect_timeout", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 3 {
			return newError("__tcp_connect_timeout 需要3个参数，得到 %d 个", len(args))
		}
		hostStr, ok := args[0].(*String)
		if !ok {
			return newError("__tcp_connect_timeout 第一个参数必须是字符串，得到 %s", args[0].Type())
		}
		portInt, ok := args[1].(*Integer)
		if !ok {
			return newError("__tcp_connect_timeout 第二个参数必须是整数，得到 %s", args[1].Type())
		}
		timeoutInt, ok := args[2].(*Integer)
		if !ok {
			return newError("__tcp_connect_timeout 第三个参数必须是整数，得到 %s", args[2].Type())
		}

		addr := fmt.Sprintf("%s:%d", hostStr.Value, portInt.Value)
		timeout := time.Duration(timeoutInt.Value) * time.Millisecond
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}

		return &TcpConnection{
			Conn:       conn,
			Reader:     bufio.NewReader(conn),
			Writer:     bufio.NewWriter(conn),
			LocalAddr:  conn.LocalAddr().String(),
			RemoteAddr: conn.RemoteAddr().String(),
		}
	}})

	// ===== TCP 连接函数 =====

	// __tcp_conn_read(conn, count) - 读取指定字节数
	env.Set("__tcp_conn_read", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__tcp_conn_read 需要2个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_read 第一个参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		if conn.Closed {
			return newError("IOException: connection is closed")
		}
		countInt, ok := args[1].(*Integer)
		if !ok {
			return newError("__tcp_conn_read 第二个参数必须是整数，得到 %s", args[1].Type())
		}

		buf := make([]byte, countInt.Value)
		n, err := conn.Reader.Read(buf)
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

	// __tcp_conn_read_line(conn) - 读取一行
	env.Set("__tcp_conn_read_line", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__tcp_conn_read_line 需要1个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_read_line 参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		if conn.Closed {
			return newError("IOException: connection is closed")
		}

		line, err := conn.Reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return newError("IOException: %s", err.Error())
		}
		// 移除换行符
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")
		return &String{Value: line}
	}})

	// __tcp_conn_read_all(conn) - 读取所有数据直到 EOF
	env.Set("__tcp_conn_read_all", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__tcp_conn_read_all 需要1个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_read_all 参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		if conn.Closed {
			return newError("IOException: connection is closed")
		}

		data, err := io.ReadAll(conn.Reader)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &String{Value: string(data)}
	}})

	// __tcp_conn_write(conn, data) - 写入字符串
	env.Set("__tcp_conn_write", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__tcp_conn_write 需要2个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_write 第一个参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		if conn.Closed {
			return newError("IOException: connection is closed")
		}
		dataStr, ok := args[1].(*String)
		if !ok {
			return newError("__tcp_conn_write 第二个参数必须是字符串，得到 %s", args[1].Type())
		}

		n, err := conn.Writer.WriteString(dataStr.Value)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Integer{Value: int64(n)}
	}})

	// __tcp_conn_write_line(conn, line) - 写入一行（自动添加换行符）
	env.Set("__tcp_conn_write_line", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__tcp_conn_write_line 需要2个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_write_line 第一个参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		if conn.Closed {
			return newError("IOException: connection is closed")
		}
		lineStr, ok := args[1].(*String)
		if !ok {
			return newError("__tcp_conn_write_line 第二个参数必须是字符串，得到 %s", args[1].Type())
		}

		n, err := conn.Writer.WriteString(lineStr.Value + "\n")
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Integer{Value: int64(n)}
	}})

	// __tcp_conn_write_bytes(conn, bytes) - 写入字节数组
	env.Set("__tcp_conn_write_bytes", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__tcp_conn_write_bytes 需要2个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_write_bytes 第一个参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		if conn.Closed {
			return newError("IOException: connection is closed")
		}
		bytesArr, ok := args[1].(*Array)
		if !ok {
			return newError("__tcp_conn_write_bytes 第二个参数必须是数组，得到 %s", args[1].Type())
		}

		buf := make([]byte, len(bytesArr.Elements))
		for i, elem := range bytesArr.Elements {
			byteInt, ok := elem.(*Integer)
			if !ok {
				return newError("__tcp_conn_write_bytes 数组元素必须是整数，得到 %s", elem.Type())
			}
			buf[i] = byte(byteInt.Value)
		}

		n, err := conn.Writer.Write(buf)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Integer{Value: int64(n)}
	}})

	// __tcp_conn_flush(conn) - 刷新缓冲区
	env.Set("__tcp_conn_flush", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__tcp_conn_flush 需要1个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_flush 参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		if conn.Closed {
			return &Null{}
		}

		err := conn.Writer.Flush()
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __tcp_conn_close(conn) - 关闭连接
	env.Set("__tcp_conn_close", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__tcp_conn_close 需要1个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_close 参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		if conn.Closed {
			return &Null{}
		}

		// 先刷新缓冲区
		conn.Writer.Flush()
		err := conn.Conn.Close()
		conn.Closed = true
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __tcp_conn_get_local_addr(conn) - 获取本地地址
	env.Set("__tcp_conn_get_local_addr", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__tcp_conn_get_local_addr 需要1个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_get_local_addr 参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		return &String{Value: conn.LocalAddr}
	}})

	// __tcp_conn_get_remote_addr(conn) - 获取远程地址
	env.Set("__tcp_conn_get_remote_addr", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__tcp_conn_get_remote_addr 需要1个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_get_remote_addr 参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		return &String{Value: conn.RemoteAddr}
	}})

	// __tcp_conn_is_closed(conn) - 是否已关闭
	env.Set("__tcp_conn_is_closed", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__tcp_conn_is_closed 需要1个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_is_closed 参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		return &Boolean{Value: conn.Closed}
	}})

	// __tcp_conn_set_timeout(conn, timeout_ms) - 设置读写超时
	env.Set("__tcp_conn_set_timeout", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__tcp_conn_set_timeout 需要2个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_set_timeout 第一个参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		if conn.Closed {
			return newError("IOException: connection is closed")
		}
		timeoutInt, ok := args[1].(*Integer)
		if !ok {
			return newError("__tcp_conn_set_timeout 第二个参数必须是整数，得到 %s", args[1].Type())
		}

		timeout := time.Duration(timeoutInt.Value) * time.Millisecond
		deadline := time.Now().Add(timeout)
		err := conn.Conn.SetDeadline(deadline)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __tcp_conn_set_read_timeout(conn, timeout_ms) - 设置读超时
	env.Set("__tcp_conn_set_read_timeout", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__tcp_conn_set_read_timeout 需要2个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_set_read_timeout 第一个参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		if conn.Closed {
			return newError("IOException: connection is closed")
		}
		timeoutInt, ok := args[1].(*Integer)
		if !ok {
			return newError("__tcp_conn_set_read_timeout 第二个参数必须是整数，得到 %s", args[1].Type())
		}

		timeout := time.Duration(timeoutInt.Value) * time.Millisecond
		deadline := time.Now().Add(timeout)
		err := conn.Conn.SetReadDeadline(deadline)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// __tcp_conn_set_write_timeout(conn, timeout_ms) - 设置写超时
	env.Set("__tcp_conn_set_write_timeout", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 2 {
			return newError("__tcp_conn_set_write_timeout 需要2个参数，得到 %d 个", len(args))
		}
		conn, ok := args[0].(*TcpConnection)
		if !ok {
			return newError("__tcp_conn_set_write_timeout 第一个参数必须是 TcpConnection，得到 %s", args[0].Type())
		}
		if conn.Closed {
			return newError("IOException: connection is closed")
		}
		timeoutInt, ok := args[1].(*Integer)
		if !ok {
			return newError("__tcp_conn_set_write_timeout 第二个参数必须是整数，得到 %s", args[1].Type())
		}

		timeout := time.Duration(timeoutInt.Value) * time.Millisecond
		deadline := time.Now().Add(timeout)
		err := conn.Conn.SetWriteDeadline(deadline)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}
		return &Null{}
	}})

	// ===== 工具函数 =====

	// __net_parse_address(address) - 解析地址为 host 和 port
	env.Set("__net_parse_address", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__net_parse_address 需要1个参数，得到 %d 个", len(args))
		}
		addrStr, ok := args[0].(*String)
		if !ok {
			return newError("__net_parse_address 参数必须是字符串，得到 %s", args[0].Type())
		}

		host, portStr, err := net.SplitHostPort(addrStr.Value)
		if err != nil {
			return newError("IOException: invalid address: %s", addrStr.Value)
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			return newError("IOException: invalid port: %s", portStr)
		}

		result := &Map{
			Pairs:     make(map[string]Object),
			KeyType:   "string",
			ValueType: "any",
		}
		result.Pairs["host"] = &String{Value: host}
		result.Pairs["port"] = &Integer{Value: int64(port)}
		return result
	}})

	// __net_resolve_host(hostname) - DNS 解析
	env.Set("__net_resolve_host", &Builtin{Fn: func(args ...Object) Object {
		if len(args) != 1 {
			return newError("__net_resolve_host 需要1个参数，得到 %d 个", len(args))
		}
		hostStr, ok := args[0].(*String)
		if !ok {
			return newError("__net_resolve_host 参数必须是字符串，得到 %s", args[0].Type())
		}

		ips, err := net.LookupHost(hostStr.Value)
		if err != nil {
			return newError("IOException: %s", err.Error())
		}

		elements := make([]Object, len(ips))
		for i, ip := range ips {
			elements[i] = &String{Value: ip}
		}
		return &Array{Elements: elements}
	}})
}

// TcpListener TCP 监听器对象
type TcpListener struct {
	Listener net.Listener
	Address  string
	Closed   bool
}

func (tl *TcpListener) Type() ObjectType { return "TCP_LISTENER" }
func (tl *TcpListener) Inspect() string  { return "TcpListener(" + tl.Address + ")" }

// TcpConnection TCP 连接对象
type TcpConnection struct {
	Conn       net.Conn
	Reader     *bufio.Reader
	Writer     *bufio.Writer
	LocalAddr  string
	RemoteAddr string
	Closed     bool
}

func (tc *TcpConnection) Type() ObjectType { return "TCP_CONNECTION" }
func (tc *TcpConnection) Inspect() string  { return "TcpConnection(" + tc.RemoteAddr + ")" }

