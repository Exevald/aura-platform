CREATE DATABASE auth_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE user_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE access_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE session_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE registerer_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER 'auth'@'%' IDENTIFIED BY 'integration:password@local';
CREATE USER 'user'@'%' IDENTIFIED BY 'integration:password@local';
CREATE USER 'access'@'%' IDENTIFIED BY 'integration:password@local';
CREATE USER 'session'@'%' IDENTIFIED BY 'integration:password@local';
CREATE USER 'registerer'@'%' IDENTIFIED BY 'integration:password@local';

GRANT ALL PRIVILEGES ON auth_test.* TO 'auth'@'%';
GRANT ALL PRIVILEGES ON user_test.* TO 'user'@'%';
GRANT ALL PRIVILEGES ON access_test.* TO 'access'@'%';
GRANT ALL PRIVILEGES ON session_test.* TO 'session'@'%';
GRANT ALL PRIVILEGES ON registerer_test.* TO 'registerer'@'%';
