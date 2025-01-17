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

// Code generated by goyacc -p sql -o parser.go sql.y. DO NOT EDIT.

//line sql.y:2
package sql

import __yyfmt__ "fmt"

//line sql.y:2

import (
	"fmt"
	"strings"
	"sync"
)

/* expr defines an expression as a Lisp list.  If len(val)>0,
   it is an atomic expression, in particular, NUMBER, IDENT,
   or STRING, defined by typ and val; otherwise, it is a
   Lisp S-expression. */
type expr struct {
	typ  int
	val  string
	sexp exprlist
}

type exprlist []*expr

/* construct an atomic expr */
func atomic(typ int, val string) *expr {
	return &expr{
		typ: typ,
		val: val,
	}
}

/* construct a funcall expr */
func funcall(name string, oprd exprlist) *expr {
	return &expr{
		sexp: append(exprlist{atomic(IDENT, name)}, oprd...),
	}
}

/* construct a unary expr */
func unary(typ int, op string, od1 *expr) *expr {
	return &expr{
		sexp: append(exprlist{atomic(typ, op)}, od1),
	}
}

/* construct a binary expr */
func binary(typ int, od1 *expr, op string, od2 *expr) *expr {
	return &expr{
		sexp: append(exprlist{atomic(typ, op)}, od1, od2),
	}
}

/* construct a variadic expr */
func variadic(typ int, op string, ods exprlist) *expr {
	return &expr{
		sexp: append(exprlist{atomic(typ, op)}, ods...),
	}
}

type extendedSelect struct {
	extended bool
	train    bool
	standardSelect
	trainClause
	predictClause
}

type standardSelect struct {
	fields exprlist
	tables []string
	where  *expr
	limit  string
}

type trainClause struct {
	estimator  string
	trainAttrs attrs
	columns    columnClause
	label      string
	save       string
}

/* If no FOR in the COLUMN, the key is "" */
type columnClause map[string]exprlist
type filedClause exprlist

type attrs map[string]*expr

type predictClause struct {
	predAttrs attrs
	model     string
	into      string
}

var parseResult *extendedSelect

func attrsUnion(as1, as2 attrs) attrs {
	for k, v := range as2 {
		if _, ok := as1[k]; ok {
			log.Panicf("attr %q already specified", as2)
		}
		as1[k] = v
	}
	return as1
}

//line sql.y:106
type sqlSymType struct {
	yys  int
	val  string /* NUMBER, IDENT, STRING, and keywords */
	flds exprlist
	tbls []string
	expr *expr
	expl exprlist
	atrs attrs
	eslt extendedSelect
	slct standardSelect
	tran trainClause
	colc columnClause
	labc string
	infr predictClause
}

const SELECT = 57346
const FROM = 57347
const WHERE = 57348
const LIMIT = 57349
const TRAIN = 57350
const PREDICT = 57351
const WITH = 57352
const COLUMN = 57353
const LABEL = 57354
const USING = 57355
const INTO = 57356
const FOR = 57357
const AS = 57358
const IDENT = 57359
const NUMBER = 57360
const STRING = 57361
const AND = 57362
const OR = 57363
const GE = 57364
const LE = 57365
const NE = 57366
const NOT = 57367
const POWER = 57368
const UMINUS = 57369

var sqlToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"SELECT",
	"FROM",
	"WHERE",
	"LIMIT",
	"TRAIN",
	"PREDICT",
	"WITH",
	"COLUMN",
	"LABEL",
	"USING",
	"INTO",
	"FOR",
	"AS",
	"IDENT",
	"NUMBER",
	"STRING",
	"AND",
	"OR",
	"'>'",
	"'<'",
	"'='",
	"'!'",
	"GE",
	"LE",
	"NE",
	"'+'",
	"'-'",
	"'*'",
	"'/'",
	"'%'",
	"NOT",
	"POWER",
	"UMINUS",
	"';'",
	"'('",
	"')'",
	"','",
	"'['",
	"']'",
	"'\"'",
	"'\\''",
}
var sqlStatenames = [...]string{}

