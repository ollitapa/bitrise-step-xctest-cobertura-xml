package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Xcode coverage report structure

// FunctionCoverageReport structure
type FunctionCoverageReport struct {
	CoveredLines    int
	ExecutableLines int
	ExecutionCount  int
	LineCoverage    float64
	LineNumber      int
	Name            string
}

// FileCoverageReport structure
type FileCoverageReport struct {
	CoveredLines    int
	ExecutableLines int
	Functions       []FunctionCoverageReport
	LineCoverage    float64
	Name            string
	Path            string
}

// TargetCoverageReport structure
type TargetCoverageReport struct {
	BuildProductPath string
	CoveredLines     int
	ExecutableLines  int
	Files            []FileCoverageReport
	LineCoverage     float64
	Name             string
}

// CoverageReport structure
type CoverageReport struct {
	ExecutableLines int
	Targets         []TargetCoverageReport
	LineCoverage    float64
	CoveredLines    int
}

// Cobertura XML coverage structure

// XMLCoverage structure
type XMLCoverage struct {
	XMLName    xml.Name `xml:"coverage"`
	LineRate   string   `xml:"line-rate,attr"`
	BranchRate string   `xml:"branch-rate,attr"`

	LinesCovered string `xml:"lines-covered,attr"`
	LinesValid   string `xml:"lines-valid,attr"`

	TimeStamp       string `xml:"timestamp,attr"`
	Vesion          string `xml:"version,attr"`
	Complexity      string `xml:"complexity,attr"`
	BranchesValid   string `xml:"branches-valid,attr"`
	BranchesCovered string `xml:"branches-covered,attr"`

	Source   []string     `xml:"sources>source"`
	Packages []XMLPackage `xml:"packages>package"`
}

// XMLPackage structure
type XMLPackage struct {
	XMLName xml.Name   `xml:"package"`
	Classes []XMLClass `xml:"classes>class"`

	Name       string `xml:"name,attr"`
	LineRate   string `xml:"line-rate,attr"`
	BranchRate string `xml:"branch-rate,attr"`
	Complexity string `xml:"complexity,attr"`
}

// XMLClass structure
type XMLClass struct {
	XMLName xml.Name `xml:"class"`

	Name       string `xml:"name,attr"`
	Filename   string `xml:"filename,attr"`
	LineRate   string `xml:"line-rate,attr"`
	BranchRate string `xml:"branch-rate,attr"`
	Complexity string `xml:"complexity,attr"`

	Lines []XMLLine `xml:"lines>line"`
}

// XMLLine structure
type XMLLine struct {
	XMLName xml.Name `xml:"line"`
	Number  string   `xml:"number,attr"`
	Branch  string   `xml:"branch,attr"`
	Hits    string   `xml:"hits,attr"`
}

func main() {
	// Inputs
	pathToXcresult := os.Getenv("path_to_xcresult")
	outputDir := os.Getenv("xml_output_dir")
	sourceDir := os.Getenv("path_to_source_dir")

	ConvertXcodeCoverageToCobetura(
		pathToXcresult,
		outputDir,
		sourceDir,
	)
}

