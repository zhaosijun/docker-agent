package models

import (
	"time"
	"errors"
	"strconv"
	"fmt"

	"github.com/fsouza/go-dockerclient"
	"github.com/astaxie/beego/orm"
    _ "github.com/go-sql-driver/mysql"
    "github.com/astaxie/beego"
)

var (
	SQL1 string
	SQL2 string
	SQL3 string
)

func init() {
	// select id,name,registry_central,repository,tag,size,pushed from
	// (select a.imagerefence_id from
	// (select imagerefence_id from(select *from imagerefence_labels)a inner join (select *from label i where i.name='test1=test1')b on(a.label_id = b.id))a
	// inner join 
	// (select imagerefence_id from(select *from imagerefence_labels)a inner join (select *from label i where i.name='test2=test2')b on(a.label_id = b.id))b 
	// on(a.imagerefence_id = b.imagerefence_id))c
	// inner join
	// (select * from imagerefence)d 
	// on(c.imagerefence_id = d.id)
	
	//Todo: optimize m2m query sql
	SQL1 = "select id,name,registry_central,repository,tag,size,pushed from (%s)c inner join (select * from imagerefence)d on(c.imagerefence_id = d.id)"
	SQL2 = "select a.imagerefence_id from (%s)a inner join (%s)b on(a.imagerefence_id = b.imagerefence_id)"
	SQL3 = "select imagerefence_id from(select *from imagerefence_labels)a inner join (select *from label i where i.name='%s')b on(a.label_id = b.id)"

	user := beego.AppConfig.DefaultString("mysqluser", "root")
	passwd := beego.AppConfig.DefaultString("mysqlpass", "111")
	//mysqlurls := beego.AppConfig.DefaultString("mysqlurls", "tcp(192.168.88.23:3306)/docker_agent_db?charset=utf8&loc=Asia%2FShanghai")
	mysqlurls := beego.AppConfig.DefaultString("mysqlurls", "tcp(192.168.88.23:3306)/docker_agent_db?charset=utf8&loc=Asia%2FShanghai")

	link := user + ":" + passwd + "@" + mysqlurls
	beego.Info(link)
	orm.DefaultTimeLoc = time.UTC
	// set mysql engine innodb supporting mvcc and set transaction isolation level repeatable read
	// tolerate phantom read
 	orm.RegisterDriver("mysql", orm.DRMySQL)
    orm.RegisterDataBase("default", "mysql", link)
    orm.RegisterModel(new(Imagereference), new(Label))
}

type RspData struct {
	Status 	int  			`json:"status"`
	Message string 			`json:"msg"`
	Data    interface{}     `json:"data"`
}

type Imagereference struct {
    Id                int
    Name            string              `orm:"unique"`
    Alias            string              `orm:"unique"`  
    RegistryCentral string  			 
    Repository      string          	 
    Tag             string
    Size			int64   			 
    Pushed          time.Time           
    Labels          []*Label            `orm:"rel(m2m)"` 
}

func (i *Imagereference) TableIndex() [][]string {
    return [][]string{
        []string{"Name"},
    }
}

func (i *Imagereference) TableName() string {
    return "imagerefence"
}

type Label struct {
    Id                int
    Name            string               `orm:"unique"`  
    //Imagereferences  []*Imagereference   `orm:"reverse(many)"` 
}

func (l *Label) TableIndex() [][]string {
    return [][]string{
        []string{"Name"},
    }
}

func (l *Label) TableName() string {
    return "label"
}

// type DeleteImageOpts struct {
// 	ImageNames      []string
// }

type BuildImageOpts struct {
	ImageName		string
	ContextDir  	string
}

type PushImageOpts struct {
	ImageName		string
	Tag				string
}

type PullImageOpts struct {
	Repository 		string
	Tag				string
}

type DeployImageOpts struct {
	ImageName		string
	DeployOpts 		*DeployOpts
}

func GenerateRspData(status int, msg string, data interface{}) (rspdata RspData) {
	rspdata.Status = status
	rspdata.Message = msg
	rspdata.Data = data
	return rspdata
}

