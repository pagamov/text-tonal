CREATE TABLE IF NOT EXISTS "log_table" (
    "id" SERIAL PRIMARY KEY,
    "date" TEXT,
    "text" TEXT,
    "label" TEXT,
    "info" TEXT
);
CREATE TABLE IF NOT EXISTS "sample_table" (
    "id" SERIAL PRIMARY KEY,
    "text_en" TEXT,
    "text_ru" TEXT DEFAULT NULL,
    "label" TEXT,
    "processed" INTEGER DEFAULT 0
);
CREATE TABLE IF NOT EXISTS "usage_table" (
    "id" SERIAL PRIMARY KEY,
    "word" TEXT NOT NULL,
    "language" TEXT NOT NULL,
    "label" TEXT NOT NULL,
    "usage" INTEGER NOT NULL DEFAULT 0
);