package validate

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/maintenance/parse"
)

func MaintenanceConfigurationID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := parse.MaintenanceConfigurationID(v); err != nil {
		errors = append(errors, err)
	}

	return
}
