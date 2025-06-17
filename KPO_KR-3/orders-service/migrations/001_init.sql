CREATE TABLE IF NOT EXISTS orders (
                                      id SERIAL PRIMARY KEY,
                                      user_id TEXT NOT NULL,
                                      amount NUMERIC NOT NULL,
                                      status TEXT NOT NULL DEFAULT 'pending',
                                      created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS outbox (
                                      id SERIAL PRIMARY KEY,
                                      topic TEXT NOT NULL,
                                      payload JSONB NOT NULL,
                                      processed BOOLEAN NOT NULL DEFAULT FALSE,
                                      created_at TIMESTAMPTZ DEFAULT now()
);
