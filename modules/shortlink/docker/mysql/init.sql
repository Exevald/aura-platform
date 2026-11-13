CREATE DATABASE IF NOT EXISTS `shortlink` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'shortlink'@'%' IDENTIFIED BY 'shortlink';
GRANT ALL PRIVILEGES ON `shortlink`.* TO 'shortlink'@'%';
