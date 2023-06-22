package context_keys

type UserKey string
type MachineKey string

var UserContextKey = UserKey("user")
var MachineContextKey = MachineKey("machine")
