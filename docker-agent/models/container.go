package models

import (
	"errors"
	"time"
	"strconv"

	"github.com/fsouza/go-dockerclient"
	//"github.com/google/cadvisor/info/v2"
)

var (
	DefaultHostIP string
)

func init() {
	DefaultHostIP = "0.0.0.0"
}

type Container struct {
	ContainerId   string
	Status        string
	Name 		  string
	// IPAddress     string
	// Gateway		  string
	// Ports		  map[docker.Port][]docker.PortBinding
	Labels		  map[string]string
	PublishAddr   string
	CPUSetCPUs    string
	Memory        int64
	Created 	  time.Time
	Env 		  []string
	ImageName     string
}

type APIContainer struct {
	ContainerId   string
	Status        string
	Name 		  string
	Labels		  map[string]string
	PublishAddr   string
	ImageName     string
}

type ContainerObject struct {
	Id   		  string
	Created       string
	Name 		  string
	State     	  State
}

type ListContainersOptions struct {
	All    		  bool
	Labels 		  map[string]string
}

type State struct {
	StartedAt  	  string
	FinishedAt    string
	Health 		  docker.Health
}

//For swagger example 
type CreateContainerOptions struct {
	Name             string   
	Config           *Config
	HostConfig		 *HostConfig  
}

//For swagger example 
type Config struct {
	Image  	string
	Labels 	map[string]string   
	Env     map[string]string
}

//For swagger example 
type HostConfig struct {
	Privileged    	bool
	PublishAllPorts bool
	//CPUQuota	  int64
	CPUSetCPUs    	string
	Memory 		  	int64
	MemorySwap    	int64 `json:"-"`
	LogConfig       LogConfig  `json:"-"`
	//PortMappings  []PortMapping  
}

type LogConfig struct {
	Type   string
	Config map[string]string
}

//For swagger example 
type StartContainerOpts struct {
	// Privileged    bool
	// CPUQuota	  int64
	// Memory 		  int64
	// MemorySwap    int64
	// PortMappings  []PortMapping
	//PortBindings  map[docker.Port][]docker.PortBinding
}

type PortMapping struct {
	ExposedPort   	  string
	PortBindings 	  []PortBinding 
}

//For swagger example 
type PortBinding struct {
 	HostIP	 string  `json:"-"` 
 	HostPort string
}

// type ContainerStats struct {
// 	Read      		time.Time
// 	Networks    	map[string]NetworkStats
// 	MemoryStats  	MemoryStats 
// 	CPUStats    	CPUStats
// }

// type NetworkStats struct {
// 	RxBytes   uint64
// 	TxBytes   uint64
// }

// type MemoryStats struct {
// 	Usage    uint64
// 	Limit    uint64
// }

// type CPUStats struct {
// 	CPUUsage struct {
// 		PercpuUsage       []uint64
// 		TotalUsage        uint64
// 	}
// 	SystemCPUUsage  uint64 
// }



func ConvertStartOptsFromFrontToDockerClient (optsSrc *CreateContainerOptions, optsDest *docker.CreateContainerOptions){

	if optsSrc == nil || optsDest == nil {
		return 
	}

	var hostConfig docker.HostConfig
	var config docker.Config
	ConvertHostConfigFromFrontToDockerClient(optsSrc.HostConfig, &hostConfig)
	ConvertConfigFromFrontToDockerClient(optsSrc.Config, &config)
	optsDest.HostConfig = &hostConfig

	optsDest.Config = &config
	optsDest.Name = optsSrc.Name
}

func ConvertConfigFromFrontToDockerClient (confSrc *Config, confDest *docker.Config) error{
	if confSrc == nil || confDest == nil {
		return errors.New("Input parameter is null")
	}

	var envs []string
	for k, v := range confSrc.Env {
		envs = append(envs, k + "=" + v) 
	} 
	confDest.Env = envs

	confDest.Image = confSrc.Image

	confDest.Labels = confSrc.Labels

	return nil
}

