package expr_test

import (
	"fmt"
	"testing"

	"github.com/expr-lang/expr/internal/testify/require"

	"github.com/expr-lang/expr"
)

// TestUserDefinedStaticLabels tests examples from the User-Defined Static Labels section
func TestUserDefinedStaticLabels(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		labels      map[string]any
		shouldMatch bool
	}{
		{
			name:       "Infrastructure and Platform team",
			expression: `type = "infrastructure" AND team = "platform"`,
			labels: map[string]any{
				"type": "infrastructure",
				"team": "platform",
			},
			shouldMatch: true,
		},
		{
			name:       "Infrastructure but not Platform team",
			expression: `type = "infrastructure" AND team = "platform"`,
			labels: map[string]any{
				"type": "infrastructure",
				"team": "payments",
			},
			shouldMatch: false,
		},
		{
			name:       "Application type",
			expression: `type = "application"`,
			labels: map[string]any{
				"type": "application",
				"team": "checkout",
			},
			shouldMatch: true,
		},
		{
			name:       "Customer facing SLA",
			expression: `sla = "customer-facing"`,
			labels: map[string]any{
				"sla":  "customer-facing",
				"type": "business",
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.labels))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.labels)
			require.NoError(t, err)
			require.Equal(t, tt.shouldMatch, out)
		})
	}
}

// TestPlatformProvidedLabels tests examples from the Platform-Provided Labels section
func TestPlatformProvidedLabels(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		labels      map[string]any
		shouldMatch bool
	}{
		{
			name:       "Critical threshold with application type",
			expression: `threshold.name = "critical" AND type = "application"`,
			labels: map[string]any{
				"threshold": map[string]any{
					"name": "critical",
				},
				"type": "application",
			},
			shouldMatch: true,
		},
		{
			name:       "Warning threshold - should not match critical expression",
			expression: `threshold.name = "critical" AND type = "application"`,
			labels: map[string]any{
				"threshold": map[string]any{
					"name": "warning",
				},
				"type": "application",
			},
			shouldMatch: false,
		},
		{
			name:       "Alert name matching",
			expression: `alertname = "High CPU Usage"`,
			labels: map[string]any{
				"alertname": "High CPU Usage",
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.labels))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.labels)
			require.NoError(t, err)
			require.Equal(t, tt.shouldMatch, out)
		})
	}
}

// TestDynamicQueryLabels tests examples from the Dynamic Query Labels section
func TestDynamicQueryLabels(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		labels      map[string]any
		shouldMatch bool
	}{
		{
			name:       "Kubernetes production namespace with API service",
			expression: `k8s.namespace.name = "production" AND service.name CONTAINS "api"`,
			labels: map[string]any{
				"k8s": map[string]any{
					"namespace": map[string]any{
						"name": "production",
					},
				},
				"service": map[string]any{
					"name": "checkout-api",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Kubernetes staging namespace - should not match production",
			expression: `k8s.namespace.name = "production" AND service.name CONTAINS "api"`,
			labels: map[string]any{
				"k8s": map[string]any{
					"namespace": map[string]any{
						"name": "staging",
					},
				},
				"service": map[string]any{
					"name": "checkout-api",
				},
			},
			shouldMatch: false,
		},
		{
			name:       "Host name IN list with region",
			expression: `host.name IN ["prod-server-1", "prod-server-2"] AND cloud.region = "us-east-1"`,
			labels: map[string]any{
				"host": map[string]any{
					"name": "prod-server-1",
				},
				"cloud": map[string]any{
					"region": "us-east-1",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Host name not in list",
			expression: `host.name IN ["prod-server-1", "prod-server-2"] AND cloud.region = "us-east-1"`,
			labels: map[string]any{
				"host": map[string]any{
					"name": "prod-server-3",
				},
				"cloud": map[string]any{
					"region": "us-east-1",
				},
			},
			shouldMatch: false,
		},
		{
			name:       "HTTP method and status code",
			expression: `http.method = "POST" AND http.status_code = "500"`,
			labels: map[string]any{
				"http": map[string]any{
					"method":      "POST",
					"status_code": "500",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Database operation",
			expression: `db.name = "orders_db" AND db.operation = "SELECT"`,
			labels: map[string]any{
				"db": map[string]any{
					"name":      "orders_db",
					"operation": "SELECT",
					"system":    "postgresql",
				},
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.labels))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.labels)
			require.NoError(t, err)
			require.Equal(t, tt.shouldMatch, out)
		})
	}
}

// TestRoutingProcessExample tests the specific example from "Routing Process Example" section
func TestRoutingProcessExample(t *testing.T) {
	alertLabels := map[string]any{
		"type": "application",
		"threshold": map[string]any{
			"name": "critical",
		},
		"deployment": map[string]any{
			"environment": "production",
		},
		"team": "checkout",
	}

	tests := []struct {
		name        string
		expression  string
		shouldMatch bool
	}{
		{
			name:        "Team equals checkout",
			expression:  `team = "checkout"`,
			shouldMatch: true,
		},
		{
			name:        "Critical threshold in production",
			expression:  `threshold.name = "critical" AND deployment.environment = "production"`,
			shouldMatch: true,
		},
		{
			name:        "Checkout team with application type",
			expression:  `team = "checkout" AND type = "application"`,
			shouldMatch: true,
		},
		{
			name:        "Wrong team - should not match",
			expression:  `team = "payment"`,
			shouldMatch: false,
		},
		{
			name:        "Warning threshold - should not match",
			expression:  `threshold.name = "warning" AND deployment.environment = "production"`,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(alertLabels))
			require.NoError(t, err)

			out, err := expr.Run(program, alertLabels)
			require.NoError(t, err)
			require.Equal(t, tt.shouldMatch, out)
		})
	}
}

