package controllers

import (
	"docker-agent/models"
	"fmt"
	"bytes"
	"encoding/json"
	"strings"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/fsouza/go-dockerclient"
	//"github.com/astaxie/beego/orm"
)

// Operations about Image
type ImageController struct {
	beego.Controller
}

// @Title Deploy
// @Description Deploy service By Image
// @Param	context		query 	string	true		"The URL for the specified docker"
// @Param	body		        body 	models.DeployImageOpts	true		"The deployment options for service"
// @Success 200 {object} models.RspData
// @Failure 500 server error
// @router /deploy [Post]
func (o *ImageController) Deploy() {

	var deployOpts models.DeployImageOpts

	json.Unmarshal(o.Ctx.Input.RequestBody, &deployOpts)
	
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
	
	containerId, err := models.DeployByImage(client, &deployOpts)
	if err != nil { 
		
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
	} else {
		
		o.Data["json"] = models.GenerateRspData(200, "no error", map[string]string{"containerId": containerId})
	}

	o.ServeJSON()
}


func (o *ImageController) Build() {
// @Title Build
// @Description build Images
// @Param	context		query 	string	true		"The URL for the specified docker"
// @Param	body		    body 	models.BuildImageOpts	true		"The Options of Pushing Image"
// @Success 200 no error
// @Failure 500 server error
// @router /build [Post]

	var buildOpts models.BuildImageOpts
	var buf bytes.Buffer
	json.Unmarshal(o.Ctx.Input.RequestBody, &buildOpts)

	endpoint := fmt.Sprintf("tcp://%s:5555", o.GetString("context"))
	client, err := docker.NewClient(endpoint)
	if err != nil {
		beego.Info(err)
		
		o.Data["json"] = err.Error()
		o.ServeJSON()
		return
	}

	err = client.BuildImage(docker.BuildImageOptions{
								Name: buildOpts.ImageName,
								ContextDir: buildOpts.ContextDir,
								OutputStream: &buf})
	if err != nil { 
		
		o.Data["json"] = buf.String() + err.Error()
	} else {
		
		o.Data["json"] = buf.String()
	}

	o.ServeJSON()
}


func (o *ImageController) Push() {
// @Title Push
// @Description Push a Image
// @Param	context		query 	string	true		"The URL for the specified docker"
// @Param	body		    body 	models.PushImageOpts	true		"The Options of Pushing Image"
// @Success 200 no error
// @Failure 500 server error
// @router /push [Post]
	var pushOpts models.PushImageOpts
	var buf bytes.Buffer

	json.Unmarshal(o.Ctx.Input.RequestBody, &pushOpts)
	beego.Info(pushOpts.ImageName, pushOpts.Tag)

	endpoint := fmt.Sprintf("tcp://%s:5555", o.GetString("context"))
	client, err := docker.NewClient(endpoint)
	if err != nil {
		beego.Info(err)
		
		o.Data["json"] = err.Error()
		o.ServeJSON()
		return
	}

	err = client.PushImage(docker.PushImageOptions{Name: pushOpts.ImageName, 
									Tag: pushOpts.Tag,
									OutputStream: &buf}, docker.AuthConfiguration{})
	if err != nil {
		o.Data["json"] = buf.String() + err.Error()
		
	} else {
		o.Data["json"] = buf.String()
		
	}
	o.ServeJSON()
}


func (o *ImageController) Pull() {
// @Title Pull
// @Description pull a image from the registry
// @Param	context		query 	string	true		"The URL for the specified docker"
// @Param	body			body 	models.PullImageOpts  "The Options of Pulling Image"
// @Success 200 no error
// @Failure 500 server error
// @router /pull [Post]
	var pullOpts models.PullImageOpts
	var buf bytes.Buffer

	json.Unmarshal(o.Ctx.Input.RequestBody, &pullOpts)
	beego.Info(pullOpts.Repository, pullOpts.Tag)

	endpoint := fmt.Sprintf("tcp://%s:5555", o.GetString("context"))
	client, err := docker.NewClient(endpoint)
	if err != nil {
		beego.Info(err)
		
		o.Data["json"] = err.Error()
		o.ServeJSON()
		return
	}

	err = client.PullImage(docker.PullImageOptions {Repository : pullOpts.Repository,
									Tag: pullOpts.Tag,
									OutputStream: &buf}, docker.AuthConfiguration{})
	if err != nil { 
		o.Data["json"] = buf.String() + err.Error()
		
	} else {
		o.Data["json"] = buf.String()
		
	}
	o.ServeJSON()
}

