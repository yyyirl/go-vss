CREATE TABLE `video-projects`
(
    `id`               int(11)      NOT NULL AUTO_INCREMENT,
    `name`             varchar(100) NOT NULL COMMENT '计划名称',
    `state`            tinyint(4)   NOT NULL DEFAULT 1 COMMENT '启用状态 0 未启用 1 启用',
    `channelUniqueIds` varchar(255) NOT NULL COMMENT '通道id集合',
    `plans`            text         NOT NULL COMMENT '计划',
    `createdAt`        bigint       NOT NULL DEFAULT 0,
    `updatedAt`        bigint       NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT '录像计划';