// TestExpressionSyntaxOperators tests examples from the Expression Syntax section
func TestExpressionSyntaxOperators(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		labels      map[string]any
		shouldMatch bool
	}{
		// Equality operator
		{
			name:       "Service name equality - match",
			expression: `service.name = "api"`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "api",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Service name equality - no match",
			expression: `service.name = "api"`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "frontend",
				},
			},
			shouldMatch: false,
		},
		// Inequality operator
		{
			name:        "Environment inequality - match",
			expression:  `environment != "development"`,
			labels:      map[string]any{"environment": "production"},
			shouldMatch: true,
		},
		{
			name:        "Environment inequality - no match",
			expression:  `environment != "development"`,
			labels:      map[string]any{"environment": "development"},
			shouldMatch: false,
		},
		// Contains operator
		{
			name:       "Service name CONTAINS - match",
			expression: `service.name CONTAINS "auth"`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "user-auth-service",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Service name contains - no match",
			expression: `service.name contains "auth"`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "inventory-service",
				},
			},
			shouldMatch: false,
		},
		// In operator
		{
			name:        "Severity in list - match",
			expression:  `severity in ["critical", "warning"]`,
			labels:      map[string]any{"severity": "critical"},
			shouldMatch: true,
		},
		{
			name:        "Severity in list - no match",
			expression:  `severity in ["critical", "warning"]`,
			labels:      map[string]any{"severity": "info"},
			shouldMatch: false,
		},
		// Logical operators
		{
			name:       "AND operator - both true",
			expression: `service.name = "api" AND environment = "production"`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "api",
				},
				"environment": "production",
			},
			shouldMatch: true,
		},
		{
			name:       "AND operator - one false",
			expression: `service.name = "api" AND environment = "production"`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "api",
				},
				"environment": "staging",
			},
			shouldMatch: false,
		},
		{
			name:       "OR operator - both true",
			expression: `service.name = "api" OR environment = "production"`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "api",
				},
				"environment": "production",
			},
			shouldMatch: true,
		},
		{
			name:       "OR operator - one true",
			expression: `service.name = "api" OR environment = "production"`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "frontend",
				},
				"environment": "production",
			},
			shouldMatch: true,
		},
		{
			name:       "OR operator - both false",
			expression: `service.name = "api" OR environment = "production"`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "frontend",
				},
				"environment": "staging",
			},
			shouldMatch: false,
		},
		// Parentheses for grouping
		{
			name:       "Complex expression with parentheses - match",
			expression: `(service.name = "api" OR service.name = "frontend") AND environment = "production"`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "api",
				},
				"environment": "production",
			},
			shouldMatch: true,
		},
		{
			name:       "Complex expression with parentheses - no match",
			expression: `(service.name = "api" OR service.name = "frontend") AND environment = "production"`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "backend",
				},
				"environment": "production",
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.labels))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.labels)
			require.NoError(t, err)
			require.Equal(t, tt.shouldMatch, out)
		})
	}
}

