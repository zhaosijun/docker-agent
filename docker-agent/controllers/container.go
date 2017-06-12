package controllers

import (
	"docker-agent/models"
	//"fmt"
	"encoding/json"
	"bytes"
	 //"time"
	 //"net/http"

	"github.com/astaxie/beego"
	"github.com/fsouza/go-dockerclient"
	clientv2 "github.com/google/cadvisor/client/v2"
	"github.com/google/cadvisor/info/v2"
)

// Operations about Container
type ContainerController struct {
	beego.Controller
}


func (o *ContainerController) CreateContainer() {
// @Title CreateContainer CreateContainerOptions
// @Description Create a container 
// @Param	context		query 	string	true		"The URL for the specified docker"
// @Param	body			body 	models.CreateContainerOptions	true		"The Name of container to create"
// @Success 201 {object} models.ContainerObject
// @Failure 400 parameter error
// @Failure 404 no such container
// @Failure 409 container already exists 
// @Failure 500 server error
// @router /create [Post]
	endpoint, err := models.GetContextEndpoint(o.GetString("context"))
	if err != nil {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
		
		o.ServeJSON()
		return
	}

	client, err := docker.NewClient(endpoint)
	if err != nil {
		beego.Info(err)
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		o.ServeJSON()
		return
	}

	var createOptsDocker docker.CreateContainerOptions
	var createOpts models.CreateContainerOptions

	json.Unmarshal(o.Ctx.Input.RequestBody, &createOpts)
	models.LogVarInfo("CreateContainer createOpts: ", createOpts)
	models.ConvertStartOptsFromFrontToDockerClient(&createOpts, &createOptsDocker)
	models.LogVarInfo("CreateContainer createOptsDocker: ", createOptsDocker)

	container, err := client.CreateContainer(createOptsDocker)
	if err != nil {
		switch err {
		case docker.ErrNoSuchImage:
			o.Data["json"] = models.GenerateRspData(404, "no such container: "+err.Error(), nil)
		case docker.ErrContainerAlreadyExists:
			o.Data["json"] = models.GenerateRspData(409, "container already exists: "+err.Error(), nil)
		default:
			o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		}
	} else {
		o.Data["json"] = models.GenerateRspData(201, "no error", container)
	}
	o.ServeJSON()
}

// @Title StartContainer
// @Description start a container 
// @Param	context		query 	string	true		"The URL for the specified docker"
// @Param   containerid     path    string  true        "The id or name of container you want to start"
// @Success 204 {object} models.RspData
// @Failure 304 container already started
// @Failure 400 parameter error
// @Failure 404 no such container
// @Failure 500 server error
// @router /:containerid/start [Post]
func (o *ContainerController) StartContainer() {

	endpoint, err := models.GetContextEndpoint(o.GetString("context"))
	if err != nil {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
		
		o.ServeJSON()
		return
	}

	client, err := docker.NewClient(endpoint)
	if err != nil {
		beego.Info(err)
		
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		o.ServeJSON()
		return
	}

	var hostConfig docker.HostConfig
	
	containerId := o.GetString(":containerid")

	err = client.StartContainer(containerId, &hostConfig)
	if err != nil {
		switch err.(type) {
		case *docker.ContainerAlreadyRunning:
			
			o.Data["json"] = models.GenerateRspData(304, "container already started: "+err.Error(), nil)
		case *docker.NoSuchContainer:
			
			o.Data["json"] = models.GenerateRspData(404, "no such container: "+err.Error(), nil)
		default:
			
			o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		}
	} else {
		
		o.Data["json"] = models.GenerateRspData(204, "no error", nil)
	}
	o.ServeJSON()
}

// @Title StopContainer
// @Description stop a container 
// @Param	context		query 	string	true		"The hostIp for the specified docker"
// @Param   containerid     path    string  true        "The id or name of container you want to delete"
// @Param	t				query 	uint	false		"The number of seconds to wait before killing the container"
// @Success 204 {object} models.RspData
// @Failure 304 container already stopped
// @Failure 400 parameter error
// @Failure 404 no such container
// @Failure 500 server error
// @router /:containerid/stop [Post]
func (o *ContainerController) StopContainer() {

	var hostConfig docker.HostConfig
	endpoint, err := models.GetContextEndpoint(o.GetString("context"))
	if err != nil {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
		
		o.ServeJSON()
		return
	}

	client, err := docker.NewClient(endpoint)
	if err != nil {
		beego.Info(err)
		
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		o.ServeJSON()
		return
	}

	json.Unmarshal(o.Ctx.Input.RequestBody, &hostConfig)
	beego.Info(hostConfig)

	t, err := o.GetInt("t")
	if err != nil {
		t = 10
	}

	containerId := o.GetString(":containerid")

	err = client.StopContainer(containerId, uint(t))
	if err != nil {
		switch err.(type) {
		case *docker.ContainerNotRunning:
			
			o.Data["json"] = models.GenerateRspData(304, "container already started: "+err.Error(), nil)
		case *docker.NoSuchContainer:
			
			o.Data["json"] = models.GenerateRspData(404, "no such container: "+err.Error(), nil)
		default:
			
			o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		}
	} else {
		
		o.Data["json"] = models.GenerateRspData(204, "no error", nil)
	}
	o.ServeJSON()
}

