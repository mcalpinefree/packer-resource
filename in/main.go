package main

import (
	"encoding/json"
	"fmt"

	"github.com/ci-pipeline/concourse-ci-resource/utils"
)

func main() {
	result := utils.VersionResult{}
	output, _ := json.Marshal(result)
	fmt.Println(string(output))
}
