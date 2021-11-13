# Cobra Flag Prompt
Cobra Flag Prompt prompts users to enter values for required flags. It is an extension of [Cobra](https://github.com/spf13/cobra), and requires that you use Cobra to set up the CLI for your program.

[GoDoc](https://pkg.go.dev/github.com/Polyapp-LLC/cobraFlagPrompt)

## User Experience Before Cobra Flag Prompt
Without Cobra Flag Prompt, the program immediately terminates if the user did not enter a required flag and prints out which flag(s) were missing and all usage text for all flags. The user must then retype the command, add the new flag(s), identify the flag(s) usage text, and choose the correct value for the required flag. This isn't a big deal with programs which only have two options, but for programs with dozens of options the process becomes tedious.

Video: https://youtu.be/EhVW5Vl9KAE

## User Experience After Cobra Flag Prompt
With Cobra Flag Prompt, the user is prompted for any missing required flags. To assist the user, the Usage information for the missing flag(s) is displayed when the user needs it.

Video: https://youtu.be/5sNhdYA5hc4

## Usage
Usage is trivial. Replace calls to `MarkFlagRequired` with `cobraFlagPrompt.MarkFlagRequired` and calls to `MarkPersistentFlagRequired` with `cobraFlagPrompt.MarkPersistentFlagRequired`.

Advanced usage can be ascertained by reading the [GoDoc](https://pkg.go.dev/github.com/Polyapp-LLC/cobraFlagPrompt) or the source code.

## Usage Walkthrough
First, get the package

`go get github.com/Polyapp-LLC/cobraFlagPrompt`

Then modify your existing code. Here is an example Cobra init() function for a program with 2 flags and which does **NOT** use cobraFlagPrompt:
```
func init() {
    rootCmd.Flags().StringVar(&EnvironmentName, "environmentName", "", `environment names can only be lowercase ascii characters and hyphens. Example: test-one`)
    rootCmd.PersistentFlags().StringVarP(&CloudName, "cloudName", "n", "", `cloud name must be one of the following strings: "AWS", "Azure", "Google", "localhost"`)
    
    err := rootCmd.MarkFlagRequired("environmentName")
    if err != nil {
        panic("rootCmd.MarkFlagRequired(environmentName): " + err.Error())
    }
    err = rootCmd.MarkPersistentFlagRequired("cloudName")
    if err != nil {
    	panic("rootCmd.MarkPersistentFlagRequired(cloudName): " + err.Error())
    }
}
```
Here is the same code, except it is using cobraFlagPrompt. Notice: `rootCmd.MarkFlagRequired` has been updated to: `cobraFlagPrompt.MarkFlagRequired`.
```
func init() {
    rootCmd.Flags().StringVar(&EnvironmentName, "environmentName", "", `environment names can only be lowercase ascii characters and hyphens. Example: test-one`)
    rootCmd.PersistentFlags().StringVarP(&CloudName, "cloudName", "n", "", `cloud name must be one of the following strings: "AWS", "Azure", "Google", "localhost"`)
    
    err := cobraFlagPrompt.MarkFlagRequired(rootCmd, "environmentName")
    if err != nil {
        panic("cobraFlagPrompt.MarkFlagRequired(environmentName): " + err.Error())
    }
    err = cobraFlagPrompt.MarkPersistentFlagRequired(rootCmd, "cloudName")
    if err != nil {
    	panic("cobraFlagPrompt.MarkPersistentFlagRequired(cloudName): " + err.Error())
    }
}
```
The results can be seen above (this is the code used in the example videos).

## Support and Issues
Cobra Flag Prompt ought to work with all flag types supported by Cobra. Tests are in [cobraFlagPrompt_test.go](./cobraFlagPrompt_test.go). Questions? Leave an Issue! Thanks :)