func parseImageDeployOptions(deployImageOpts *DeployImageOpts, deployContextOpts *DeployContextOpts) error {
	
	if deployImageOpts == nil || deployContextOpts == nil {
		return errors.New("Input parameter is null")
	}

	deployContextOpts.ImageName = deployImageOpts.ImageName

	err := ParseDeployOptions(deployImageOpts.DeployOpts, deployContextOpts)
	if err != nil {
		return err
	}

    return nil
}

func DeployByImage(client *docker.Client, deployImageOpts *DeployImageOpts) (string, error) {

	var deployContextOpts DeployContextOpts
	err := parseImageDeployOptions(deployImageOpts, &deployContextOpts)
	if err != nil {
		return "", err
	}

	container, err := createContainer(client, &deployContextOpts)
	if err != nil {
		return "", err
	}

	err = startContainer(client, container.ID, &deployContextOpts)
	if err != nil {
		return "", err
	}

	return container.ID, nil
}

func InspectImage(image *Imagereference) (err error) {

	o := orm.NewOrm()

	err = o.Begin() //repeatable read, Name is unique, engine of storage is InnoDB supportting mvcc
    if err != nil {
	    return err
	}

    defer func() {
		if err != nil {
			if errRollback := o.Rollback(); errRollback != nil {
				err = errRollback
			}
		}
	}()

    err = o.Read(image)
    if err != nil {
    	if err == orm.ErrNoRows {
    		beego.Info("ErrNoRows")
    		return errors.New("The image id do not exists: " + strconv.Itoa(image.Id))
    	}
    	return err
    }
    LogVarInfo("InspectImage: ", image)
    
    _, err = o.LoadRelated(image, "Labels")
    if err != nil {
        return err
    }

    return o.Commit()
}

func ListImage(limit int64, label string) (images []*Imagereference, err error) {

	o := orm.NewOrm()

	imageObj := Imagereference{}
    qs := o.QueryTable(imageObj.TableName())
    
    err = o.Begin() //repeatable read, Name is unique, engine of storage is InnoDB supportting mvcc
    if err != nil {
	    return nil, err
	}

    defer func() {
		if err != nil {
			if errRollback := o.Rollback(); errRollback != nil {
				err = errRollback
			}
		}
	}()

    if label != "" {
    	_, err := qs.Filter("labels__label__name", label).Limit(limit).All(&images)
    	if err != nil {
    		return nil, err
    	}
    } else {
    	_, err = qs.Limit(limit).All(&images)
    	if err != nil {
    		return nil, err
    	}
    }
    
    for _, image := range images {
	    _, err = o.LoadRelated(image, "Labels")
	    if err != nil {
	        return nil, err
	    }
    }

    return images, o.Commit()
}

func GenerateSqlWithLabels(labels []string) string {
	beego.Info("GenerateSqlWithLabels: ", labels)
	len := len(labels)
	switch len {
	case 1:
		return fmt.Sprintf(SQL3, labels[0])
	default:
		beego.Info("lenth of lables tail index is", len/2)
		sql3_1 := GenerateSqlWithLabels(labels[0:len/2])
		sql3_2 := GenerateSqlWithLabels(labels[len/2:len])
		return fmt.Sprintf(SQL2, sql3_1, sql3_2)
	}
}

func ListImageByLabels(limit int64, labels []string) (images []*Imagereference, err error) {
	
	var sql string
	o := orm.NewOrm()
	if len(labels) != 0 {
    	sqlWithLabels := GenerateSqlWithLabels(labels)
		sql = fmt.Sprintf(SQL1, sqlWithLabels)
		beego.Info(sql)
	}

	err = o.Begin() //repeatable read, Name is unique, engine of storage is InnoDB supportting mvcc
    if err != nil {
	    return nil, err
	}

    defer func() {
		if err != nil {
			if errRollback := o.Rollback(); errRollback != nil {
				err = errRollback
			}
		}
	}()

	if sql != "" {
		num, err := o.Raw(sql).QueryRows(&images)
		if err != nil {
		    return nil, err
		}
		beego.Info("query number of images is", num)
    } else {
    	imageObj := Imagereference{}
    	_, err = o.QueryTable(imageObj.TableName()).Limit(limit).All(&images)
    	if err != nil {
    		return nil, err
    	}
    }

    for _, image := range images {
	    _, err = o.LoadRelated(image, "Labels")
	    if err != nil {
	        return nil, err
	    }
    }

    return images, o.Commit()
}

