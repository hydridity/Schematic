package constraints

import (
	"fmt"
	"github.com/hydridity/Schematic/pkg/parser"
	"github.com/hydridity/Schematic/pkg/schema"
)

type constraintBuilder = func(part *parser.Part) schema.Constraint

var constraintBuilders = []constraintBuilder{
	func(part *parser.Part) schema.Constraint { return TryCreateLiteralConstraint(part) },
	func(part *parser.Part) schema.Constraint { return TryCreateVariableConstraint(part) },
	func(part *parser.Part) schema.Constraint { return TryCreateVariableSetConstraint(part) },
	func(part *parser.Part) schema.Constraint { return TryCreateWildcardMultiConstraint(part) },
	func(part *parser.Part) schema.Constraint { return TryCreateWildcardSingleConstraint(part) },
}

func CreateConstraint(part *parser.Part) (schema.Constraint, error) {
	var constraint schema.Constraint = nil
	for _, builder := range constraintBuilders {
		constraint = builder(part)
		if constraint != nil {
			break
		}
	}
	if constraint == nil {
		return nil, fmt.Errorf("failed to compile constraints: cannot translate AST part %s", part.String())
	}
	return constraint, nil
}
