# for 循环

LongLang 的 for 循环语法与 Go 语言完全一致，支持三种形式：

## 1. while 式循环

使用 `for condition { ... }` 形式，相当于其他语言的 `while(condition)`：

```longlang
i := 0
for i < 5 {
    fmt.Println("i =", i)
    i++
}
```

## 2. 无限循环

使用 `for { ... }` 形式，相当于 `while(true)`：

```longlang
count := 0
for {
    fmt.Println("无限循环", count)
    count++
    if count >= 3 {
        break
    }
}
```

## 3. 传统 for 循环

使用 `for init; condition; post { ... }` 形式：

```longlang
for j := 0; j < 5; j++ {
    fmt.Println("j =", j)
}
```

## break 和 continue

### break

`break` 语句用于立即跳出当前循环：

```longlang
for i := 0; i < 10; i++ {
    if i == 5 {
        break  // 当 i 等于 5 时跳出循环
    }
    fmt.Println(i)
}
// 输出: 0 1 2 3 4
```

### continue

`continue` 语句用于跳过当前迭代，继续下一次迭代：

```longlang
for i := 0; i < 5; i++ {
    if i == 2 {
        continue  // 跳过 i 等于 2 的情况
    }
    fmt.Println(i)
}
// 输出: 0 1 3 4
```

## 自增/自减运算符

支持 `++` 和 `--` 运算符：

```longlang
i := 0
i++  // i 变为 1
i--  // i 变回 0
```

## 注意事项

1. for 循环的条件表达式可以省略，省略时表示无限循环
2. 初始化语句和 post 语句也可以省略
3. break 只能跳出最内层的循环
4. continue 只能跳过最内层循环的当前迭代

