CREATE TABLE `rules` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `Object` varchar(100) NOT NULL,
  `AlarmCategoryOrder` int NOT NULL,
  `AlarmLogic` varchar(255) NOT NULL,
  `TriggerValue` varchar(255) NOT NULL,
  `AlarmCategory` varchar(255) NOT NULL,
  `AlamrMessage` varchar(255) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT now()
);

CREATE TABLE `history` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `eventID` varchar(255),
  `Object` varchar(100) NOT NULL,
  `AlarmCategory` varchar(255) NOT NULL,
  `AckMessage` varchar(255),
  `created_at` timestamp NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX `rules_index_0` ON `rules` (`Object`, `AlarmCategoryOrder`);

CREATE INDEX `history_index_1` ON `history` (`Object`);
