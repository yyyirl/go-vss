CREATE TABLE `permissions`
(
    `uniqueId`       char(32)        NOT NULL,
    `name`           varchar(50)     NOT NULL COMMENT '权限名称',
    `path`           varchar(255)    NOT NULL COMMENT '权限路径',
    `method`         char(50)        NOT NULL COMMENT '请求方式',
    `state`          tinyint(5)      DEFAULT 0 COMMENT '权限使用状态',
    `super`          tinyint(4)      DEFAULT 0 COMMENT '超级管理员使用',
    `type`           tinyint(4)      DEFAULT 0 COMMENT '0 权限 1 分组',
    `common`         tinyint(4)      DEFAULT 0 COMMENT '通用',
    `parentUniqueId` char(32)        NOT NULL DEFAULT '' COMMENT '父级id',
    `relationship`   varchar(255)    NOT NULL DEFAULT '' COMMENT '前端映射 路由',
    `sort`           DECIMAL(10, 6)  NOT NULL DEFAULT 0 COMMENT '排序',
    `createdAt`      bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
    `updatedAt`      bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
    UNIQUE KEY `uix_permissions_path_method` (`path`, `method`),
    PRIMARY KEY (`uniqueId`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT '权限资源列表';
