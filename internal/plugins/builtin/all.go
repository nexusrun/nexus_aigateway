// Package builtin lists the plugins compiled into GoModel.
package builtin

import (
	"github.com/nexusrun/nexus_aigateway/internal/plugins/builtin/headeredit"
	"github.com/nexusrun/nexus_aigateway/internal/plugins/builtin/llmaltering"
	"github.com/nexusrun/nexus_aigateway/internal/plugins/builtin/llmjudge"
	"github.com/nexusrun/nexus_aigateway/internal/plugins/builtin/presidio"
	"github.com/nexusrun/nexus_aigateway/internal/plugins/builtin/routeexample"
	"github.com/nexusrun/nexus_aigateway/internal/plugins/builtin/stringreplace"
	"github.com/nexusrun/nexus_aigateway/internal/plugins/builtin/systemprompt"
	"github.com/nexusrun/nexus_aigateway/pluginapi"
)

// All returns a factory per built-in plugin, in registration order.
func All() []func() pluginapi.Plugin {
	return []func() pluginapi.Plugin{
		systemprompt.New,
		llmaltering.New,
		stringreplace.New,
		headeredit.New,
		llmjudge.New,
		presidio.New,
		routeexample.New,
	}
}
