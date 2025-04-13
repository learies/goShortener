package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/learies/goShortener/cmd/staticlint/analyzer"
)

func main() {
	// Стандартные анализаторы из golang.org/x/tools/go/analysis/passes
	standardAnalyzers := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		structtag.Analyzer,
	}

	// Анализаторы из staticcheck.io
	var staticcheckAnalyzers []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		staticcheckAnalyzers = append(staticcheckAnalyzers, v.Analyzer)
	}
	for _, v := range simple.Analyzers {
		staticcheckAnalyzers = append(staticcheckAnalyzers, v.Analyzer)
	}
	for _, v := range stylecheck.Analyzers {
		staticcheckAnalyzers = append(staticcheckAnalyzers, v.Analyzer)
	}

	// Собственный анализатор
	customAnalyzers := []*analysis.Analyzer{
		analyzer.OSExitAnalyzer,
	}

	// Объединяем все анализаторы
	allAnalyzers := append(standardAnalyzers, staticcheckAnalyzers...)
	allAnalyzers = append(allAnalyzers, customAnalyzers...)

	// Запуск multichecker
	multichecker.Main(allAnalyzers...)
}
