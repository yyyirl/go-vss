CREATE TABLE `users`
(
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID，主键',
    `username`      VARCHAR(50)     NOT NULL COMMENT '用户名',
    `password`      VARCHAR(255)    NOT NULL COMMENT '密码哈希值',
    `email`         VARCHAR(100)    NOT NULL COMMENT '电子邮箱',
    `phone`         VARCHAR(20)              DEFAULT NULL COMMENT '手机号码',
    `avatar`        VARCHAR(500)             DEFAULT NULL COMMENT '头像URL',
    `nickname`      VARCHAR(100)             DEFAULT NULL COMMENT '昵称',
    `gender`        TINYINT                  DEFAULT 0 COMMENT '性别：0-未知，1-男，2-女',
    `birthday`      BIGINT UNSIGNED          DEFAULT 0 COMMENT '出生日期',
    `status`        TINYINT         NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用，2-未激活',
    `emailVerified` TINYINT         NOT NULL DEFAULT 0 COMMENT '邮箱是否已验证：0-否，1-是',
    `phoneVerified` TINYINT         NOT NULL DEFAULT 0 COMMENT '手机是否已验证：0-否，1-是',
    `isDel`         TINYINT         NOT NULL DEFAULT 0 COMMENT '是否删除：0-否，1-是',
    `loginCount`    INT UNSIGNED             DEFAULT 0 COMMENT '累计登录次数',
    `lastLoginIp`   VARCHAR(45)              DEFAULT '' COMMENT '最后登录IP（支持IPv6）',
    `lastLoginTime` BIGINT UNSIGNED          DEFAULT 0 COMMENT '最后登录时间（时间戳）',
    `createdAt`     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间（时间戳）',
    `updatedAt`     BIGINT UNSIGNED          DEFAULT 0 COMMENT '更新时间（时间戳）',
    `deletedAt`     BIGINT UNSIGNED          DEFAULT 0 COMMENT '软删除时间（0表示未删除）',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_email` (`email`),
    UNIQUE KEY `uk_phone` (`phone`),
    KEY `idx_status` (`status`),
    KEY `idx_deleted_at` (`deletedAt`),
    KEY `idx_created_at` (`createdAt`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT = '用户表';