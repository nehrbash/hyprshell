package core

import "encoding/json"

type DockAction string
type DockSignal struct {
	Type DockAction
	Msg  string
}

// PrettyPrint to print struct in a readable way
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