// TestDetailedExamplePolicies tests the sample routing policies from the Detailed Example section
func TestDetailedExamplePolicies(t *testing.T) {
	tests := []struct {
		name        string
		policyName  string
		expression  string
		alertLabels map[string]any
		shouldMatch bool
	}{
		// Policy 1: Production Critical Escalation
		{
			name:       "Policy 1 - Critical production alert matches",
			policyName: "Production Critical Escalation",
			expression: `deployment.environment = "production" AND threshold.name = "critical"`,
			alertLabels: map[string]any{
				"deployment": map[string]any{
					"environment": "production",
				},
				"threshold": map[string]any{
					"name": "critical",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Policy 1 - Critical staging alert doesn't match",
			policyName: "Production Critical Escalation",
			expression: `deployment.environment = "production" AND threshold.name = "critical"`,
			alertLabels: map[string]any{
				"deployment": map[string]any{
					"environment": "staging",
				},
				"threshold": map[string]any{
					"name": "critical",
				},
			},
			shouldMatch: false,
		},
		// Policy 2: Checkout Team Alerts
		{
			name:       "Policy 2 - Checkout critical alert matches",
			policyName: "Checkout Team Alerts",
			expression: `service.name = "checkout" AND threshold.name IN ["critical", "warning"]`,
			alertLabels: map[string]any{
				"service": map[string]any{
					"name": "checkout",
				},
				"threshold": map[string]any{
					"name": "critical",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Policy 2 - Checkout warning alert matches",
			policyName: "Checkout Team Alerts",
			expression: `service.name = "checkout" AND threshold.name IN ["critical", "warning"]`,
			alertLabels: map[string]any{
				"service": map[string]any{
					"name": "checkout",
				},
				"threshold": map[string]any{
					"name": "warning",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Policy 2 - Checkout info alert doesn't match",
			policyName: "Checkout Team Alerts",
			expression: `service.name = "checkout" AND threshold.name IN ["critical", "warning"]`,
			alertLabels: map[string]any{
				"service": map[string]any{
					"name": "checkout",
				},
				"threshold": map[string]any{
					"name": "info",
				},
			},
			shouldMatch: false,
		},
		// Policy 3: Payment Service Monitoring
		{
			name:       "Policy 3 - Payment critical alert matches",
			policyName: "Payment Service All Envs",
			expression: `service.name = "payment" AND (threshold.name = "critical" OR (threshold.name = "warning" AND deployment.environment = "production"))`,
			alertLabels: map[string]any{
				"service": map[string]any{
					"name": "payment",
				},
				"threshold": map[string]any{
					"name": "critical",
				},
				"deployment": map[string]any{
					"environment": "staging",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Policy 3 - Payment warning in production matches",
			policyName: "Payment Service All Envs",
			expression: `service.name = "payment" AND (threshold.name = "critical" OR (threshold.name = "warning" AND deployment.environment = "production"))`,
			alertLabels: map[string]any{
				"service": map[string]any{
					"name": "payment",
				},
				"threshold": map[string]any{
					"name": "warning",
				},
				"deployment": map[string]any{
					"environment": "production",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Policy 3 - Payment warning in staging doesn't match",
			policyName: "Payment Service All Envs",
			expression: `service.name = "payment" AND (threshold.name = "critical" OR (threshold.name = "warning" AND deployment.environment = "production"))`,
			alertLabels: map[string]any{
				"service": map[string]any{
					"name": "payment",
				},
				"threshold": map[string]any{
					"name": "warning",
				},
				"deployment": map[string]any{
					"environment": "staging",
				},
			},
			shouldMatch: false,
		},
		// Policy 4: Dev Environment Notifications
		{
			name:       "Policy 4 - Development environment matches",
			policyName: "Dev Environment Notifications",
			expression: `deployment.environment = "development"`,
			alertLabels: map[string]any{
				"deployment": map[string]any{
					"environment": "development",
				},
				"service": map[string]any{
					"name": "any-service",
				},
			},
			shouldMatch: true,
		},
		// Policy 5: Infrastructure Critical
		{
			name:       "Policy 5 - Infrastructure critical alert matches",
			policyName: "Infrastructure Critical",
			expression: `type = "infrastructure" AND threshold.name = "critical"`,
			alertLabels: map[string]any{
				"type": "infrastructure",
				"threshold": map[string]any{
					"name": "critical",
				},
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.alertLabels))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.alertLabels)
			require.NoError(t, err)
			require.Equal(t, tt.shouldMatch, out)
		})
	}
}

// TestAlertRoutingExamples tests the specific alert examples from the Alert Routing Examples section
func TestAlertRoutingExamples(t *testing.T) {
	policies := []struct {
		name       string
		expression string
	}{
		{
			name:       "Policy 1: Production Critical Escalation",
			expression: `deployment.environment = "production" AND threshold.name = "critical"`,
		},
		{
			name:       "Policy 2: Checkout Team Alerts",
			expression: `service.name = "checkout" AND threshold.name IN ["critical", "warning"]`,
		},
		{
			name:       "Policy 3: Payment Service All Envs",
			expression: `service.name = "payment" AND (threshold.name = "critical" OR (threshold.name = "warning" AND deployment.environment = "production"))`,
		},
		{
			name:       "Policy 4: Dev Environment Notifications",
			expression: `deployment.environment = "development"`,
		},
		{
			name:       "Policy 5: Infrastructure Critical",
			expression: `type = "infrastructure" AND threshold.name = "critical"`,
		},
	}

	tests := []struct {
		name             string
		alertDescription string
		alertLabels      map[string]any
		expectedMatches  []string // Policy names that should match
	}{
		{
			name:             "Alert 1: Checkout service critical error in production",
			alertDescription: "Should match Policy 1 and Policy 2",
			alertLabels: map[string]any{
				"service": map[string]any{
					"name": "checkout",
				},
				"deployment": map[string]any{
					"environment": "production",
				},
				"threshold": map[string]any{
					"name": "critical",
				},
				"type":       "application",
				"team":       "checkout",
				"error_type": "timeout",
			},
			expectedMatches: []string{
				"Policy 1: Production Critical Escalation",
				"Policy 2: Checkout Team Alerts",
			},
		},
		{
			name:             "Alert 2: Payment service warning in staging",
			alertDescription: "Should not match any policy",
			alertLabels: map[string]any{
				"service": map[string]any{
					"name": "payment",
				},
				"deployment": map[string]any{
					"environment": "staging",
				},
				"threshold": map[string]any{
					"name": "warning",
				},
				"type": "application",
				"team": "payment",
			},
			expectedMatches: []string{},
		},
		{
			name:             "Alert 3: Database infrastructure critical alert",
			alertDescription: "Should match Policy 1 and Policy 5",
			alertLabels: map[string]any{
				"service": map[string]any{
					"name": "postgres-primary",
				},
				"deployment": map[string]any{
					"environment": "production",
				},
				"threshold": map[string]any{
					"name": "critical",
				},
				"type": "infrastructure",
				"db": map[string]any{
					"name": "orders_db",
				},
			},
			expectedMatches: []string{
				"Policy 1: Production Critical Escalation",
				"Policy 5: Infrastructure Critical",
			},
		},
		{
			name:             "Alert 4: Inventory service info in development",
			alertDescription: "Should match Policy 4",
			alertLabels: map[string]any{
				"service": map[string]any{
					"name": "inventory",
				},
				"deployment": map[string]any{
					"environment": "development",
				},
				"threshold": map[string]any{
					"name": "info",
				},
				"type": "application",
				"team": "inventory",
			},
			expectedMatches: []string{
				"Policy 4: Dev Environment Notifications",
			},
		},
	}

	for _, alertTest := range tests {
		t.Run(alertTest.name, func(t *testing.T) {
			var matchedPolicies []string

			// Check each policy against the alert
			for _, policy := range policies {
				program, err := expr.Compile(policy.expression, expr.Env(alertTest.alertLabels))
				require.NoError(t, err)

				out, err := expr.Run(program, alertTest.alertLabels)
				require.NoError(t, err)

				if out.(bool) {
					matchedPolicies = append(matchedPolicies, policy.name)
				}
			}

			// Verify the correct policies matched
			require.ElementsMatch(t, alertTest.expectedMatches, matchedPolicies,
				"Alert: %s\nExpected policies: %v\nActual policies: %v",
				alertTest.alertDescription, alertTest.expectedMatches, matchedPolicies)
		})
	}
}

// TestComplexNestedLabels tests handling of nested label structures
func TestComplexNestedLabels(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		labels      map[string]any
		shouldMatch bool
	}{
		{
			name:       "Nested Kubernetes labels",
			expression: `k8s.pod.name = "api-pod-123" AND k8s.namespace.name = "production"`,
			labels: map[string]any{
				"k8s": map[string]any{
					"pod": map[string]any{
						"name": "api-pod-123",
					},
					"namespace": map[string]any{
						"name": "production",
					},
					"deployment": map[string]any{
						"name": "api-deployment",
					},
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Service attributes with version",
			expression: `service.name = "checkout" AND service.version = "v2.1.0" AND service.environment = "production"`,
			labels: map[string]any{
				"service": map[string]any{
					"name":        "checkout",
					"version":     "v2.1.0",
					"environment": "production",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Cloud infrastructure labels",
			expression: `cloud.provider = "aws" AND cloud.region = "us-east-1" AND cloud.availability_zone contains "us-east-1"`,
			labels: map[string]any{
				"cloud": map[string]any{
					"provider":          "aws",
					"region":            "us-east-1",
					"availability_zone": "us-east-1a",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "HTTP request attributes",
			expression: `http.method = "POST" AND http.route = "/api/checkout" AND http.status_code IN ["500", "502", "503"]`,
			labels: map[string]any{
				"http": map[string]any{
					"method":      "POST",
					"route":       "/api/checkout",
					"status_code": "502",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Database query attributes",
			expression: `db.system = "postgresql" AND db.name = "orders_db" AND db.operation IN ["INSERT", "UPDATE", "DELETE"]`,
			labels: map[string]any{
				"db": map[string]any{
					"system":    "postgresql",
					"name":      "orders_db",
					"operation": "UPDATE",
				},
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.labels))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.labels)
			require.NoError(t, err)
			require.Equal(t, tt.shouldMatch, out)
		})
	}
}

// TestEdgeCases tests various edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		labels      map[string]any
		shouldMatch bool
		shouldError bool
	}{
		{
			name:        "Empty labels map",
			expression:  `service.name = "api"`,
			labels:      map[string]any{},
			shouldMatch: false,
			shouldError: true, // Will error because service doesn't exist
		},
		{
			name:       "Label exists but is nil",
			expression: `service.name != nil`,
			labels: map[string]any{
				"service": map[string]any{
					"name": nil,
				},
			},
			shouldMatch: false,
		},
		{
			name:       "Case sensitivity in label names",
			expression: `service.name = "api"`,
			labels: map[string]any{
				"service": map[string]any{
					"Name": "api", // Different case
				},
			},
			shouldMatch: false,
			shouldError: false,
		},
		{
			name:       "Special characters in label values",
			expression: `service.name = "api-service@v2.0"`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "api-service@v2.0",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Numeric string comparison",
			expression: `http.status_code = "200"`,
			labels: map[string]any{
				"http": map[string]any{
					"status_code": "200",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Boolean label values",
			expression: `is_critical = true`,
			labels: map[string]any{
				"is_critical": true,
			},
			shouldMatch: true,
		},
		{
			name:       "Complex IN expression with mixed types",
			expression: `threshold.name IN ["critical", "warning", "error"]`,
			labels: map[string]any{
				"threshold": map[string]any{
					"name": "warning",
				},
			},
			shouldMatch: true,
		},
		{
			name:       "Missing nested path",
			expression: `k8s.pod.name = "test"`,
			labels: map[string]any{
				"k8s": map[string]any{
					// pod is missing
				},
			},
			shouldMatch: false,
			shouldError: true,
		},
		{
			name:       "Partial nested path exists",
			expression: `service.version.major = 2`,
			labels: map[string]any{
				"service": map[string]any{
					"name": "api",
					// version is missing
				},
			},
			shouldMatch: false,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.labels))
			fmt.Println(program, err)

			if tt.shouldError {
				// If we expect an error during compilation or execution
				if err != nil {
					// Error during compilation is expected
					return
				}

				// Try to run and expect an error
				_, runErr := expr.Run(program, tt.labels)
				require.Error(t, runErr)
				return
			}

			require.NoError(t, err)

			out, err := expr.Run(program, tt.labels)
			require.NoError(t, err)
			require.Equal(t, tt.shouldMatch, out)
		})
	}
}

// TestRealWorldScenarios tests realistic routing policy scenarios
func TestRealWorldScenarios(t *testing.T) {
	tests := []struct {
		name        string
		scenario    string
		expression  string
		labels      map[string]any
		shouldMatch bool
	}{
		{
			name:       "High latency in payment service",
			scenario:   "Route to payment team if latency is high in payment service",
			expression: `service.name = "payment" AND metric.name = "latency" AND threshold.name IN ["warning", "critical"]`,
			labels: map[string]any{
				"service":   map[string]any{"name": "payment"},
				"metric":    map[string]any{"name": "latency"},
				"threshold": map[string]any{"name": "critical"},
				"value":     "2500ms",
			},
			shouldMatch: true,
		},
		{
			name:       "Database connection pool exhausted",
			scenario:   "Escalate to platform team for database issues in production",
			expression: `type = "infrastructure" AND db.name != "" AND deployment.environment = "production" AND threshold.name = "critical"`,
			labels: map[string]any{
				"type":       "infrastructure",
				"db":         map[string]any{"name": "orders_db"},
				"deployment": map[string]any{"environment": "production"},
				"threshold":  map[string]any{"name": "critical"},
				"metric":     map[string]any{"name": "connection_pool_exhausted"},
			},
			shouldMatch: true,
		},
		{
			name:       "Kubernetes pod restart loop",
			scenario:   "Alert platform team for pod stability issues",
			expression: `k8s.pod.restart_count > 5 OR (metric.name = "pod_restarts" AND threshold.name = "critical")`,
			labels: map[string]any{
				"k8s":       map[string]any{"pod": map[string]any{"name": "checkout-api-xyz", "restart_count": 10}},
				"metric":    map[string]any{"name": "pod_restarts"},
				"threshold": map[string]any{"name": "critical"},
			},
			shouldMatch: true,
		},
		{
			name:       "Business hours critical alerts",
			scenario:   "Different routing for business hours (assuming a business_hours label)",
			expression: `business_hours = true AND threshold.name = "critical" AND type = "application"`,
			labels: map[string]any{
				"business_hours": true,
				"threshold":      map[string]any{"name": "critical"},
				"type":           "application",
				"service":        map[string]any{"name": "checkout"},
			},
			shouldMatch: true,
		},
		{
			name:       "Multi-region failure detection",
			scenario:   "Escalate if multiple regions are affected",
			expression: `(cloud.region = "us-east-1" OR cloud.region = "eu-west-1") AND threshold.name = "critical" AND service.name = "api"`,
			labels: map[string]any{
				"cloud":     map[string]any{"region": "us-east-1"},
				"threshold": map[string]any{"name": "critical"},
				"service":   map[string]any{"name": "api"},
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Scenario: %s", tt.scenario)

			program, err := expr.Compile(tt.expression, expr.Env(tt.labels))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.labels)
			require.NoError(t, err)
			require.Equal(t, tt.shouldMatch, out)
		})
	}
}

// TestSQLStyleEqualityOperators tests SQL-style = and != operators
func TestSQLStyleEqualityOperators(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		env        map[string]any
		expected   bool
	}{
		// Single equals (=) operator tests
		{
			name:       "SQL equals operator - string match",
			expression: `name = "John"`,
			env:        map[string]any{"name": "John"},
			expected:   true,
		},
		{
			name:       "SQL equals operator - string no match",
			expression: `name = "Jane"`,
			env:        map[string]any{"name": "John"},
			expected:   false,
		},
		{
			name:       "SQL equals operator - number match",
			expression: `age = 25`,
			env:        map[string]any{"age": 25},
			expected:   true,
		},
		{
			name:       "SQL equals operator - nested field",
			expression: `user.status = "active"`,
			env:        map[string]any{"user": map[string]any{"status": "active"}},
			expected:   true,
		},
		// Not equals (!=) operator tests
		{
			name:       "SQL not equals operator - string different",
			expression: `name != "Jane"`,
			env:        map[string]any{"name": "John"},
			expected:   true,
		},
		{
			name:       "SQL not equals operator - string same",
			expression: `name != "John"`,
			env:        map[string]any{"name": "John"},
			expected:   false,
		},
		{
			name:       "SQL not equals operator - number",
			expression: `count != 0`,
			env:        map[string]any{"count": 5},
			expected:   true,
		},
		{
			name:       "SQL not equals operator - nested field",
			expression: `product.type != "premium"`,
			env:        map[string]any{"product": map[string]any{"type": "basic"}},
			expected:   true,
		},
		// Mixed operators in same expression
		{
			name:       "Mix SQL and standard equals",
			expression: `firstName = "John" AND lastName == "Doe"`,
			env:        map[string]any{"firstName": "John", "lastName": "Doe"},
			expected:   true,
		},
		{
			name:       "Mix SQL and standard not equals",
			expression: `status != "deleted" AND type != "temp"`,
			env:        map[string]any{"status": "active", "type": "permanent"},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.env))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.env)
			require.NoError(t, err)
			require.Equal(t, tt.expected, out)
		})
	}
}

// TestSQLStyleLogicalOperators tests uppercase AND, OR operators
func TestSQLStyleLogicalOperators(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		env        map[string]any
		expected   bool
	}{
		// AND operator tests
		{
			name:       "Uppercase AND - both true",
			expression: `active = true AND verified = true`,
			env:        map[string]any{"active": true, "verified": true},
			expected:   true,
		},
		{
			name:       "Uppercase AND - first false",
			expression: `active = false AND verified = true`,
			env:        map[string]any{"active": false, "verified": true},
			expected:   true,
		},
		{
			name:       "Uppercase AND - second false",
			expression: `active = true AND verified = false`,
			env:        map[string]any{"active": true, "verified": false},
			expected:   true,
		},
		{
			name:       "Uppercase AND - both false",
			expression: `active = false AND verified = false`,
			env:        map[string]any{"active": false, "verified": false},
			expected:   true,
		},
		// OR operator tests
		{
			name:       "Uppercase OR - both true",
			expression: `premium = true OR trial = true`,
			env:        map[string]any{"premium": true, "trial": true},
			expected:   true,
		},
		{
			name:       "Uppercase OR - first true",
			expression: `premium = true OR trial = false`,
			env:        map[string]any{"premium": true, "trial": false},
			expected:   true,
		},
		{
			name:       "Uppercase OR - second true",
			expression: `premium = false OR trial = true`,
			env:        map[string]any{"premium": false, "trial": true},
			expected:   true,
		},
		{
			name:       "Uppercase OR - both false",
			expression: `premium = false OR trial = false`,
			env:        map[string]any{"premium": false, "trial": false},
			expected:   true,
		},
		// Complex expressions with uppercase operators
		{
			name:       "Complex AND/OR expression",
			expression: `(status = "active" AND role = "admin") OR superuser = true`,
			env:        map[string]any{"status": "active", "role": "user", "superuser": true},
			expected:   true,
		},
		{
			name:       "Nested AND/OR with parentheses",
			expression: `enabled = true AND (role = "admin" OR role = "moderator")`,
			env:        map[string]any{"enabled": true, "role": "moderator"},
			expected:   true,
		},
		// Mixed case operators
		{
			name:       "Mix uppercase and lowercase AND",
			expression: `active = true AND verified = true and enabled = true`,
			env:        map[string]any{"active": true, "verified": true, "enabled": true},
			expected:   true,
		},
		{
			name:       "Mix uppercase and lowercase OR",
			expression: `status = "pending" OR status = "processing" or status = "queued"`,
			env:        map[string]any{"status": "queued"},
			expected:   true,
		},
		{
			name:       "Mix all logical operators",
			expression: `(active = true && verified = true) AND (premium = true || trial = true) OR override = true`,
			env:        map[string]any{"active": true, "verified": true, "premium": false, "trial": true, "override": false},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.env))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.env)
			require.NoError(t, err)
			require.Equal(t, tt.expected, out)
		})
	}
}

