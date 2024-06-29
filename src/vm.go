package glox

import "fmt"

const UINT8_MAX = 255
const UINT16_MAX = 255 * 255

type VM struct {
	chunk   *Chunk
	ip      int              // INSTRUCTION POINTER - pointer to the the current instruction-nr were working on in the chunk
	stack   []Value          // Stack that holds all currently 'in memory' Values
	globals map[string]Value // HashMap (key: identifiers, value=global values)
	strings map[string]Value // to enable string-interning we store all active-string variables in this table
}

func NewVM(base_chunk *Chunk) *VM {
	return &VM{
		chunk:   base_chunk,
		stack:   make([]Value, 0, UINT8_MAX+1),
		globals: map[string]Value{},
		strings: map[string]Value{},
	}
}

func (vm *VM) push(v Value) {
	if len(vm.stack) >= UINT8_MAX*UINT8_MAX {
		panic("stack full")
	}
	vm.stack = append(vm.stack, v)
}

func (vm *VM) pop() Value {
	n := len(vm.stack)
	res := vm.stack[n-1] // we can decrement first then retreive the top/removed element
	vm.stack = vm.stack[0 : n-1]
	return res
}
func (vm *VM) peek(distance int) Value {
	n := len(vm.stack)
	return vm.stack[n-1-distance] // we can decrement first then retreive the top/removed element
}

func (vm *VM) readByte() byte {
	res := vm.chunk.bytecode[vm.ip]
	vm.ip += 1
	return res
}

func (vm *VM) readUint16() uint16 {
	res := (uint16(vm.chunk.bytecode[vm.ip]) << 8) | uint16(vm.chunk.bytecode[vm.ip+1])
	vm.ip += 2
	return res
}

func (vm *VM) readConst() Value {
	return vm.chunk.consts[vm.readByte()]
}

func (vm *VM) readString() StringObject {
	return vm.readConst().(StringObject)
}

func (vm *VM) binary(b, a Value, op OpCode) {
	ok := false
	switch a.(type) {
	case IntValue:
		_, ok = b.(IntValue)
	case FloatValue:
		_, ok = b.(FloatValue)
	case BoolValue:
		_, ok = b.(BoolValue)
	case StringObject:
		_, ok = b.(StringObject)
	}
	if !ok {
		panic("different type")
	}
	pnc := false
	switch op {
	case OP_ADD:
		switch a.(type) {
		case IntValue:
			vm.push(a.(IntValue) + b.(IntValue))
		case FloatValue:
			vm.push(a.(FloatValue) + b.(FloatValue))
		default:
			pnc = true
		}
	case OP_SUB:
		switch a.(type) {
		case IntValue:
			vm.push(a.(IntValue) - b.(IntValue))
		case FloatValue:
			vm.push(a.(FloatValue) - b.(FloatValue))
		default:
			pnc = true
		}
	case OP_MULT:
		switch a.(type) {
		case IntValue:
			vm.push(a.(IntValue) * b.(IntValue))
		case FloatValue:
			vm.push(a.(FloatValue) * b.(FloatValue))
		default:
			pnc = true
		}
	case OP_DIV:
		switch a.(type) {
		case IntValue:
			vm.push(a.(IntValue) / b.(IntValue))
		case FloatValue:
			vm.push(a.(FloatValue) / b.(FloatValue))
		default:
			pnc = true
		}
	case OP_MOD:
		switch a.(type) {
		case IntValue:
			vm.push(a.(IntValue) % b.(IntValue))
		default:
			pnc = true
		}
	case OP_LOR:
		switch a.(type) {
		case BoolValue:
			vm.push(a.(BoolValue) || b.(BoolValue))
		default:
			pnc = true
		}
	case OP_LAND:
		switch a.(type) {
		case BoolValue:
			vm.push(a.(BoolValue) && b.(BoolValue))
		default:
			pnc = true
		}
	case OP_EQL:
		switch a.(type) {
		case IntValue:
			vm.push(BoolValue(a.(IntValue) == b.(IntValue)))
		case FloatValue:
			vm.push(BoolValue(a.(FloatValue) == b.(FloatValue)))
		case BoolValue:
			vm.push(BoolValue(a.(BoolValue) == b.(BoolValue)))
		default:
			pnc = true
		}
	case OP_GTR:
		switch a.(type) {
		case IntValue:
			vm.push(BoolValue(a.(IntValue) > b.(IntValue)))
		case FloatValue:
			vm.push(BoolValue(a.(FloatValue) > b.(FloatValue)))
		default:
			pnc = true
		}
	case OP_LSS:
		switch a.(type) {
		case IntValue:
			vm.push(BoolValue(a.(IntValue) < b.(IntValue)))
		case FloatValue:
			vm.push(BoolValue(a.(FloatValue) < b.(FloatValue)))
		default:
			pnc = true
		}
	case OP_OR:
		switch a.(type) {
		case IntValue:
			vm.push(a.(IntValue) | b.(IntValue))
		default:
			pnc = true
		}
	case OP_XOR:
		switch a.(type) {
		case IntValue:
			vm.push(a.(IntValue) ^ b.(IntValue))
		default:
			pnc = true
		}
	case OP_LSH:
		switch a.(type) {
		case IntValue:
			vm.push(a.(IntValue) << b.(IntValue))
		default:
			pnc = true
		}
	case OP_RSH:
		switch a.(type) {
		case IntValue:
			vm.push(a.(IntValue) >> b.(IntValue))
		default:
			pnc = true
		}
	case OP_AND:
		switch a.(type) {
		case IntValue:
			vm.push(a.(IntValue) & b.(IntValue))
		default:
			pnc = true
		}
	default:
		pnc = true
	}
	if pnc {
		panic("unsupported type")
	}
}

