package constraints

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hydridity/Schematic/pkg/parser"
	"github.com/hydridity/Schematic/pkg/schema/context"
)

type VariableModifier struct {
	FuncName string
	Args     []string
}
type VariableModifierInstance struct {
	Modifier VariableModifier
	Function context.VariableModifierFunction
}
type VariableConstraint struct {
	VariableName string
	Modifiers    []VariableModifier
}

func (c *VariableConstraint) Consume(path []string, context *context.ValidationContext) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}
	variable, found := context.VariableStore.GetVariable(c.VariableName)
	if !found {
		return nil, fmt.Errorf("variable '%s' not found in store", c.VariableName)
	}

	// Get modifier functions referenced in the constraint
	modifierInstances := make([]VariableModifierInstance, 0, len(c.Modifiers))
	for _, modifier := range c.Modifiers {
		fun, found := context.VariableModifiers[modifier.FuncName]
		if !found {
			return nil, fmt.Errorf("modifier '%s' not found in context modifiers", modifier.FuncName)
		}
		modifierInstances = append(modifierInstances, VariableModifierInstance{
			Modifier: modifier,
			Function: fun,
		})
	}

	// Apply the modifier functions to the variable in order
	var err error
	parts := strings.Split(variable, "/")
	for _, mod := range modifierInstances {
		parts, err = mod.Function(parts, mod.Modifier.Args)
		if err != nil {
			return nil, fmt.Errorf("modifier '%s' application failed: %v", mod.Modifier.FuncName, err)
		}
	}

	// Validate the input with modified variable parts

	if len(path) < len(parts) {
		return nil, fmt.Errorf("path too short for variable '%s'", variable)
	}
	for i, part := range parts {
		if path[i] != part {
			return nil, fmt.Errorf("invalid variable constraint value at part %d, variable '%s'", i, variable)
		}
	}
	return path[len(parts):], nil
}

func (c *VariableConstraint) String() string {
	return fmt.Sprintf("VariableConstraint(%s)", c.VariableName)
}

func (c *VariableConstraint) GetVariableName() string {
	return c.VariableName
}

func TryCreateVariableConstraint(part *parser.Part) *VariableConstraint {
	if part.Var != nil {
		return nil
	}
	modifiers := make([]VariableModifier, 0)
	if part.Var.Modifier != nil {
		modifiers = append(modifiers, VariableModifier{
			FuncName: part.Var.Modifier.Func,
			Args:     part.Var.Modifier.Args,
		})
	}
	return &VariableConstraint{VariableName: part.Var.Name, Modifiers: modifiers}
}
