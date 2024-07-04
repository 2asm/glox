package glox

import (
	"fmt"
)

const UINT8_MAX = 255
const UINT16_MAX = 255 * 255
const CALLFRAME_MAX = 255 * 255

type CallFrame struct {
	function  *FuntionObject
	ip        int
	start_ind int // first index in value stack that this function can use
}

type VM struct {
	frames []*CallFrame
	// chunk   *Chunk
	// ip      int
	stack   []Value
	globals map[string]Value
	isPanic bool
}

func NewVM() *VM {
	return &VM{
		frames:  make([]*CallFrame, 0),
		stack:   make([]Value, 0, UINT8_MAX+1),
		globals: map[string]Value{},
	}
}

func (vm *VM) call(function *FuntionObject, args_count int) {
	if len(vm.frames) >= CALLFRAME_MAX {
		vm.isPanic = true
		return
	}
	if function.arity != args_count {
		vm.isPanic = true
		return
	}
	frame := &CallFrame{
		function:  function,
		ip:        0,
		start_ind: len(vm.stack) - args_count,
	}
	vm.frames = append(vm.frames, frame)
}

func (vm *VM) push(v Value) {
	if len(vm.stack) >= UINT8_MAX*UINT8_MAX {
		vm.isPanic = true
		return
	}
	vm.stack = append(vm.stack, v)
}

func (vm *VM) pop() Value {
	n := len(vm.stack)
	res := vm.stack[n-1]
	vm.stack = vm.stack[0 : n-1]
	return res
}
func (vm *VM) peek(distance int) Value {
	n := len(vm.stack)
	return vm.stack[n-1-distance]
}

func (vm *VM) cur_frame() *CallFrame {
	return vm.frames[len(vm.frames)-1]
}

func (vm *VM) readByte() byte {
	frame := vm.cur_frame()
	res := frame.function.chunk.bytecode[frame.ip]
	frame.ip += 1
	return res
}

func (vm *VM) readUint16() uint16 {
	frame := vm.cur_frame()
	res := (uint16(frame.function.chunk.bytecode[frame.ip]) << 8) | uint16(frame.function.chunk.bytecode[frame.ip+1])
	frame.ip += 2
	return res
}

func (vm *VM) readConst() Value {
	frame := vm.cur_frame()
	return frame.function.chunk.consts[vm.readByte()]
}

func (vm *VM) readString() StringObject {
	return vm.readConst().(StringObject)
}

func (vm *VM) binary(b, a Value, op OpCode) error {
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
		return fmt.Errorf("different type")
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
		return fmt.Errorf("unsupported type")
	}
	return nil
}

func (vm *VM) run() error {
	for {
		lin := vm.cur_frame().function.chunk.lines[vm.cur_frame().ip]
		instruciton := OpCode(vm.readByte())
		switch instruciton {
		case OP_CONST:
			vm.push(vm.readConst())
		case OP_POP:
			vm.pop()

		case OP_DEF_GLOBAL:
			name := vm.readString().inner
			vm.globals[*name] = vm.peek(0)
			vm.pop()
		case OP_GET_GLOBAL:
			name := vm.readString().inner
			if val, ok := vm.globals[*name]; ok {
				vm.push(val)
			} else {
				return fmt.Errorf("variable not found:%v", lin)
			}
		case OP_SET_GLOBAL:
			name := vm.readString().inner
			if _, ok := vm.globals[*name]; ok {
				vm.globals[*name] = vm.peek(0)
				vm.pop()
			} else {
				return fmt.Errorf("variable not found:%v", lin)
			}
		case OP_GET_LOCAL:
			ind := vm.readByte()
			vm.push(vm.stack[vm.cur_frame().start_ind+int(ind)])
		case OP_SET_LOCAL:
			ind := vm.readByte()
			vm.stack[vm.cur_frame().start_ind+int(ind)] = vm.peek(0)
			// vm.pop()
		case OP_UNARY_ADD:
		case OP_UNARY_SUB:
			switch val := vm.pop(); val.(type) {
			case IntValue:
				vm.push(-val.(IntValue))
			case FloatValue:
				vm.push(-val.(FloatValue))
			default:
				return fmt.Errorf("invalid unary sub operation:%v", lin)
			}
		case OP_UNARY_NOT:
			if val, ok := vm.pop().(BoolValue); ok {
				vm.push(!val)
			} else {
				return fmt.Errorf("invalid unary not operation:%v", lin)
			}
		case OP_UNARY_TILDE:
			if val, ok := vm.pop().(IntValue); ok {
				vm.push(^val)
			} else {
				return fmt.Errorf("invalid tilde operation:%v", lin)
			}
		case OP_GTR, OP_LSS, OP_EQL, OP_LOR, OP_LAND, OP_ADD, OP_SUB, OP_OR, OP_XOR, OP_MULT, OP_DIV, OP_MOD, OP_LSH, OP_RSH, OP_AND:
			if e := vm.binary(vm.pop(), vm.pop(), instruciton); e != nil {
				return e
			}
		case OP_PRINT:
			cnt := vm.readByte()
			cnt2 := cnt
			for cnt > 0 {
				cnt -= 1
				r := vm.peek(int(cnt))
				fmt.Printf("%v ", r)
			}
			fmt.Println()
			for cnt2 > 0 {
				vm.pop()
				cnt2 -= 1
			}
		case OP_JUMP:
			offset := vm.readUint16()
			vm.cur_frame().ip += int(offset)
		case OP_JUMP_IF_FALSE:
			val, ok := vm.peek(0).(BoolValue)
			if !ok {
				return fmt.Errorf("invalid bool type operation:%v", lin)
			}
			offset := vm.readUint16()
			if val == false {
				vm.cur_frame().ip += int(offset)
			}
		case OP_JUMP_BACK:
			offset := vm.readUint16()
			vm.cur_frame().ip -= int(offset)
		case OP_NIL:
			vm.push(NilObject{})
		case OP_CALL:
			args_count := vm.readByte()
			f, ok := vm.peek(int(args_count)).(*FuntionObject)
			if !ok {
				return fmt.Errorf("expected function:%v", lin)
			}
			vm.call(f, int(args_count))
		case OP_RETURN:
			result := vm.pop()
			vm.stack = vm.stack[0:vm.cur_frame().start_ind]
			vm.frames = vm.frames[0 : len(vm.frames)-1]
			if len(vm.frames) == 0 {
				return nil
			}
			vm.pop()
			vm.push(result)
		}
	}
}

func Interpret(input string) error {
	p := NewParser(input)
	p.compile()
	if p.isPanic == true {
		return fmt.Errorf("error")
	}
	vm := NewVM()
	vm.call(p.function, 0)
	err := vm.run()
	if vm.isPanic && err == nil {
		err = fmt.Errorf("error")
	}
	return err
}
