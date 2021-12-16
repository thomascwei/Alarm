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

CREATE TABLE `history_event` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `Object` varchar(100) NOT NULL,
  `AlarmCategoryOrder` int NOT NULL,
  `HighestAlarmCategory` varchar(255) NOT NULL,
  `AckMessage` varchar(255) NOT NULL,
  `start_time` timestamp NOT NULL DEFAULT now(),
  `end_time` timestamp
);

CREATE TABLE `history_event_detail` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `Event_id` int NOT NULL,
  `Object` varchar(100) NOT NULL,
  `AlarmCategory` varchar(255) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT now()
);

ALTER TABLE `history_event_detail` ADD FOREIGN KEY (`Event_id`) REFERENCES `history_event` (`id`);

CREATE UNIQUE INDEX `rules_index_0` ON `rules` (`Object`, `AlarmCategoryOrder`);

CREATE INDEX `history_event_index_1` ON `history_event` (`Object`);

CREATE INDEX `history_event_detail_index_2` ON `history_event_detail` (`id`);
