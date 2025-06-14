package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Define a simple AST for a schema like: $gitlab_path.strip_prefix("helm-")/$[technologies]/+
type Schema struct {
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
	Func string `@Ident "("`
	Arg  string `@String ")"`
}

var schemaLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_-]*`},
	{Name: "String", Pattern: `"(?:\\.|[^"])*"`},
	{Name: "Slash", Pattern: `/`},
	{Name: "Dot", Pattern: `\.`},
	{Name: "Plus", Pattern: `\+`},
	{Name: "Star", Pattern: `\*`},
	{Name: "Dollar", Pattern: `\$`},
	{Name: "LParen", Pattern: `\(`},
	{Name: "RParen", Pattern: `\)`},
	{Name: "LBracket", Pattern: `\[`},
	{Name: "RBracket", Pattern: `\]`},
	{Name: "Whitespace", Pattern: `[ \t\n\r]+`},
})

func NewParser() (*participle.Parser[Schema], error) {
	return participle.Build[Schema](
		participle.Lexer(schemaLexer),
		participle.Unquote("String"),
	)
}
