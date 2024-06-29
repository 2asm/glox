package glox

import (
	"fmt"
)

type Scanner struct {
	Input []rune
	Index int
	Len   int
	Line  int
}

func NewScanner(input string) *Scanner {
	input_runes := []rune(input)
	return &Scanner{
		Input: input_runes,
		Line:  1,
		Index: 0,
		Len:   len(input_runes),
	}
}

func (sc *Scanner) lookahead(ind int) rune {
	if sc.Index+ind < sc.Len {
		return sc.Input[sc.Index+ind]
	}
	return 0
}

func (sc *Scanner) consume(c rune) {
	ch := sc.lookahead(0)
	if ch != c {
		panic(fmt.Sprintf("illegal character '%v' on line %v", ch, sc.Line))
	}
	sc.Index += 1
	if ch == '\n' {
		sc.Line += 1
	}
}

func (sc *Scanner) readInt() ([]rune, bool) {
	str := []rune{}
	ch := sc.lookahead(0)
	var max_underscore = 0
	var cur_underscore = 0
	var end_char = '_'
	for isDigit(ch) || ch == '_' {
		if ch == '_' {
			cur_underscore += 1
			if cur_underscore > max_underscore {
				max_underscore = cur_underscore
			}
		} else {
			cur_underscore = 0
            str = append(str, ch)
		}
		end_char = ch
		sc.consume(ch)
		ch = sc.lookahead(0)
	}
	if max_underscore > 1 || end_char == '_' {
		return str, false
	}
	return str, true
}

func (sc *Scanner) lexNumber() Token {
	lin := sc.Line
	left, ok := sc.readInt()
	if !ok {
		errorMsg := fmt.Sprintf("invalid number literal %s", string(left))
		return NewToken(ILLEGAL, &errorMsg, lin)
	}
	NumToken := INT_LIT
	if sc.lookahead(0) == rune('.') && isDigit(sc.lookahead(1)) {
		sc.consume(sc.lookahead(0))
		right, ok := sc.readInt()
		if !ok {
			errorMsg := fmt.Sprintf("invalid number literal %s", string(left))
			return NewToken(ILLEGAL, &errorMsg, lin)
		}
		left = append(left, '.')
		left = append(left, right...)
		NumToken = FLOAT_LIT
	}
	value_lit := string(left)
	return NewToken(NumToken, &value_lit, lin)
}

func (sc *Scanner) AssignOp(op, aop TokenKind, lin int) Token {
	sc.consume(sc.lookahead(0))
	if sc.lookahead(0) == '=' {
		sc.consume('=')
		return NewToken(aop, nil, lin)
	}
	return NewToken(op, nil, lin)
}

