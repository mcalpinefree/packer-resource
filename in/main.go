package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ci-pipeline/concourse-ci-resource/utils"
	"github.com/concourse/atc"
)

func main() {
	utils.Logln("Running in")
	destination := os.Args[1]
	input := utils.GetInput()
	utils.Logln(input.Version)

	var packer_type string
	var packer_version string
	for k, v := range input.Version.(map[string]interface{}) {
		packer_type = k
		packer_version = v.(string)
	}
	err := ioutil.WriteFile(destination+"/"+packer_type, []byte(packer_version), 0644)
	if err != nil {
		panic(err)
	}

	metadata := []atc.MetadataField{atc.MetadataField{Name: "Test", Value: "Value"}}
	result := utils.VersionResult{
		Version:  atc.Version{packer_type: packer_version},
		Metadata: metadata,
	}

	output, _ := json.Marshal(result)
	fmt.Printf("%s", string(output))
}
