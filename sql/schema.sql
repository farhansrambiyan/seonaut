# ************************************************************
# Sequel Ace SQL dump
# Version 20029
#
# https://sequel-ace.com/
# https://github.com/Sequel-Ace/Sequel-Ace
#
# Host: 0.0.0.0 (MySQL 8.0.28)
# Database: seo
# Generation Time: 2022-02-25 11:50:05 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE='NO_AUTO_VALUE_ON_ZERO', SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


# Dump of table crawls
# ------------------------------------------------------------

DROP TABLE IF EXISTS `crawls`;

CREATE TABLE `crawls` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `project_id` int unsigned NOT NULL,
  `start` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `end` timestamp NULL DEFAULT NULL,
  `total_urls` int NOT NULL DEFAULT '0',
  `total_issues` int NOT NULL DEFAULT '0',
  `issues_end` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `crawl_project` (`project_id`),
  CONSTRAINT `crawl_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# Dump of table external_links
# ------------------------------------------------------------

DROP TABLE IF EXISTS `external_links`;

CREATE TABLE `external_links` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `crawl_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `rel` varchar(100) DEFAULT NULL,
  `text` varchar(1024) DEFAULT NULL,
  `nofollow` tinyint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `external_links_pagereport` (`pagereport_id`),
  KEY `external_links_crawl` (`crawl_id`),
  CONSTRAINT `external_links_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `external_links_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# Dump of table hreflangs
# ------------------------------------------------------------

DROP TABLE IF EXISTS `hreflangs`;

CREATE TABLE `hreflangs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `crawl_id` int unsigned NOT NULL,
  `from_lang` varchar(10) DEFAULT NULL,
  `to_url` varchar(2048) NOT NULL DEFAULT '',
  `to_lang` varchar(10) DEFAULT NULL,
  `from_hash` varchar(256) NOT NULL DEFAULT '',
  `to_hash` varchar(256) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `hreflangs_from_hash` (`from_hash`),
  KEY `hreflangs_to_hash` (`to_hash`),
  KEY `hreflangs_pagereport` (`pagereport_id`),
  KEY `hreflangs_crawl` (`crawl_id`),
  KEY `hreflangs_crawl_from_to` (`crawl_id`,`from_hash`,`to_hash`),
  CONSTRAINT `hreflangs_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `hreflangs_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# Dump of table images
# ------------------------------------------------------------

DROP TABLE IF EXISTS `images`;

CREATE TABLE `images` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `alt` varchar(1024) DEFAULT NULL,
  `crawl_id` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `images_pagereport` (`pagereport_id`),
  KEY `images_crawl` (`crawl_id`),
  CONSTRAINT `images_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `images_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# Dump of table issue_types
# ------------------------------------------------------------

DROP TABLE IF EXISTS `issue_types`;

CREATE TABLE `issue_types` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(256) DEFAULT NULL,
  `priority` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

LOCK TABLES `issue_types` WRITE;
/*!40000 ALTER TABLE `issue_types` DISABLE KEYS */;

INSERT INTO `issue_types` (`id`, `type`, `priority`)
VALUES
	(1,'ERROR_30x',1),
	(2,'ERROR_40x',1),
	(3,'ERROR_50x',1),
	(4,'ERROR_DUPLICATED_TITLE',2),
	(5,'ERROR_DUPLICATED_DESCRIPTION',2),
	(6,'ERROR_EMPTY_TITLE',2),
	(7,'ERROR_SHORT_TITLE',2),
	(8,'ERROR_LONG_TITLE',2),
	(9,'ERROR_EMPTY_DESCRIPTION',2),
	(10,'ERROR_SHORT_DESCRIPTION',2),
	(11,'ERROR_LONG_DESCRIPTION',2),
	(12,'ERROR_LITTLE_CONTENT',3),
	(13,'ERROR_IMAGES_NO_ALT',2),
	(14,'ERROR_REDIRECT_CHAIN',1),
	(15,'ERROR_NO_H1',2),
	(16,'ERROR_NO_LANG',3),
	(17,'ERROR_HTTP_LINKS',2),
	(18,'ERROR_HREFLANG_RETURN',2),
	(19,'ERROR_TOO_MANY_LINKS',3),
	(20,'ERROR_INTERNAL_NOFOLLOW',3),
	(21,'ERROR_EXTERNAL_WITHOUT_NOFOLLOW',3),
	(22,'ERROR_CANONICALIZED_NON_CANONICAL',2),
	(23,'ERROR_REDIRECT_LOOP',1),
	(24,'ERROR_NOT_VALID_HEADINGS',2);

/*!40000 ALTER TABLE `issue_types` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table issues
# ------------------------------------------------------------

DROP TABLE IF EXISTS `issues`;

CREATE TABLE `issues` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `crawl_id` int unsigned NOT NULL,
  `issue_type_id` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `issue_crawl` (`crawl_id`),
  KEY `issue_pagereport` (`pagereport_id`),
  KEY `issues_issue_type` (`issue_type_id`),
  CONSTRAINT `issue_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `issue_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE,
  CONSTRAINT `issues_issue_type` FOREIGN KEY (`issue_type_id`) REFERENCES `issue_types` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# Dump of table links
# ------------------------------------------------------------

DROP TABLE IF EXISTS `links`;

CREATE TABLE `links` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `crawl_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `scheme` varchar(5) NOT NULL,
  `rel` varchar(100) DEFAULT NULL,
  `text` varchar(1024) DEFAULT NULL,
  `url_hash` varchar(256) NOT NULL DEFAULT '',
  `nofollow` tinyint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `links_external` (`pagereport_id`),
  KEY `links_hash` (`url_hash`),
  KEY `links_crawl` (`crawl_id`),
  CONSTRAINT `links_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `links_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# Dump of table pagereports
