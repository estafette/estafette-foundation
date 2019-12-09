package foundation

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/rs/zerolog/log"
)

// HandleError logs a fatal when the error is not nil
func HandleError(err error) {
	if err != nil {
		log.Fatal().Err(err).Msg("Fatal error")
	}
}

// RunCommand runs a full command string and replaces placeholders with the arguments; it logs a fatal on error
// RunCommand("kubectl logs -l app=%v -n %v", app, namespace)
func RunCommand(ctx context.Context, command string, args ...interface{}) {
	err := RunCommandExtended(ctx, command, args...)
	HandleError(err)
}

// RunCommandExtended runs a full command string and replaces placeholders with the arguments; it returns an error if command execution failed
// err := RunCommandExtended("kubectl logs -l app=%v -n %v", app, namespace)
func RunCommandExtended(ctx context.Context, command string, args ...interface{}) error {
	command = fmt.Sprintf(command, args...)
	log.Debug().Msg(aurora.Sprintf(aurora.Gray(18, "> %v"), command))

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

	cmd := exec.CommandContext(ctx, c, a...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

// RunCommandWithArgs runs a single command and passes the arguments; it logs a fatal on error
// RunCommandWithArgs("kubectl", []string{"logs", "-l", "app="+app, "-n", namespace)
func RunCommandWithArgs(ctx context.Context, command string, args []string) {
	err := RunCommandWithArgsExtended(ctx, command, args)
	HandleError(err)
}

// RunCommandWithArgsExtended runs a single command and passes the arguments; it returns an error if command execution failed
// err := RunCommandWithArgsExtended("kubectl", []string{"logs", "-l", "app="+app, "-n", namespace)
func RunCommandWithArgsExtended(ctx context.Context, command string, args []string) error {
	log.Debug().Msg(aurora.Sprintf(aurora.Gray(18, "> %v %v"), command, strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

// GetCommandWithArgsOutput runs a single command and passes the arguments; it returns the output as a string and an error if command execution failed
// output, err := GetCommandWithArgsOutput("kubectl", []string{"logs", "-l", "app="+app, "-n", namespace)
func GetCommandWithArgsOutput(ctx context.Context, command string, args []string) (string, error) {
	log.Debug().Msg(aurora.Sprintf(aurora.Gray(18, "> %v %v"), command, strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()

	return string(output), err
}