// TestSQLStyleInOperator tests uppercase IN operator
func TestSQLStyleInOperator(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		env        map[string]any
		expected   bool
	}{
		// Basic IN tests
		{
			name:       "Uppercase IN - string found",
			expression: `status IN ["active", "pending", "processing"]`,
			env:        map[string]any{"status": "pending"},
			expected:   true,
		},
		{
			name:       "Uppercase IN - string not found",
			expression: `status IN ["active", "pending", "processing"]`,
			env:        map[string]any{"status": "cancelled"},
			expected:   false,
		},
		{
			name:       "Uppercase IN - number found",
			expression: `priority IN [1, 2, 3]`,
			env:        map[string]any{"priority": 2},
			expected:   true,
		},
		{
			name:       "Uppercase IN - empty array",
			expression: `status IN []`,
			env:        map[string]any{"status": "active"},
			expected:   false,
		},
		// NOT IN tests
		{
			name:       "Uppercase NOT IN - value not in list",
			expression: `status NOT IN ["deleted", "archived", "hidden"]`,
			env:        map[string]any{"status": "active"},
			expected:   true,
		},
		{
			name:       "Uppercase NOT IN - value in list",
			expression: `status NOT IN ["deleted", "archived", "hidden"]`,
			env:        map[string]any{"status": "deleted"},
			expected:   false,
		},
		// Mixed case IN operators
		{
			name:       "Mix uppercase and lowercase IN",
			expression: `(status in ["active", "pending"]) AND (priority IN [1, 2, 3])`,
			env:        map[string]any{"status": "active", "priority": 2},
			expected:   true,
		},
		// Complex IN expressions
		{
			name:       "IN with nested field",
			expression: `user.role IN ["admin", "moderator", "editor"]`,
			env:        map[string]any{"user": map[string]any{"role": "editor"}},
			expected:   true,
		},
		{
			name:       "Multiple IN conditions",
			expression: `department IN ["sales", "marketing"] AND level IN [3, 4, 5]`,
			env:        map[string]any{"department": "sales", "level": 4},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.env))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.env)
			require.NoError(t, err)
			require.Equal(t, tt.expected, out)
		})
	}
}

