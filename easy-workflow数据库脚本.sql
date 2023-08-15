CREATE DATABASE easy_workflow;

DROP TABLE `proc_def`;
CREATE TABLE `proc_def` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流程ID',
  `name` VARCHAR(250) DEFAULT NULL COMMENT '流程名字',
  `version` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '版本号',
  `resource` TEXT NOT NULL COMMENT '流程定义模板',
  `user_id` VARCHAR(250) NOT NULL COMMENT '创建者ID',
  `source` VARCHAR(250) NOT NULL COMMENT '来源(引擎可能被多个系统、组件等使用，这里记下从哪个来源创建的流程)',
  `create_time` DATETIME DEFAULT NOW() COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ;

CREATE UNIQUE INDEX uix_name_source ON proc_def(`name`,source);


DROP TABLE `hist_proc_def`;
CREATE TABLE `hist_proc_def` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, 
  `proc_id` INT UNSIGNED NOT NULL COMMENT '流程ID',
  `name` VARCHAR(250) DEFAULT NULL COMMENT '流程名字',
  `version` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '版本号',
  `resource` TEXT NOT NULL COMMENT '流程定义模板',
  `user_id` VARCHAR(250) NOT NULL COMMENT '创建者ID',
  `source` VARCHAR(250) NOT NULL COMMENT '来源(引擎可能被多个系统、组件等使用，这里记下从哪个来源创建的流程)',
  `create_time` DATETIME DEFAULT NOW() COMMENT '创建时间'
) 




DROP TABLE `proc_inst`;
CREATE TABLE `proc_inst` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流程实例ID',
  `proc_id` INT NOT NULL COMMENT '流程ID',
  `proc_version` INT UNSIGNED NOT NULL COMMENT '流程版本号', 
  `business_id` VARCHAR(250) DEFAULT NULL COMMENT '业务ID',
  `current_node_id` VARCHAR(250) NOT NULL COMMENT '当前进行节点ID',  
  `create_time` DATETIME DEFAULT NOW(),  
   `status` TINYINT DEFAULT 0 COMMENT '0:未完成 1:已完成 2:撤销',
  PRIMARY KEY (`id`)
) ;


DROP TABLE `hist_proc_inst`;
CREATE TABLE `hist_proc_inst` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  proc_inst_id INT UNSIGNED NOT NULL COMMENT '流程实例ID',
  `proc_id` INT NOT NULL COMMENT '流程ID',
  `proc_version` INT UNSIGNED NOT NULL COMMENT '流程版本号', 
  `business_id` VARCHAR(250) DEFAULT NULL COMMENT '业务ID',
  `current_node_id` VARCHAR(250) NOT NULL COMMENT '当前进行节点ID',  
  `create_time` DATETIME DEFAULT NOW(),  
   `status` TINYINT DEFAULT 0 COMMENT '0:未完成 1:已完成 2:撤销',
  PRIMARY KEY (`id`)
) ;



#
DROP TABLE `task`;
CREATE TABLE `task` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `proc_id` INT UNSIGNED NOT NULL COMMENT '流程ID,冗余字段，偷懒用',
  `proc_inst_id` INT UNSIGNED NOT NULL COMMENT '流程实例ID',
  `node_id` VARCHAR(250) NOT NULL COMMENT '节点ID',  
  `prev_node_id` VARCHAR(250) DEFAULT NULL COMMENT '上个处理节点ID,注意这里和execution中的上一个节点不一样，这里是实际审批处理时上个已处理节点的ID',  
  `is_cosigned` TINYINT DEFAULT 0 COMMENT '0:任意一人通过即可 1:会签',
  `batch_code` VARCHAR(50) DEFAULT NULL COMMENT '批次码.节点会被驳回，一个节点可能产生多批task,用此码做分别',
   `user_id` VARCHAR(250) NOT NULL COMMENT '分配用户ID',
  `is_passed` TINYINT DEFAULT NULL COMMENT '任务是否通过 0:驳回 1:通过',
  `is_finished` TINYINT DEFAULT 0 COMMENT '0:任务未处理 1:处理完成.任务未必都是用户处理的，比如会签时一人驳回，其他任务系统自动设为已处理',
  `create_time` DATETIME DEFAULT NOW() COMMENT '系统创建任务时间',
  `finished_time` DATETIME DEFAULT NULL COMMENT '处理任务时间',
  PRIMARY KEY (`id`)
) 

#思考 task表中是否需要加 prenodeid，应该需要。
#因为一个节点的上级节点可能不是一个，所以节点驳回的时候，就需要知道往哪个节点驳回
#是否会签可以冗余在表中否？如果冗余，则可以不用去`proc_execution`表读取
#task表中是否需要加一个子线程id？如果会签，会产生N个task，第一次可以通过nodeid获取所有的task
#但是第二次呢？就会把上一次的task也获取到


