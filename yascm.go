package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	FIXNUM int = iota
	FLOATNUM
	BOOL
	CHAR
	STRING
	PAIR
	SYMBOL
	KEYWORD
	PRIM
	COMPOUND_PROC
	ENV
	OTHER
)

type Pair struct {
	Car *Object
	Cdr *Object
}

type CompoundProc struct {
	Parameters *Object
	Body       *Object
	Env        *Object
}

type EnvFrame struct {
	Vars *Object
	Up   *Object
}

// TODO: As Go has no union, it waste memory
type Object struct {
	Type     int
	IntVal   int64
	FloatVal float64
	BoolVal  bool
	CharVal  rune
	StrVal   string
	Pair
	CompoundProc
	EnvFrame
	Func Primitive
}

type Primitive func(env, args *Object) *Object

var (
	NOT_END bool   = true
	FALSE   Object = Object{
		Type:    BOOL,
		BoolVal: false,
	}
	TRUE Object = Object{
		Type:    BOOL,
		BoolVal: true,
	}
	NIL         *Object = &Object{}
	ELSE        *Object = &Object{}
	OK          *Object = &Object{}
	UNSPECIFIED *Object = &Object{}
	SymbolTable *Object
)

func eof_handle() {
	NOT_END = false
}

func createObject(typ int) *Object {
	return &Object{Type: typ}
}

func car(pair *Object) *Object {
	return pair.Car
}

func cdr(pair *Object) *Object {
	return pair.Cdr
}

func cadr(pair *Object) *Object {
	return car(cdr(pair))
}

func cdar(pair *Object) *Object {
	return cdr(car(pair))
}

func caar(pair *Object) *Object {
	return car(car(pair))
}

func caddr(pair *Object) *Object {
	return car(cdr(cdr(pair)))
}

func cons(car, cdr *Object) *Object {
	pair := createObject(PAIR)
	pair.Car = car
	pair.Cdr = cdr
	return pair
}

/* return ((x . y) . a) */
func acons(x, y, a *Object) *Object {
	return cons(cons(x, y), a)
}

func addVariable(env, sym, val *Object) {
	env.Vars = acons(sym, val, env.Vars)
}

func makeBool(val bool) *Object {
	if val {
		return &TRUE
	}
	return &FALSE
}

func makeChar(val rune) *Object {
	obj := createObject(CHAR)
	obj.CharVal = val
	return obj
}

func makeString(val string) *Object {
	obj := createObject(STRING)
	obj.StrVal = val
	return obj
}

func makeFixnum(val int64) *Object {
	obj := createObject(FIXNUM)
	obj.IntVal = val
	return obj
}

func makeEmptylist() *Object {
	return NIL
}

func _makeSymbol(name string) *Object {
	obj := createObject(SYMBOL)
	obj.StrVal = name
	return obj
}

func makeSymbol(name string) *Object {
	var p *Object
	for p = SymbolTable; p != NIL; p = p.Cdr {
		if strings.Compare(name, p.Car.StrVal) == 0 {
			return p.Car
		}
	}
	sym := _makeSymbol(name)
	SymbolTable = cons(sym, SymbolTable)
	return sym
}

func makeQuote(obj *Object) *Object {
	return cons(makeSymbol("quote"), cons(obj, NIL))
}

func makeFunction(parameters, body, env *Object) *Object {
	function := createObject(COMPOUND_PROC)
	function.Parameters = parameters
	function.Body = body
	function.Env = env
	return function
}

func lookupVariableVal(v, env *Object) *Object {
	var p, cell *Object
	for p = env; p != nil; p = p.Up {
		for cell = p.Vars; cell != NIL; cell = cell.Cdr {
			bind := cell.Car
			if v == bind.Car {
				return bind
			}
		}
	}
	return nil
}

func listOfVal(args, env *Object) *Object {
	if args == NIL {
		return NIL
	}
	return cons(eval(env, car(args)), listOfVal(cdr(args), env))
}

