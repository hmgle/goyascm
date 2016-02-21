package main

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

type Env struct {
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
	Env
	Func func(env, args *Object) *Object
}

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

func main() {
}
