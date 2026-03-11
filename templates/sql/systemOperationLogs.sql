CREATE TABLE `system-operation-logs`
(
    `id`        int(11)    NOT NULL AUTO_INCREMENT,
    `userid`    int(11)         NOT NULL COMMENT '管理员id',
    `type`      tinyint(5) NOT NULL COMMENT '操作类型',
    `data`      json                NOT NULL COMMENT '操作数据内容',
    `ip`        char(30)            NOT NULL DEFAULT '' COMMENT 'ip',
    `mac`       char(30)            NOT NULL DEFAULT '' COMMENT 'mac地址',
    `createdAt` bigint     NOT NULL DEFAULT 0,
    `updatedAt` bigint     NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE,
    INDEX `userid` (`userid`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT '系统操作日志';
