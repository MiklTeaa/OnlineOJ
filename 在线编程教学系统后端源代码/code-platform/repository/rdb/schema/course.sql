CREATE TABLE `course` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `teacher_id` BIGINT UNSIGNED NOT NULL,
    `name` VARCHAR(50) NOT NULL DEFAULT '',
    `description` TEXT NOT NULL,
    `pic_url` VARCHAR(200) NOT NULL DEFAULT '',
    `secret_key` VARCHAR(20) DEFAULT NULL COMMENT '加入课程的密码',
    `need_audit` TINYINT(1) NOT NULL DEFAULT 0,
    `is_closed` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '结课标志',
    `language` TINYINT NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_teacherid`(`teacher_id`),
    FULLTEXT KEY `fidx_name_description` (`name`, `description`) WITH PARSER NGRAM
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;