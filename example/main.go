package main

import (

	. "github.com/Bunny3th/easy-workflow/workflow/config"
	. "github.com/Bunny3th/easy-workflow/workflow/engine"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
	."github.com/Bunny3th/easy-workflow/workflow/web_api"
	"github.com/gin-gonic/gin"
	"log"
)

//这里创建了一个角色-用户的人员库，用来模拟数据库中存储的角色-用户对应关系
var RoleUser = make(map[string][]string)

type Event struct{}

func (e *Event) MyEvent_End(ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error {
	//可以做一些处理，比如通知流程开始人，节点到了哪个步骤
	processName,err:=GetProcessNameByInstanceID(ProcessInstanceID)
	if err!=nil{
		return err
	}
	log.Printf("--------流程[%s]节点[%s]结束-------",processName, CurrentNode.NodeName)
	return nil
}

func (e *Event) MyEvent_Notify(ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error {
	processName,err:=GetProcessNameByInstanceID(ProcessInstanceID)
	if err!=nil{
		return err
	}
	log.Printf("--------流程[%s]节点[%s]，通知节点中对应人员--------",processName, CurrentNode.NodeName)
	if CurrentNode.NodeType==EndNode{
		log.Printf("============================== 流程[%s]结束 ==============================",processName)
		variables,err:= ResolveVariables(ProcessInstanceID, []string{"$starter"})
		if err!=nil{
			return err
		}
		log.Printf("通知流程创建人%s,流程[%s]已完成",variables["$starter"],processName)

	}else{
		for _, user := range CurrentNode.UserIDs {
			log.Printf("通知用户[%s],抓紧去处理", user)
		}
	}
	return nil
}

//解析角色
func (e *Event) MyEvent_ResolveRoles(ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error {
	processName,err:=GetProcessNameByInstanceID(ProcessInstanceID)
	if err!=nil{
		return err
	}
	log.Printf("--------流程[%s]节点[%s],开始解析角色--------",processName, CurrentNode.NodeName)
	//把用户库中对应角色的用户全部放到CurrentNode.UserIDs中去
	for _, role := range CurrentNode.Roles {
		if users, ok := RoleUser[role]; ok {
			CurrentNode.UserIDs = append(CurrentNode.UserIDs, users...)
		}
	}
	return nil
}

func init() {
	//初始化人事数据
	RoleUser["主管"] = []string{"张经理"}
	RoleUser["人事经理"]=[]string{"人事老刘"}
	RoleUser["老板"] = []string{"李老板","老板娘"}
	RoleUser["副总"] = []string{"赵总", "钱总", "孙总"}
}

func DBConfig() {
	DBConnect.DBConnectString = "goeasy:sNd%sLDjd*12@tcp(172.16.18.18:3306)/easy_workflow?charset=utf8mb4&parseTime=True&loc=Local"
}


func main() {
	//----------------------------开启流程引擎----------------------------
	//StartWorkFlow(DBConfig, &Event{})
	StartWorkFlow(DBConfig,&Event{})

	//----------------------------开启web api----------------------------
	//本项目采用gin运行web api，首先生成一个gin.Engine
	//这里需要注意：如果你的业务系统也是使用gin框架，同时使用了swagger
	//你希望业务系统的swagger页面(以下简称“业务swagger”)与easy-workflow内置web api的swagger（以下简称“工作流swagger”）同时开启
	//必须做到：
	//1、业务swagger与工作流swagger必须使用同一个访问路由，即假如业务swagger访问路由是"/swagger/*any",则工作流swagger必须也是这个路由
	//2、由于同一个端口下不能起两个同名的路由地址，所以业务系统web api与easy-workflow内置web api必须使用不同的端口
	//我做过多种尝试，希望在同一个端口下，业务系统swagger使用路由“/X ”,工作流swagger使用路由"/Y"，但x、y只能存活一个，必然有一个不能访问
	//swagger-ui.css、swagger-ui-bundle.js、swagger-ui-standalone-preset.js这几个文件，猜测其原因，虽然module是两个，swag库是全局的
	//在此情况下，css文件等只能适配一个路由地址。以上是我的尝试与猜测，有知道的同学可以联系我告知正确的打开方式
	engine := gin.New()
	//这里定义中间件
	engine.Use(gin.Logger())      //gin的默认log，默认输出是os.Stdout，即屏幕
	engine.Use(gin.Recovery())    //从任何panic中恢复，并在出现panic时返回http 500
	StartWebApi(engine,"/process",true,"/swagger/*any",":8180")
}
