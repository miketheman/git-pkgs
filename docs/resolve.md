# Resolve

`git pkgs resolve` runs the detected package manager's dependency graph command and parses the output into a normalized dependency tree. Every dependency gets a [PURL](https://github.com/package-url/purl-spec) (Package URL), a standard identifier that encodes the ecosystem, name, and version in one string.

```bash
$ git pkgs resolve
npm (npm)
├── express@4.18.2
│   ├── accepts@1.3.8
│   └── body-parser@1.20.1
└── lodash@4.17.21
```

Pass `-f json` for machine-readable output:

```bash
$ git pkgs resolve -f json
{
  "Manager": "npm",
  "Ecosystem": "npm",
  "Direct": [
    {
      "PURL": "pkg:npm/express@4.18.2",
      "Name": "express",
      "Version": "4.18.2",
      "Deps": [
        {
          "PURL": "pkg:npm/accepts@1.3.8",
          "Name": "accepts",
          "Version": "1.3.8",
          "Deps": []
        }
      ]
    }
  ]
}
```

The output goes to stdout. Status lines (detected manager, command being run) go to stderr, so you can pipe output directly into other tools.

## Output structure

Each result contains:

- **Manager** - the package manager that ran (npm, cargo, gomod, etc.)
- **Ecosystem** - the package ecosystem (npm, cargo, golang, pypi, etc.)
- **Direct** - the top-level dependencies, each with transitive deps nested under `Deps`

For managers that produce tree output (npm, cargo, go, maven, uv, etc.), `Deps` contains the transitive dependency tree. For managers that only produce flat lists (pip, conda, bundler, helm, nuget, conan), `Deps` is null.

## PURLs

