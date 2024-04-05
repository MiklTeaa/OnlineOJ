CREATE TABLE `detection_report` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `lab_id` BIGINT UNSIGNED NOT NULL,
    `data` LONGBLOB NOT NULL,
    `created_at` DATETIME(3) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uidx_labid_created_at`(`lab_id`, `created_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;