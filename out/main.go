package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ci-pipeline/concourse-ci-resource/utils"
	"github.com/concourse/atc"
)

func main() {
	utils.Logln("Change to build directory")
	utils.GoToBuildDirectory()
	cwd, _ := os.Getwd()
	utils.Logln(cwd)
	input := utils.GetInput()
	utils.Logln(input)
	metadata := []atc.MetadataField{atc.MetadataField{Name: "Test", Value: "Value"}}
	result := utils.VersionResult{Metadata: metadata}
	output, _ := json.Marshal(result)
	fmt.Printf("%s", string(output))
}
