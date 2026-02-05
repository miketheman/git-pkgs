package cmd

import (
	"testing"

	"github.com/git-pkgs/git-pkgs/internal/osv"
)

func TestBuildVersRange(t *testing.T) {
	tests := []struct {
		name      string
		ranges    []osv.Range
		ecosystem string
		want      string
	}{
		{
			name: "single introduced/fixed pair",
			ranges: []osv.Range{
				{
					Type: "ECOSYSTEM",
					Events: []osv.Event{
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
			ranges: []osv.Range{
				{
					Type: "ECOSYSTEM",
					Events: []osv.Event{
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
			ranges: []osv.Range{
				{
					Type: "ECOSYSTEM",
					Events: []osv.Event{
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
			ranges: []osv.Range{
				{
					Type: "ECOSYSTEM",
					Events: []osv.Event{
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
			ranges: []osv.Range{
				{
					Type: "ECOSYSTEM",
					Events: []osv.Event{
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
			ranges: []osv.Range{
				{
					Type: "ECOSYSTEM",
					Events: []osv.Event{
						{Introduced: "3.0.0"},
					},
				},
			},
			ecosystem: "npm",
			want:      "vers:npm/>=3.0.0",
		},
		{
			name: "all versions affected (introduced 0 with no fix)",
			ranges: []osv.Range{
				{
					Type: "ECOSYSTEM",
					Events: []osv.Event{
						{Introduced: "0"},
					},
				},
			},
			ecosystem: "npm",
			want:      "vers:npm/*",
		},
		{
			name: "multiple ranges",
			ranges: []osv.Range{
				{
					Type: "ECOSYSTEM",
					Events: []osv.Event{
						{Introduced: "1.0.0"},
						{Fixed: "1.1.0"},
					},
				},
				{
					Type: "ECOSYSTEM",
					Events: []osv.Event{
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
			ranges:    []osv.Range{},
			ecosystem: "npm",
			want:      "",
		},
		{
			name: "skip GIT range type",
			ranges: []osv.Range{
				{
					Type: "GIT",
					Events: []osv.Event{
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
			got := buildVersRange(tt.ranges, tt.ecosystem)
			if got != tt.want {
				t.Errorf("buildVersRange() = %q, want %q", got, tt.want)
			}
		})
	}
}
