package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"],
		beego.ControllerComments{
			Method: "StartContainer",
			Router: `/:containerid/start`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"],
		beego.ControllerComments{
			Method: "StopContainer",
			Router: `/:containerid/stop`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"],
		beego.ControllerComments{
			Method: "RemoveContainer",
			Router: `/:containerid`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"],
		beego.ControllerComments{
			Method: "InspectContainer",
			Router: `/:containerid/inspect`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"],
		beego.ControllerComments{
			Method: "LogsContainer",
			Router: `/:containerid/logs`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"],
		beego.ControllerComments{
			Method: "StatsContainer",
			Router: `/:containerid/stats`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:ContainerController"],
		beego.ControllerComments{
			Method: "ListContainer",
			Router: `/list`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:ImageController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:ImageController"],
		beego.ControllerComments{
			Method: "Deploy",
			Router: `/deploy`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:ImageController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:ImageController"],
		beego.ControllerComments{
			Method: "List",
			Router: `/list`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:ImageController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:ImageController"],
		beego.ControllerComments{
			Method: "InspectImage",
			Router: `/:id`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:ImageController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:ImageController"],
		beego.ControllerComments{
			Method: "DeleteOneImage",
			Router: `/:id`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:ImageController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:ImageController"],
		beego.ControllerComments{
			Method: "DeleteImages",
			Router: `/delete`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:JarController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:JarController"],
		beego.ControllerComments{
			Method: "Deploy",
			Router: `/deploy`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:JarController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:JarController"],
		beego.ControllerComments{
			Method: "Push",
			Router: `/push`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:NodeController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:NodeController"],
		beego.ControllerComments{
			Method: "ListNodes",
			Router: `/list`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["docker-agent/controllers:NodeController"] = append(beego.GlobalControllerRouter["docker-agent/controllers:NodeController"],
		beego.ControllerComments{
			Method: "StatsNode",
			Router: `/stats`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

}
