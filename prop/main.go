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
  "strings"
  "encoding/json"
  "io/ioutil"
)

/*
  Propositional logic with variables and equations.

  Example usage: ./prop environment.json 'NOT (a AND b)'

  The environment.json is just one object where the keys
  are names of variables of the formula and the values are
  either bools or strings. The resulting formula will be
  simplified.

  Identifier   := [a-zA-Z_][a-zA-Z0-9_]* except for keywords
  String       := ["]([^"]|[\\]["])*["]
  Bool         := TRUE | FALSE
  Atom         := Identifier
                | Bool
                | "(" Expression ")"
  Eqn          := Atom ("=" Atom)?
  Not          := "NOT"* Atom
  And          := Not (("AND") Not)*
  Expression   := And (("OR") And)*

  Important: No left-recursion!
  Important: No overlapping rules/alternatives!

  Keywords of the language: NOT, AND, OR, TRUE, FALSE

 */
func main () {
  if len (os.Args) != 2 && len (os.Args) != 3 {
    fmt.Printf (licenceNotice)
    return
  }
  var env = readEnvironment ()
  var input = StringToInput (os.Args[len (os.Args) - 1])
  var parserResult = ParseOr (input)
  var term, isTerm = parserResult.Result.(Term)
  if isTerm {
    presentResult (term, env, parserResult)
  } else {
    fmt.Printf ("Can't parse the input!\n")
  }
}

/* presentResult shows to the user the parsed term, the simplified term
  and remaining input from the formula text */
func presentResult (term Term, env map[string] Value,
                    parserResult ParserResult) {
  fmt.Printf ("parsed expression = %s\n", term.String ())
  fmt.Printf ("simplified result = %s\n", term.Simplify (env).String ())
  if nil != parserResult.RemainingInput {
    var inp = parserResult.RemainingInput.(RuneArrayInput)
    var rest = inp.Text[inp.CurrentPosition:]
    fmt.Printf ("There's some remaining input: %s\n", string (rest))
  }
}

/* readEnvironment reads the contents of the JSON-file supplied by the
  user into an enviornment map */
func readEnvironment () map[string] Value {
  var env map[string] Value = make (map[string] Value)
  if len (os.Args) == 3 {
    var jsonFile, err = os.Open (os.Args[1])
    if err != nil {
      panic (err)
    }
    var jsonContent map[string] interface{}
    var bytes, _ = ioutil.ReadAll (jsonFile)
    json.Unmarshal (bytes, &jsonContent)
    for key, value := range jsonContent {
      readEnvEntry (key, value, env)
    }
    defer jsonFile.Close ()
  }
  return env
}

/* readEnvEntry reads one key/value-pair of the JSON-file supplied by
  the user and puts it into the environment.
*/
func readEnvEntry (key string, value interface{}, env map[string] Value) {
  var valueBool, isBool = value.(bool)
  var valueString, isString = value.(string)
  if isBool {
    env[key] = &BoolValue { valueBool }
  } else if isString {
    env[key] = &StringValue { valueString }
  } else {
    fmt.Printf ("Invalid value type for key '%s'!\n", key)
  }
}

/* Value is a term that doesn't consist of any further parts
  and that can't be simplified any further */
type Value interface {
  /* String returns something that ParseValue can parse */
  String () string
  /* Equals is structural equality */
  Equals (other Value) bool
}

/* StringValue just wraps the string type */
type StringValue struct { Value string }
/* BoolValue just wraps the bool type */
type BoolValue struct { Value bool }

/* String wraps the contents of the string with quotes and escapes quotes
  inside the string */
func (str *StringValue) String () string {
  return "\"" + strings.NewReplacer ("\"", "\\\"").Replace (str.Value) + "\""
}

/* String converts true to "TRUE" and false to "FALSE" */
func (b *BoolValue) String () string {
  if b.Value {
    return "TRUE"
  }
  return "FALSE"
}

/* Equals returns true iff the string values are the same */
func (str *StringValue) Equals (other Value) bool {
  var otherStringValue, isStringValue = other.(*StringValue)
  return isStringValue && str.Value == otherStringValue.Value
}

/* Equals returns true iff the bool values are the same */
func (b *BoolValue) Equals (other Value) bool {
  var otherBoolValue, isBoolValue = other.(*BoolValue)
  return isBoolValue && b.Value == otherBoolValue.Value
}

