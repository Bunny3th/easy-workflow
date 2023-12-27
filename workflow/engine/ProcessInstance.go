package engine

import (
	"errors"
	"fmt"
	"github.com/Bunny3th/easy-workflow/workflow/database"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
)

//map [NodeID]Node
type ProcNodes map[string]Node

//定义流程cache其结构为 map [ProcID]ProcNodes
var ProcCache = make(map[int]ProcNodes)

//从缓存中获取流程节点定义
func GetProcCache(ProcessID int) (ProcNodes, error) {
	if nodes, ok := ProcCache[ProcessID]; ok {
		return nodes, nil
	} else {
		process, err := GetProcessDefine(ProcessID)
		if err != nil {
			return nil, err
		}
		pn := make(ProcNodes)
		for _, n := range process.Nodes {
			pn[n.NodeID] = n
		}
		ProcCache[ProcessID] = pn
	}
	return ProcCache[ProcessID], nil
}

//1、流程实例初始化 2、保存实例变量 返回:流程实例ID、开始节点
func instanceInit(ProcessID int, BusinessID string, VariableJson string) (int, Node, error) {
	//获取流程定义(流程中所有node)
	nodes, err := GetProcCache(ProcessID)
	if err != nil {
		return 0, Node{}, err
	}

	//检查流程节点中的事件是否都已经注册
	err = VerifyEvents(ProcessID, nodes)
	if err != nil {
		return 0, Node{}, err
	}

	//获取流程开始节点ID
	type result struct {
		Node_ID string
	}
	var r result
	_, err = ExecSQL("SELECT node_id FROM `proc_execution` WHERE proc_id=? AND node_type=0", &r, ProcessID)
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
	DB.Where("id=?", ProcessID).First(&procDef)

	//开启事务
	tx := DB.Begin()
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

	//获取流程起始人
	users,err:=resolveNodeUser(procInst.ID,StartNode)
	if err!=nil{
		tx.Rollback()
		return 0, StartNode, err
	}

	if len(users) < 1 {
		tx.Rollback()
		return 0,StartNode, errors.New("未指定处理人，无法处理节点:" + StartNode.NodeName)
	}

	//更新起始人到流程实例表
	re=tx.Model(&database.ProcInst{}).Where("id=?",procInst.ID).Update("starter",users[0])
	if re.Error != nil {
		tx.Rollback()
		return 0, StartNode, re.Error
	}

	//关闭事务
	tx.Commit()

	return procInst.ID, StartNode, nil
}

//开始流程实例 返回流程实例ID
func InstanceStart(ProcessID int, BusinessID string, Comment string, VariablesJson string) (int, error) {
	//实例初始化
	InstanceID, StartNode, err := instanceInit(ProcessID, BusinessID, VariablesJson)
	if err != nil {
		return 0, err
	}

	//开始节点处理
	err = startNodeHandle(InstanceID, &StartNode, Comment, VariablesJson)
	if err != nil {
		//需要删除刚才已经建立的实例记录、变量记录、任务记录。
		//这里已经没有必要对数据库执行做错误判断了，能删就删，删不掉也没多大关系
		DB.Where("id=?",InstanceID).Delete(database.ProcInst{})
		DB.Where("proc_inst_id=?",InstanceID).Delete(database.ProcInstVariable{})
		DB.Where("proc_inst_id=?",InstanceID).Delete(database.ProcTask{})

		return InstanceID, err
	}

	return InstanceID, nil
}

