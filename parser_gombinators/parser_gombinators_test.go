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

package parser_gombinators

import (
  "testing"
  "container/list"
)

func testExpectedCharacterInInput (t *testing.T,
                                   codePoint rune, input ParserInput) {
  if input == nil {
    t.Errorf ("Expected %c but the input is nil!\n", codePoint)
  } else if codePoint != input.CurrentCodePoint () {
    t.Errorf ("Expected %c, got %c!\n", codePoint, input.CurrentCodePoint ())
  }
}

func TestRuneArrayInput (t *testing.T) {
  var input = StringToInput ("大熊猫")
  testExpectedCharacterInInput (t, '大', input)
  input = input.RemainingInput ()
  var input2 = input
  testExpectedCharacterInInput (t, '熊', input)
  input = input.RemainingInput ()
  testExpectedCharacterInInput (t, '猫', input)
  testExpectedCharacterInInput (t, '熊', input2)
  input = input.RemainingInput ()
  testExpectedCharacterInInput (t, '\x00', input)
  input = input.RemainingInput ()
  if input != nil {
    t.Errorf (
      "Expected the remaining input to be nil after the end of the input!")
  }
}

func TestExpectCodePoints (t *testing.T) {
  var input = StringToInput ("大熊猫")
  var result = ExpectCodePoints ([]rune ("大熊")) (input)
  if result.RemainingInput.CurrentCodePoint () != '猫' {
    t.Errorf ("Expected the remaining input be \"猫\"!")
  }
  result = ExpectCodePoints ([]rune ("大熊猫ABC")) (input)
  if result.Result != nil {
    t.Errorf ("Expected failed parse!")
  }
  result = ExpectCodePoints ([]rune ("ABC")) (input)
  if result.Result != nil {
    t.Errorf ("Expected failed parse!")
  }

}

func TestRepeated (t *testing.T) {
  var parser = ExpectCodePoint (rune ('A')).Repeated ()
  var input = StringToInput ("AAABCD")
  var result = parser (input)
  if result.Result.(*list.List).Len () != 3 ||
     result.RemainingInput.CurrentCodePoint () != rune ('B') {
     t.Errorf ("Expected the parser to parse the As until the character B.")
  }
  input = StringToInput ("B")
  result = parser (input)
  if result.Result.(*list.List).Len () != 0 ||
     result.RemainingInput.CurrentCodePoint () != rune ('B') {
    t.Errorf ("Expected the parser to successfully parse nothing!")
  }
  input = StringToInput ("")
  result = parser (input)
  if result.Result.(*list.List).Len () != 0 ||
     result.RemainingInput.CurrentCodePoint () != rune ('\x00') {
    t.Errorf ("Expected the parser to successfully parse nothing!")
  }
}

func TestOnceOrMore (t *testing.T) {
  var parser = ExpectCodePoint (rune ('A')).OnceOrMore ()
  var input = StringToInput ("AAABCD")
  var result = parser (input)
  if result.Result.(*list.List).Len () != 3 ||
     result.RemainingInput.CurrentCodePoint () != rune ('B') {
     t.Errorf ("Expected the parser to parse the As until the character B.")
  }
  input = StringToInput ("B")
  result = parser (input)
  if result.Result != nil {
    t.Errorf ("Expected the parser to fail!")
  }
  input = StringToInput ("")
  result = parser (input)
  if result.Result != nil {
    t.Errorf ("Expected the parser to fail!")
  }
}

func TestOrElse (t *testing.T) {
  var parser = ExpectString ("A").OrElse (ExpectString ("B"))
  var input = StringToInput ("AC")
  var result = parser (input)
  if result.Result != "A" ||
    result.RemainingInput.CurrentCodePoint () != 'C' {
    t.Errorf ("Expected the parser to parse the First code point!")
  }

  input = StringToInput ("BC")
  result = parser (input)
  if result.Result != "B" ||
    result.RemainingInput.CurrentCodePoint () != 'C' {
    t.Errorf ("Expected the parser to parse the First code point! %s", result)
  }

}

