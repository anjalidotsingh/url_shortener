-- Create the database 'url' if it doesn't exist
CREATE DATABASE IF NOT EXISTS url_shortner;

-- Use the 'url' database
USE url_shortner;

-- Create the 'url_mapping' table
CREATE TABLE IF NOT EXISTS url_mapping (
  id INT AUTO_INCREMENT PRIMARY KEY,
  actual_url VARCHAR(255) NOT NULL,
  reference_key VARCHAR(255) NOT NULL,
  UNIQUE (reference_key)
);

-- Create the 'url_count' table
CREATE TABLE IF NOT EXISTS url_count (
  id INT AUTO_INCREMENT PRIMARY KEY,
  domain_name VARCHAR(255) NOT NULL,
  count INT NOT NULL
);

-- Grant privileges to 'test' user
GRANT SELECT, INSERT, UPDATE, DELETE ON url_shortner.* TO 'test'@'%';
FLUSH PRIVILEGES;