/* Term is a propositional formula with extras */
type Term interface {
  /* String returns something that ParseOr can parse */
  String () string
  /* Simplify substitutes the values from the environment
    and removes redundant sub-formulas */
  Simplify (env map[string] Value) Term
  /* Equals is structural equality */
  Equals (other Term) bool
}

/* ValueTerm just wraps Values */
type ValueTerm  struct { Value Value             }
/* Identifiers are the variables of the formula */
type Identifier struct { /* Name can a key into the env */ Name  string }
/* Equation is a term of the form X=Y */
type Equation   struct { Left  Term ; Right Term }
/* Not is negation */
type Not        struct { Arg   Term              }
/* And is conjunction */
type And        struct { Left  Term ; Right Term }
/* Or is disjunction */
type Or         struct { Left  Term ; Right Term }

/* String delegates to Value.String () */
func (value *ValueTerm) String () string {
  return value.Value.String ()
}

func (ident *Identifier) String () string {
  return ident.Name
}

func (equals *Equation) String () string {
  return equals.Left.String () + "=" + equals.Right.String ()
}

func (not *Not) String () string {
  return "(NOT " + not.Arg.String () + ")"
}

func (and *And) String () string {
  return "(" + and.Left.String () + " AND " + and.Right.String () + ")"
}

func (or *Or) String () string {
  return "(" + or.Left.String () + " OR " + or.Right.String () + ")"
}

/* Simplify has no effect on values */
func (value *ValueTerm) Simplify (env map[string] Value) Term {
  return value
}

/* Simplify looks up identifiers in the environment */
func (ident *Identifier) Simplify (env map[string] Value) Term {
  if value, ok := env[ident.Name] ; ok {
    return &ValueTerm { value }
  }
  return ident
}

/* Simplify calls Equals to check whether the left and right side of the
  equation are equal */
func (eqn *Equation) Simplify (env map[string] Value) Term {
  var left = eqn.Left.Simplify (env)
  var right = eqn.Right.Simplify (env)
  var _, leftIsValue = left.(*ValueTerm)
  var _, rightIsValue = right.(*ValueTerm)
  if left.Equals (right) {
    return &ValueTerm { &BoolValue { true } }
  } else if leftIsValue && rightIsValue {
    return &ValueTerm { &BoolValue { false } }
  }
  // Leave the term unchanged if possible
  if left == eqn.Left && right == eqn.Right {
    return eqn
  }
  return &Equation { left, right }
}

/* Simplify removes double-negations */
func (not *Not) Simplify (env map[string] Value) Term {
  var arg = not.Arg.Simplify (env)
  var argValue, isValue = arg.(*ValueTerm)
  if isValue {
    var argBool, isBool = argValue.Value.(*BoolValue)
    if isBool {
      return &ValueTerm { &BoolValue { !argBool.Value } }
    }
  }

  var argNot, isNot = arg.(*Not)
  if isNot {
    return argNot.Arg
  }
  // Leave the term unchanged if possible
  if arg == not.Arg {
    return not
  }
  return &Not { arg }
}

/* Simplify converts (true AND x) into x, (false AND x) into false
  and vice versa. */
func (and *And) Simplify (env map[string] Value) Term {
  var left = and.Left.Simplify (env)
  var leftValue, isValue = left.(*ValueTerm)
  if isValue {
    var leftBool, isBool = leftValue.Value.(*BoolValue)
    if isBool {
      if leftBool.Value {
        return and.Right.Simplify (env)
      }
      return left
    }
  }
  var right = and.Right.Simplify (env)
  var rightValue, isRightValue = right.(*ValueTerm)
  if isRightValue {
    var rightBool, isBool = rightValue.Value.(*BoolValue)
    if isBool {
      if !rightBool.Value {
        return right
      }
      return left
    }
  }
  // Leave the term unchanged if possible
  if left == and.Left && right == and.Right {
    return and
  }
  return &And { left, right }
}

/* Simplify converts (true OR x) into true, (false OR x) into x
  and vice versa. */
func (or *Or) Simplify (env map[string] Value) Term {
  var left = or.Left.Simplify (env)
  var leftValue, isValue = left.(*ValueTerm)
  if isValue {
    var leftBool, isBool = leftValue.Value.(*BoolValue)
    if isBool {
      if !leftBool.Value {
        return or.Right.Simplify (env)
      }
      return left
    }
  }
  var right = or.Right.Simplify (env)
  var rightValue, isRightValue = right.(*ValueTerm)
  if isRightValue {
    var rightBool, isBool = rightValue.Value.(*BoolValue)
    if isBool {
      if rightBool.Value {
        return right
      }
      return left
    }
  }
  // Leave the term unchanged if possible
  if left == or.Left && right == or.Right {
    return or
  }
  return &Or { left, right }
}