func TestAndThen (t *testing.T) {
  var parser = ExpectString ("A").AndThen (ExpectString ("B"))
  var input = StringToInput ("A")
  var result = parser (input)
  if result.Result != nil {
    t.Errorf ("Expected the parser to fail!")
  }
  input = StringToInput ("AB")
  result = parser (input)
  if result.Result.(Pair).First != "A" ||
    result.Result.(Pair).Second != "B" ||
    result.RemainingInput.CurrentCodePoint () != rune ('\x00') {
    t.Errorf ("Expected the parser to parse (A, B)!")
  }
  input = StringToInput ("ABC")
  result = parser (input)
  if result.Result.(Pair).First != "A" ||
    result.Result.(Pair).Second != "B" ||
    result.RemainingInput.CurrentCodePoint () != rune ('C') {
    t.Errorf ("Expected the parser to parse (A, B)!")
  }
}

func TestConvertResult (t *testing.T) {
  var parser = ExpectString ("A").Convert (
    func (result interface{}) interface{} {
      return 42
    })
  var input = StringToInput ("AB")
  var result = parser (input)
  if result.Result != 42 {
    t.Errorf ("Expected the result to be 42!")
  }
}

func TestExpectSpaces (t *testing.T) {
  var parser = ExpectSpaces
  var input = StringToInput ("  \nABC")
  var result = parser (input)
  if result.Result != "  \n" ||
      result.RemainingInput.CurrentCodePoint () != rune ('A') {
    t.Errorf ("Expected the parser to parse the whitespace!")
  }
  input = StringToInput ("ABC")
  result = parser (input)
  if result.Result != (Nothing{}) ||
      result.RemainingInput.CurrentCodePoint () != rune ('A') {
    t.Errorf ("Expected the parser to parse nothing!")
  }

}

func TestExpectIdentifier (t *testing.T) {
  var parser = ExpectIdentifier
  var input = StringToInput ("b1_la blub")
  var result = parser (input)
  if result.Result != "b1_la" || result.RemainingInput.CurrentCodePoint () != rune (' ') {
    t.Errorf ("Expected the parser to read the First identifier!")
  }
  input = StringToInput ("1b_la blub")
  result = parser (input)
  if result.Result != nil || result.RemainingInput.CurrentCodePoint () != rune ('1') {
    t.Errorf ("Expected the parser to fail!")
  }
}

func TestRepeatAndFoldLeft (t *testing.T) {
  var parser = ExpectString ("a").RepeatAndFoldLeft ("a",
        func (lhs interface{}, rhs interface{}) interface{} {
          return "(" + lhs.(string) + " " + rhs.(string) + ")"
        })
  var input = StringToInput ("aaabcde")
  var result = parser (input)
  if result.Result != "(((a a) a) a)" ||
    result.RemainingInput.CurrentCodePoint () != rune ('b') {
    t.Errorf ("Expected the parser to parse the four a chars!")
  }
  input = StringToInput ("abcde")
  result = parser (input)
  if result.Result != "(a a)" ||
    result.RemainingInput.CurrentCodePoint () != rune ('b') {
    t.Errorf ("Expected the parser to parse the first char!")
  }
  input = StringToInput ("bcde")
  result = parser (input)
  if result.Result != "a" ||
    result.RemainingInput.CurrentCodePoint () != rune ('b') {
    t.Errorf ("Expected the parser to return the accumulator!")
  }
}

func TestBind (t *testing.T) {
  var parser = ExpectIdentifier.Bind (func (arg interface{}) Parser {
      return MaybeSpacesBefore (ExpectString (arg.(string)))
    })
  var input = StringToInput ("ning ning")
  var result = parser (input)
  if result.Result != "ning" ||
    result.RemainingInput.CurrentCodePoint () != '\x00' {
    t.Errorf ("Expected the parser to eat up the whole input!")
  }
}
