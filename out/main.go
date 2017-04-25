package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/ci-pipeline/concourse-ci-resource/utils"
	"github.com/ci-pipeline/packer-resource/docker"
	"github.com/concourse/atc"
	"github.com/mitchellh/mapstructure"
)

type Source struct {
	Type string `json:"type"`
}

type DockerParams struct {
	BuildDir           string `mapstructure:"build_dir"`
	PackerJson         string `mapstructure:"packer_json"`
	VersionDir         string `mapstructure:"version_dir"`
	VarFile            string `mapstructure:"var_file"`
	AwsAccessKeyId     string `mapstructure:"aws_access_key_id"`
	AwsSecretAccessKey string `mapstructure:"aws_secret_access_key"`
}

func getNameServer() string {
	b, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		panic(err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if fields[0] == "nameserver" {
			nameserver := net.ParseIP(fields[1])
			if nameserver != nil {
				return nameserver.String()
			}
		}
	}
	return ""
}

func main() {
	os.Chdir(os.Args[1])
	input := utils.GetInput()

	var source Source
	err := mapstructure.Decode(input.Source, &source)
	if err != nil {
		panic(err)
	}


	if source.Type == "docker" {
		var params DockerParams
		if err := mapstructure.Decode(input.Params, &params); err != nil {
			panic(err)
		}
		docker.CgroupfsMount()
		cmd := docker.StartDocker()
		var b []byte
		var err error
		b, err = ioutil.ReadFile(params.VersionDir + "/version")
		if err != nil {
			panic(err)
		}
		version := strings.TrimSpace(string(b))

		nameserver := getNameServer()

		os.Chdir(params.BuildDir)
		commonArgs := []string{}
		commonArgs = append(commonArgs, "-only=docker")
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "version="+version)
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "aws_access_key_id="+params.AwsAccessKeyId)
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "aws_secret_access_key="+params.AwsSecretAccessKey)
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "nameserver="+nameserver)
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

		//metadata := []atc.MetadataField{atc.MetadataField{Name: "Test", Value: "Value"}}
		result := utils.VersionResult{
			Version:  atc.Version{"docker": version},
			//Metadata: metadata,
		}
		output, _ := json.Marshal(result)
		fmt.Printf("%s", string(output))
	}
}

