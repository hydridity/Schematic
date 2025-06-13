package main

import (
	"fmt"
	"log"
	"schematic/parser"
	"schematic/schema"

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
	processConfig(config)
	return config
}

func processConfig(config Config) {
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

func debug() { //This is simulation of application that imports the parser and schema packages and handles context
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
