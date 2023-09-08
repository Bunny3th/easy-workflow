package engine

import (
	"errors"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/dao"
	"github.com/Bunny3th/easy-workflow/workflow/database"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
	. "github.com/Bunny3th/easy-workflow/workflow/util"
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

	//-----------------------------------开始处理数据-----------------------------------
	//1、在proc_inst表中生成一条记录
	//2、在proc_inst_variable表中记录流程实例的变量
	//3、返回proc_inst_id(流程实例ID)

	//获取流程定义信息
	var procDef database.ProcDef
	dao.DB.Where("id=?", ProcessID).First(&procDef)

	//开启事务
	tx := dao.DB.Begin()
	//在实例表中生成一条数据
	procInst := database.ProcInst{ProcID: ProcessID, ProcVersion: procDef.Version, BusinessID: BusinessID, CurrentNodeID: StartNode.NodeID}
	re := tx.Create(&procInst)
	if re.Error != nil {
		tx.Rollback()
		return 0, StartNode, re.Error
	}

	//保存流程变量
	err = InstanceVariablesSave(procInst.ID, VariableJson)
	if err != nil {
		tx.Rollback()
		return 0, StartNode, err
	}

	//关闭事务
	tx.Commit()

	return procInst.ID, StartNode, nil
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

//流程实例变量存入数据库
func InstanceVariablesSave(ProcessInstanceID int, VariablesJson string) error {
	type Variable struct {
		Key   string
		Value string
	}
	//获取变量数组
	var variables []Variable
	Json2Struct(VariablesJson, &variables)

	tx := dao.DB.Begin()
	for _, v := range variables {
		var ProcInstVariable database.ProcInstVariable
		result := tx.Raw("SELECT * FROM proc_inst_variable WHERE proc_inst_id=? AND `key`=? ORDER BY id LIMIT 1",
			ProcessInstanceID, v.Key).Scan(&ProcInstVariable)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
		if ProcInstVariable.ID == 0 { //说明数据库中无此数据
			//插入
			result := tx.Create(&database.ProcInstVariable{ProcInstID: ProcessInstanceID, Key: v.Key, Value: v.Value})
			if result.Error != nil {
				tx.Rollback()
				return result.Error
			}
		} else { //数据库中已有数据
			//更新
			result := tx.Model(&database.ProcInstVariable{}).
				Where("proc_inst_id=? and `key`=?", ProcessInstanceID, v.Key).Update("value", v.Value)
			if result.Error != nil {
				tx.Rollback()
				return result.Error
			}
		}
	}

	tx.Commit()

	return nil
}
