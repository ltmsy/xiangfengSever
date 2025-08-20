-- +migrate Up

-- 添加图片类型配置项
INSERT INTO `admin_system_config` (`config_key`, `config_value`, `config_type`, `description`, `category`, `is_editable`, `is_public`, `sort_order`, `version`, `created_by`, `updated_by`) VALUES
-- UI图片配置
('ui.login_background', '/assets/images/login_bg.jpg', 'image', '登录页面背景图片', 'ui', 1, 1, 400, 1, 'system', 'system'),
('ui.app_logo', '/assets/images/logo.png', 'image', '应用Logo图片', 'ui', 1, 1, 401, 1, 'system', 'system'),
('ui.favicon', '/assets/images/favicon.ico', 'image', '网站图标', 'ui', 1, 1, 402, 1, 'system', 'system'),
('ui.default_avatar', '/assets/images/default_avatar.png', 'image', '默认用户头像', 'ui', 1, 1, 403, 1, 'system', 'system'),
('ui.banner_image', '/assets/images/banner.png', 'image', '首页横幅图片', 'ui', 1, 1, 404, 1, 'system', 'system'),
('ui.footer_logo', '/assets/images/footer_logo.png', 'image', '页脚Logo图片', 'ui', 1, 1, 405, 1, 'system', 'system'),
('ui.chat_background', '/assets/images/chat_bg.jpg', 'image', '聊天背景图片', 'ui', 1, 1, 406, 1, 'system', 'system'),
('ui.group_avatar', '/assets/images/default_group_avatar.png', 'image', '默认群组头像', 'ui', 1, 1, 407, 1, 'system', 'system');

-- +migrate Down

-- 删除图片类型配置项
DELETE FROM `admin_system_config` WHERE `config_key` IN (
  'ui.login_background',
  'ui.app_logo', 
  'ui.favicon',
  'ui.default_avatar',
  'ui.banner_image',
  'ui.footer_logo',
  'ui.chat_background',
  'ui.group_avatar'
); 