func (sc *Scanner) Next() Token {
	for isWhitespace(sc.lookahead(0)) {
		sc.consume(sc.lookahead(0))
	}

	lin := sc.Line
	ch := sc.lookahead(0)
	if isNameStart(ch) {
		sc.consume(ch)
		name := []rune{ch}
		ch = sc.lookahead(0)
		for isNameStart(ch) || isDigit(ch) {
			name = append(name, ch)
			sc.consume(ch)
			ch = sc.lookahead(0)
		}
		name_str := string(name)
		if name_str == "true" || name_str == "false" {
			return NewToken(BOOL_LIT, &name_str, lin)
		}
		if KeywordKind, isKeyword := Keywords[string(name)]; isKeyword {
			return NewToken(KeywordKind, &name_str, lin)
		}
		if BuiltinKind, isBuiltin := Builtins[string(name)]; isBuiltin {
			return NewToken(BuiltinKind, &name_str, lin)
		}
		return NewToken(IDENT, &name_str, lin)
	}

	switch ch {
	case rune(0):
		sc.consume(ch)
		return NewToken(EOF, nil, lin)
	case ';':
		sc.consume(ch)
		return NewToken(SEMI, nil, lin)
	case '\n':
		sc.consume(ch)
		return sc.Next()
		// return NewToken(Nline, []rune{}, lin)
	case '+':
		return sc.AssignOp(ADD, ADD_ASSIGN, lin)
	case '-':
		return sc.AssignOp(SUB, SUB_ASSIGN, lin)
	case '*':
		return sc.AssignOp(MUL, MUL_ASSIGN, lin)
	case '/':
		// line comment
		if sc.lookahead(1) == '/' {
			sc.consume('/')
			sc.consume('/')
			comment := []rune{}
			for sc.lookahead(0) != '\n' && sc.lookahead(0) != 0 {
				comment = append(comment, sc.lookahead(0))
				sc.consume(sc.lookahead(0))
			}
			return sc.Next() // skip comment
		}
		// multiline comment
		if sc.lookahead(1) == '*' {
			sc.consume('/')
			sc.consume('*')
			comment := []rune{}
			for sc.lookahead(0) != '*' && sc.lookahead(1) != '/' {
				if sc.lookahead(0) == 0 {
					errorMsg := fmt.Sprintf("ILLEGAL character('%v')", string(ch))
					return NewToken(ILLEGAL, &errorMsg, lin)
				}
				comment = append(comment, sc.lookahead(0))
				sc.consume(sc.lookahead(0))
			}
			sc.consume('*')
			sc.consume('/')
			return sc.Next() // skip comment
		}
		return sc.AssignOp(DIV, DIV_ASSIGN, lin)
	case '%':
		return sc.AssignOp(MOD, MOD_ASSIGN, lin)
	case '|':
		return sc.AssignOp(OR, OR_ASSIGN, lin)
	case '&':
		return sc.AssignOp(AND, AND_ASSIGN, lin)
	case '^':
		return sc.AssignOp(XOR, XOR_ASSIGN, lin)
	case '<':
		sc.consume(ch)
		if sc.lookahead(0) == '<' {
			return sc.AssignOp(LSH, LSH_ASSIGN, lin)
		}
		if sc.lookahead(0) == '=' {
			sc.consume('=')
			return NewToken(LEQ, nil, lin)
		}
		return NewToken(LSS, nil, lin)
	case '>':
		sc.consume(ch)
		if sc.lookahead(0) == '>' {
			return sc.AssignOp(RSH, RSH_ASSIGN, lin)
		}
		if sc.lookahead(0) == '=' {
			sc.consume('=')
			return NewToken(GEQ, nil, lin)
		}
		return NewToken(GTR, nil, lin)
	case '=':
		sc.consume(ch)
		if sc.lookahead(0) == '=' {
			sc.consume('=')
			return NewToken(EQL, nil, lin)
		}
		return NewToken(ASSIGN, nil, lin)
	case '!':
		sc.consume(ch)
		if sc.lookahead(0) == '=' {
			sc.consume('=')
			return NewToken(NEQ, nil, lin)
		}
		return NewToken(NOT, nil, lin)
	case '(':
		sc.consume(ch)
		return NewToken(LPAREN, nil, lin)
	case ')':
		sc.consume(ch)
		return NewToken(RPAREN, nil, lin)
	case '{':
		sc.consume(ch)
		return NewToken(LBRACE, nil, lin)
	case '}':
		sc.consume(ch)
		return NewToken(RBRACE, nil, lin)
	case '~':
		sc.consume(ch)
		return NewToken(TILDE, nil, lin)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return sc.lexNumber()
	case '"':
		sc.consume(ch)
		str := []rune{}
		ch = sc.lookahead(0)
		for ch != '"' {
			if ch == '\n' || ch == 0 {
				return NewToken(ILLEGAL, nil, lin)
			}
			str = append(str, sc.lookahead(0))
			sc.consume(ch)
			ch = sc.lookahead(0)
		}
		sc.consume(ch)
		str_str := string(str)
		return NewToken(STR_LIT, &str_str, lin)
	}
	sc.consume(ch) // keep going even if character is invalid
	errorMsg := fmt.Sprintf("invalid character '%v'", string(ch))
	return NewToken(ILLEGAL, &errorMsg, lin)
}

func (sc *Scanner) Peek(d int) Token {
	pos := sc.Index
	lin := sc.Line
	var tok Token
	for _ = range d + 1 {
		tok = sc.Next()
	}
	sc.Index = pos
	sc.Line = lin
	return tok
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isAlpha(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// new line is a Token
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\r'
}

func isNameStart(ch rune) bool {
	return ch == '_' || isAlpha(ch)
}
