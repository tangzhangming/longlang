package interpreter

import (
	"sync"

	"github.com/tangzhangming/longlang/internal/parser"
)

// evalGoStatement 执行 go 语句，启动一个新的协程
func (i *Interpreter) evalGoStatement(node *parser.GoStatement) Object {
	// 克隆当前环境，协程使用独立的环境副本
	envCopy := i.env.Clone()

	// 保存当前命名空间和其他状态
	currentNamespace := i.currentNamespace
	stdlibPath := i.stdlibPath
	projectRoot := i.projectRoot
	projectConfig := i.projectConfig
	namespaceMgr := i.namespaceMgr
	loadedNamespaces := make(map[string]bool)
	for k, v := range i.loadedNamespaces {
		loadedNamespaces[k] = v
	}

	// 启动 Go 协程
	go func() {
		// 创建新的解释器实例用于协程执行
		goroutineInterpreter := &Interpreter{
			env:              envCopy,
			currentNamespace: currentNamespace,
			stdlibPath:       stdlibPath,
			projectRoot:      projectRoot,
			projectConfig:    projectConfig,
			namespaceMgr:     namespaceMgr,
			loadedNamespaces: loadedNamespaces,
		}

		// 执行表达式
		goroutineInterpreter.Eval(node.Call)
	}()

	return nil
}

// ========== Channel 实现 ==========

// ChannelObject Channel 对象
// 封装 Go 的 channel
type ChannelObject struct {
	ch       chan Object
	capacity int
	closed   bool
	mu       sync.RWMutex
}

func (c *ChannelObject) Type() ObjectType { return "CHANNEL" }
func (c *ChannelObject) Inspect() string {
	if c.capacity == 0 {
		return "Channel(unbuffered)"
	}
	return "Channel(capacity=" + string(rune(c.capacity+'0')) + ")"
}

// NewChannel 创建新的 Channel
func NewChannel(capacity int) *ChannelObject {
	return &ChannelObject{
		ch:       make(chan Object, capacity),
		capacity: capacity,
		closed:   false,
	}
}

// Send 发送数据到 Channel
func (c *ChannelObject) Send(value Object) bool {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return false
	}
	c.mu.RUnlock()

	c.ch <- value
	return true
}

// Receive 从 Channel 接收数据（阻塞）
func (c *ChannelObject) Receive() (Object, bool) {
	value, ok := <-c.ch
	return value, ok
}

// TryReceive 尝试从 Channel 接收数据（非阻塞）
func (c *ChannelObject) TryReceive() (Object, bool) {
	select {
	case value, ok := <-c.ch:
		return value, ok
	default:
		return nil, false
	}
}

// Close 关闭 Channel
func (c *ChannelObject) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.closed = true
		close(c.ch)
	}
}

// IsClosed 检查 Channel 是否已关闭
func (c *ChannelObject) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// Len 返回 Channel 中当前的元素数量
func (c *ChannelObject) Len() int {
	return len(c.ch)
}

// Cap 返回 Channel 的容量
func (c *ChannelObject) Cap() int {
	return c.capacity
}

// ========== WaitGroup 实现 ==========

// WaitGroupObject WaitGroup 对象
type WaitGroupObject struct {
	wg sync.WaitGroup
}

func (w *WaitGroupObject) Type() ObjectType { return "WAITGROUP" }
func (w *WaitGroupObject) Inspect() string  { return "WaitGroup" }

// NewWaitGroup 创建新的 WaitGroup
func NewWaitGroup() *WaitGroupObject {
	return &WaitGroupObject{}
}

// Add 添加等待计数
func (w *WaitGroupObject) Add(delta int) {
	w.wg.Add(delta)
}

// Done 完成一个等待
func (w *WaitGroupObject) Done() {
	w.wg.Done()
}

// Wait 等待所有完成
func (w *WaitGroupObject) Wait() {
	w.wg.Wait()
}

// ========== Mutex 实现 ==========

// MutexObject Mutex 对象
type MutexObject struct {
	mu sync.Mutex
}

func (m *MutexObject) Type() ObjectType { return "MUTEX" }
func (m *MutexObject) Inspect() string  { return "Mutex" }

