package constraints

import (
	"errors"
	"fmt"

	"github.com/hydridity/Schematic/pkg/parser"
	"github.com/hydridity/Schematic/pkg/schema/context"
)

type VariableSetConstraint struct {
	VariableName string
}

func (c *VariableSetConstraint) Consume(path []string, context *context.ValidationContext) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}
	variable, found := context.VariableStore.GetVariableSet(c.VariableName)
	if !found {
		return nil, fmt.Errorf("variable '%s' not found in store", c.VariableName)
	}
	if len(variable) == 0 {
		return nil, fmt.Errorf("variable set '%s' is empty", c.VariableName)
	}
	foundInSet := false
	for _, v := range variable {
		if path[0] == v {
			foundInSet = true
			break
		}
	}
	if !foundInSet {
		return nil, errors.New("invalid variable set constraint value")
	}
	return path[1:], nil
}

func (c *VariableSetConstraint) String() string {
	return fmt.Sprintf("VariableSetConstraint(%s)", c.VariableName)
}

func (c *VariableSetConstraint) GetVariableName() string {
	return c.VariableName
}

func TryCreateVariableSetConstraint(part *parser.Part) *VariableSetConstraint {
	if part.VarSet == nil {
		return nil
	}
	return &VariableSetConstraint{VariableName: part.VarSet.Name}
}
