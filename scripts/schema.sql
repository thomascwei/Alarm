USE alarm;
CREATE TABLE `rules` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `Object` varchar(100) NOT NULL,
  `AlarmCategoryOrder` int NOT NULL,
  `AlarmLogic` varchar(10) NOT NULL,
  `TriggerValue` varchar(100) NOT NULL,
  `AlarmCategory` varchar(100) NOT NULL,
  `AlarmMessage` varchar(255) NOT NULL,
  `AckMethod` varchar(100) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT now()
);

CREATE TABLE `history_event` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `Object` varchar(100) NOT NULL,
  `AlarmCategoryOrder` int NOT NULL,
  `HighestAlarmCategory` varchar(100) NOT NULL,
  `AlarmMessage` varchar(255) NOT NULL,
  `AckMessage` varchar(255) NOT NULL,
  `start_time` datetime NOT NULL DEFAULT now(),
  `end_time` datetime
);

CREATE TABLE `history_event_detail` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `Event_id` int NOT NULL,
  `Object` varchar(100) NOT NULL,
  `AlarmCategory` varchar(100) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT now()
);

ALTER TABLE `history_event_detail` ADD FOREIGN KEY (`Event_id`) REFERENCES `history_event` (`id`);

CREATE UNIQUE INDEX `rules_index_0` ON `rules` (`Object`, `AlarmCategoryOrder`);

CREATE INDEX `history_event_index_1` ON `history_event` (`Object`);

CREATE INDEX `history_event_detail_index_2` ON `history_event_detail` (`id`);
