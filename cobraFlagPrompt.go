package cobraFlagPrompt

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io"
	"os"
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
	orderedAllFlags = make([]requiredFlag, 0)
)

// MarkFlagRequired causes the command to prompt the user if this flag is not provided as a command line argument.
func MarkFlagRequired(cmd *cobra.Command, name string) error {
	f := cmd.Flag(name)
	if f == nil {
		return fmt.Errorf("no such flag -%v", name)
	}

	cmd.PreRunE = preRun(cmd.PreRunE, cmd.PreRun)
	// cmd.PreRun = nil // not necessary

	// TODO verify that this flag is a regular flag and not a persistent flag???
	flagsRequired = append(flagsRequired, requiredFlag{name: name, cmdToWhichFlagIsAttached: cmd})
	orderedAllFlags = append(orderedAllFlags, requiredFlag{name: name, cmdToWhichFlagIsAttached: cmd})
	return nil
}

func MarkPersistentFlagRequired(cmd *cobra.Command, name string) error {
	f := cmd.Flag(name)
	if f == nil {
		return fmt.Errorf("no such flag -%v", name)
	}

	cmd.PreRunE = preRun(cmd.PreRunE, cmd.PreRun)
	// cmd.PreRun = nil

	// TODO verify that this flag is a persistent flag and not a regular flag???
	persistentFlagsRequired = append(persistentFlagsRequired, requiredFlag{name: name})
	orderedAllFlags = append(orderedAllFlags, requiredFlag{name: name})
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

		// preRun will be called several times from MarkFlagRequired and/or MarkPersistentFlagRequired
		// (once for each required flag). CobraFlagPromptPreRunE *could* be written so that it could be run
		// multiple times in sequence or in parallel for the same Cmd. However, if the developer decides to
		// call CobraFlagPromptPreRunE directly, then it will only be executed once. To avoid having the developer's
		// way of calling CobraFlagPromptPreRunE != the cobraFlagPrompt way, we must ensure preRun also only
		// calls CobraFlagPromptPreRunE once.
		preRunMux.Lock()
		defer preRunMux.Unlock()
		if (hasBeenCalled[cmd]) {
			return nil
		}

		err = CobraFlagPromptPreRunE(cmd, args, os.Stdin, os.Stdout)
		if err != nil {
			return fmt.Errorf("cobraFlagPromptPreRunE: %w", err)
		}
		hasBeenCalled[cmd] = true
		return nil
	}
}

// CobraFlagPromptPreRunE is the cobraFlagPrompt PreRunE code which will prompt the user to enter information for missing flags.
// cobraFlagPrompt makes an effort to call this function for you. Most developers will not need to call it manually.
// However, if you do find yourself needing to trigger this function manually, make sure you call it exactly once per
// cmd *cobra.Command.
//
// Inputs: cmd where we can find the flags; args from the program; os.StdOut (pass this in to help with testing)
//
// If there are no cobraFlagPrompt required flags registered at the cmd, this func does nothing.
//
// Developer note: This code is automatically attached to PreRunE when you call MarkFlagRequired or MarkPersistentFlagRequired
// BUT if you set PreRunE *after* calling MarkFlagRequired or MarkPersistentFlagRequired, that will overwrite this PreRunE.
// In that scenario, you will need to manually call CobraFlagPromptPreRunE *at the very end of your PreRunE*.
// Because Cobra prefers PreRunE over PreRun (it's an if / else if), if you set your own PreRun after calling
// MarkFlagRequired or MarkPersistentFlagRequired then that PreRun will be ignored (unless you also cleared out PreRunE).
func CobraFlagPromptPreRunE(cmd *cobra.Command, args []string, stdIn io.Reader, stdOut io.Writer) error {
	if cmd == nil {
		return errors.New("CobraFlagPromptPreRunE saw cmd == nil")
	}
	var err error
	flags := cmd.Flags()

	// prompts should occur in the order the flags were added, not in alphabetical order!
	for _, flagName := range orderedAllFlags {
		flag := flags.Lookup(flagName.name)
		if flag == nil {
			continue
		}

		// If the developer specifies a field is required yet the required field has a default value, I'm left
		// with the question: What to do? It does not make sense to have a required field which is populated automatically.
		// In that case there is no point in specifying it as a required field since it will always have a value.
		//
		// The only logical reason I can come up with for doing this is to provide a "suggested value" to the user.
		//
		// For "suggested value" cases, a developer may want the user to verify that a "suggested value" is
		// what they intended while still providing them with a reasonable default.
		if !flag.Changed || flag.NoOptDefVal == flag.Value.String() {
			// user did not set the value -> we want to capture it from them.
			// value was not set to default -> we want to capture it from them.
			// second part of the 'if' is the "suggested value" case.
			err = PromptForFlag(flag, stdIn, stdOut)
			if err != nil {
				return fmt.Errorf("PromptForFlag flag name (%v): %w", flag.Name, err)
			}
		}

		if flag.Value == nil {
			// I'm not sure if this is possible, but if it is, we want to prompt for the value
			err = PromptForFlag(flag, stdIn, stdOut)
			if err != nil {
				return fmt.Errorf("PromptForFlag flag name (%v): %w", flag.Name, err)
			}
		}
	}

	return nil
}

