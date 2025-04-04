package config

// These arguments and semantics are used for all command-line utilities that deal directly with autograder resources
// (which is almost all utilities).
// ConfigArgs should be embedded into whatever Kong argument structure you are using.
// HandleConfigArgs() should be called with the parsed args,
// and this will handle initing the entire configuration infrastructure.
//
// Configurations will be loaded in the following order (later options override earlier ones):
//  0) The command-line options are checked for BASE_DIR.
//  1) Load options from environmental variables.
//  2) Options are loaded from WORK_DIR/config (config.json then secrets.json).
//  3) Options are loaded from the current working directory (config.json then secrets.json).
//  4) Options are loaded from any files specified with --config-path (ordered by appearance).
//  5) Options are loaded from the command-line (--config / -c).

import (
	"fmt"
	"os"

	"github.com/edulinq/autograder/internal/log"
)

// A Kong-style struct for adding on all the config-related options to a CLI.
type ConfigArgs struct {
	ConfigPath  []string          `help:"Path to config file to load." type:"existingfile"`
	Config      map[string]string `help:"Config options." short:"c"`
	LogLevel    string            `help:"Set the text (command line) logging level. Shortcut for '-c log.text.level'."`
	UnitTesting bool              `help:"Enable all options for unit testing and load test data/courses." default:"false"`
}

func HandleConfigArgs(args ConfigArgs) error {
	return HandleConfigArgsFull(args, "", false)
}

func HandleConfigArgsFull(args ConfigArgs, cwd string, skipEnv bool) error {
	defer InitSystemWithConfig()

	if cwd == "" {
		cwd = shouldGetCWD()
	}

	// Step 0: Check the command-line options for BASE_DIR.
	value, ok := args.Config[BASE_DIR.Key]
	if ok {
		BASE_DIR.Set(value)
	}

	// Step 1: Load options from environmental variables.
	if !skipEnv {
		LoadEnv()
	}

	// Step 2: Load options from WORK_DIR.
	dir := GetConfigDir()
	err := LoadConfigFromDir(dir)
	if err != nil {
		return fmt.Errorf("Failed to load config from work dir ('%s'): '%w'.", dir, err)
	}

	// Step 3: Load options from the current working directory.
	dir = cwd
	err = LoadConfigFromDir(dir)
	if err != nil {
		return fmt.Errorf("Failed to load config from work dir ('%s'): '%w'.", dir, err)
	}

	// Step 4: Load files from --config-path.
	for _, path := range args.ConfigPath {
		err := LoadFile(path)
		if err != nil {
			return err
		}
	}

	// Step 5: Load options from the command-line (--config).
	for key, value := range args.Config {
		Set(key, value)
	}

	// Set special options.

	if args.LogLevel != "" {
		LOG_TEXT_LEVEL.Set(args.LogLevel)
	}

	if args.UnitTesting {
		err := EnableUnitTestingModeFull(false)
		if err != nil {
			return err
		}
	}

	return nil
}

func shouldGetCWD() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Error("Failed to get working directory.", err)
		return "."
	}

	return cwd
}
