package controllers

import (
	"docker-agent/models"
	"fmt"
	//"encoding/json"
	//"bytes"
	 //"time"
	 //"net/http"

	"github.com/astaxie/beego"
	"github.com/fsouza/go-dockerclient"
	//"github.com/docker/docker/api/types/swarm"
	//"github.com/google/cadvisor/info/v2"
	//clientv2 "github.com/google/cadvisor/client/v2"
)

// Operations about Node
type NodeController struct {
	beego.Controller
}

// @Title ListNodes
// @Description List Nodes 
// @Param	labels			query	[]string	false "Filter nodes by labels, i.e. "key=value" of a node label"
// @Success 200 {object} models.RspData
// @Failure 500 server error
// @router /list [Get]
func (o *NodeController) ListNodes() {

	endpoint, _ := models.GetContextEndpoint("cluster")

	client, err := docker.NewClient(endpoint)
	if err != nil {
		beego.Info(err)
		
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		o.ServeJSON()
		return
	}

	labels := o.GetStrings("labels")
	models.LogVarInfo("labels: ", labels)

	var filters map[string][]string
	filters = make(map[string][]string)
	filters["label"] = labels
	
	nodes, err := client.ListNodes(docker.ListNodesOptions{Filters: filters})

	if err != nil {
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
	} else {
		var nodesInfo []models.NodeInfo
		models.ConvertNodeInfoFromDockerClientToFront(&nodes, &nodesInfo)
		beego.Info("nodes base Info : ", nodesInfo)
		for index, _ := range nodesInfo {
			nodeAttr, err := models.GetNodeAttribute(nodesInfo[index].HostName)
			if err == nil {
				models.ConvertNodeAttrFromDockerClientToFront(nodeAttr, &nodesInfo[index])
				beego.Info("node Info with attributes: ", nodesInfo[index])
			} else {
				beego.Info("GetNodeAttribute with error: ", err)
			}
		}
		o.Data["json"] = models.GenerateRspData(200, "no error", nodesInfo)
	}
	o.ServeJSON()
}

// @Title StatsNode
// @Description Stats Node 
// @Param	hostname		query	string	    true  "Specify which node to stat"
// @Param	count			query	int		    false "Number of stats samples to be reported. Default is 1"
// @Success 200 {object} models.RspData
// @Failure 400 parameter error
// @Failure 500 server error
// @router /stats [Get]
func (o *NodeController) StatsNode() {

	hostName := o.GetString("hostname")
	models.LogVarInfo("hostName: ", hostName)
	if hostName == "" {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: no hostname", nil)
		o.ServeJSON()
		return
	}

	count, err := o.GetInt("count")
	if err != nil {
		count = 1
	}

	baseUrl := fmt.Sprintf("http://%s:%s/api/v2.1/machinestats", hostName, models.CadvisorPort)

	u := fmt.Sprintf("%s?count=%d", baseUrl, count)
	beego.Info("node cadvisor baseUrl: ", u)
	stats, err := models.GetMachineStats(u)
	if err != nil {
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
	} else {
		var statsFront []models.MachineStats
		models.ConvertMachineStatsFromCadvisorClientToFront(&stats, &statsFront)
		o.Data["json"] = models.GenerateRspData(200, "no error", statsFront)
	}
	o.ServeJSON()
}

