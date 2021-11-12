package cobraFlagPrompt

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
	"testing"
)

// TestNoCobraFlagPrompt is an example of when you import cobraFlagPrompt but don't use it
func TestNoCobraFlagPrompt(t *testing.T) {
	singleStringValueCmd := &cobra.Command{
		Use:   "use",
		Short: "short",
		Long:  `long`,
	}
	singleStringValueCmd.Flags().String("singleStringValueCmd", "", "usage")

	// we never set up any required flags so we should not do anything to cobra.Command
	var testOut bytes.Buffer
	var testIn bytes.Buffer
	err := CobraFlagPromptPreRunE(singleStringValueCmd, []string{}, &testIn, &testOut)
	if err != nil {
		t.Errorf("CobraFlagPromptPreRunE: %v", err.Error())
	}
	if testOut.Len() > 0 {
		t.Error("testOut.Len() > 0. Value: " + testOut.String())
	}
}

func TestMixedFlags(t *testing.T) {
	var err error
	twoStringValueCmd := &cobra.Command{
		Use:   "use",
		Short: "short",
		Long:  `long`,
	}
	twoStringValueCmd.Flags().String("singleStringValueCmd", "", "usage")
	secondFlagPtr := twoStringValueCmd.Flags().String("second", "", "usage")
	err = MarkFlagRequired(twoStringValueCmd, "second")
	if err != nil {
		t.Fatalf("MarkFlagRequired: %v", err.Error())
	}

	var testOut bytes.Buffer
	var testIn bytes.Buffer
	testIn.WriteString("test value\nsecond\nthird\n\n")
	err = CobraFlagPromptPreRunE(twoStringValueCmd, []string{}, &testIn, &testOut)
	if err != nil {
		t.Errorf("CobraFlagPromptPreRunE: %v", err.Error())
	}
	if !strings.Contains(testOut.String(), fmt.Sprintf("Flag --%v is required. Please enter a value for this flag.", "second")) {
		t.Error("testOut should have prompted for the flag named second")
	}
	if strings.Contains(testOut.String(), fmt.Sprintf("singleStringValueCmd")) {
		t.Errorf("testOut should NOT have prompted for the flag named singleStringValueCmd")
	}
	flagOut := twoStringValueCmd.Flag("second").Value
	if flagOut.String() != "test value" {
		t.Errorf("flag's new value should have been recorded as the first newline delimited value available at stdIn. Instead it was: " + flagOut.String())
	}
	if *secondFlagPtr != "test value" {
		t.Errorf("flag's new value should hvae been recorded as the first newline deliminated value available at stdIn. Instead it was: " + *secondFlagPtr)
	}
}

func TestBoolPointerFlags(t *testing.T) {
	var err error
	twoStringValueCmd := &cobra.Command{
		Use:   "use",
		Short: "short",
		Long:  `long`,
	}
	var boolPtr bool
	twoStringValueCmd.Flags().BoolVar(&boolPtr, "boolPtr", false, "usage")
	err = MarkFlagRequired(twoStringValueCmd, "boolPtr")
	if err != nil {
		t.Fatalf("MarkFlagRequired: %v", err.Error())
	}

	var testOut bytes.Buffer
	var testIn bytes.Buffer
	testIn.WriteString("true\nstring value\nthird\n\n")
	err = CobraFlagPromptPreRunE(twoStringValueCmd, []string{}, &testIn, &testOut)
	if err != nil {
		t.Errorf("CobraFlagPromptPreRunE: %v", err.Error())
	}
	if !strings.Contains(testOut.String(), fmt.Sprintf("Flag --%v is required. Please enter a value for this flag.", "boolPtr")) {
		t.Error("testOut should have prompted for the flag named boolPtr")
	}
	if !strings.Contains(testOut.String(), fmt.Sprintf("boolPtr")) {
		t.Errorf("testOut should have prompted for the flag named boolPtr")
	}
	if boolPtr != true {
		t.Errorf("bool's value should be true as set by stdin. Instead it was: " + strconv.FormatBool(boolPtr))
	}
}

func TestStringPointerFlags(t *testing.T) {
	var err error
	twoStringValueCmd := &cobra.Command{
		Use:   "use",
		Short: "short",
		Long:  `long`,
	}
	var stringPtr string
	twoStringValueCmd.Flags().StringVar(&stringPtr, "stringPtr", "default value", "usage")
	err = MarkFlagRequired(twoStringValueCmd, "stringPtr")
	if err != nil {
		t.Fatalf("MarkFlagRequired: %v", err.Error())
	}

	var testOut bytes.Buffer
	var testIn bytes.Buffer
	testIn.WriteString("string value\nthird\n\n")
	err = CobraFlagPromptPreRunE(twoStringValueCmd, []string{}, &testIn, &testOut)
	if err != nil {
		t.Errorf("CobraFlagPromptPreRunE: %v", err.Error())
	}
	if !strings.Contains(testOut.String(), fmt.Sprintf("Flag --%v is required. Please enter a value for this flag.", "stringPtr")) {
		t.Error("testOut should have prompted for the flag named stringPtr")
	}
	if !strings.Contains(testOut.String(), fmt.Sprintf("stringPtr")) {
		t.Errorf("testOut should have prompted for the flag named stringPtr")
	}
	if stringPtr != "string value" {
		t.Errorf("string's value should be true as set by stdin. Instead it was: " + stringPtr)
	}
}

func TestListStringPointerFlags(t *testing.T) {
	var err error
	twoStringValueCmd := &cobra.Command{
		Use:   "use",
		Short: "short",
		Long:  `long`,
	}
	var stringPtr []string
	twoStringValueCmd.Flags().StringSliceVar(&stringPtr, "stringPtr", []string{"default value"}, "usage")
	err = MarkFlagRequired(twoStringValueCmd, "stringPtr")
	if err != nil {
		t.Fatalf("MarkFlagRequired: %v", err.Error())
	}

	var testOut bytes.Buffer
	var testIn bytes.Buffer
	testIn.WriteString("string value\nsecond\n")
	err = CobraFlagPromptPreRunE(twoStringValueCmd, []string{}, &testIn, &testOut)
	if err != nil {
		t.Errorf("CobraFlagPromptPreRunE: %v", err.Error())
	}
	if !strings.Contains(testOut.String(), fmt.Sprintf("Flag --%v is required. Please enter a value for this flag.", "stringPtr")) {
		t.Error("testOut should have prompted for the flag named stringPtr")
	}
	if !strings.Contains(testOut.String(), fmt.Sprintf("stringPtr")) {
		t.Errorf("testOut should have prompted for the flag named stringPtr")
	}
	if len(stringPtr) != 2 {
		t.Fatal("len(stringPtr) should have been 2 since we put in 2 inputs")
	}
	if stringPtr[0] != "string value" {
		t.Errorf("string's value should be true as set by stdin. Instead it was: " + stringPtr[0])
	}
	if stringPtr[1] != "second" {
		t.Errorf("strings' second value should have been second. Instead it was: " + stringPtr[1])
	}
}