func apply(env, fn, args *Object) *Object {
	var evalArgs *Object
	if args != NIL && args.Type != PAIR {
		panic("args must be a list")
	}
	if fn.Type == PRIM {
		evalArgs = listOfVal(args, env)
		return fn.Func(env, evalArgs)
	} else if fn.Type == KEYWORD {
		return fn.Func(env, args)
	}
	log.Println("error")
	return nil
}

func makeEnv(v, up *Object) *Object {
	env := createObject(ENV)
	env.Vars = v
	env.Up = up
	return env
}

func extendEnv(vars, vals, baseEnv *Object) *Object {
	newEnv := makeEnv(NIL, baseEnv)
	for ; vars != NIL && vals != NIL; vars, vals = vars.Cdr, vals.Cdr {
		if vars.Type == SYMBOL {
			addVariable(newEnv, vars, vals)
			return newEnv
		}
		addVariable(newEnv, vars.Car, vals.Car)
	}
	return newEnv
}

func isTheLastArg(args *Object) bool {
	if args.Cdr == NIL {
		return true
	}
	return false
}

func eval(env, obj *Object) *Object {
	var bind, fn, args, newEnv, newObj, evalArgs *Object
	if obj == NIL {
		return NIL
	}
	switch obj.Type {
	case SYMBOL:
		bind = lookupVariableVal(obj, env)
		if bind == nil {
			return NIL
		}
		return bind.Cdr
	case PAIR:
		fn = eval(env, obj.Car)
		args = obj.Cdr
		if fn.Type == PRIM || fn.Type == KEYWORD {
			return apply(env, fn, args)
		} else if fn.Type == COMPOUND_PROC {
			evalArgs = listOfVal(args, env)
			newEnv = extendEnv(fn.Parameters, evalArgs, fn.Env)
			for newObj = fn.Body; !isTheLastArg(newObj); newObj = newObj.Cdr {
				eval(newEnv, newObj.Car)
			}
			return eval(newEnv, newObj.Car)
		} else {
			return obj // list
		}
	default:
		return obj
	}
}

func pairPrint(pair *Object) {
	carPair, cdrPair := pair.Car, pair.Cdr
	objectPrint(carPair)
	if cdrPair == NIL {
		return
	} else if cdrPair.Type == PAIR {
		print(" ")
		pairPrint(cdrPair)
	} else {
		print(" . ")
		objectPrint(cdrPair)
	}
}

func objectPrint(obj *Object) {
	if obj == nil {
		return
	}
	switch obj.Type {
	case FIXNUM:
		print(obj.IntVal)
	case KEYWORD:
		print("<keyword>")
	case PRIM:
		print("<primitive>")
	case BOOL:
		if obj.BoolVal {
			print("#t")
		} else {
			print("#f")
		}
	case CHAR:
		fmt.Printf("#\\%c", obj.CharVal)
	case STRING:
		fmt.Printf("\"%s\"", obj.StrVal)
	case SYMBOL:
		print(obj.StrVal)
	case PAIR:
		print("(")
		pairPrint(obj)
		print(")")
	case COMPOUND_PROC:
		print("<proc>")
	default:
		if obj == NIL {
			print("()")
		} else if obj == OK {
			fmt.Fprint(os.Stderr, "; ok")
		} else if obj == UNSPECIFIED {
		} else {
			fmt.Printf("type: %d", obj.Type)
		}
	}
}

func addPrimitive(env *Object, name string, fun Primitive, typ int) {
	sym := makeSymbol(name)
	prim := createObject(typ)
	prim.Func = fun
	addVariable(env, sym, prim)
}

func primPlus(env, args *Object) *Object {
	var ret int64
	for ret = 0; args != NIL; args = args.Cdr {
		ret += car(args).IntVal
	}
	return makeFixnum(ret)
}

func listLength(list *Object) int {
	length := int(0)
	for {
		if list == NIL {
			return length
		}
		list = list.Cdr
		length += 1
	}
}

