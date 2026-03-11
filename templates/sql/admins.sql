CREATE TABLE `admins`
(
    `id`        int(11) NOT NULL AUTO_INCREMENT,
    `username`  varchar(50)      NOT NULL,
    `password`  varchar(255)     NOT NULL,
    `nickname`  varchar(50)      NOT NULL,
    `email`     varchar(100)     NOT NULL,
    `mobile`    varchar(11)      NOT NULL,
    `state`     tinyint(5)       DEFAULT 0,
    `depIds`    json                      DEFAULT (json_array()) COMMENT '组织部门id集合',
    `remark`    varchar(255)              DEFAULT '',
    `sex`       tinyint(5)       DEFAULT 0,
    `head`      varchar(255)              DEFAULT '',
    `super`     tinyint(4)                DEFAULT 0,
    `isDel`     tinyint(4)                DEFAULT 0,
    `createdAt` bigint  NOT NULL DEFAULT 0,
    `updatedAt` bigint  NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `username` (`username`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;
