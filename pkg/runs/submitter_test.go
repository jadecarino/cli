/*
 * Copyright contributors to the Galasa project
 */
package runs

import (
	"testing"

	"github.com/galasa.dev/cli/pkg/launcher"
	"github.com/galasa.dev/cli/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCanWriteAndReadBackThrottleFile(t *testing.T) {

	mockFileSystem := utils.NewMockFileSystem()
	err := writeThrottleFile(mockFileSystem, "throttle", 101)
	if err != nil {
		assert.Fail(t, "Should not have failed to write a throttle file. "+err.Error())
	}

	isThrottleFileExists, err := mockFileSystem.Exists("throttle")
	if err != nil {
		assert.Fail(t, "Should not have failed to check for the existence of a throttle file. "+err.Error())
	}

	assert.True(t, isThrottleFileExists, "throttle file does not exist!")

	var readBackThrottle int
	readBackThrottle, err = readThrottleFile(mockFileSystem, "throttle")
	if err != nil {
		assert.Fail(t, "Should not have failed to read from a throttle file. "+err.Error())
	}
	assert.Equal(t, readBackThrottle, 101, "read back the wrong throttle value")
}

func TestReadBackThrottleFileFailsIfNoThrottleFileThere(t *testing.T) {

	var err error
	mockFileSystem := utils.NewMockFileSystem()

	_, err = readThrottleFile(mockFileSystem, "throttle")
	if err == nil {
		assert.Fail(t, "Should have failed to read from a throttle file. "+err.Error())
	}
	assert.Contains(t, err.Error(), "GAL1048", "Error returned should contain GAL1048 error indicating read throttle file failed."+err.Error())
}

func TestReadBackThrottleFileFailsIfFileContainsInvalidInt(t *testing.T) {

	var err error
	mockFileSystem := utils.NewMockFileSystem()

	mockFileSystem.WriteTextFile("throttle", "abc")

	_, err = readThrottleFile(mockFileSystem, "throttle")
	if err == nil {
		assert.Fail(t, "Should have failed to read from a throttle file. "+err.Error())
	}
	assert.Contains(t, err.Error(), "GAL1049E", "Error returned should contain GAL1049E error indicating read invalid throttle file content."+err.Error())
}

func TestUpdateThrottleFromFileIfDifferentChangesValueWhenDifferent(t *testing.T) {

	mockFileSystem := utils.NewMockFileSystem()

	mockFileSystem.WriteTextFile("throttle", "10")
	newValue, isLost := updateThrottleFromFileIfDifferent(mockFileSystem, "throttle", 20, false)

	assert.Equal(t, 10, newValue)
	assert.False(t, isLost)
}

func TestUpdateThrottleFromFileIfDifferentDoesntChangeIfFileMissing(t *testing.T) {

	mockFileSystem := utils.NewMockFileSystem()

	// mockFileSystem.WriteTextFile("throttle", "10") - file is missing now.
	newValue, isLost := updateThrottleFromFileIfDifferent(mockFileSystem, "throttle", 20, false)

	assert.Equal(t, 20, newValue)
	assert.True(t, isLost)
}

func TestOverridesReadFromOverridesFile(t *testing.T) {

	fileProps := make(map[string]interface{})
	fileProps["c"] = "d"

	fs := utils.NewMockFileSystem()
	utils.WritePropertiesFile(fs, "/tmp/temp.properties", fileProps)

	commandParameters := utils.RunsSubmitCmdParameters{
		Overrides:        []string{"a=b"},
		OverrideFilePath: "/tmp/temp.properties",
	}

	overrides, err := buildOverrideMap(fs, commandParameters)

	assert.Nil(t, err)
	assert.NotNil(t, overrides)
	assert.Contains(t, overrides, "a", "command-line override wasn't used.")
	assert.Equal(t, overrides["a"], "b", "command-line override not passed correctly.")
	assert.Contains(t, overrides, "c", "file-based override wasn't used")
	assert.Equal(t, overrides["c"], "d", "file-based override value wasn't passed correctly.")
}

func TestOverridesWithoutOverridesFile(t *testing.T) {

	fs := utils.NewMockFileSystem()

	commandParameters := utils.RunsSubmitCmdParameters{
		Overrides:        []string{"a=b"},
		OverrideFilePath: "",
	}

	// Make sure it doesn't blow up if there is no .galasa folder
	overrides, err := buildOverrideMap(fs, commandParameters)

	assert.Nil(t, err)
	assert.NotNil(t, overrides)
}

func TestOverridesWithDashFileDontReadFromAnyFile(t *testing.T) {

	fs := utils.NewMockFileSystem()

	commandParameters := utils.RunsSubmitCmdParameters{
		Overrides:        []string{"a=b"},
		OverrideFilePath: "-",
	}

	overrides, err := buildOverrideMap(fs, commandParameters)

	assert.Nil(t, err)
	assert.NotNil(t, overrides)
	assert.Contains(t, overrides, "a", "command-line override wasn't used.")
	assert.Equal(t, overrides["a"], "b", "command-line override not passed correctly.")
}

func TestValidateAndCorrectParametersSetsDefaultOverrideFile(t *testing.T) {

	fs := utils.NewMockFileSystem()

	commandParameters := &utils.RunsSubmitCmdParameters{
		Overrides:        []string{"a=b"},
		OverrideFilePath: "",
	}

	regexSelectValue := false
	submitSelectionFlags := &TestSelectionFlags{
		bundles:     new([]string),
		packages:    new([]string),
		tests:       new([]string),
		tags:        new([]string),
		classes:     new([]string),
		stream:      "myStream",
		regexSelect: &regexSelectValue,
	}

	mockLauncher := launcher.NewMockLauncher()

	err := validateAndCorrectParams(fs, commandParameters, mockLauncher, submitSelectionFlags)

	assert.Nil(t, err)
	assert.NotEmpty(t, commandParameters.OverrideFilePath)
}
