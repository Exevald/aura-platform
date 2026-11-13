CREATE DATABASE IF NOT EXISTS `realtime` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'realtime'@'%' IDENTIFIED BY 'realtime';
GRANT ALL PRIVILEGES ON `realtime`.* TO 'realtime'@'%';
