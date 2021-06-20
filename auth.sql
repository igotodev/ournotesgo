CREATE TABLE `auth` (
  `login` varchar(100) NOT NULL UNIQUE,
  `pass` varchar(255) NOT NULL,
  `cookie` varchar(255) NOT NULL,
  `time` varchar(50) NOT NULL,
  `invite` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`login`),
  UNIQUE KEY `login` (`login`)
  
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

