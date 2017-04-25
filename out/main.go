package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/ci-pipeline/concourse-ci-resource/utils"
	"github.com/ci-pipeline/packer-resource/docker"
	"github.com/concourse/atc"
)

type Source struct {
	Type string `json:"type"`
}

type DockerParams struct {
	BuildDir           string `json:"build_dir"`
	PackerJson         string `json:"packer_json"`
	VersionDir         string `json:"version_dir"`
	VarFile            string `json:"var_file"`
	AwsAccessKeyId     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
}

func main() {
	os.Chdir(os.Args[1])
	input := utils.GetInput()
	utils.Logln(input)

	source := input.Source.(Source)

	if source.Type == "docker" {
		params := input.Params.(DockerParams)
		utils.Logln(params)
		docker.CgroupfsMount()
		cmd := docker.StartDocker()

		b, err := ioutil.ReadFile(params.VersionDir + "/version")
		if err != nil {
			panic(err)
		}
		version := strings.TrimSpace(string(b))

		os.Chdir(params.BuildDir)

		commonArgs := []string{}
		commonArgs = append(commonArgs, "-only=docker")
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "version="+version)
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "aws_access_key="+params.AwsAccessKeyId)
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "aws_secret_key="+params.AwsSecretAccessKey)
		commonArgs = append(commonArgs, params.PackerJson)
		if docker.RunCmd("packer", append([]string{"validate"}, commonArgs...)...) != 0 {
			utils.Logln("packer script was not validated")
			if err := cmd.Process.Kill(); err != nil {
				log.Fatal("failed to kill: ", err)
			}
			os.Exit(1)
		}
		if docker.RunCmd("packer", append([]string{"build"}, commonArgs...)...) != 0 {
			utils.Logln("Was not built")
			if err := cmd.Process.Kill(); err != nil {
				log.Fatal("failed to kill: ", err)
			}
			os.Exit(1)
		}

	}

	metadata := []atc.MetadataField{atc.MetadataField{Name: "Test", Value: "Value"}}
	result := utils.VersionResult{
		Version:  atc.Version{"docker": "sha12312"},
		Metadata: metadata,
	}
	output, _ := json.Marshal(result)
	fmt.Printf("%s", string(output))
}
