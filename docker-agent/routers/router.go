// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"docker-agent/controllers"

	"github.com/astaxie/beego"
        "github.com/astaxie/beego/plugins/cors"
)

func init() {
        beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
                AllowAllOrigins:  true,
                AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS","HEAD","PATCH"},
                AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type","Access-Token"},
                ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
                AllowCredentials: true,
            }))

	ns := beego.NewNamespace("/deployservice/v1",
		beego.NSNamespace("/images",
                        beego.NSInclude(
                                &controllers.ImageController{},
                        ),
                ),
                beego.NSNamespace("/containers",
                        beego.NSInclude(
                                &controllers.ContainerController{},
                        ),
                ),
                beego.NSNamespace("/jars",
                        beego.NSInclude(
                                &controllers.JarController{},
                        ),
                ),
                beego.NSNamespace("/nodes",
                        beego.NSInclude(
                                &controllers.NodeController{},
                        ),
                ),
	)
	beego.AddNamespace(ns)
}
