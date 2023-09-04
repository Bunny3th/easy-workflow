package router

import (
	_ "github.com/Bunny3th/easy-workflow/workflow/web_api/docs" // 导入swagger文档用的
	. "github.com/Bunny3th/easy-workflow/workflow/web_api/service"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(engine *gin.Engine,GinMode string) *gin.Engine {
	gin.SetMode(GinMode)

	//只有在debug模式下才开启swagger
	if GinMode == "debug" {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	engine.POST("/process/def/save",ProcDef_Save)
	engine.GET("/process/def/list",ProcDef_ListBySource)
	engine.GET("/process/def/get",ProcDef_GetProcDefByID)

	engine.POST("/process/inst/start",ProcInst_Start)
	engine.POST("/process/inst/revoke",ProcInst_Revoke)
	engine.GET("/process/inst/task_history",ProcInst_TaskHistory)

	engine.POST("/process/task/pass",Task_Pass)
	engine.POST("/process/task/pass/directly",Task_Pass_DirectlyToWhoRejectedMe)

	engine.POST("/process/task/reject",Task_Reject)
	engine.POST("/process/task/reject/free",Task_FreeRejectToUpstreamNode)
	engine.GET("/process/task/todo",Task_ToDoList)
	engine.GET("/process/task/finished",Task_FinishedList)
	engine.GET("/process/task/upstream",Task_UpstreamNodeList)

	return engine
}
