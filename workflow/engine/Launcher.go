package engine

import (
	"log"
)

type DataBaseConfigurator func()

//传入参数
//1、DBConnConfigurator:数据库连接配置方法，方法签名func()
//2、ignoreEventError:是否忽略事件错误
//3、EventStructs:动态参数，事件函数所关联的struct，可传多个
func StartWorkFlow(DBConnConfigurator DataBaseConfigurator, ignoreEventError bool, EventStructs ...any) {
	//配置数据库连接选项
	DBConnConfigurator()

	//数据库连接
	err := DBConnect()
	if err != nil {
		log.Fatalln("easy workflow 数据库连接失败，错误:", err)
	}

	//初始化数据库表
	err = DatabaseInitialize()
	if err != nil {
		log.Fatalln("easy workflow 初始化数据表失败，错误:", err)
	}

	//是否忽略事件错误
	IgnoreEventError = ignoreEventError

	//注册事件函数
	for _, s := range EventStructs {
		if s != nil {
			RegisterEvents(s)
		}
	}

	log.Println("========================== easy workflow 启动成功  create by 兔老三 ========================== ")
	log.Print("\n███████╗ █████╗ ███████╗██╗   ██╗    ████████╗ ██████╗      ██████╗  ██████╗ \n██╔════╝██╔══██╗██╔════╝╚██╗ ██╔╝    ╚══██╔══╝██╔═══██╗    ██╔════╝ ██╔═══██╗\n█████╗  ███████║███████╗ ╚████╔╝        ██║   ██║   ██║    ██║  ███╗██║   ██║\n██╔══╝  ██╔══██║╚════██║  ╚██╔╝         ██║   ██║   ██║    ██║   ██║██║   ██║\n███████╗██║  ██║███████║   ██║          ██║   ╚██████╔╝    ╚██████╔╝╚██████╔╝\n╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝          ╚═╝    ╚═════╝      ╚═════╝  ╚═════╝ \n                                                                             \n")
}
