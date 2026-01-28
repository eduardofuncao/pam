package params

import (
	"fmt"
)

func ResolveParameters(paramDefs, cliValues map[string]string) map[string]string {
	result := make(map[string]string)

	// First, add all defaults
	for name, defaultValue := range paramDefs {
		result[name] = defaultValue
	}

	// Then override with CLI values (higher priority)
	for name, cliValue := range cliValues {
		if _, exists := result[name]; exists {
			result[name] = cliValue
		}
	}

	return result
}

func GetMissingRequired(paramDefs, currentValues map[string]string) []string {
	var missing []string

	for name, defaultValue := range paramDefs {
		// If default is empty, it's required
		if defaultValue == "" {
			if value, exists := currentValues[name]; !exists || value == "" {
				missing = append(missing, name)
			}
		}
	}

	return missing
}

func ValidateCLIValues(cliValues, paramDefs map[string]string) error {
	for name := range cliValues {
		if _, exists := paramDefs[name]; !exists {
			return fmt.Errorf("unknown parameter: %s", name)
		}
	}
	return nil
}

// Reserved flags that cannot be used as parameter names
var reservedFlags = map[string]bool{
	"edit":    true,
	"last":    true,
	"l":       true,
	"help":    true,
	"h":       true,
	"version": true,
	"v":       true,
}

func ValidateParamNames(paramDefs map[string]string) error {
	for name := range paramDefs {
		if reservedFlags[name] {
			return fmt.Errorf("parameter name '%s' conflicts with reserved flag", name)
		}
	}
	return nil
}
