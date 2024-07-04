package glox

import (
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
breakStmt

--- declaration ---
structDecl
funcDecl
varDecl
statement


// TODO
// --- top level declaration ---
// structDecl
// funcDecl
// varDecl
*/

type Compiler struct {
	enclosing *Compiler // Each Compiler points back to the Comnpiler for the function that encloses it
	function  *FuntionObject
	top_level bool
	//function.chunk   *Chunk
	locals     []Local
	scopeDepth int
	loopDepth  int
	breaks     []int
}

type Parser struct {
	*Scanner
	*Compiler
}

type Local struct {
	name  Token
	depth int
}

func NewParser(input string) *Parser {
	return &Parser{Scanner: NewScanner(input), Compiler: NewCompiler(nil, true)}
}

func NewCompiler(enclosing *Compiler, top_level bool) *Compiler {
	return &Compiler{
		enclosing:  enclosing,
		function:   &FuntionObject{chunk: NewChunk()},
		top_level:  top_level,
		locals:     make([]Local, 0, 256),
		scopeDepth: 0,
	}
}

func (p *Parser) consume(kind TokenKind) Token {
	t := p.Next()
	if t.Kind != kind {
		p.isPanic = true
	}
	return t
}

func (p *Parser) emitByte(bs ...byte) {
	for _, b := range bs {
		p.function.chunk.Write(b, p.Line)
	}
}

func (p *Parser) emitConst(val Value) {
	ind := p.function.chunk.AddConst(val)
	p.emitByte(byte(OP_CONST), byte(ind))
}

func (p *Parser) emitJumpBack(start int) {
	p.emitByte(byte(OP_JUMP_BACK))
	offset := len(p.function.chunk.bytecode) - start + 2
	if offset > UINT16_MAX {
		p.isPanic = true
		return
	}
	p.emitByte(byte(offset>>8), byte(offset&255))
}

func (p *Parser) emitJump(instruction OpCode) int {
	p.emitByte(byte(instruction), 0, 0)
	return len(p.function.chunk.bytecode) - 2
}

func (p *Parser) patchJump(offset int) {
	jump := len(p.function.chunk.bytecode) - offset - 2
	if jump > UINT16_MAX {
		p.isPanic = true
		return
	}
	p.function.chunk.bytecode[offset] = byte(jump >> 8)
	p.function.chunk.bytecode[offset+1] = byte(jump & 255)
}

// add new variable
func (p *Parser) addVar(name_token Token) {
	if p.scopeDepth == 0 && p.top_level == true { // global
		ind := p.function.chunk.AddConst(StringObject{inner: name_token.Lit})
		p.emitByte(byte(OP_DEF_GLOBAL), byte(ind))
	} else { // local
		lcl := Local{name: name_token, depth: p.scopeDepth}
		if len(p.locals) >= UINT8_MAX {
			p.isPanic = true
			return
		}
		p.locals = append(p.locals, lcl)
	}
}

func (p *Parser) varDecl() {
	p.consume(LET)
	name_token := p.consume(IDENT)
	p.consume(ASSIGN)
	p.parseExpr(LOWEST_PREC + 1)
	p.consume(SEMI)

	p.addVar(name_token)
}

func (p *Parser) funcDecl() {
	p.consume(FUNC)
	name_token := p.consume(IDENT)
	new_compiler := NewCompiler(p.Compiler, false)
	p.Compiler = new_compiler
	p.scopeDepth++

	p.consume(LPAREN)
	for {
		t := p.Peek(0)
		if t.Kind == RPAREN || t.Kind == EOF {
			break
		}
		var_token := p.consume(IDENT)
		p.addVar(var_token)
		p.function.arity += 1
		t = p.Peek(0)
		if t.Kind == RPAREN || t.Kind == EOF {
			break
		}
		p.consume(COMMA)
	}
	p.consume(RPAREN)

	p.block()

	p.emitByte(byte(OP_NIL), byte(OP_RETURN))
	f := p.Compiler.function

	p.Compiler = p.Compiler.enclosing

	p.emitConst(f)
	p.addVar(name_token)
}

func (p *Parser) stmt() {
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
	case BREAK:
		p.breakStmt()
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
	}
}

func (p *Parser) printStmt() {
	p.consume(PRINT)

	args_count := 0
	for {
		t := p.Peek(0)
		if t.Kind == SEMI || t.Kind == EOF {
			break
		}
		p.parseExpr(LOWEST_PREC + 1)
		args_count += 1
		t = p.Peek(0)
		if t.Kind == SEMI || t.Kind == EOF {
			break
		}
		p.consume(COMMA)
	}
	p.consume(SEMI)
	p.emitByte(byte(OP_PRINT), byte(args_count))
}
func (p *Parser) emptyStmt() {
	p.consume(SEMI)
}

func (p *Parser) exprStmt() {
	p.parseExpr(LOWEST_PREC + 1)
	p.consume(SEMI)
	p.emitByte(byte(OP_POP))
}

