package foundation

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	goVersion = runtime.Version()

	// seed random number
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// InitLoggingFromEnv initalializes a logger with format specified in envvar ESTAFETTE_LOG_FORMAT and outputs a startup message
func InitLoggingFromEnv(appgroup, app, version, branch, revision, buildDate string) {
	InitLoggingByFormat(appgroup, app, version, branch, revision, buildDate, os.Getenv("ESTAFETTE_LOG_FORMAT"))
}

// InitLoggingByFormat initalializes a logger with specified format and outputs a startup message
func InitLoggingByFormat(appgroup, app, version, branch, revision, buildDate string, logFormat string) {

	// configure logger and output startup message
	switch logFormat {
	case LogFormatJSON:
		initLoggingJSON(appgroup, app, version, branch, revision, buildDate)
		logStartupMessage(appgroup, app, version, branch, revision, buildDate)
	case LogFormatStackdriver:
		initLoggingStackdriver(appgroup, app, version, branch, revision, buildDate)
		logStartupMessage(appgroup, app, version, branch, revision, buildDate)
	case LogFormatV3:
		initLoggingV3(appgroup, app, version, branch, revision, buildDate)
		logStartupMessageV3(appgroup, app, version, branch, revision, buildDate)
	case LogFormatConsole:
		initLoggingConsole(appgroup, app, version, branch, revision, buildDate)
		logStartupMessageConsole(appgroup, app, version, branch, revision, buildDate)
	default: // LogFormatPlainText
		initLoggingPlainText(appgroup, app, version, branch, revision, buildDate)
		logStartupMessage(appgroup, app, version, branch, revision, buildDate)
	}
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