// @Title List
// @Description List images by label
// @Param	labels			query	[]string	false		"Filter Images, i.e. "key=value" of a Image label"
// @Param	limit			query	int 	false		"The number of Listing Images, default to 10"
// @Success 200 {object} models.RspData 
// @Failure 500 server error
// @router /list [Get]
func (o *ImageController) List() {

	var limit int64
	labels := o.GetStrings("labels")
	limit, err := o.GetInt64("limit")
	if err != nil {
		limit = 10
	}

	//images, err := models.ListImage(limit, label)
	images, err := models.ListImageByLabels(limit, labels)
	if err != nil {
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		
	} else {
		o.Data["json"] = models.GenerateRspData(200, "no error", images)
		
	}
	o.ServeJSON()
}

// @Title InspectImage
// @Description Inspect image 
// @Param	Id			path	int 	true		"The Id of Image to Inspect"
// @Success 200 {object} models.RspData
// @Failure 400 parameter error
// @Failure 500 server error
// @router /:id [Get]
func (o *ImageController) InspectImage() {
	
	imageId, err := o.GetInt(":id")
	if err != nil {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
		
		o.ServeJSON()
		return
	}
	beego.Info("InspectImage:", imageId)
	
	image := models.Imagereference{Id: imageId}
	err = models.InspectImage(&image)
	if err != nil {
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		
	} else {
		o.Data["json"] = models.GenerateRspData(200, "no error", image)
		
	}
	o.ServeJSON()
}

// @Title DeleteOneImage
// @Description Delete one image 
// @Param	Id			path	int 	true		"The Id of Image to Delete"
// @Success 200 no error
// @Failure 400 parameter error
// @Failure 500 server error
// @router /:id [Delete]
func (o *ImageController) DeleteOneImage() {
	
	imageId, err := o.GetInt(":id")
	if err != nil {
		o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
		
		o.ServeJSON()
		return
	}
	beego.Info("DeleteOneImage:", imageId)
	
	image := models.Imagereference{Id: imageId}
	err = models.DeleteImage(&image)
	if err != nil {
		o.Data["json"] = models.GenerateRspData(500, "server error: "+err.Error(), nil)
		
	} else {
		o.Data["json"] = models.GenerateRspData(200, "no error", nil)
		
	}
	o.ServeJSON()
}

// @Title DeleteImages
// @Description Delete images 
// @Param	ids			query	string	true  "The Ids of Images to Delete i.e. 1,2,3"
// @Success 200 no error 
// @Failure 400 parameter error
// @Failure 500 server error
// @router /delete [Get]
func (o *ImageController) DeleteImages() {

	imageIdsStr := o.GetString("ids")
	beego.Info("DeleteImages:", imageIdsStr)
	imageIdStrs := strings.Split(imageIdsStr, ",")

	var images []*models.Imagereference
	for _, idstr := range imageIdStrs {
		id, err := strconv.Atoi(idstr)
		if err != nil {
			o.Data["json"] = models.GenerateRspData(400, "parameter error: "+err.Error(), nil)
			
			o.ServeJSON()
			return
		}
		image := models.Imagereference{Id: id}
		images = append(images, &image)
	}

	err := models.DeleteImages(images)
	if err != nil {
		o.Data["json"] =  models.GenerateRspData(500, "server error: "+err.Error(), nil)
		
	} else {
		o.Data["json"] =  models.GenerateRspData(200, "no error", nil)
		
	}
	o.ServeJSON()
}


