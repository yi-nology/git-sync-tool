package appinfo

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
	AppName   = "git-manage-service"
)

// Set sets the app info at startup, typically from ldflags-injected values.
func Set(version, buildTime, gitCommit string) {
	Version = version
	BuildTime = buildTime
	GitCommit = gitCommit
}
