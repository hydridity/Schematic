package constraints

import (
	"errors"

	"github.com/hydridity/Schematic/pkg/parser"
	"github.com/hydridity/Schematic/pkg/schema/context"
)

type WildcardSingleConstraint struct{}

func (c *WildcardSingleConstraint) Consume(path []string, context *context.ValidationContext) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}
	return path[1:], nil
}
func (c *WildcardSingleConstraint) String() string {
	return "WildcardSingleConstraint"
}

func (c *WildcardSingleConstraint) GetVariableName() string {
	return ""
}

func TryCreateWildcardSingleConstraint(part *parser.Part) *WildcardSingleConstraint {
	if part.Wildcard == nil {
		return nil
	}
	if *part.Wildcard != "+" {
		return nil
	}
	return &WildcardSingleConstraint{}
}
