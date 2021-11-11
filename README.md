# Cobra Flag Prompt
Cobra Flag Prompt [GoDoc](TODO) prompts users to enter values for required flags. It is an extension of [Cobra](https://github.com/spf13/cobra), and requires that you use Cobra to set up the CLI for your program.

## User Experience Before Cobra Flag Prompt

## User Experience After Cobra Flag Prompt

## Usage
First, get the package

`go get github.com/Polyapp-LLC/cobraFlagPrompt`

Then modify your existing Cobra code. Example Cobra init() function for a program with 2 flags and which does **NOT** use cobraFlagPrompt:
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
Here is the same code, except `rootCmd.MarkFlagRequired` -> `cobraFlagPrompt.MarkFlagRequired`.
```
func init() {
    rootCmd.Flags().StringVar(&EnvironmentName, "environmentName", "", `environment names can only be lowercase ascii characters and hyphens. Example: test-one`)
    rootCmd.PersistentFlags().StringVarP(&CloudName, "cloudName", "n", "", `cloud name must be one of the following strings: "AWS", "Azure", "Google", "localhost"`)
    
    // Warning: You MUST define your rootCmd.PreRunE or rootCmd.PreRun PRIOR to this next line of code.
    // cobraFlagPrompt only works because it attaches some code in PreRunE.
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
This code supports all native types in Go. Tests are in [cobraFlagPrompt_test.go](./cobraFlagPrompt_test.go). Questions? Leave an Issue! Thanks :)