// @Title RemoveContainer
// @Description stop a container 
// @Param	context		query 	string	true		"The hostIp for the specified docker"
// @Param	containerid		path	int 	true		"The Id of Container to Delete"
// @Param	force			query	bool 	false		"If the container is running, kill it before removing it. Default to true"
// @Success 204 no error
// @Failure 400 parameter error
// @Failure 404 no such container
// @Failure 500 server error
// @router /:containerid [Delete]
func (o *ContainerController) RemoveContainer() {

	var removeOpts docker.RemoveContainerOptions 
	endpoint, err := models.GetContextEndpoint(o.GetString("context"))
	if err != nil {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
		
		o.ServeJSON()
		return
	}

	client, err := docker.NewClient(endpoint)
	if err != nil {
		beego.Info(err)
		
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		o.ServeJSON()
		return
	}

	removeOpts.Force, err = o.GetBool("force")
	if err != nil {
		removeOpts.Force = true
	}
	removeOpts.ID = o.GetString(":containerid")
	
	err = client.RemoveContainer(removeOpts)
	if err != nil {
		switch err.(type) {
		case *docker.NoSuchContainer:
			
			o.Data["json"] = models.GenerateRspData(404, "no such container: "+err.Error(), nil)
		default:
			
			o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		}
	} else {
		
		o.Data["json"] = models.GenerateRspData(204, "no error", nil)
	}
	o.ServeJSON()
}

// @Title InspectContainer
// @Description Inspect a container info
// @Param	context		query 	string	true		"The URL for the specified docker"
// @Param	containerid		path 	string	true		"The id or name for the specified container"
// @Success 200 {object} models.RspData
// @Failure 400 parameter error
// @Failure 404 no such container
// @Failure 500 server error
// @router /:containerid/inspect [Get]
func (o *ContainerController) InspectContainer() {

	endpoint, err := models.GetContextEndpoint(o.GetString("context"))
	if err != nil {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
		
		o.ServeJSON()
		return
	}

	client, err := docker.NewClient(endpoint)
	if err != nil {
		beego.Info(err)
		
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		o.ServeJSON()
		return
	}

	containerId := o.GetString(":containerid")
	beego.Info("containerId:", containerId)

	container, err := client.InspectContainer(containerId)
	if err != nil {
		switch err.(type) {
		case *docker.NoSuchContainer:
			
			o.Data["json"] = models.GenerateRspData(404, "no such container: "+err.Error(), nil)
		default:
			
			o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		}
	} else {
		
		var con models.Container
		models.ConvertContainerFromDockerClientToFront(container, &con, o.GetString("context"))
		o.Data["json"] = models.GenerateRspData(200, "no error", con)
	}
	o.ServeJSON()
}

// @Title LogsContainer
// @Description Get logs of a container
// @Param	context		query 	string	true		"The URL for the specified docker"
// @Param	containerid		path 	string	true		"The id or name for the specified container"
// @Param	tail			query	string	false		"Output specified number of lines at the end of logs: all or <number>. Default all"
// @Success 200 {object} models.RspData
// @Failure 400 parameter error
// @Failure 404 no such container
// @Failure 500 server error
// @router /:containerid/logs [Get]
func (o *ContainerController) LogsContainer() {

	endpoint, err := models.GetContextEndpoint(o.GetString("context"))
	if err != nil {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
		
		o.ServeJSON()
		return
	}

	client, err := docker.NewClient(endpoint)
	if err != nil {
		beego.Info(err)
		
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		o.ServeJSON()
		return
	}

	containerId := o.GetString(":containerid")
	beego.Info("containerId:", containerId)

	tail := o.GetString("tail")
	if tail == "" {
		tail = "20"
	}
	var buf bytes.Buffer
	err = client.Logs(docker.LogsOptions{Container: containerId,
													 OutputStream: &buf,
													 Tail: tail,
													 Stdout: true,
													 Stderr: true,
													 Timestamps: true})
	if err != nil {
		switch err.(type) {
		case *docker.NoSuchContainer:
			
			o.Data["json"] = models.GenerateRspData(404, "no such container: "+err.Error(), nil)
		default:
			
			o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		}
	} else {
		
		o.Data["json"] = models.GenerateRspData(200, "no error", map[string]string{"containerLogs": buf.String()})
	}
	o.ServeJSON()
}