func (vm *VM) run() {
	for {
		instruciton := OpCode(vm.readByte())
		switch instruciton {
		case OP_CONST:
			vm.push(vm.readConst())
		case OP_POP:
			vm.pop() // pop value from stack and forget it.

		case OP_DEF_GLOBAL:
			name := vm.readString().inner
			vm.globals[*name] = vm.peek(0)
			vm.pop()
		case OP_GET_GLOBAL:
			name := vm.readString().inner
			if val, ok := vm.globals[*name]; ok {
				vm.push(val)
			} else {
				panic("variable not found")
			}
		case OP_SET_GLOBAL:
			name := vm.readString().inner
			if _, ok := vm.globals[*name]; ok {
				vm.globals[*name] = vm.peek(0)
				vm.pop()
			} else {
				panic("variable not found")
			}
		case OP_GET_LOCAL:
			ind := vm.readByte()
			vm.push(vm.stack[ind])
		case OP_SET_LOCAL:
			ind := vm.readByte()
			vm.stack[ind] = vm.peek(0)
			// vm.pop()
		case OP_UNARY_ADD:
		case OP_UNARY_SUB:
			switch val := vm.pop(); val.(type) {
			case IntValue:
				vm.push(-val.(IntValue))
			case FloatValue:
				vm.push(-val.(FloatValue))
			default:
				panic("invalid unary sub operation")
			}
		case OP_UNARY_NOT:
			if val, ok := vm.pop().(BoolValue); ok {
				vm.push(!val)
			} else {
				panic("invalid unary not operation")
			}
		case OP_UNARY_TILDE:
			if val, ok := vm.pop().(IntValue); ok {
				vm.push(^val)
			} else {
				panic("invalid tilde operation")
			}
		case OP_GTR, OP_LSS, OP_EQL, OP_LOR, OP_LAND, OP_ADD, OP_SUB, OP_OR, OP_XOR, OP_MULT, OP_DIV, OP_MOD, OP_LSH, OP_RSH, OP_AND:
			vm.binary(vm.pop(), vm.pop(), instruciton)
		case OP_PRINT:
			fmt.Println(vm.pop())
		case OP_JUMP:
			offset := vm.readUint16()
			vm.ip += int(offset)
		case OP_JUMP_IF_FALSE:
			val, ok := vm.peek(0).(BoolValue)
			if !ok {
				panic("invalid bool type operation")
			}
            offset := vm.readUint16()
			if val == false {
				vm.ip += int(offset)
			}
		case OP_JUMP_BACK:
			offset := vm.readUint16()
			vm.ip -= int(offset)
		case OP_RETURN:
			return
		}
	}
}

// takes the source-code string (from file or repl) and interprets/runs it
func Interpret(input string) {
	pc := NewParserCompiler(input)
	pc.compile()
	vm := NewVM(pc.curChunk)
	vm.run()
}
