package engine

import (
	"easy-workflow/workflow/dao"
	. "easy-workflow/workflow/model"
	"easy-workflow/workflow/util"
	"encoding/json"
	"errors"
)

//流程定义解析(json->struct)
func ProcessParse(Resource string) ([]Node, error) {
	var nodes []Node
	err := util.Json2Struct(Resource, &nodes)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

//待办这里要写一个func，检查解析后的node结构，比如是否只有一个开始和结束节点

//流程定义保存,返回 流程ID、error
func ProcessSave(ProcessName string, Resource string, CreateUserID string, Source string) (int, error) {
	if ProcessName == "" || Source == "" || CreateUserID == "" || Resource == "" {
		return 0, errors.New("流程名称、资源定义、来源、创建人ID不能为空")
	}

	//解析传入的json，获得node列表
	nodes, err := ProcessParse(Resource)
	if err != nil {
		return 0, err
	}

	//解析node之间的关系，转为json，以便存入数据库
	execution, err := Nodes2Execution(nodes)
	if err != nil {
		return 0, err
	}

	type result struct {
		ID    int
		Error string
	}
	//存入数据库
	var r result
	_, err = dao.ExecSQL("CALL sp_proc_def_save(?,?,?,?,?)", &r, ProcessName, Resource, execution, CreateUserID, Source)
	if err != nil {
		return 0, err
	}

	if r.Error != "" {
		return 0, errors.New(r.Error)
	}

	//移除cache中对应流程ID的内容(如果有)
	delete(ProcCache, r.ID)

	return r.ID, nil
}

//将Node转为可被数据库表记录的执行步骤。节点的PrevNodeID可能是n个，则在数据库表中需要存n行
func Nodes2Execution(nodes []Node) (string, error) {
	var executions []Execution
	for _, n := range nodes {
		if len(n.PrevNodeIDs) <= 1 { //上级节点数<=1的情况下
			var PrevNodeID string
			if len(n.PrevNodeIDs) == 0 { //开始节点没有上级
				PrevNodeID = ""
			} else {
				PrevNodeID = n.PrevNodeIDs[0]
			}
			executions = append(executions, Execution{
				NodeID:     n.NodeID,
				NodeName:   n.NodeName,
				PrevNodeID: PrevNodeID,
				NodeType:   int(n.NodeType),
				IsCosigned: int(n.IsCosigned),
			})
		} else { //上级节点>1的情况下，则每一个上级节点都要生成一行
			for _, prev := range n.PrevNodeIDs {
				executions = append(executions, Execution{
					NodeID:     n.NodeID,
					NodeName:   n.NodeName,
					PrevNodeID: prev,
					NodeType:   int(n.NodeType),
					IsCosigned: int(n.IsCosigned),
				})
			}
		}
	}
	//转为json
	json, err := json.Marshal(executions)
	if err != nil {
		return "", err
	}
	return string(json), nil
}

//获取流程ID by 流程名、来源
func GetProcessIDByProcessName(ProcessName string, Source string) (int, error) {
	var ID int
	_, err := dao.ExecSQL("SELECT id FROM proc_def where name=? and source=?", &ID, ProcessName, Source)
	if err != nil {
		return 0, err
	}

	return ID, nil
}

//获取流程ID by 流程实例ID
func GetProcessIDByInstanceID(ProcessInstanceID int) (int, error) {
	var ID int
	_, err := dao.ExecSQL("SELECT proc_id FROM `proc_inst` WHERE id=?", &ID, ProcessInstanceID)
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
	_, err := dao.ExecSQL(sql, &r, ProcessInstanceID)
	if err != nil {
		return "", err
	}

	return r.Name, nil
}

//获取流程定义（返回流程中所有节点） by 流程ID
func GetProcessDefine(ProcessID int) ([]Node, error) {
	type result struct {
		Resource string
	}
	var r result
	_, err := dao.ExecSQL("SELECT resource FROM proc_def WHERE ID=?", &r, ProcessID)
	if err != nil {
		return nil, err
	}

	return ProcessParse(r.Resource)
}

//获得某个source下所有流程信息
func GetProcessList(Source string) ([]ProcessDefine, error) {
	var ProcessDefine []ProcessDefine
	_, err := dao.ExecSQL("select * from proc_def where source=?", &ProcessDefine, Source)
	return ProcessDefine, err
}
