// Package main is a loader test fixture whose GoModelPlugin symbol has the
// wrong type.
package main

import _ "github.com/nexusrun/nexus_aigateway/pluginapi"

// GoModelPlugin is deliberately not a constructor or a pluginapi.Plugin.
var GoModelPlugin = 42

func main() {}
