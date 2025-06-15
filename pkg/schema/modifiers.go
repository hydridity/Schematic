package schema

import (
	"fmt"
	"strings"

	"github.com/hydridity/Schematic/pkg/schema/common"
)

func modifierStripLastPrefix(variable []string, args []string) ([]string, error) {
	if len(args) <= 0 {
		return nil, fmt.Errorf("strip_last_prefix: expected at least 1 argument, found %d", len(args))
	}

	if len(variable) <= 0 {
		return variable, nil // stripping prefix from an empty variable results in an empty variable, no error
	}

	for _, prefix := range args {
		lastIndex := len(variable) - 1
		if strings.HasPrefix(variable[lastIndex], prefix) {
			variable[lastIndex] = variable[lastIndex][len(prefix):]
			break // Strip only one prefix
		}
	}

	return variable, nil
}

func getPredefinedModifiers() map[string]common.VariableModifierFunction {
	return map[string]common.VariableModifierFunction{
		"strip_last_prefix": modifierStripLastPrefix,
	}
}