const sqlEofCode = 1
const sqlErrCode = 2
const sqlInitialStackSize = 16

//line sql.y:281

/* Like Lisp's builtin function cdr. */
func (e *expr) cdr() (r []string) {
	for i := 1; i < len(e.sexp); i++ {
		r = append(r, e.sexp[i].String())
	}
	return r
}

/* Convert exprlist to string slice. */
func (el exprlist) Strings() (r []string) {
	for i := 0; i < len(el); i++ {
		r = append(r, el[i].String())
	}
	return r
}

func (e *expr) String() string {
	if e.typ == 0 { /* a compound expression */
		switch e.sexp[0].typ {
		case '+', '*', '/', '%', '=', '<', '>', '!', LE, GE, AND, OR:
			if len(e.sexp) != 3 {
				log.Panicf("Expecting binary expression, got %.10q", e.sexp)
			}
			return fmt.Sprintf("%s %s %s", e.sexp[1], e.sexp[0].val, e.sexp[2])
		case '-':
			switch len(e.sexp) {
			case 2:
				return fmt.Sprintf(" -%s", e.sexp[1])
			case 3:
				return fmt.Sprintf("%s - %s", e.sexp[1], e.sexp[2])
			default:
				log.Panicf("Expecting either unary or binary -, got %.10q", e.sexp)
			}
		case '(':
			if len(e.sexp) != 2 {
				log.Panicf("Expecting ( ) as unary operator, got %.10q", e.sexp)
			}
			return fmt.Sprintf("(%s)", e.sexp[1])
		case '[':
			return "[" + strings.Join(e.cdr(), ", ") + "]"
		case NOT:
			return fmt.Sprintf("NOT %s", e.sexp[1])
		case IDENT: /* function call */
			return e.sexp[0].val + "(" + strings.Join(e.cdr(), ", ") + ")"
		}
	} else {
		return fmt.Sprintf("%s", e.val)
	}

	log.Panicf("Cannot print an unknown expression")
	return ""
}

func (s standardSelect) String() string {
	r := "SELECT "
	if len(s.fields) == 0 {
		r += "*"
	} else {
		for i := 0; i < len(s.fields); i++ {
			r += s.fields[i].String()
			if i != len(s.fields)-1 {
				r += ", "
			}
		}
	}
	r += "\nFROM " + strings.Join(s.tables, ", ")
	if s.where != nil {
		r += fmt.Sprintf("\nWHERE %s", s.where)
	}
	if len(s.limit) > 0 {
		r += fmt.Sprintf("\nLIMIT %s", s.limit)
	}
	return r
}

// sqlReentrantParser makes sqlParser, generated by goyacc and using a
// global variable parseResult to return the result, reentrant.
type sqlSyncParser struct {
	pr sqlParser
}

func newParser() *sqlSyncParser {
	return &sqlSyncParser{sqlNewParser()}
}

var mu sync.Mutex

func (p *sqlSyncParser) Parse(s string) (r *extendedSelect, e error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			e, ok = r.(error)
			if !ok {
				e = fmt.Errorf("%v", r)
			}
		}
	}()

	mu.Lock()
	defer mu.Unlock()

	p.pr.Parse(newLexer(s))
	return parseResult, nil
}

//line yacctab:1
var sqlExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const sqlPrivate = 57344

const sqlLast = 181

