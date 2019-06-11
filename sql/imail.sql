/*
 Navicat Premium Data Transfer

 Source Server         : 172.16.1.11
 Source Server Type    : MySQL
 Source Server Version : 50621
 Source Host           : 172.16.1.11:3306
 Source Schema         : aq_petro

 Target Server Type    : MySQL
 Target Server Version : 50621
 File Encoding         : 65001

 Date: 27/05/2019 15:44:29
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for ap_device
-- ----------------------------
DROP TABLE IF EXISTS `ap_device`;
CREATE TABLE `ap_device` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `orgroot` char(50) DEFAULT NULL,
  `orgcode` char(50) DEFAULT NULL,
  `gpsno` varchar(50) NOT NULL,
  `status` tinyint(4) NOT NULL COMMENT '状态',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`,`gpsno`,`status`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=401 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for ap_logs
-- ----------------------------
DROP TABLE IF EXISTS `ap_logs`;
CREATE TABLE `ap_logs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `type` char(50) NOT NULL,
  `log` text,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`,`type`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=74 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for ap_pm
-- ----------------------------
DROP TABLE IF EXISTS `ap_pm`;
CREATE TABLE `ap_pm` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(60) NOT NULL COMMENT '姓名',
  `identity` varchar(50) DEFAULT NULL COMMENT '身份',
  `identity_type` tinyint(4) DEFAULT NULL COMMENT '身份类型 1:承包商工程师,、2承包商项目人员、3安保人员、4稽查大队、5其他',
  `orgroot` varchar(50) DEFAULT NULL COMMENT '所属机构',
  `orgcode` varchar(50) DEFAULT NULL COMMENT '子机构',
  `orgname` varchar(50) DEFAULT NULL COMMENT '机构名称',
  `carrier` varchar(50) DEFAULT NULL COMMENT '所属承运商',
  `job_number` varchar(50) DEFAULT '' COMMENT '工号',
  `id_number` varchar(50) DEFAULT '' COMMENT '身份证号',
  `telephone` varchar(50) DEFAULT NULL COMMENT '联系电话',
  `is_bind` tinyint(4) DEFAULT NULL COMMENT '是否绑定 1:绑定 2:未绑定',
  `gpsno` varchar(50) DEFAULT NULL COMMENT '卡号',
  `binding_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '绑定时间',
  `status` tinyint(4) DEFAULT '0' COMMENT '状态0:正常，1:删除',
  `created_at` datetime NOT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `identity_type` (`identity_type`),
  KEY `orgroot` (`orgroot`) USING BTREE,
  KEY `gpsno` (`gpsno`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=74 DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT;

-- ----------------------------
-- Table structure for ap_pm_type
-- ----------------------------
DROP TABLE IF EXISTS `ap_pm_type`;
CREATE TABLE `ap_pm_type` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL,
  `status` tinyint(4) DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`,`name`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;

SET FOREIGN_KEY_CHECKS = 1;
