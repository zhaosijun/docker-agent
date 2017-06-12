package models

import (
	"errors"
	"io"
	"bytes"
	"encoding/json"
	"strings"
    "net/http"
    "os"
    "time"
    "strconv"
    "fmt"
	
	"github.com/astaxie/beego"
	"github.com/fsouza/go-dockerclient"
)
var (
	Registry string
	ContextRootDir string
	DefaultTag string
	DockerfileTemplateDir string

	DevEnvHost string
	TestEnvHost string
	DeployEnvHost string
	ClusterEnvHost string
)

func init() {
	Registry = "gcr.io"
	DefaultTag = "latest"
	ContextRootDir = "build-package"
	DockerfileTemplateDir = "conf/dockerfile-template"

	DevEnvHost = beego.AppConfig.DefaultString("develophost", "0.0.0.0")
	TestEnvHost = beego.AppConfig.DefaultString("testhost", "0.0.0.0")
	DeployEnvHost = beego.AppConfig.DefaultString("deployhost", "0.0.0.0")
	ClusterEnvHost = beego.AppConfig.DefaultString("clusterhost", "0.0.0.0")
}

type DeployJarOpts struct {
	JarUrl	   			string
	DeployOpts 			*DeployOpts
}

type DeployOpts struct {
	InstanceName		string
	Labels     			map[string]string
	HostConfig  		HostConfig
	Env 				map[string]string
}

type PushAndBuildJarOpts struct {
	JarUrl	   			string
	ImageAlias 			string
	Labels     			map[string]string
}

type DeployContextOpts struct {
	JarUrl	   			string
	DockerfileTemplate  string
	ContextDir  		string
	Repository   		string
	Registry    		string
	ImageName 			string
	ImageAlias 			string
	Tag         		string
	InstanceName		string
	Labels 				map[string]string
	HostConfig  		*docker.HostConfig
	Config 				*docker.Config
}

func GetContextEndpoint(context string) (string, error) {
	switch context {
	case "test":
		return fmt.Sprintf("http://%s:5555", TestEnvHost), nil
	case "develop":
		return fmt.Sprintf("http://%s:5555", DevEnvHost), nil
	case "deploy":
		return fmt.Sprintf("http://%s:5555", DeployEnvHost), nil
	case "cluster":
		return fmt.Sprintf("http://%s:5555", ClusterEnvHost), nil
	case "test/cadvisor":
		return fmt.Sprintf("http://%s:%s", TestEnvHost, CadvisorPort), nil
	case "develop/cadvisor":
		return fmt.Sprintf("http://%s:%s", DevEnvHost, CadvisorPort), nil
	case "deploy/cadvisor":
		return fmt.Sprintf("http://%s:%s", DeployEnvHost, CadvisorPort), nil
	default:
		return "", errors.New("context is error")
	}
}

func DeployByJar(endpoint string, deployJarOpts *DeployJarOpts) (containerId string, err error) {
	
	deployContextOpts, err := parseJarDeployOptions(deployJarOpts)
	if err != nil {
		return "", err
	}

	err = downloadJarPackageAndDockerfile(deployContextOpts)
	if err != nil {
		return "", err
	}

	client, err := docker.NewClient(endpoint)
	if err != nil {
		return "", err
	}

	err = build(client, deployContextOpts)
	if err != nil {
		return "", err
	}

	// err = push(client, deployContextOpts)
	// if err != nil {
	// 	return "", err
	// }

	// err = insertImageLabel(client, deployContextOpts)
	// if err != nil {
	// 	return "", err
	// }

	container, err := createContainer(client, deployContextOpts)
	if err != nil {
		return "", err
	}

	err = startContainer(client, container.ID, deployContextOpts)
	if err != nil {
		return "", err
	}

	return container.ID, nil
}

