package glox

import "fmt"

type Value interface {
	isValue()
}

type BoolValue bool
type IntValue int64
type FloatValue float64
type ObjectValue Object

type Object interface{ isObject() }
type NilObject struct{}

type StringObject struct{ inner *string }
type FuntionObject struct {
	name  *string
	arity int
	chunk *Chunk
}

type StructObject map[*string]Value

func (v BoolValue) isValue()     {}
func (v IntValue) isValue()      {}
func (v FloatValue) isValue()    {}
func (v NilObject) isValue()     {}
func (v StringObject) isValue()  {}
func (v FuntionObject) isValue() {}
func (v StructObject) isValue()  {}

func (v NilObject) isObject()     {}
func (v StringObject) isObject()  {}
func (v FuntionObject) isObject() {}
func (v StructObject) isObject()  {}

func (v NilObject) String() string    { return "nil" }
func (v StringObject) String() string { return *v.inner }
func (v StructObject) String() string {
	res := "struct{\n"
	for key, val := range v {
		res += fmt.Sprint("\t", *key, "=", val, ",\n")
	}
	res += "}\n"
	return res
}

// func (v FuntionObject) String() string { return "fn " + k + "(" + string(v.arity) + ")" }
