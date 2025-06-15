package main

import (
	"fmt"
	"github.com/hydridity/Schematic/pkg/schema"
	"log"
	"os"

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

func main() {
	config := loadConfig()

	variableStore := BuildVariableStore(config)
	fmt.Printf("Variable Store: %#v\n", variableStore)
	context := schema.ValidationContext{
		VariableStore:     &variableStore,
		VariableModifiers: nil,
	}
	schemaCompiled, err := schema.CreateSchema(config.Schema)
	if err != nil {
		log.Fatal(err)
	}

	inputStr := "deployment/group1/helm-project1/postgres/admin" // TODO: Some inputs for raw API Vault paths will have "data" after mounth path
	//TODO Example: "deployment/data/group1/helm-project1/postgres/admin"
	fmt.Println("Input to validate:", inputStr)
	err = schemaCompiled.Validate(inputStr, &context)
	if err != nil {
		fmt.Printf("Validation failed: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Validation succeeded")
	}
}
