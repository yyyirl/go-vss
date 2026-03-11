CREATE TABLE `cascade`
(
    `id`                int(11)      NOT NULL AUTO_INCREMENT,
    `name`              varchar(50)  NOT NULL COMMENT '平台名称',
    `protocol`          tinyint(4)   NOT NULL DEFAULT 1 COMMENT '信令传输协议 1 TCP 2 UDP',
    `sipId`             varchar(100) NOT NULL COMMENT 'SIP服务国标编码',
    `sipDomain`         varchar(100) NOT NULL COMMENT 'SIP服务国标域',
    `sipIp`             varchar(100) NOT NULL COMMENT 'SIP服务IP',
    `sipPort`           tinyint(4)   NOT NULL COMMENT 'SIP服务端口',
    `username`          varchar(100) NOT NULL DEFAULT '' COMMENT 'SIP认证用户',
    `password`          varchar(100) NOT NULL DEFAULT '' COMMENT 'SIP认证密码',
    `localIp`           varchar(100) NOT NULL COMMENT '本地级联IP',
    `keepaliveInterval` tinyint(4)   NOT NULL DEFAULT 60 COMMENT '心跳间隔(秒)',
    `registerInterval`  tinyint(4)   NOT NULL DEFAULT 60 COMMENT '注册间隔(秒)',
    `registerTimeout`   tinyint(4)   NOT NULL DEFAULT 3600 COMMENT '注册有效期(秒)',
    `commandTransport`  tinyint(4)   NOT NULL DEFAULT 3600 COMMENT '信令传输 TCP UDP',
    `state`             tinyint(4)   NOT NULL DEFAULT 1 COMMENT '启用状态 0 未启用 1 启用',
    `online`            tinyint(4)   NOT NULL DEFAULT 1 COMMENT '在线状态 0 不在线 1 在线',
    `createdAt`         bigint       NOT NULL DEFAULT 0,
    `updatedAt`         bigint       NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT '平台级联';