DROP TABLE `hist_task`;
CREATE TABLE `hist_task` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  task_id INT UNSIGNED NOT NULL COMMENT '任务ID',
  `proc_id` INT UNSIGNED NOT NULL COMMENT '流程ID,冗余字段，偷懒用',
  `proc_inst_id` INT UNSIGNED NOT NULL COMMENT '流程实例ID',
  `node_id` VARCHAR(250) NOT NULL COMMENT '节点ID',  
  `prev_node_id` VARCHAR(250) DEFAULT NULL COMMENT '上一节点ID',  
  `is_cosigned` TINYINT DEFAULT 0 COMMENT '0:任意一人通过即可 1:会签',
  `batch_code` VARCHAR(50) DEFAULT NULL COMMENT '批次码.节点会被驳回，一个节点可能产生多批task,用此码做分别',
   `user_id` VARCHAR(250) NOT NULL COMMENT '分配用户ID',
  `is_passed` TINYINT DEFAULT NULL COMMENT '任务是否通过 0:驳回 1:通过',
  `is_finished` TINYINT DEFAULT 0 COMMENT '0:任务未处理 1:处理完成.任务未必都是用户处理的，比如会签时一人驳回，其他任务系统自动设为已处理',
  `create_time` DATETIME DEFAULT NOW() COMMENT '系统创建任务时间',
  `finished_time` DATETIME DEFAULT NULL COMMENT '处理任务时间',
  PRIMARY KEY (`id`)
) 



CREATE TABLE task_comment
(
`id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
task_id INT UNSIGNED NOT NULL COMMENT '任务ID',
`comment` TEXT COMMENT '任务备注'
);

CREATE INDEX ix_task_id ON task_comment(task_id);

DROP TABLE proc_execution;
CREATE TABLE proc_execution(
`id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
`proc_id` INT NOT NULL COMMENT '流程ID',
`proc_version` INT UNSIGNED NOT NULL COMMENT '流程版本号', 
`node_id` VARCHAR(250) NOT NULL COMMENT '节点ID',  
`node_name` VARCHAR(250) NOT NULL COMMENT '节点名称',
`prev_node_id` VARCHAR(250) DEFAULT NULL COMMENT '上级节点ID',  
`node_type` TINYINT NOT NULL COMMENT '流程类型 0:开始节点 1:任务节点 2:网关节点 3:结束节点',
#`gateway` VARCHAR(500) DEFAULT NULL COMMENT '网关定义(只有在nodetype为2时才会有)',
`is_cosigned` TINYINT NOT NULL COMMENT '是否会签',
#`pre_events`  VARCHAR(500) DEFAULT NULL COMMENT '前置事件',
#`exit_events` VARCHAR(500) DEFAULT NULL COMMENT '退出事件',
`create_time` DATETIME DEFAULT NOW() COMMENT '创建时间'
);

DROP TABLE hist_proc_execution;
CREATE TABLE hist_proc_execution(
`id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
`proc_id` INT NOT NULL COMMENT '流程ID',
`proc_version` INT UNSIGNED NOT NULL COMMENT '流程版本号', 
`node_id` VARCHAR(250) NOT NULL COMMENT '节点ID',  
`node_name` VARCHAR(250) NOT NULL COMMENT '节点名称',
`prev_node_id` VARCHAR(250) DEFAULT NULL COMMENT '上级节点ID',  
`node_type` TINYINT NOT NULL COMMENT '流程类型 0:开始节点 1:任务节点 2:网关节点 3:结束节点',
#`gateway` VARCHAR(500) DEFAULT NULL COMMENT '网关定义(只有在nodetype为2时才会有)',
`is_cosigned` TINYINT NOT NULL COMMENT '是否会签',
#`pre_events`  VARCHAR(500) DEFAULT NULL COMMENT '前置事件',
#`exit_events` VARCHAR(500) DEFAULT NULL COMMENT '退出事件',
`create_time` DATETIME DEFAULT NOW() COMMENT '创建时间'
);



#需增加proc_inst_variables表
DROP TABLE proc_inst_variable;
CREATE TABLE proc_inst_variable(
`id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
`proc_inst_id` INT UNSIGNED NOT NULL COMMENT '流程实例ID',
`key` VARCHAR(250) NOT NULL COMMENT '变量key',
`value` VARCHAR(250) NOT NULL COMMENT '变量value'
);

CREATE INDEX ix_proc_inst_id ON proc_inst_variable(proc_inst_id)


DROP TABLE hist_proc_inst_variable;
CREATE TABLE hist_proc_inst_variable(
`id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
`proc_inst_id` INT UNSIGNED NOT NULL COMMENT '流程实例ID',
`key` VARCHAR(250) NOT NULL COMMENT '变量key',
`value` VARCHAR(250) NOT NULL COMMENT '变量value'
);





#2023-08-07 待办
#proc_inst表中，current_node_id 需要更新  ok  更新到最后一个task node

#20230809待办 
#解决
#sp_task_next_opt_node 逻辑不对，在多次驳回提交后，上下节点的判断有误
#歧义出现在分支节点，因为有分支，所以不知道流向是什么
#但是感觉可以倒推，倒推就能不受分支干扰




