Every dependency includes a PURL string. PURLs follow the [package-url spec](https://github.com/package-url/purl-spec) and look like `pkg:npm/%40scope/name@1.0.0`. They're useful for cross-referencing against vulnerability databases like [OSV](https://osv.dev) and for identifying packages across tools that speak PURL.

Scoped npm packages get URL-encoded: `@babel/core` becomes `pkg:npm/%40babel/core@7.23.0`.

## Supported parsers

24 package managers are supported. The manager name in the first column is what you pass to `-m`.

| Manager | Ecosystem | Output format |
|---------|-----------|---------------|
| npm | npm | JSON tree |
| pnpm | npm | JSON tree |
| yarn | npm | NDJSON tree |
| bun | npm | Text tree |
| cargo | cargo | JSON graph |
| gomod | golang | Edge list |
| pip | pypi | JSON flat |
| uv | pypi | Text tree |
| poetry | pypi | Text tree |
| conda | conda | JSON flat |
| bundler | gem | Text flat |
| maven | maven | Text tree |
| gradle | maven | Text tree |
| composer | packagist | Text tree |
| nuget | nuget | Tabular |
| swift | swift | JSON tree |
| pub | pub | Text tree |
| mix | hex | Text tree |
| rebar3 | hex | Text tree |
| stack | hackage | JSON flat |
| lein | clojars | Text tree |
| conan | conan | Custom |
| deno | deno | JSON flat |
| helm | helm | Tabular |

## Flags

```
-f, --format     Output format: text, json (default text)
-m, --manager    Override detected package manager
-e, --ecosystem  Filter to specific ecosystem
    --raw        Print raw manager output instead of parsed JSON
    --dry-run    Show what would be run without executing
-x, --extra      Extra arguments to pass to package manager
-t, --timeout    Timeout for resolve operation (default 5m)
-q, --quiet      Suppress status output on stderr
```

## Raw mode

`--raw` skips parsing and prints the manager's output as-is. Useful for debugging or when you need the original format:

```bash
$ git pkgs resolve --raw
{
  "version": "1.0.0",
  "name": "my-project",
  "dependencies": {
    "express": {
      "version": "4.18.2",
      ...
    }
  }
}
```

## Multi-ecosystem projects

If your project has multiple lockfiles, resolve runs for each detected manager:

```bash
$ git pkgs resolve
bundler (gem)
├── rails@7.1.0
│   └── actionpack@7.1.0
└── puma@6.4.0

npm (npm)
├── express@4.18.2
└── lodash@4.17.21
```

With `-f json`, each manager produces a separate JSON object:

```bash
$ git pkgs resolve -q -f json
{"Manager":"bundler","Ecosystem":"gem","Direct":[...]}
{"Manager":"npm","Ecosystem":"npm","Direct":[...]}
```

Filter to one ecosystem with `-e`:

```bash
git pkgs resolve -e npm
```

## Examples

The JSON format works well with [jq](https://jqlang.github.io/jq/) and standard unix tools.

### List all dependency names and versions

```bash
git pkgs resolve -q -f json | jq -r '.Direct[] | "\(.Name) \(.Version)"'
```

```
express 4.18.2
react 18.2.0
```

### Extract just the PURLs

```bash
git pkgs resolve -q -f json | jq -r '.. | .PURL? // empty'
```

```
pkg:npm/express@4.18.2
pkg:npm/accepts@1.3.8
pkg:npm/react@18.2.0
```

### Count total dependencies (including transitive)

```bash
git pkgs resolve -q -f json | jq '[.. | .PURL? // empty] | length'
```

### Check a specific package against OSV

Grab a PURL from resolve output and query the [OSV API](https://osv.dev) for known vulnerabilities:

```bash
git pkgs resolve -q -f json \
  | jq -r '.. | .PURL? // empty' \
  | while read purl; do
      curl -s "https://api.osv.dev/v1/query" \
        -d "{\"package\":{\"purl\":\"$purl\"}}" \
        | jq -r --arg p "$purl" 'select(.vulns) | "\($p) \(.vulns | length) vulns"'
    done
```

### Find all packages matching a name

```bash
git pkgs resolve -q -f json | jq '[.. | select(.Name? == "lodash")]'
```

### Show why a transitive dependency is in the tree

Find every path from a direct dependency down to a specific package. This tells you which of your dependencies pulled it in:

```bash
git pkgs resolve -q -f json | jq --arg pkg "mime-types" '
  def paths_to($name):
    if .Name == $name then [.Name]
    elif (.Deps // []) | length > 0 then
      .Name as $n | .Deps[] | paths_to($name) | select(length > 0) | [$n] + .
    else empty
    end;
  [.Direct[] | paths_to($pkg)] | unique[] | join(" > ")
'
```

```
express > accepts > mime-types
```

This walks the dependency tree recursively and prints each chain that leads to the package. If `mime-types` appears under multiple direct dependencies, you'll see all paths.

### Diff resolved dependencies between branches

```bash
diff <(git stash && git pkgs resolve -q -f json | jq -r '.. | .PURL? // empty' | sort) \
     <(git stash pop && git pkgs resolve -q -f json | jq -r '.. | .PURL? // empty' | sort)
```

### Save a snapshot for later comparison

```bash
git pkgs resolve -q -f json > deps-$(date +%Y%m%d).json
```

### Feed into a Go program

The output matches the `resolve.Result` struct from [github.com/git-pkgs/resolve](https://github.com/git-pkgs/resolve), so you can decode it directly:

```go
import (
	"encoding/json"
	"os/exec"

	"github.com/git-pkgs/resolve"
)

out, _ := exec.Command("git", "pkgs", "resolve", "-q", "-f", "json").Output()
var result resolve.Result
json.Unmarshal(out, &result)

for _, dep := range result.Direct {
	fmt.Println(dep.PURL)
}
```

## How it works

The resolve command calls the manager's dependency graph command (defined in the [managers](https://github.com/git-pkgs/managers) library), captures stdout, and passes it to the [resolve](https://github.com/git-pkgs/resolve) library for parsing. Each parser knows the output format for its manager and builds a normalized `Result`.

The parsers use an init-registration pattern similar to `database/sql` drivers. The `resolve` package defines the types and `Parse()` function, and the `resolve/parsers` subpackage registers all parser implementations at import time.
