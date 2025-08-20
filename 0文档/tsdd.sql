/*
 Navicat Premium Dump SQL

 Source Server         : 本地mysql8
 Source Server Type    : MySQL
 Source Server Version : 80042 (8.0.42)
 Source Host           : localhost:3306
 Source Schema         : tsdd

 Target Server Type    : MySQL
 Target Server Version : 80042 (8.0.42)
 File Encoding         : 65001

 Date: 17/08/2025 23:22:43
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for app
-- ----------------------------
DROP TABLE IF EXISTS `app`;
CREATE TABLE `app` (
  `app_id` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'app id',
  `app_key` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'app key',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态 0.禁用 1.可用',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `app_name` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'app名字',
  `app_logo` varchar(400) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'app logo',
  UNIQUE KEY `app_id` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for app_config
-- ----------------------------
DROP TABLE IF EXISTS `app_config`;
CREATE TABLE `app_config` (
  `id` int NOT NULL AUTO_INCREMENT,
  `rsa_private_key` varchar(4000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `rsa_public_key` varchar(4000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `version` int NOT NULL DEFAULT '0',
  `super_token` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `super_token_on` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `revoke_second` smallint NOT NULL DEFAULT '0' COMMENT '消息可撤回时长',
  `welcome_message` varchar(2000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '登录欢迎语',
  `new_user_join_system_group` smallint NOT NULL DEFAULT '1' COMMENT '注册用户是否默认加入系统群',
  `search_by_phone` smallint NOT NULL DEFAULT '0' COMMENT '是否可通过手机号搜索',
  `register_invite_on` smallint NOT NULL DEFAULT '0' COMMENT '是否开启注册邀请',
  `send_welcome_message_on` smallint NOT NULL DEFAULT '1' COMMENT '是否开启登录欢迎语',
  `invite_system_account_join_group_on` smallint NOT NULL DEFAULT '0' COMMENT '是否开启系统账号进入群聊',
  `register_user_must_complete_info_on` smallint NOT NULL DEFAULT '0' COMMENT '注册用户是否必须完善信息',
  `channel_pinned_message_max_count` smallint NOT NULL DEFAULT '10' COMMENT '频道最多置顶消息数量',
  `can_modify_api_url` smallint NOT NULL DEFAULT '0' COMMENT '是否能修改服务器地址',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for app_module
-- ----------------------------
DROP TABLE IF EXISTS `app_module`;
CREATE TABLE `app_module` (
  `id` int NOT NULL AUTO_INCREMENT,
  `sid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `name` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `desc` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `status` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `app_module_sid_idx` (`sid`)
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for app_version
-- ----------------------------
DROP TABLE IF EXISTS `app_version`;
CREATE TABLE `app_version` (
  `id` int NOT NULL AUTO_INCREMENT,
  `app_version` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `os` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `is_force` smallint NOT NULL DEFAULT '0',
  `update_desc` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `download_url` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `signature` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '二进制包的签名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for channel_offset
-- ----------------------------
DROP TABLE IF EXISTS `channel_offset`;
CREATE TABLE `channel_offset` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid_channel_idx` (`uid`,`channel_id`,`channel_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for channel_offset1
-- ----------------------------
DROP TABLE IF EXISTS `channel_offset1`;
CREATE TABLE `channel_offset1` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid_channel_idx` (`uid`,`channel_id`,`channel_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for channel_offset2
-- ----------------------------
DROP TABLE IF EXISTS `channel_offset2`;
CREATE TABLE `channel_offset2` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid_channel_idx` (`uid`,`channel_id`,`channel_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for channel_setting
-- ----------------------------
DROP TABLE IF EXISTS `channel_setting`;
CREATE TABLE `channel_setting` (
  `id` int NOT NULL AUTO_INCREMENT,
  `channel_id` varchar(80) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `channel_type` smallint NOT NULL DEFAULT '0',
  `parent_channel_id` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `parent_channel_type` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `msg_auto_delete` int NOT NULL DEFAULT '0' COMMENT '消息定时删除时间',
  `offset_message_seq` int NOT NULL DEFAULT '0' COMMENT 'channel消息删除偏移seq',
  PRIMARY KEY (`id`),
  UNIQUE KEY `channel_setting_uidx` (`channel_id`,`channel_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for chat_bg
-- ----------------------------
DROP TABLE IF EXISTS `chat_bg`;
CREATE TABLE `chat_bg` (
  `id` int NOT NULL AUTO_INCREMENT,
  `cover` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `url` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `is_svg` smallint NOT NULL DEFAULT '1',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for conversation_extra
-- ----------------------------
DROP TABLE IF EXISTS `conversation_extra`;
CREATE TABLE `conversation_extra` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '所属用户',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` smallint NOT NULL DEFAULT '0' COMMENT '频道类型',
  `browse_to` bigint NOT NULL DEFAULT '0' COMMENT '预览到的位置，与会话保持位置不同的是 预览到的位置是用户读到的最大的messageSeq。跟未读消息数量有关系',
  `keep_message_seq` bigint NOT NULL DEFAULT '0' COMMENT '会话保持的位置',
  `keep_offset_y` int NOT NULL DEFAULT '0' COMMENT '会话保持的位置的偏移量',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  `draft` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '草稿',
  `version` bigint NOT NULL DEFAULT '0' COMMENT '数据版本',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid_channel_idx` (`uid`,`channel_id`,`channel_type`),
  KEY `uid_idx` (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=99 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for device
-- ----------------------------
DROP TABLE IF EXISTS `device`;
CREATE TABLE `device` (
  `id` int NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `device_id` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `device_name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `device_model` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `last_login` int NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `device_uid_device_id` (`uid`,`device_id`),
  KEY `device_uid` (`uid`),
  KEY `device_device_id` (`device_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for device_flag
-- ----------------------------
DROP TABLE IF EXISTS `device_flag`;
CREATE TABLE `device_flag` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `device_flag` smallint NOT NULL DEFAULT '0' COMMENT '设备标记 0. app 1.Web 2.PC',
  `weight` int NOT NULL DEFAULT '0' COMMENT '设备权重 值越大越优先',
  `remark` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `udx_device_flag` (`device_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for device_offset
-- ----------------------------
DROP TABLE IF EXISTS `device_offset`;
CREATE TABLE `device_offset` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `device_uuid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid_device_offset_unidx` (`uid`,`device_uuid`,`channel_id`,`channel_type`),
  KEY `uid_device_offset_idx` (`uid`,`device_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for event
-- ----------------------------
DROP TABLE IF EXISTS `event`;
CREATE TABLE `event` (
  `id` int NOT NULL AUTO_INCREMENT,
  `event` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `type` smallint NOT NULL DEFAULT '0',
  `data` varchar(10000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `status` smallint NOT NULL DEFAULT '0',
  `reason` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `version_lock` int NOT NULL DEFAULT '0' COMMENT '乐观锁',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `event_key` (`event`),
  KEY `event_type` (`type`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for friend
-- ----------------------------
DROP TABLE IF EXISTS `friend`;
CREATE TABLE `friend` (
  `id` int NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户UID',
  `to_uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '好友uid',
  `remark` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '对好友的备注 TODO: 此字段不再使用，已经迁移到user_setting表',
  `flag` smallint NOT NULL DEFAULT '0' COMMENT '好友标示',
  `version` bigint NOT NULL DEFAULT '0' COMMENT '版本号',
  `vercode` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '验证码 加好友来源',
  `source_vercode` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '好友来源',
  `is_deleted` smallint NOT NULL DEFAULT '0' COMMENT '是否已删除',
  `is_alone` smallint NOT NULL DEFAULT '0' COMMENT '单项好友',
  `initiator` smallint NOT NULL DEFAULT '0' COMMENT '加好友发起方',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `to_uid_uid` (`uid`,`to_uid`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for friend_apply_record
-- ----------------------------
DROP TABLE IF EXISTS `friend_apply_record`;
CREATE TABLE `friend_apply_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `to_uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `remark` varchar(200) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `status` smallint NOT NULL DEFAULT '1',
  `token` varchar(200) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `friend_apply_record_uid_touidx` (`uid`,`to_uid`),
  KEY `friend_apply_record_uidx` (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for gitee_user
-- ----------------------------
DROP TABLE IF EXISTS `gitee_user`;
CREATE TABLE `gitee_user` (
  `id` bigint NOT NULL DEFAULT '0' COMMENT '用户 ID',
  `login` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户名',
  `name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户姓名',
  `email` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户邮箱',
  `bio` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户简介',
  `avatar_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户头像 URL',
  `blog` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户博客 URL',
  `events_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户事件 URL',
  `followers` int NOT NULL DEFAULT '0' COMMENT '用户粉丝数',
  `followers_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户粉丝 URL',
  `following` int NOT NULL DEFAULT '0' COMMENT '用户关注数',
  `following_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户关注 URL',
  `gists_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户 Gist URL',
  `html_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户主页 URL',
  `member_role` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户角色',
  `organizations_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户组织 URL',
  `public_gists` int NOT NULL DEFAULT '0' COMMENT '用户公开 Gist 数',
  `public_repos` int NOT NULL DEFAULT '0' COMMENT '用户公开仓库数',
  `received_events_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户接收事件 URL',
  `remark` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '企业备注名',
  `repos_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户仓库 URL',
  `stared` int NOT NULL DEFAULT '0' COMMENT '用户收藏数',
  `starred_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户收藏 URL',
  `subscriptions_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户订阅 URL',
  `url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户 URL',
  `watched` int NOT NULL DEFAULT '0' COMMENT '用户关注的仓库数',
  `weibo` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户微博 URL',
  `type` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户类型',
  `gitee_created_at` varchar(30) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'gitee用户创建时间',
  `gitee_updated_at` varchar(30) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'gitee用户更新时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `gitee_user_login` (`login`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for github_user
-- ----------------------------
DROP TABLE IF EXISTS `github_user`;
CREATE TABLE `github_user` (
  `id` bigint NOT NULL DEFAULT '0' COMMENT '用户 ID',
  `login` varchar(100) COLLATE utf8mb4_general_ci NOT NULL COMMENT '登录名',
  `node_id` varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '节点ID',
  `avatar_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT '头像URL',
  `gravatar_id` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'Gravatar ID',
  `url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'GitHub URL',
  `html_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'GitHub HTML URL',
  `followers_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT '关注者URL',
  `following_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT '被关注者URL',
  `gists_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT '代码片段URL',
  `starred_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT '收藏URL',
  `subscriptions_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT '订阅URL',
  `organizations_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT '组织URL',
  `repos_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT '仓库URL',
  `events_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT '事件URL',
  `received_events_url` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL COMMENT '接收事件URL',
  `type` varchar(100) COLLATE utf8mb4_general_ci NOT NULL COMMENT '用户类型',
  `site_admin` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否为管理员',
  `name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '姓名',
  `company` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '公司',
  `blog` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '博客',
  `location` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '所在地',
  `email` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '电子邮件',
  `hireable` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否可被雇佣',
  `bio` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '个人简介',
  `twitter_username` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'Twitter 用户名',
  `public_repos` int NOT NULL DEFAULT '0' COMMENT '公共仓库数量',
  `public_gists` int NOT NULL DEFAULT '0' COMMENT '公共代码片段数量',
  `followers` int NOT NULL DEFAULT '0' COMMENT '关注者数量',
  `following` int NOT NULL DEFAULT '0' COMMENT '被关注者数量',
  `github_created_at` varchar(30) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '创建时间',
  `github_updated_at` varchar(30) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '更新时间',
  `private_gists` int NOT NULL DEFAULT '0' COMMENT '私有代码片段数量',
  `total_private_repos` int NOT NULL DEFAULT '0' COMMENT '私有仓库总数',
  `owned_private_repos` int NOT NULL DEFAULT '0' COMMENT '拥有的私有仓库数量',
  `disk_usage` int NOT NULL DEFAULT '0' COMMENT '磁盘使用量',
  `collaborators` int NOT NULL DEFAULT '0' COMMENT '协作者数量',
  `two_factor_authentication` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否启用两步验证',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `github_user_login` (`login`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for gorp_migrations
-- ----------------------------
DROP TABLE IF EXISTS `gorp_migrations`;
CREATE TABLE `gorp_migrations` (
  `id` varchar(255) NOT NULL,
  `applied_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

-- ----------------------------
-- Table structure for group
-- ----------------------------
DROP TABLE IF EXISTS `group`;
CREATE TABLE `group` (
  `id` int NOT NULL AUTO_INCREMENT,
  `group_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `name` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `creator` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `status` smallint NOT NULL DEFAULT '0',
  `forbidden` smallint NOT NULL DEFAULT '0' COMMENT '群禁言',
  `invite` smallint NOT NULL DEFAULT '0' COMMENT '群邀请开关',
  `forbidden_add_friend` smallint NOT NULL DEFAULT '0',
  `allow_view_history_msg` smallint NOT NULL DEFAULT '1',
  `version` bigint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `notice` varchar(400) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `avatar` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '群头像',
  `is_upload_avatar` smallint NOT NULL DEFAULT '0' COMMENT '群头像是否已经被用户上传',
  `group_type` smallint NOT NULL DEFAULT '0' COMMENT '群类型 0.普通群 1.超大群',
  `category` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '0' COMMENT '群分类',
  `allow_member_pinned_message` smallint NOT NULL DEFAULT '0' COMMENT '允许成员置顶聊天消息 0.不允许 1.允许',
  PRIMARY KEY (`id`),
  UNIQUE KEY `group_groupNo` (`group_no`),
  KEY `group_creator` (`creator`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for group_invite
-- ----------------------------
DROP TABLE IF EXISTS `group_invite`;
CREATE TABLE `group_invite` (
  `id` int NOT NULL AUTO_INCREMENT,
  `invite_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '邀请唯一编号',
  `group_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '群唯一编号',
  `inviter` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '邀请者uid',
  `remark` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '备注',
  `status` smallint NOT NULL DEFAULT '0' COMMENT '状态： 0.待确认 1.已确认',
  `allower` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '允许此次操作的用户uid',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for group_member
-- ----------------------------
DROP TABLE IF EXISTS `group_member`;
CREATE TABLE `group_member` (
  `id` int NOT NULL AUTO_INCREMENT,
  `group_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `remark` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `role` smallint NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  `is_deleted` smallint NOT NULL DEFAULT '0',
  `status` smallint NOT NULL DEFAULT '1',
  `vercode` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `robot` smallint NOT NULL DEFAULT '0',
  `invite_uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `forbidden_expir_time` int NOT NULL DEFAULT '0' COMMENT '群成员禁言时长',
  PRIMARY KEY (`id`),
  UNIQUE KEY `group_no_uid` (`group_no`,`uid`),
  KEY `group_member_groupNo` (`group_no`),
  KEY `group_member_uid` (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for group_setting
-- ----------------------------
DROP TABLE IF EXISTS `group_setting`;
CREATE TABLE `group_setting` (
  `id` int NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `group_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `remark` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `mute` smallint NOT NULL DEFAULT '0',
  `top` smallint NOT NULL DEFAULT '0',
  `show_nick` smallint NOT NULL DEFAULT '0',
  `save` smallint NOT NULL DEFAULT '0',
  `chat_pwd_on` smallint NOT NULL DEFAULT '0',
  `revoke_remind` smallint NOT NULL DEFAULT '1',
  `join_group_remind` smallint NOT NULL DEFAULT '0',
  `screenshot` smallint NOT NULL DEFAULT '0',
  `receipt` smallint NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `flame` smallint NOT NULL DEFAULT '0' COMMENT '阅后即焚是否开启 1.开启 0.未开启',
  `flame_second` smallint NOT NULL DEFAULT '0' COMMENT '阅后即焚销毁秒数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `groupsetting_group_no_uid` (`group_no`,`uid`),
  KEY `group_setting_groupNo` (`group_no`),
  KEY `group_setting_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for invite_item
-- ----------------------------
DROP TABLE IF EXISTS `invite_item`;
CREATE TABLE `invite_item` (
  `id` int NOT NULL AUTO_INCREMENT,
  `invite_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '邀请唯一编号',
  `group_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '群唯一编号',
  `inviter` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '邀请者uid',
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '被邀请者uid',
  `status` smallint NOT NULL DEFAULT '0' COMMENT '状态： 0.待确认 1.已确认',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for login_log
-- ----------------------------
DROP TABLE IF EXISTS `login_log`;
CREATE TABLE `login_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) NOT NULL DEFAULT '' COMMENT '用户OpenId',
  `login_ip` varchar(40) NOT NULL DEFAULT '' COMMENT '最后一次登录ip',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for member_readed
-- ----------------------------
DROP TABLE IF EXISTS `member_readed`;
CREATE TABLE `member_readed` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `clone_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `message_uid_idx` (`message_id`,`uid`),
  KEY `channel_idx` (`channel_id`,`channel_type`),
  KEY `uid_idx` (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for message
-- ----------------------------
DROP TABLE IF EXISTS `message`;
CREATE TABLE `message` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `client_msg_no` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `header` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `setting` smallint NOT NULL DEFAULT '0',
  `signal` smallint NOT NULL DEFAULT '0',
  `from_uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `timestamp` bigint NOT NULL DEFAULT '0',
  `payload` mediumblob NOT NULL,
  `is_deleted` smallint NOT NULL DEFAULT '0',
  `voice_status` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `expire` int NOT NULL DEFAULT '0' COMMENT '消息过期时长',
  `expire_at` bigint NOT NULL DEFAULT '0' COMMENT '消息过期时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `message_id` (`message_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for message1
-- ----------------------------
DROP TABLE IF EXISTS `message1`;
CREATE TABLE `message1` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `client_msg_no` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `setting` smallint NOT NULL DEFAULT '0',
  `signal` smallint NOT NULL DEFAULT '0',
  `header` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `from_uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `timestamp` bigint NOT NULL DEFAULT '0',
  `payload` mediumblob NOT NULL,
  `is_deleted` smallint NOT NULL DEFAULT '0',
  `voice_status` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `expire` int NOT NULL DEFAULT '0' COMMENT '消息过期时长',
  `expire_at` bigint NOT NULL DEFAULT '0' COMMENT '消息过期时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `message_id` (`message_id`)
) ENGINE=InnoDB AUTO_INCREMENT=36 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for message2
-- ----------------------------
DROP TABLE IF EXISTS `message2`;
CREATE TABLE `message2` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `client_msg_no` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `setting` smallint NOT NULL DEFAULT '0',
  `signal` smallint NOT NULL DEFAULT '0',
  `header` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `from_uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `timestamp` bigint NOT NULL DEFAULT '0',
  `payload` mediumblob NOT NULL,
  `is_deleted` smallint NOT NULL DEFAULT '0',
  `voice_status` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `expire` int NOT NULL DEFAULT '0' COMMENT '消息过期时长',
  `expire_at` bigint NOT NULL DEFAULT '0' COMMENT '消息过期时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `message_id` (`message_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for message3
-- ----------------------------
DROP TABLE IF EXISTS `message3`;
CREATE TABLE `message3` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `client_msg_no` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `setting` smallint NOT NULL DEFAULT '0',
  `signal` smallint NOT NULL DEFAULT '0',
  `header` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `from_uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `timestamp` bigint NOT NULL DEFAULT '0',
  `payload` mediumblob NOT NULL,
  `is_deleted` smallint NOT NULL DEFAULT '0',
  `voice_status` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `expire` int NOT NULL DEFAULT '0' COMMENT '消息过期时长',
  `expire_at` bigint NOT NULL DEFAULT '0' COMMENT '消息过期时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `message_id` (`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for message4
-- ----------------------------
DROP TABLE IF EXISTS `message4`;
CREATE TABLE `message4` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `client_msg_no` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `setting` smallint NOT NULL DEFAULT '0',
  `signal` smallint NOT NULL DEFAULT '0',
  `header` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `from_uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `timestamp` bigint NOT NULL DEFAULT '0',
  `payload` mediumblob NOT NULL,
  `is_deleted` smallint NOT NULL DEFAULT '0',
  `voice_status` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `expire` int NOT NULL DEFAULT '0' COMMENT '消息过期时长',
  `expire_at` bigint NOT NULL DEFAULT '0' COMMENT '消息过期时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `message_id` (`message_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for message_extra
-- ----------------------------
DROP TABLE IF EXISTS `message_extra`;
CREATE TABLE `message_extra` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `from_uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `revoke` smallint NOT NULL DEFAULT '0',
  `revoker` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `clone_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `version` bigint NOT NULL DEFAULT '0',
  `readed_count` int NOT NULL DEFAULT '0',
  `is_deleted` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `content_edit` text COLLATE utf8mb4_general_ci COMMENT '编辑后的正文',
  `content_edit_hash` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '编辑正文的hash值，用于重复判断',
  `edited_at` int NOT NULL DEFAULT '0' COMMENT '编辑时间 时间戳（秒）',
  `is_pinned` smallint NOT NULL DEFAULT '0' COMMENT '消息是否置顶',
  PRIMARY KEY (`id`),
  UNIQUE KEY `message_id` (`message_id`),
  KEY `from_uid_idx` (`from_uid`),
  KEY `channel_idx` (`channel_id`,`channel_type`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for message_user_extra
-- ----------------------------
DROP TABLE IF EXISTS `message_user_extra`;
CREATE TABLE `message_user_extra` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `voice_readed` smallint NOT NULL DEFAULT '0',
  `message_is_deleted` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid_message_idx` (`uid`,`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for message_user_extra1
-- ----------------------------
DROP TABLE IF EXISTS `message_user_extra1`;
CREATE TABLE `message_user_extra1` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `voice_readed` smallint NOT NULL DEFAULT '0',
  `message_is_deleted` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid_message_idx` (`uid`,`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for message_user_extra2
-- ----------------------------
DROP TABLE IF EXISTS `message_user_extra2`;
CREATE TABLE `message_user_extra2` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `voice_readed` smallint NOT NULL DEFAULT '0',
  `message_is_deleted` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid_message_idx` (`uid`,`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for pinned_message
-- ----------------------------
CREATE TABLE `pinned_message` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '消息唯一ID（全局唯一）',
  `message_seq` bigint NOT NULL DEFAULT '0' COMMENT '消息序列号(非严格递增)',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '频道ID（保持兼容性）',
  `original_channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '原始频道ID（消息实际存储的频道ID）',
  `current_channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '当前用户频道ID（发起置顶操作的用户频道ID）',
  `channel_type` smallint NOT NULL DEFAULT '0' COMMENT '频道类型',
  `is_deleted` smallint NOT NULL DEFAULT '0' COMMENT '是否已删除',
  `version` bigint NOT NULL DEFAULT '0' COMMENT '同步版本号',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `pinned_message_unique_idx` (`message_id`,`channel_id`),
  KEY `pinned_message_channelx` (`channel_id`,`channel_type`),
  KEY `pinned_message_original_channel_idx` (`original_channel_id`,`channel_type`),
  KEY `pinned_message_current_channel_idx` (`current_channel_id`,`channel_type`),
  KEY `pinned_message_message_current_idx` (`message_id`,`current_channel_id`)
) ENGINE=InnoDB AUTO_INCREMENT=161 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;-- ----------------------------
-- Table structure for prohibit_words
-- ----------------------------
DROP TABLE IF EXISTS `prohibit_words`;
CREATE TABLE `prohibit_words` (
  `id` int NOT NULL AUTO_INCREMENT,
  `is_deleted` smallint NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  `content` text COLLATE utf8mb4_general_ci,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for reaction_users
-- ----------------------------
DROP TABLE IF EXISTS `reaction_users`;
CREATE TABLE `reaction_users` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `seq` bigint NOT NULL DEFAULT '0',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `name` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `emoji` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `is_deleted` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `reaction_user_message_channel` (`message_id`,`uid`,`emoji`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for reminder_done
-- ----------------------------
DROP TABLE IF EXISTS `reminder_done`;
CREATE TABLE `reminder_done` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `reminder_id` bigint NOT NULL DEFAULT '0' COMMENT '提醒事项的id',
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '完成的用户uid',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `reminder_done_uid_reminder_id_uidx` (`uid`,`reminder_id`),
  KEY `reminder_done_reminder_id_idx` (`reminder_id`),
  KEY `reminder_done_created_at_idx` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for reminders
-- ----------------------------
DROP TABLE IF EXISTS `reminders`;
CREATE TABLE `reminders` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` smallint NOT NULL DEFAULT '0' COMMENT '频道类型',
  `reminder_type` int NOT NULL DEFAULT '0' COMMENT '提醒类型 1.有人@我 2.草稿',
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '提醒的用户uid，如果此字段为空则表示 提醒项为整个频道内的成员',
  `text` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '提醒内容',
  `data` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '自定义数据',
  `is_locate` smallint NOT NULL DEFAULT '0' COMMENT ' 是否需要定位',
  `message_seq` bigint NOT NULL DEFAULT '0' COMMENT '消息序列号',
  `message_id` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '消息唯一ID（全局唯一）',
  `version` bigint NOT NULL DEFAULT '0' COMMENT ' 数据版本',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_msg_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '消息client msg no',
  `is_deleted` smallint NOT NULL DEFAULT '0' COMMENT '是否被删除',
  `publisher` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '提醒项发布者uid',
  PRIMARY KEY (`id`),
  KEY `channel_uid_uidx` (`uid`,`channel_id`,`channel_type`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for report
-- ----------------------------
DROP TABLE IF EXISTS `report`;
CREATE TABLE `report` (
  `id` int NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '举报用户',
  `category_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '类别编号',
  `channel_id` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '频道ID',
  `channel_type` smallint NOT NULL DEFAULT '0' COMMENT '频道类型',
  `imgs` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '图片集合',
  `remark` varchar(800) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for report_category
-- ----------------------------
DROP TABLE IF EXISTS `report_category`;
CREATE TABLE `report_category` (
  `id` int NOT NULL AUTO_INCREMENT,
  `category_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '类别编号',
  `category_name` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '类别名称',
  `parent_category_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '父类别编号',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `category_ename` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '英文类别名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `report_category_no_idx` (`category_no`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for robot
-- ----------------------------
DROP TABLE IF EXISTS `robot`;
CREATE TABLE `robot` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `robot_id` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `token` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `version` bigint NOT NULL DEFAULT '0',
  `status` smallint NOT NULL DEFAULT '1',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `inline_on` smallint NOT NULL DEFAULT '0' COMMENT '是否开启行内搜索',
  `placeholder` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '输入框占位符，开启行内搜索有效',
  `username` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '机器人的username',
  `app_id` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '机器人所属app id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `robot_id_robot_index` (`robot_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for robot_menu
-- ----------------------------
DROP TABLE IF EXISTS `robot_menu`;
CREATE TABLE `robot_menu` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `robot_id` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `cmd` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `remark` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `type` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `bot_id_robot_menu_index` (`robot_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for send_history
-- ----------------------------
DROP TABLE IF EXISTS `send_history`;
CREATE TABLE `send_history` (
  `id` int NOT NULL AUTO_INCREMENT,
  `receiver` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `receiver_name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `receiver_channel_type` smallint NOT NULL DEFAULT '0',
  `sender` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `sender_name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `handler_uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `handler_name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `content` text COLLATE utf8mb4_general_ci,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for seq
-- ----------------------------
DROP TABLE IF EXISTS `seq`;
CREATE TABLE `seq` (
  `id` int NOT NULL AUTO_INCREMENT,
  `key` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `min_seq` bigint NOT NULL DEFAULT '1000000',
  `step` int NOT NULL DEFAULT '1000',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `seq_uidx` (`key`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for shortno
-- ----------------------------
DROP TABLE IF EXISTS `shortno`;
CREATE TABLE `shortno` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `shortno` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '唯一短编号',
  `used` smallint NOT NULL DEFAULT '0' COMMENT '是否被用',
  `hold` smallint NOT NULL DEFAULT '0' COMMENT '保留，保留的号码将不会再被分配',
  `locked` smallint NOT NULL DEFAULT '0' COMMENT '是否被锁定，锁定了的短编号将不再被分配,直到解锁',
  `business` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '被使用的业务，比如 user',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `udx_shortno` (`shortno`)
) ENGINE=InnoDB AUTO_INCREMENT=10232 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for signal_identities
-- ----------------------------
DROP TABLE IF EXISTS `signal_identities`;
CREATE TABLE `signal_identities` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `registration_id` bigint NOT NULL DEFAULT '0',
  `identity_key` text COLLATE utf8mb4_general_ci NOT NULL,
  `signed_prekey_id` int NOT NULL DEFAULT '0',
  `signed_pubkey` text COLLATE utf8mb4_general_ci NOT NULL,
  `signed_signature` text COLLATE utf8mb4_general_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `identities_index_id` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for signal_onetime_prekeys
-- ----------------------------
DROP TABLE IF EXISTS `signal_onetime_prekeys`;
CREATE TABLE `signal_onetime_prekeys` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `key_id` int NOT NULL DEFAULT '0',
  `pubkey` text COLLATE utf8mb4_general_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `key_id_uid_index_id` (`uid`,`key_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` int NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `short_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `short_status` smallint NOT NULL DEFAULT '0',
  `sex` smallint NOT NULL DEFAULT '0',
  `robot` smallint NOT NULL DEFAULT '0',
  `category` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `role` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `username` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `password` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `zone` varchar(20) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `phone` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `chat_pwd` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `lock_screen_pwd` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `lock_after_minute` int NOT NULL DEFAULT '0',
  `vercode` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `is_upload_avatar` smallint NOT NULL DEFAULT '0',
  `qr_vercode` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `device_lock` smallint NOT NULL DEFAULT '0',
  `search_by_phone` smallint NOT NULL DEFAULT '1',
  `search_by_short` smallint NOT NULL DEFAULT '1',
  `new_msg_notice` smallint NOT NULL DEFAULT '1',
  `msg_show_detail` smallint NOT NULL DEFAULT '1',
  `voice_on` smallint NOT NULL DEFAULT '1',
  `shock_on` smallint NOT NULL DEFAULT '1',
  `mute_of_app` smallint NOT NULL DEFAULT '0',
  `offline_protection` smallint NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  `status` smallint NOT NULL DEFAULT '1',
  `bench_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `app_id` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'app id',
  `email` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'email地址',
  `is_destroy` smallint NOT NULL DEFAULT '0' COMMENT '是否已销毁',
  `wx_openid` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '微信openid',
  `wx_unionid` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '微信unionid',
  `gitee_uid` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'gitee的用户id',
  `github_uid` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'github的用户id',
  `web3_public_key` varchar(200) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'web3公钥',
  `msg_expire_second` bigint NOT NULL DEFAULT '0' COMMENT '消息过期时长(单位秒)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid` (`uid`),
  UNIQUE KEY `short_no_udx` (`short_no`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for user_last_offset
-- ----------------------------
DROP TABLE IF EXISTS `user_last_offset`;
CREATE TABLE `user_last_offset` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_type` smallint NOT NULL DEFAULT '0',
  `message_seq` bigint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid_user_last_offset_unidx` (`uid`,`channel_id`,`channel_type`),
  KEY `uid_user_last_offset_idx` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for user_maillist
-- ----------------------------
DROP TABLE IF EXISTS `user_maillist`;
CREATE TABLE `user_maillist` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `phone` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `zone` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `name` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `vercode` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid_maillist_index` (`uid`,`zone`,`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for user_online
-- ----------------------------
DROP TABLE IF EXISTS `user_online`;
CREATE TABLE `user_online` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `device_flag` smallint NOT NULL DEFAULT '0',
  `last_online` int NOT NULL DEFAULT '0',
  `last_offline` int NOT NULL DEFAULT '0',
  `online` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uid_device` (`uid`,`device_flag`),
  KEY `online_idx` (`online`),
  KEY `uid_idx` (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for user_red_dot
-- ----------------------------
DROP TABLE IF EXISTS `user_red_dot`;
CREATE TABLE `user_red_dot` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `count` smallint NOT NULL DEFAULT '0',
  `category` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `is_dot` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_red_dot_uid_categoryx` (`uid`,`category`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for user_setting
-- ----------------------------
DROP TABLE IF EXISTS `user_setting`;
CREATE TABLE `user_setting` (
  `id` int NOT NULL AUTO_INCREMENT,
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `to_uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `mute` smallint NOT NULL DEFAULT '0',
  `top` smallint NOT NULL DEFAULT '0',
  `blacklist` smallint NOT NULL DEFAULT '0',
  `chat_pwd_on` smallint NOT NULL DEFAULT '0',
  `screenshot` smallint NOT NULL DEFAULT '1',
  `revoke_remind` smallint NOT NULL DEFAULT '1',
  `receipt` smallint NOT NULL DEFAULT '1',
  `version` bigint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `remark` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户备注',
  `flame` smallint NOT NULL DEFAULT '0' COMMENT '阅后即焚是否开启 1.开启 0.未开启',
  `flame_second` smallint NOT NULL DEFAULT '0' COMMENT '阅后即焚销毁秒数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `to_uid_uid` (`uid`,`to_uid`),
  KEY `uid_idx` (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for workplace_app
-- ----------------------------
DROP TABLE IF EXISTS `workplace_app`;
CREATE TABLE `workplace_app` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `app_id` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `icon` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `description` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `app_category` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `status` smallint NOT NULL DEFAULT '1',
  `jump_type` smallint NOT NULL DEFAULT '0',
  `app_route` varchar(200) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `web_route` varchar(200) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `is_paid_app` smallint NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `workplace_app_appid` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for workplace_app_user_record
-- ----------------------------
DROP TABLE IF EXISTS `workplace_app_user_record`;
CREATE TABLE `workplace_app_user_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `app_id` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `count` int NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `workplace_app_user_record_uid_appid` (`uid`,`app_id`),
  KEY `workplace_app_user_record_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for workplace_banner
-- ----------------------------
DROP TABLE IF EXISTS `workplace_banner`;
CREATE TABLE `workplace_banner` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `banner_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `cover` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `title` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `description` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `jump_type` smallint NOT NULL DEFAULT '0',
  `route` varchar(200) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `sort_num` int NOT NULL DEFAULT '0' COMMENT '排序号',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for workplace_category
-- ----------------------------
DROP TABLE IF EXISTS `workplace_category`;
CREATE TABLE `workplace_category` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `category_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `sort_num` int NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for workplace_category_app
-- ----------------------------
DROP TABLE IF EXISTS `workplace_category_app`;
CREATE TABLE `workplace_category_app` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `category_no` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `app_id` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `sort_num` int NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `workplace_category_app_cno_aid` (`category_no`,`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for workplace_user_app
-- ----------------------------
DROP TABLE IF EXISTS `workplace_user_app`;
CREATE TABLE `workplace_user_app` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `app_id` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `sort_num` int NOT NULL DEFAULT '0',
  `uid` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `workplace_user_app_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for admin_system_config
-- ----------------------------
DROP TABLE IF EXISTS `admin_system_config`;
CREATE TABLE `admin_system_config` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `config_key` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '配置键',
  `config_value` text COLLATE utf8mb4_general_ci COMMENT '配置值',
  `config_type` varchar(50) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'string' COMMENT '配置类型：string、number、boolean、json、image等',
  `description` varchar(500) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '配置描述',
  `category` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'system' COMMENT '配置分类：system、security、notification、ui等',
  `is_editable` smallint NOT NULL DEFAULT '1' COMMENT '是否可编辑：0.不可编辑 1.可编辑',
  `is_public` smallint NOT NULL DEFAULT '0' COMMENT '是否公开：0.私有配置 1.公开配置',
  `sort_order` int NOT NULL DEFAULT '0' COMMENT '排序顺序',
  `version` bigint NOT NULL DEFAULT '0' COMMENT '版本号',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `created_by` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '创建人UID',
  `updated_by` varchar(40) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '更新人UID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `admin_system_config_key_uidx` (`config_key`),
  KEY `admin_system_config_category_idx` (`category`),
  KEY `admin_system_config_type_idx` (`config_type`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='后台管理系统配置表';

-- ----------------------------
-- Initial data for admin_system_config
-- ----------------------------
INSERT INTO `admin_system_config` (`config_key`, `config_value`, `config_type`, `description`, `category`, `is_editable`, `is_public`, `sort_order`, `version`, `created_by`, `updated_by`) VALUES
-- 群组配置
('group.notify_member_join', '1', 'boolean', '群成员加入时是否通知其他成员', 'group', 1, 1, 100, 1, 'system', 'system'),
('group.notify_member_leave', '1', 'boolean', '群成员离开时是否通知其他成员', 'group', 1, 1, 101, 1, 'system', 'system'),
('group.notify_member_kick', '1', 'boolean', '群成员被踢出时是否通知其他成员', 'group', 1, 1, 102, 1, 'system', 'system'),
('group.new_member_see_history', '1', 'boolean', '新成员是否可见历史消息', 'group', 1, 1, 103, 1, 'system', 'system'),
('group.allow_member_invite', '1', 'boolean', '是否允许普通成员邀请新成员', 'group', 1, 1, 104, 1, 'system', 'system'),
('group.allow_member_pinned_message', '0', 'boolean', '是否允许普通成员置顶消息', 'group', 1, 1, 105, 1, 'system', 'system'),
('group.max_member_count', '500', 'number', '群组最大成员数量', 'group', 1, 1, 106, 1, 'system', 'system'),
('group.auto_delete_inactive_days', '30', 'number', '自动删除不活跃群组的天数', 'group', 1, 0, 107, 1, 'system', 'system'),

-- 系统配置
('system.site_name', '唐僧叨叨', 'string', '网站名称', 'system', 1, 1, 200, 1, 'system', 'system'),
('system.site_logo', '/assets/images/logo.png', 'image', '网站Logo图片', 'system', 1, 1, 201, 1, 'system', 'system'),
('system.site_favicon', '/assets/images/favicon.ico', 'image', '网站图标图片', 'system', 1, 1, 202, 1, 'system', 'system'),
('system.default_avatar', '/assets/images/default_avatar.png', 'image', '默认用户头像图片', 'system', 1, 1, 203, 1, 'system', 'system'),
('system.default_group_avatar', '/assets/images/default_group_avatar.png', 'image', '默认群组头像图片', 'system', 1, 1, 204, 1, 'system', 'system'),
('system.welcome_message', '欢迎使用唐僧叨叨！', 'string', '新用户注册欢迎语', 'system', 1, 1, 205, 1, 'system', 'system'),
('system.maintenance_mode', '0', 'boolean', '系统维护模式', 'system', 1, 0, 206, 1, 'system', 'system'),
('system.maintenance_message', '系统正在维护中，请稍后再试', 'string', '维护模式提示信息', 'system', 1, 0, 207, 1, 'system', 'system'),

-- 安全配置
('security.password_min_length', '6', 'number', '密码最小长度', 'security', 1, 0, 300, 1, 'system', 'system'),
('security.password_require_special_char', '0', 'boolean', '密码是否要求特殊字符', 'security', 1, 0, 301, 1, 'system', 'system'),
('security.login_fail_max_count', '5', 'number', '登录失败最大次数', 'security', 1, 0, 302, 1, 'system', 'system'),
('security.login_lock_minutes', '30', 'number', '登录锁定时间（分钟）', 'security', 1, 0, 303, 1, 'system', 'system'),
('security.session_timeout_minutes', '1440', 'number', '会话超时时间（分钟）', 'security', 1, 0, 304, 1, 'system', 'system'),
('security.allow_register', '1', 'boolean', '是否允许用户注册', 'security', 1, 0, 305, 1, 'system', 'system'),
  ('security.require_phone_verify', '0', 'boolean', '注册是否要求手机验证', 'security', 1, 0, 307, 1, 'system', 'system');

SET FOREIGN_KEY_CHECKS = 1;
