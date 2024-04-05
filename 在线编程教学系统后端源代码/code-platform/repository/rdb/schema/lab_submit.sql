CREATE TABLE `lab_submit` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `lab_id` BIGINT UNSIGNED NOT NULL,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `report_url` VARCHAR(200) NOT NULL COMMENT '存放实验报告pdf的url',
    `score` INT DEFAULT NULL,
    `is_finish` TINYINT(1) NOT NULL,
    `comment` TEXT NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uidx_userid_labid` (`user_id`, `lab_id`),
    UNIQUE KEY `uidx_labid_userid` (`lab_id`, `user_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;