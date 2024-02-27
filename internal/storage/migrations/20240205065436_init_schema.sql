-- +goose Up
-- +goose StatementBegin
BEGIN;

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'absence_code') THEN
            create type absence_code AS ENUM ('ДО', 'Б', 'К', 'ОТ', 'ОД', 'У', 'НН');
        END IF;
    END
$$;

CREATE TABLE IF NOT EXISTS "users"
(
    id bigserial primary key,
    last_name varchar(50) NOT NULL,
    first_name varchar(50) NOT NULL,
    middle_name varchar(50) NOT NULL,
    birthday date NOT NULL,
    "position" varchar(150) NOT NULL,
    service_number integer NOT NULL,
    created_at timestamptz DEFAULT (now()),
    updated_at timestamptz
);

CREATE TABLE IF NOT EXISTS "absences"
(
    id bigserial primary key,
    user_id bigint NOT NULL,
    "type" absence_code NOT NULL,
    date_begin date NOT NULL,
    date_end date,
    created_at timestamptz DEFAULT (now()),
    updated_at timestamptz
);

CREATE OR REPLACE FUNCTION update_updated_at()
    RETURNS TRIGGER
    LANGUAGE plpgsql
    AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE OR REPLACE TRIGGER trigger_users_set_updated_at
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE OR REPLACE TRIGGER trigger_absences_set_updated_at
    BEFORE UPDATE
    ON absences
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

ALTER TABLE "absences"
    ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
BEGIN;

DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS absences;

DROP TYPE IF EXISTS absence_code;

DROP FUNCTION IF EXISTS update_updated_at() CASCADE;

COMMIT;
-- +goose StatementEnd
