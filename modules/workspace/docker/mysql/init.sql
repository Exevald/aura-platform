CREATE DATABASE IF NOT EXISTS `workspace` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'workspace'@'%' IDENTIFIED BY 'workspace';
GRANT ALL PRIVILEGES ON `workspace`.* TO 'workspace'@'%';
