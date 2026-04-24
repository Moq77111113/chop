package block

// RPC method names exposed by every block over stdio JSON-RPC.
// The supervisor calls these; block.RunBlock registers handlers for them.
// Wire contract — change here = change on both sides.
const (
	MethodInfo     = "info"
	MethodSnapshot = "snapshot"
	MethodApply    = "apply"
	MethodAction   = "action"
)
