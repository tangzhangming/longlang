# 接口

接口定义了一组方法签名，类可以通过 `implements` 关键字实现接口。

## 定义接口

使用 `interface` 关键字定义接口：

```longlang
interface Printable {
    function toString():string
}

interface Comparable {
    function compare(other:any):int
}
```

接口只包含方法签名，不包含实现。

## 实现接口

使用 `implements` 关键字让类实现接口：

```longlang
interface Flyable {
    function fly():string
    function getSpeed():int
}

class Bird implements Flyable {
    public speed int

    public function __construct(speed:int) {
        this.speed = speed
    }

    // 必须实现接口中定义的所有方法
    public function fly():string {
        return "鸟在飞翔"
    }

    public function getSpeed():int {
        return this.speed
    }
}
```

## 实现多个接口

一个类可以实现多个接口，用逗号分隔：

```longlang
interface Flyable {
    function fly():string
}

interface Swimmable {
    function swim():string
}

// 鸭子同时实现两个接口
class Duck implements Flyable, Swimmable {
    public function fly():string {
        return "鸭子在飞"
    }

    public function swim():string {
        return "鸭子在游泳"
    }
}
```

## 继承与接口

类可以同时继承父类和实现接口：

```longlang
interface Printable {
    function toString():string
}

class Entity {
    public id int

    public function __construct(id:int) {
        this.id = id
    }
}

// 同时继承和实现接口
class User extends Entity implements Printable {
    public name string

    public function __construct(id:int, name:string) {
        this.id = id
        this.name = name
    }

    public function toString():string {
        return "User[" + this.id + ", " + this.name + "]"
    }
}
```

## 接口方法签名

接口方法可以定义参数和返回类型：

```longlang
interface Calculator {
    function add(a:int, b:int):int
    function subtract(a:int, b:int):int
    function multiply(a:int, b:int):int
    function divide(a:int, b:int):float
}
```

## 接口契约

实现接口的类**必须**提供接口中定义的所有方法，否则会报错：

```longlang
interface Speakable {
    function speak():string
}

// ❌ 错误：没有实现 speak 方法
class Silent implements Speakable {
    // 缺少 speak() 方法的实现
}

// ✅ 正确：实现了所有接口方法
class Talker implements Speakable {
    public function speak():string {
        return "Hello!"
    }
}
```

## 完整示例

```longlang
package main

import "fmt"

// 定义接口
interface Drawable {
    function draw():string
}

interface Resizable {
    function resize(factor:int)
}

// 基类
class Shape {
    public name string

    public function __construct(name:string) {
        this.name = name
    }
}

// 实现单个接口
class Circle extends Shape implements Drawable {
    public radius int

    public function __construct(radius:int) {
        this.name = "圆形"
        this.radius = radius
    }

    public function draw():string {
        return "绘制圆形，半径: " + this.radius
    }
}

// 实现多个接口
class Rectangle extends Shape implements Drawable, Resizable {
    public width int
    public height int

    public function __construct(width:int, height:int) {
        this.name = "矩形"
        this.width = width
        this.height = height
    }

    public function draw():string {
        return "绘制矩形，宽: " + this.width + "，高: " + this.height
    }

    public function resize(factor:int) {
        this.width = this.width * factor
        this.height = this.height * factor
    }
}

fn main() {
    fmt.Println("=== 接口示例 ===")

    circle := new Circle(5)
    fmt.Println(circle.draw())

    fmt.Println("")

    rect := new Rectangle(10, 6)
    fmt.Println(rect.draw())
    
    rect.resize(2)
    fmt.Println("放大2倍后:")
    fmt.Println(rect.draw())
}
```

## 接口 vs 继承

| 特性 | 继承 | 接口 |
|------|------|------|
| 关键字 | `extends` | `implements` |
| 数量限制 | 只能单继承 | 可实现多个 |
| 内容 | 包含属性和方法实现 | 只有方法签名 |
| 用途 | 代码复用 | 定义契约/规范 |

## 相关文档

- [类基础](class-basics.md)
- [类继承](class-inheritance.md)






