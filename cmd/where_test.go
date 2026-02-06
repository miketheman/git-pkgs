package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchFileForPackage(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		packageName string
		wantLines   []int
	}{
		{
			name:        "matches package name in dependency line",
			content:     `    "six": "^1.0.0",`,
			packageName: "six",
			wantLines:   []int{1},
		},
		{
			name:        "does not match inside integrity hash",
			content:     `      "integrity": "sha512-abc123SIxia456def==",`,
			packageName: "six",
			wantLines:   nil,
		},
		{
			name: "matches real dependency but not hash containing same text",
			content: `{
  "node_modules/six": {
    "version": "1.16.0",
    "resolved": "https://registry.npmjs.org/six/-/six-1.16.0.tgz",
    "integrity": "sha512-ySIxiAbcSIxcdefgSIxyz=="
  }
}`,
			packageName: "six",
			wantLines:   []int{2, 4},
		},
		{
			name:        "case insensitive match",
			content:     `    "Six": "^2.0.0",`,
			packageName: "six",
			wantLines:   []int{1},
		},
		{
			name:        "matches with special regex characters in name",
			content:     `    "@scope/my.pkg": "^1.0.0",`,
			packageName: "@scope/my.pkg",
			wantLines:   []int{1},
		},
		{
			name:        "no match when package name is substring of another word",
			content:     `    "sixteenth": "^1.0.0",`,
			packageName: "six",
			wantLines:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "package-lock.json")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			matches, err := searchFileForPackage(path, "package-lock.json", tt.packageName, "npm", 0)
			if err != nil {
				t.Fatal(err)
			}

			if len(matches) != len(tt.wantLines) {
				t.Fatalf("got %d matches, want %d", len(matches), len(tt.wantLines))
			}

			for i, m := range matches {
				if m.LineNumber != tt.wantLines[i] {
					t.Errorf("match %d: got line %d, want %d", i, m.LineNumber, tt.wantLines[i])
				}
			}
		})
	}
}
