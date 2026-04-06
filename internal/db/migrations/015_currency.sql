ALTER TABLE characters ADD COLUMN currency_balance INTEGER NOT NULL DEFAULT 0;
ALTER TABLE characters ADD COLUMN currency_label TEXT NOT NULL DEFAULT 'Gold';
