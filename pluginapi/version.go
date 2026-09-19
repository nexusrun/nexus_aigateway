// Package pluginapi is the public contract between AIGateway and its plugins.
//
// A plugin imports this package and nothing else from AIGateway. The package
// depends on the standard library only, so a plugin built as a shared object
// shares exactly the standard library and this package with the host.
//
// Plugins implement [Plugin] plus any of the optional hook interfaces
// ([RequestHook], [PromptHook], [ResponseHook], [StreamHook], [RouteStrategy],
// [CompleteHook]). Hooks receive an [Exchange]: a unified, dialect-neutral view
// of the request ([Prompt]) and response ([Completion]) that AIGateway maps back
// onto the wire format after the hook returns.
package pluginapi

// Version is the pluginapi contract version. It is informational: AIGateway
// reports it in diagnostics and stamps it into [BuildInfo], but never uses it
// to accept or reject a plugin (the Go toolchain already enforces that a
// shared object was built from identical sources).
const Version = "0.2.0"