// ConvertXcodeCoverageToCobetura Converter
// Converts xcode coverage file at `pathToXcresult` to
// cobertura compatible xml format. Output is saved to
// given directory. Provide also base directory of the
// code source, so it can be referenced in the xml.
func ConvertXcodeCoverageToCobetura(
	pathToXcresult string,
	outputDir string,
	sourceDir string,
) {

	outputJSON := filepath.Join(outputDir, "coverage.json")
	outputXML := filepath.Join(outputDir, "cobertura.xml")

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

	jsonData, jsonErr := ioutil.ReadFile(outputJSON)
	//fmt.Println("Contents of file:", string(jsonData))
	//fmt.Printf("%+v\n", report)
	if jsonErr != nil {
		fmt.Printf("Failed to read coverage json, error: %#v", jsonErr.Error())
		os.Exit(1)
	}

	// Decode Xcode coverage report from JSON
	var report CoverageReport
	json.Unmarshal(jsonData, &report)

	// Create cobertura main coverage item
	xmlCov := &XMLCoverage{
		LineRate:        fmt.Sprintf("%f", report.LineCoverage),
		TimeStamp:       fmt.Sprintf("%d", time.Now().Unix()),
		LinesCovered:    fmt.Sprintf("%d", report.CoveredLines),
		LinesValid:      fmt.Sprintf("%d", report.ExecutableLines),
		Vesion:          "diff_coverage 1.0",
		BranchRate:      "1.0", // Not present in xcresult
		Complexity:      "0.0", // Not present in xcresult
		BranchesValid:   "1.0", // Not present in xcresult
		BranchesCovered: "1.0", // Not present in xcresult
		Source:          []string{sourceDir},
	}

	var packs []XMLPackage

	// Go through each target and treat as package
	for _, target := range report.Targets {

		// If package is empty, skip
		if len(target.Files) < 1 {
			continue
		}

		// Package name
		targetPath, _ := filepath.Split(target.Files[0].Path)
		packageName := strings.ReplaceAll(targetPath, "/", ".")
		packageName = strings.Trim(packageName, ".")
		pack := XMLPackage{
			Name:       packageName,
			LineRate:   fmt.Sprintf("%f", target.LineCoverage),
			BranchRate: "1.0", // Not present in xcresult
			Complexity: "0.0", // Not present in xcresult
		}

		var covClasses = []XMLClass{}

		// Go through each file, which will represent class in xml coverage
		for _, file := range target.Files {

			var covClass = XMLClass{
				Name:       packageName + filenameWithoutExtension(file.Name),
				Filename:   strings.Replace(file.Path, sourceDir+"/", "", -1),
				LineRate:   fmt.Sprintf("%f", file.LineCoverage),
				BranchRate: "1.0", // Not present in xcresult
				Complexity: "0.0", // Not present in xcresult
			}

			var covLines = []XMLLine{}

			// Go through each line in each function.
			for _, function := range file.Functions {
				for lineIdx := 0; lineIdx < function.ExecutableLines; lineIdx++ {
					// Function coverage report won't be 100% reliable without parsing it by file
					// (would need to use xccov view --file filePath currentDirectory + Build/Logs/Test/*.xccovarchive)
					lineHits := 0
					if lineIdx < function.CoveredLines {
						lineHits = function.ExecutionCount
					}
					covLine := XMLLine{
						Number: fmt.Sprintf("%d", function.LineNumber+lineIdx),
						Branch: "false",
						Hits:   fmt.Sprintf("%d", lineHits),
					}
					covLines = append(covLines, covLine)
				}
			}

			covClass.Lines = covLines
			covClasses = append(covClasses, covClass)

		}

		pack.Classes = covClasses
		packs = append(packs, pack)
	}

	xmlCov.Packages = packs

	// Decode XML to a file
	out, _ := xml.MarshalIndent(xmlCov, "", "    ")
	//fmt.Printf("%+v\n", out)
	err = writeToFile(outputXML, xml.Header+xmlDTD+string(out))
	if err != nil {
		fmt.Printf("Failed to write xml, error: %#v", err.Error())
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

func writeToFile(outfile string, text string) error {
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	_, err = f.WriteString(text)
	if err != nil {
		f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func filenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

const xmlDTD = `<!DOCTYPE coverage SYSTEM "http://cobertura.sourceforge.net/xml/coverage-04.dtd" [
    <!ELEMENT coverage (sources?, packages)>
    <!ATTLIST coverage line-rate CDATA #REQUIRED>
    <!ATTLIST coverage branch-rate CDATA #REQUIRED>
    <!ATTLIST coverage lines-covered CDATA #REQUIRED>
    <!ATTLIST coverage lines-valid CDATA #REQUIRED>
    <!ATTLIST coverage branches-covered CDATA #REQUIRED>
    <!ATTLIST coverage branches-valid CDATA #REQUIRED>
    <!ATTLIST coverage complexity CDATA #REQUIRED>
    <!ATTLIST coverage version CDATA #REQUIRED>
    <!ATTLIST coverage timestamp CDATA #REQUIRED>
    <!ELEMENT sources (source)*>
    <!ELEMENT source (#PCDATA)>
    <!ELEMENT packages (package)*>
    <!ELEMENT package (classes)>
    <!ATTLIST package name CDATA #REQUIRED>
    <!ATTLIST package line-rate CDATA #REQUIRED>
    <!ATTLIST package branch-rate CDATA #REQUIRED>
    <!ATTLIST package complexity CDATA #REQUIRED>
    <!ELEMENT classes (class)*>
    <!ELEMENT class (methods, lines)>
    <!ATTLIST class name CDATA #REQUIRED>
    <!ATTLIST class filename CDATA #REQUIRED>
    <!ATTLIST class line-rate CDATA #REQUIRED>
    <!ATTLIST class branch-rate CDATA #REQUIRED>
    <!ATTLIST class complexity CDATA #REQUIRED>
    <!ELEMENT methods (method)*>
    <!ELEMENT method (lines)>
    <!ATTLIST method name CDATA #REQUIRED>
    <!ATTLIST method signature CDATA #REQUIRED>
    <!ATTLIST method line-rate CDATA #REQUIRED>
    <!ATTLIST method branch-rate CDATA #REQUIRED>
    <!ATTLIST method complexity CDATA #REQUIRED>
    <!ELEMENT lines (line)*>
    <!ELEMENT line (conditions)*>
    <!ATTLIST line number CDATA #REQUIRED>
    <!ATTLIST line hits CDATA #REQUIRED>
    <!ATTLIST line branch CDATA "false">
    <!ATTLIST line condition-coverage CDATA "100%">
    <!ELEMENT conditions (condition)*>
    <!ELEMENT condition EMPTY>
    <!ATTLIST condition number CDATA #REQUIRED>
    <!ATTLIST condition type CDATA #REQUIRED>
    <!ATTLIST condition coverage CDATA #REQUIRED>
]>

`
