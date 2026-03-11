CREATE TABLE `alarms`
(
    `id`             int(11)      NOT NULL AUTO_INCREMENT,
    `deviceUniqueId` char(70)     NOT NULL COMMENT '设备id',
    `channelUniqueId` char(70)     NOT NULL COMMENT '通道id',
    `level` int(11)     NOT NULL COMMENT '通道id',
    `type` int(11)     NOT NULL COMMENT '报警方式',
    `snapshot` varchar(255)     NOT NULL COMMENT '快照',
    `video` varchar(255)     NOT NULL COMMENT '录像',
    `createdAt`      bigint       NOT NULL DEFAULT 0,
    `updatedAt`      bigint       NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE,
    INDEX deviceUniqueId (`deviceUniqueId`),
    INDEX channelUniqueId (`channelUniqueId`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT '报警记录';
