package cmd

import (
	"bytes"
	"testing"

	"github.com/git-pkgs/resolve"
)

func TestWriteResolveTree(t *testing.T) {
	result := &resolve.Result{
		Manager:   "npm",
		Ecosystem: "npm",
		Direct: []*resolve.Dep{
			{
				PURL:    "pkg:npm/express@4.18.2",
				Name:    "express",
				Version: "4.18.2",
				Deps: []*resolve.Dep{
					{
						PURL:    "pkg:npm/accepts@1.3.8",
						Name:    "accepts",
						Version: "1.3.8",
					},
					{
						PURL:    "pkg:npm/body-parser@1.20.1",
						Name:    "body-parser",
						Version: "1.20.1",
					},
				},
			},
			{
				PURL:    "pkg:npm/lodash@4.17.21",
				Name:    "lodash",
				Version: "4.17.21",
			},
		},
	}

	var buf bytes.Buffer
	writeResolveTree(&buf, result)

	expected := `npm (npm)
├── express@4.18.2
│   ├── accepts@1.3.8
│   └── body-parser@1.20.1
└── lodash@4.17.21
`
	if buf.String() != expected {
		t.Errorf("unexpected tree output:\ngot:\n%s\nwant:\n%s", buf.String(), expected)
	}
}

func TestWriteResolveTreeSingleDep(t *testing.T) {
	result := &resolve.Result{
		Manager:   "cargo",
		Ecosystem: "cargo",
		Direct: []*resolve.Dep{
			{
				PURL:    "pkg:cargo/serde@1.0.0",
				Name:    "serde",
				Version: "1.0.0",
			},
		},
	}

	var buf bytes.Buffer
	writeResolveTree(&buf, result)

	expected := `cargo (cargo)
└── serde@1.0.0
`
	if buf.String() != expected {
		t.Errorf("unexpected tree output:\ngot:\n%s\nwant:\n%s", buf.String(), expected)
	}
}

func TestWriteResolveTreeDeepNesting(t *testing.T) {
	result := &resolve.Result{
		Manager:   "npm",
		Ecosystem: "npm",
		Direct: []*resolve.Dep{
			{
				PURL:    "pkg:npm/a@1.0.0",
				Name:    "a",
				Version: "1.0.0",
				Deps: []*resolve.Dep{
					{
						PURL:    "pkg:npm/b@2.0.0",
						Name:    "b",
						Version: "2.0.0",
						Deps: []*resolve.Dep{
							{
								PURL:    "pkg:npm/c@3.0.0",
								Name:    "c",
								Version: "3.0.0",
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	writeResolveTree(&buf, result)

	expected := `npm (npm)
└── a@1.0.0
    └── b@2.0.0
        └── c@3.0.0
`
	if buf.String() != expected {
		t.Errorf("unexpected tree output:\ngot:\n%s\nwant:\n%s", buf.String(), expected)
	}
}

func TestWriteResolveTreeNoDeps(t *testing.T) {
	result := &resolve.Result{
		Manager:   "pip",
		Ecosystem: "pypi",
		Direct:    nil,
	}

	var buf bytes.Buffer
	writeResolveTree(&buf, result)

	expected := "pip (pypi)\n"
	if buf.String() != expected {
		t.Errorf("unexpected tree output:\ngot:\n%s\nwant:\n%s", buf.String(), expected)
	}
}
