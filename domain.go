package foundation

import "runtime"

// ApplicationInfo contains basic information about an application
type ApplicationInfo struct {
	AppGroup  string
	App       string
	Version   string
	Branch    string
	Revision  string
	BuildDate string
}

// OperatingSystem returns the current operating system
func (ai *ApplicationInfo) OperatingSystem() string {
	return runtime.GOOS
}

// GoVersion returns the golang version used to build an application
func (ai *ApplicationInfo) GoVersion() string {
	return runtime.Version()
}

// NewApplicationInfo returns an ApplicationInfo object
func NewApplicationInfo(appgroup, app, version, branch, revision, buildDate string) ApplicationInfo {
	return ApplicationInfo{
		AppGroup:  appgroup,
		App:       app,
		Version:   version,
		Branch:    branch,
		Revision:  revision,
		BuildDate: buildDate,
	}
}
