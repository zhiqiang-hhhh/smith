package version

import "runtime/debug"

// Build-time parameters set via -ldflags.

var (
	Version   = "devel"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// A user may install crush using `go install github.com/zhiqiang-hhhh/smith@latest`.
// without -ldflags, in which case the version above is unset. As a workaround
// we use the embedded build version that *is* set when using `go install` (and
// is only set for `go install` and not for `go build`).
func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	mainVersion := info.Main.Version
	if mainVersion != "" && mainVersion != "(devel)" {
		Version = mainVersion
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if Commit == "unknown" && s.Value != "" {
				Commit = s.Value
			}
		case "vcs.time":
			if BuildDate == "unknown" && s.Value != "" {
				BuildDate = s.Value
			}
		}
	}
}

// Full returns a human-readable version string including commit and build
// date.
func Full() string {
	s := Version
	if Commit != "unknown" {
		c := Commit
		if len(c) > 8 {
			c = c[:8]
		}
		s += " (" + c + ")"
	}
	if BuildDate != "unknown" {
		s += " built " + BuildDate
	}
	return s
}
