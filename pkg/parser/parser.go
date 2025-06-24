package parser

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Define a simple AST for a schema like: $gitlab_path.strip_prefix("helm-")/$[technologies]/+
type SchemaAST struct {
	Parts []*Part `@@ ("/" @@)*`
}

type Part struct {
	Var      *Var      `  @@`
	VarSet   *VarSet   `| @@`
	Wildcard *Wildcard `| @@`
	Literal  *string   `| @Ident`
}

type Wildcard struct {
	Symbol     string      `@("+" | "*")`
	Quantifier *Quantifier `@@?`
}

type Quantifier struct {
	Min int `"{" @Int`
	Max int `( "," @Int )? "}"`
}

type Var struct {
	Name     string    `"$" @Ident`
	Modifier *Modifier `( "." @@ )?`
}

type VarSet struct {
	Name string `"$""[" @Ident "]"`
}

type Modifier struct {
	Func string   `@Ident "("`
	Args []string `( Whitespace? @String ( Whitespace? "," Whitespace? @String )* Whitespace? ) ")"`
}

var schemaLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_-]*`},
	{Name: "String", Pattern: `"(?:\\.|[^"])*"`},
	{Name: "Slash", Pattern: `/`},
	{Name: "Dot", Pattern: `\.`},
	{Name: "Comma", Pattern: `\,`},
	{Name: "Plus", Pattern: `\+`},
	{Name: "Star", Pattern: `\*`},
	{Name: "Dollar", Pattern: `\$`},
	{Name: "LParen", Pattern: `\(`},
	{Name: "RParen", Pattern: `\)`},
	{Name: "LBracket", Pattern: `\[`},
	{Name: "RBracket", Pattern: `\]`},
	{Name: "LCBracker", Pattern: `\{`},
	{Name: "RCBracker", Pattern: `\}`},
	{Name: "Int", Pattern: `[0-9]+`},
	{Name: "Whitespace", Pattern: `[ \t\n\r]+`},
})

func NewParser() (*participle.Parser[SchemaAST], error) {
	return participle.Build[SchemaAST](
		participle.Lexer(schemaLexer),
		participle.Unquote("String"),
	)
}

func (p *Part) String() string {
	builder := strings.Builder{}

	switch {
	case p.Var != nil:
		builder.WriteString("Variable: ")
		builder.WriteString(p.Var.Name)
		if p.Var.Modifier != nil {
			args := strings.Join(p.Var.Modifier.Args, ", ")
			builder.WriteString(fmt.Sprintf("\n    Modifier: %s(%s), Arguments: %s", p.Var.Modifier.Func, args, args))
		}
	case p.VarSet != nil:
		builder.WriteString("VarSet:")
		builder.WriteString(p.VarSet.Name)
	case p.Wildcard != nil:
		builder.WriteString("Wildcard:")
		builder.WriteString(p.Wildcard.Symbol)
		if p.Wildcard.Quantifier != nil {
			if p.Wildcard.Quantifier.Min != 0 {
				builder.WriteString(fmt.Sprintf("\n Quantifier Min: %d\n", p.Wildcard.Quantifier.Min))
			} else {
				builder.WriteString("Quantifier Min is Missing!")
			}

			if p.Wildcard.Quantifier.Max != 0 {
				builder.WriteString(fmt.Sprintf("\n Quantifier Max: %d\n", p.Wildcard.Quantifier.Max))
			} else {
				builder.WriteString(" Quantifier Max is Missing!")
			}
		}
	case p.Literal != nil:
		builder.WriteString("Literal:")
		builder.WriteString(*p.Literal)
	}

	return builder.String()
}

func (s *SchemaAST) String() string {
	builder := strings.Builder{}

	for i, part := range s.Parts {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(part.String())
	}

	return builder.String()
}
