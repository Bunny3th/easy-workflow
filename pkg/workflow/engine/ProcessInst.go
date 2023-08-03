package engine

import (
	"easy-workflow/pkg/dao"
	"easy-workflow/pkg/workflow/model/node"
	. "easy-workflow/pkg/workflow/model/variables"
	"log"

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

//返回流程实例ID
func ProcessStart(ProcessID int, Comment string, v Variables) (int, error) {

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

	log.Println("nodeid:", r.Node_ID)

	StartNode := nodes[r.Node_ID]

	log.Printf("获得开始节点-%+v", StartNode)

	//在proc_inst表中生成一条记录,并返回proc_inst_id(流程实例ID)
	type result2 struct {
		Proc_Inst_ID int
	}

	var r2 result2

	_, err = dao.ExecSQL("call sp_proc_inst_init(?,?)", &r2, ProcessID, StartNode.NodeID)
	if err != nil {
		return 0, err
	}

	return r2.Proc_Inst_ID, nil

}

//处理节点,如：生成task、进行条件判断、处理结束节点等，返回下一个节点
//func ProcessNode(ProcessInstanceID int, node node.Node) node.Node {
//	return nil
//}
