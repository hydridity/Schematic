package constraints

import (
	"fmt"
	"github.com/hydridity/Schematic/pkg/parser"
	"github.com/hydridity/Schematic/pkg/schema/context"
)

// The basic interface of our constraints.
type Constraint interface {
	Consume([]string, *context.ValidationContext) ([]string, error)
	String() string
	GetVariableName() string
}

func CompileConstraints(schemaAst *parser.SchemaAST) ([]Constraint, error) {
	res := make([]Constraint, 0, len(schemaAst.Parts))
	for _, part := range schemaAst.Parts {
		constraint, err := CreateConstraint(part)
		if err != nil {
			return nil, fmt.Errorf("error creating constraint: %v", err)
		}
		res = append(res, constraint)
	}

	return res, nil
}

type constraintBuilder = func(part *parser.Part) (Constraint, bool)

var constraintBuilders = []constraintBuilder{
	func(part *parser.Part) (Constraint, bool) { return TryCreateLiteralConstraint(part) },
	func(part *parser.Part) (Constraint, bool) { return TryCreateVariableConstraint(part) },
	func(part *parser.Part) (Constraint, bool) { return TryCreateVariableSetConstraint(part) },
	func(part *parser.Part) (Constraint, bool) { return TryCreateWildcardMultiConstraint(part) },
	func(part *parser.Part) (Constraint, bool) { return TryCreateWildcardSingleConstraint(part) },
}

func CreateConstraint(part *parser.Part) (Constraint, error) {
	var constraint Constraint = nil
	var success = false
	for _, builder := range constraintBuilders {
		constraint, success = builder(part)
		if success {
			break
		}
	}
	if !success {
		return nil, fmt.Errorf("failed to compile constraints: cannot translate AST part %s", part.String())
	}
	return constraint, nil
}
