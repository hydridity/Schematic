package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/hydridity/Schematic/pkg/schema"

	"github.com/hydridity/Schematic/pkg/parser"

	"github.com/alecthomas/repr"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

type Environment struct {
	Name string `hcl:"name,label"`
	From string `hcl:"from"`
}

type Variable_Set struct {
	Content []string `hcl:"content"`
}

type Input struct {
	Name   string   `hcl:"name,label"`
	Type   string   `hcl:"type,label"`
	Remain hcl.Body `hcl:",remain"`
}

type Config struct {
	Schema string  `hcl:"schema"`
	Inputs []Input `hcl:"input,block"`
}

type VariableStore struct {
	Environments map[string]Environment
	VariableSets map[string]Variable_Set
}

func BuildVariableStore(config Config) VariableStore {
	envs := make(map[string]Environment)
	sets := make(map[string]Variable_Set)
	for _, input := range config.Inputs {
		switch input.Type {
		case "environment":
			var envInput Environment
			diags := gohcl.DecodeBody(input.Remain, nil, &envInput)
			if diags.HasErrors() {
				log.Fatalf("Failed to decode environment input: %s", diags.Error())
			}
			envs[input.Name] = envInput
		case "variable_set":
			var vsInput Variable_Set
			diags := gohcl.DecodeBody(input.Remain, nil, &vsInput)
			if diags.HasErrors() {
				log.Fatalf("Failed to decode variable_set input: %s", diags.Error())
			}
			sets[input.Name] = vsInput
		}
	}
	return VariableStore{
		Environments: envs,
		VariableSets: sets,
	}
}

func loadConfig() Config {
	var config Config
	err := hclsimple.DecodeFile("./cmd/internal/example-config.hcl", nil, &config)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}
	log.Printf("Configuration is %#v", config)
	debugProcessConfig(config)
	return config
}

func debugProcessConfig(config Config) {
	for _, input := range config.Inputs {
		switch input.Type {
		case "environment":
			var envInput Environment
			diags := gohcl.DecodeBody(input.Remain, nil, &envInput)
			if diags.HasErrors() {
				log.Fatalf("Failed to decode environment input: %s", diags.Error())
			}
			fmt.Printf("Env Input: name=%s, from=%s\n", input.Name, envInput.From)
		case "variable_set":
			var vsInput Variable_Set
			diags := gohcl.DecodeBody(input.Remain, nil, &vsInput)
			if diags.HasErrors() {
				log.Fatalf("Failed to decode variable_set input: %s", diags.Error())
			}
			fmt.Printf("Variable Set Input: name=%s, content=%v\n", input.Name, vsInput.Content)
		default:
			fmt.Printf("Unknown input type: %s\n", input.Type)
		}
	}
}

// This is simulation of application that imports the parser and schema packages and handles context
func debug() {
	config := loadConfig()

	parser, err := parser.NewParser()
	if err != nil {
		panic(err)
	}

	VariableStore := BuildVariableStore(config)
	fmt.Printf("Variable Store: %#v\n", VariableStore)

	//fmt.Println(parser.String())
	schemaStr := config.Schema
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
		fmt.Println("Requests:", part.GetVariableName())
	}
}

func Validate(input string, constraints []schema.Constraint, variableStore VariableStore) error {
	inputSegments := strings.Split(strings.Trim(input, "/"), "/")
	for _, constraint := range constraints {
		if remainder, err := constraint.Consume(inputSegments); err != nil {
			if len(remainder) > 0 {
				fmt.Printf("input '%s' did not fully consume all segments, remaining: %v\n", input, remainder)
			}
			return fmt.Errorf("validation failed for input '%s': %w", input, err)
		}
	}
	return nil
}

func main() {
	toDebug := false
	if toDebug {
		debug()
	}

	config := loadConfig()
	parser, err := parser.NewParser()
	if err != nil {
		panic(err)
	}

	schemaStr := config.Schema
	schemaAst, err := parser.ParseString("", schemaStr)
	if err != nil {
		panic(err)
	}

	VariableStore := BuildVariableStore(config)
	fmt.Printf("Variable Store: %#v\n", VariableStore)
	schema := schema.CompileSchema(schemaAst, nil)

	inputStr := "deployment/group1/helm-project1/postgres/admin" // TODO: Some inputs for raw API Vault paths will have "data" after mounth path
	//TODO Example: "deployment/data/group1/helm-project1/postgres/admin"
	fmt.Println("Input to validate:", inputStr)
	Validate(inputStr, schema, VariableStore)

}