// @Title StatsContainer
// @Description Stats Container 
// @Param	context		query 	string	true		"The URL for the specified docker"
// @Param	containerid		path 	string	true		"The id or name for the specified container"
// @Param	count			query	int		    false "Number of stats samples to be reported. Default is 1"
// @Success 200 {object} models.RspData
// @Failure 400 parameter error
// @Failure 500 server error
// @router /:containerid/stats [Get]
func (o *ContainerController) StatsContainer() {

	containerId := o.GetString(":containerid")
	if containerId == "" {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: no container id", nil)
		o.ServeJSON()
		return
	}

	baseUrl, err:= models.GetContextEndpoint(o.GetString("context") + "/cadvisor")
	if err != nil {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
		o.ServeJSON()
		return
	}

	count, err := o.GetInt("count")
	if err != nil {
		count = 1
	}

	beego.Info("node cadvisor baseUrl: ", baseUrl)
	client, err := clientv2.NewClient(baseUrl)
	if err != nil {
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		o.ServeJSON()
		return
	}

	containsMap, err := client.Stats(containerId, &v2.RequestOptions{IdType: "docker",
																Count: count,
																Recursive: false})
	if err != nil {
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
	} else {
		var containerStats models.ContainerStatsInfo
		beego.Info("containsMap: ", containsMap)
		err := models.ConvertContainerStatsInfoFromDockerClientToFront(&containsMap, &containerStats)
		beego.Info("containerStats: ", containerStats)
		if err != nil {
			o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		} else {
			o.Data["json"] = models.GenerateRspData(200, "no error", &containerStats)
		}
	}
	o.ServeJSON()

	
	// endpoint, err := models.GetContextEndpoint(o.GetString("context"))
	// if err != nil {
	// 	o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
		
	// 	o.ServeJSON()
	// 	return
	// }

	// containerId := o.GetString(":containerid")
	// beego.Info("containerId:", containerId)

	// dockerUrl := fmt.Sprintf("%s/v1.24/containers/%s/stats?stream=false", endpoint, containerId)
	// response, err := http.Get(dockerUrl)

	// if err != nil {
	// 	switch err.(type) {
	// 	case *docker.NoSuchContainer:
			
	// 		o.Data["json"] = models.GenerateRspData(404, "no such container: "+err.Error(), nil)
	// 	default:
	// 		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
	// 	}
	// } else {
	// 	decoder := json.NewDecoder(response.Body)
	// 	stats := new(docker.Stats)
	// 	err := decoder.Decode(stats)
	// 	if err != nil {
	// 		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
	// 	} else {
	// 		var containerStat models.ContainerStats
	// 		models.ConvertContainerStatFromDockerClientToFront(stats, &containerStat)
	// 		o.Data["json"] = models.GenerateRspData(200, "no error", &containerStat)
	// 	}
	// }
	// o.ServeJSON()
}

// @Title ListContainer
// @Description List containers 
// @Param	context		query 	string	true		"The hostIp for the specified docker"
// @Param	all			    query	bool	false		"Show all containers, this Default to false"
// @Param	labels			query	[]string	false "Filter containers by labels, i.e. "key=value" of a container label"
// @Param	status			query	[]string	false "Filter containers by status, i.e. status = (created restarting running paused exited dead)
// @Success 200 {object} models.RspData
// @Failure 400 parameter error
// @Failure 500 server error
// @router /list [Get]
func (o *ContainerController) ListContainer() {

	endpoint, err := models.GetContextEndpoint(o.GetString("context"))
	if err != nil {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
		
		o.ServeJSON()
		return
	}

	client, err := docker.NewClient(endpoint)
	if err != nil {
		beego.Info(err)
		
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		o.ServeJSON()
		return
	}

	all, err := o.GetBool("all")
	labels := o.GetStrings("labels")
	status := o.GetStrings("status")
	models.LogVarInfo("labels: ", labels)

	var filters map[string][]string
	filters = make(map[string][]string)
	filters["label"] = labels
	filters["status"] = status
	
	containers, err := client.ListContainers(docker.ListContainersOptions{Filters: filters, All: all})

	if err != nil {
		
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
	} else {
		
		cons, _ := models.ConvertAPIContainerFromDockerClientToFront(&containers, o.GetString("context"))
		o.Data["json"] = models.GenerateRspData(200, "no error", cons)
	}
	o.ServeJSON()
}