//func BuildAndPushImageByJar(endpoint string, deployJarOpts *DeployJarOpts) error {
func BuildAndPushImageByJar(endpoint string, pushAndBuildOpts *PushAndBuildJarOpts) error {

	deployContextOpts, err := parsePushAndBuildOptions(pushAndBuildOpts)
	if err != nil {
		return err
	}

	err = downloadJarPackageAndDockerfile(deployContextOpts)
	if err != nil {
		return err
	}

	client, err := docker.NewClient(endpoint)
	if err != nil {
		return err
	}

	err = build(client, deployContextOpts)
	if err != nil {
		return err
	}

	err = push(client, deployContextOpts)
	if err != nil {
		return err
	}

	err = insertImageLabel(client, deployContextOpts)
	if err != nil {
		return err
	}

	return nil
}

func LogVarInfo(info string, v interface{}) {
	b, _ := json.MarshalIndent(v, "", "   ")
	beego.Info(info + string(b))
}

func createContainer(client *docker.Client, deployContextOpts *DeployContextOpts) (*docker.Container, error) {
	LogVarInfo("createContainer: ", deployContextOpts)
	var createOpts docker.CreateContainerOptions
	
	createOpts.Config = deployContextOpts.Config
	createOpts.HostConfig = deployContextOpts.HostConfig
	createOpts.Name = deployContextOpts.InstanceName

	LogVarInfo("createOpts: ", &createOpts)
	return client.CreateContainer(createOpts)
}

func startContainer(client *docker.Client, containerId string, deployContextOpts *DeployContextOpts) error {
	beego.Info("startContainer: " + "containerId: " + containerId)
	LogVarInfo("deployContextOpts: ", deployContextOpts)
	var hostConfig docker.HostConfig

	LogVarInfo("hostConfig: ", &hostConfig)
	return client.StartContainer(containerId, &hostConfig)
}

func build(client *docker.Client, deployContextOpts *DeployContextOpts) error {
	LogVarInfo("build: ", deployContextOpts)

	var buf bytes.Buffer
	var buildOpt docker.BuildImageOptions

	buildOpt.Name = deployContextOpts.ImageName
	buildOpt.ContextDir = deployContextOpts.ContextDir
	buildOpt.Labels = deployContextOpts.Labels
	buildOpt.BuildArgs = append(buildOpt.BuildArgs, docker.BuildArg{"JAR", deployContextOpts.Repository})
	buildOpt.OutputStream = &buf

	LogVarInfo("buildOpt: ", &buildOpt)
	err := client.BuildImage(buildOpt)
	if err != nil {
		beego.Info(err)
		beego.Info(buf.String())
		return errors.New(err.Error() + "\n" + buf.String())
	}

	return nil
}

func push(client *docker.Client, deployContextOpts *DeployContextOpts) error {
	LogVarInfo("build: ", deployContextOpts)

	var buf bytes.Buffer
	var pushOpt docker.PushImageOptions

	pushOpt.Name = deployContextOpts.ImageName
	pushOpt.Tag = deployContextOpts.Tag
	pushOpt.OutputStream = &buf

	LogVarInfo("pushOpt: ", &pushOpt)
	err := client.PushImage(pushOpt, docker.AuthConfiguration{})
	if err != nil {
		beego.Info(err)
		beego.Info(buf.String())
		return errors.New(err.Error() + "\n" + buf.String())
	}

	return nil
}

func insertImageLabel(client *docker.Client, deployContextOpts *DeployContextOpts) error {
	pImage, err := client.InspectImage(deployContextOpts.ImageName)
	if err != nil {
		beego.Info(err)
		return err
	}

	image := Imagereference{Name: deployContextOpts.ImageName, 
							Tag: deployContextOpts.Tag, 
							RegistryCentral: deployContextOpts.Registry,
							Repository: deployContextOpts.Repository,
							Size: pImage.Size,
							Pushed: time.Now(),
							Alias: deployContextOpts.ImageAlias}

	var labels []*Label
	for key, value := range deployContextOpts.Labels {
		label := Label{Name: key + "=" + value}
		labels = append(labels, &label)
	}
	
	err = AddImageM2MLabel(&image, labels)
	if err != nil {
		return err
	}

	return nil
}

