package router

import (
	_ "easy-workflow/api/docs" // 导入swagger文档用的
	. "easy-workflow/api/service"
	. "easy-workflow/pkg/config"
	"easy-workflow/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter() *gin.Engine {
	gin.SetMode(Gin.GinMode)
	r := gin.New()

	//这里定义了一些中间件。中间件可以看作是拦截器，请求传入后，需要经过
	r.Use(gin.Logger())      //gin的默认log，默认输出是os.Stdout，即屏幕
	r.Use(log.MyGinlogger()) //自定义的日志记录,在方法执行完毕后记录在日志文件中
	r.Use(gin.Recovery())    //从任何panic中恢复，并在出现panic时返回http 500


	//只有在debug模式下才开启swagger
	if Gin.GinMode == "debug" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}


	r.POST("/process/def/save",ProcDef_Save)
	r.GET("/process/def/list",ProcDef_ListBySource)
    r.GET("/process/def/get",ProcDef_GetProcDefByID)

	r.POST("/process/inst/start",ProcInst_Start)

	r.POST("/process/task/pass",Task_Pass)
	r.POST("/process/task/reject",Task_Reject)
	r.GET("/process/task/todo",Task_ToDoList)
	r.GET("/process/task/finished",Task_FinishedList)



	return r
}