func (value *ValueTerm) Equals (other Term) bool {
  var otherValueTerm, isValueTerm = other.(*ValueTerm)
  return isValueTerm && value.Value.Equals (otherValueTerm.Value)
}

func (identifier *Identifier) Equals (other Term) bool {
  var otherIdentifier, isIdentifier = other.(*Identifier)
  return isIdentifier && identifier.Name == otherIdentifier.Name
}

func (eqn *Equation) Equals (other Term) bool {
  var otherEqn, isEquation = other.(*Equation)
  return isEquation && eqn.Left.Equals (otherEqn.Left) &&
    eqn.Right.Equals (otherEqn.Right)
}

func (not *Not) Equals (other Term) bool {
  var otherNot, isNot = other.(*Not)
  return isNot && not.Arg.Equals (otherNot.Arg)
}

func (and *And) Equals (other Term) bool {
  var otherAnd, isAnd = other.(*And)
  return isAnd && and.Left.Equals (otherAnd.Left) &&
    and.Right.Equals (otherAnd.Right)
}

func (or *Or) Equals (other Term) bool {
  var otherOr, isOr = other.(*Or)
  return isOr && or.Left.Equals (otherOr.Left) &&
    or.Right.Equals (otherOr.Right)
}

/* ParseIdent parses identifiers, excludes keywords and prints to stdout
  if the user almost hits a keyword. */
func ParseIdent (input ParserInput) ParserResult {
  return Parser (MaybeSpacesBefore (ExpectIdentifier)).
    Convert (func (arg interface{}) interface{} {
        var text, isText = arg.(string)
        if isText {
          // Exclude all the keywords, they're not valid identifiers!
          if text == "TRUE" || text == "FALSE" || text == "NOT" ||
             text == "OR" || text == "AND" {
            return nil
          }
          // Warn the user about almost-Keywords!
          if text == "true" || text == "True" || text == "false" ||
            text == "False" || text == "not" || text == "Not" ||
            text == "or" || text == "Or" || text == "and" || text == "And" {
            fmt.Printf (
              "You probably don't want to use \"%s\" as a variable name!\n",
              text)
            fmt.Printf ("This language is case sensitive.\n")
            fmt.Printf (
              "Use all uppper case letters for logical expressions.\n\n")
          }
          return &Identifier { text }
        }
        return nil
      }) (input)
}

/* ParseBool parses TRUE and FALSE. */
func ParseBool (input ParserInput) ParserResult {
  return MaybeSpacesBefore (ExpectIdentifier).
    Convert (func (identifier interface{}) interface{} {
      var text, isText = identifier.(string)
      if isText {
        if text == "TRUE" {
          return true
        } else if text == "FALSE" {
          return false
        }
      }
      return nil
    }) (input)
}

/* ParseString parses a string literal */
func ParseString (input ParserInput) ParserResult {
  return ExpectCodePoint ('"').AndThen (
            (ExpectCodePoint ('\\').AndThen (ExpectCodePoint ('"')).Second ().
             OrElse (ExpectNotCodePoint ([]rune { '"' }))).
              RepeatAndFoldLeft ("",
                func (acc interface{}, char interface{}) interface{} {
                  return acc.(string) + string (char.(rune))
                })).Second ().
            AndThen (ExpectCodePoint ('"')).First () (input)
}

/* ParseValue parses a string literal or a boolean */
func ParseValue (input ParserInput) ParserResult {
  return MaybeSpacesBefore (Parser (ParseBool).
    Convert (func (arg interface{}) interface{} {
            var argBool, isBool = arg.(bool)
            if !isBool {
              return nil
            }
            return &ValueTerm { &BoolValue { argBool } }
          }).OrElse (
         Parser (ParseString).Convert (func (arg interface{}) interface{} {
            var argString, isString = arg.(string)
            if !isString {
              return nil
            }
            return &ValueTerm { &StringValue { argString } }
          }))) (input)
}

