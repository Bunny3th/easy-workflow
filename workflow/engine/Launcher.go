package engine

import (
	"github.com/Bunny3th/easy-workflow/workflow/dao"
	"log"
)

type DataBaseConfigurator func()

//传入参数1、数据库配置方法，方法签名func(Params ...string) 2、动态参数，事件函数所关联的struct，可传多个
func StartWorkFlow(DBConfigurator DataBaseConfigurator, EventStructs ...any) {
	//配置数据库
	DBConfigurator()

	//数据库连接初始化
	dao.DBInit()

	if len(EventStructs) != 0 {
		//注册事件函数
		for _, s := range EventStructs {
			if s != nil {
				RegisterEvents(s)
			}
		}
	}

	log.Println("========================== easy workflow 启动成功  create by 兔老三 ========================== ")
	log.Print("\n███████╗ █████╗ ███████╗██╗   ██╗    ████████╗ ██████╗      ██████╗  ██████╗ \n██╔════╝██╔══██╗██╔════╝╚██╗ ██╔╝    ╚══██╔══╝██╔═══██╗    ██╔════╝ ██╔═══██╗\n█████╗  ███████║███████╗ ╚████╔╝        ██║   ██║   ██║    ██║  ███╗██║   ██║\n██╔══╝  ██╔══██║╚════██║  ╚██╔╝         ██║   ██║   ██║    ██║   ██║██║   ██║\n███████╗██║  ██║███████║   ██║          ██║   ╚██████╔╝    ╚██████╔╝╚██████╔╝\n╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝          ╚═╝    ╚═════╝      ╚═════╝  ╚═════╝ \n                                                                             \n")
}
