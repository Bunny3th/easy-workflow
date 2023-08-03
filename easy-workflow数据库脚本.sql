CREATE DATABASE easy_workflow;

DROP TABLE `proc_def`;
CREATE TABLE `proc_def` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流程ID',
  `name` VARCHAR(250) DEFAULT NULL COMMENT '流程名字',
  `version` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '版本号',
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
  `version` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '版本号',
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
   `is_completed` TINYINT DEFAULT 0 COMMENT '0:未完成 1:已完成',
  PRIMARY KEY (`id`)
) ;

#
CREATE TABLE `task` (
  `id` INT UNSIGNED NOT NULL COMMENT '任务ID',
  `proc_inst_id` INT UNSIGNED NOT NULL COMMENT '流程实例ID',
  `node_id` VARCHAR(250) NOT NULL COMMENT '节点ID',  
  `user_id` VARCHAR(250) NOT NULL COMMENT '分配用户ID',
  `is_cosigned` TINYINT DEFAULT 0 COMMENT '0:任意一人通过即可 1:会签',
  `is_passed` TINYINT DEFAULT NULL COMMENT '任务是否通过 0:驳回 1:通过',
  `is_finished` TINYINT DEFAULT 0 COMMENT '0:未执行 1:执行完成',
  `create_time` DATETIME DEFAULT NOW() COMMENT '系统创建任务时间',
  `finished_time` DATETIME DEFAULT NULL COMMENT '用户完成任务时间',
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
`gateway` VARCHAR(500) DEFAULT NULL COMMENT '网关定义(只有在nodetype为2时才会有)',
`is_cosigned` TINYINT NOT NULL COMMENT '是否会签',
`pre_events`  VARCHAR(500) DEFAULT NULL COMMENT '前置事件',
`exit_events` VARCHAR(500) DEFAULT NULL COMMENT '退出事件',
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
`gateway` VARCHAR(500) DEFAULT NULL COMMENT '网关定义(只有在nodetype为2时才会有)',
`is_cosigned` TINYINT NOT NULL COMMENT '是否会签',
`pre_events`  VARCHAR(500) DEFAULT NULL COMMENT '前置事件',
`exit_events` VARCHAR(500) DEFAULT NULL COMMENT '退出事件',
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


