CREATE TABLE user (
    `id` CHAR(36) NOT NULL,
    `created_at` DATETIME(6) NOT NULL,
    `updated_at` DATETIME(6) NOT NULL,
    `name` VARCHAR(255) NOT NULL,
    `email` VARCHAR(255) NOT NULL,
    `country` CHAR(2) NOT NULL,

    INDEX `by_country` (`country`, `id`),
    PRIMARY KEY (`id`)
) ENGINE=InnoDB;