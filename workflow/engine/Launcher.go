package engine

import (
	"easy-workflow/workflow/dao"
	. "easy-workflow/workflow/event"
	"log"
)

type DataBaseConfigurator func()

//传入参数1、数据库配置方法，方法签名func(Params ...string) 2、动态参数，事件函数所在的struct，可传多个
func StartWorkFlow(DBConfigurator DataBaseConfigurator, EventStructs ...any) {
	//配置数据库
	DBConfigurator()

	//数据库连接初始化
	dao.DBInit()

	//注册事件函数
	for _, s := range EventStructs {
		RegisterEvents(s)
	}
	log.Println("================== easy workflow 启动成功 ================== ")
	log.Print("                                                                                                            \n  ██████▓                       ▒                                                              █████        \n  ▒█████░                    ████░   ▒░                                                        ░████        \n   ████▓        ░███▓███▒   ███████      ░██▓▒▓███       ▓████▓░███  ░███████▓▒    ░▓████████   ████░ █████░\n   ████▓       ▓███▒  ████  ░████░       ████▒▒▒░         ▒█████▓▓▒ ████▓  ▒████  ▓████   ▒▓▓   ████░ ▒██░  \n  ░████▓  ░██▓ ████▒         ████▒         ░▒▒▓███▓       ▒████     ████▒  ░████  ████▓   ░     ████▒▓███▓  \n ░███████████▓  ▒███▓▒███▒   ▓█████      ▓██▓▒▒███░      ▓██████░    ▒████████▓    ▒████▓███▒  █████░  ▓███░\n                                                                                                            \n\n")

}
