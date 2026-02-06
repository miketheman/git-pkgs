package cmd

import (
	"testing"

	"github.com/git-pkgs/vulns"
)

func TestBuildVersRange(t *testing.T) {
	tests := []struct {
		name      string
		ranges    []vulns.Range
		ecosystem string
		want      string
	}{
		{
			name: "single introduced/fixed pair",
			ranges: []vulns.Range{
				{
					Type: "ECOSYSTEM",
					Events: []vulns.Event{
						{Introduced: "1.0.0"},
						{Fixed: "1.5.0"},
					},
				},
			},
			ecosystem: "npm",
			want:      "vers:npm/>=1.0.0|<1.5.0",
		},
		{
			name: "multiple introduced/fixed pairs in one range",
			ranges: []vulns.Range{
				{
					Type: "ECOSYSTEM",
					Events: []vulns.Event{
						{Introduced: "1.0.0"},
						{Fixed: "1.5.0"},
						{Introduced: "2.0.0"},
						{Fixed: "2.5.0"},
					},
				},
			},
			ecosystem: "npm",
			want:      "vers:npm/>=1.0.0|<1.5.0|>=2.0.0|<2.5.0",
		},
		{
			name: "introduced from zero with fix",
			ranges: []vulns.Range{
				{
					Type: "ECOSYSTEM",
					Events: []vulns.Event{
						{Introduced: "0"},
						{Fixed: "1.2.0"},
					},
				},
			},
			ecosystem: "npm",
			want:      "vers:npm/<1.2.0",
		},
		{
			name: "introduced from zero then reintroduced",
			ranges: []vulns.Range{
				{
					Type: "ECOSYSTEM",
					Events: []vulns.Event{
						{Introduced: "0"},
						{Fixed: "1.2.0"},
						{Introduced: "2.0.0"},
						{Fixed: "2.3.0"},
					},
				},
			},
			ecosystem: "npm",
			want:      "vers:npm/<1.2.0|>=2.0.0|<2.3.0",
		},
		{
			name: "lastAffected instead of fixed",
			ranges: []vulns.Range{
				{
					Type: "ECOSYSTEM",
					Events: []vulns.Event{
						{Introduced: "1.0.0"},
						{LastAffected: "1.9.9"},
					},
				},
			},
			ecosystem: "PyPI",
			want:      "vers:PyPI/>=1.0.0|<=1.9.9",
		},
		{
			name: "no upper bound",
			ranges: []vulns.Range{
				{
					Type: "ECOSYSTEM",
					Events: []vulns.Event{
						{Introduced: "3.0.0"},
					},
				},
			},
			ecosystem: "npm",
			want:      "vers:npm/>=3.0.0",
		},
		{
			name: "all versions affected (introduced 0 with no fix)",
			ranges: []vulns.Range{
				{
					Type: "ECOSYSTEM",
					Events: []vulns.Event{
						{Introduced: "0"},
					},
				},
			},
			ecosystem: "npm",
			want:      "vers:npm/*",
		},
		{
			name: "multiple ranges",
			ranges: []vulns.Range{
				{
					Type: "ECOSYSTEM",
					Events: []vulns.Event{
						{Introduced: "1.0.0"},
						{Fixed: "1.1.0"},
					},
				},
				{
					Type: "ECOSYSTEM",
					Events: []vulns.Event{
						{Introduced: "2.0.0"},
						{Fixed: "2.1.0"},
					},
				},
			},
			ecosystem: "npm",
			want:      "vers:npm/>=1.0.0|<1.1.0|>=2.0.0|<2.1.0",
		},
		{
			name:      "empty ranges",
			ranges:    []vulns.Range{},
			ecosystem: "npm",
			want:      "",
		},
		{
			name: "skip GIT range type",
			ranges: []vulns.Range{
				{
					Type: "GIT",
					Events: []vulns.Event{
						{Introduced: "abc123"},
						{Fixed: "def456"},
					},
				},
			},
			ecosystem: "npm",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &vulns.Vulnerability{
				Affected: []vulns.Affected{{
					Package: vulns.Package{
						Ecosystem: tt.ecosystem,
						Name:      "test-pkg",
					},
					Ranges: tt.ranges,
				}},
			}
			got := buildVersRange(v, tt.ecosystem, "test-pkg")
			if got != tt.want {
				t.Errorf("buildVersRange() = %q, want %q", got, tt.want)
			}
		})
	}
}
