package engine

import . "github.com/Bunny3th/easy-workflow/workflow/database"

//初始化数据库表
func DatabaseInitialize() error {
	err:=DB.AutoMigrate(&ProcDef{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(&HistProcDef{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(&ProcInst{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(&HistProcInst{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(&ProcTask{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(&HistProcTask{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(&ProcExecution{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(&HistProcExecution{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(&ProcInstVariable{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(&HistProcInstVariable{})
	if err != nil {
		return err
	}
	return nil
}