/* expectIdent parses an identifier like "TRUE" or "FALSE".
  The reason for why we can't just use the parser expect (..)
  is that some identifiers might start with the text we try
  to parse. The parser expect ("(") parses the text "(fu" into
  the result "(" and the rest of the input "fu".
  However, we don't want "TRUEfu" to become the result
  "TRUE" and then "fu". That's why we need another expectIdent
  parser that firstly parses the whole identifier and then
  checks whether it's the correct one. */
func expectIdent (text string) Parser {
  return MaybeSpacesBefore (
    func (input ParserInput) ParserResult {
      var result = ExpectIdentifier (input)
      if (result.Result == text) {
        return result
      } else {
        return ParserResult { nil, input }
      }
    })
}

/* ParseAtom parse values, identifiers and expressions
  within parenthesis */
func ParseAtom (input ParserInput) ParserResult {
  return Parser (ParseValue).OrElse (ParseIdent).OrElse (
    expect ("(").AndThen (ParseOr).AndThen (expect (")")).
      First().Second()) (input)
}

/* ParseEqn parses equations and atoms */
func ParseEqn (input ParserInput) ParserResult {
  return Parser (ParseAtom).AndThen (
    expect ("=").AndThen (ParseAtom).Second ().Optional ()).
      Convert (func (arg interface{}) interface{} {
          var pair = arg.(Pair)
          var lhs = pair.First.(Term)
          if pair.Second == (Nothing{}) {
            return lhs
          }
          return &Equation { lhs, pair.Second.(Term) }
        }) (input)
}

/* ParseNot subsumes ParseEqn and parses negations and drops double negations.
  This is okay here because this is not a theorem prover with a constructive
  logic. */
func ParseNot (input ParserInput) ParserResult {
  // see the comment on expectIdent for why not expect("NOT")
  return expectIdent ("NOT").RepeatAndFoldLeft (false,
      func (negate interface{}, _ interface{}) interface{} { return !negate.(bool) }).
    AndThen (ParseEqn).Convert (func (arg interface{}) interface{} {
        var pair, _ = arg.(Pair)
        var negate = pair.First.(bool)
        var term = pair.Second.(Term)
        if negate {
          return &Not { term }
        } else {
          return term
        }
      }) (input)
}

/* ParseAnd subsumes ParseNot and parses conjunctions */
func ParseAnd (input ParserInput) ParserResult {
  // see the comment on expectIdent for why not expect("AND")
  return Parser (ParseNot).AndThen (
      expectIdent ("AND").AndThen (ParseAnd).Second ().Optional ()).
        Convert (func (arg interface{}) interface{} {
                  var pair, _ = arg.(Pair)
                  if pair.Second == (Nothing{}) {
                    return pair.First.(Term)
                  }
                  return &And { pair.First.(Term), pair.Second.(Term) }
                }) (input)
}

/* ParseOr subsumes ParseAnd and parses disjunctions */
func ParseOr (input ParserInput) ParserResult {
  // see the comment on expectIdent for why not expect("OR")
  return Parser (ParseAnd).AndThen (
      expectIdent ("OR").AndThen (ParseOr).Second ().Optional ()).
        Convert (func (arg interface{}) interface{} {
                  var pair, _ = arg.(Pair)
                  if pair.Second == (Nothing{}) {
                    return pair.First.(Term)
                  }
                  return &Or { pair.First.(Term), pair.Second.(Term) }
                }) (input)
}

/* expect trys to find a certain text at the beginning of the input */
func expect (text string) Parser {
  return MaybeSpacesBefore (ExpectString (text))
}

/* licenceNotice contains the usage and GPL3 text */
var licenceNotice =
    "Usage:\n" +
    "  prop [name-of-environment.json] 'expression'\n" +
    "    The name-of-environment.json is optional.\n\n" +
    "This program is a basic boolean expression simplifier.\n" +
    "  It lets you evaluate boolean expressions in an environment.\n\n" +
    "License:\n" +
    "  Copyright (C) 2018  Armin Heller\n\n" +
    "  This program is free software: you can redistribute it and/or modify\n" +
    "  it under the terms of the GNU General Public License as published by\n" +
    "  the Free Software Foundation, either version 3 of the License, or\n" +
    "  (at your option) any later version.\n\n" +
    "  This program is distributed in the hope that it will be useful,\n" +
    "  but WITHOUT ANY WARRANTY; without even the implied warranty of\n" +
    "  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the\n" +
    "  GNU General Public License for more details.\n\n" +
    "  You should have received a copy of the GNU General Public License\n" +
    "  along with this program.  If not, see <https://www.gnu.org/licenses/>.\n\n"
