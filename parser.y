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
   | END_OF_FILE {eof_handle();}

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
      | FALSE_T		{$$ = makeBool(false);}
      | number		{$$ = makeFixnum($1);}
      | CHAR_T		{$$ = makeChar($1);}
      | string		{$$ = makeString($1);}
      | SYMBOL_T	{$$ = makeSymbol($1);}
      | emptylist	{$$ = makeEmptylist();}
      | quote_list	{$$ = $1;}
      | list		{$$ = $1;}

%%

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
	 return parseBool(l.input, lval)
      }
   }
   return -1
}

func (l *scmLex) Error(e string) {
    log.Fatal(e)
}

func parseBool(input bufio.Reader, lval *scmSymType) int {
   var nextRune rune
   nextRune, _, _ =input.ReadRune()
   switch nextRune {
   case 'f':
      lval.v = makeBool(false)
      return FALSE_T
   case 't':
      lval.v = makeBool(true)
      return TRUE_T
   }
   fmt.Errorf("syntax error\n")
   return FALSE_T
}

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}
