package engine

import (
	"errors"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/dao"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
)

//map [NodeID]Node
type ProcNodes map[string]Node

//定义流程cache其结构为 map [ProcID]ProcNodes
var ProcCache = make(map[int]ProcNodes)

//从缓存中获取流程定义
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

//1、流程实例初始化 2、保存实例变量 返回:流程实例ID、开始节点
func InstanceInit(ProcessID int, BusinessID string, VariableJson string) (int, Node, error) {
	//获取流程定义(流程中所有node)
	nodes, err := GetProcCache(ProcessID)
	if err != nil {
		return 0, Node{}, err
	}

	//检查流程节点中的事件是否都已经注册
	for _, node := range nodes {
		err = CheckIfEventRegistered(node)
		if err != nil {
			return 0, Node{}, err
		}
	}

	//获取流程开始节点ID
	type result struct {
		Node_ID string
	}
	var r result
	_, err = dao.ExecSQL("SELECT node_id FROM `proc_execution` WHERE proc_id=? AND node_type=0", &r, ProcessID)
	if err != nil {
		return 0, Node{}, err
	}
	if r.Node_ID == "" {
		return 0, Node{}, fmt.Errorf("无法获取流程ID为%d的开始节点", ProcessID)
	}
	//获得开始节点
	StartNode := nodes[r.Node_ID]

	//1、在proc_inst表中生成一条记录
	//2、在proc_inst_variable表中记录流程实例的变量
	//3、返回proc_inst_id(流程实例ID)
	type result2 struct {
		ID    int
		Error string
	}
	var r2 result2

	_, err = dao.ExecSQL("call sp_proc_inst_init(?,?,?,?)", &r2, ProcessID, BusinessID, StartNode.NodeID, VariableJson)
	if err != nil {
		return 0, Node{}, err
	}

	if r2.Error != "" {
		return 0, Node{}, errors.New(r2.Error)
	}

	return r2.ID, StartNode, nil
}

//开始流程实例 返回流程实例ID
func InstanceStart(ProcessID int, BusinessID string, Comment string, VariablesJson string) (int, error) {
	//实例初始化
	InstanceID, StartNode, err := InstanceInit(ProcessID, BusinessID, VariablesJson)
	if err != nil {
		return 0, err
	}

	//开始节点处理
	err = StartNodeHandle(InstanceID, &StartNode, Comment, VariablesJson)
	if err != nil {
		return InstanceID, err
	}

	return InstanceID, nil
}

//撤销流程实例 参数说明:InstanceID 实例ID,Force 是否强制撤销，若为false,则只有流程回到发起人这里才能撤销
func InstanceRevoke(ProcessInstanceID int, Force bool) error {
	if !Force {
		//这段SQL判断是否当前Node就是开始Node
		sql := "SELECT a.id FROM proc_inst a " +
			"JOIN proc_execution b ON a.proc_id=b.proc_id AND a.current_node_id=b.node_id " +
			"WHERE a.id=? AND b.prev_node_id IS NULL LIMIT 1"
		var id int
		_, err := dao.ExecSQL(sql, &id, ProcessInstanceID)
		if err != nil {
			return err
		}
		if id == 0 {
			return errors.New("当前流程所在节点不是发起节点，无法撤销!")
		}
	}
	//调用EndNodeHandle,做数据清理归档
	err := EndNodeHandle(ProcessInstanceID, 2)
	return err
}
