package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	prompt "github.com/c-bata/go-prompt"
	"github.com/google/uuid"
	"github.com/neblic/platform/cmd/neblictl/internal"
	"github.com/neblic/platform/cmd/neblictl/internal/controlplane"
	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
	promptImpl "github.com/neblic/platform/cmd/neblictl/internal/prompt"
	"github.com/neblic/platform/controlplane/client"
	"github.com/neblic/platform/logging"
	"github.com/pkg/term/termios"
	"golang.org/x/sys/unix"
)

var (
	controlURL string
)

var (
	controlPlaneCommands *controlplane.Commands
	writer               *internal.Writer
	fd                   int
	originalTermios      *unix.Termios = &unix.Termios{}
)

func createTokanizedCommand(inputText string) *interpoler.TokanizedCommand {
	parts := strings.Split(inputText, " ")

	// Remove empty parts result of having multiple contiguous spaces
	removedSpacesParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}

		removedSpacesParts = append(removedSpacesParts, part)
	}

	hasTrailingSpace := false
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		hasTrailingSpace = true
	}

	return interpoler.NewTokanizedCommand(removedSpacesParts, hasTrailingSpace)

}

func executor(in string) {
	// restore the original settings to allow ctrl-c to generate signal
	if err := termios.Tcsetattr(uintptr(fd), termios.TCSANOW, (*unix.Termios)(originalTermios)); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			cancel()
		}
	}()

	// Don't print anything if input is empty
	if in == "" {
		return
	}

	// Sanitize input
	tokanizedCommand := createTokanizedCommand(in)

	// Execute command and show output/error
	err := promptImpl.Execute(ctx, controlPlaneCommands.Commands, tokanizedCommand, writer)
	if err != nil {
		writer.WriteStringf("Error: %v\n", err)
	}

}

func completer(in prompt.Document) []prompt.Suggest {
	inputText := in.TextBeforeCursor()

	// Don't print anything if input is empty
	if inputText == "" {
		return []prompt.Suggest{}
	}

	// Tokanize the input
	tokanizedCommand := createTokanizedCommand(inputText)

	// Execute command and show output/error
	suggestions := promptImpl.Suggestions(context.Background(), controlPlaneCommands.Commands, tokanizedCommand)
	return suggestions
}

func fail(format string, a ...any) {
	println(fmt.Sprintf(format, a...))
	os.Exit(1)
}

func main() {
	// Initialize logger
	logger, err := logging.NewZapDev()
	if err != nil {
		fail(err.Error())
	}

	// Initialize configuration controller
	configController, err := internal.NewConfigurationController()
	if err != nil {
		fail("error initializing configuration controller: %v", err)
	}
	config, err := configController.Configuration()
	if err != nil {
		fail("error reading configuration file: %v", err)
	}

	// Detect if a simple 'init' command has to be executed
	if len(os.Args) == 2 && os.Args[1] == "init" {
		if config.ClientUID == "" {
			config.ClientUID = uuid.NewString()
			err := configController.SetConfiguration(config)
			if err != nil {
				fail("error writing configuration file: %v", err)
			}
		}
		fmt.Println("user configuration file:", configController.ConfigurationPath())
		os.Exit(0)
	}

	// Parse arguments
	host := flag.String("host", "localhost", "OpenTelemetry collector host")
	controlPort := flag.Uint64("control-port", 8899, "OpenTelemetry collector control port")
	tls := flag.Bool("tls", false, "Enable TLS encryption in the server connection")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	controlURL = fmt.Sprintf("%s:%d", *host, *controlPort)
	writer = internal.NewFormattedWriter(os.Stdout)

	// Initialize clientUID if needed
	if config.ClientUID == "" {
		config.ClientUID = uuid.NewString()
		err := configController.SetConfiguration(config)
		if err != nil {
			fail("error writing configuration file: %v", err)
		}
	}

	// Initialize terminal variables
	fd, err = syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	err = termios.Tcgetattr(uintptr(fd), originalTermios)
	if err != nil {
		panic(err)
	}

	// Initialize control plane client
	opts := []client.Option{}
	// Set logger
	var clientLogger logging.Logger
	clientLogger = logging.NewNopLogger()
	if *debug {
		clientLogger = logger
	}
	opts = append(opts, client.WithLogger(clientLogger))
	// Set TLS
	if *tls {
		opts = append(opts, client.WithTLS())
	}
	// Set token
	if config.Token != "" {
		opts = append(opts, client.WithAuthBearer(config.Token))
	}
	controlPlaneClient, err := controlplane.NewClient(config.ClientUID, controlURL, opts...)
	if err != nil {
		fail(err.Error())
	}

	// Initialize executor
	controlPlaneExecutors := controlplane.NewExecutors(controlPlaneClient)

	// Initialize completer
	controlPlaneCompleters := controlplane.NewCompleters(controlPlaneClient)

	// Initialize commands
	controlPlaneCommands = controlplane.NewCommands(controlPlaneExecutors, controlPlaneCompleters)

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("live-prefix-example"),
		prompt.OptionSwitchKeyBindMode(prompt.CommonKeyBind),
	)
	p.Run()
}
