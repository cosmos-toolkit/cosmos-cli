package rules

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	modulePathRegex   = regexp.MustCompile(`^[a-z0-9]+([-._][a-z0-9]+)*/[a-z0-9]+([-._][a-z0-9]+)*(/[a-z0-9]+([-._][a-z0-9]+)*)*$`)
	projectNameRegex  = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	templateNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

func ValidateModulePath(module string) error {
	if module == "" {
		return fmt.Errorf("module path is required")
	}

	if !modulePathRegex.MatchString(module) {
		return fmt.Errorf("invalid module path format: %s", module)
	}

	return nil
}

func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name is required")
	}

	if !projectNameRegex.MatchString(name) {
		return fmt.Errorf("project name can only contain alphanumeric characters, hyphens, and underscores")
	}

	return nil
}

func ValidateType(typeName string) error {
	validTypes := []string{"api", "worker", "cli"}
	for _, t := range validTypes {
		if t == typeName {
			return nil
		}
	}
	return fmt.Errorf("invalid type: %s. Valid types: %s", typeName, strings.Join(validTypes, ", "))
}

func ValidateTemplateName(name string) error {
	if name == "" {
		return fmt.Errorf("template name is required")
	}

	if !templateNameRegex.MatchString(name) {
		return fmt.Errorf("template name can only contain alphanumeric characters, hyphens, and underscores")
	}

	return nil
}

func ValidateTypeCompatibility(templateTypes []string, requestedType string) error {
	if len(templateTypes) == 0 {
		return nil // No types specified means supports all
	}

	for _, t := range templateTypes {
		if t == requestedType {
			return nil
		}
	}

	return fmt.Errorf("template does not support type '%s'. Supported types: %s", requestedType, strings.Join(templateTypes, ", "))
}
