CREATE TABLE `configs`
(
    `id`                        int(11)    NOT NULL AUTO_INCREMENT,
    `apProtocol`                tinyint(4) NOT NULL DEFAULT 0 COMMENT 'access platform 接入平台配置 - 接入平台 3 GB28181协议 4 EHOME协议',
    `apSipId`                   varchar(100)        NOT NULL DEFAULT '' COMMENT 'access platform 接入平台配置 - sip id',
    `apSipRealm`                varchar(100)        NOT NULL DEFAULT '' COMMENT 'access platform 接入平台配置 - sip 域',
    `apSipHost`                 varchar(100)        NOT NULL DEFAULT '' COMMENT 'access platform 接入平台配置 - sip host',
    `apSipPort`                 varchar(100)        NOT NULL DEFAULT '' COMMENT 'access platform 接入平台配置 - sip 端口',
    `apDeviceAccessPassword`    varchar(100)        NOT NULL DEFAULT '' COMMENT 'access platform 接入平台配置 - 设备统一接入密码',
    `apPublicStreamAccessState` tinyint(4) NOT NULL DEFAULT 0 COMMENT 'access platform 接入平台配置 - 开启公网收流 0 未启用 1 启用',
    `apExternalIp`              char(30)            NOT NULL DEFAULT '' COMMENT 'access platform 接入平台配置 - 公网ip',
    `apStreamingPortRangeTcp`   varchar(50)         NOT NULL DEFAULT '' COMMENT 'access platform 接入平台配置 - 收流端口区间 TCP',
    `apStreamingPortRangeUdp`   varchar(50)         NOT NULL DEFAULT '' COMMENT 'access platform 接入平台配置 - 收流端口区间 UDP',
    `apBlacklistIds`            varchar(255)        NOT NULL DEFAULT '' COMMENT 'access platform 接入平台配置 - 和名单id',
    `apBlacklistIps`            varchar(255)        NOT NULL DEFAULT '' COMMENT 'access platform 接入平台配置 - 黑名单ip',
    `apWhitelistState`          tinyint(4) NOT NULL DEFAULT 0 COMMENT 'access platform 接入平台配置 - 白名单开起状态 0 未启用 1 启用',
    `apAutoRetrieveState`       tinyint(4) NOT NULL DEFAULT 0 COMMENT 'access platform 接入平台配置 - 自动检索状态 0 未启用 1 启用',
    `updatedAt`                 bigint     NOT NULL DEFAULT 0
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT '配置表';