// ========== Atomic 实现（用于协程间共享数据） ==========

// AtomicObject 原子值对象
// 支持协程间安全地共享数据
type AtomicObject struct {
	value Object
	mu    sync.RWMutex
}

func (a *AtomicObject) Type() ObjectType { return "ATOMIC" }
func (a *AtomicObject) Inspect() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.value == nil {
		return "Atomic(null)"
	}
	return "Atomic(" + a.value.Inspect() + ")"
}

// NewAtomic 创建新的 Atomic
func NewAtomic(value Object) *AtomicObject {
	return &AtomicObject{value: value}
}

// Get 获取值
func (a *AtomicObject) Get() Object {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.value
}

// Set 设置值
func (a *AtomicObject) Set(value Object) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.value = value
}

// CompareAndSwap 比较并交换
func (a *AtomicObject) CompareAndSwap(expected, newValue Object) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if atomicObjectsEqual(a.value, expected) {
		a.value = newValue
		return true
	}
	return false
}

// atomicObjectsEqual 比较两个对象是否相等
func atomicObjectsEqual(a, b Object) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *Integer:
		if bv, ok := b.(*Integer); ok {
			return av.Value == bv.Value
		}
	case *Float:
		if bv, ok := b.(*Float); ok {
			return av.Value == bv.Value
		}
	case *String:
		if bv, ok := b.(*String); ok {
			return av.Value == bv.Value
		}
	case *Boolean:
		if bv, ok := b.(*Boolean); ok {
			return av.Value == bv.Value
		}
	case *Null:
		_, ok := b.(*Null)
		return ok
	}
	// 对于其他类型，比较指针
	return a == b
}

// NewMutex 创建新的 Mutex
func NewMutex() *MutexObject {
	return &MutexObject{}
}

// Lock 加锁
func (m *MutexObject) Lock() {
	m.mu.Lock()
}

// Unlock 解锁
func (m *MutexObject) Unlock() {
	m.mu.Unlock()
}

// TryLock 尝试加锁（非阻塞）
func (m *MutexObject) TryLock() bool {
	return m.mu.TryLock()
}

// ========== 绑定方法对象 ==========

// BoundChannelMethod Channel 方法绑定
type BoundChannelMethod struct {
	Channel    *ChannelObject
	MethodName string
}

func (b *BoundChannelMethod) Type() ObjectType { return FUNCTION_OBJ }
func (b *BoundChannelMethod) Inspect() string  { return "channel method " + b.MethodName }

// BoundWaitGroupMethod WaitGroup 方法绑定
type BoundWaitGroupMethod struct {
	WaitGroup  *WaitGroupObject
	MethodName string
}

func (b *BoundWaitGroupMethod) Type() ObjectType { return FUNCTION_OBJ }
func (b *BoundWaitGroupMethod) Inspect() string  { return "waitgroup method " + b.MethodName }

// BoundMutexMethod Mutex 方法绑定
type BoundMutexMethod struct {
	Mutex      *MutexObject
	MethodName string
}

func (b *BoundMutexMethod) Type() ObjectType { return FUNCTION_OBJ }
func (b *BoundMutexMethod) Inspect() string  { return "mutex method " + b.MethodName }

// BoundAtomicMethod Atomic 方法绑定
type BoundAtomicMethod struct {
	Atomic     *AtomicObject
	MethodName string
}

func (b *BoundAtomicMethod) Type() ObjectType { return FUNCTION_OBJ }
func (b *BoundAtomicMethod) Inspect() string  { return "atomic method " + b.MethodName }

// ========== 方法调用实现 ==========

