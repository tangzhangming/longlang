# return 语句

`return` 语句用于从函数或方法中返回值，并结束当前函数/方法的执行。

## 基本用法

### 1. 返回单个值

```longlang
fn add(a:int, b:int) int {
    return a + b
}
```

### 2. 无返回值

```longlang
fn sayHello(name:string) {
    fmt.Println("Hello,", name)
    return  // 可以省略，但显式 return 可用于提前结束函数
}
```

### 3. 条件返回

```longlang
fn abs(x:int) int {
    if x < 0 {
        return -x
    }
    return x
}
```

## 在类方法中使用

### 实例方法

```longlang
class Calculator {
    private value int
    
    public function getValue() int {
        return this.value
    }
    
    public function isPositive() bool {
        return this.value > 0
    }
}
```

### 静态方法

```longlang
class MathUtils {
    public static function max(a:int, b:int) int {
        if a > b {
            return a
        }
        return b
    }
}
```

## 提前退出

`return` 可以在函数任意位置使用，用于提前结束函数执行：

```longlang
fn processData(data:string) {
    if data == "" {
        return  // 数据为空，提前退出
    }
    
    // 处理数据...
    fmt.Println("Processing:", data)
}
```

## 在循环中使用

`return` 会直接结束整个函数，而不仅仅是跳出循环：

```longlang
fn findFirst(target:int) int {
    for i := 0; i < 100; i++ {
        if i == target {
            return i  // 找到后直接返回，结束函数
        }
    }
    return -1  // 未找到
}
```

## 注意事项

1. `return` 会立即结束当前函数/方法的执行
2. 如果函数声明了返回类型，`return` 必须返回相应类型的值
3. 在循环中使用 `return` 会直接结束整个函数，而不仅仅是循环
4. 可以使用不带值的 `return` 来提前结束无返回值的函数

