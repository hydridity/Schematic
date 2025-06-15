package parser

import (
	"fmt"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"strings"
)

// Define a simple AST for a schema like: $gitlab_path.strip_prefix("helm-")/$[technologies]/+
type SchemaAST struct {
	Parts []*Part `@@ ("/" @@)*`
}

type Part struct {
	Var      *Var    `  @@`
	VarSet   *VarSet `| @@`
	Wildcard *string `| @("+" | "*")`
	Literal  *string `| @Ident`
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
		builder.WriteString(*p.Wildcard)
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
