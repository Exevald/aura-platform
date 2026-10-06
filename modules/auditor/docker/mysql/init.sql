CREATE DATABASE IF NOT EXISTS `auditor` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'auditor'@'%' IDENTIFIED BY 'auditor';
GRANT ALL PRIVILEGES ON `auditor`.* TO 'auditor'@'%';
