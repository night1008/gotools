CREATE TABLE `permissions` (
  `name` text,
  `title` text,
  `domain` text,
  `resource` text,
  `action` text,
  `created_at` integer,
  PRIMARY KEY (`name`)
);
CREATE UNIQUE INDEX `idx_permissions_resource` ON `permissions`(`domain`,`resource`,`action`);


CREATE TABLE `permission_groups` (
  `name` text,
  `domain` text,
  `title` text,
  `group_index` integer,
  `parent_name` text,
  `created_at` integer,
  PRIMARY KEY (`name`)
);


CREATE TABLE `permission_group_permissions` (
  `permission_group_name` text,
  `permission_name` text,
  `created_at` integer,
  PRIMARY KEY (`permission_group_name`,`permission_name`)
);


CREATE TABLE `roles` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `roleable_type` text,
  `roleable_id` integer,
  `name` text,
  `title` text,
  `description` text,
  `creator_user_id` integer,
  `created_at` integer,
  `updated_at` integer
);
CREATE UNIQUE INDEX `idx_roles_name` ON `roles`(`roleable_type`,`roleable_id`,`name`);


CREATE TABLE `role_permission_groups` (
  `role_id` integer,
  `permission_group_name` text,
  `created_at` integer,
  PRIMARY KEY (`role_id`,`permission_group_name`)
);


CREATE TABLE `user_roles` (
  `user_id` integer,
  `role_id` integer,
  `created_at` integer,
  PRIMARY KEY (`user_id`,`role_id`)
);