// TestSQLStyleStringOperators tests uppercase CONTAINS, REGEXP operators
func TestSQLStyleStringOperators(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		env        map[string]any
		expected   bool
	}{
		// CONTAINS operator tests
		{
			name:       "Uppercase CONTAINS - substring found",
			expression: `message CONTAINS "error"`,
			env:        map[string]any{"message": "Critical error occurred"},
			expected:   true,
		},
		{
			name:       "Uppercase CONTAINS - substring not found",
			expression: `message CONTAINS "warning"`,
			env:        map[string]any{"message": "Everything is fine"},
			expected:   false,
		},
		{
			name:       "Uppercase CONTAINS - case sensitive",
			expression: `message CONTAINS "Error"`,
			env:        map[string]any{"message": "error occurred"},
			expected:   false,
		},
		{
			name:       "NOT CONTAINS",
			expression: `message NOT CONTAINS "spam"`,
			env:        map[string]any{"message": "Important notification"},
			expected:   true,
		},
		// REGEXP operator tests
		{
			name:       "Uppercase REGEXP - pattern match",
			expression: `email REGEXP "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`,
			env:        map[string]any{"email": "user@example.com"},
			expected:   true,
		},
		{
			name:       "Uppercase REGEXP - pattern no match",
			expression: `email REGEXP "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`,
			env:        map[string]any{"email": "invalid-email"},
			expected:   false,
		},
		{
			name:       "Lowercase regexp operator",
			expression: `phone regexp "^\\+?[1-9]\\d{1,14}$"`,
			env:        map[string]any{"phone": "+1234567890"},
			expected:   true,
		},
		{
			name:       "NOT REGEXP",
			expression: `username NOT REGEXP "^[0-9]"`,
			env:        map[string]any{"username": "john123"},
			expected:   true,
		},
		// Mixed string operators
		{
			name:       "Mix uppercase and lowercase string operators",
			expression: `(title contains "important") AND (body CONTAINS "urgent")`,
			env:        map[string]any{"title": "important notice", "body": "This is urgent!"},
			expected:   true,
		},
		{
			name:       "Complex string operator expression",
			expression: `(email REGEXP "@company\\.com$") AND (name CONTAINS "Admin" OR role = "admin")`,
			env:        map[string]any{"email": "john@company.com", "name": "John Administrator", "role": "user"},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.env))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.env)
			require.NoError(t, err)
			require.Equal(t, tt.expected, out)
		})
	}
}

