package foundation

import (
	"fmt"
	stdlog "log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type v3Error struct {
	Message string `json:"message"`
}

var (
	goVersion = runtime.Version()

	// seed random number
	r = rand.New(rand.NewSource(time.Now().UnixNano()))

	sequenceID uint64 = 0
)

type messageIDHook struct{}

func (h messageIDHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Str("messageuniqueid", uuid.New().String())
	e.Uint64("sequenceid", atomic.AddUint64(&sequenceID, 1))
}

// InitV3Logging initializes logging to log everything as json in v3 log format
func InitV3Logging(appgroup, app, version, branch, revision, buildDate string) {

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
		appgroup,
		app,
		version,
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

	LogStartupMessage(appgroup, app, version, branch, revision, buildDate)
}

// InitStackdriverLogging initializes logging to log everything as json optimized for Stackdriver logging
func InitStackdriverLogging(appgroup, app, version, branch, revision, buildDate string) {

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

	LogStartupMessage(appgroup, app, version, branch, revision, buildDate)
}

// InitConsoleLogging initializes a console logger for use by Estafette CI extensions
func InitConsoleLogging(appgroup, app, version, branch, revision, buildDate string) {

	output := zerolog.ConsoleWriter{
		Out: os.Stdout,
		PartsOrder: []string{
			zerolog.LevelFieldName,
			zerolog.MessageFieldName,
		},
	}
	output.FormatTimestamp = func(i interface{}) string {
		return ""
	}
	output.FormatCaller = func(i interface{}) string {
		return ""
	}
	output.FormatLevel = func(i interface{}) string {
		if ll, ok := i.(string); ok {
			switch ll {
			case "debug":
				return colorizeStart(colorizeGray)
			case "info":
				return colorizeStart(colorizeBold)
			case "warn",
				"error",
				"fatal",
				"panic":
			default:
			}
		}
		return colorizeStart(colorizeReset)
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s%s", i, colorizeEnd())
	}

	log.Logger = zerolog.New(output).With().Logger()

	// use zerolog for any logs sent via standard log library
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)

	LogStartupMessage(appgroup, app, version, branch, revision, buildDate)
}

// LogStartupMessage logs a default startup message for any Estafette application
func LogStartupMessage(appgroup, app, version, branch, revision, buildDate string) {
	startupProps := struct {
		Branch    string `json:"branch"`
		Revision  string `json:"revision"`
		BuildDate string `json:"buildDate"`
		GoVersion string `json:"goVersion"`
		Os        string `json:"os"`
	}{
		branch,
		revision,
		buildDate,
		goVersion,
		runtime.GOOS,
	}

	log.Info().
		Interface("payload", startupProps).
		Msgf("Starting %v version %v...", app, version)
}

type colorizeFormat int

const (
	colorizeWhite colorizeFormat = 0
	colorizeGray  colorizeFormat = 1
	colorizeBold  colorizeFormat = 2
	colorizeReset colorizeFormat = 3
)

func colorizeStart(c colorizeFormat) string {

	switch c {
	case colorizeWhite:
		return "\x1b[37m"
	case colorizeGray:
		return "\x1b[38;5;250m"
	case colorizeBold:
		return "\x1b[1m"
	}

	return "\x1b[0m"
}

func colorizeEnd() string {
	return fmt.Sprintf("\x1b[0m")
}

// InitMetrics initializes the prometheus endpoint /metrics on port 9101
func InitMetrics() {
	InitMetricsWithPort(9101)
}

// InitMetricsWithPort initializes the prometheus endpoint /metrics on specified port
func InitMetricsWithPort(port int) {
	// start prometheus
	go func() {
		portString := fmt.Sprintf(":%v", port)
		log.Debug().
			Str("port", portString).
			Msg("Serving Prometheus metrics...")

		http.Handle("/metrics", promhttp.Handler())

		if err := http.ListenAndServe(portString, nil); err != nil {
			log.Fatal().Err(err).Msg("Starting Prometheus listener failed")
		}
	}()
}

// InitGracefulShutdownHandling generates the channel that listens to SIGTERM and a waitgroup to use for finishing work when shutting down
func InitGracefulShutdownHandling() (gracefulShutdown chan os.Signal, waitGroup *sync.WaitGroup) {

	// define channel used to gracefully shutdown the application
	gracefulShutdown = make(chan os.Signal)

	signal.Notify(gracefulShutdown, syscall.SIGTERM, syscall.SIGINT)

	waitGroup = &sync.WaitGroup{}

	return gracefulShutdown, waitGroup
}

