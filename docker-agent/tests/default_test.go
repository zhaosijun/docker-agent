package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"runtime"
	"path/filepath"
	"io/ioutil"
	"bytes"
	"fmt"
	"encoding/json"

	"docker-agent/models"
	_ "docker-agent/routers"


	"github.com/astaxie/beego"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/fsouza/go-dockerclient"
)

func init() {
	_, file, _, _ := runtime.Caller(1)
	apppath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".." + string(filepath.Separator))))
	beego.TestBeegoInit(apppath)
}

func TestOperateContainer(t *testing.T) {
	var containers []docker.APIContainers

	//List containers by labels
	hostIp := beego.AppConfig.DefaultString("hostIp", "0.0.0.0")
	url := fmt.Sprintf("/v1/containers/list?dockerhost=%s&labels=env%stest", hostIp, "%3D")
	
	r, _ := http.NewRequest("GET", url, nil)
	beego.Info(r)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	json.Unmarshal(w.Body.Bytes(), &containers)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())
	Convey("Subject: List containers\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//Delete Containers
	for _, con := range containers {
		url := fmt.Sprintf("/v1/containers/%s?dockerhost=%s&force=true", con.ID, hostIp)
		r, _ := http.NewRequest("DELETE", url, nil)
		w := httptest.NewRecorder()
		beego.BeeApp.Handlers.ServeHTTP(w, r)

		beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

		Convey("Subject: Delete Container\n", t, func() {
		        Convey("Status Code Should Be 204", func() {
		                So(w.Code, ShouldEqual, 204)
		        })
		        Convey("The Result Should Not Be Empty", func() {
		                So(w.Body.Len(), ShouldBeGreaterThan, 0)
		        })
		})
	}
	
	//Deploy container by jar
	// type ConID struct {
	// 	containerId string
	// }
	var conId map[string]string

	bytess, err := ioutil.ReadFile("./tests/json/deploy_jar.json")
	if err != nil {
		beego.Trace(err)
	}
	buf := bytes.NewBuffer(bytess)

	hostIp = beego.AppConfig.DefaultString("hostIp", "0.0.0.0")
	url = fmt.Sprintf("/v1/jar/deploy?dockerhost=%s", hostIp)

	r, _ = http.NewRequest("POST", url, buf)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	json.Unmarshal(w.Body.Bytes(), &conId)
	beego.Trace(conId)
	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Deploy container by jar\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//List containers by labels
	hostIp = beego.AppConfig.DefaultString("hostIp", "0.0.0.0")
	url = fmt.Sprintf("/v1/containers/list?dockerhost=%s&labels=env%stest", hostIp, "%3D")
	beego.Info(url)
	r, _ = http.NewRequest("GET", url, nil)
	beego.Info(r)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	json.Unmarshal(w.Body.Bytes(), &containers)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())
	Convey("Subject: List containers\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//Inspect Container
	url = fmt.Sprintf("/v1/containers/%s/inspect?dockerhost=%s", conId["containerId"], hostIp)
	beego.Trace(url)
	r, _ = http.NewRequest("GET", url, nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Inspect Container\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//Stop Container
	url = fmt.Sprintf("/v1/containers/%s/stop?dockerhost=%s&t=2", conId["containerId"], hostIp)
	r, _ = http.NewRequest("POST", url, nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Stop Container\n", t, func() {
	        Convey("Status Code Should Be 204", func() {
	                So(w.Code, ShouldEqual, 204)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//Start Container
	url = fmt.Sprintf("/v1/containers/%s/start?dockerhost=%s", conId["containerId"], hostIp)
	r, _ = http.NewRequest("POST", url, nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Start Container\n", t, func() {
	        Convey("Status Code Should Be 204", func() {
	                So(w.Code, ShouldEqual, 204)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//Delete Container
	url = fmt.Sprintf("/v1/containers/%s?dockerhost=%s&force=true", conId["containerId"], hostIp)
	r, _ = http.NewRequest("DELETE", url, nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Delete Container\n", t, func() {
	        Convey("Status Code Should Be 204", func() {
	                So(w.Code, ShouldEqual, 204)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//Deploy container by image
	bytess, err = ioutil.ReadFile("./tests/json/deploy_image.json")
	if err != nil {
		beego.Trace(err)
	}
	buf = bytes.NewBuffer(bytess)

	hostIp = beego.AppConfig.DefaultString("hostIp", "0.0.0.0")
	url = fmt.Sprintf("/v1/images/deploy?dockerhost=%s", hostIp)

	r, _ = http.NewRequest("POST", url, buf)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	json.Unmarshal(w.Body.Bytes(), &conId)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Deploy container by image\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//Delete Container
	url = fmt.Sprintf("/v1/containers/%s?dockerhost=%s&force=true", conId["containerId"], hostIp)
	r, _ = http.NewRequest("DELETE", url, nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Delete Container\n", t, func() {
	        Convey("Status Code Should Be 204", func() {
	                So(w.Code, ShouldEqual, 204)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

}

func TestOperateImage(t *testing.T) {

	//build image1 by jar and push 
	bytess, err := ioutil.ReadFile("./tests/json/build_push_jar_1.json")
	if err != nil {
		beego.Trace(err)
	}
	buf := bytes.NewBuffer(bytess)

	hostIp := beego.AppConfig.DefaultString("hostIp", "0.0.0.0")
	url := fmt.Sprintf("/v1/jar/push?dockerhost=%s", hostIp)

	r, _ := http.NewRequest("POST", url, buf)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Build image by jar and Push\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//build image2 by jar and push
	bytess, err = ioutil.ReadFile("./tests/json/build_push_jar_2.json")
	if err != nil {
		beego.Trace(err)
	}
	buf = bytes.NewBuffer(bytess)

	r, _ = http.NewRequest("POST", url, buf)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Build image by jar and Push\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//build image3 by jar and push
	bytess, err = ioutil.ReadFile("./tests/json/build_push_jar_3.json")
	if err != nil {
		beego.Trace(err)
	}
	buf = bytes.NewBuffer(bytess)

	r, _ = http.NewRequest("POST", url, buf)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Build image by jar and Push\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//List images
	var images []models.Imagereference
	r, _ = http.NewRequest("GET", "/v1/images/list?labels=env%3Dtest", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	json.Unmarshal(w.Body.Bytes(), &images)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Test Station Endpoint\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//Inpsect Image
	url = fmt.Sprintf("/v1/images/%d", images[0].Id)
	r, _ = http.NewRequest("GET", url, nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Inpsect Image\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//Delete one image
	url = fmt.Sprintf("/v1/images/%d", images[0].Id)
	r, _ = http.NewRequest("DELETE", url, nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Delete one image\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})

	//Delete images
	url = fmt.Sprintf("/v1/images/delete?Ids=%d%s%d", images[1].Id, "%2C", images[2].Id)
	r, _ = http.NewRequest("GET", url, nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestGet", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Delete images\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})
}