# ------------------------------------------------------------

DROP TABLE IF EXISTS `pagereports`;

CREATE TABLE `pagereports` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `crawl_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `scheme` varchar(5) DEFAULT NULL,
  `redirect_url` varchar(2048) DEFAULT NULL,
  `refresh` varchar(2048) DEFAULT NULL,
  `status_code` int NOT NULL,
  `content_type` varchar(100) DEFAULT NULL,
  `media_type` varchar(100) DEFAULT NULL,
  `lang` varchar(10) DEFAULT NULL,
  `title` varchar(2048) DEFAULT NULL,
  `description` varchar(2048) DEFAULT NULL,
  `robots` varchar(100) DEFAULT NULL,
  `canonical` varchar(2048) DEFAULT NULL,
  `h1` varchar(1024) DEFAULT NULL,
  `h2` varchar(1024) DEFAULT NULL,
  `words` int DEFAULT NULL,
  `size` int DEFAULT NULL,
  `url_hash` varchar(256) NOT NULL DEFAULT '',
  `redirect_hash` varchar(256) DEFAULT NULL,
  `valid_headings` tinyint NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `pagereport_crawl` (`crawl_id`),
  KEY `pagereport_hash` (`url_hash`),
  KEY `redirect_hash` (`redirect_hash`),
  CONSTRAINT `pagereport_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# Dump of table projects
# ------------------------------------------------------------

DROP TABLE IF EXISTS `projects`;

CREATE TABLE `projects` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int unsigned DEFAULT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `ignore_robotstxt` tinyint NOT NULL DEFAULT '0',
  `use_javascript` tinyint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  CONSTRAINT `projects_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# Dump of table scripts
# ------------------------------------------------------------

DROP TABLE IF EXISTS `scripts`;

CREATE TABLE `scripts` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `crawl_id` int unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `scripts_pagereport` (`pagereport_id`),
  KEY `scripts_crawl` (`crawl_id`),
  CONSTRAINT `scripts_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `scripts_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# Dump of table styles
# ------------------------------------------------------------

DROP TABLE IF EXISTS `styles`;

CREATE TABLE `styles` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `pagereport_id` int unsigned NOT NULL,
  `url` varchar(2048) NOT NULL DEFAULT '',
  `crawl_id` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `styles_pagereport` (`pagereport_id`),
  KEY `styles_crawl` (`crawl_id`),
  CONSTRAINT `styles_crawl` FOREIGN KEY (`crawl_id`) REFERENCES `crawls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `styles_pagereport` FOREIGN KEY (`pagereport_id`) REFERENCES `pagereports` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;



# Dump of table users
# ------------------------------------------------------------

DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `email` varchar(256) NOT NULL DEFAULT '',
  `password` varchar(512) NOT NULL DEFAULT '',
  `stripe_session_id` varchar(256) DEFAULT NULL,
  `stripe_customer_id` varchar(256) DEFAULT NULL,
  `period_end` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;




/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
