package validate_test

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/services/containers/validate"
)

func TestContainerRegistryScopeMapName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "four",
			ErrCount: 1,
		},
		{
			Value:    "5five",
			ErrCount: 0,
		},
		{
			Value:    "five-123",
			ErrCount: 0,
		},
		{
			Value:    "hello-world",
			ErrCount: 0,
		},
		{
			Value:    "hello_world",
			ErrCount: 1,
		},
		{
			Value:    "helloWorld",
			ErrCount: 0,
		},
		{
			Value:    "helloworld12",
			ErrCount: 0,
		},
		{
			Value:    "hello@world",
			ErrCount: 1,
		},

		{
			Value:    "qfvbdsbvipqdbwsbddbdcwqffewsqwcdw21ddwqwd3324120",
			ErrCount: 0,
		},
		{
			Value:    "qfvbdsbvipqdbwsbddbdcwqffewsqwcdw21ddwqwd33241202",
			ErrCount: 0,
		},
		{
			Value:    "qfvbdsbvipqdbwsbddbdcwqfjjfewsqwcdw21ddwqwd3324120fadfadf",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validate.ContainerRegistryScopeMapName(tc.Value, "azurerm_container_registry_scope_map")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Azure RM Container Registry Token Name to trigger a validation error: %v", errors)
		}
	}
}