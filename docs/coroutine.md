# LongLang 协程

LongLang 提供了类似 Go 语言的协程支持，让你可以轻松编写并发程序。

## 启动协程

使用 `go` 关键字启动一个新的协程：

```longlang
// 方式 1：定义闭包后启动
task := fn() {
    fmt.println("在协程中执行")
}
go task()

// 方式 2：直接启动命名函数
go myFunction()
```

## Channel 通信

Channel 是协程间通信的主要方式：

```longlang
// 创建无缓冲 Channel
ch := new Channel()

// 创建带缓冲 Channel（容量为 10）
bufCh := new Channel(10)

// 发送数据
ch.send("Hello")

// 接收数据（阻塞直到有数据）
msg := ch.receive()

// 尝试接收（非阻塞，如果没有数据返回 null）
msg2 := ch.tryReceive()

// 关闭 Channel
ch.close()

// 检查是否已关闭
if ch.isClosed() {
    fmt.println("Channel 已关闭")
}

// 获取 Channel 长度和容量
fmt.println("长度: " + toString(ch.len()))
fmt.println("容量: " + toString(ch.cap()))

// 遍历 Channel（直到关闭）
ch.forEach(fn(value: any) {
    fmt.println("收到: " + toString(value))
})
```

### Channel 方法

| 方法 | 说明 |
|------|------|
| `send(value)` | 发送数据到 Channel |
| `receive()` | 接收数据（阻塞） |
| `tryReceive()` | 尝试接收（非阻塞） |
| `close()` | 关闭 Channel |
| `isClosed()` | 检查是否已关闭 |
| `isEmpty()` | 检查是否为空 |
| `len()` | 获取当前元素数量 |
| `cap()` | 获取容量 |
| `forEach(callback)` | 遍历所有元素直到关闭 |

## WaitGroup

WaitGroup 用于等待一组协程完成：

```longlang
wg := new WaitGroup()

// 添加等待计数
wg.add(3)

for i := 0; i < 3; i++ {
    task := fn() {
        // 执行任务...
        wg.done()  // 完成时调用
    }
    go task()
}

// 等待所有协程完成
wg.wait()
fmt.println("所有任务已完成")
```

### WaitGroup 方法

| 方法 | 说明 |
|------|------|
| `add(delta)` | 添加等待计数 |
| `done()` | 完成一个任务（计数减1） |
| `wait()` | 等待所有任务完成 |

## Mutex 互斥锁

Mutex 用于保护共享资源：

```longlang
mutex := new Mutex()

// 加锁
mutex.lock()
// 操作共享资源...
mutex.unlock()

// 尝试加锁（非阻塞）
if mutex.tryLock() {
    // 成功获取锁
    mutex.unlock()
}

// 使用 withLock 自动管理锁
mutex.withLock(fn() {
    // 在锁保护下执行
    // 自动释放锁
})
```

### Mutex 方法

| 方法 | 说明 |
|------|------|
| `lock()` | 加锁（阻塞） |
| `unlock()` | 解锁 |
| `tryLock()` | 尝试加锁（非阻塞） |
| `withLock(callback)` | 在锁保护下执行回调 |

## Atomic 原子操作

Atomic 提供协程安全的数据共享：

```longlang
// 创建原子整数
counter := new Atomic(0)

// 获取值
value := counter.get()

// 设置值
counter.set(100)

// 原子加
counter.add(10)

// 原子增减
counter.increment()
counter.decrement()

// 比较并交换
if counter.compareAndSwap(100, 200) {
    fmt.println("交换成功")
}

// 使用回调更新值
counter.update(fn(current: any) {
    return current * 2
})
```

### Atomic 方法

| 方法 | 说明 |
|------|------|
| `get()` | 获取当前值 |
| `set(value)` | 设置新值 |
| `add(delta)` | 原子加（仅整数） |
| `increment()` | 原子加1（仅整数） |
| `decrement()` | 原子减1（仅整数） |
| `compareAndSwap(expected, newValue)` | 比较并交换 |
| `update(callback)` | 使用回调更新值 |

## sleep 函数

`sleep` 函数用于休眠当前协程：

```longlang
// 休眠 100 毫秒
sleep(100)

// 休眠 1 秒
sleep(1000)
```

## 完整示例

### 生产者-消费者模式

```longlang
namespace Example

class Main {
    public static function main() {
        ch := new Channel(10)
        wg := new WaitGroup()
        
        // 生产者
        wg.add(1)
        producer := fn() {
            for i := 0; i < 5; i++ {
                ch.send("消息 " + toString(i))
                sleep(10)
            }
            ch.close()
            wg.done()
        }
        go producer()
        
        // 消费者
        wg.add(1)
        consumer := fn() {
            ch.forEach(fn(msg: any) {
                fmt.println("收到: " + toString(msg))
            })
            wg.done()
        }
        go consumer()
        
        wg.wait()
        fmt.println("完成")
    }
}
```

### 并发计数器

```longlang
namespace Example

class Main {
    public static function main() {
        counter := new Atomic(0)
        wg := new WaitGroup()
        
        wg.add(10)
        for i := 0; i < 10; i++ {
            task := fn() {
                for j := 0; j < 100; j++ {
                    counter.increment()
                }
                wg.done()
            }
            go task()
        }
        
        wg.wait()
        fmt.println("最终值: " + toString(counter.get()))
        // 输出: 最终值: 1000
    }
}
```

### 工作池模式

```longlang
namespace Example

class Main {
    public static function main() {
        jobs := new Channel(100)
        results := new Channel(100)
        wg := new WaitGroup()
        
        // 启动 3 个 worker
        wg.add(3)
        for w := 0; w < 3; w++ {
            workerId := w
            worker := fn() {
                jobs.forEach(fn(job: any) {
                    // 处理任务
                    result := "Worker " + toString(workerId) + " 处理了任务 " + toString(job)
                    results.send(result)
                })
                wg.done()
            }
            go worker()
        }
        
        // 发送任务
        for j := 0; j < 9; j++ {
            jobs.send(j)
        }
        jobs.close()
        
        // 收集结果
        collector := fn() {
            wg.wait()
            results.close()
        }
        go collector()
        
        results.forEach(fn(result: any) {
            fmt.println(toString(result))
        })
        
        fmt.println("所有任务完成")
    }
}
```

## 注意事项

1. **闭包变量捕获**：在循环中创建协程时，注意变量捕获问题。建议在循环内创建新变量来捕获当前值。

2. **共享数据**：协程之间不共享普通变量。使用 `Atomic` 或通过 `Channel` 传递数据来实现协程间通信。

3. **避免死锁**：
   - 使用带缓冲的 Channel 可以减少死锁风险
   - 确保 `lock()` 和 `unlock()` 配对使用
   - 使用 `withLock()` 自动管理锁的生命周期

4. **资源释放**：
   - 完成后关闭 Channel
   - 使用 WaitGroup 确保所有协程完成
   - 使用 defer 或 withLock 确保锁被释放

5. **调试**：协程执行顺序不确定，调试并发问题时可以使用 `sleep()` 来控制时序。










