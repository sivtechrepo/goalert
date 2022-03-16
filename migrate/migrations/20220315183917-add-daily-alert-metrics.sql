-- +migrate Up

UPDATE engine_processing_versions
SET "version" = 2
WHERE type_id = 'metrics';

CREATE TABLE daily_alert_metrics (
    id BIGSERIAL PRIMARY KEY,
    date DATE NOT NULL,
    service_id UUID NOT NULL,
    avg_time_to_ack INTERVAL,
    avg_time_to_close INTERVAL,
    escalated_count INT DEFAULT 0 NOT NULL,
    UNIQUE(service_id, date)
);

ALTER TABLE alert_metrics ADD COLUMN closed_at TIMESTAMP WITH TIME ZONE;

TRUNCATE alert_metrics;


-- +migrate Down

ALTER TABLE alert_metrics DROP COLUMN closed_at;

DROP TABLE daily_alert_metrics;

UPDATE engine_processing_versions
SET "version" = 1
WHERE type_id = 'metrics';
