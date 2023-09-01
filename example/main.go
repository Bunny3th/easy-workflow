package main

import (
	"easy-workflow/example/router"
	. "easy-workflow/workflow/config"
	. "easy-workflow/workflow/engine"
	. "easy-workflow/workflow/model"
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

func DBConfig() {
	DBConnect.DBConnectString = "goeasy:sNd%sLDjd*12@tcp(172.16.18.18:3306)/easy_workflow?charset=utf8mb4&parseTime=True&loc=Local"
}

func init() {
	//初始化人事数据
	RoleUser["人事主管"] = []string{"张经理"}
	RoleUser["老板"] = []string{"李老板","老板娘"}
	RoleUser["副总"] = []string{"赵总", "钱总", "孙总"}
}

// @title easy-workflow工作流引擎API
// @version 1.0.0
// @description 演示说明文档
// @contact.name go-swagger帮助文档
// @contact.url https://github.com/swaggo/swag/blob/master/README_zh-CN.md
// @host localhost:8180
// @BasePath /
func main() {

	//开启流程引擎
	StartWorkFlow(DBConfig, &Event{})

	//开启web api
	router := router.NewRouter()
	router.Run(":8180")

	//如何使用swagger生成文档
	//一般在main包所在目录执行 swag init
	//但本项目中，swagger命令需要在api目录中加上-d参数执行，如下
	//swag init -d ./,../workflow/model
}
