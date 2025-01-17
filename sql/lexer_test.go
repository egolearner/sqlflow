// Copyright 2019 The SQLFlow Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLexer(t *testing.T) {
	a := assert.New(t)
	l := newLexer("")
	var n sqlSymType
	ret := l.Lex(&n)
	a.Equal(0, ret)
}

func TestNextAndBackup(t *testing.T) {
	a := assert.New(t)
	l := newLexer("ab")
	a.Equal('a', l.next())
	l.backup()
	a.Equal('a', l.next())
	a.Equal('b', l.next())
	a.Equal(eof, l.next())
	a.Equal(eof, l.next())
	l.backup()
	a.Equal(eof, l.next())
}

func TestSkipSpaces(t *testing.T) {
	a := assert.New(t)
	l := newLexer("ab")
	l.skipSpaces()
	a.Equal('a', rune(l.input[l.start]))
	a.Equal('a', l.next())
	l.skipSpaces()
	a.Equal('b', rune(l.input[l.start]))
	a.Equal('b', l.next())
}

func TestLexNumber(t *testing.T) {
	a := assert.New(t)
	l := newLexer("123.4")
	var n sqlSymType
	a.Equal(NUMBER, l.Lex(&n))
	a.Equal("123.4", n.val)

	l = newLexer("[5,10]")
	typs := []int{'[', NUMBER, ',', NUMBER, ']'}
	vals := []string{"[", "5", ",", "10", "]"}
	i := 0
	for {
		typ := l.Lex(&n)
		if typ == 0 {
			break
		}
		a.Equal(typs[i], typ)
		a.Equal(vals[i], n.val)
		i++
	}
}

func TestLexString(t *testing.T) {
	a := assert.New(t)
	l := newLexer(`  "\""  `)
	var n sqlSymType
	a.Equal(STRING, l.Lex(&n))
	a.Equal(`"\""`, n.val)
}

func TestLexOperator(t *testing.T) {
	a := assert.New(t)
	l := newLexer(`+-***/%()[]{}<<==,;`)

	typs := []int{
		'+', '-', POWER, '*', '/', '%', '(', ')', '[', ']', '{', '}',
		'<', LE, '=', ',', ';'}
	vals := []string{
		"+", "-", "**", "*", "/", "%", "(", ")", "[", "]",
		"{", "}", "<", "<=", "=", ",", ";"}
	i := 0
	var n sqlSymType
	for {
		typ := l.Lex(&n)
		if typ == 0 {
			break
		}
		a.Equal(typs[i], typ)
		a.Equal(vals[i], n.val)
		i++
	}
}

func TestLexIdentOrKeyword(t *testing.T) {
	a := assert.New(t)
	vals := []string{"a1_2b", "x.y", "x.y.z", "Select", "froM", "where", "tRain", "colUmn",
		"and", "or", "not"}
	typs := []int{IDENT, IDENT, IDENT, SELECT, FROM, WHERE, TRAIN, COLUMN,
		AND, OR, NOT}
	var n sqlSymType
	for i, it := range vals {
		l := newLexer(it)
		a.Equal(typs[i], l.Lex(&n))
		a.Equal(vals[i], n.val)
	}
}

func TestLexSQL(t *testing.T) {
	a := assert.New(t)
	l := newLexer("  Select * from a_table where a_table.col_1 > 100;")
	typs := []int{
		SELECT, '*', FROM, IDENT, WHERE, IDENT, '>', NUMBER, ';'}
	vals := []string{
		"Select", "*", "from", "a_table", "where",
		"a_table.col_1", ">", "100", ";"}
	var n sqlSymType
	for i := range typs {
		a.Equal(typs[i], l.Lex(&n))
		a.Equal(vals[i], n.val)
	}
}

func TestLexSQLWithOperators(t *testing.T) {
	a := assert.New(t)
	l := newLexer(`  CREATE TABLE tmp AS
SELECT a,b,c,d
FROM x.y
WHERE c=20190806
AND b IS NOT NULL AND b != "-" and COALESCE(d, "-")<>"-";`)
	var n sqlSymType
	typs := []int{
		IDENT, IDENT, IDENT, AS,
		SELECT, IDENT, ',', IDENT, ',', IDENT, ',', IDENT,
		FROM, IDENT, WHERE, IDENT, '=', NUMBER,
		AND, IDENT, IDENT, NOT, IDENT, AND, IDENT, NE, STRING, AND,
		IDENT, '(', IDENT, ',', STRING, ')', NE, STRING, ';'}
	vals := []string{
		"CREATE", "TABLE", "tmp", "AS", "SELECT",
		"a", ",", "b", ",", "c", ",", "d", "FROM", "x.y",
		"WHERE", "c", "=", "20190806", "AND", "b",
		"IS", "NOT", "NULL", "AND", "b", "!=", "\"-\"", "and",
		"COALESCE", "(", "d", ",", "\"-\"", ")", "<>", "\"-\"", ";"}

	for i := range typs {
		a.Equal(typs[i], l.Lex(&n))
		a.Equal(vals[i], n.val)
	}
}

func TestLexSQLWithSingleQuote(t *testing.T) {
	a := assert.New(t)
	l := newLexer(`  SELECT a,b
FROM x.y
WHERE b='20190806';`)
	var n sqlSymType
	typs := []int{
		SELECT, IDENT, ',', IDENT,
		FROM, IDENT, WHERE, IDENT, '=', STRING, ';'}
	vals := []string{
		"SELECT", "a", ",", "b", "FROM", "x.y",
		"WHERE", "b", "=", "'20190806'", ";"}

	for i := range typs {
		a.Equal(typs[i], l.Lex(&n))
		a.Equal(vals[i], n.val)
	}
}
