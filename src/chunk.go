package glox

type OpCode byte

const (
	OP_CONST OpCode = iota

	OP_TRUE
	OP_FALSE

	OP_POP // pop top value of the stack and disregard it

	OP_GET_LOCAL // read current local value and push it on the stack
	OP_SET_LOCAL // writes to local-variable the top value on the stack

	OP_DEF_GLOBAL // define a global variable
	OP_GET_GLOBAL // read current global val and push it on stack
	OP_SET_GLOBAL // writes to existing global variable

	OP_LOR // ||

	OP_LAND // &&

	OP_EQL // ==
	OP_GTR // >
	OP_LSS // <

	OP_ADD // +
	OP_SUB // -
	OP_OR  // |
	OP_XOR // ^

	OP_MULT // *
	OP_DIV  // /
	OP_MOD  // %
	OP_LSH  // <<
	OP_RSH  // >>

	OP_AND // &

	// unary operators
	OP_UNARY_NOT
	OP_UNARY_ADD
	OP_UNARY_SUB
	OP_UNARY_TILDE

	OP_PRINT  // print expression
	OP_RETURN // return from the current function

	OP_JUMP
	OP_JUMP_IF_FALSE
	OP_JUMP_BACK
)

func (o OpCode) String() string {
	strs := []string{
		"OP_CONST ", "OP_TRUE ", "OP_FALSE ",
		"OP_POP ", "OP_GET_LOCAL", "OP_SET_LOCAL",
		"OP_GET_GLOBAL ", "OP_DEF_GLOBAL ", "OP_SET_GLOBAL ",
		"OP_LOR ", "OP_LAND ", "OP_EQL ", "OP_GTR ", "OP_LSS ",
		"OP_ADD ", "OP_SUB ", "OP_OR  ", "OP_XOR ", "OP_MULT ", "OP_DIV  ", "OP_MOD ", "OP_LSH ", "OP_RSH  ", "OP_AND ",
		"OP_UNARY_NOT ", "OP_UNARY_ADD ", "OP_UNARY_SUB ", "OP_UNARY_TILDE ",
		"OP_PRINT ", "OP_RETURN ", "OP_JUMP", "OP_JUMP_IF_FALSE", "OP_JUMP_BACK",
	}
	return strs[o]
}

type Chunk struct {
	bytecode []byte // instructions
	lines    []int
	consts   []Value
}

func NewChunk() *Chunk { return &Chunk{} }

func (c *Chunk) Write(b byte, line int) {
	c.bytecode = append(c.bytecode, b)
	c.lines = append(c.lines, line)
}

func (c *Chunk) AddConst(_const Value) int {
	ind := len(c.consts)
	if ind >= UINT8_MAX {
		panic("error: too many constants in one chunk")
	}
	c.consts = append(c.consts, _const)
	return ind
}
