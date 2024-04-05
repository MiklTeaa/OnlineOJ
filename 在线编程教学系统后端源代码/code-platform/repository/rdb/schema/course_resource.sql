CREATE TABLE `course_resource` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'id',
    `course_id` BIGINT UNSIGNED NOT NULL,
    `title` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '标题，限100字',
    `content` TEXT NOT NULL COMMENT '公告内容',
    `attachment_url` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '附件url',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_courseid` (`course_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '课程资源，包括公告，附件等等';