//撤销流程实例 参数说明:
//1、InstanceID 实例ID
//2、Force 是否强制撤销，若为false,则只有流程回到发起人这里才能撤销
//3、撤销发起人用户ID
func InstanceRevoke(ProcessInstanceID int, Force bool,RevokeUserID string) error {
	if !Force {
		//这段SQL判断是否当前Node就是开始Node
		sql := "SELECT a.id FROM proc_inst a " +
			"JOIN proc_execution b ON a.proc_id=b.proc_id AND a.current_node_id=b.node_id " +
			"WHERE a.id=? AND (b.prev_node_id IS NULL OR b.prev_node_id='') LIMIT 1"
		var id int
		_, err := ExecSQL(sql, &id, ProcessInstanceID)
		if err != nil {
			return err
		}
		if id == 0 {
			return errors.New("当前流程所在节点不是发起节点，无法撤销!")
		}
	}

	//-----------------------------执行流程撤销事件 start-----------------------------
	//流程ID
	ProcID,err:=GetProcessIDByInstanceID(ProcessInstanceID)
	if err!=nil{
		return err
	}
	//流程定义
	process,err:=GetProcessDefine(ProcID)
	if err!=nil{
		return err
	}

	err=RunProcEvents(process.RevokeEvents,ProcessInstanceID,RevokeUserID)
	if err!=nil{
		return err
	}
	//-----------------------------执行流程撤销事件 end-----------------------------

	//调用EndNodeHandle,做数据清理归档
	err = EndNodeHandle(ProcessInstanceID, 2)
	return err
}

//流程实例变量存入数据库
func InstanceVariablesSave(ProcessInstanceID int, VariablesJson string) error {
	//获取变量数组
	var variables []Variable
	Json2Struct(VariablesJson, &variables)

	tx := DB.Begin()
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

//获取流程实例信息
func GetInstanceInfo(ProcessInstanceID int) (Instance, error) {
	var procInst Instance
	//历史信息也要兼顾
	sql:="WITH tmp_procinst AS\n" +
		"(SELECT id,proc_id,proc_version,business_id,starter,current_node_id,\n" +
		"create_time,`status`\n" +
		"FROM proc_inst \n" +
		"WHERE id=?\n" +
		"UNION ALL\n" +
		"SELECT proc_inst_id AS id,proc_id,proc_version,business_id,starter,current_node_id,\n" +
		"create_time,`status` \n" +
		"FROM hist_proc_inst \n" +
		"WHERE proc_inst_id=?)\n" +

		"SELECT a.id,a.proc_id,a.proc_version,a.business_id,a.starter,\n" +
		"a.current_node_id,a.create_time,a.`status`,b.name \n" +
		"FROM  tmp_procinst a\n" +
		"LEFT JOIN proc_def b ON a.proc_id=b.id"

	_, err := ExecSQL(sql, &procInst, ProcessInstanceID,ProcessInstanceID)
	if err != nil {
		return procInst, err
	}

	return procInst, nil
}

//获取起始人为特定用户的流程实例。参数说明：
//UserID:用户ID 传入空则获取所有用户的流程实例
//ProcessName:指定流程名称,传入""则为全部
//StartIndex:分页用,开始index
//MaxRows:分页用,最大返回行数
func GetInstanceStartByUser(UserID string,ProcessName string,StartIndex int,MaxRows int) ([]Instance,error){
	var procInsts []Instance
	//历史信息也要兼顾
	sql:="WITH tmp_procinst AS\n " +
		"(SELECT id,proc_id,proc_version,business_id,starter,current_node_id,\n" +
		"create_time,`status`\n" +
		"FROM proc_inst \n" +
		"WHERE CASE WHEN ''=@userid THEN TRUE ELSE starter=@userid END\n" +
		"UNION ALL\n" +
		"SELECT proc_inst_id AS id,proc_id,proc_version,business_id,starter,current_node_id,\n" +
		"create_time,`status` \n" +
		"FROM hist_proc_inst \n" +
		"WHERE CASE WHEN ''=@userid THEN TRUE ELSE starter=@userid END)\n" +

		"SELECT a.id,a.proc_id,a.proc_version,a.business_id,\n" +
		"a.starter,a.current_node_id,a.create_time,a.`status`,b.name\n" +
		"FROM tmp_procinst a\n" +
		"JOIN proc_def b ON a.proc_id=b.id\n" +
		"WHERE CASE WHEN ''=@procname THEN TRUE ELSE b.name=@procname END\n"+
		"ORDER BY a.id limit @index,@rows"

	condition:=map[string]interface{}{"userid":UserID,"procname":ProcessName,"index":StartIndex,"rows":MaxRows}

	_, err := ExecSQL(sql, &procInsts, condition)
	if err != nil {
		return procInsts, err
	}

	return procInsts, nil
}