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

package parse

import (
  "strings"
  "container/list"
  "io"
  "bufio"
  "os"
)

// Parser parses its input and produces some result.
type Parser func (ParserInput) ParserResult

// ParserInput is anything that can produce a sequence of code points.
// RuneArrayInput is one implementation that you can use. See StringToInput
// if you want to create ParserInput directly from a string.
type ParserInput interface {

  // CurrentCodePoint returns the rune at the beginning of this input
  CurrentCodePoint () rune

  // RemainingInput returns everything that comes after the current code point.
  RemainingInput () ParserInput
}

// ParserResult is the result of a parse along with the input that remains to
// be parsed.
type ParserResult struct {

  // Result can be anything except for nil which indicates that parsing failed.
  // Parsing failed <==> Result == nil
  Result interface{}

  // RemainingInput is the rest of the input after the successful parse of
  // Result. If the parse failed then it's just the input from before the
  // parsing attempt.
  RemainingInput ParserInput
}

// ExpectCodePoint expects exactly one rune in the input. If the input
// starts with this rune it will become the result.
func ExpectCodePoint (expectedCodePoint rune) Parser {
  return func (input ParserInput) ParserResult {
    if expectedCodePoint == input.CurrentCodePoint () {
      return ParserResult { expectedCodePoint, input.RemainingInput () }
    }
    return ParserResult { nil, input }
  }
}

