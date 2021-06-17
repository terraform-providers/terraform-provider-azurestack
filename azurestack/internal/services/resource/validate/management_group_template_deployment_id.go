package validate

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/resource/parse"
)

func ManagementGroupTemplateDeploymentID(input interface{}, key string) (warnings []string, errors []error) {
	v, ok := input.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string", key))
		return
	}

	if _, err := parse.ManagementGroupTemplateDeploymentID(v); err != nil {
		errors = append(errors, err)
	}

	return
}
