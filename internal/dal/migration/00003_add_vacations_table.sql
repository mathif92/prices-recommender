-- +goose Up
CREATE TABLE IF NOT EXISTS "vacations" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    affected_levels TEXT NOT NULL,
    year INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT vacations_name_year_unique UNIQUE (name, year)
);

CREATE INDEX IF NOT EXISTS idx_vacations_year ON vacations(year);
CREATE INDEX IF NOT EXISTS idx_vacations_start_date ON vacations(start_date);

-- +goose Down
DROP TABLE IF EXISTS "vacations";
