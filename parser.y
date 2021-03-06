%{
package main

import (
	"fmt"
	"log"
	"bufio"
	"io"
)
%}

%union{
	v *Object
	n int64
	d float64
	c rune
	s string
}

%token <v> LP
%token RP
%token DOT
%token QUOTE
%token <n> FIXNUM_T
%token <d> FLOATNUM_T
%token <v> FALSE_T
%token <v> TRUE_T
%token <c> CHAR_T
%token <s> STRING_T
%token DOUBLE_QUOTE
%token <s> SYMBOL_T
%token END_OF_FILE

%type <s> string
%type <v> object
%type <n> number
%type <v> emptylist
%type <v> quote_list
%type <v> pair
%type <v> list_item
%type <v> list_end
%type <v> list

%start top

%%

top: object {
		scmlex.(*scmLex).obj = $1;
		return 0
	}
	| END_OF_FILE {eof_handle(); return 0}

string: DOUBLE_QUOTE STRING_T DOUBLE_QUOTE {$$ = $2;}
	| DOUBLE_QUOTE DOUBLE_QUOTE {$$ = "\"";}

number: FIXNUM_T {$$ = $1;}
	| FLOATNUM_T {print("float: not support now");}

emptylist: LP RP 

quote_list: QUOTE object {$$ = makeQuote($2);}

pair: object DOT object {$$ = cons($1, $3);}

list_item: object {$$ = cons($1, makeEmptylist());}
	| object list_item {$$ = cons($1, $2);}
	| pair {$$ = $1;}

list_end: list_item RP {$$ = $1;}

list: LP list_end {$$ = $2;}

object: TRUE_T		{$$ = makeBool(true);}
	| FALSE_T	{$$ = makeBool(false);}
	| number	{$$ = makeFixnum($1);}
	| CHAR_T	{$$ = makeChar($1);}
	| string	{$$ = makeString($1);}
	| SYMBOL_T	{$$ = makeSymbol($1);}
	| emptylist	{$$ = makeEmptylist();}
	| quote_list	{$$ = $1;}
	| list		{$$ = $1;}

%%

const (
	Initial = iota
	StringStatus
)

var status = Initial

type token struct {
	tok int
	val interface{}
}

type scmLex struct {
	input bufio.Reader
	obj *Object
}

func (l *scmLex) Lex(lval *scmSymType) int {
	var nextRune rune
	var err error
	for {
		nextRune, _, err = l.input.ReadRune()
		if err == io.EOF {
			return END_OF_FILE
		} else if err != nil {
			fmt.Errorf("syntax error: %v\n", err)
		}
		if status == Initial {
			switch nextRune {
			case '(':
				return LP
			case ')':
				return RP
			case '.':
				return DOT
			case '\'':
				return QUOTE
			case '#':
				return parseBoolOrChar(&l.input, lval)
			case ' ', '\t', '\n': // do nothing with white space
				continue
			case ';': // Scheme comment
				eatLine(&l.input)
				continue
			case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				return parseNumOrSub(nextRune, &l.input, lval)
			case '"':
				status = StringStatus
				return DOUBLE_QUOTE 
			default:
				// FIXME: symbol [!$%&*+\-./:<=>?@^_~[:alnum:]]
				return parseSymbol(nextRune, &l.input, lval)
			}
		} else if status == StringStatus {
			switch nextRune {
			case '"':
				status = Initial
				return DOUBLE_QUOTE 
			default:
				return parseStr(nextRune, &l.input, lval)
			}
		}
	}
	return -1
}

func (l *scmLex) Error(e string) {
	log.Fatal(e)
}

func eatLine(input *bufio.Reader) {
	input.ReadLine()
}

func isDigit(val rune) bool {
	if val >= '0' && val <= '9' {
		return true
	} else {
		return false
	}
}

func isSpace(val rune) bool {
	if val == '\t' || val == '\n' || val == '\r' || val == ' ' {
		return true
	}
	return false
}

// EOF is also a delimiter
func isDelimiter(val rune) bool {
	if isSpace(val) || val == '(' || val == ')' ||
		val == '"' || val == ';' || val == 0 {
		return true
	} else {
		return false
	}
}

func isAlpha(val rune) bool {
	if (val >= 'a' && val <= 'z') ||
		(val >= 'A' && val <= 'Z') {
		return true
	} else {
		return false
	}
}

func isInitial(val rune) bool {
	if isAlpha(val) ||
		val == '*' || val == '/' ||
		val == '+' || val == '-' ||
		val == '>' || val == '<' ||
		val == '=' || val == '?' ||
		val == '%' || val == '&' ||
		val == ':' || val == '~' ||
		val == '!' || val == '^' {
		return true
	} else {
		return false
	}
}

func parseSymbol(c rune, input *bufio.Reader, lval *scmSymType) int {
	var n rune
	buf := string(c)
	for {
		n, _, _ = input.ReadRune()
		if !(isInitial(n) || isDigit(n)) {
			input.UnreadRune()
			break
		}
		buf += string(n)
	}
	lval.s =  buf
	return SYMBOL_T
}

func parseStr(c rune, input *bufio.Reader, lval *scmSymType) int {
	buf := ""
	for {
		n, _, _ := input.ReadRune()
		if n == '\\' {
			buf += string(n)
			n2, _, _ := input.ReadRune()
			buf += string(n2)
			continue
		}
		if n == '"' {
			input.UnreadRune()
			break
		}
		buf += string(n)
	}
	lval.s = string(c) + buf
	return STRING_T
}

func parseNumOrSub(c rune, input *bufio.Reader, lval *scmSymType) int {
	sign := 1
	if c == '-' {
		x, _, _ := input.ReadRune()
		input.UnreadRune()
		if isDelimiter(x) {
			lval.s =  "-"
			return SYMBOL_T
		}
		sign = -1
	} else {
		input.UnreadRune()
	}
	var num int64
	var n rune
	for {
		n, _, _ = input.ReadRune()
		if !isDigit(n) {
			break
		}
		diff := n - '0'
		num = (num * 10) + int64(diff)
	}
	if isDelimiter(n) {
		input.UnreadRune()
	}
	lval.n = num * int64(sign)
	return FIXNUM_T
}

func parseBoolOrChar(input *bufio.Reader, lval *scmSymType) int {
	var nextRune rune
	nextRune, _, _ = input.ReadRune()
	switch nextRune {
	case 'f':
		return FALSE_T
	case 't':
		return TRUE_T
	case '\\':
		c, _, _ := input.ReadRune()
		lval.c = c
		return CHAR_T
	}
	fmt.Errorf("syntax error\n")
	return FALSE_T
}

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}
