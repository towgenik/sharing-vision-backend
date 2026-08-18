-- 0001_create_posts.up.sql
CREATE TABLE posts (
  `Id`           INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `Title`        VARCHAR(200) NOT NULL,
  `Content`      TEXT,
  `Category`     VARCHAR(100),
  `Created_date` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `Updated_date` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `Status`       VARCHAR(100) DEFAULT 'draft' CHECK (Status IN ('publish','draft','thrash'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
