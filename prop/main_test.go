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
  . "github.com/QAhell/Parser-Gombinators/parse"
  "testing"
)

func eqnError (t *testing.T, eqn *Equation, value string) {
  t.Errorf ("The equation %s should evaluate to %s!\n", eqn, value)
}

func TestEquation (t *testing.T) {
  var fu = &ValueTerm { &StringValue { "熊猫" } }
  var bar = &ValueTerm { &StringValue { "大狗" } }
  var eqnFuFu = &Equation { fu, fu }
  var eqnFuBar = &Equation { fu, bar }
  var shouldBeTrue = eqnFuFu.Simplify (make(map[string]Value))
  var shouldBeFalse = eqnFuBar.Simplify (make(map[string]Value))
  var shouldBeTrueValue, isFuFuValue = shouldBeTrue.(*ValueTerm)
  var shouldBeFalseValue, isFuBarValue = shouldBeFalse.(*ValueTerm)
  if !isFuFuValue {
    eqnError (t, eqnFuFu, "true")
  }
  if !isFuBarValue {
    eqnError (t, eqnFuBar, "false")
  }
  var shouldBeTrueBool, isFuFuBool = shouldBeTrueValue.Value.(*BoolValue)
  var shouldBeFalseBool, isFuBarBool = shouldBeFalseValue.Value.(*BoolValue)
  if !isFuFuBool || !shouldBeTrueBool.Value {
    eqnError (t, eqnFuFu, "true")
  }
if !isFuBarBool || shouldBeFalseBool.Value {
    eqnError (t, eqnFuBar, "false")
  }
}

var trew = &BoolValue { true }
var fawlz = &BoolValue { false }

func TestNot (t *testing.T) {
  var env = make (map[string] Value)
  var eggs = &Identifier { "x" }
  var naughtEggs = &Not { &Not { &Not { eggs } } }
  if "(NOT x)" != naughtEggs.Simplify (env).String () {
    t.Errorf ("Expected %s to evaluate to (NOT x).", naughtEggs)
  }
  env["x"] = trew
  if "FALSE" != naughtEggs.Simplify (env).String () {
    t.Errorf ("Expected %s to evaluate to FALSE with x=TRUE.", naughtEggs)
  }
  env["x"] = fawlz
  if "TRUE" != naughtEggs.Simplify (env).String () {
    t.Errorf ("Expected %s to evaluate to TRUE with x=FALSE.", naughtEggs)
  }
}

func TestAnd (t *testing.T) {
  var env = make (map[string] Value)
  var and = &And { &Identifier { "x" }, &Identifier { "y" } }
  env["x"] = trew
  if "y" != and.Simplify (env).String () {
    t.Errorf ("Expected %s to evaluate to y with x=TRUE", and)
  }
  env["x"] = fawlz
  if "FALSE" != and.Simplify (env).String () {
    t.Errorf ("Expected %s to evaluate to FALSE with x=FALSE", and)
  }
  delete (env, "x")
  env["y"] = trew
  if "x" != and.Simplify (env).String () {
    t.Errorf ("Expected %s to evaluate to x with y=TRUE", and)
  }
  env["y"] = fawlz
  if "FALSE" != and.Simplify (env).String () {
    t.Errorf ("Expected %s to evaluate to FALSE with y=FALSE", and)
  }
}

func TestOr (t *testing.T) {
  var env = make (map[string] Value)
  var or = &Or { &Identifier { "x" }, &Identifier { "y" } }
  env["x"] = trew
  if "TRUE" != or.Simplify (env).String () {
    t.Errorf ("Expected %s to evaluate to TRUE with x=TRUE", or)
  }
  env["x"] = fawlz
  if "y" != or.Simplify (env).String () {
    t.Errorf ("Expected %s to evaluate to y with x=FALSE", or)
  }
  delete (env, "x")
  env["y"] = trew
  if "TRUE" != or.Simplify (env).String () {
    t.Errorf ("Expected %s to evaluate to TRUE with y=TRUE", or)
  }
  env["y"] = fawlz
  if "x" != or.Simplify (env).String () {
    t.Errorf ("Expected %s to evaluate to x with y=FALSE", or)
  }
}

func TestParser (t *testing.T) {
  var text = "NOT x=\"y\" OR TRUE AND (FALSE OR z)"
  var expected =
    &Or { &Not { &Equation { &Identifier { "x" },
        &ValueTerm { &StringValue { "y" } } } },
      &And { &ValueTerm { trew },
        &Or { &ValueTerm { fawlz }, &Identifier { "z" } } } }
  var input = StringToInput (text)
  var result = ParseOr (input)
  var term, isTerm = result.Result.(Term)
  if !isTerm {
    t.Errorf ("Expected the parser to parse '%s' well.\n", text)
  }
  if !expected.Equals (term) {
    t.Errorf ("Expected the text '%s' to become the " +
      "term %s.", text, term)
  }
}
