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
package main