// PromptForFlag prompts the user to enter a value for pflag. It can handle any type of flag supported by cobra.
//
// The value received from the user is stored in the same way cobra stores it and is retrievable in the same way.
func PromptForFlag(flag *pflag.Flag, stdIn io.Reader, stdOut io.Writer) error {
	var err error
	err = infoPrompt(stdOut, flag)
	if err != nil {
		return fmt.Errorf("infoPrompt: %w", err)
	}

	// There are a LOT of implemented flag types from pflag's flag.go
	// We will only want to test a subset of them, but this might work for others.

	// Slices should implement the SliceValue interface.
	sliceValue, ok := flag.Value.(pflag.SliceValue)
	tries := 0
	maxTries := 5
	if ok {
		// case where we need to take a list of inputs
		// reset the slice to remove any defaults
		err = sliceValue.Replace([]string{})
		if err != nil {
			return fmt.Errorf("sliceValue.Replace to reset the flag: %w", err)
		}

		_, err = fmt.Fprintf(stdOut, "This flag is a list. Each line you type will be one element in the list. To terminate the list, press Enter.\n")
		if err != nil {
			return fmt.Errorf("fmt.Fprintf at 'this flag is a list': %w", err)
		}
		receivedStringsCount := 0
		scanner := bufio.NewScanner(stdIn)
		for {
			if tries > maxTries {
				return errors.New("max tries exceeded")
			}
			var stringReceiver string
			if scanner.Scan() {
				stringReceiver = scanner.Text()
			}
			err = scanner.Err()
			if err == nil && stringReceiver == "" {
				if receivedStringsCount > 0 {
					// OK pressed enter
					break
				} else {
					// The user entered nothing for this list. This is not allowed since it is a required flag.
					_, err = fmt.Fprintf(stdOut, "You must enter at least one value in this list because this flag is required.\n")
					if err != nil {
						return fmt.Errorf("fmt.Fprintf at 'list flag is required': %w", err)
					}
					err = infoPrompt(stdOut, flag)
					if err != nil {
						return fmt.Errorf("infoPrompt: %w", err)
					}
					_, err = fmt.Fprintf(stdOut, "This flag is a list. Each line you type will be one element in the list. To terminate the list, press Enter.\n")
					if err != nil {
						return fmt.Errorf("fmt.Fprintf at the second 'this flag is a list' notification: %w", err)
					}
					tries++
					continue
				}
			}
			if err != nil {
				return fmt.Errorf("fmt.Fscanln: %w", err)
			}
			err = sliceValue.Append(stringReceiver)
			if err != nil {
				_, err = fmt.Fprintf(stdOut, "error processing input with pflag.SliceValue.Append(): %v\n", err.Error())
				if err != nil {
					return fmt.Errorf("fmt.Fprintf: %w", err)
				}
				// we want to give the user another chance to input a value
				continue
			}

			receivedStringsCount++
		}
	} else {
		var stringReceiver string
		scanner := bufio.NewScanner(stdIn)
		for {
			if tries > maxTries {
				return errors.New("max tries exceeded")
			}
			if scanner.Scan() {
				stringReceiver = scanner.Text()
			}
			err = scanner.Err()
			if err == nil && stringReceiver == "" {
				err = infoPrompt(stdOut, flag)
				if err != nil {
					return fmt.Errorf("infoPrompt: %w", err)
				}
				// force the user to enter a value
				continue
			}
			if err != nil {
				return fmt.Errorf("fmt.Fscanln: %w", err)
			}
			err = flag.Value.Set(stringReceiver)
			if err != nil {
				_, err = fmt.Fprintf(stdOut, "error processing input with pflag.Value.Set(): %v\n", err.Error())
				if err != nil {
					return fmt.Errorf("fmt.Fprintf: %w", err)
				}
				tries++
				// we want to give the user another chance to input a value
				continue
			}
			break
		}
	}

	return nil
}

func infoPrompt(stdOut io.Writer, pflag *pflag.Flag) error {
	var err error
	_, err = fmt.Fprintf(stdOut, "Flag --%v is required. Please enter a value for this flag.\n", pflag.Name)
	if err != nil {
		return fmt.Errorf("fmt.Fprintf: %w", err)
	}
	_, err = fmt.Fprintf(stdOut, "Usage: %v\n", pflag.Usage)
	if err != nil {
		return fmt.Errorf("fmt.Fprintf: %w", err)
	}
	return nil
}