func GetJVMOpts(hostConfig *docker.HostConfig) string {

	if hostConfig == nil {
		return ""
	}
	
	if hostConfig.Memory == 0 {
		return ""
	}

	// if hostConfig.MemorySwap == 0 {
	// 	hostConfig.MemorySwap = 2*hostConfig.Memory
	// }

	memoryLimitValue := strconv.Itoa(int((hostConfig.Memory)*7/10/1024/1024)) + "m"

	return "JAVA_OPTIONS=-Xmx" + memoryLimitValue
}

func ParseDeployOptions(deployOpts *DeployOpts, deployContextOpts *DeployContextOpts) error {

	if deployOpts == nil || deployContextOpts == nil {
		return errors.New("Input parameter is null")
	}

	var hostConfig docker.HostConfig
    ConvertHostConfigFromFrontToDockerClient(&deployOpts.HostConfig, &hostConfig)
    deployContextOpts.HostConfig = &hostConfig

    var config docker.Config
    var envs []string
    config.Image = deployContextOpts.ImageName
	for k, v := range deployOpts.Env {
		if k != "JAVA_OPTIONS" {
			envs = append(envs, k + "=" + v) 
		}
	}
	javaOpt := GetJVMOpts(&hostConfig)
	if javaOpt != "" {
		envs = append(envs, javaOpt)
	}

	config.Env = envs
    config.Labels = deployOpts.Labels

    deployContextOpts.Config = &config

    deployContextOpts.InstanceName = deployOpts.InstanceName
	deployContextOpts.Labels = deployOpts.Labels

	return nil
}

func parseJarUrlOptions(url string, deployContextOpts *DeployContextOpts) error {

	if deployContextOpts == nil {
		return errors.New("deployContextOpts is null")
	}

	pathStrs := strings.Split(url, "/")
    jarName := strings.ToLower(pathStrs[len(pathStrs) - 1])

    deployContextOpts.ContextDir = ContextRootDir + "/" + jarName
	deployContextOpts.Repository = jarName
	deployContextOpts.Registry = Registry
	deployContextOpts.Tag = DefaultTag
	deployContextOpts.ImageName = Registry + "/" + jarName + ":" + deployContextOpts.Tag
	deployContextOpts.JarUrl = url

    nameStrs := strings.Split(jarName, ".")
    if nameStrs[len(nameStrs) - 1] == "jar" {
    	deployContextOpts.DockerfileTemplate = DockerfileTemplateDir + "/jar/Dockerfile"
    } else if nameStrs[len(nameStrs) - 1] == "war" {
    	deployContextOpts.DockerfileTemplate = DockerfileTemplateDir + "/war/Dockerfile"
    } else {
    	beego.Info("The Url of jar package is error")
    	return errors.New("The Url of jar package is error")
    }

    return nil	
}

func parsePushAndBuildOptions(pushAndBuildOpts *PushAndBuildJarOpts) (*DeployContextOpts, error) {
	LogVarInfo("parsePushAndBuildOptions: ", pushAndBuildOpts)

	if pushAndBuildOpts == nil {
		return nil, errors.New("pushAndBuildOpts is null")
	}

	var deployContextOpts DeployContextOpts
	err := parseJarUrlOptions(pushAndBuildOpts.JarUrl, &deployContextOpts)
	if err != nil {
		return nil, err
	}

	deployContextOpts.Labels = pushAndBuildOpts.Labels
	deployContextOpts.ImageAlias = pushAndBuildOpts.ImageAlias
	
	return &deployContextOpts, nil
}

