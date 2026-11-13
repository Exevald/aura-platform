CREATE DATABASE IF NOT EXISTS `registerer` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'registerer'@'%' IDENTIFIED BY 'registerer';
GRANT ALL PRIVILEGES ON `registerer`.* TO 'registerer'@'%';
