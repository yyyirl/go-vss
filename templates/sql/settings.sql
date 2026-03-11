CREATE TABLE `settings`
(
    `id`      int(11) NOT NULL AUTO_INCREMENT,
    `content` json NOT NULL DEFAULT (json_object()),
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT '设置表';