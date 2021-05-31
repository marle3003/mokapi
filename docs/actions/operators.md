# Operators
Mokapi Actions provides a number of operators allowing you to perform basic operations with values.

- [Arithmetic operators](#arithmetic)
- [Comparison operators](#comparison)
- [Equality operators](#equality)
- [Member access operators](#member)
- [Lambda operator](#lambda)
- [Range operator](#range)
- [Operator precedence](#precedence)

## Unary plus and minus operators <span id="arithmetic"></span>
The unary *+* operator returns the value of its operand. The unary *-* operator computes the numeric negation of its operand.

```yaml
env:
  myVar: ${{ +2 }}
  myOtherVar: ${{ -2 }}
```

## Multiplication operator *
The multiplication operator *\** computes the product of its operands

```yaml
env:
  myVar: ${{ 2 * 3 }}
```

## Division operator /
The division operator */* divides its left-hand operator by its right-hand operand.

### Integer division
For the operands of integer types, the result of the */* operator is of an integer type and equals the quotient of the two operands rounded towards zero.

```yaml
env:
  myVar: ${{ 9 / 5 }}
```
In the above example the value of the environment variable *myVar* will be *1*

### Floating-point division
If one of the operands is a floating number, the result of the division operator */* will be a floating number.

```yaml
env:
  myVar: ${{ 7.4 / 5 }}
  myOtherVar: ${{ 7 / 5.0 }}
```

## Remainder operator %
The remainder operator *%* computes the remainder after dividing its integer left-hand operand by its integer right-hand operand. Floating operands are not supported.

```yaml
env:
  myVar: ${{ 7 / 5 }}
```

## Adding operator +
The addition operator *+* computes the sum of its operands.

```yaml
env:
  myVar: ${{ 3 + 9 }}
```

## Substraction operator -
The substraction operator *-* substracts its right-hand operand from its left-hand operand

```yaml
env:
  myVar: ${{ 47 - 5 }}
```

## Less than operator < <span id="comparison"></span>
The *<* operator returns *true* if its left-hand operand is less than its right-hand operand, false otherwise.

```yaml
env:
  myVar: ${{ 47 < 5 }}
```

## Greater than operator >
The *>* operator returns *true* if its left-hand operand is greater than its right-hand operand, false otherwise

```yaml
env:
  myVar: ${{ 47 > 5 }}
```

## Less than or equal operator <=
The *<=* operator returns *true* if its left-hand operand is less than or equal to its right-hand operand, false otherwise

```yaml
env:
  myVar: ${{ 47 <= 5 }}
```

## Greater than or equal operator >=
The *>=* operator returns *true* if its left-hand operand is greater than or equal to its right-hand operand, false otherwise

```yaml
env:
  myVar: ${{ 47 >= 5 }}
```

## Equality operator == <span id="equality"></span>
The equality operator *==* returns *true* if its operands are equal, *false* otherwise.

```yaml
env:
  myVar: ${{ 47 == 5 }}
```

## Inequality operator !=
The inequality operator *!=* returns *true* if its operands are not equal, *false* otherwise.

```yaml
env:
  myVar: ${{ 47 != 5 }}
```

## Member access expression . <span id="member"></span>
Use *.* to access type members

```yaml
env:
  myVar: ${{ env.otherValue }}
```

## Index operator []
Square brackets *[]* are used for array or indexer

### Index access
```yaml
env:
  myVar: ${{ array[1] }}
```

### Array
```yaml
env:
  myVar: ${{ find([1,2,3,4], x => x == 3) }}
```

## Invocation expression ()
Use parentheses *()* to call a function.  You also use parentheses to adjust the order in which to evaluate operations in an expression.

```yaml
env:
  myVar: ${{ find([1,2,3,4], x => x == 3) }}
```

## Lambda expression <span id="lambda"></span>
You use a lambda expression to create an anonymous function. Use the operator *=>* to separate the lambda's parameter list from its body

```yaml
env:
  myVar: ${{ find([1,2,3,4], x => x == 3) }}
```

## Range operator .. <span id="range"></span>
Range operator allows you to create a list of sequential values.
```yaml
env:
  myVar: ${{ [1..4] }}
```

## Operator precedence <span id="precedence"></span>
In an expression with multiple operators, the operators with higher precedence are evaluated before the operators with lower precedence. The following list lists the operators starting with the highest precedence to the lowest. The operators within each row have the same precedence. 

1. x.y, f(x), x\[y]
2. +x, -x, !x
3. x..y
4. x * y, x / y, x % y
5. x + y, x - y
6. x == y, x != y
7. x && y   
8. x || y

When operators have the same precedence, they are evaluated in order from left to right. Use parentheses to change the order of evaluation.

```yaml
env:
  myVar: ${{ (1 + 2) * 3 }}
```