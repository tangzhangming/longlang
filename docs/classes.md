# 类与面向对象

LongLang 支持面向对象编程，包括类的定义、实例化、继承等特性。

## 类的定义

使用 `class` 关键字定义类：

```longlang
class Person {
    // 成员变量
    public name string
    public age int

    // 构造函数
    public function __construct(name:string, age:int) {
        this.name = name
        this.age = age
    }

    // 实例方法
    public function greet():string {
        return "你好，我是 " + this.name
    }
}
```

## 访问修饰符

| 修饰符 | 说明 |
|-------|------|
| `public` | 公开成员，可以从类外部访问 |
| `private` | 私有成员，只能在类内部访问 |
| `protected` | 受保护成员，只能在类及其子类中访问 |

## 实例化

使用 `new` 关键字创建类的实例：

```longlang
person := new Person("张三", 25)
fmt.Println(person.name)    // 张三
fmt.Println(person.greet()) // 你好，我是 张三
```

## this 关键字

在类方法中，使用 `this` 关键字引用当前实例：

```longlang
public function setName(name:string) {
    this.name = name
}
```

## 静态方法

使用 `static` 关键字定义静态方法：

```longlang
class MathUtil {
    public static function add(a:int, b:int):int {
        return a + b
    }
}

// 调用静态方法
result := MathUtil::add(1, 2)
```

## 继承

使用 `extends` 关键字实现类的继承：

```longlang
// 父类
class Animal {
    public name string
    
    public function __construct(name:string) {
        this.name = name
    }
    
    public function speak():string {
        return "动物叫声"
    }
}

// 子类
class Dog extends Animal {
    public breed string
    
    public function __construct(name:string, breed:string) {
        this.name = name
        this.breed = breed
    }
    
    // 重写父类方法
    public function speak():string {
        return "汪汪汪!"
    }
}
```

### 继承的特性

1. **方法继承**：子类自动继承父类的所有公开和受保护方法
2. **属性继承**：子类自动继承父类的所有公开和受保护属性
3. **方法重写**：子类可以重写父类的方法
4. **多层继承**：支持多层继承链（A → B → C）

### 继承示例

```longlang
// 基类
class Shape {
    public name string
    
    public function describe():string {
        return "这是一个" + this.name
    }
}

// 二维形状
class Shape2D extends Shape {
    public function area():int {
        return 0
    }
}

// 矩形
class Rectangle extends Shape2D {
    public width int
    public height int
    
    public function __construct(w:int, h:int) {
        this.name = "矩形"
        this.width = w
        this.height = h
    }
    
    public function area():int {
        return this.width * this.height
    }
}

// 使用
rect := new Rectangle(5, 3)
fmt.Println(rect.describe())  // 继承自 Shape
fmt.Println(rect.area())      // 重写的方法，返回 15
```

## super 关键字

在子类中，可以使用 `super` 关键字引用父类：

```longlang
class Child extends Parent {
    public function method() {
        // super 指向父类
        // 可用于调用父类的静态方法
    }
}
```

## 构造函数

构造函数使用特殊名称 `__construct`：

```longlang
class Person {
    public name string
    
    public function __construct(name:string) {
        this.name = name
    }
}
```

### 构造函数继承

如果子类没有定义构造函数，会自动使用父类的构造函数（如果存在）。

## self 关键字

在类方法中，使用 `self` 关键字引用当前类（用于调用静态方法）：

```longlang
class Calculator {
    public static function double(n:int):int {
        return n * 2
    }
    
    public function quadruple(n:int):int {
        return self::double(self::double(n))
    }
}
```

## 完整示例

```longlang
package main

import "fmt"

class Animal {
    public name string
    public age int

    public function __construct(name:string, age:int) {
        this.name = name
        this.age = age
    }

    public function speak():string {
        return "..."
    }

    public function info():string {
        return this.name + ", " + this.age + "岁"
    }
}

class Dog extends Animal {
    public breed string

    public function __construct(name:string, age:int, breed:string) {
        this.name = name
        this.age = age
        this.breed = breed
    }

    public function speak():string {
        return "汪汪汪!"
    }
}

fn main() {
    dog := new Dog("旺财", 3, "金毛")
    fmt.Println(dog.info())   // 旺财, 3岁
    fmt.Println(dog.speak())  // 汪汪汪!
}
```

