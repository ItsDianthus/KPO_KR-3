CREATE TABLE IF NOT EXISTS accounts (
                                        user_id TEXT PRIMARY KEY,
                                        balance NUMERIC NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS inbox (
                                     message_id TEXT PRIMARY KEY,
                                     payload JSONB NOT NULL,
                                     processed BOOLEAN NOT NULL DEFAULT FALSE,
                                     received_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS outbox (
                                      id SERIAL PRIMARY KEY,
                                      topic TEXT NOT NULL,
                                      payload JSONB NOT NULL,
                                      processed BOOLEAN NOT NULL DEFAULT FALSE,
                                      created_at TIMESTAMPTZ DEFAULT now()
);