func primSub(env, args *Object) *Object {
	var ret int64
	if listLength(args) == 1 {
		return makeFixnum(-car(args).IntVal)
	}
	ret = car(args).IntVal
	for args = args.Cdr; args != NIL; args = args.Cdr {
		ret -= car(args).IntVal
	}
	return makeFixnum(ret)
}

func primMul(env, args *Object) *Object {
	var ret int64
	for ret = 1; args != NIL; args = args.Cdr {
		ret *= car(args).IntVal
	}
	return makeFixnum(ret)
}

func primQuotient(env, args *Object) *Object {
	return makeFixnum(args.Car.IntVal / cadr(args).IntVal)
}

func primCons(env, args *Object) *Object {
	return cons(car(args), cadr(args))
}

func primCar(env, args *Object) *Object {
	return caar(args)
}

func primCdr(env, args *Object) *Object {
	return cdar(args)
}

func primSetCar(env, args *Object) *Object {
	pair := car(args)
	pair.Car = cadr(args)
	return UNSPECIFIED
}

func primSetCdr(env, args *Object) *Object {
	pair := car(args)
	pair.Cdr = cadr(args)
	return UNSPECIFIED
}

func primList(env, args *Object) *Object {
	return args
}

func primQuote(env, args *Object) *Object {
	if listLength(args) != 1 {
		panic("quote")
	}
	return args.Car
}

func defVar(args *Object) *Object {
	if args.Car.Type == SYMBOL {
		return car(args)
	}
	return caar(args) // PAIR
}

func defVal(args, env *Object) *Object {
	if car(args).Type == PAIR {
		return makeFunction(cdar(args), cdr(args), env)
	}
	return eval(env, cadr(args))
}

func defineVariable(vari, val, env *Object) {
	var cell *Object
	for cell = env.Vars; cell != NIL; cell = cell.Cdr {
		oldvar := cell.Car
		if vari == oldvar.Car {
			oldvar.Cdr = val
			return
		}
	}
	addVariable(env, vari, val)
}

func setVar(args *Object) *Object {
	return car(args)
}

func setVal(args *Object) *Object {
	return cadr(args)
}

func setVarVal(vari, val, env *Object) {
	oldVar := lookupVariableVal(vari, env)
	if oldVar == nil {
		panic("unbound variable")
	}
	oldVar.Cdr = val
}

func primDefine(env, args *Object) *Object {
	defineVariable(defVar(args), defVal(args, env), env)
	return OK
}

func primLambda(env, args *Object) *Object {
	return makeFunction(car(args), cdr(args), env)
}

func primLet(env, args *Object) *Object {
	var parameters, exps *Object
	paraEnd := &parameters
	expsEnd := &exps
	letVarExp := car(args)
	letBody := cdr(args)
	for letVarExp != NIL {
		*paraEnd = cons(caar(letVarExp), NIL)
		paraEnd = &((*paraEnd).Cdr)
		*expsEnd = cons(car(cdar(letVarExp)), NIL)
		expsEnd = &((*expsEnd).Cdr)
		letVarExp = cdr(letVarExp)
	}
	lambda := makeFunction(parameters, letBody, env)
	return eval(env, cons(lambda, exps))
}

func primSet(env, args *Object) *Object {
	setVarVal(setVar(args), eval(env, setVal(args)), env)
	return OK
}

func isFalse(obj *Object) bool {
	if obj.Type != BOOL || obj.BoolVal {
		return false
	}
	return true
}

func isTrue(obj *Object) bool {
	return !isFalse(obj)
}

func primAnd(env, args *Object) *Object {
	var obj *Object
	ret := &TRUE
	for ; args != NIL; args = args.Cdr {
		obj = car(args)
		ret = eval(env, obj)
		if isFalse(ret) {
			return ret
		}
	}
	return ret
}

func primOr(env, args *Object) *Object {
	var obj *Object
	ret := &FALSE
	for ; args != NIL; args = args.Cdr {
		obj = car(args)
		ret = eval(env, obj)
		if isTrue(ret) {
			return ret
		}
	}
	return ret
}