func AddImageM2MLabel(image *Imagereference, labels []*Label) (err error) {
    
    var ptrimage *Imagereference
    if image != nil {
    	ptrimage = &Imagereference{Name: image.Name, 
							Tag: image.Tag, 
							RegistryCentral: image.RegistryCentral,
							Repository: image.Repository,
							Size: image.Size,
							Pushed: image.Pushed,
							Alias: image.Alias}
    }

    o := orm.NewOrm()
    err = o.Begin() //repeatable read, Name is unique, engine of storage is InnoDB supportting mvcc
    if err != nil {
	    return err
	}

    defer func() {
		if err != nil {
			if errRollback := o.Rollback(); errRollback != nil {
				err = errRollback
			}
		}
	}()

    if created, id, err := o.ReadOrCreate(image, "Name"); err == nil {
        if created {
            beego.Info("New Insert an object. id:", id)
        } else { 
            beego.Info("Get an object. id:", id)
            ptrimage.Id = image.Id
            if num, err := o.Update(ptrimage); err == nil {
				beego.Info("Number of records updated in database:", num)
			} else {
    			return err
			}
        }
    } else {
    	return err
    }

    m2m := o.QueryM2M(image, "Labels")
    _, err = m2m.Clear()
	if err != nil {
	    return err
	}

    for _, label := range labels {
	    if created, id, err := o.ReadOrCreate(label, "Name"); err == nil {
	        if created {
	            beego.Info("New Insert an object. id:", id)
	        } else {
	            beego.Info("Get an object. id:", id)
	        }
	    } else {
	        beego.Info("rollback")
	    	return err
		}

	    if !m2m.Exist(label) {
	        _, err = m2m.Add(label)
	        if err != nil {
	        	beego.Info("rollback")
	        	return err
	        }
	    }
	}

    return o.Commit()
}

func DeleteImage(image *Imagereference) (err error) {
    
    o := orm.NewOrm()
    
    err = o.Begin() //repeatable read, Name is unique, engine of storage is InnoDB supportting mvcc
    if err != nil {
	    return err
	}

    defer func() {
		if err != nil {
			if errRollback := o.Rollback(); errRollback != nil {
				err = errRollback
			}
		}
	}()

    err = o.Read(image)
    if err != nil {
    	if err == orm.ErrNoRows {
    		beego.Info("ErrNoRows")
    		return errors.New("The image id do not exists: " + strconv.Itoa(image.Id))
    	} else {
    		return err
    	}
    }
    LogVarInfo("DeleteImage: ", image)

    m2m := o.QueryM2M(image, "Labels")
    
    num, err := m2m.Clear()
    beego.Info("clear labels num:", num)
	if err != nil {
		beego.Info("rollback")
	    return err
	}

	num, err = o.Delete(image)
	beego.Info("Delete image num:", num)
	if err != nil {
		beego.Info("rollback")
	    return err
	}
  
    return o.Commit()
}

func DeleteImages(images []*Imagereference) (err error) {
    
    o := orm.NewOrm()
    
    err = o.Begin() //repeatable read, Name is unique, engine of storage is InnoDB supportting mvcc
    if err != nil {
	    return err
	}

    defer func() {
		if err != nil {
			if errRollback := o.Rollback(); errRollback != nil {
				err = errRollback
			}
		}
	}()

    for _, image := range images {
	    err = o.Read(image)
	    if err != nil {
	    	if err == orm.ErrNoRows {
	    		beego.Info("ErrNoRows")
	    		return errors.New("The image id do not exists: " + strconv.Itoa(image.Id))
	    	} else {
		    	return err
		    }
	    } 
	}

	for _, image := range images {
		m2m := o.QueryM2M(image, "Labels")
    
	    _, err := m2m.Clear()
		if err != nil {
			beego.Info("rollback")
		    return err
		}

		_, err = o.Delete(image)
		if err != nil {
			beego.Info("rollback")
		    return err
		}
	}
    
    return o.Commit()
}
