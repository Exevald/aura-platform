CREATE DATABASE IF NOT EXISTS `usergroup` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'usergroup'@'%' IDENTIFIED BY 'usergroup';
GRANT ALL PRIVILEGES ON `usergroup`.* TO 'usergroup'@'%';
