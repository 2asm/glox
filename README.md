# GLOX
The goal of this project was to create a simple programming language and learn about bytecode-compiler/virtual-machine. It is heavily inspired by Robert Nystrom's book "crafting interpreters" and implements features from chapter 14 to 24. But I made a few changes, which are listed below.

## changes
- there are two type of numbers float64 and int64
- few keywords are different `var` is `let` and `fun` is `fn`
- new variable can't be declared without an initial value
```
   let num = 5; 
```
- Assign(=) is not an operator, it's a statement
- operator precedence order is copied from Golang
- supports all the prefix(~, +, -) and infix(+, -, *, /, %, &, |, ^, <<, >>, ==, !=, <, <=, >, >=, ||, &&) operators
- supports all the `Op=` type statements like +=, -=, *=, /= etc. 
- {} after `if cond` and `while cond` is must, similar to Golang and parenthesis around `cond` is not necessary
```
if 1<2 {
    print "true";
} else {
    print "false";
}
```
- `print` statement can take multiple arguments
` print "hello", "world"; `
- no support for `for` loop because `while` can do it all
- added support for `break` statement
- doesn't follow the exact same implementation details from the book
- no support for string interning
- no jump in logical expressions

## sample code
```
fn fact(n) {
    if n <= 1 {
        return 1;
    }
    return n * fact(n-1);
}

let n = 6;
print "factorial of", n, "is", fact(n);

let i = 0;
let sm = 0;
while i <= 1_000_000 {
    i += 1;
    sm += i;
}
print sm;
```

## try it out
``` Console
$ git clone github.com/2asm/glox
$ cd glox
$ go build .
$ ./glox test.glox
```

## Useful resources
https://craftinginterpreters.com/ <br>
https://interpreterbook.com/ <br>
https://matklad.github.io/2020/04/13/simple-but-powerful-pratt-parsing.html <br>
https://en.wikipedia.org/wiki/Operator-precedence_parser <br>
https://github.com/golang/go/tree/master/src/go <br>

