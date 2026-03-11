CREATE TABLE `devices`
(
    `id`             int(11)    NOT NULL AUTO_INCREMENT,
    `name`           varchar(50)         NOT NULL COMMENT '设备名称',
    `accessProtocol` tinyint(4) NOT NULL DEFAULT 0 COMMENT '接入协议 1 流媒体源 2 RTMP推流 3 GB28181协议 4 EHOME协议',
    `protocol`       tinyint(4) NOT NULL DEFAULT 0 COMMENT '协议 1 TCP 2 UDP',
    `deviceUniqueId` char(20)            NOT NULL DEFAULT '' COMMENT '设备id GB28181协议, EHOME协议',
    `state`          tinyint(4) NOT NULL DEFAULT 1 COMMENT '启用状态 0 未启用 1 启用',
    `online`         tinyint(4) NOT NULL DEFAULT 1 COMMENT '在线状态 0 不在线 1 在线',
    `createdAt`      bigint     NOT NULL DEFAULT 0,
    `updatedAt`      bigint     NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT '设备';
