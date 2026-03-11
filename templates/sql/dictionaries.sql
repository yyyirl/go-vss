CREATE TABLE `dictionaries`
(
    `id`        int(11)    NOT NULL AUTO_INCREMENT,
    `name`      varchar(100)        NOT NULL COMMENT '名称',
    `uniqueId`  char(30)            NOT NULL COMMENT '唯一id 标注值',
    `parentId`  int(11)    NOT NULL DEFAULT 0 COMMENT '父级id',
    `type`      tinyint(4) NOT NULL COMMENT '类型 1 fod分类自定义标签',
    `state`     tinyint(4) NOT NULL DEFAULT 0 COMMENT '启用状态 0 未启用 1 启用',
    `createdAt` bigint     NOT NULL DEFAULT 0,
    `updatedAt` bigint     NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY uniqueId(`uniqueId`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT '字典';