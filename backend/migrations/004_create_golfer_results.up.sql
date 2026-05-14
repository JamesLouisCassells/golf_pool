CREATE TABLE golfer_results (
    year INT NOT NULL,
    golfer_name TEXT NOT NULL,
    position TEXT NOT NULL,
    score TEXT NOT NULL DEFAULT '',
    today TEXT NOT NULL DEFAULT '',
    thru TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (year, golfer_name)
);
