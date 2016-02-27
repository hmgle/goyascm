package main

import "bufio"

func ungetc(reader *bufio.Reader) {
	err := reader.UnreadByte()
	if err != nil {
		panic(err)
	}
}

func isSpace(val byte) bool {
	if val == '\t' || val == '\n' || val == '\r' || val == ' ' {
		return true
	}
	return false
}

func eatWhiteSpace(reader *bufio.Reader) {
	for {
		c, err := reader.ReadByte()
		if err != nil {
			break
		}
		if isSpace(c) {
			continue
		} else if c == ';' { // comments are whitespace also
			for {
				v, err := reader.ReadByte()
				if err != nil || v == '\n' {
					break
				}
			}
			continue
		}
		ungetc(reader)
		break
	}

}

// EOF is also a delimiter
func isDelimiter(val byte) bool {
	if isSpace(val) || val == '(' || val == ')' ||
		val == '"' || val == ';' || val == 0 {
		return true
	} else {
		return false
	}
}

func isDigit(val byte) bool {
	if val >= '0' && val <= '9' {
		return true
	} else {
		return false
	}
}

func isAlpha(val byte) bool {
	if (val >= 'a' && val <= 'z') ||
		(val >= 'A' && val <= 'Z') {
		return true
	} else {
		return false
	}
}

func isInitial(val byte) bool {
	if isAlpha(val) ||
		val == '*' || val == '/' ||
		val == '+' || val == '-' ||
		val == '>' || val == '<' ||
		val == '=' || val == '?' ||
		val == '!' {
		return true
	} else {
		return false
	}
}

func peekc(reader *bufio.Reader) byte {
	c, err := reader.Peek(1)
	if err != nil {
		//EOF
		return 0
	}
	return c[0]
}

func readc(reader *bufio.Reader) byte {
	c, _ := reader.ReadByte()
	return c
}

func readChar(reader *bufio.Reader) *Object {
	c, err := reader.ReadByte()
	if err != nil {
		panic("incomplete char literal\n")
	}
	if !isDelimiter(peekc(reader)) {
		panic("character not followed by delimiter\n")
	}
	return makeChar(rune(c))
}

func readPair(reader *bufio.Reader) *Object {
	eatWhiteSpace(reader)
	c := readc(reader)
	if c == ')' {
		return makeEmptylist()
	}
	reader.UnreadByte()
	carObj := read(reader)
	eatWhiteSpace(reader)
	c = readc(reader)
	if c == '.' {
	} else {
		reader.UnreadByte()
		cdrObj := readPair(reader)
		return cons(carObj, cdrObj)
	}
	return nil
}

func read(reader *bufio.Reader) *Object {
	eatWhiteSpace(reader)
	c, _ := reader.ReadByte()
	if c == '#' {
		c, _ := reader.ReadByte()
		switch c {
		case 't':
			return makeBool(true)
		case 'f':
			return makeBool(false)
		case '\\':
			return readChar(reader)
		default:
			panic("unknown boolean or character literal\n")
		}
	} else if isDigit(c) || (c == '-' && (isDigit(peekc(reader)))) {
		//make a number
		sign := 1
		if c == '-' {
			sign = -1
		} else {
			reader.UnreadByte()
		}
		num := 0
		n := c
		for {
			n = readc(reader)
			if !isDigit(n) {
				break
			}
			num = (num * 10) + (int(n) - '0')
		}
		num *= sign
		if isDelimiter(n) {
			reader.UnreadByte()
			return makeFixnum(int64(num))
		} else {
			panic("number not followed by delimiter\n")
		}
	} else if c == '"' { // read a string
		buf := ""
		for {
			n := readc(reader)
			if n == '"' {
				break
			}
			buf += string(n)
		}
		return makeString(buf)
	} else if isInitial(c) {
		n := c
		buf := string(c)
		for {
			n = readc(reader)
			if !(isInitial(n) || isDigit(n)) {
				break
			}
			buf += string(n)
		}
		if isDelimiter(n) {
			ungetc(reader)
			return makeSymbol(buf)
		}
	} else if c == '\'' {
		return makeQuote(read(reader))
	} else if c == '(' {
		return readPair(reader)
	}
	return nil
}
