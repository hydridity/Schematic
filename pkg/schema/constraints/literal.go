package constraints

import (
	"errors"
	"fmt"

	"github.com/hydridity/Schematic/pkg/parser"
	"github.com/hydridity/Schematic/pkg/schema/common"
)

type LiteralConstraint struct {
	Literal string
}

func (c *LiteralConstraint) Consume(path []string, context *common.ValidationContext) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}
	if path[0] != c.Literal {
		return nil, fmt.Errorf("expected '%s', got '%s'", c.Literal, path[0])
	}
	return path[1:], nil
}

func (c *LiteralConstraint) String() string {
	return fmt.Sprintf("LiteralConstraint(%s)", c.Literal)
}

func (c *LiteralConstraint) GetVariableName() string {
	return "" // Literal constraints do not have a variable name
}

func TryCreateLiteralConstraint(part *parser.Part) *LiteralConstraint {
	if part.Literal == nil {
		return nil
	}
	return &LiteralConstraint{Literal: *part.Literal}
}