func primBegin(env, args *Object) *Object {
	var obj, ret *Object
	for ; args != NIL; args = args.Cdr {
		obj = car(args)
		ret = eval(env, obj)
	}
	return ret
}

func primIf(env, args *Object) *Object {
	predicate := eval(env, car(args))
	if isTrue(predicate) {
		return eval(env, cadr(args))
	}
	if isTheLastArg(args.Cdr) {
		return UNSPECIFIED
	}
	return eval(env, caddr(args))
}

func isElse(sym *Object) bool {
	if sym == ELSE {
		return true
	}
	return false
}

func evalArgsList(env, argsList *Object) *Object {
	var arg *Object
	for arg = argsList; !isTheLastArg(arg); arg = arg.Cdr {
		eval(env, arg.Car)
	}
	return eval(env, arg.Car)
}

func primCond(env, args *Object) *Object {
	var predicate *Object
	pairs := args
	for pairs = args; pairs != NIL; pairs = pairs.Cdr {
		predicate = eval(env, caar(pairs))
		if !isTheLastArg(pairs) {
			if predicate.BoolVal {
				return evalArgsList(env, cdar(pairs))
			}
		} else {
			if isElse(predicate) || predicate.BoolVal {
				return evalArgsList(env, cdar(pairs))
			}
		}
	}
	return UNSPECIFIED
}

func primIsNull(env, args *Object) *Object {
	if args.Car == NIL {
		return makeBool(true)
	}
	return makeBool(false)
}

func primIsX(args *Object, typ ...int) *Object {
	for _, t := range typ {
		if args.Car.Type == t {
			return makeBool(true)
		}
	}
	return makeBool(false)
}

func primIsBoolean(env, args *Object) *Object {
	return primIsX(args, BOOL)
}

func primIsPair(env, args *Object) *Object {
	return primIsX(args, PAIR)
}

func primIsSymbol(env, args *Object) *Object {
	return primIsX(args, SYMBOL)
}

func primIsNumber(env, args *Object) *Object {
	return primIsX(args, FIXNUM, FLOATNUM)
}

func primIsChar(env, args *Object) *Object {
	return primIsX(args, CHAR)
}

func primIsString(env, args *Object) *Object {
	return primIsX(args, STRING)
}

func primIsProcedure(env, args *Object) *Object {
	return primIsX(args, COMPOUND_PROC, PRIM)
}

func primIsEq(env, args *Object) *Object {
	var obj *Object
	first := args.Car
	for ; args != NIL; args = args.Cdr {
		obj = args.Car
		if obj.Type != first.Type {
			return makeBool(false)
		}
		switch first.Type {
		case FIXNUM:
			if obj.IntVal != first.IntVal {
				return makeBool(false)
			}
		case BOOL:
			if obj.BoolVal != first.BoolVal {
				return makeBool(false)
			}
		case CHAR:
			if obj.CharVal != first.CharVal {
				return makeBool(false)
			}
		case STRING:
			if obj.StrVal != first.StrVal {
				return makeBool(false)
			}
		default:
			if obj != first {
				return makeBool(false)
			}
		}
	}
	return makeBool(true)
}

func primIsNumEq(env, args *Object) *Object {
	var next, first int64
	if args == NIL {
		return makeBool(true)
	}
	first = args.Car.IntVal
	for args = args.Cdr; args != NIL; first, args = next, args.Cdr {
		next = args.Car.IntVal
		if next != first {
			return makeBool(false)
		}
	}
	return makeBool(true)
}

func primIsNumGt(env, args *Object) *Object {
	var next, first int64
	if args == NIL {
		return makeBool(true)
	}
	first = args.Car.IntVal
	for args = args.Cdr; args != NIL; first, args = next, args.Cdr {
		next = args.Car.IntVal
		if next >= first {
			return makeBool(false)
		}
	}
	return makeBool(true)
}

