-- ============================================================
-- Sharing Vision test — DATABASE MANUAL TASK (80 pts)
-- Run against the `article` database:  mysql article < manual_create.sql
-- Creates the posts table AND inserts two sample rows.
-- ============================================================
USE article;

CREATE TABLE posts (
  `Id`           INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `Title`        VARCHAR(200) NOT NULL,
  `Content`      TEXT,
  `Category`     VARCHAR(100),
  `Created_date` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `Updated_date` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `Status`       VARCHAR(100) DEFAULT 'draft' CHECK (Status IN ('publish','draft','thrash'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO posts (Title, Content, Category, Status) VALUES
  ('Welcome to ShareVision', 'Hello from the demo database. This row was inserted by the manual SQL script.', 'Welcome', 'publish'),
  ('Draft ideas', 'A private draft article. It should not appear on the public preview.', 'Draft', 'draft');
