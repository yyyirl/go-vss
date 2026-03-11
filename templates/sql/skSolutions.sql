CREATE TABLE `skSolutions`
(
    `id`          int(11)     NOT NULL AUTO_INCREMENT,
    `title`       varchar(50) NOT NULL,
    `state`       tinyint(5)           DEFAULT 0,
    `createdAt`   bigint      NOT NULL DEFAULT 0,
    `updatedAt`   bigint      NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;