var sqlAct = [...]int{

	30, 106, 60, 105, 13, 88, 87, 84, 85, 83,
	86, 39, 22, 37, 47, 46, 45, 59, 49, 48,
	50, 40, 41, 42, 43, 44, 100, 85, 93, 85,
	53, 38, 119, 116, 56, 57, 7, 9, 8, 10,
	11, 64, 99, 69, 70, 71, 72, 73, 74, 75,
	76, 77, 78, 79, 80, 81, 67, 117, 117, 24,
	23, 25, 97, 40, 41, 42, 43, 44, 4, 96,
	91, 98, 32, 18, 17, 108, 31, 42, 43, 44,
	27, 66, 92, 33, 55, 28, 29, 54, 94, 107,
	15, 96, 114, 122, 115, 21, 120, 118, 109, 111,
	89, 110, 104, 109, 16, 90, 113, 24, 23, 25,
	68, 65, 35, 34, 20, 36, 112, 61, 109, 121,
	32, 102, 103, 3, 31, 24, 23, 25, 27, 12,
	26, 33, 58, 28, 29, 19, 63, 14, 32, 62,
	6, 101, 31, 95, 5, 2, 27, 1, 0, 33,
	0, 28, 29, 51, 52, 47, 46, 45, 0, 49,
	48, 50, 40, 41, 42, 43, 44, 51, 52, 47,
	46, 45, 82, 49, 48, 50, 40, 41, 42, 43,
	44,
}
var sqlPact = [...]int{

	119, -1000, 31, 73, -1000, 37, 36, 97, 77, 108,
	96, 95, -1000, 99, -27, -7, -1000, -1000, -1000, -29,
	-1000, -1000, 147, -1000, -7, -1000, -1000, 108, 68, 65,
	-1000, 108, 108, 90, 107, 126, 3, 94, 42, 93,
	108, 108, 108, 108, 108, 108, 108, 108, 108, 108,
	108, 108, 108, 133, -34, -37, -1000, -1000, -1000, -32,
	147, 83, 88, 83, 108, -1000, -1000, -11, -1000, 46,
	46, -1000, -1000, -1000, 34, 34, 34, 34, 34, 34,
	-8, -8, -1000, -1000, -1000, 108, -1000, 51, -1000, 47,
	-1000, 29, -13, -1000, 147, 110, 83, 58, 108, 82,
	-1000, 102, 58, 75, -1000, 18, -1000, -1000, -7, -1000,
	147, -1000, 80, 17, -1000, -1000, 79, 58, -1000, 76,
	-1000, -1000, -1000,
}
var sqlPgo = [...]int{

	0, 147, 145, 144, 143, 141, 140, 137, 135, 2,
	0, 1, 17, 130, 3, 129, 5, 6,
}
var sqlR1 = [...]int{

	0, 1, 1, 1, 2, 2, 2, 2, 3, 6,
	6, 4, 4, 4, 15, 15, 7, 7, 7, 11,
	11, 11, 14, 14, 5, 5, 8, 8, 16, 17,
	17, 10, 10, 12, 12, 13, 13, 9, 9, 9,
	9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
	9, 9, 9, 9, 9, 9, 9, 9, 9, 9,
}
var sqlR2 = [...]int{

	0, 2, 3, 3, 2, 3, 3, 3, 8, 4,
	6, 2, 4, 5, 5, 1, 1, 1, 3, 1,
	1, 1, 1, 3, 2, 2, 1, 3, 3, 1,
	3, 3, 4, 1, 3, 2, 3, 1, 1, 1,
	1, 3, 3, 3, 1, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 2, 2,
}
var sqlChk = [...]int{

	-1000, -1, -2, 4, 37, -3, -6, 5, 7, 6,
	8, 9, -15, -10, -7, 17, 31, 37, 37, -8,
	17, 18, -9, 18, 17, 19, -13, 38, 43, 44,
	-10, 34, 30, 41, 17, 17, 16, 40, 38, 40,
	29, 30, 31, 32, 33, 24, 23, 22, 27, 26,
	28, 20, 21, -9, 19, 19, -9, -9, 42, -12,
	-9, 10, 13, 10, 38, 17, 39, -12, 17, -9,
	-9, -9, -9, -9, -9, -9, -9, -9, -9, -9,
	-9, -9, 39, 43, 44, 40, 42, -17, -16, 17,
	17, -17, -12, 39, -9, -4, 40, 11, 24, 13,
	39, -5, 11, 12, -16, -14, -11, 31, 17, -10,
	-9, 17, 14, -14, 17, 19, 15, 40, 17, 15,
	17, -11, 17,
}
var sqlDef = [...]int{

	0, -2, 0, 0, 1, 0, 0, 0, 0, 0,
	0, 0, 4, 0, 15, 17, 16, 2, 3, 5,
	26, 6, 7, 37, 38, 39, 40, 0, 0, 0,
	44, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 58, 59, 35, 0,
	33, 0, 0, 0, 0, 18, 31, 0, 27, 45,
	46, 47, 48, 49, 50, 51, 52, 53, 54, 55,
	56, 57, 41, 42, 43, 0, 36, 0, 29, 0,
	9, 0, 0, 32, 34, 0, 0, 0, 0, 0,
	14, 0, 0, 0, 30, 11, 22, 19, 20, 21,
	28, 10, 0, 0, 24, 25, 0, 0, 8, 0,
	12, 23, 13,
}
var sqlTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 25, 43, 3, 3, 33, 3, 44,
	38, 39, 31, 29, 40, 30, 3, 32, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 37,
	23, 24, 22, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 41, 3, 42,
}
var sqlTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	26, 27, 28, 34, 35, 36,
}
var sqlTok3 = [...]int{
	0,
}

var sqlErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	sqlDebug        = 0
	sqlErrorVerbose = false
)

type sqlLexer interface {
	Lex(lval *sqlSymType) int
	Error(s string)
}

type sqlParser interface {
	Parse(sqlLexer) int
	Lookahead() int
}

type sqlParserImpl struct {
	lval  sqlSymType
	stack [sqlInitialStackSize]sqlSymType
	char  int
}

func (p *sqlParserImpl) Lookahead() int {
	return p.char
}

func sqlNewParser() sqlParser {
	return &sqlParserImpl{}
}

const sqlFlag = -1000

func sqlTokname(c int) string {
	if c >= 1 && c-1 < len(sqlToknames) {
		if sqlToknames[c-1] != "" {
			return sqlToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func sqlStatname(s int) string {
	if s >= 0 && s < len(sqlStatenames) {
		if sqlStatenames[s] != "" {
			return sqlStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func sqlErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !sqlErrorVerbose {
		return "syntax error"
	}

	for _, e := range sqlErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + sqlTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := sqlPact[state]
	for tok := TOKSTART; tok-1 < len(sqlToknames); tok++ {
		if n := base + tok; n >= 0 && n < sqlLast && sqlChk[sqlAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if sqlDef[state] == -2 {
		i := 0
		for sqlExca[i] != -1 || sqlExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; sqlExca[i] >= 0; i += 2 {
			tok := sqlExca[i]
			if tok < TOKSTART || sqlExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if sqlExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += sqlTokname(tok)
	}
	return res
}

func sqllex1(lex sqlLexer, lval *sqlSymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = sqlTok1[0]
		goto out
	}
	if char < len(sqlTok1) {
		token = sqlTok1[char]
		goto out
	}
	if char >= sqlPrivate {
		if char < sqlPrivate+len(sqlTok2) {
			token = sqlTok2[char-sqlPrivate]
			goto out
		}
	}
	for i := 0; i < len(sqlTok3); i += 2 {
		token = sqlTok3[i+0]
		if token == char {
			token = sqlTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = sqlTok2[1] /* unknown char */
	}
	if sqlDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", sqlTokname(token), uint(char))
	}
	return char, token
}

func sqlParse(sqllex sqlLexer) int {
	return sqlNewParser().Parse(sqllex)
}

func (sqlrcvr *sqlParserImpl) Parse(sqllex sqlLexer) int {
	var sqln int
	var sqlVAL sqlSymType
	var sqlDollar []sqlSymType
	_ = sqlDollar // silence set and not used
	sqlS := sqlrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	sqlstate := 0
	sqlrcvr.char = -1
	sqltoken := -1 // sqlrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		sqlstate = -1
		sqlrcvr.char = -1
		sqltoken = -1
	}()
	sqlp := -1
	goto sqlstack

ret0:
	return 0

ret1:
	return 1

sqlstack:
	/* put a state and value onto the stack */
	if sqlDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", sqlTokname(sqltoken), sqlStatname(sqlstate))
	}

	sqlp++
	if sqlp >= len(sqlS) {
		nyys := make([]sqlSymType, len(sqlS)*2)
		copy(nyys, sqlS)
		sqlS = nyys
	}
	sqlS[sqlp] = sqlVAL
	sqlS[sqlp].yys = sqlstate

sqlnewstate:
	sqln = sqlPact[sqlstate]
	if sqln <= sqlFlag {
		goto sqldefault /* simple state */
	}
	if sqlrcvr.char < 0 {
		sqlrcvr.char, sqltoken = sqllex1(sqllex, &sqlrcvr.lval)
	}
	sqln += sqltoken
	if sqln < 0 || sqln >= sqlLast {
		goto sqldefault
	}
	sqln = sqlAct[sqln]
	if sqlChk[sqln] == sqltoken { /* valid shift */
		sqlrcvr.char = -1
		sqltoken = -1
		sqlVAL = sqlrcvr.lval
		sqlstate = sqln
		if Errflag > 0 {
			Errflag--
		}
		goto sqlstack
	}

sqldefault:
	/* default state action */
	sqln = sqlDef[sqlstate]
	if sqln == -2 {
		if sqlrcvr.char < 0 {
			sqlrcvr.char, sqltoken = sqllex1(sqllex, &sqlrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if sqlExca[xi+0] == -1 && sqlExca[xi+1] == sqlstate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			sqln = sqlExca[xi+0]
			if sqln < 0 || sqln == sqltoken {
				break
			}
		}
		sqln = sqlExca[xi+1]
		if sqln < 0 {
			goto ret0
		}
	}
	if sqln == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			sqllex.Error(sqlErrorMessage(sqlstate, sqltoken))
			Nerrs++
			if sqlDebug >= 1 {
				__yyfmt__.Printf("%s", sqlStatname(sqlstate))
				__yyfmt__.Printf(" saw %s\n", sqlTokname(sqltoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for sqlp >= 0 {
				sqln = sqlPact[sqlS[sqlp].yys] + sqlErrCode
				if sqln >= 0 && sqln < sqlLast {
					sqlstate = sqlAct[sqln] /* simulate a shift of "error" */
					if sqlChk[sqlstate] == sqlErrCode {
						goto sqlstack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if sqlDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", sqlS[sqlp].yys)
				}
				sqlp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if sqlDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", sqlTokname(sqltoken))
			}
			if sqltoken == sqlEofCode {
				goto ret1
			}
			sqlrcvr.char = -1
			sqltoken = -1
			goto sqlnewstate /* try again in the same state */
		}
	}

	/* reduction by production sqln */
	if sqlDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", sqln, sqlStatname(sqlstate))
	}

	sqlnt := sqln
	sqlpt := sqlp
	_ = sqlpt // guard against "declared and not used"

	sqlp -= sqlR2[sqln]
	// sqlp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if sqlp+1 >= len(sqlS) {
		nyys := make([]sqlSymType, len(sqlS)*2)
		copy(nyys, sqlS)
		sqlS = nyys
	}
	sqlVAL = sqlS[sqlp+1]

	/* consult goto table to find next state */
	sqln = sqlR1[sqln]
	sqlg := sqlPgo[sqln]
	sqlj := sqlg + sqlS[sqlp].yys + 1

	if sqlj >= sqlLast {
		sqlstate = sqlAct[sqlg]
	} else {
		sqlstate = sqlAct[sqlj]
		if sqlChk[sqlstate] != -sqln {
			sqlstate = sqlAct[sqlg]
		}
	}
	// dummy call; replaced with literal code
	switch sqlnt {

	case 1:
		sqlDollar = sqlS[sqlpt-2 : sqlpt+1]
//line sql.y:148
		{
			parseResult = &extendedSelect{
				extended:       false,
				standardSelect: sqlDollar[1].slct}
		}
	case 2:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:153
		{
			parseResult = &extendedSelect{
				extended:       true,
				train:          true,
				standardSelect: sqlDollar[1].slct,
				trainClause:    sqlDollar[2].tran}
		}
	case 3:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:160
		{
			parseResult = &extendedSelect{
				extended:       true,
				train:          false,
				standardSelect: sqlDollar[1].slct,
				predictClause:  sqlDollar[2].infr}
		}
	case 4:
		sqlDollar = sqlS[sqlpt-2 : sqlpt+1]
//line sql.y:170
		{
			sqlVAL.slct.fields = sqlDollar[2].expl
		}
	case 5:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:171
		{
			sqlVAL.slct.tables = sqlDollar[3].tbls
		}
	case 6:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:172
		{
			sqlVAL.slct.limit = sqlDollar[3].val
		}
	case 7:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:173
		{
			sqlVAL.slct.where = sqlDollar[3].expr
		}
	case 8:
		sqlDollar = sqlS[sqlpt-8 : sqlpt+1]
//line sql.y:177
		{
			sqlVAL.tran.estimator = sqlDollar[2].val
			sqlVAL.tran.trainAttrs = sqlDollar[4].atrs
			sqlVAL.tran.columns = sqlDollar[5].colc
			sqlVAL.tran.label = sqlDollar[6].labc
			sqlVAL.tran.save = sqlDollar[8].val
		}
	case 9:
		sqlDollar = sqlS[sqlpt-4 : sqlpt+1]
//line sql.y:187
		{
			sqlVAL.infr.into = sqlDollar[2].val
			sqlVAL.infr.model = sqlDollar[4].val
		}
	case 10:
		sqlDollar = sqlS[sqlpt-6 : sqlpt+1]
//line sql.y:188
		{
			sqlVAL.infr.into = sqlDollar[2].val
			sqlVAL.infr.predAttrs = sqlDollar[4].atrs
			sqlVAL.infr.model = sqlDollar[6].val
		}
	case 11:
		sqlDollar = sqlS[sqlpt-2 : sqlpt+1]
//line sql.y:192
		{
			sqlVAL.colc = map[string]exprlist{"feature_columns": sqlDollar[2].expl}
		}
	case 12:
		sqlDollar = sqlS[sqlpt-4 : sqlpt+1]
//line sql.y:193
		{
			sqlVAL.colc = map[string]exprlist{sqlDollar[4].val: sqlDollar[2].expl}
		}
	case 13:
		sqlDollar = sqlS[sqlpt-5 : sqlpt+1]
//line sql.y:194
		{
			sqlVAL.colc[sqlDollar[5].val] = sqlDollar[3].expl
		}
	case 14:
		sqlDollar = sqlS[sqlpt-5 : sqlpt+1]
//line sql.y:198
		{
			sqlVAL.expl = exprlist{sqlDollar[1].expr, atomic(IDENT, "AS"), funcall("", sqlDollar[4].expl)}
		}
	case 15:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:201
		{
			sqlVAL.expl = sqlDollar[1].flds
		}
	case 16:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:205
		{
			sqlVAL.flds = append(sqlVAL.flds, atomic(IDENT, "*"))
		}
	case 17:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:206
		{
			sqlVAL.flds = append(sqlVAL.flds, atomic(IDENT, sqlDollar[1].val))
		}
	case 18:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:207
		{
			sqlVAL.flds = append(sqlDollar[1].flds, atomic(IDENT, sqlDollar[3].val))
		}
	case 19:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:211
		{
			sqlVAL.expr = atomic(IDENT, "*")
		}
	case 20:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:212
		{
			sqlVAL.expr = atomic(IDENT, sqlDollar[1].val)
		}
	case 21:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:213
		{
			sqlVAL.expr = sqlDollar[1].expr
		}
	case 22:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:217
		{
			sqlVAL.expl = exprlist{sqlDollar[1].expr}
		}
	case 23:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:218
		{
			sqlVAL.expl = append(sqlDollar[1].expl, sqlDollar[3].expr)
		}
	case 24:
		sqlDollar = sqlS[sqlpt-2 : sqlpt+1]
//line sql.y:222
		{
			sqlVAL.labc = sqlDollar[2].val
		}
	case 25:
		sqlDollar = sqlS[sqlpt-2 : sqlpt+1]
//line sql.y:223
		{
			sqlVAL.labc = sqlDollar[2].val[1 : len(sqlDollar[2].val)-1]
		}
	case 26:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:227
		{
			sqlVAL.tbls = []string{sqlDollar[1].val}
		}
	case 27:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:228
		{
			sqlVAL.tbls = append(sqlDollar[1].tbls, sqlDollar[3].val)
		}
	case 28:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:232
		{
			sqlVAL.atrs = attrs{sqlDollar[1].val: sqlDollar[3].expr}
		}
	case 29:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:236
		{
			sqlVAL.atrs = sqlDollar[1].atrs
		}
	case 30:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:237
		{
			sqlVAL.atrs = attrsUnion(sqlDollar[1].atrs, sqlDollar[3].atrs)
		}
	case 31:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:241
		{
			sqlVAL.expr = funcall(sqlDollar[1].val, nil)
		}
	case 32:
		sqlDollar = sqlS[sqlpt-4 : sqlpt+1]
//line sql.y:242
		{
			sqlVAL.expr = funcall(sqlDollar[1].val, sqlDollar[3].expl)
		}
	case 33:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:246
		{
			sqlVAL.expl = exprlist{sqlDollar[1].expr}
		}
	case 34:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:247
		{
			sqlVAL.expl = append(sqlDollar[1].expl, sqlDollar[3].expr)
		}
	case 35:
		sqlDollar = sqlS[sqlpt-2 : sqlpt+1]
//line sql.y:251
		{
			sqlVAL.expl = nil
		}
	case 36:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:252
		{
			sqlVAL.expl = sqlDollar[2].expl
		}
	case 37:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:256
		{
			sqlVAL.expr = atomic(NUMBER, sqlDollar[1].val)
		}
	case 38:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:257
		{
			sqlVAL.expr = atomic(IDENT, sqlDollar[1].val)
		}
	case 39:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:258
		{
			sqlVAL.expr = atomic(STRING, sqlDollar[1].val)
		}
	case 40:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:259
		{
			sqlVAL.expr = variadic('[', "square", sqlDollar[1].expl)
		}
	case 41:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:260
		{
			sqlVAL.expr = unary('(', "paren", sqlDollar[2].expr)
		}
	case 42:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:261
		{
			sqlVAL.expr = unary('"', "quota", atomic(STRING, sqlDollar[2].val))
		}
	case 43:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:262
		{
			sqlVAL.expr = unary('\'', "quota", atomic(STRING, sqlDollar[2].val))
		}
	case 44:
		sqlDollar = sqlS[sqlpt-1 : sqlpt+1]
//line sql.y:263
		{
			sqlVAL.expr = sqlDollar[1].expr
		}
	case 45:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:264
		{
			sqlVAL.expr = binary('+', sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 46:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:265
		{
			sqlVAL.expr = binary('-', sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 47:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:266
		{
			sqlVAL.expr = binary('*', sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 48:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:267
		{
			sqlVAL.expr = binary('/', sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 49:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:268
		{
			sqlVAL.expr = binary('%', sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 50:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:269
		{
			sqlVAL.expr = binary('=', sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 51:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:270
		{
			sqlVAL.expr = binary('<', sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 52:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:271
		{
			sqlVAL.expr = binary('>', sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 53:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:272
		{
			sqlVAL.expr = binary(LE, sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 54:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:273
		{
			sqlVAL.expr = binary(GE, sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 55:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:274
		{
			sqlVAL.expr = binary(NE, sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 56:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:275
		{
			sqlVAL.expr = binary(AND, sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 57:
		sqlDollar = sqlS[sqlpt-3 : sqlpt+1]
//line sql.y:276
		{
			sqlVAL.expr = binary(OR, sqlDollar[1].expr, sqlDollar[2].val, sqlDollar[3].expr)
		}
	case 58:
		sqlDollar = sqlS[sqlpt-2 : sqlpt+1]
//line sql.y:277
		{
			sqlVAL.expr = unary(NOT, sqlDollar[1].val, sqlDollar[2].expr)
		}
	case 59:
		sqlDollar = sqlS[sqlpt-2 : sqlpt+1]
//line sql.y:278
		{
			sqlVAL.expr = unary('-', sqlDollar[1].val, sqlDollar[2].expr)
		}
	}
	goto sqlstack /* stack new state and value */
}