func (p *Parser) localIndex(name Token) int {
	for i := len(p.locals) - 1; i >= 0; i -= 1 {
		if *name.Lit == *p.locals[i].name.Lit {
			return i
		}
	}
	return -1
}

func (p *Parser) opAssign(assign TokenKind) {
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

func (p *Parser) getVar(is_local bool, ind int) {
	if is_local {
		p.emitByte(byte(OP_GET_LOCAL), byte(ind))
	} else {
		p.emitByte(byte(OP_GET_GLOBAL), byte(ind))
	}
}

func (p *Parser) setVar(is_local bool, ind int) {
	if is_local {
		p.emitByte(byte(OP_SET_LOCAL), byte(ind))
	} else {
		p.emitByte(byte(OP_SET_GLOBAL), byte(ind))
	}
}

func (p *Parser) assignStmt(assign TokenKind) {
	t := p.Next()
	ind := p.localIndex(t)
	is_local := ind != -1
	if !is_local {
		ind = p.function.chunk.AddConst(StringObject{inner: t.Lit})
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

func (p *Parser) block() {
	p.consume(LBRACE)

	for {
		t := p.Peek(0)
		if t.Kind == RBRACE || t.Kind == EOF {
			break
		}
		p.decl()
	}
	p.consume(RBRACE)
}

func (p *Parser) blockStmt() {
	p.scopeDepth++

	p.block()

	p.scopeDepth--
	n := len(p.locals) - 1
	for n >= 0 && p.locals[n].depth > p.scopeDepth {
		p.emitByte(byte(OP_POP))
		n -= 1
	}
	p.locals = p.locals[0 : n+1]
}

func (p *Parser) ifStmt() {
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

func (p *Parser) breakStmt() {
	if p.loopDepth == 0 {
		p.isPanic = true
		return
	}
	p.consume(BREAK)
	p.consume(SEMI)
	exit := p.emitJump(OP_JUMP)
	p.breaks = append(p.breaks, exit)
}

func (p *Parser) whileStmt() {
	p.consume(WHILE)
	start := len(p.function.chunk.bytecode)
	p.parseExpr(LOWEST_PREC + 1)

	exit := p.emitJump(OP_JUMP_IF_FALSE)
	p.emitByte(byte(OP_POP))

	p.loopDepth += 1
	p.blockStmt()
	p.loopDepth -= 1

	p.emitJumpBack(start)
	p.patchJump(exit)
	p.emitByte(byte(OP_POP))
	for _, end := range p.breaks {
		p.patchJump(end)
	}
	p.breaks = p.breaks[:0]
}

func (p *Parser) returnStmt() {
	if p.top_level == true {
		p.isPanic = true
		return
	}
	p.consume(RETURN)
	if p.Peek(0).Kind == SEMI {
		p.emitByte(byte(OP_NIL), byte(OP_RETURN))
	} else {
		p.parseExpr(LOWEST_PREC + 1)
		p.emitByte(byte(OP_RETURN))
	}
	p.consume(SEMI)
}

func (p *Parser) decl() {
	t := p.Peek(0)
	switch t.Kind {
	case FUNC:
		p.funcDecl()
	case LET:
		p.varDecl()
	default:
		p.stmt()
	}
}

// pratt parser
func (p *Parser) parseExpr(mprec int) {
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
			p.isPanic = true
			return
		}
		p.emitConst(IntValue(ival))
	case FLOAT_LIT:
		fval, err := strconv.ParseFloat(*lt.Lit, 10)
		if err != nil {
			p.isPanic = true
			return
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
			ind = p.function.chunk.AddConst(StringObject{inner: lt.Lit})
		}
		p.getVar(is_local, ind)
	case NIL:
		p.emitConst(NilObject{})
	default: // unreachable
		p.isPanic = true
		return
	}
	for {
		op := p.Peek(0)
		cprec := op.Kind.Prec()
		if mprec > cprec {
			return
		}
		p.Next()
		switch op.Kind {
		case LPAREN:
			args_count := 0
			for {
				t := p.Peek(0)
				if t.Kind == RPAREN || t.Kind == EOF {
					break
				}
				p.parseExpr(LOWEST_PREC + 1)
				args_count += 1
				t = p.Peek(0)
				if t.Kind == RPAREN || t.Kind == EOF {
					break
				}
				p.consume(COMMA)
			}
			p.consume(RPAREN)
			p.emitByte(byte(OP_CALL), byte(args_count))
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
			p.isPanic = true
			return
		}
	}
	// parse primary expression
}

func (p *Parser) compile() {
	for {
		if p.Peek(0).Kind == EOF {
			break
		}
		if p.Peek(0).Kind == ILLEGAL {
			p.isPanic = true
			return
		}
		p.decl()
	}
	p.emitByte(byte(OP_NIL), byte(OP_RETURN))
}