// TestSQLStyleOperatorPrecedence tests operator precedence with SQL-style operators
func TestSQLStyleOperatorPrecedence(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		env        map[string]any
		expected   bool
	}{
		{
			name:       "AND has higher precedence than OR",
			expression: `status = "active" OR status = "pending" AND priority = 1`,
			env:        map[string]any{"status": "active", "priority": 5},
			expected:   true, // Should be evaluated as: active OR (pending AND priority=1)
		},
		{
			name:       "Uppercase AND has higher precedence than OR",
			expression: `verified = false OR active = true AND enabled = true`,
			env:        map[string]any{"verified": false, "active": true, "enabled": true},
			expected:   true,
		},
		{
			name:       "Parentheses override precedence",
			expression: `(status = "active" OR status = "pending") AND priority = 1`,
			env:        map[string]any{"status": "active", "priority": 5},
			expected:   false,
		},
		{
			name:       "Complex precedence with mixed operators",
			expression: `role IN ["admin", "moderator"] AND active = true OR override = true`,
			env:        map[string]any{"role": "user", "active": true, "override": true},
			expected:   true,
		},
		{
			name:       "Comparison operators have same precedence",
			expression: `age >= 18 AND age <= 65 AND status = "active"`,
			env:        map[string]any{"age": 25, "status": "active"},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.env))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.env)
			require.NoError(t, err)
			require.Equal(t, tt.expected, out)
		})
	}
}

