# 类继承

LongLang 支持单继承机制，使用 `extends` 关键字实现。

## 基本语法

```longlang
class Parent {
    public name string
    
    public function greet():string {
        return "Hello from Parent"
    }
}

class Child extends Parent {
    public age int
    
    // 重写父类方法
    public function greet():string {
        return "Hello from Child"
    }
}
```

## 继承的特性

### 1. 属性继承

子类自动继承父类的所有公开和受保护属性：

```longlang
class Animal {
    public name string
    public age int
}

class Dog extends Animal {
    public breed string  // 子类新增的属性
}

// 使用
dog := new Dog()
dog.name = "旺财"  // 继承自 Animal
dog.age = 3        // 继承自 Animal
dog.breed = "金毛" // Dog 自己的属性
```

### 2. 方法继承

子类自动继承父类的所有公开和受保护方法：

```longlang
class Animal {
    public function speak():string {
        return "动物叫声"
    }
    
    public function info():string {
        return "这是一个动物"
    }
}

class Dog extends Animal {
    // Dog 自动拥有 speak() 和 info() 方法
}

dog := new Dog()
fmt.Println(dog.info())  // 输出: 这是一个动物
```

### 3. 方法重写

子类可以重写父类的方法：

```longlang
class Animal {
    public function speak():string {
        return "动物叫声"
    }
}

class Dog extends Animal {
    public function speak():string {
        return "汪汪汪!"  // 重写父类方法
    }
}

class Cat extends Animal {
    public function speak():string {
        return "喵喵喵!"  // 重写父类方法
    }
}

dog := new Dog()
cat := new Cat()
fmt.Println(dog.speak())  // 汪汪汪!
fmt.Println(cat.speak())  // 喵喵喵!
```

## 多层继承

支持多层继承链：

```longlang
class A {
    public function methodA():string {
        return "A"
    }
}

class B extends A {
    public function methodB():string {
        return "B"
    }
}

class C extends B {
    public function methodC():string {
        return "C"
    }
}

// C 同时拥有 methodA、methodB、methodC
c := new C()
fmt.Println(c.methodA())  // A
fmt.Println(c.methodB())  // B
fmt.Println(c.methodC())  // C
```

## 构造函数与继承

子类构造函数需要手动初始化父类属性：

```longlang
class Animal {
    public name string
    
    public function __construct(name:string) {
        this.name = name
    }
}

class Dog extends Animal {
    public breed string
    
    public function __construct(name:string, breed:string) {
        this.name = name    // 初始化父类属性
        this.breed = breed  // 初始化子类属性
    }
}
```

## super 关键字

在子类中可以使用 `super` 关键字引用父类：

```longlang
class Parent {
    public static function helper():string {
        return "Parent helper"
    }
}

class Child extends Parent {
    public function useParent():string {
        // super 指向父类
        return "Using parent"
    }
}
```

## 完整示例

```longlang
package main

import "fmt"

// 基类：形状
class Shape {
    public name string

    public function __construct(name:string) {
        this.name = name
    }

    public function describe():string {
        return "这是一个" + this.name
    }

    public function area():int {
        return 0
    }
}

// 子类：矩形
class Rectangle extends Shape {
    public width int
    public height int

    public function __construct(width:int, height:int) {
        this.name = "矩形"
        this.width = width
        this.height = height
    }

    public function area():int {
        return this.width * this.height
    }

    public function perimeter():int {
        return 2 * (this.width + this.height)
    }
}

// 子类：正方形（继承自矩形）
class Square extends Rectangle {
    public function __construct(side:int) {
        this.name = "正方形"
        this.width = side
        this.height = side
    }
}

fn main() {
    rect := new Rectangle(5, 3)
    fmt.Println(rect.describe())
    fmt.Println("面积: " + rect.area())
    fmt.Println("周长: " + rect.perimeter())

    fmt.Println("")

    square := new Square(4)
    fmt.Println(square.describe())
    fmt.Println("面积: " + square.area())
    fmt.Println("周长: " + square.perimeter())
}
```

## 相关文档

- [类基础](class-basics.md)
- [接口](class-interface.md)










