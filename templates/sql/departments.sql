CREATE TABLE `departments`
(
    `id`        int(11)    NOT NULL AUTO_INCREMENT,
    `name`      varchar(100)        NOT NULL COMMENT '组织部门名称',
    `remark`    varchar(255)        NOT NULL COMMENT '备注',
    `parentId`  int(11)             NOT NULL COMMENT '父级部门',
    `roleIds`   json                         DEFAULT (json_array()) COMMENT '角色id集合',
    `state`     tinyint(4) NOT NULL DEFAULT 1 COMMENT '启用状态 0 未启用 1 启用',
    `createdAt` bigint     NOT NULL DEFAULT 0,
    `updatedAt` bigint     NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT '组织机构';
