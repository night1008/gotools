CREATE TABLE permissions (
  name character varying(256) PRIMARY KEY,
  title text,
  domain character varying(128),
  resource character varying(256),
  action character varying(64),
  created_at bigint
);
CREATE UNIQUE INDEX permissions_pkey ON permissions(name text_ops);
CREATE UNIQUE INDEX idx_permissions_resource ON permissions(domain text_ops,resource text_ops,action text_ops);


CREATE TABLE permission_groups (
  name character varying(256) PRIMARY KEY,
  domain text,
  title text,
  group_index bigint,
  parent_name text,
  created_at bigint
);
CREATE UNIQUE INDEX permission_groups_pkey ON permission_groups(name text_ops);


CREATE TABLE permission_group_permissions (
  permission_group_name character varying(256),
  permission_name character varying(256),
  created_at bigint,
  CONSTRAINT permission_group_permissions_pkey PRIMARY KEY (permission_group_name, permission_name)
);
CREATE UNIQUE INDEX permission_group_permissions_pkey ON permission_group_permissions(permission_group_name text_ops,permission_name text_ops);


CREATE TABLE roles (
  id BIGSERIAL PRIMARY KEY,
  roleable_type character varying(128),
  roleable_id bigint,
  name character varying(256),
  title text,
  description text,
  creator_user_id bigint,
  created_at bigint,
  updated_at bigint
);
CREATE UNIQUE INDEX roles_pkey ON roles(id int8_ops);
CREATE UNIQUE INDEX idx_roles_name ON roles(roleable_type text_ops,roleable_id int8_ops,name text_ops);


CREATE TABLE role_permission_groups (
  role_id bigint,
  permission_group_name character varying(256),
  created_at bigint,
  CONSTRAINT role_permission_groups_pkey PRIMARY KEY (role_id, permission_group_name)
);
CREATE UNIQUE INDEX role_permission_groups_pkey ON role_permission_groups(role_id int8_ops,permission_group_name text_ops);


CREATE TABLE user_roles (
  user_id bigint,
  role_id bigint,
  created_at bigint,
  CONSTRAINT user_roles_pkey PRIMARY KEY (user_id, role_id)
);
CREATE UNIQUE INDEX user_roles_pkey ON user_roles(user_id int8_ops,role_id int8_ops);