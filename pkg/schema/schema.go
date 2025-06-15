package schema

import (
	schemaInternal "github.com/hydridity/Schematic/internal/pkg/schema"
)

type VariableStore = schemaInternal.VariableStore

// VariableModifierFunction represents a Modifier. It accepts a context variable value, split by '/', along with
// a set of schema-provided arguments, and should modify the variable however it wants.
// It should return the modified variable slice.
type VariableModifierFunction = schemaInternal.VariableModifierFunction

type ValidationContext = schemaInternal.ValidationContext

type Schema = schemaInternal.Schema

var CreateSchema = schemaInternal.CreateSchema
