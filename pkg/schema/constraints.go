package schema

import (
	"log"

	"github.com/hydridity/Schematic/pkg/parser"
	"github.com/hydridity/Schematic/pkg/schema/constraints"
	"github.com/hydridity/Schematic/pkg/schema/context"
)

// The basic interface of our constraints.
type Constraint interface {
	Consume([]string, *ValidationContext) ([]string, error)
	String() string
	GetVariableName() string
}

func CompileConstraints(schemaAst *parser.SchemaAST) []Constraint {
	res := make([]Constraint, 0, len(schemaAst.Parts))
	for _, part := range schemaAst.Parts {
		constraint, err := constraints.CreateConstraint(part)
		if err != nil {
			log.Printf("Error creating constraint: %v", err)
		}
		res = append(res, constraint)
	}

	return res
}
