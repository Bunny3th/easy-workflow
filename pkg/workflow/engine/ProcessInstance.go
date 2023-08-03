package engine

import (
	"easy-workflow/pkg/dao"
	"easy-workflow/pkg/workflow/model/node"
	"encoding/json"
)

//map [NodeID]Node
type ProcNodes map[string]node.Node

//map [ProcID]ProcNodes
var ProcCache = make(map[int]ProcNodes)

func GetProcCache(ProcessID int) (ProcNodes, error) {
	if nodes, ok := ProcCache[ProcessID]; ok {
		return nodes, nil
	} else {
		nodes, err := GetProcessDefine(ProcessID)
		if err != nil {
			return nil, err
		}
		pn := make(ProcNodes)
		for _, n := range nodes {
			pn[n.NodeID] = n
		}
		ProcCache[ProcessID] = pn

	}
	return ProcCache[ProcessID], nil
}

//1、流程实例初始化 2、保存实例变量 返回:流程实例ID
func InstanceInit(ProcessID int,BusinessID string,Variables map[string]string) (int, error) {

	nodes, err := GetProcCache(ProcessID)
	if err != nil {
		return 0, err
	}

	//获取开始节点
	type result struct {
		Node_ID string
	}
	var r result
	_, err = dao.ExecSQL("SELECT node_id FROM `proc_execution` WHERE proc_id=? AND node_type=0", &r, ProcessID)
	if err != nil {
		return 0, err
	}
	StartNode := nodes[r.Node_ID]

	type kv struct {
		Key string
		Value string
	}

	var vars []kv

	for k,v:=range Variables{
		vars=append(vars,kv{Key:k,Value: v})
	}


	kvs,err:=json.Marshal(vars)
	if err!=nil{
		return 0,err
	}


	//在proc_inst表中生成一条记录,并返回proc_inst_id(流程实例ID)
	type result2 struct {
		ID int
		Error string
	}
	var r2 result2

	_, err = dao.ExecSQL("call sp_proc_inst_init(?,?,?,?)", &r2, ProcessID,BusinessID,StartNode.NodeID,string(kvs))
	if err != nil {
		return 0, err
	}

	return r2.ID, nil
}


func InstanceStart(){

}




