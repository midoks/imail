SET NAMES utf8mb4;
-- ----------------------------
-- Table structure for im_mail
-- ----------------------------
DROP TABLE IF EXISTS `im_mail`;
CREATE TABLE `im_mail` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `mail_from` varchar(255) NOT NULL DEFAULT '' COMMENT '邮件来源',
  `mail_to` varchar(255) NOT NULL DEFAULT '' COMMENT '发送给谁',
  `content` text NOT NULL COMMENT '邮件内容',
  `size` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '内容大小',
  `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '状态',
  `create_time` bigint(20) unsigned NOT NULL COMMENT '创建时间',
  `update_time` bigint(20) unsigned NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COMMENT='邮件内容表';

-- ----------------------------
-- Table structure for im_user
-- ----------------------------
DROP TABLE IF EXISTS `im_user`;
CREATE TABLE `im_user` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(50) NOT NULL COMMENT '用户名',
  `password` varchar(32) NOT NULL COMMENT '密码',
  `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '状态',
  `create_time` bigint(20) unsigned NOT NULL COMMENT '创建时间',
  `update_time` bigint(20) unsigned NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- ----------------------------
-- Table structure for im_user_box
-- ----------------------------
DROP TABLE IF EXISTS `im_user_box`;
CREATE TABLE `im_user_box` (
  `id` bigint(255) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `uid` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
  `mid` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '邮件ID',
  `size` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '邮件字节大小',
  `type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '类型， 0:接收邮件;1:发送邮件',
  `create_time` bigint(20) unsigned NOT NULL COMMENT '创建时间',
  `update_time` bigint(20) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COMMENT='用户BOX';

-- ----------------------------
-- Table structure for im_class
-- ----------------------------
DROP TABLE IF EXISTS `im_class`;
CREATE TABLE `im_class` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `tag` varchar(50) NOT NULL COMMENT 'TAG',
  `name` varchar(50) NOT NULL COMMENT '名字',
  `type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '类型,0:通用:1,用户自定义',
  `uid` bigint(20) NOT NULL DEFAULT '0' COMMENT '用户ID',
  `update_time` bigint(20) NOT NULL COMMENT '更新时间',
  `create_time` bigint(20) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COMMENT='分类';

-- ----------------------------
-- Records of im_class
-- ----------------------------
BEGIN;
INSERT INTO `im_class` VALUES (1, '收件箱', 0, 0, 1560332405, 1560332405);
INSERT INTO `im_class` VALUES (2, '发件箱', 0, 0, 1560332405, 1560332405);
INSERT INTO `im_class` VALUES (3, '已删除', 0, 0, 1560332405, 1560332405);
INSERT INTO `im_class` VALUES (4, '广告邮件', 0, 0, 1560332405, 1560332405);
INSERT INTO `im_class` VALUES (5, '垃圾邮件', 0, 0, 1560332405, 1560332405);
COMMIT;

