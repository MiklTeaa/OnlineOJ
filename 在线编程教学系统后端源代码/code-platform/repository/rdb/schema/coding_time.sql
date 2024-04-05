CREATE TABLE `coding_time` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `lab_id` BIGINT UNSIGNED NOT NULL,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `duration` INT UNSIGNED NOT NULL COMMENT '编码时间，分钟为单位',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_at_date` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_labid_userid_duration` (`lab_id`, `user_id`, `duration`),
    KEY `idx_userid_createdatdate_duration`(`user_id`, `created_at_date`, `duration`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;