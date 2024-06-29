package glox

import (
	"fmt"
	"strconv"
)

/*
--- statement ---
exprStmt
ifStmt
printStmt
returnStmt
whileStmt
blockStmt
emptyStmt
assignStmt

--- declaration ---
structDecl
funDecl
varDecl
statement


// TODO
// --- top level declaration ---
// structDecl
// funDecl
// varDecl
*/

type Local struct {
	name  Token
	depth int
}

type ParserCompiler struct {
	*Scanner
	curChunk   *Chunk // current compiling chunk
	locals     []Local
	scopeDepth int
}

func NewParserCompiler(input string) *ParserCompiler {
	return &ParserCompiler{
		Scanner:  NewScanner(input),
		curChunk: NewChunk(),
		locals:   make([]Local, 0, 256),
	}
}

func (p *ParserCompiler) consume(kind TokenKind) Token {
	t := p.Next()
	if t.Kind != kind {
		panic(fmt.Sprintf("error invalid token %v at line %v, expeted  %v", t, p.Line, kind))
	}
	return t
}

func (p *ParserCompiler) emitByte(bs ...byte) {
	for _, b := range bs {
		p.curChunk.Write(b, p.Line)
	}
}

func (p *ParserCompiler) emitConst(val Value) {
	ind := p.curChunk.AddConst(val)
	p.emitByte(byte(OP_CONST), byte(ind))
}

func (p *ParserCompiler) emitJumpBack(start int) {
	p.emitByte(byte(OP_JUMP_BACK))
	offset := len(p.curChunk.bytecode) - start + 2
	if offset > UINT16_MAX {
		panic("jump is too large")
	}
	p.emitByte(byte(offset>>8), byte(offset&255))
}

func (p *ParserCompiler) emitJump(instruction OpCode) int {
	p.emitByte(byte(instruction), 0, 0)
	return len(p.curChunk.bytecode) - 2
}

func (p *ParserCompiler) patchJump(offset int) {
	jump := len(p.curChunk.bytecode) - offset - 2
	if jump > UINT16_MAX {
		panic("jump is too large")
	}
	p.curChunk.bytecode[offset] = byte(jump >> 8)
	p.curChunk.bytecode[offset+1] = byte(jump & 255)
}

func (p *ParserCompiler) varDecl() {
	p.consume(LET)
	name_token := p.consume(IDENT)
	p.consume(ASSIGN)
	p.parseExpr(LOWEST_PREC + 1)
	p.consume(SEMI)
	if p.scopeDepth == 0 {
		ind := p.curChunk.AddConst(StringObject{inner: name_token.Lit})
		p.emitByte(byte(OP_DEF_GLOBAL), byte(ind))
	} else {
		lcl := Local{name: name_token, depth: p.scopeDepth}
		if len(p.locals) >= UINT8_MAX {
			panic("too many local variables")
		}
		p.locals = append(p.locals, lcl)
	}
}

func (p *ParserCompiler) stmt() {
	t := p.Peek(0)
	switch t.Kind {
	case PRINT:
		p.printStmt()
	case LBRACE:
		p.blockStmt()
	case IF:
		p.ifStmt()
	case WHILE:
		p.whileStmt()
	case RETURN:
		p.returnStmt()
	case SEMI:
		p.emptyStmt()
	case IDENT:
		assign := p.Peek(1).Kind
		switch assign {
		case ASSIGN, MUL_ASSIGN, DIV_ASSIGN, MOD_ASSIGN, ADD_ASSIGN, SUB_ASSIGN, LSH_ASSIGN, RSH_ASSIGN, AND_ASSIGN, XOR_ASSIGN, OR_ASSIGN:
			p.assignStmt(assign)
		default:
			p.exprStmt()
		}
	default:
		p.exprStmt()
		// panic("not valid statement")
	}
}

func (p *ParserCompiler) printStmt() {
	p.consume(PRINT)
	p.parseExpr(LOWEST_PREC + 1)
	p.consume(SEMI)
	p.emitByte(byte(OP_PRINT))
}
func (p *ParserCompiler) emptyStmt() {
	p.consume(SEMI)
}

func (p *ParserCompiler) exprStmt() {
	p.parseExpr(LOWEST_PREC + 1)
	p.consume(SEMI)
	p.emitByte(byte(OP_POP))
}

func (p *ParserCompiler) localIndex(name Token) int {
	for i := len(p.locals) - 1; i >= 0; i -= 1 {
		if *name.Lit == *p.locals[i].name.Lit {
			return i
		}
	}
	return -1
}

func (p *ParserCompiler) opAssign(assign TokenKind) {
	p.parseExpr(LOWEST_PREC + 1)
	switch assign {
	case ADD_ASSIGN:
		p.emitByte(byte(OP_ADD))
	case SUB_ASSIGN:
		p.emitByte(byte(OP_SUB))
	case MUL_ASSIGN:
		p.emitByte(byte(OP_MULT))
	case DIV_ASSIGN:
		p.emitByte(byte(OP_DIV))
	case MOD_ASSIGN:
		p.emitByte(byte(OP_MOD))
	case LSH_ASSIGN:
		p.emitByte(byte(OP_LSH))
	case RSH_ASSIGN:
		p.emitByte(byte(OP_RSH))
	case AND_ASSIGN:
		p.emitByte(byte(OP_AND))
	case OR_ASSIGN:
		p.emitByte(byte(OP_OR))
	case XOR_ASSIGN:
		p.emitByte(byte(OP_XOR))
	}
}

func (p *ParserCompiler) getVar(is_local bool, ind int) {
	if is_local {
		p.emitByte(byte(OP_GET_LOCAL), byte(ind))
	} else {
		p.emitByte(byte(OP_GET_GLOBAL), byte(ind))
	}
}

