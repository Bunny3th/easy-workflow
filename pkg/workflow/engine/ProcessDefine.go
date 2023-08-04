package engine

import (
	"easy-workflow/pkg/dao"
	. "easy-workflow/pkg/workflow/model/datatables"
	. "easy-workflow/pkg/workflow/model/node"
	"easy-workflow/pkg/workflow/util"
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

//这里要写一个func，检查解析后的node结构，比如是否只有一个开始和结束节点

//流程定义保存,返回 流程ID、error
func ProcessSave(ProcessName string, Resource string, CreateUserID string, Source string) (int,error) {
	if ProcessName == "" || Source == "" || CreateUserID == "" {
		return 0,errors.New("流程名称、来源、创建人ID不能为空")
	}

	nodes, err := ProcessParse(Resource)
	if err != nil {
		return 0,err
	}



	execution, err := Nodes2Execution(nodes)
	if err != nil {
		return 0,err
	}

	type result struct {
		ID int
		Error string
	}

	var r result
	_, err = dao.ExecSQL("CALL sp_proc_def_save(?,?,?,?,?)", &r, ProcessName, Resource, execution, CreateUserID, Source)

	if err != nil || r.Error != "" {
		return 0,errors.New(err.Error() + r.Error)
	}

	//移除cache中对应流程ID的内容
	delete(ProcCache,r.ID)

	return r.ID,nil
}

//将Node转为可被数据库表记录的执行步骤。因为节点的PrevNodeID可能是n个，则在数据库表中需要存n行
func Nodes2Execution(nodes []Node) (string, error) {
	var executions []Execution
	for _, n := range nodes {
		if len(n.PrevNodeIDs) <= 1 {
			var PrevNodeID string
			if len(n.PrevNodeIDs) == 0 {
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
		} else {
			for _, pre := range n.PrevNodeIDs {
				executions = append(executions, Execution{
					NodeID:     n.NodeID,
					NodeName:   n.NodeName,
					PrevNodeID: pre,
					NodeType:   int(n.NodeType),
					IsCosigned: int(n.IsCosigned),
				})
			}
		}

	}

	json, err := json.Marshal(executions)
	if err != nil {
		return "", err
	}
	return string(json), nil
}

//获取流程ID
func GetProcessIDByProcessName(ProcessName string, Source string) (int, error) {
	var ID int
	_, err := dao.ExecSQL("SELECT id FROM proc_def where name=? and source=?", &ID, ProcessName, Source)

	if err != nil {
		return 0, err
	}

	return ID, nil
}

//获取流程定义
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
