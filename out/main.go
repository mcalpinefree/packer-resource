package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/ci-pipeline/concourse-ci-resource/utils"
	"github.com/ci-pipeline/packer-resource/docker"
	"github.com/concourse/atc"
	"github.com/mitchellh/mapstructure"
)

type Source struct {
	Type string `json:"type"`
}

type AmazonEbsParams struct {
	SourceAmiPath      string `mapstructure:"source_ami_path"`
	BuildDir           string `mapstructure:"build_dir"`
	PackerJson         string `mapstructure:"packer_json"`
	VersionDir         string `mapstructure:"version_dir"`
	VarFile            string `mapstructure:"var_file"`
	AwsAccessKeyId     string `mapstructure:"aws_access_key_id"`
	AwsSecretAccessKey string `mapstructure:"aws_secret_access_key"`
	VpcId              string `mapstructure:"vpc_id"`
	SubnetId           string `mapstructure:"subnet_id"`
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

	var result utils.VersionResult

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
		if _, exitStatus := docker.RunCmd("packer", append([]string{"validate"}, commonArgs...)...); exitStatus != 0 {
			utils.Logln("packer script was not validated")
			if err := cmd.Process.Kill(); err != nil {
				log.Fatal("failed to kill: ", err)
			}
			os.Exit(1)
		}
		packerOutput, exitStatus := docker.RunCmd("packer", append([]string{"build"}, commonArgs...)...)

		if exitStatus != 0 {
			utils.Logln("Was not built")
			if err := cmd.Process.Kill(); err != nil {
				log.Fatal("failed to kill: ", err)
			}
			os.Exit(1)
		}

		re := regexp.MustCompile("--> docker: Imported Docker image: (sha256:[a-z0-9]+)")
		var dockerImage string
		for _, line := range strings.Split(packerOutput, "\n") {
			match := re.FindStringSubmatch(line)
			if match != nil && len(match[1]) > 0 {
				dockerImage = match[1]
				break
			}
		}

		//metadata := []atc.MetadataField{atc.MetadataField{Name: "Test", Value: "Value"}}
		result = utils.VersionResult{
			Version: atc.Version{"docker": dockerImage},
			//Metadata: metadata,
		}
	} else if source.Type == "amazon-ebs" {
		var params AmazonEbsParams
		if err := mapstructure.Decode(input.Params, &params); err != nil {
			panic(err)
		}
		var b []byte
		var err error
		b, err = ioutil.ReadFile(params.VersionDir + "/version")
		if err != nil {
			panic(err)
		}
		version := strings.TrimSpace(string(b))

		b, err = ioutil.ReadFile(params.SourceAmiPath)
		if err != nil {
			panic(err)
		}
		sourceAmi := strings.TrimSpace(string(b))

		os.Chdir(params.BuildDir)
		commonArgs := []string{}
		commonArgs = append(commonArgs, "-only=amazon-ebs")
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "version="+version)
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "aws_access_key_id="+params.AwsAccessKeyId)
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "aws_secret_access_key="+params.AwsSecretAccessKey)
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "vpc_id="+params.VpcId)
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "subnet_id="+params.SubnetId)
		commonArgs = append(commonArgs, "-var")
		commonArgs = append(commonArgs, "source_ami="+sourceAmi)
		commonArgs = append(commonArgs, params.PackerJson)
		if _, exitStatus := docker.RunCmd("packer", append([]string{"validate"}, commonArgs...)...); exitStatus != 0 {
			utils.Logln("packer script was not validated")
			os.Exit(1)
		}
		packerOutput, exitStatus := docker.RunCmd("packer", append([]string{"build"}, commonArgs...)...)

		if exitStatus != 0 {
			utils.Logln("Was not built")
			os.Exit(1)
		}

		re := regexp.MustCompile("[a-z]+-[a-z]+-[0-9]: (ami-[a-z0-9]+)")
		var ami string
		for _, line := range strings.Split(packerOutput, "\n") {
			match := re.FindStringSubmatch(line)
			if match != nil && len(match[1]) > 0 {
				ami = match[1]
				break
			}
		}

		//metadata := []atc.MetadataField{atc.MetadataField{Name: "Test", Value: "Value"}}
		result = utils.VersionResult{
			Version: atc.Version{"ami": ami},
			//Metadata: metadata,
		}
	}

	output, _ := json.Marshal(result)
	fmt.Printf("%s", string(output))
}
