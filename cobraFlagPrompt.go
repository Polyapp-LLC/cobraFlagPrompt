package cobraFlagPrompt

import (
	"fmt"
	"github.com/spf13/cobra"
	"sync"
)

type requiredFlag struct {
	name string
	// cmdToWhichFlagIsAttached is either the relevant command OR it is nil, which implies this is a persistent required flag.
	cmdToWhichFlagIsAttached *cobra.Command
}

var (
	// flagsRequired is a list of flag names which are required. They can include both persistent and non-persistent flags.
	flagsRequired           = make([]requiredFlag, 0)
	persistentFlagsRequired = make([]requiredFlag, 0)
)

// MarkFlagRequired causes the command to prompt the user if this flag is not provided as a command line argument.
func MarkFlagRequired(cmd *cobra.Command, name string) error {
	f := cmd.Flag(name)
	if f == nil {
		return fmt.Errorf("no such flag -%v", name)
	}

	cmd.PreRunE = preRun(cmd.PreRunE, cmd.PreRun)
	// cmd.PreRun = nil // not necessary

	// TODO verify that this flag is a regular flag and not a persistent flag
	flagsRequired = append(flagsRequired, requiredFlag{name: name, cmdToWhichFlagIsAttached: cmd})
	return nil
}

func MarkPersistentFlagRequired(cmd *cobra.Command, name string) error {
	f := cmd.Flag(name)
	if f == nil {
		return fmt.Errorf("no such flag -%v", name)
	}

	cmd.PreRunE = preRun(cmd.PreRunE, cmd.PreRun)
	// cmd.PreRun = nil

	// TODO verify that this flag is a persistent flag and not a regular flag
	flagsRequired = append(persistentFlagsRequired, requiredFlag{name: name})
	return nil
}

var (
	// I doubt this is necessary but I'm going to add it just in case someone is doing some weird stuff with cobra.
	preRunMux sync.Mutex

	hasBeenCalled = make(map[*cobra.Command]bool)
)

// preRun ensures the existing PreRunE or PreRun command which could be defined by the developer consuming this library is executed.
//
// Why is this set up like this? In Cobra, PreRunE and PreRun execute just prior to `func validateRequiredFlags` which is where
// missing flags can be identified and the user can be notified of their error.
// To ensure cobraFlagPrompt closely mimics Cobra's behavior, we need to execute our code at the same time validateRequiredFlags
// would run. Therefore, we want to run our code just after PreRun or PreRunE execute.
// The easiest way to do this is to attach our own PreRunE. A simple implementation would involve either
// asking the developer to call our code within their own PreRunE code or overwriting PreRunE, but then the
// developer could forget to call our code, the developer could call it in the wrong order, or our code could
// overwrite the developer's code.
//
// The solution to call the existing PreRunE has the drawback that if the developer defines their own PreRunE after
// our PreRunE is defined our code will not run. However, the documentation for Cobra suggests defining PreRunE in
// the cobra.Command{} struct definition and defining flags after that, so I am hopeful developers will naturally
// not encounter this problem. In the event that they DO encounter this problem, they can call CobraFlagPromptPreRunE directly.
//
// Testing note: I do not intend to write any tests for this. All testing was done by manually defining PreRun and PreRunE
// which printed out "PreRunE" on github.com/Polyapp-LLC/gendeploy and making sure the text was printed to the cmd line.
func preRun(existingPreRunE func(cmd *cobra.Command, args []string) error, existingPreRun func(cmd *cobra.Command, args []string)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error
		if existingPreRunE != nil {
			err = existingPreRunE(cmd, args)
			if err != nil {
				return fmt.Errorf("%w", err)
			}
		} else if existingPreRun != nil {
			existingPreRun(cmd, args)
		}

		// preRun will be called several times (once for each required flag). But CobraFlagPromptPreRunE only wants
		// to be executed once. Indeed, if the developer decides to call CobraFlagPromptPreRunE directly, then
		// it only WILL be executed once. To avoid having the developer's way of calling the code != the
		// cobraFlagPrompt way of calling this code, let's only allow this code to be called once per cmd.
		preRunMux.Lock()
		defer preRunMux.Unlock()
		if (hasBeenCalled[cmd]) {
			return nil
		}

		err = CobraFlagPromptPreRunE(cmd, args)
		if err != nil {
			return fmt.Errorf("cobraFlagPromptPreRunE: %w", err)
		}
		hasBeenCalled[cmd] = true
		return nil
	}
}

// CobraFlagPromptPreRunE runs our PreRunE command, which will prompt the user to enter information for missing flags.
// This function can be called multiple times, but it will only run once.
//
// Developer note: This code is automatically attached to PreRunE when you call MarkFlagRequired or MarkPersistentFlagRequired
// BUT if you set PreRunE *after* calling MarkFlagRequired or MarkPersistentFlagRequired, that will overwrite this PreRunE.
// In that scenario, you will need to manually call CobraFlagPromptPreRunE *at the very end of your PreRunE*.
// Because Cobra prefers PreRunE over PreRun (it's an if / else if), if you set your own PreRun after calling
// MarkFlagRequired or MarkPersistentFlagRequired then that PreRun will be ignored (unless you also cleared out PreRunE).
func CobraFlagPromptPreRunE(cmd *cobra.Command, args []string) error {
	fmt.Println("CobraFlagPromptPreRunE")
	return nil
}
