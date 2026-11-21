CREATE DATABASE IF NOT EXISTS `publication` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'publication'@'%' IDENTIFIED BY 'publication';
GRANT ALL PRIVILEGES ON `publication`.* TO 'publication'@'%';
