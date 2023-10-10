package engine

import ."github.com/Bunny3th/easy-workflow/workflow/database"

//初始化数据库表
func DatabaseInitialize(){
	DB.AutoMigrate(&ProcDef{})
	DB.AutoMigrate(&HistProcDef{})
	DB.AutoMigrate(&ProcInst{})
	DB.AutoMigrate(&HistProcInst{})
	DB.AutoMigrate(&ProcTask{})
	DB.AutoMigrate(&HistProcTask{})
	DB.AutoMigrate(&ProcExecution{})
	DB.AutoMigrate(&HistProcExecution{})
	DB.AutoMigrate(&ProcInstVariable{})
	DB.AutoMigrate(&HistProcInstVariable{})

}