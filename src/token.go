package glox

import (
	"fmt"
)

type TokenKind uint

const (
	ILLEGAL TokenKind = iota
	EOF

	// identifier
	IDENT // variable name

	// Lit
	STR_LIT
	BOOL_LIT
	INT_LIT
	FLOAT_LIT

	// operator
	TILDE // ~
	// unary +-
	NOT // !

	// left to right
	MUL    // *
	DIV    // /
	MOD    // %
	ADD    // +
	SUB    // -
	LSH    // <<
	RSH    // >>
	LSS    // <
	LEQ    // <=
	GTR    // >
	GEQ    // >=
	EQL    // ==
	NEQ    // !=
	AND    // &
	XOR    // ^
	OR     // |
	LAND   // &&
	LOR    // ||
	ASSIGN // =

	// AssignOp // Op=
	MUL_ASSIGN // *=
	DIV_ASSIGN // /=
	MOD_ASSIGN // %=

	ADD_ASSIGN // +=
	SUB_ASSIGN // -=

	LSH_ASSIGN // <<=
	RSH_ASSIGN // >>=

	AND_ASSIGN // &=
	XOR_ASSIGN // ^=
	OR_ASSIGN  // |=

	// delimiters
	LPAREN // (
	RPAREN // )
	LBRACE // {
	RBRACE // }

	SEMI  // ;
	COMMA // ,

	// keywords
	LET      // let
	FUNC     // fn
	RETURN   // return
	IF       // if
	ELSE     // else
	WHILE    // while
	BREAK    // break
	CONTINUE // continue

	// builtin
	PRINT // print
)

func (tk TokenKind) String() string {
	strs := []string{
		"ILLEGAL",
		"EOF",
		"IDENT",
		"STR_LIT",
		"BOOL_LIT",
		"INT_LIT",
		"FLOAT_LIT",

		"TILDE",

		"NOT",
		"MUL",
		"DIV",
		"MOD",
		"ADD",
		"SUB",
		"LSH",
		"RSH",
		"LSS",
		"LEQ",
		"GTR",
		"GEQ",
		"EQL",
		"NEQ",
		"AND",
		"XOR",
		"OR",
		"LAND",
		"LOR",
		"ASSIGN",

		"MUL_ASSIGN",
		"DIV_ASSIGN",
		"MOD_ASSIGN",

		"ADD_ASSIGN",
		"SUB_ASSIGN",

		"LSH_ASSIGN",
		"RSH_ASSIGN",

		"AND_ASSIGN",
		"XOR_ASSIGN",
		"OR_ASSIGN",

		"LPAREN",
		"RPAREN",

		"LBRACE",
		"RBRACE",

		"SEMI",
		"COMMA",

		"LET",
		"FUNC",
		"RETURN",
		"IF",
		"ELSE",
		"WHILE",

		"BREAK",
		"CONTINUE",

		"PRINT",
	}
	if int(tk) < len(strs) {
		return strs[tk]
	}
	return "ERROR: TokenKind not found"
}

func (tk TokenKind) IsOp() bool {
	switch tk {
	case TILDE, NOT, MUL, DIV, MOD, ADD, SUB, LSH, RSH, LSS, LEQ, GTR, GEQ, EQL, NEQ, AND, XOR, OR, LAND, LOR:
		return true
	}
	return false
}

func (tk TokenKind) IsBuiltin() bool {
	switch tk {
	case PRINT:
		return true
	}
	return false
}

var Keywords = map[string]TokenKind{
	"let":      LET,
	"fn":       FUNC,
	"return":   RETURN,
	"if":       IF,
	"else":     ELSE,
	"while":    WHILE,
	"break":    BREAK,
	"continue": CONTINUE,
}

var Builtins = map[string]TokenKind{
	"print": PRINT,
}

const (
	LOWEST_PREC  = 0 // non-operators
	UNARY_PREC   = 6
	HIGHEST_PREC = 7
)

// precedence order copied from golang
func (op TokenKind) Prec() int {
	switch op {
	case LOR:
		return 1
	case LAND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB, OR, XOR:
		return 4
	case MUL, DIV, MOD, LSH, RSH, AND:
		return 5
	}
	return LOWEST_PREC
}

type Token struct {
	Kind TokenKind
	Lit  *string
	Line int
}

func (tk Token) String() string {
	return fmt.Sprintf("Token{%v, \"%v\", %v}", tk.Kind, *tk.Lit, tk.Line)
}

func NewToken(kind TokenKind, value *string, line int) Token {
	return Token{
		Kind: kind,
		Lit:  value,
		Line: line,
	}
}
