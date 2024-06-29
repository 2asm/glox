package glox

type Value interface{ isValue() }

type BoolValue bool
type IntValue int64
type FloatValue float64
type ObjectValue Object

func test() {
}

type Object interface{ isObject() }
type NilObject struct{}

type StringObject struct{ inner *string }
type FuntionObject struct {
	name        *string
	param_names []*string
	param_types []*string
	result_type *string
}

func (v BoolValue) isValue()     {}
func (v IntValue) isValue()      {}
func (v FloatValue) isValue()    {}
func (v NilObject) isValue()     {}
func (v StringObject) isValue()  {}
func (v FuntionObject) isValue() {}

func (v NilObject) isObject()     {}
func (v StringObject) isObject()  {}
func (v FuntionObject) isObject() {}

func (v StringObject) String() string { return *v.inner }

func (v NilObject) type_() string    { return "NIL" }
func (v BoolValue) type_() string    { return "BOOL" }
func (v IntValue) type_() string     { return "INT" }
func (v FloatValue) type_() string   { return "FLOAT" }
func (v StringObject) type_() string { return "STRING" }
func (v FuntionObject) type_() string {
	t := "fn "
	t += *v.name + " "
	t += "("
	for _, y := range v.param_types {
		t += *y + ","
	}
	t += ") "
	t += *v.result_type
	return t
}
