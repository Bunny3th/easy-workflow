package engine

import (
	"errors"
	"github.com/Bunny3th/easy-workflow/workflow/database"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
	"gorm.io/gorm"
)

//流程定义解析(json->struct)
func ProcessParse(Resource string) (Process, error) {
	var process Process
	err := Json2Struct(Resource, &process)
	if err != nil {
		return Process{}, err
	}
	return process, nil
}

//todo:这里要写一个func，检查解析后的node结构，比如是否只有一个开始和结束节点

//流程定义保存,返回 流程ID、error
func ProcessSave(Resource string, CreateUserID string) (int, error) {
	//解析传入的json，获得process数据结构
	process, err := ProcessParse(Resource)
	if err != nil {
		return 0, err
	}

	if process.ProcessName == "" || process.Source == "" || CreateUserID == "" {
		return 0, errors.New("流程名称、来源、创建人ID不能为空")
	}

	//首先判断此工作流是否已定义
	ProcID, Version, err := GetProcessIDByProcessName(nil, process.ProcessName, process.Source)
	if err != nil {
		return 0, err
	}

	//开启事务
	tx := DB.Begin()
	//判断工作流是否已经定义
	if ProcID != 0 { //已有老版本
		//需要将老版本移到历史表中
		result := tx.Exec("INSERT INTO hist_proc_def(proc_id,NAME,`version`,resource,user_id,source,create_time)\n        "+
			"SELECT id,NAME,`version`,resource,user_id,source,create_time\n"+
			"FROM proc_def WHERE NAME=? AND source=?;", process.ProcessName, process.Source)
		if result.Error != nil {
			tx.Rollback()
			return 0, result.Error
		}
		//而后更新现有定义
		result = tx.Model(&database.ProcDef{}).
			Where("name=? AND source=?", process.ProcessName, process.Source).
			Updates(database.ProcDef{Version: Version + 1, Resource: Resource, UserID: CreateUserID, CreatTime: database.LTime.Now()})
		if result.Error != nil {
			tx.Rollback()
			return 0, result.Error
		}
	} else {
		//若没有老版本，则直接插入
		procDef := database.ProcDef{Name: process.ProcessName, Resource: Resource, UserID: CreateUserID, Source: process.Source}
		result := tx.Create(&procDef)
		if result.Error != nil {
			tx.Rollback()
			return 0, result.Error
		}
	}
	//重新获得流程ID、版本号,此时因为是在事务中，所以需要传入tx
	ProcID, Version, err = GetProcessIDByProcessName(tx, process.ProcessName, process.Source)
	if err != nil {
		return 0, err
	}

	//将proc_execution表对应数据移到历史表中
	result := tx.Exec("INSERT INTO hist_proc_execution(proc_id,proc_version,node_id,node_name,\n"+
		"prev_node_id,node_type,is_cosigned,create_time)\n"+
		"SELECT proc_id,proc_version,node_id,node_name,\n"+
		"prev_node_id,node_type,is_cosigned,create_time\n"+
		"FROM proc_execution WHERE proc_id=?;", ProcID)
	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	}

	//而后删除proc_execution表对应数据
	result = tx.Where("proc_id=?", ProcID).Delete(&database.ProcExecution{})
	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	}

	//解析node之间的关系，流程节点执行关系定义记录
	Execution := nodes2Execution(ProcID, Version, process.Nodes)

	//将Execution定义插入proc_execution表
	result = tx.Create(&Execution)
	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	}

	//事务提交
	tx.Commit()

	//移除cache中对应流程ID的内容(如果有)
	delete(ProcCache, ProcID)

	return ProcID, nil
}

//将Node转为可被数据库表记录的执行步骤。节点的PrevNodeID可能是n个，则在数据库表中需要存n行
func nodes2Execution(ProcID int, ProcVersion int, nodes []Node) []database.ProcExecution {
	var executions []database.ProcExecution
	for _, n := range nodes {
		if len(n.PrevNodeIDs) <= 1 { //上级节点数<=1的情况下
			var PrevNodeID string
			if len(n.PrevNodeIDs) == 0 { //开始节点没有上级
				PrevNodeID = ""
			} else {
				PrevNodeID = n.PrevNodeIDs[0]
			}
			executions = append(executions, database.ProcExecution{
				ProcID:      ProcID,
				ProcVersion: ProcVersion,
				NodeID:      n.NodeID,
				NodeName:    n.NodeName,
				PrevNodeID:  PrevNodeID,
				NodeType:    int(n.NodeType),
				IsCosigned:  int(n.IsCosigned),
				CreateTime:  database.LTime.Now(),
			})
		} else { //上级节点>1的情况下，则每一个上级节点都要生成一行
			for _, prev := range n.PrevNodeIDs {
				executions = append(executions, database.ProcExecution{
					ProcID:      ProcID,
					ProcVersion: ProcVersion,
					NodeID:      n.NodeID,
					NodeName:    n.NodeName,
					PrevNodeID:  prev,
					NodeType:    int(n.NodeType),
					IsCosigned:  int(n.IsCosigned),
					CreateTime:  database.LTime.Now(),
				})
			}
		}
	}
	return executions
}

//获取流程ID、Version by 流程名、来源
//设置传入参数db，是因为此函数可能在事务中执行。当在事务中执行时，需要传入对应的*gorm.DB
//若db传参为nil，则默认使用当前默认的*gorm.DB
func GetProcessIDByProcessName(db *gorm.DB, ProcessName string, Source string) (int, int, error) {
	type Result struct {
		ID      int
		Version int
	}
	var result Result

	var d *gorm.DB
	if db == nil {
		d = DB
	} else {
		d = db
	}

	r := d.Raw("SELECT id,version FROM proc_def where name=? and source=?", ProcessName, Source).Scan(&result)
	if r.Error != nil {
		return 0, 0, r.Error
	}

	return result.ID, result.Version, nil
}

//获取流程ID by 流程实例ID
func GetProcessIDByInstanceID(ProcessInstanceID int) (int, error) {
	var ID int
	_, err := ExecSQL("SELECT proc_id FROM `proc_inst` WHERE id=?", &ID, ProcessInstanceID)
	if err != nil {
		return 0, err
	}

	return ID, nil
}

//获取流程名称 by 流程实例ID
func GetProcessNameByInstanceID(ProcessInstanceID int) (string, error) {
	sql := "SELECT b.name FROM proc_inst a JOIN proc_def b ON a.proc_id=b.id WHERE a.id=?"
	type result struct {
		Name string
	}
	var r result
	_, err := ExecSQL(sql, &r, ProcessInstanceID)
	if err != nil {
		return "", err
	}

	return r.Name, nil
}

//获取流程定义 by 流程ID
func GetProcessDefine(ProcessID int) (Process, error) {
	type result struct {
		Resource string
	}
	var r result
	_, err := ExecSQL("SELECT resource FROM proc_def WHERE ID=?", &r, ProcessID)
	if err != nil {
		return Process{}, err
	}

	//如果数据库中没有找到ProcessID对应的流程,则直接报错
	if r.Resource==""{
		return Process{},errors.New("未找到对应流程定义")
	}

	return ProcessParse(r.Resource)
}

//获得某个source下所有流程信息
func GetProcessList(Source string) ([]database.ProcDef, error) {
	var ProcessDefine []database.ProcDef
	_, err := ExecSQL("select * from proc_def where source=?", &ProcessDefine, Source)
	return ProcessDefine, err
}
