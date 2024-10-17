package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	exitcheck "github.com/igortoigildin/go-metrics-altering/cmd/staticlint/checker"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/unused"
)

// Config — config file name.
const Config = `config.json`

// ConfigData defines struct with config file.
type ConfigData struct {
	Staticcheck []string
}

func main() {
	appfile, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	if err != nil {
		log.Fatal(err)
	}

	// cfg - JSON file with statickcheck id of checks needed for multichecker.
	var cfg ConfigData
	if err = json.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}

	// mychecks collects all analyzers inclueded in multichecker.
	mychecks := []*analysis.Analyzer{
		exitcheck.ExitCheckAnalyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		nilness.Analyzer,
		errcheck.Analyzer,
		unused.Analyzer.Analyzer,
	}
	checks := make(map[string]bool)
	for _, v := range cfg.Staticcheck {
		checks[v] = true
	}
	// Add analyzers which stated in config fail.
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	// Muiltichecker examines Go source code and reports suspicious constructs,
	// such as Printf calls whose arguments do not align with the format
	// string. It uses heuristics that do not guarantee all reports are
	// genuine problems, but it can find errors not caught by the compilers.
	//
	// Multichecker registers the following analyzers:
	// exitcheck	reports os.Exit function usage
	// printf 		checks consistency of Printf format strings and arguments
	// shadow		checks for shadowed variables
	// structtag 	checks struct field tags are well formed
	// errcheck		checks unchecked errors in Go code
	// unused		finds unused code
	// nilness		inspects the control-flow graph of an SSA function and reports
	// errors such as nil pointer dereferences and degenerate nil pointer comparisons
	//
	// By default all analyzers are run.
	// To select specific analyzers, use the -NAME flag for each one,
	// or -NAME=false to run all analyzers not explicitly disabled.
	//
	// For basic usage, just give the package path of interest as the first argument:
	//
	// multichecker cmd/testdata
	//
	// To check all packages beneath the current directory:
	// multichecker ./...
	multichecker.Main(
		mychecks...,
	)
}
