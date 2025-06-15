package schema

import (
	"github.com/hydridity/Schematic/pkg/parser"
	"github.com/hydridity/Schematic/pkg/schema/constraints"
	"log"
)

// The basic interface of our constraints.
type Constraint interface {
	Consume([]string, *ValidationContext) ([]string, error)
	String() string
	GetVariableName() string
}

type constraintBuilder = func(part *parser.Part) Constraint

func getConstraintBuilders() []constraintBuilder {
	return []constraintBuilder{
		func(part *parser.Part) Constraint { return constraints.TryCreateLiteralConstraint(part) },
		func(part *parser.Part) Constraint { return constraints.TryCreateVariableConstraint(part) },
		func(part *parser.Part) Constraint { return constraints.TryCreateVariableSetConstraint(part) },
		func(part *parser.Part) Constraint { return constraints.TryCreateWildcardMultiConstraint(part) },
		func(part *parser.Part) Constraint { return constraints.TryCreateWildcardSingleConstraint(part) },
	}
}

func CompileConstraints(schemaAst *parser.SchemaAST) []Constraint {
	res := make([]Constraint, 0, len(schemaAst.Parts))
	builders := getConstraintBuilders()

	for _, part := range schemaAst.Parts {
		var constraint Constraint = nil
		for _, builder := range builders {
			constraint = builder(part)
			if constraint != nil {
				break
			}
		}
		if constraint == nil {
			log.Fatalf("Failed to compile constraints: cannot translate AST part %s", part.String())
		}
	}

	return res
}
