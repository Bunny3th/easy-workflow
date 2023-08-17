package main

import (
	"easy-workflow/api/router"
	. "easy-workflow/workflow/config"
	."easy-workflow/workflow/engine"
	. "easy-workflow/workflow/model"
	"log"
)

// @title easy-workflow工作流引擎API
// @version 1.0.0
// @description 演示说明文档
// @contact.name go-swagger帮助文档
// @contact.url https://github.com/swaggo/swag/blob/master/README_zh-CN.md
// @host localhost:8180
// @BasePath /
func main() {


	StartWorkFlow(DBConfig,&Event{})


	router := router.NewRouter()
	router.Run(":8180")

	//这里演示如何编译成linux下可执行文件 -o 参数指定生成文件名
	//set GOARCH=amd64
	//set GOOS=linux
	//go build -o easy_workflow_linux main.go

	//如何使用swagger生成文档
	//一般在main包所在目录执行 swag init
	//但本项目中，swagger命令需要在api目录中加上-d参数执行，如下
	//swag init -d ./,../workflow/model

}


type Event struct{}

func (e *Event) Fuck(ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error {
	log.Println("fucking shit!!!!!")
	CurrentNode.UserIDs=[]string{"fucker"}
	//return errors.New("事件报错啦！！！")
	return nil
}

func DBConfig() {
	DBConnect.DBConnectString = "goeasy:sNd%sLDjd*12@tcp(172.16.18.18:3306)/easy_workflow?charset=utf8mb4&parseTime=True&loc=Local"
}