package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Inputs
	pathToXcresult := os.Getenv("path_to_xcresult")
	outputDir := os.Getenv("xml_output_dir")
	stepSourceDir := os.Getenv("BITRISE_STEP_SOURCE_DIR")

	outputJSON := filepath.Join(outputDir, "coverage.json")
	outputXML := filepath.Join(outputDir, "cobertura.xml")
	conversionScript := filepath.Join(stepSourceDir, "xccov-json-to-cobertura-xml.swift")

	// for _, pair := range os.Environ() {
	// 	fmt.Println(pair)
	// }

	// Generate JSON from XCResult
	fmt.Println("Generating json from 'path_to_xcresult':", pathToXcresult)

	cmd1 := exec.Command("xcrun", "xccov", "view", "--report", "--json", pathToXcresult)
	err := runAndSaveToFile(cmd1, outputJSON)
	if err != nil {
		fmt.Printf("Failed to generate coverage json, error: %#v", err.Error())
		os.Exit(1)
	}

	// Generate XML from JSON
	fmt.Println("Generating xml from", outputJSON)

	cmd2 := exec.Command("xcrun", "swift", conversionScript, outputJSON)
	err2 := runAndSaveToFile(cmd2, outputXML)
	if err2 != nil {
		fmt.Printf("Failed to generate cobertura xml, error: %#v", err2.Error())
		os.Exit(1)
	}

	//
	// --- Step Outputs: Export Environment Variables for other Steps:
	// You can export Environment Variables for other Steps with
	//  envman, which is automatically installed by `bitrise setup`.
	// A very simple example:
	cmdLog3, err3 := exec.Command("bitrise", "envman", "add", "--key", "COVERAGE_XML_TEST_RESULT_PATH", "--value", outputXML).CombinedOutput()
	if err3 != nil {
		fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err3.Error(), cmdLog3)
		os.Exit(1)
	}
	cmdLog4, err4 := exec.Command("bitrise", "envman", "add", "--key", "COVERAGE_JSON_TEST_RESULT_PATH", "--value", outputJSON).CombinedOutput()
	if err4 != nil {
		fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err4.Error(), cmdLog4)
		os.Exit(1)
	}
	//
	// --- Exit codes:
	// The exit code of your Step is very important. If you return
	//  with a 0 exit code `bitrise` will register your Step as "successful".
	// Any non zero exit code will be registered as "failed" by `bitrise`.
	os.Exit(0)
}

func runAndSaveToFile(cmd *exec.Cmd, outfile string) error {
	stdOutStream, err := os.Create(outfile)
	if err != nil {
		return err
	}
	cmd.Stdout = stdOutStream

	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	stdOutStream.Close()
	return nil
}
