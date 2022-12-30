package module

import "fmt"

const (
	MajorVersion = 1
	MinorVersion = 0
	FixVersion   = 0
	CommitHash   = ""
	VERSION      = "1.0.0"
)

func GetVersion() string {
	if CommitHash != "" {
		return fmt.Sprintf("meta-lib-v%v.%v.%v-%s", MajorVersion, MinorVersion, FixVersion, CommitHash)
	} else {
		return fmt.Sprintf("meta-lib-v%v.%v.%v", MajorVersion, MinorVersion, FixVersion)
	}
}
