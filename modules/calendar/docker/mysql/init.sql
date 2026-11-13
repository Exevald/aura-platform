CREATE DATABASE IF NOT EXISTS `calendar` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'calendar'@'%' IDENTIFIED BY 'calendar';
GRANT ALL PRIVILEGES ON `calendar`.* TO 'calendar'@'%';