func parseJarDeployOptions(deployJarOpts *DeployJarOpts) (*DeployContextOpts, error) {
	LogVarInfo("parseDeployOptions: ", deployJarOpts)

	if deployJarOpts == nil {
		return nil, errors.New("deployJarOpts is null")
	}
	var deployContextOpts DeployContextOpts
	err := parseJarUrlOptions(deployJarOpts.JarUrl, &deployContextOpts)
	if err != nil {
		return nil, err
	}
	// pathStrs := strings.Split(deployJarOpts.JarUrl, "/")
 //    jarName := pathStrs[len(pathStrs) - 1]

 //    deployContextOpts.ContextDir = ContextRootDir + "/" + jarName
	// deployContextOpts.Repository = jarName
	// deployContextOpts.Registry = Registry
	// deployContextOpts.Tag = DefaultTag
	// deployContextOpts.ImageName = Registry + "/" + jarName + ":" + deployContextOpts.Tag

 //    nameStrs := strings.Split(jarName, ".")
 //    if nameStrs[len(nameStrs) - 1] == "jar" {
 //    	deployContextOpts.DockerfileTemplate = DockerfileTemplateDir + "/jar/Dockerfile"
 //    } else if nameStrs[len(nameStrs) - 1] != "war" {
 //    	deployContextOpts.DockerfileTemplate = DockerfileTemplateDir + "/war/Dockerfile"
 //    } else {
 //    	beego.Info("The Url of jar package is error")
 //    	return nil, errors.New("The Url of jar package is error")
 //    }	


	err = ParseDeployOptions(deployJarOpts.DeployOpts, &deployContextOpts)
	if err != nil {
		return nil, err
	}
    
 //    var hostConfig docker.HostConfig
 //    ConvertHostConfigFromFrontToDockerClient(&deployJarOpts.HostConfig, &hostConfig)
 //    deployContextOpts.HostConfig = &hostConfig

 //    var config docker.Config
 //    var envs []string
 //    config.Image = deployContextOpts.ImageName
	// for k, v := range deployJarOpts.Env {
	// 	if k != "JAVA_OPTIONS" {
	// 		envs = append(envs, k + "=" + v) 
	// 	}
	// }
	// javaOpt := GetJVMOpts(&hostConfig)
	// if javaOpt != "" {
	// 	envs = append(envs, javaOpt)
	// }

	// config.Env = envs
 //    config.Labels = deployJarOpts.Labels

 //    deployContextOpts.Config = &config

 //    deployContextOpts.InstanceName = deployJarOpts.InstanceName
	// deployContextOpts.Labels = deployJarOpts.Labels

    return &deployContextOpts, nil
}

func downloadJarPackageAndDockerfile(deployContextOpts *DeployContextOpts) error {
	LogVarInfo("downloadJarPackageAndDockerfile: ", deployContextOpts)
	err := os.MkdirAll(deployContextOpts.ContextDir, 0755) 
    if err != nil { 
    	if !os.IsExist(err) {
    		return err
    	}
	}

    jarFile, err := os.Create(deployContextOpts.ContextDir + "/" + deployContextOpts.Repository)
    if err != nil {
        beego.Info(err)
        return err
    }
    defer jarFile.Close()

	rsp, err := http.Get(deployContextOpts.JarUrl)
    if err != nil {
        beego.Info(err)
        return err   
    }

    _, err = io.Copy(jarFile, rsp.Body)
    if err != nil {
        beego.Info(err)
        return err
    }

    dockerFile, err := os.Create(deployContextOpts.ContextDir + "/Dockerfile")
    if err != nil {
        beego.Info(err)
        return err
    }
    defer dockerFile.Close()

    dockerFileTemplate, err := os.Open(deployContextOpts.DockerfileTemplate)
    if err != nil {
        beego.Info(err)
        return err
    }
    defer dockerFile.Close()

    _, err = io.Copy(dockerFile, dockerFileTemplate)
    if err != nil {
        beego.Info(err)
        return err
    }

    return nil
}


