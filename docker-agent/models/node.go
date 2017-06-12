package models

import (
	"time"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"errors"

	"github.com/astaxie/beego"
	//"github.com/fsouza/go-dockerclient"
	//"github.com/google/cadvisor/info/v1"
	"github.com/google/cadvisor/info/v2"
	clientv2 "github.com/google/cadvisor/client/v2"
	"github.com/docker/docker/api/types/swarm"
	
)

var (
	CadvisorPort string
)

type NodeInfo struct {
	Id  			string
	HostName  		string
	Architecture   string
	OS			   string
	NumCores	   int
	CpuFrequency   uint64
	MemoryCapacity uint64
}

type MachineStats struct {
	Timestamp      	time.Time
	NetworkStats    NetworkStats
	MemoryStats  	MemoryStats 
	CPUStats    	CPUStats
}

type ContainerStatsInfo struct {
	ContainerSpec 	ContainerSpec
	ContainerStats  []*ContainerStats
}

type ContainerStats struct {
	Timestamp      	time.Time
	NetworkStats    NetworkStats
	MemoryStats  	MemoryStats 
	CPUStats    	CPUStats
}

type ContainerSpec struct {
	CPUSetCPUs	string		
	MemoryLimit uint64
}

type NetworkStats struct {
	InterfaceStats []InterfaceStats
}

type InterfaceStats struct {
	Name string
	RxBytes   uint64
	TxBytes   uint64
}

type MemoryStats struct {
	Usage    uint64
}

type CPUStats struct {
	CpuUsage CpuUsage
}

type CpuUsage struct {
	Total uint64
	PerCpu []uint64
}

func init() {
	CadvisorPort = beego.AppConfig.DefaultString("cadvisorport", "9094")
}

func ConvertNodeInfoFromDockerClientToFront (nodesSrc *[]swarm.Node, nodesDest *[]NodeInfo){

	if nodesSrc == nil || nodesDest == nil {
		return 
	}

	for index, _ := range *nodesSrc {
		nodeSrc := (*nodesSrc)[index]
		var nodeDest NodeInfo
		nodeDest.Id = nodeSrc.ID
		nodeDest.HostName = nodeSrc.Description.Hostname
		nodeDest.Architecture = nodeSrc.Description.Platform.Architecture
		nodeDest.OS = nodeSrc.Description.Platform.OS
		*nodesDest = append(*nodesDest, nodeDest)
	}
}

func ConvertNodeAttrFromDockerClientToFront (nodeSrc *v2.Attributes, nodeDest *NodeInfo){

	if nodeSrc == nil || nodeDest == nil {
		return 
	}

	nodeDest.NumCores = nodeSrc.NumCores
	nodeDest.CpuFrequency = nodeSrc.CpuFrequency 
	nodeDest.MemoryCapacity = nodeSrc.MemoryCapacity 
}

func GetNodeAttribute(hostName string) (*v2.Attributes, error) {
	baseUrl := fmt.Sprintf("http://%s:%s/", hostName, CadvisorPort)
	beego.Info("node cadvisor baseUrl: ", baseUrl)

	client, err := clientv2.NewClient(baseUrl)
	if err != nil {
		beego.Info("Fail to get node attributes due to ", err)
		return nil, err
	}

	nodeAttr, err := client.Attributes()
	if err != nil {
		beego.Info("Fail to get node attributes due to ", err)
		return nil, err
	}

	return nodeAttr, nil
}

func ConvertMachineStatsFromCadvisorClientToFront (statssSrc *[]v2.MachineStats, statssDest *[]MachineStats){

	if statssSrc == nil || statssDest == nil {
		return 
	}

	for index, _ := range *statssSrc {
		statsSrc := (*statssSrc)[index]
		var statsDest MachineStats
		statsDest.Timestamp = statsSrc.Timestamp
		for _, eth := range statsSrc.Network.Interfaces {
			interf := InterfaceStats{Name: eth.Name, RxBytes: eth.RxBytes, TxBytes: eth.TxBytes}
			statsDest.NetworkStats.InterfaceStats = append(statsDest.NetworkStats.InterfaceStats, interf)
		}
		statsDest.MemoryStats.Usage = statsSrc.Memory.Usage
		statsDest.CPUStats.CpuUsage.Total = statsSrc.Cpu.Usage.Total
		statsDest.CPUStats.CpuUsage.PerCpu = statsSrc.Cpu.Usage.PerCpu

		*statssDest = append(*statssDest, statsDest)
	}
}

func ConvertContainerStatsInfoFromDockerClientToFront (statsMapSrc *map[string]v2.ContainerInfo, statsDest *ContainerStatsInfo) error{

	if statsMapSrc == nil || statsDest == nil {
		return errors.New("stats is null")
	}

	var statssSrc []v2.ContainerInfo
	for _, stats := range *statsMapSrc {
		statssSrc = append(statssSrc, stats)
	}

	if len(statssSrc) != 1 {
		return errors.New("stats is null or error ")
	}

	statsSrc := statssSrc[0]

	statsDest.ContainerSpec.CPUSetCPUs = statsSrc.Spec.Cpu.Mask
	statsDest.ContainerSpec.MemoryLimit = statsSrc.Spec.Memory.Limit
	
	for index, _ := range statsSrc.Stats {
		stats := statsSrc.Stats[index]
		var containerStats ContainerStats
		containerStats.Timestamp = stats.Timestamp
		for _, eth := range stats.Network.Interfaces {
			interf := InterfaceStats{Name: eth.Name, RxBytes: eth.RxBytes, TxBytes: eth.TxBytes}
			containerStats.NetworkStats.InterfaceStats = append(containerStats.NetworkStats.InterfaceStats, interf)
		}
		containerStats.MemoryStats.Usage = stats.Memory.Usage
		containerStats.CPUStats.CpuUsage.Total = stats.Cpu.Usage.Total
		containerStats.CPUStats.CpuUsage.PerCpu = stats.Cpu.Usage.PerCpu

		statsDest.ContainerStats = append(statsDest.ContainerStats, &containerStats)
	}
	return nil
}

func GetMachineStats(url string) ([]v2.MachineStats, error) {
	var ret []v2.MachineStats
	err := httpGetJsonData(&ret, nil, url, "machine stats")
	return ret, err
}

func httpGetJsonData(data, postData interface{}, url, infoName string) error {
	body, err := httpGetResponse(postData, url, infoName)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, data); err != nil {
		err = fmt.Errorf("unable to unmarshal %q (Body: %q) from %q with error: %v", infoName, string(body), url, err)
		return err
	}
	return nil
}

func httpGetResponse(postData interface{}, urlPath, infoName string) ([]byte, error) {
	var resp *http.Response
	var err error

	if postData != nil {
		data, marshalErr := json.Marshal(postData)
		if marshalErr != nil {
			return nil, fmt.Errorf("unable to marshal data: %v", marshalErr)
		}
		resp, err = http.Post(urlPath, "application/json", bytes.NewBuffer(data))
	} else {
		resp, err = http.Get(urlPath)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to post %q to %q: %v", infoName, urlPath, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("received empty response for %q from %q", infoName, urlPath)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("unable to read all %q from %q: %v", infoName, urlPath, err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request %q failed with error: %q", urlPath, strings.TrimSpace(string(body)))
	}
	return body, nil
}
