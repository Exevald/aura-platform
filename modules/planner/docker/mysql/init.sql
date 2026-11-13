CREATE DATABASE IF NOT EXISTS `planner` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'planner'@'%' IDENTIFIED BY 'planner';
GRANT ALL PRIVILEGES ON `planner`.* TO 'planner'@'%';
