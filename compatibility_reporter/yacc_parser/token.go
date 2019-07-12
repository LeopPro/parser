package yacc_parser

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
)

const delimiter string = "|:"
const (
	or    = '|'
	colon = ':'
)

type token interface {
	toString() string
}
type eof struct{}

func (*eof) toString() string {
	return "EOF"
}

type operator struct {
	val string
}

func (op *operator) toString() string {
	return op.val
}

type keyword struct {
	val string
}

func (kw *keyword) toString() string {
	return kw.val
}

type nonTerminal struct {
	val string
}

func (nt *nonTerminal) toString() string {
	return nt.val
}

func Tokenize(reader *bufio.Reader) func() token {
	return func() token {
		var r rune
		var err error
		// Skip spaces.
		for {
			r, _, err = reader.ReadRune()
			panicIfNonEOF(err)
			if err == io.EOF {
				return &eof{}
			}
			if !unicode.IsSpace(r) {
				break
			}
		}

		// Handle delimiter.
		if r == ':' || r == '|' {
			return &operator{string(r)}
		}

		// Handle identifier.
		stringBuf := string(r)
		for {
			r, _, err = reader.ReadRune()
			panicIfNonEOF(err)
			if err == io.EOF || unicode.IsSpace(r) || isDelimiter(r) {
				reader.UnreadRune()
				break
			}
			stringBuf += string(r)
		}
		if allCapital(stringBuf) {
			return &keyword{stringBuf}
		} else {
			return &nonTerminal{stringBuf}
		}
	}
}

func panicIfNonEOF(err error) {
	if err != nil && err != io.EOF {
		panic(fmt.Sprintf("unknown error: %v", err))
	}
}

func isDelimiter(r rune) bool {
	return r == '|' || r == ':'
}

func allCapital(str string) bool {
	for _, c := range str {
		if !unicode.IsUpper(c) {
			return false
		}
	}
	return true
}

func isEOF(tkn token) bool {
	_, ok := tkn.(*eof)
	return ok
}
