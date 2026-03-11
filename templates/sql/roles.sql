CREATE TABLE `roles`
(
    `id`                  int(11) NOT NULL AUTO_INCREMENT,
    `name`                varchar(100)     NOT NULL COMMENT '角色名称',
    `permissionUniqueIds` json                DEFAULT (json_array()) COMMENT '权限id集合',
    `state`               tinyint(5) DEFAULT 0 COMMENT '启用状态 0 默认不启用 1 启用',
    `remark`              varchar(255)        DEFAULT '' COMMENT '备注',
    `isDel`               tinyint(5) DEFAULT 0,
    `createdAt`           bigint     DEFAULT 0,
    `updatedAt`           bigint     DEFAULT 0,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;
