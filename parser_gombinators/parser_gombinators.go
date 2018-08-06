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
  "strings"
  "container/list"
)

type Parser func (ParserInput) ParserResult

type ParserInput interface {
  CurrentCodePoint () rune
  RemainingInput () ParserInput
}

type ParserResult struct {
  Result interface{} // Parsing failed <==> result == nil
  RemainingInput ParserInput
}

func ExpectCodePoint (expectedCodePoint rune) Parser {
  return func (input ParserInput) ParserResult {
    if expectedCodePoint == input.CurrentCodePoint () {
      return ParserResult { expectedCodePoint, input.RemainingInput () }
    }
    return ParserResult { nil, input }
  }
}

func ExpectCodePoints (expectedCodePoints []rune) Parser {
  return func (input ParserInput) ParserResult {
    var RemainingInput = input
    for _, expectedCodePoint := range expectedCodePoints {
      if nil == RemainingInput.RemainingInput () {
        return ParserResult { nil, RemainingInput }
      }
      var result = ExpectCodePoint (expectedCodePoint) (RemainingInput)
      if result.Result == nil {
        return ParserResult { nil, RemainingInput }
      }
      RemainingInput = result.RemainingInput
    }
    return ParserResult { expectedCodePoints, RemainingInput }
  }
}

func ExpectString(expectedString string) Parser {
  return func(input ParserInput) ParserResult {
    var result = ExpectCodePoints ([]rune (expectedString)) (input)
    var runes, isRuneArray = result.Result.([]rune)
    if isRuneArray {
      result.Result = string (runes)
    }
    return result
  }
}

func (parser Parser) Repeated () Parser {
  return func (input ParserInput) ParserResult {
    var result = ParserResult { list.New (), input }
    for result.RemainingInput != nil {
      var oneMoreResult = parser (result.RemainingInput)
      if oneMoreResult.Result == nil {
        return result
      }
      result.Result.(*list.List).PushBack (oneMoreResult.Result)
      result.RemainingInput = oneMoreResult.RemainingInput
    }
    return result
  }
}

func (parser Parser) OnceOrMore () Parser {
  return func (input ParserInput) ParserResult {
    var result = parser.Repeated () (input)
    if result.Result.(*list.List).Len () > 0 {
      return result
    }
    return ParserResult { nil, input }
  }
}

func (parser Parser) RepeatAndFoldLeft (accumulator interface{},
                                combine func (interface{},
                                              interface{}) interface{}) Parser {
  return func (input ParserInput) ParserResult {
    var result = ParserResult { accumulator, input }
    for result.RemainingInput != nil {
      var oneMoreResult = parser (result.RemainingInput)
      if oneMoreResult.Result == nil {
        return result
      }
      result.Result = combine (result.Result, oneMoreResult.Result)
      result.RemainingInput = oneMoreResult.RemainingInput
    }
    return result
  }
}

func (parser Parser) Bind (constructor func (interface{}) Parser) Parser {
  return func (input ParserInput) ParserResult {
    var firstResult = parser (input)
    var secondParser = constructor (firstResult.Result)
    return secondParser (firstResult.RemainingInput)
  }
}

func (parser Parser) OrElse (alternativeParser Parser) Parser {
  return func (input ParserInput) ParserResult {
    var FirstResult = parser (input)
    if FirstResult.Result != nil {
      return FirstResult
    }
    return alternativeParser (input)
  }
}

type Pair struct {
  First interface {}
  Second interface {}
}

func GetSecond (argument interface {}) interface {} {
  var pair, isPair = argument.(Pair)
  if isPair {
    return pair.Second
  }
  return argument
}

func GetFirst (argument interface {}) interface {} {
  var pair, isPair = argument.(Pair)
  if isPair {
    return pair.First
  }
  return argument
}

func (FirstParser Parser) AndThen (SecondParser Parser) Parser {
  return func (input ParserInput) ParserResult {
    var FirstResult = FirstParser (input)
    if FirstResult.Result != nil {
      var SecondResult = SecondParser (FirstResult.RemainingInput)
      if SecondResult.Result != nil {
        return ParserResult {
          Pair { FirstResult.Result, SecondResult.Result },
          SecondResult.RemainingInput }
      }
      return SecondResult
    }
    return FirstResult
  }
}

func (parser Parser) Convert (
                        Converter func (interface {}) interface {}) Parser {
  return func (input ParserInput) ParserResult {
    var result = parser (input)
    if result.Result != nil {
      result.Result = Converter (result.Result)
    }
    return result
  }
}

func (parser Parser) First () Parser {
  return parser.Convert (GetFirst)
}

func (parser Parser) Second () Parser {
  return parser.Convert (GetSecond)
}

type Nothing struct {}

func (parser Parser) optional () Parser {
  return func (input ParserInput) ParserResult {
    var result = parser (input)
    if result.Result == nil {
      result.Result = Nothing {}
    }
    return result
  }
}

type RuneArrayInput struct {
  Text []rune
  CurrentPosition int
}

func StringToInput (Text string) ParserInput {
  return &RuneArrayInput { []rune(Text), 0 }
}

func (input RuneArrayInput) RemainingInput () ParserInput {
  if input.CurrentPosition >= len (input.Text) {
    return nil
  }
  return RuneArrayInput { input.Text, input.CurrentPosition + 1 }
}

func (input RuneArrayInput) CurrentCodePoint () rune {
  if input.CurrentPosition >= len (input.Text) {
    return '\x00'
  }
  return input.Text[input.CurrentPosition]
}


func isIdentifierStartChar (FirstCodePoint rune) bool {
  return rune ('a') <= FirstCodePoint && FirstCodePoint <= rune ('z') ||
      rune ('A') <= FirstCodePoint && FirstCodePoint <= rune ('Z') ||
      rune ('_') == FirstCodePoint
}

func isDigit (codePoint rune) bool {
  return rune ('0') <= codePoint && codePoint <= rune ('9')
}

func isIdentifierChar (codePoint rune) bool {
  return isIdentifierStartChar (codePoint) || isDigit (codePoint)
}

func isSpaceChar (codePoint rune) bool {
  return codePoint == rune (' ') || codePoint == rune ('\n') ||
        codePoint == rune ('\r') || codePoint == rune ('\t')
}

func ExpectSeveral (isFirstChar func (rune) bool,
                    isLaterChar func (rune) bool) Parser {
  return func (input ParserInput) ParserResult {
    var FirstCodePoint = input.CurrentCodePoint ()
    if !isFirstChar (FirstCodePoint) {
      return ParserResult { nil, input }
    }
    var builder strings.Builder
    var codePoint = FirstCodePoint
    var RemainingInput = input
    for isLaterChar (codePoint) {
      builder.WriteRune (codePoint)
      RemainingInput = RemainingInput.RemainingInput ()
      codePoint = RemainingInput.CurrentCodePoint ()
    }
    return ParserResult { builder.String (), RemainingInput }
  }
}

var ExpectIdentifier Parser =
  ExpectSeveral (isIdentifierStartChar, isIdentifierChar)

var ExpectSpaces Parser =
  ExpectSeveral (isSpaceChar, isSpaceChar).optional ()

var ExpectNumber Parser =
  ExpectSeveral (isDigit, isDigit)

func MaybeSpacesBefore (parser Parser) Parser {
  return Parser (ExpectSpaces).AndThen (parser).Second ()
}
