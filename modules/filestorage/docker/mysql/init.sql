CREATE DATABASE IF NOT EXISTS `filestorage` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'filestorage'@'%' IDENTIFIED BY 'filestorage';
GRANT ALL PRIVILEGES ON `filestorage`.* TO 'filestorage'@'%';