// ExpectCodePoints expects exactly the code points from the slice
// expectedCodePoints at the beginning of the input in the given order.
// If the input begins with these code points then expectedCodePoints will
// be the result of the parse.
func ExpectCodePoints (expectedCodePoints []rune) Parser {
  return func (input ParserInput) ParserResult {
    var RemainingInput = input
    for _, expectedCodePoint := range expectedCodePoints {
      if nil == RemainingInput {
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

// ExpectString expects the input to begin with the code points from the
// expectedString in the given order. If the input starts with these code
// points then expectedString will be the result of the parse.
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

// Repeated applies a parser zero or more times and accumulates the results
// of the parses in a list. This parse always produces a non-nil result.
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

// OnceOrMore is like Repeated except that it doesn't allow parsing zero times.
func (parser Parser) OnceOrMore () Parser {
  return func (input ParserInput) ParserResult {
    var result = parser.Repeated () (input)
    if result.Result.(*list.List).Len () > 0 {
      return result
    }
    return ParserResult { nil, input }
  }
}

// RepeatAndFoldLeft is like Repeat except that it doesn't produce a list.
// You can make RepeatAndFoldLeft implement Repeat by using the empty list as
// the accumulator and PushBack as the combine function. The accumulator is
// the initial value and every result of the parser will be added to the
// accumulator using the combine function. See the calculator example for
// an idiomatic use-case.
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

// Bind uses the result of a first parser to construct a second parser that
// will parse the left-over input from the first parser. You can use this
// to implement syntax annotations.
func (parser Parser) Bind (constructor func (interface{}) Parser) Parser {
  return func (input ParserInput) ParserResult {
    var firstResult = parser (input)
    var secondParser = constructor (firstResult.Result)
    return secondParser (firstResult.RemainingInput)
  }
}

// OrElse uses the first parser to parse the input. If this fails it will
// use the second parser to parse the same input. Only use non-overlapping
// parsers with this combinator! For the most part it's the usual alternative
// except that it's first-come, first-served: if the first parser succeeds,
// then it will not attempt to use the second parser and there's no
// back-tracking. This is in contrast to most regex-libs where the longest
// match wins. The first match wins here, please keep this in mind.
func (parser Parser) OrElse (alternativeParser Parser) Parser {
  return func (input ParserInput) ParserResult {
    var FirstResult = parser (input)
    if FirstResult.Result != nil {
      return FirstResult
    }
    return alternativeParser (input)
  }
}

// Pair is a simple pair. Please use it only as an intermediate data structure.
// If you know what you're parsing then convert your pairs into structs with
// more meaningful names.
type Pair struct {

  // First is the first component of the pair.
  First interface {}

  // Second is the second component of the pair.
  Second interface {}
}

// GetSecond extracts the second component of a pair or
// returns the argument if it is not a pair.
func GetSecond (argument interface {}) interface {} {
  var pair, isPair = argument.(Pair)
  if isPair {
    return pair.Second
  }
  return argument
}

// GetFirst extracts the first component of a pair or
// returns the argument if it is not a pair.
func GetFirst (argument interface {}) interface {} {
  var pair, isPair = argument.(Pair)
  if isPair {
    return pair.First
  }
  return argument
}

// AndThen applies the firstParser to the input and then the
// secondParser. The result will be a Pair containing the results
// of both parsers.
func (firstParser Parser) AndThen (secondParser Parser) Parser {
  return func (input ParserInput) ParserResult {
    var firstResult = firstParser (input)
    if firstResult.Result != nil {
      var secondResult = secondParser (firstResult.RemainingInput)
      if secondResult.Result != nil {
        return ParserResult {
          Pair { firstResult.Result, secondResult.Result },
          secondResult.RemainingInput }
      }
      return secondResult
    }
    return firstResult
  }
}

// Convert applies the converter to the result of a successful parse.
// If the parser fails then Convert won't do anything.
func (parser Parser) Convert (
                        converter func (interface {}) interface {}) Parser {
  return func (input ParserInput) ParserResult {
    var result = parser (input)
    if result.Result != nil {
      result.Result = converter (result.Result)
    }
    return result
  }
}

// First extracts the first component of the result of a successful parse.
// If the parser fails then First won't do anything.
func (parser Parser) First () Parser {
  return parser.Convert (GetFirst)
}

// Second extracts the second component of the result of a successful parse.
// If the parser fails then Second won't do anything.
func (parser Parser) Second () Parser {
  return parser.Convert (GetSecond)
}

// Nothing is the result of successfully parsing nothing at all.
// Don't confuse it with nil which indicates failure.
// This Nothing type means that the parser has explicitly allowed
// an empty input to be valid and to produce this result!
type Nothing struct {}

// Optional applies the parser zero or one times to the input.
// If the parser itself would fail then the Optional parser can still
// produce a successful parse with the result Nothing{}.
func (parser Parser) Optional () Parser {
  return func (input ParserInput) ParserResult {
    var result = parser (input)
    if result.Result == nil {
      result.Result = Nothing {}
      result.RemainingInput = input
    }
    return result
  }
}

// FileInput is an implementation of ParserInput
// You can use FileToInput to create instances of this type directly
// from a path.
type FileInput struct {

  // File is the underlying file of this parser input
  File        io.RuneReader

  // CurrentRune is the current character
  CurrentRune rune

  // RestOfInput is what remains after the CurrentRune
  RestOfInput *FileInput
}

// FileToInput converts a RuneReader into a ParserInput.
func FileToInput (file io.RuneReader) FileInput {
  var r, _, err = file.ReadRune ()
  if err != nil {
    return FileInput { file, '\x00', nil }
  }
  return FileInput { file, r, nil }
}

// FilenameToInput opens a file and converts it into ParserInput.
func FilenameToInput (filename string) FileInput {
  var file, err = os.Open (filename)
  if err != nil {
    panic (err)
  }
  return FileToInput (bufio.NewReader (file))
}

// RemainingInput is necessary for FileInput to implement ParserInput
func (input FileInput) RemainingInput () ParserInput {
  if input.RestOfInput != nil {
    return *(input.RestOfInput)
  }
  if input.File == nil {
    return nil
  }
  var r, _, err = input.File.ReadRune ()
  if err != nil {
    input.File = nil
    return nil
  }
  input.RestOfInput = &FileInput { input.File, r, nil }
  return *input.RestOfInput
}

// CurrentCodePoint is necessary for FileInput to implement ParserInput
func (input FileInput) CurrentCodePoint () rune {
  return input.CurrentRune
}

// RuneArrayInput is an implementation of ParserInput.
// You can use StringToInput to create instances of this type directly
// from strings.
type RuneArrayInput struct {

  // Text is the whole input text. Please keep it unchanged while parsers are
  // working on it.
  Text []rune

  // CurrentPosition points to the current code point in the Text
  CurrentPosition int
}

// RemainingInput is necessary for RuneArrayInput to implement ParserInput
func (input RuneArrayInput) RemainingInput () ParserInput {
  if input.CurrentPosition + 1 >= len (input.Text) {
    return nil
  }
  return RuneArrayInput { input.Text, input.CurrentPosition + 1 }
}

// CurrentCodePoint is necessary for RuneArrayInput to implement ParserInput
func (input RuneArrayInput) CurrentCodePoint () rune {
  if input.CurrentPosition >= len (input.Text) {
    return '\x00' // only happens with empty input now?!
  }
  return input.Text[input.CurrentPosition]
}

// StringToInput converts a string to a RuneArrayInput so you can use parsers
// on it.
func StringToInput (Text string) ParserInput {
  return &RuneArrayInput { []rune(Text), 0 }
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

// ExpectSeveral accepts the first code point from the input if isFirstChar
// returns true. After reading the first character, it takes all following code
// points as long as they satisfy isLaterChar. It stops parsing the input at the
// first code point that doesn't satisfy isLaterChar. ExpectSeveral will only
// fail if the first character from the input doesn't satisfy isFirstChar!
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

// ExpectIdentifier parses a [a-zA-Z_][a-zA-Z0-9_]* from the input.
var ExpectIdentifier Parser =
  ExpectSeveral (isIdentifierStartChar, isIdentifierChar)

// ExpectSpaces parses a [ \t\n\r]* from the input.
var ExpectSpaces Parser =
  ExpectSeveral (isSpaceChar, isSpaceChar).Optional ()

// ExpectNumber parses a [0-9]+ from the input and the result will be a string.
// You need to convert it into your favorite number type by yourself.
var ExpectNumber Parser =
  ExpectSeveral (isDigit, isDigit)

// MaybeSpacesBefore allows and ignores space characters before applying the
// parser from the argument.
func MaybeSpacesBefore (parser Parser) Parser {
  return Parser (ExpectSpaces).AndThen (parser).Second ()
}
