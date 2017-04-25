package main

import (
	"io/ioutil"
	"os"

	"github.com/ci-pipeline/concourse-ci-resource/utils"
)

func main() {
	utils.Logln("Running in")
	destination := os.Args[1]
	input := utils.GetInput()
	utils.Logln(input.Version)

	for k, v := range input.Version.(map[string]interface{}) {
		err := ioutil.WriteFile(destination+"/"+k, []byte(v.(string)), 0644)
		if err != nil {
			panic(err)
		}
	}

}