func loadFile(fileName string, env *Object) *Object {
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fopen() %s fail: %v\n", fileName, err)
		return NIL
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	yascmLex := &scmLex{
		input: *reader,
	}
	for NOT_END {
		scmParse(yascmLex)
		obj := eval(env, yascmLex.obj)
		if obj == nil || obj == NIL {
			break
		}
		objectPrint(obj)
	}
	fmt.Fprintf(os.Stderr, "; done loading %s\n", fileName)
	return OK
}

func primEval(env, args *Object) *Object {
	return eval(env, car(args))
}

func primLoad(env, args *Object) *Object {
	return loadFile(car(args).StrVal, env)
}

// TODO
// func primRead(env, args *Object) *Object {
// }

func primDisplay(env, args *Object) *Object {
	if args.Car.Type == STRING {
		print(args.Car.StrVal)
	} else {
		objectPrint(car(args))
	}
	return UNSPECIFIED
}

func primNewline(env, args *Object) *Object {
	println()
	return OK
}

func definePrim(env *Object) {
	addPrimitive(env, "define", primDefine, KEYWORD)
	addPrimitive(env, "lambda", primLambda, KEYWORD)
	addPrimitive(env, "let", primLet, KEYWORD)
	addPrimitive(env, "set!", primSet, KEYWORD)
	addPrimitive(env, "and", primAnd, KEYWORD)
	addPrimitive(env, "or", primOr, KEYWORD)
	addPrimitive(env, "begin", primBegin, KEYWORD)
	addPrimitive(env, "if", primIf, KEYWORD)
	addPrimitive(env, "cond", primCond, KEYWORD)
	addPrimitive(env, "quote", primQuote, KEYWORD)

	addPrimitive(env, "+", primPlus, PRIM)
	addPrimitive(env, "-", primSub, PRIM)
	addPrimitive(env, "*", primMul, PRIM)
	addPrimitive(env, "quotient", primQuotient, PRIM)
	addPrimitive(env, "cons", primCons, PRIM)
	addPrimitive(env, "car", primCar, PRIM)
	addPrimitive(env, "cdr", primCdr, PRIM)
	addPrimitive(env, "set-car!", primSetCar, PRIM)
	addPrimitive(env, "set-cdr!", primSetCdr, PRIM)
	addPrimitive(env, "list", primList, PRIM)
	addPrimitive(env, "null?", primIsNull, PRIM)
	addPrimitive(env, "boolean?", primIsBoolean, PRIM)
	addPrimitive(env, "pair?", primIsPair, PRIM)
	addPrimitive(env, "symbol?", primIsSymbol, PRIM)
	addPrimitive(env, "number?", primIsNumber, PRIM)
	addPrimitive(env, "char?", primIsChar, PRIM)
	addPrimitive(env, "string?", primIsString, PRIM)
	addPrimitive(env, "procedure?", primIsProcedure, PRIM)
	addPrimitive(env, "eq?", primIsEq, PRIM)
	addPrimitive(env, "=", primIsNumEq, PRIM)
	addPrimitive(env, ">", primIsNumGt, PRIM)
	addPrimitive(env, "eval", primEval, PRIM)
	addPrimitive(env, "load", primLoad, PRIM)
	// addPrimitive(env, "read", primRead, PRIM)
	addPrimitive(env, "display", primDisplay, PRIM)
	addPrimitive(env, "newline", primNewline, PRIM)
}

func main() {
	NIL = createObject(OTHER)
	ELSE = createObject(OTHER)
	OK = createObject(OTHER)
	UNSPECIFIED = createObject(OTHER)
	genv := makeEnv(NIL, nil)
	SymbolTable = NIL
	definePrim(genv)
	addVariable(genv, makeSymbol("else"), ELSE)
	loadFile("stdlib.scm", genv)
	fmt.Fprint(os.Stderr, "welcome\n> ")
	reader := bufio.NewReader(os.Stdin)
	yascmLex := &scmLex{
		input: *reader,
	}
	NOT_END = true
	for NOT_END {
		scmParse(yascmLex)
		objectPrint(eval(genv, yascmLex.obj))
		fmt.Printf("\n> ")
	}
}
