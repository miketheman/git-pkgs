package osv

import (
	"math"
	"testing"
)

func TestParseCVSSScore(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "numeric score",
			input:    "9.8",
			expected: 9.8,
		},
		{
			name:     "CVSS 3.1 critical vector",
			input:    "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
			expected: 9.8,
		},
		{
			name:     "CVSS 3.0 high vector",
			input:    "CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H",
			expected: 8.8,
		},
		{
			name:     "CVSS 3.1 medium vector",
			input:    "CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:U/C:L/I:L/A:N",
			expected: 5.4,
		},
		{
			name:     "CVSS 3.1 low vector",
			input:    "CVSS:3.1/AV:L/AC:L/PR:L/UI:N/S:U/C:L/I:N/A:N",
			expected: 3.3,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "CVSS 3.1 scope changed",
			input:    "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
			expected: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseCVSSScore(tt.input)
			if math.Abs(got-tt.expected) > 0.1 {
				t.Errorf("ParseCVSSScore(%q) = %f, want %f", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetSeverityLevel(t *testing.T) {
	tests := []struct {
		name     string
		vuln     *Vulnerability
		expected string
	}{
		{
			name: "critical from CVSS vector",
			vuln: &Vulnerability{
				Severity: []Severity{
					{Type: "CVSS_V3", Score: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"},
				},
			},
			expected: "critical",
		},
		{
			name: "high from CVSS vector",
			vuln: &Vulnerability{
				Severity: []Severity{
					{Type: "CVSS_V3", Score: "CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H"},
				},
			},
			expected: "high",
		},
		{
			name: "medium from CVSS vector",
			vuln: &Vulnerability{
				Severity: []Severity{
					{Type: "CVSS_V3", Score: "CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:U/C:L/I:L/A:N"},
				},
			},
			expected: "medium",
		},
		{
			name: "low from CVSS vector",
			vuln: &Vulnerability{
				Severity: []Severity{
					{Type: "CVSS_V3", Score: "CVSS:3.1/AV:L/AC:L/PR:L/UI:N/S:U/C:L/I:N/A:N"},
				},
			},
			expected: "low",
		},
		{
			name: "database_specific fallback",
			vuln: &Vulnerability{
				DatabaseSpecific: map[string]any{
					"severity": "HIGH",
				},
			},
			expected: "high",
		},
		{
			name:     "unknown when no severity info",
			vuln:     &Vulnerability{},
			expected: "unknown",
		},
		{
			name: "numeric score string",
			vuln: &Vulnerability{
				Severity: []Severity{
					{Type: "CVSS_V3", Score: "7.5"},
				},
			},
			expected: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSeverityLevel(tt.vuln)
			if got != tt.expected {
				t.Errorf("GetSeverityLevel() = %q, want %q", got, tt.expected)
			}
		})
	}
}
