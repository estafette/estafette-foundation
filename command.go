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
// RunCommand(ctx, "kubectl logs -l app=%v -n %v", app, namespace)
func RunCommand(ctx context.Context, command string, args ...interface{}) {
	c, a := getSeparateCommandAndArgs(ctx, command, args)
	RunCommandWithArgs(ctx, c, a)
}

// RunCommandExtended runs a full command string and replaces placeholders with the arguments; it returns an error if command execution failed
// err := RunCommandExtended(ctx, "kubectl logs -l app=%v -n %v", app, namespace)
func RunCommandExtended(ctx context.Context, command string, args ...interface{}) error {
	c, a := getSeparateCommandAndArgs(ctx, command, args)
	return RunCommandWithArgsExtended(ctx, c, a)
}

// GetCommandOutput runs a full command string and replaces placeholders with the arguments; it returns the output as a string and an error if command execution failed
// output, err := GetCommandOutput(ctx, "kubectl logs -l app=%v -n %v", app, namespace)
func GetCommandOutput(ctx context.Context, command string, args ...interface{}) (string, error) {
	c, a := getSeparateCommandAndArgs(ctx, command, args)
	return GetCommandWithArgsOutput(ctx, c, a)
}

// RunCommandWithArgs runs a single command and passes the arguments; it logs a fatal on error
// RunCommandWithArgs(ctx, "kubectl", []string{"logs", "-l", "app="+app, "-n", namespace)
func RunCommandWithArgs(ctx context.Context, command string, args []string) {
	err := RunCommandWithArgsExtended(ctx, command, args)
	HandleError(err)
}

// RunCommandWithArgsExtended runs a single command and passes the arguments; it returns an error if command execution failed
// err := RunCommandWithArgsExtended(ctx, "kubectl", []string{"logs", "-l", "app="+app, "-n", namespace)
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
// output, err := GetCommandWithArgsOutput(ctx, "kubectl", []string{"logs", "-l", "app="+app, "-n", namespace)
func GetCommandWithArgsOutput(ctx context.Context, command string, args []string) (string, error) {
	log.Debug().Msg(aurora.Sprintf(aurora.Gray(18, "> %v %v"), command, strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()

	return string(output), err
}

// RunCommandInDirectory runs a full command string and replaces placeholders with the arguments from the specified directory; it logs a fatal on error
// RunCommandInDirectory(ctx, "directory other than working dir", "kubectl logs -l app=%v -n %v", app, namespace)
func RunCommandInDirectory(ctx context.Context, dir string, command string, args ...interface{}) {
	c, a := getSeparateCommandAndArgs(ctx, command, args)
	RunCommandInDirectoryWithArgs(ctx, dir, c, a)
}

// RunCommandInDirectoryExtended runs a full command string and replaces placeholders with the arguments from the specified directory; it returns an error if command execution failed
// err := RunCommandInDirectoryExtended(ctx, "directory other than working dir", "kubectl logs -l app=%v -n %v", app, namespace)
func RunCommandInDirectoryExtended(ctx context.Context, dir string, command string, args ...interface{}) error {
	c, a := getSeparateCommandAndArgs(ctx, command, args)
	return RunCommandInDirectoryWithArgsExtended(ctx, dir, c, a)
}

// RunCommandInDirectoryWithArgs runs a single command and passes the arguments from the specified directory; it logs a fatal on error
// RunCommandInDirectoryWithArgs(ctx, "directory other than working dir", "kubectl", []string{"logs", "-l", "app="+app, "-n", namespace)
func RunCommandInDirectoryWithArgs(ctx context.Context, dir string, command string, args []string) {
	err := RunCommandInDirectoryWithArgsExtended(ctx, dir, command, args)
	HandleError(err)
}

// RunCommandInDirectoryWithArgsExtended runs a single command and passes the arguments from the specified directory; it returns an error if command execution failed
// err := RunCommandInDirectoryWithArgsExtended(ctx, "directory other than working dir", "kubectl", []string{"logs", "-l", "app="+app, "-n", namespace)
func RunCommandInDirectoryWithArgsExtended(ctx context.Context, dir string, command string, args []string) error {
	log.Debug().Msg(aurora.Sprintf(aurora.Gray(18, "[%v] > %v %v"), dir, command, strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir

	err := cmd.Run()

	return err
}

// GetCommandInDirectoryOutput runs a full command string and replaces placeholders with the arguments from the specified directory; it returns the output as a string and an error if command execution failed
// output, err := GetCommandInDirectoryOutput(ctx, "directory other than working dir", "kubectl logs -l app=%v -n %v", app, namespace)
func GetCommandInDirectoryOutput(ctx context.Context, dir string, command string, args ...interface{}) (string, error) {
	c, a := getSeparateCommandAndArgs(ctx, command, args)
	return GetCommandWithArgsInDirectoryOutput(ctx, dir, c, a)
}

// GetCommandWithArgsInDirectoryOutput runs a single command and passes the arguments from the specified directory; it returns the output as a string and an error if command execution failed
// output, err := GetCommandWithArgsOutput(ctx, "directory other than working dir", "kubectl", []string{"logs", "-l", "app="+app, "-n", namespace)
func GetCommandWithArgsInDirectoryOutput(ctx context.Context, dir string, command string, args []string) (string, error) {
	log.Debug().Msg(aurora.Sprintf(aurora.Gray(18, "[%v] > %v %v"), dir, command, strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = os.Environ()
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()

	return string(output), err
}

func getSeparateCommandAndArgs(ctx context.Context, command string, args []interface{}) (c string, a []string) {
	command = fmt.Sprintf(command, args...)

	// trim spaces and de-dupe spaces in string
	command = strings.ReplaceAll(command, "  ", " ")
	command = strings.Trim(command, " ")

	// split into actual command and arguments
	commandArray := strings.Split(command, " ")

	// remove empty items
	var cleanedCommandArray []string
	for _, str := range commandArray {
		if str != "" {
			cleanedCommandArray = append(cleanedCommandArray, str)
		}
	}

	if len(cleanedCommandArray) > 0 {
		c = cleanedCommandArray[0]
	}
	if len(cleanedCommandArray) > 1 {
		a = cleanedCommandArray[1:]
	}

	return
}
