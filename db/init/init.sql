
-- Включаем расширение, если его нет
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Создаем таблицу valuations
CREATE TABLE IF NOT EXISTS valuations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    well_id TEXT NOT NULL,
    npv_usd DECIMAL(15, 2),
    market_value_usd DECIMAL(15, 2),
    discount_pct DECIMAL(5, 2),
    confidence DECIMAL(3, 2),
    remaining_reserves_bbl DECIMAL(15, 2),
    calculated_at TIMESTAMP WITH TIME ZONE,
    oil_price_usd DECIMAL(8, 2) DEFAULT 75.00,
    operating_cost_per_bbl DECIMAL(8, 2) DEFAULT 15.00,
    discount_rate DECIMAL(5, 4) DEFAULT 0.10,
    royalty_rate DECIMAL(5, 4) DEFAULT 0.20,
    valuation_date DATE DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создаем индекс для ускорения поиска по well_id
CREATE INDEX IF NOT EXISTS idx_valuations_well_id_btree ON valuations(well_id);

-- Создаем таблицу wells
CREATE TABLE IF NOT EXISTS wells (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    well_id TEXT NOT NULL UNIQUE,
    well_name TEXT NOT NULL,
    api_number TEXT UNIQUE,
    state TEXT NOT NULL DEFAULT 'TX',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создаем индекс для ускорения поиска по well_id в wells
CREATE INDEX IF NOT EXISTS idx_wells_well_id_btree ON wells(well_id);

-- Очищаем таблицы перед вставкой новых данных
TRUNCATE TABLE valuations, wells RESTART IDENTITY;

-- Вставляем моковые данные в wells
INSERT INTO wells (well_id, well_name, api_number) VALUES
(
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Starbelly #1H',
    '42-383-35125'
),
(
    'c2e1f3b4-8d5a-4f2e-9c6b-7d8a9b0c1d2e',
    'Midnight Oil A-25',
    '42-123-98765'
);

-- Вставляем моковые данные в valuations
INSERT INTO valuations (
    well_id, npv_usd, market_value_usd, confidence, remaining_reserves_bbl, created_at, updated_at
) VALUES
(
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', -- Этот well_id можно использовать для теста
    1200000.50,
    1100000.00,
    0.85,
    50000.00,
    '2024-07-20 10:00:00+00', -- Старая запись
    '2024-07-20 10:00:00+00'
),
(
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', -- Тот же well_id
    1250000.75,
    1150000.25,
    0.90,
    48000.00,
    '2024-07-21 11:30:00+00', -- Новая запись (эта должна вернуться по API)
    '2024-07-21 11:30:00+00'
),
(
    'c2e1f3b4-8d5a-4f2e-9c6b-7d8a9b0c1d2e', -- Другой well_id для теста
    2500000.00,
    2300000.00,
    0.95,
    120000.00,
    '2024-07-21 12:00:00+00',
    '2024-07-21 12:00:00+00'
);