func ConvertHostConfigFromFrontToDockerClient (confSrc *HostConfig, confDest *docker.HostConfig) error{

	if confSrc == nil || confDest == nil {
		return errors.New("Input parameter is null")
	}

	// var portBindings map[docker.Port][]docker.PortBinding
	// portBindings = make(map[docker.Port][]docker.PortBinding)
	// for _, portMapping := range confSrc.PortMappings {
	// 	var portBindingSlice []docker.PortBinding
	// 	for _, portBinding := range portMapping.PortBindings {
	// 		portBindingDocker := docker.PortBinding{DefaultHostIP, portBinding.HostPort}
	// 		portBindingSlice = append(portBindingSlice, portBindingDocker)
	// 	}
	// 	portBindings[docker.Port(portMapping.ExposedPort)] = portBindingSlice
	// }

	// confDest.PortBindings = portBindings
	confDest.Privileged = confSrc.Privileged
	confDest.PublishAllPorts = confSrc.PublishAllPorts
	//confDest.CPUQuota = confSrc.CPUQuota
	confDest.CPUSetCPUs = confSrc.CPUSetCPUs
	confDest.Memory = confSrc.Memory
	confDest.MemorySwap = confSrc.MemorySwap
	//confDest.LogConfig.Type = "json-file"
	
	return nil
}

func ConvertContainerFromDockerClientToFront(containerSrc *docker.Container, containerDest *Container, context string) error {
	
	if containerDest == nil || containerSrc == nil {
		return errors.New("Input parameter is null")
	}

	var hostIp string
	switch context {
	case "test":
		hostIp = TestEnvHost
	case "develop":
		hostIp = DevEnvHost
	case "deploy":
		hostIp = DeployEnvHost
	default:
		return errors.New("Input parameter is error")
	}

	containerDest.ContainerId = containerSrc.ID
	containerDest.Status = containerSrc.State.Status
	containerDest.Name = containerSrc.Name
	// containerDest.IPAddress = containerSrc.NetworkSettings.IPAddress
	// containerDest.Gateway = containerSrc.NetworkSettings.Gateway
	// containerDest.Ports = containerSrc.NetworkSettings.Ports
	containerDest.Labels = containerSrc.Config.Labels
	containerDest.Env = containerSrc.Config.Env
	containerDest.ImageName = containerSrc.Config.Image

	containerDest.CPUSetCPUs = containerSrc.HostConfig.CPUSetCPUs
	containerDest.Memory = containerSrc.HostConfig.Memory
	containerDest.Created = containerSrc.Created
	

	ports := containerSrc.NetworkSettings.Ports[docker.Port("8080/tcp")]
	if ports != nil && len(ports) > 0 {
		containerDest.PublishAddr = hostIp + ":" + ports[0].HostPort
	} else {
		containerDest.PublishAddr = "unknow"
	}
	return nil
}

func ConvertAPIContainerFromDockerClientToFront(containersSrc *[]docker.APIContainers, context string) (containersDest []APIContainer, err error) {
	
	if containersSrc == nil {
		return containersDest, errors.New("Input parameter is null")
	}

	var hostIp string
	switch context {
	case "test":
		hostIp = TestEnvHost
	case "develop":
		hostIp = DevEnvHost
	case "deploy":
		hostIp = DeployEnvHost
	default:
		return containersDest, errors.New("Input parameter is error")
	}

	for _, apicontainer := range *containersSrc {
		var containerDest APIContainer
		containerDest.ContainerId = apicontainer.ID
		containerDest.Status = apicontainer.State
		if len(apicontainer.Names) != 0 {
			containerDest.Name = apicontainer.Names[0]
		} else {
			containerDest.Name = ""
		}

		containerDest.Labels = apicontainer.Labels
		containerDest.ImageName = apicontainer.Image

		if len(apicontainer.Ports) > 0 {
			containerDest.PublishAddr = hostIp + ":" + strconv.Itoa(int(apicontainer.Ports[0].PublicPort))
		} else {
			containerDest.PublishAddr = "unknow"
		}

		containersDest = append(containersDest, containerDest)
	}
	
	return containersDest, nil
} 
