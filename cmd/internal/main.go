package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hydridity/Schematic/pkg/schema"
	"gopkg.in/yaml.v3"

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

func (vs VariableStore) GetVariable(name string) (string, bool) {
	env, ok := vs.Environments[name]
	if !ok {
		return "", false
	}
	value, found := os.LookupEnv(env.From)
	if !found {
		return "", false
	}
	return value, true
}

func (vs VariableStore) GetVariableSet(name string) ([]string, bool) {
	set, ok := vs.VariableSets[name]
	if !ok {
		return nil, false
	}
	return set.Content, true
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
	err := hclsimple.DecodeFile("./example-config.hcl", nil, &config)
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
		var err error
		inputSegments, err = constraint.Consume(inputSegments, variableStore)
		if err != nil {
			return fmt.Errorf("failed to consume input '%s' with constraint '%s': %w", input, constraint.Debug(), err)
		}
	}
	if len(inputSegments) > 0 {
		return fmt.Errorf("input '%s' did not fully consume all segments, remaining: %v", input, inputSegments)
	}
	return nil
}

func ExtractFromYaml(path string) ([]string, error) {
	var matches []string
	re := regexp.MustCompile(`<path:[^>]+>`)
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !(strings.HasSuffix(p, ".yaml") || strings.HasSuffix(p, ".yml")) {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		var node yaml.Node
		if err := yaml.Unmarshal(data, &node); err != nil {
			return err
		}
		var walk func(n *yaml.Node)
		walk = func(n *yaml.Node) {
			if n.Kind == yaml.ScalarNode && n.Tag == "!!str" {
				found := re.FindAllString(n.Value, -1)
				matches = append(matches, found...)
			}
			for _, c := range n.Content {
				walk(c)
			}
		}
		walk(&node)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
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

	repr.Println(schemaAst)

	inputs, err := ExtractFromYaml(".")
	if err != nil {
		log.Fatalf("Failed to extract paths from YAML files: %s", err)
	}
	fmt.Println("Extracted paths from YAML files")

	for _, input := range inputs {
		fmt.Println("Input:", input)
	}

	VariableStore := BuildVariableStore(config)
	fmt.Printf("Variable Store: %#v\n", VariableStore)
	schema := schema.CompileSchema(schemaAst, nil)

	inputStr := "deployment/backend/postgres/admin" // TODO: Some inputs for raw API Vault paths will have "data" after mounth path
	//TODO Example: "deployment/data/group1/helm-project1/postgres/admin"
	fmt.Println("Input to validate:", inputStr)
	err = Validate(inputStr, schema, VariableStore)
	if err != nil {
		fmt.Printf("Validation failed: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Validation succeeded")
	}

}
