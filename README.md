# Parser-Gombinators
Simple Parser Combinators in Go

This library implements simple parser combinators in the Go programming
language. Parser combinators allow you to parse texts of many 
deterministic context-free languages. Parser combinators are designed
to make the parser code mimic the grammar.

**Avoid left-recursion!**

**Avoid overlapping prefixes in alternatives!**

The following grammar for primary school arithmetic expression
satisfies the above two constraints.

```
Multiplicand := Number
              | "(" Expression ")"
Adddend      := Multiplicand (("*" | "/") Multiplicand)*
Expression   := Addend       (("+" | "-") Addend)*
```

Using Parser-Gombinators the source code of the parser reads almost like
the grammar itself.

```go
func Multiplicand (input ParserInput) ParserResult {
  return ExpectNumber.Convert(atoi).OrElse (
      expect ("(").AndThen (Expression).AndThen (expect (")")).
        First().Second()) (input)
}

func Addend (input ParserInput) ParserResult {
  return Parser (Multiplicand).Bind (func (firstResult interface{}) Parser {
      return expect ("*").OrElse (expect ("/")).AndThen (Multiplicand).
        RepeatAndFoldLeft (firstResult, multiply)
    }) (input)
}

func Expression (input ParserInput) ParserResult {
  return Parser (Addend).Bind (func (firstResult interface{}) Parser {
      return expect ("+").OrElse (expect ("-")).AndThen (Addend).
        RepeatAndFoldLeft (firstResult, add)
    }) (input)
}
```

See the calculator example for the full source code.
