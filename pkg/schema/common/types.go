package common

// VariableModifierFunction represents a Modifier. It accepts a context variable value, split by '/', along with
// a set of schema-provided arguments, and should modify the variable however it wants.
// It should return the modified variable slice.
type VariableModifierFunction func(variable []string, args []string) ([]string, error)

type ValidationContext struct {
	VariableStore     VariableStore
	VariableModifiers map[string]VariableModifierFunction
}

// The basic interface of our constraints.
type Constraint interface {
	Consume([]string, *ValidationContext) ([]string, error)
	String() string
	GetVariableName() string
}

type VariableStore interface {
	GetVariable(name string) (string, bool)
	GetVariableSet(name string) ([]string, bool)
}
