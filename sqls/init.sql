-- LandcPermission Database Schema

CREATE TABLE IF NOT EXISTS `roles` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(100) NOT NULL COMMENT 'Role display name',
    `code` VARCHAR(100) NOT NULL COMMENT 'Role unique code',
    `description` VARCHAR(500) DEFAULT '' COMMENT 'Role description',
    `status` TINYINT DEFAULT 1 COMMENT '1=active, 0=disabled',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`),
    UNIQUE KEY `uk_code` (`code`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `permissions` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(200) NOT NULL COMMENT 'Permission display name',
    `code` VARCHAR(200) NOT NULL COMMENT 'Permission unique code',
    `resource` VARCHAR(200) NOT NULL COMMENT 'Resource identifier (e.g., user.profile, order.*)',
    `action` VARCHAR(50) NOT NULL COMMENT 'Action: read/write/delete/admin/*',
    `description` VARCHAR(500) DEFAULT '' COMMENT 'Permission description',
    `status` TINYINT DEFAULT 1 COMMENT '1=active, 0=disabled',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`),
    KEY `idx_resource` (`resource`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `role_permissions` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `role_id` BIGINT UNSIGNED NOT NULL,
    `permission_id` BIGINT UNSIGNED NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_permission` (`role_id`, `permission_id`),
    KEY `idx_role_id` (`role_id`),
    KEY `idx_permission_id` (`permission_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_roles` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `role_id` BIGINT UNSIGNED NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_role` (`user_id`, `role_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_role_id` (`role_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- 用户直接权限：绕过角色的定向授权（应急提权 / 外部协作 / 例外授权）。
-- expire_at 为 NULL 表示永久有效；非 NULL 且已过期时在校验阶段被过滤，
-- 正确性不依赖清理任务（定期清理仅用于回收空间）。
CREATE TABLE IF NOT EXISTS `user_permissions` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `permission_id` BIGINT UNSIGNED NOT NULL,
    `expire_at` DATETIME NULL DEFAULT NULL COMMENT 'NULL=永久有效',
    `reason` VARCHAR(500) DEFAULT '' COMMENT '授权原因，供审计追溯',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_permission` (`user_id`, `permission_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_expire_at` (`expire_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- Default roles
INSERT INTO
    `roles` (`name`, `code`, `description`)
VALUES (
        'Super Admin',
        'super_admin',
        'Full system access'
    ),
    (
        'Admin',
        'admin',
        'Administrative access'
    ),
    (
        'User',
        'user',
        'Standard user access'
    );

-- Default permissions
INSERT INTO
    `permissions` (
        `name`,
        `code`,
        `resource`,
        `action`,
        `description`
    )
VALUES (
        'All Access',
        'all_access',
        '*',
        'admin',
        'Full access to all resources'
    ),
    (
        'User Read',
        'user_read',
        'user',
        'read',
        'Read user information'
    ),
    (
        'User Write',
        'user_write',
        'user',
        'write',
        'Modify user information'
    ),
    (
        'Role Manage',
        'role_manage',
        'role',
        'admin',
        'Manage roles'
    ),
    (
        'Permission Manage',
        'permission_manage',
        'permission',
        'admin',
        'Manage permissions'
    );

-- Assign all permissions to super_admin
INSERT INTO
    `role_permissions` (`role_id`, `permission_id`)
SELECT 1, id
FROM `permissions`;

-- Assign user read/write to user role
INSERT INTO
    `role_permissions` (`role_id`, `permission_id`)
VALUES (3, 2),
    (3, 3);