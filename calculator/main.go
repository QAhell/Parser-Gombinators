/*
    © 2018 Armin Heller

    This file is part of Parser-Gombinators.

    Parser-Gombinators is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Parser-Gombinators is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Parser-Gombinators. If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
  "os"
  "fmt"
  . "github.com/QAhell/Parser-Gombinators/parse"
  . "strconv"
)

/*
  Primary school arithmetic, left-associative, only binary operators

  Digit        := "0" | .. | "9"
  Number       := Digit Digit*
  Multiplicand := Number
                | "(" Expression ")"
  Adddend      := Multiplicand (("*" | "/") Multiplicand)*
  Expression   := Addend       (("+" | "-") Addend)*

  Important: No left-recursion!
  Important: No overlapping rules/alternatives!

 */

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

var licence_notice = "Parsing-Gombinators: An Example Calculator.\n" +
    "  Evaluate primary school arithmetic expressions.\n" +
    "  Usage: calculator 'expression'\n\n" +
    "Copyright (C) 2018  Armin Heller\n\n" +
    "This program is free software: you can redistribute it and/or modify\n" +
    "it under the terms of the GNU General Public License as published by\n" +
    "the Free Software Foundation, either version 3 of the License, or\n" +
    "(at your option) any later version.\n\n" +
    "This program is distributed in the hope that it will be useful,\n" +
    "but WITHOUT ANY WARRANTY; without even the implied warranty of\n" +
    "MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the\n" +
    "GNU General Public License for more details.\n\n" +
    "You should have received a copy of the GNU General Public License\n" +
    "along with this program.  If not, see <https://www.gnu.org/licenses/>.\n\n"

func main () {
  if len (os.Args) != 2 {
    fmt.Printf (licence_notice)
  } else {
    var input = StringToInput (os.Args[1])
    var parserResult = Expression (input)
    var result, isInteger = parserResult.Result.(int)
    if isInteger {
      fmt.Printf ("result = %d\n", result)
      if nil != parserResult.RemainingInput {
        var inp = parserResult.RemainingInput.(RuneArrayInput)
        var rest = inp.Text[inp.CurrentPosition:]
        fmt.Printf ("There's some remaining input: %s\n", string (rest))
      }
    } else {
      fmt.Printf ("Couldn't read the input!\n")
    }
  }
}

func expect (text string) Parser {
  return MaybeSpacesBefore (ExpectString (text))
}

func multiply (lhs interface{}, rhs interface{}) interface{} {
  if "*" == GetFirst (rhs).(string) {
    return lhs.(int) * GetSecond (rhs).(int)
  }
  return lhs.(int) / GetSecond (rhs).(int)
}

func add (lhs interface{}, rhs interface{}) interface{} {
  if "+" == GetFirst (rhs).(string) {
    return lhs.(int) + GetSecond (rhs).(int)
  }
  return lhs.(int) - GetSecond (rhs).(int)
}

func atoi (arg interface{}) interface{} {
  var result, _ = Atoi(arg.(string))
  return result
}