// HandleGracefulShutdown waits for SIGTERM to unblock gracefulShutdown and waits for the waitgroup to await pending work
func HandleGracefulShutdown(gracefulShutdown chan os.Signal, waitGroup *sync.WaitGroup, functionsOnShutdown ...func()) {

	signalReceived := <-gracefulShutdown
	log.Info().
		Msgf("Received signal %v. Waiting for running tasks to finish...", signalReceived)

	// execute any passed function
	for _, f := range functionsOnShutdown {
		f()
	}

	waitGroup.Wait()

	log.Info().Msg("Shutting down...")
}

// ApplyJitter adds +-25% jitter to the input
func ApplyJitter(input int) (output int) {

	deviation := int(0.25 * float64(input))

	return input - deviation + r.Intn(2*deviation)
}

// WatchForFileChanges waits for a change to the provided file path and then executes the function
func WatchForFileChanges(filePath string, functionOnChange func(fsnotify.Event)) {
	// copied from https://github.com/spf13/viper/blob/v1.3.1/viper.go#L282-L348
	initWG := sync.WaitGroup{}
	initWG.Add(1)
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal().Err(err).Msg("Creating file system watcher failed")
		}
		defer watcher.Close()

		// we have to watch the entire directory to pick up renames/atomic saves in a cross-platform way
		file := filepath.Clean(filePath)
		fileDir, _ := filepath.Split(file)
		realFile, _ := filepath.EvalSymlinks(filePath)

		eventsWG := sync.WaitGroup{}
		eventsWG.Add(1)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok { // 'Events' channel is closed
						eventsWG.Done()
						return
					}
					currentFile, _ := filepath.EvalSymlinks(filePath)
					// we only care about the key file with the following cases:
					// 1 - if the key file was modified or created
					// 2 - if the real path to the key file changed (eg: k8s ConfigMap/Secret replacement)
					const writeOrCreateMask = fsnotify.Write | fsnotify.Create
					if (filepath.Clean(event.Name) == file &&
						event.Op&writeOrCreateMask != 0) ||
						(currentFile != "" && currentFile != realFile) {
						realFile = currentFile

						functionOnChange(event)
					} else if filepath.Clean(event.Name) == file &&
						event.Op&fsnotify.Remove&fsnotify.Remove != 0 {
						eventsWG.Done()
						return
					}

				case err, ok := <-watcher.Errors:
					if ok { // 'Errors' channel is not closed
						log.Printf("watcher error: %v\n", err)
					}
					eventsWG.Done()
					return
				}
			}
		}()
		watcher.Add(fileDir)
		initWG.Done()   // done initalizing the watch in this go routine, so the parent routine can move on...
		eventsWG.Wait() // now, wait for event loop to end in this go-routine...
	}()
	initWG.Wait() // make sure that the go routine above fully ended before returning
}

// HandleError logs a fatal when the error is not nil
func HandleError(err error) {
	if err != nil {
		log.Fatal().Err(err)
	}
}

// RunCommand runs a full command string and replaces placeholders with the arguments; it logs a fatal on error
// RunCommand("kubectl logs -l app=%v -n %v", app, namespace)
func RunCommand(command string, args ...interface{}) {
	err := RunCommandExtended(command, args...)
	HandleError(err)
}

// RunCommandExtended runs a full command string and replaces placeholders with the arguments; it returns an error if command execution failed
// err := RunCommandExtended("kubectl logs -l app=%v -n %v", app, namespace)
func RunCommandExtended(command string, args ...interface{}) error {
	command = fmt.Sprintf(command, args...)
	log.Debug().Msgf("> %v", command)

	// trim spaces and de-dupe spaces in string
	command = strings.ReplaceAll(command, "  ", " ")
	command = strings.Trim(command, " ")

	// split into actual command and arguments
	commandArray := strings.Split(command, " ")
	var c string
	var a []string
	if len(commandArray) > 0 {
		c = commandArray[0]
	}
	if len(commandArray) > 1 {
		a = commandArray[1:]
	}

	cmd := exec.Command(c, a...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

// RunCommandWithArgs runs a single command and passes the arguments; it logs a fatal on error
// RunCommandWithArgs("kubectl", []string{"logs", "-l", "app="+app, "-n", namespace)
func RunCommandWithArgs(command string, args []string) {
	err := RunCommandWithArgsExtended(command, args)
	HandleError(err)
}

// RunCommandWithArgsExtended runs a single command and passes the arguments; it returns an error if command execution failed
// err := RunCommandWithArgsExtended("kubectl", []string{"logs", "-l", "app="+app, "-n", namespace)
func RunCommandWithArgsExtended(command string, args []string) error {
	log.Debug().Msgf("> %v %v", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
