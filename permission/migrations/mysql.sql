CREATE TABLE `permissions` (
  `name` varchar(256) NOT NULL,
  `title` longtext,
  `domain` varchar(128) DEFAULT NULL,
  `resource` varchar(256) DEFAULT NULL,
  `action` varchar(64) DEFAULT NULL,
  `created_at` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`name`),
  UNIQUE KEY `idx_permissions_resource` (`domain`,`resource`,`action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `permission_groups` (
  `name` varchar(256) NOT NULL,
  `domain` longtext,
  `title` longtext,
  `group_index` bigint(20) DEFAULT NULL,
  `parent_name` longtext,
  `created_at` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `permission_group_permissions` (
  `permission_group_name` varchar(256) NOT NULL,
  `permission_name` varchar(256) NOT NULL,
  `created_at` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`permission_group_name`,`permission_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `roles` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `roleable_type` varchar(128) DEFAULT NULL,
  `roleable_id` bigint(20) DEFAULT NULL,
  `name` varchar(256) DEFAULT NULL,
  `title` longtext,
  `description` longtext,
  `creator_user_id` bigint(20) DEFAULT NULL,
  `created_at` bigint(20) DEFAULT NULL,
  `updated_at` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_roles_name` (`roleable_type`,`roleable_id`,`name`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4;


 CREATE TABLE `role_permission_groups` (
  `role_id` bigint(20) NOT NULL
  `permission_group_name` varchar(256) NOT NULL
  `created_at` bigint(20) DEFAULT NULL
  PRIMARY KEY (`role_id`,`permission_group_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `user_roles` (
  `user_id` bigint(20) NOT NULL,
  `role_id` bigint(20) NOT NULL,
  `created_at` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`user_id`,`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4