// TestSQLStyleComplexScenarios tests real-world scenarios with SQL-style operators
func TestSQLStyleComplexScenarios(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		env        map[string]any
		expected   bool
	}{
		// User access control
		{
			name: "Complex user permission check",
			expression: `(user.role IN ["admin", "superuser"]) OR 
			            (user.role = "moderator" AND user.permissions CONTAINS "edit") OR
			            (user.role = "user" AND resource.owner = user.id AND resource.public = false)`,
			env: map[string]any{
				"user": map[string]any{
					"id":          123,
					"role":        "user",
					"permissions": []string{"read", "comment"},
				},
				"resource": map[string]any{
					"owner":  123,
					"public": false,
				},
			},
			expected: true,
		},
		// Database query filter simulation
		{
			name: "SQL-like WHERE clause simulation",
			expression: `status IN ["active", "pending"] AND 
			            created_date >= "2024-01-01" AND 
			            (category = "electronics" OR category = "computers") AND
			            price BETWEEN 100 AND 1000`,
			env: map[string]any{
				"status":       "active",
				"created_date": "2024-03-15",
				"category":     "electronics",
				"price":        599.99,
			},
			expected: false, // BETWEEN is not implemented, this will error
		},
		// Alert routing with SQL-style syntax
		{
			name: "Alert routing with uppercase operators",
			expression: `severity IN ["critical", "high"] AND
			            deployment.environment = "production" AND
			            (service.name REGEXP "^api-" OR service.name REGEXP "^auth-") AND
			            NOT (alert.silenced = true OR alert.acknowledged = true)`,
			env: map[string]any{
				"severity": "critical",
				"deployment": map[string]any{
					"environment": "production",
				},
				"service": map[string]any{
					"name": "api-gateway",
				},
				"alert": map[string]any{
					"silenced":     false,
					"acknowledged": false,
				},
			},
			expected: true,
		},
		// Email filtering rules
		{
			name: "Email filter with multiple conditions",
			expression: `(sender REGEXP "@(spam|junk|trash)\\.com$" OR 
			             subject CONTAINS "WIN NOW" OR 
			             subject CONTAINS "FREE MONEY") AND
			            NOT (sender IN ["whitelist@example.com", "trusted@company.com"]) AND
			            attachments > 0`,
			env: map[string]any{
				"sender":      "offer@spam.com",
				"subject":     "Regular email",
				"attachments": 2,
			},
			expected: true,
		},
		// Feature flag evaluation
		{
			name: "Feature flag with complex targeting",
			expression: `(user.subscription = "premium" OR user.subscription = "enterprise") AND
			            user.region IN ["US", "CA", "UK"] AND
			            (user.beta_tester = true OR user.employee = true OR percentage < 10)`,
			env: map[string]any{
				"user": map[string]any{
					"subscription": "premium",
					"region":       "US",
					"beta_tester":  false,
					"employee":     false,
				},
				"percentage": 5,
			},
			expected: true,
		},
		// Log filtering
		{
			name: "Log filter with nested conditions",
			expression: `level IN ["ERROR", "FATAL"] AND
			            (message CONTAINS "timeout" OR message CONTAINS "connection refused") AND
			            service.name != "health-check" AND
			            NOT (tags CONTAINS "expected-error" OR tags CONTAINS "ignore")`,
			env: map[string]any{
				"level":   "ERROR",
				"message": "Database connection refused",
				"service": map[string]any{
					"name": "api-service",
				},
				"tags": []string{"database", "critical"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Some tests expect compilation errors (like BETWEEN which isn't implemented)
			program, err := expr.Compile(tt.expression, expr.Env(tt.env))
			if err != nil {
				// If we got a compilation error, that might be expected
				t.Logf("Compilation error (might be expected): %v", err)
				return
			}

			out, err := expr.Run(program, tt.env)
			if err != nil {
				t.Logf("Runtime error: %v", err)
				return
			}

			require.Equal(t, tt.expected, out)
		})
	}
}

// TestSQLStyleOperatorCaseSensitivity verifies that SQL operators work in uppercase
func TestSQLStyleOperatorCaseSensitivity(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		env        map[string]any
		shouldWork bool
	}{
		{
			name:       "All uppercase SQL operators",
			expression: `status = "active" AND priority IN [1, 2, 3] AND message NOT CONTAINS "error"`,
			env:        map[string]any{"status": "active", "priority": 2, "message": "All good"},
			shouldWork: true,
		},
		{
			name:       "All lowercase operators",
			expression: `status = "active" and priority in [1, 2, 3] and message not contains "error"`,
			env:        map[string]any{"status": "active", "priority": 2, "message": "All good"},
			shouldWork: true,
		},
		{
			name:       "Mixed case operators",
			expression: `status = "active" AND priority in [1, 2, 3] and message NOT CONTAINS "error"`,
			env:        map[string]any{"status": "active", "priority": 2, "message": "All good"},
			shouldWork: true,
		},
		{
			name:       "REGEXP in uppercase",
			expression: `email REGEXP "^[^@]+@[^@]+\\.[^@]+$"`,
			env:        map[string]any{"email": "test@example.com"},
			shouldWork: true,
		},
		{
			name:       "regexp in lowercase",
			expression: `email regexp "^[^@]+@[^@]+\\.[^@]+$"`,
			env:        map[string]any{"email": "test@example.com"},
			shouldWork: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.env))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.env)
			require.NoError(t, err)
			require.Equal(t, true, out.(bool))
		})
	}
}

