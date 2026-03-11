CREATE TABLE `crontab`
(
    `uniqueId`  char(32)    NOT NULL,
    `title`     varchar(100) NOT NULL COMMENT '标题',
    `interval`  int(11)     NOT NULL COMMENT '执行周期单位/s',
    `logs`      json                 DEFAULT (json_array()) COMMENT 'log',
    `createdAt` bigint      NOT NULL DEFAULT 0 COMMENT '创建时间',
    `updatedAt` bigint      NOT NULL DEFAULT 0 COMMENT '更新时间',
    PRIMARY KEY (`uniqueId`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT '任务';
