package controllers

import (
	"docker-agent/models"
	"fmt"
	"encoding/json"

	"github.com/astaxie/beego"
)

// Operations about Jar
type JarController struct {
	beego.Controller
}

// @Title Deploy
// @Description deploy Jar or war package
// @Param	context		query 	string	true		"The URL for the specified docker"
// @Param	body		    body 	models.DeployJarOpts	true		"The deployment options for service"
// @Success 200 {object} models.RspData
// @Failure 400 parameter error
// @Failure 500 server error
// @router /deploy [Post]
func (o *JarController) Deploy() {

	var deployJarOpts models.DeployJarOpts

	json.Unmarshal(o.Ctx.Input.RequestBody, &deployJarOpts)
	models.LogVarInfo("deployJarOpts: ", deployJarOpts)
	
	endpoint, err := models.GetContextEndpoint(o.GetString("context"))
	if err != nil {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
		
		o.ServeJSON()
		return
	}
	
	containerId, err := models.DeployByJar(endpoint, &deployJarOpts)
	if err != nil { 
		
		o.Data["json"] =  models.GenerateRspData(500, "server error: "+err.Error(), nil)
	} else {

		o.Data["json"] =  models.GenerateRspData(200, "no error", map[string]string{"containerId": containerId})
	}

	o.ServeJSON()
}

// @Title Push
// @Description build image by Jar or war package and push to registry
// @Param	body		    body 	models.PushAndBuildJarOpts	true		"The deployment options for service"
// @Success 200 no error
// @Failure 500 server error
// @router /push [Post]
func (o *JarController) Push() {

	var pushAndBuildOpts models.PushAndBuildJarOpts
	json.Unmarshal(o.Ctx.Input.RequestBody, &pushAndBuildOpts)

	
	endpoint := fmt.Sprintf("tcp://%s:5555", models.DevEnvHost)
	
	err := models.BuildAndPushImageByJar(endpoint, &pushAndBuildOpts)
	if err != nil { 
		
		o.Data["json"] =  models.GenerateRspData(500, "server error: "+err.Error(), nil)
	} else {
		o.Data["json"] =  models.GenerateRspData(200, "no error", nil)
	}

	o.ServeJSON()
}
