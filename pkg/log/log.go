package log

import (
	. "easy-workflow/pkg/config"
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

var logger *logrus.Logger
var logpath string

func init() {
	logger = logrus.New()
	logpath = App.LogPath
	logger.SetFormatter(&logrus.TextFormatter{})
}

//日志文件以日期为命名格式，在配置文件指定路径下生成
func SetLoggerOut(logpath string, source string) {
	logname := time.Now().Format("2006-01-02") + "-" + source //日志文件以日期为命名格式
	//生成或附加到日志文件
	logfile, err := os.OpenFile(logpath+logname+".log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0755)
	if err != nil { //生成日志文件出错，则进行标准输出
		logger.Out = os.Stdout
	} else { //输出到日志文件
		logger.Out = logfile
	}
}

//gin.Context.writer等价于ResponseWriter,实现一个自定义的结构体
type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

//实现ResponseWriter的Write方法
func (w CustomResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

//实现ResponseWriter的WriteString方法
func (w CustomResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

//GIN中间件 记录执行日志
func MyGinlogger() gin.HandlerFunc {
	SetLoggerOut(logpath, "API")
	logger.SetLevel(logrus.DebugLevel)

	return func(c *gin.Context) {
		crw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = crw
		start_time := time.Now()
		c.Next()                                                    //执行下一个HandlerFunc
		execution_time := time.Now().Sub(start_time).Milliseconds() //执行时间,毫秒
		ClientIP := c.ClientIP()                                    //用户IP
		ReqMethod := c.Request.Method                               //请求方法
		ReqURL := c.Request.URL                                     //请求路由
		StatusCode := c.Writer.Status()                             //返回状态码
		Response := crw.body.String()                               //返回body

		var resp string
		if StatusCode == 200 { //正常返回不记录Response,节约空间
			resp = "..."
		} else { //异常返回记录Response
			resp = Response
		}

		info_model := "ClientIP:%s - Method:%s - URL:%s - Code:%d - Duration:%d - Response:%s"
		logger.Errorf(info_model, ClientIP, ReqMethod, ReqURL, StatusCode, execution_time, resp)

	}
}

func MyMicroServicelogger(ClientIP string, Method string, Response interface{}, Err error) {
	SetLoggerOut(logpath, "MicroService")
	logger.SetLevel(logrus.DebugLevel)

	var resp interface{}
	if Err == nil { //如果没有报错，则不用记下详细的response，节约空间
		resp = "..."
	} else {
		resp = Response
	}

	info_model := "ClientIP:%s - Method:%s - Response:%s - Error:%s"
	logger.Infof(info_model, ClientIP, Method, resp, Err)

}
