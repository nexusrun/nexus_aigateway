// Package main is a loader test fixture whose AIGatewayPlugin symbol has the
// wrong type.
package main

import _ "github.com/nexusrun/nexus_aigateway/pluginapi"

// AIGatewayPlugin is deliberately not a constructor or a pluginapi.Plugin.
var AIGatewayPlugin = 42

func main() {}
