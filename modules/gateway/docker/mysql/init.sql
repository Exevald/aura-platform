CREATE DATABASE IF NOT EXISTS `gateway` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'gateway'@'%' IDENTIFIED BY 'gateway';
GRANT ALL PRIVILEGES ON `gateway`.* TO 'gateway'@'%';