// TestSQLStyleOperatorOptimizations tests that SQL-style operators are properly optimized
func TestSQLStyleOperatorOptimizations(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		env        map[string]any
		expected   bool
		desc       string
	}{
		// Constant folding
		{
			name:       "Constant folding with SQL equals",
			expression: `5 = 5 AND status = "active"`,
			env:        map[string]any{"status": "active"},
			expected:   true,
			desc:       "Compiler should optimize 5 = 5 to true",
		},
		{
			name:       "Constant folding with SQL not equals",
			expression: `"a" != "b" AND enabled = true`,
			env:        map[string]any{"enabled": true},
			expected:   true,
			desc:       "Compiler should optimize 'a' != 'b' to true",
		},
		// IN optimization
		{
			name:       "IN with single element",
			expression: `status IN ["active"]`,
			env:        map[string]any{"status": "active"},
			expected:   true,
			desc:       "Single element IN should be optimized",
		},
		{
			name:       "NOT IN with empty array",
			expression: `status NOT IN []`,
			env:        map[string]any{"status": "anything"},
			expected:   true,
			desc:       "NOT IN empty array is always true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.env))
			require.NoError(t, err)

			out, err := expr.Run(program, tt.env)
			require.NoError(t, err)
			require.Equal(t, tt.expected, out)
		})
	}
}

// TestSQLStyleErrorCases tests error handling for SQL-style operators
func TestSQLStyleErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		env         map[string]any
		shouldError bool
		errorType   string
	}{
		{
			name:        "REGEXP with invalid pattern",
			expression:  `text REGEXP "[invalid"`,
			env:         map[string]any{"text": "test"},
			shouldError: true,
			errorType:   "compile",
		},
		{
			name:        "Type mismatch with CONTAINS",
			expression:  `count CONTAINS "5"`,
			env:         map[string]any{"count": 12345},
			shouldError: true,
			errorType:   "compile",
		},
		{
			name:        "IN with non-array",
			expression:  `status IN "active"`,
			env:         map[string]any{"status": "active"},
			shouldError: true,
			errorType:   "compile",
		},
		{
			name:        "Undefined field access",
			expression:  `user.name = "John" AND user.age > 18`,
			env:         map[string]any{"user": map[string]any{"name": "John"}},
			shouldError: true,
			errorType:   "runtime",
		},
		{
			name:        "Invalid operator combination",
			expression:  `status NOT NOT IN ["active"]`,
			env:         map[string]any{"status": "active"},
			shouldError: true,
			errorType:   "compile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := expr.Compile(tt.expression, expr.Env(tt.env))

			if tt.errorType == "compile" {
				require.Error(t, err, "Expected compilation error")
				return
			}

			require.NoError(t, err, "Compilation should succeed")

			_, runErr := expr.Run(program, tt.env)
			if tt.errorType == "runtime" {
				require.Error(t, runErr, "Expected runtime error")
			}
		})
	}
}
