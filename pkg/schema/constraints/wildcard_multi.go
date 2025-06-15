package constraints

import (
	"github.com/hydridity/Schematic/pkg/parser"
	"github.com/hydridity/Schematic/pkg/schema/context"
)

type WildcardMultiConstraint struct{}

func (c *WildcardMultiConstraint) Consume(path []string, context *context.ValidationContext) ([]string, error) {
	return []string{}, nil
}

func (c *WildcardMultiConstraint) String() string {
	return "WildcardMultiConstraint"
}

func (c *WildcardMultiConstraint) GetVariableName() string {
	return ""
}

func CreateWildcardMultiConstraint() *WildcardMultiConstraint {
	return &WildcardMultiConstraint{}
}

func TryCreateWildcardMultiConstraint(part *parser.Part) (*WildcardMultiConstraint, bool) {
	if part.Wildcard == nil {
		return nil, false
	}
	if *part.Wildcard != "*" {
		return nil, false
	}
	return &WildcardMultiConstraint{}, true
}
