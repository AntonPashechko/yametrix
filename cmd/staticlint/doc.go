// staticlint implements static code checks
//
// - All checks from golang.org/x/tools/go/analysis/passes
//
// - All SA checks from https://staticcheck.io/docs/checks/
//
// - S1004 check from https://staticcheck.io/docs/checks/#ST1019
//
// - Check wrapping errors https://github.com/fatih/errwrap
//
// - Checking for unchecked errors https://github.com/kisielk/errcheck
//
// - Custom check for calling os.Exit in main/main
//
// Example:
//
//	staticlint -SA1000 <project path>
//
// Perform SA1000 analysis for given project.
// For more details run:
//
//	staticlint -help
//
// osexit investigates main package for calling os.Exit from main function. Run this check with following command:
//
//	staticlint -osexit
package main
