package foundation

import (
	stdlog "log"
	"os"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/logrusorgru/aurora"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// LogFormatPlainText outputs logs in plain text without colorization and with timestamp; is the default if log format isn't specified
	LogFormatPlainText = "plaintext"
	// LogFormatConsole outputs logs in plain text with colorization and without timestamp
	LogFormatConsole = "console"
	// LogFormatJSON outputs logs in json including appgroup, app, appversion and other metadata
	LogFormatJSON = "json"
	// LogFormatStackdriver outputs a format similar to JSON format but with 'severity' instead of 'level' field
	LogFormatStackdriver = "stackdriver"
	// LogFormatV3 ouputs an internal format used at Travix in JSON format with nested payload and a specific set of required metadata
	LogFormatV3 = "v3"
)

// initLoggingStackdriver outputs a format similar to JSON format but with 'severity' instead of 'level' field
func initLoggingStackdriver(applicationInfo ApplicationInfo) {

	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999Z"
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "severity"

	// set some default fields added to all logs
	log.Logger = zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()

	// use zerolog for any logs sent via standard log library
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)
}

// initLoggingJSON outputs logs in json including appgroup, app, appversion and other metadata
func initLoggingJSON(applicationInfo ApplicationInfo) {

	// set some default fields added to all logs
	log.Logger = zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()

	// use zerolog for any logs sent via standard log library
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)
}

// initLoggingConsole outputs logs in plain text with colorization and without timestamp
func initLoggingConsole(applicationInfo ApplicationInfo) {

	output := zerolog.ConsoleWriter{
		Out:     os.Stdout,
		NoColor: false,
	}
	output.FormatTimestamp = func(i interface{}) string {
		return ""
	}
	output.FormatCaller = func(i interface{}) string {
		return ""
	}
	output.FormatLevel = func(i interface{}) string {
		return ""
	}

	log.Logger = zerolog.New(output).With().Logger()

	// use zerolog for any logs sent via standard log library
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)
}

// initLoggingPlainText outputs logs in plain text without colorization and with timestamp; is the default if log format isn't specified
func initLoggingPlainText(applicationInfo ApplicationInfo) {
	output := zerolog.ConsoleWriter{
		Out:     os.Stdout,
		NoColor: true,
	}

	log.Logger = zerolog.New(output).With().Logger()

	// use zerolog for any logs sent via standard log library
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)
}

var (
	sequenceID uint64
)

type v3Error struct {
	Message string `json:"message"`
}

type messageIDHook struct{}

func (h messageIDHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Str("messageuniqueid", uuid.New().String())
	e.Uint64("sequenceid", atomic.AddUint64(&sequenceID, 1))
}

// initLoggingV3 ouputs an internal format used at Travix in JSON format with nested payload and a specific set of required metadata
func initLoggingV3(applicationInfo ApplicationInfo) {

	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999Z"
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "loglevel"

	zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
		switch l {
		case zerolog.DebugLevel:
			return "DEBUG"
		case zerolog.InfoLevel:
			return "INFO"
		case zerolog.WarnLevel:
			return "WARN"
		case zerolog.ErrorLevel:
			return "ERROR"
		case zerolog.FatalLevel:
			return "FATAL"
		case zerolog.PanicLevel:
			return "PANIC"
		case zerolog.NoLevel:
			return ""
		}
		return ""
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	source := struct {
		AppGroup   string `json:"appgroup"`
		AppName    string `json:"appname"`
		AppVersion string `json:"appversion"`
		Hostname   string `json:"hostname"`
	}{
		applicationInfo.AppGroup,
		applicationInfo.App,
		applicationInfo.Version,
		hostname,
	}

	// set some default fields added to all logs
	log.Logger = zerolog.New(os.Stdout).Hook(messageIDHook{}).With().
		Timestamp().
		Str("logformat", "v3").
		Str("messagetype", "estafette").
		Str("messagetypeversion", "0.0.0").
		Interface("source", source).
		Logger()

	// Have the error message under and object in "error" instead of in a raw string.
	zerolog.ErrorMarshalFunc = func(err error) interface{} {
		return v3Error{err.Error()}
	}

	// use zerolog for any logs sent via standard log library
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)
}

// logStartupMessage logs a default startup message for any Estafette application
func logStartupMessage(applicationInfo ApplicationInfo) {
	log.Info().
		Str("branch", applicationInfo.Branch).
		Str("revision", applicationInfo.Revision).
		Str("buildDate", applicationInfo.BuildDate).
		Str("goVersion", applicationInfo.GoVersion()).
		Str("os", applicationInfo.OperatingSystem()).
		Msgf("Starting %v version %v...", applicationInfo.App, applicationInfo.Version)
}

// logStartupMessageConsole logs a default startup message for any Estafette application in bold
func logStartupMessageConsole(applicationInfo ApplicationInfo) {
	log.Info().
		Str("branch", applicationInfo.Branch).
		Str("revision", applicationInfo.Revision).
		Str("buildDate", applicationInfo.BuildDate).
		Str("goVersion", applicationInfo.GoVersion()).
		Str("os", applicationInfo.OperatingSystem()).
		Msg(aurora.Sprintf("Starting %v version %v...", aurora.Bold(applicationInfo.App), aurora.Bold(applicationInfo.Version)))
}

// logStartupMessageV3 logs a v3 startup message for any Estafette application
func logStartupMessageV3(applicationInfo ApplicationInfo) {
	startupProps := struct {
		Branch    string `json:"branch"`
		Revision  string `json:"revision"`
		BuildDate string `json:"buildDate"`
		GoVersion string `json:"goVersion"`
		Os        string `json:"os"`
	}{
		applicationInfo.Branch,
		applicationInfo.Revision,
		applicationInfo.BuildDate,
		applicationInfo.GoVersion(),
		applicationInfo.OperatingSystem(),
	}

	log.Info().
		Interface("payload", startupProps).
		Msgf("Starting %v version %v...", applicationInfo.App, applicationInfo.Version)
}
