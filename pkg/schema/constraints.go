package schema

import (
	"log"

	"github.com/hydridity/Schematic/pkg/parser"
	"github.com/hydridity/Schematic/pkg/schema/constraints"
	"github.com/hydridity/Schematic/pkg/schema/context"
)

type constraintBuilder = func(part *parser.Part) context.Constraint

func getConstraintBuilders() []constraintBuilder {
	return []constraintBuilder{
		func(part *parser.Part) context.Constraint { return constraints.TryCreateLiteralConstraint(part) },
		func(part *parser.Part) context.Constraint { return constraints.TryCreateVariableConstraint(part) },
		func(part *parser.Part) context.Constraint { return constraints.TryCreateVariableSetConstraint(part) },
		func(part *parser.Part) context.Constraint { return constraints.TryCreateWildcardMultiConstraint(part) },
		func(part *parser.Part) context.Constraint { return constraints.TryCreateWildcardSingleConstraint(part) },
	}
}

func CompileConstraints(schemaAst *parser.SchemaAST) []context.Constraint {
	res := make([]context.Constraint, 0, len(schemaAst.Parts))
	builders := getConstraintBuilders()

	for _, part := range schemaAst.Parts {
		var constraint context.Constraint = nil
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
