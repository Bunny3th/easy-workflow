package database

import "github.com/Bunny3th/easy-workflow/workflow/dao"

//初始化数据库表
func DatabaseInitialize(){
	dao.DB.AutoMigrate(&ProcDef{})
	dao.DB.AutoMigrate(&HistProcDef{})
	dao.DB.AutoMigrate(&ProcInst{})
	dao.DB.AutoMigrate(&HistProcInst{})
	dao.DB.AutoMigrate(&Task{})
	dao.DB.AutoMigrate(&HistTask{})
	dao.DB.AutoMigrate(&TaskComment{})
	dao.DB.AutoMigrate(&HistTaskComment{})
	dao.DB.AutoMigrate(&ProcExecution{})
	dao.DB.AutoMigrate(&HistProcExecution{})
	dao.DB.AutoMigrate(&ProcInstVariable{})
	dao.DB.AutoMigrate(&HistProcInstVariable{})

}