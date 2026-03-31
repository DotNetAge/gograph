// Package main provides the entry point for c-archive build.
// All CGO functions are implemented in the binding package.
package main

/*
#include "../gograph_c.h"
*/
import "C"

// Import binding package to make all functions available
import _ "github.com/DotNetAge/gograph/binding"

func main() {}
