-- +goose Up
-- +goose StatementBegin
BEGIN;
ALTER TYPE absence_code RENAME TO absence_code_old;
CREATE TYPE absence_code AS ENUM ('К', 'ОТ', 'ОД', 'У', 'ДО', 'Б', 'НН');
ALTER TABLE absences ALTER COLUMN "type" TYPE absence_code USING "type"::text::absence_code;
DROP TYPE absence_code_old;
COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'No changes';
-- +goose StatementEnd
