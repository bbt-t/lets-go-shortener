// Run analyzer: letsanalyze [directory].
package main

import (
	"strings"

	"myanalyzer/os_exit_analyzer"

	"github.com/gostaticanalysis/elseless"
	"github.com/gostaticanalysis/nakedreturn"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	// Checks all 'SA' checks from static check.
	var myChecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		if strings.Contains(v.Analyzer.Name, "SA") {
			myChecks = append(myChecks, v.Analyzer)
		}
	}

	for _, v := range simple.Analyzers {
		// Omit comparison with boolean constant.
		if v.Analyzer.Name == "S1002" {
			myChecks = append(myChecks, v.Analyzer)
			break
		}
	}

	// Checks printf.Analyzer, shadow.Analyzer, os_exit_analyzer.OsExitAnalyzer.
	myChecks = append(myChecks, printf.Analyzer, shadow.Analyzer, os_exit_analyzer.OsExitAnalyzer)
	myChecks = append(myChecks, nakedreturn.Analyzer)
	myChecks = append(myChecks, elseless.Analyzer)
	multichecker.Main(myChecks...)
}
