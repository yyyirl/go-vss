CREATE TABLE `channels`
(
    `id`                int(11)    NOT NULL AUTO_INCREMENT,
    `name`              varchar(100)        NOT NULL COMMENT '通道名称',
    `protocol`          tinyint(4) NOT NULL DEFAULT 0 COMMENT '协议类型 1 RTSP 2 RTMP 3 HTTP 4 ONVIF',
    `onvifIP`           char(30)            NOT NULL DEFAULT '' COMMENT 'invif 探测ip',
    `onvifUsername`     varchar(100)        NOT NULL DEFAULT '' COMMENT 'invif 探测用户名',
    `onvifPassword`     varchar(100)        NOT NULL DEFAULT '' COMMENT 'invif 探测密码',
    `cameraUsername`    varchar(100)        NOT NULL DEFAULT '' COMMENT '摄像机用户名',
    `cameraPassword`    varchar(100)        NOT NULL DEFAULT '' COMMENT '摄像机密码',
    `rtspUrl`           varchar(100)        NOT NULL DEFAULT '' COMMENT '接入主码流',
    `cdnState`          tinyint(4) NOT NULL DEFAULT 0 COMMENT 'cdn开启状态 0 未开启 1 开启',
    `cdnUrl`            varchar(100)        NOT NULL DEFAULT '' COMMENT 'cdn地址',
    `longitude`         decimal(11, 8)      NOT NULL DEFAULT '' COMMENT '通道经度',
    `latitude`          decimal(11, 8)      NOT NULL DEFAULT '' COMMENT '通道纬度',
    `onDemandLiveState` tinyint(4) NOT NULL DEFAULT 0 COMMENT '按需直播 0 未开启 1 开启',
    `audioState`        tinyint(4) NOT NULL DEFAULT 0 COMMENT '开启音频 0 未开启 1 开启',
    `transcodedState`   tinyint(4) NOT NULL DEFAULT 0 COMMENT '是否转码 0 未开启 1 开启',
    `state`             tinyint(4) NOT NULL DEFAULT 1 COMMENT '启用状态 0 未启用 1 启用',
    `online`            tinyint(4) NOT NULL DEFAULT 1 COMMENT '在线状态 0 不在线 1 在线',
    `deviceId`          int(11)    NOT NULL COMMENT '设备id',
    `createdAt`         bigint     NOT NULL DEFAULT 0,
    `updatedAt`         bigint     NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE,
    INDEX deviceId (`deviceId`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT '通道';
