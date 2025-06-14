package main

import (
	"fmt"

	"github.com/hydridity/Schematic/pkg/schema"

	"github.com/hydridity/Schematic/pkg/parser"

	"github.com/alecthomas/repr"
)

func main() {
	parser, err := parser.NewParser()
	if err != nil {
		panic(err)
	}

	//fmt.Println(parser.String())

	//schemaStr := `$gitlab_path.strip_prefix("helm-")/+/postgres/$[test]`
	schemaStr := `$gitlab_path.strip_prefix("helm-")/+/postgres/$[test]`
	schemaAst, err := parser.ParseString("", schemaStr)
	if err != nil {
		panic(err)
	}

	schema := schema.CompileSchema(schemaAst, nil)

	fmt.Println("Schema AST:")
	repr.Println(schemaAst)

	fmt.Println("Parsed AST schema:")
	for _, part := range schemaAst.Parts {
		switch {
		case part.Var != nil:
			fmt.Println("Var:", part.Var.Name)
			if part.Var.Modifier != nil {
				fmt.Printf("Modifier: %s(%s), Argument: %s\n", part.Var.Modifier.Func, part.Var.Modifier.Arg, part.Var.Modifier.Arg)
			}
		case part.VarSet != nil:
			fmt.Println("VarSet:", part.VarSet.Name)
		case part.Wildcard != nil:
			fmt.Println("Wildcard:", *part.Wildcard)
		case part.Literal != nil:
			fmt.Println("Literal:", *part.Literal)
		}
	}

	fmt.Printf("Compiled schema constraints: %#v\n", schema)

	for _, part := range schema {
		fmt.Println("Part:", part.Debug())
	}
}