func (p *ParserCompiler) setVar(is_local bool, ind int) {
	if is_local {
		p.emitByte(byte(OP_SET_LOCAL), byte(ind))
	} else {
		p.emitByte(byte(OP_SET_GLOBAL), byte(ind))
	}
}

func (p *ParserCompiler) assignStmt(assign TokenKind) {
	t := p.Next()
	ind := p.localIndex(t)
	is_local := ind != -1
	if !is_local {
		ind = p.curChunk.AddConst(StringObject{inner: t.Lit})
	}
	p.consume(assign)
	switch assign {
	case ASSIGN:
		p.parseExpr(LOWEST_PREC + 1)
	default:
		p.getVar(is_local, ind)
		p.opAssign(assign)
	}
	p.consume(SEMI)
	p.setVar(is_local, ind)
}

func (p *ParserCompiler) blockStmt() {
	p.scopeDepth++

	p.consume(LBRACE)

	for {
		t := p.Peek(0)
		if t.Kind == RBRACE || t.Kind == EOF {
			break
		}
		p.decl()
	}
	p.consume(RBRACE)

	p.scopeDepth--
	n := len(p.locals) - 1
	for n >= 0 && p.locals[n].depth > p.scopeDepth {
		p.emitByte(byte(OP_POP))
		n -= 1
	}
	p.locals = p.locals[0 : n+1]
}

func (p *ParserCompiler) ifStmt() {
	p.consume(IF)
	p.parseExpr(LOWEST_PREC + 1)

	Jump_index_false := p.emitJump(OP_JUMP_IF_FALSE)
	p.emitByte(byte(OP_POP))
	p.blockStmt()
	jump_index_true := p.emitJump(OP_JUMP)
	p.patchJump(Jump_index_false)
	p.emitByte(byte(OP_POP))
	if p.Peek(0).Kind == ELSE {
		p.consume(ELSE)
		if p.Peek(0).Kind == IF {
			p.ifStmt()
		} else if p.Peek(0).Kind == LBRACE {
			p.blockStmt()
		}
	}
	p.patchJump(jump_index_true)
}
func (p *ParserCompiler) whileStmt() {
	p.consume(WHILE)
	start := len(p.curChunk.bytecode)
	p.parseExpr(LOWEST_PREC + 1)
	exit := p.emitJump(OP_JUMP_IF_FALSE)
	p.emitByte(byte(OP_POP))
	p.blockStmt()
	p.emitJumpBack(start)
	p.patchJump(exit)
	p.emitByte(byte(OP_POP))
}
func (p *ParserCompiler) returnStmt() { panic("not implemented") }

func (p *ParserCompiler) decl() {
	t := p.Peek(0)
	switch t.Kind {
	case LET:
		p.varDecl()
	default:
		p.stmt()
	}
}

// pratt parser
func (p *ParserCompiler) parseExpr(mprec int) {
	lt := p.Next()
	switch lt.Kind {
	case ADD:
		p.parseExpr(HIGHEST_PREC)
		p.emitByte(byte(OP_UNARY_ADD))
	case SUB:
		p.parseExpr(HIGHEST_PREC)
		p.emitByte(byte(OP_UNARY_SUB))
	case TILDE:
		p.parseExpr(HIGHEST_PREC)
		p.emitByte(byte(OP_UNARY_TILDE))
	case NOT:
		p.parseExpr(HIGHEST_PREC)
		p.emitByte(byte(OP_UNARY_NOT))
	case LPAREN:
		p.parseExpr(LOWEST_PREC + 1)
		p.consume(RPAREN)
	case INT_LIT:
		ival, err := strconv.ParseInt(*lt.Lit, 10, 64)
		if err != nil {
			panic(err.Error())
		}
		p.emitConst(IntValue(ival))
	case FLOAT_LIT:
		fval, err := strconv.ParseFloat(*lt.Lit, 10)
		if err != nil {
			panic(err.Error())
		}
		p.emitConst(FloatValue(fval))
	case STR_LIT:
		p.emitConst(StringObject{inner: lt.Lit})
	case BOOL_LIT:
		p.emitConst(BoolValue(*lt.Lit == "true"))
	case IDENT:
		ind := p.localIndex(lt)
		is_local := ind != -1
		if !is_local {
			ind = p.curChunk.AddConst(StringObject{inner: lt.Lit})
		}
		p.getVar(is_local, ind)
	default: // unreachable
		fmt.Println(lt)
		panic("unreachable")
	}
	for {
		op := p.Peek(0)
		cprec := op.Kind.Prec()
		if mprec > cprec {
			return
		}
		p.Next()
		switch op.Kind {
		case MUL:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_MULT))
		case DIV:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_DIV))
		case MOD:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_MOD))
		case ADD:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_ADD))
		case SUB:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_SUB))
		case LSH:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_LSH))
		case RSH:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_RSH))
		case LSS:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_LSS))
		case LEQ:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_GTR), byte(OP_UNARY_NOT))
		case GTR:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_GTR))
		case GEQ:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_LSS), byte(OP_UNARY_NOT))
		case EQL:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_EQL))
		case NEQ:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_EQL), byte(OP_UNARY_NOT))
		case AND:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_AND))
		case XOR:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_XOR))
		case OR:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_OR))
		case LAND:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_LAND))
		case LOR:
			p.parseExpr(cprec + 1)
			p.emitByte(byte(OP_LOR))
		default: //unreachable
			panic("unreachable")
		}
	}
	// parse primary expression
}

func (p *ParserCompiler) compile() {
	for {
		if p.Peek(0).Kind == EOF {
			break
		}
		if p.Peek(0).Kind == ILLEGAL {
			panic("error invalid character")
		}
		p.decl()
	}
	p.emitByte(byte(OP_RETURN))
}
