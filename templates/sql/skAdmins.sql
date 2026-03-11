CREATE TABLE `skAdmins`
(
    `id`        int(11) NOT NULL AUTO_INCREMENT,
    `username`  varchar(50)      NOT NULL,
    `password`  varchar(255)     NOT NULL,
    `createdAt` bigint  NOT NULL DEFAULT 0,
    `updatedAt` bigint  NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `username` (`username`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;
