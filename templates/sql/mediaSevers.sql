CREATE TABLE `mediaServers`
(
    `id`                      int(11)     NOT NULL AUTO_INCREMENT,
    `name`                    varchar(50) NOT NULL COMMENT '设备名称',
    `ip`                      char(30)    NOT NULL DEFAULT '' COMMENT '服务ip',
    `port`                    tinyint(4)  NOT NULL DEFAULT '' COMMENT '服务端口',
    `mediaServerStreamPortMin` tinyint(4)  NOT NULL DEFAULT 15000 COMMENT '推流端口范围最小值',
    `mediaServerStreamPortMax` tinyint(4)  NOT NULL DEFAULT 19000 COMMENT '推流端口范围最大值',
    `state`                   tinyint(4)  NOT NULL DEFAULT 1 COMMENT '启用状态 0 未启用 1 启用',
    `createdAt`               bigint      NOT NULL DEFAULT 0,
    `updatedAt`               bigint      NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT 'media server管理';