// evalChannelMethodCall 执行 Channel 方法调用
func (i *Interpreter) evalChannelMethodCall(ch *ChannelObject, methodName string, args []Object) Object {
	switch methodName {
	case "send":
		if len(args) != 1 {
			return newError("send() 需要1个参数")
		}
		if ch.IsClosed() {
			return newError("不能向已关闭的通道发送数据")
		}
		ch.Send(args[0])
		return &Null{}

	case "receive":
		value, ok := ch.Receive()
		if !ok {
			return &Null{}
		}
		return value

	case "tryReceive":
		value, ok := ch.TryReceive()
		if !ok || value == nil {
			return &Null{}
		}
		return value

	case "close":
		ch.Close()
		return &Null{}

	case "isClosed":
		return &Boolean{Value: ch.IsClosed()}

	case "isEmpty":
		return &Boolean{Value: ch.Len() == 0}

	case "len":
		return &Integer{Value: int64(ch.Len())}

	case "cap":
		return &Integer{Value: int64(ch.Cap())}

	case "forEach":
		if len(args) != 1 {
			return newError("forEach() 需要1个回调函数参数")
		}
		callback := args[0]
		
		// 遍历通道直到关闭
		for {
			value, ok := ch.Receive()
			if !ok {
				break
			}
			// 调用回调函数
			i.applyFunction(callback, []Object{value}, nil)
		}
		return &Null{}

	default:
		return newError("Channel 没有方法: %s", methodName)
	}
}

// evalWaitGroupMethodCall 执行 WaitGroup 方法调用
func (i *Interpreter) evalWaitGroupMethodCall(wg *WaitGroupObject, methodName string, args []Object) Object {
	switch methodName {
	case "add":
		if len(args) != 1 {
			return newError("add() 需要1个参数")
		}
		delta, ok := args[0].(*Integer)
		if !ok {
			return newError("add() 参数必须是整数")
		}
		wg.Add(int(delta.Value))
		return &Null{}

	case "done":
		wg.Done()
		return &Null{}

	case "wait":
		wg.Wait()
		return &Null{}

	default:
		return newError("WaitGroup 没有方法: %s", methodName)
	}
}

// evalMutexMethodCall 执行 Mutex 方法调用
func (i *Interpreter) evalMutexMethodCall(m *MutexObject, methodName string, args []Object) Object {
	switch methodName {
	case "lock":
		m.Lock()
		return &Null{}

	case "unlock":
		m.Unlock()
		return &Null{}

	case "tryLock":
		return &Boolean{Value: m.TryLock()}

	case "withLock":
		if len(args) != 1 {
			return newError("withLock() 需要1个回调函数参数")
		}
		callback := args[0]
		
		m.Lock()
		defer m.Unlock()
		
		return i.applyFunction(callback, []Object{}, nil)

	default:
		return newError("Mutex 没有方法: %s", methodName)
	}
}

// evalAtomicMethodCall 执行 Atomic 方法调用
func (i *Interpreter) evalAtomicMethodCall(a *AtomicObject, methodName string, args []Object) Object {
	switch methodName {
	case "get":
		return a.Get()

	case "set":
		if len(args) != 1 {
			return newError("set() 需要1个参数")
		}
		a.Set(args[0])
		return &Null{}

	case "add":
		if len(args) != 1 {
			return newError("add() 需要1个参数")
		}
		delta, ok := args[0].(*Integer)
		if !ok {
			return newError("add() 参数必须是整数")
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		if current, ok := a.value.(*Integer); ok {
			a.value = &Integer{Value: current.Value + delta.Value}
			return a.value
		}
		return newError("Atomic 值不是整数，无法使用 add()")

	case "increment":
		a.mu.Lock()
		defer a.mu.Unlock()
		if current, ok := a.value.(*Integer); ok {
			a.value = &Integer{Value: current.Value + 1}
			return a.value
		}
		return newError("Atomic 值不是整数，无法使用 increment()")

	case "decrement":
		a.mu.Lock()
		defer a.mu.Unlock()
		if current, ok := a.value.(*Integer); ok {
			a.value = &Integer{Value: current.Value - 1}
			return a.value
		}
		return newError("Atomic 值不是整数，无法使用 decrement()")

	case "compareAndSwap":
		if len(args) != 2 {
			return newError("compareAndSwap() 需要2个参数")
		}
		return &Boolean{Value: a.CompareAndSwap(args[0], args[1])}

	case "update":
		if len(args) != 1 {
			return newError("update() 需要1个回调函数参数")
		}
		callback := args[0]
		
		a.mu.Lock()
		defer a.mu.Unlock()
		
		result := i.applyFunction(callback, []Object{a.value}, nil)
		if isError(result) {
			return result
		}
		a.value = result
		return result

	default:
		return newError("Atomic 没有方法: %s", methodName)
	}
}

