create table if not exists public.contracts
(
    id         serial
        primary key,
    address    varchar
        unique,
    owner      varchar,
    code       bytea,
    abi        jsonb,
    created_at timestamp,
    type       varchar
);

alter table public.contracts
    owner to gnduser;

create table if not exists public.tokens
(
    id           serial
        primary key,
    contract_id  integer
        references public.contracts,
    standard     varchar,
    symbol       varchar
        constraint unique_token_symbol
            unique,
    name         varchar,
    decimals     integer,
    total_supply numeric
);

alter table public.tokens
    owner to gnduser;

create table if not exists public.accounts
(
    id         serial
        primary key,
    address    varchar
        unique,
    balance    numeric,
    nonce      bigint,
    stake      numeric,
    created_at timestamp
);

alter table public.accounts
    owner to gnduser;

create index if not exists idx_accounts_address
    on public.accounts (address);

create table if not exists public.api_keys
(
    id          serial
        primary key,
    user_id     integer,
    key         varchar
        unique,
    permissions jsonb,
    created_at  timestamp,
    expires_at  timestamp
);

alter table public.api_keys
    owner to gnduser;

create table if not exists public.oracles
(
    id         serial
        primary key,
    name       varchar,
    url        varchar,
    status     varchar,
    last_check timestamp
);

alter table public.oracles
    owner to gnduser;

create table if not exists public.metrics
(
    id        serial
        primary key,
    metric    varchar,
    value     numeric,
    timestamp timestamp
);

alter table public.metrics
    owner to gnduser;

create table if not exists public.validators
(
    id             serial
        primary key,
    address        varchar
        unique,
    pubkey         varchar,
    stake          numeric,
    status         varchar,
    last_active    timestamp,
    consensus_type varchar not null,
    authority_info jsonb,
    metadata       jsonb
);

alter table public.validators
    owner to gnduser;

create table if not exists public.pos_validators
(
    id              serial
        primary key,
    validator_id    integer not null
        references public.validators
            on delete cascade,
    total_stake     numeric,
    delegated_stake numeric,
    commission_rate numeric,
    uptime          numeric,
    slashing_events integer,
    rewards         numeric,
    last_reward_at  timestamp,
    pos_metadata    jsonb
);

alter table public.pos_validators
    owner to gnduser;

create index if not exists idx_pos_validators_validator_id
    on public.pos_validators (validator_id);

create table if not exists public.poa_validators
(
    id                  serial
        primary key,
    validator_id        integer not null
        references public.validators
            on delete cascade,
    legal_name          varchar,
    registration_number varchar,
    jurisdiction        varchar,
    contact_info        jsonb,
    approval_date       timestamp,
    poa_metadata        jsonb
);

alter table public.poa_validators
    owner to gnduser;

create index if not exists idx_poa_validators_validator_id
    on public.poa_validators (validator_id);

create table if not exists public.blocks
(
    id           serial
        primary key,
    hash         varchar
        unique,
    prev_hash    varchar,
    timestamp    timestamp,
    validator_id integer
        references public.validators,
    tx_count     integer,
    state_root   varchar,
    signature    varchar,
    index        integer not null
        constraint unique_block_index
            unique,
    miner        varchar not null,
    gas_used     bigint  not null,
    gas_limit    bigint  not null,
    consensus    varchar not null,
    nonce        varchar not null,
    reward       numeric(78,0) not null default 0
);

alter table public.blocks
    owner to gnduser;

create table if not exists public.transactions
(
    id          integer   not null,
    block_id    integer
        references public.blocks,
    hash        varchar,
    sender      varchar,
    recipient   varchar,
    value       numeric,
    fee         numeric,
    nonce       bigint,
    type        varchar,
    contract_id integer
        references public.contracts,
    payload     jsonb,
    status      varchar,
    timestamp   timestamp not null,
    primary key (id, timestamp)
)
    partition by RANGE ("timestamp");

alter table public.transactions
    owner to gnduser;

create index if not exists idx_transactions_hash
    on public.transactions (hash);

create index if not exists idx_transactions_block_id
    on public.transactions (block_id);

create index if not exists idx_transactions_sender
    on public.transactions (sender);

create index if not exists idx_transactions_recipient
    on public.transactions (recipient);

create index if not exists idx_transactions_timestamp
    on public.transactions (timestamp);

create table if not exists public.transactions_2025_06
    partition of public.transactions
        FOR VALUES FROM ('2025-06-01 00:00:00') TO ('2025-07-01 00:00:00');

alter table public.transactions_2025_06
    owner to gnduser;

create table if not exists public.transactions_2025_07
    partition of public.transactions
        FOR VALUES FROM ('2025-07-01 00:00:00') TO ('2025-08-01 00:00:00');

alter table public.transactions_2025_07
    owner to gnduser;

create table if not exists public.transactions_2025_08
    partition of public.transactions
        FOR VALUES FROM ('2025-08-01 00:00:00') TO ('2025-09-01 00:00:00');

alter table public.transactions_2025_08
    owner to gnduser;

create table if not exists public.transactions_2025_09
    partition of public.transactions
        FOR VALUES FROM ('2025-09-01 00:00:00') TO ('2025-10-01 00:00:00');

alter table public.transactions_2025_09
    owner to gnduser;

create table if not exists public.transactions_2025_10
    partition of public.transactions
        FOR VALUES FROM ('2025-10-01 00:00:00') TO ('2025-11-01 00:00:00');

alter table public.transactions_2025_10
    owner to gnduser;

create table if not exists public.transactions_2025_11
    partition of public.transactions
        FOR VALUES FROM ('2025-11-01 00:00:00') TO ('2025-12-01 00:00:00');

alter table public.transactions_2025_11
    owner to gnduser;

create table if not exists public.transactions_2025_12
    partition of public.transactions
        FOR VALUES FROM ('2025-12-01 00:00:00') TO ('2026-01-01 00:00:00');

alter table public.transactions_2025_12
    owner to gnduser;

create table if not exists public.transactions_2026_01
    partition of public.transactions
        FOR VALUES FROM ('2026-01-01 00:00:00') TO ('2026-02-01 00:00:00');

alter table public.transactions_2026_01
    owner to gnduser;

create table if not exists public.transactions_2026_02
    partition of public.transactions
        FOR VALUES FROM ('2026-02-01 00:00:00') TO ('2026-03-01 00:00:00');

alter table public.transactions_2026_02
    owner to gnduser;

create table if not exists public.transactions_2026_03
    partition of public.transactions
        FOR VALUES FROM ('2026-03-01 00:00:00') TO ('2026-04-01 00:00:00');

alter table public.transactions_2026_03
    owner to gnduser;

create table if not exists public.transactions_2026_04
    partition of public.transactions
        FOR VALUES FROM ('2026-04-01 00:00:00') TO ('2026-05-01 00:00:00');

alter table public.transactions_2026_04
    owner to gnduser;

create table if not exists public.transactions_2026_05
    partition of public.transactions
        FOR VALUES FROM ('2026-05-01 00:00:00') TO ('2026-06-01 00:00:00');

alter table public.transactions_2026_05
    owner to gnduser;

create table if not exists public.logs
(
    id           integer   not null,
    tx_id        integer   not null,
    tx_timestamp timestamp not null,
    contract_id  integer
        references public.contracts,
    event        varchar,
    data         jsonb,
    timestamp    timestamp not null,
    primary key (id, timestamp),
    foreign key (tx_id, tx_timestamp) references public.transactions
)
    partition by RANGE ("timestamp");

alter table public.logs
    owner to gnduser;

create index if not exists idx_logs_tx_id
    on public.logs (tx_id);

create index if not exists idx_logs_contract_id
    on public.logs (contract_id);

create index if not exists idx_logs_timestamp
    on public.logs (timestamp);

create table if not exists public.logs_2025_06
    partition of public.logs
        FOR VALUES FROM ('2025-06-01 00:00:00') TO ('2025-07-01 00:00:00');

alter table public.logs_2025_06
    owner to gnduser;

create table if not exists public.logs_2025_07
    partition of public.logs
        FOR VALUES FROM ('2025-07-01 00:00:00') TO ('2025-08-01 00:00:00');

alter table public.logs_2025_07
    owner to gnduser;

create table if not exists public.logs_2025_08
    partition of public.logs
        FOR VALUES FROM ('2025-08-01 00:00:00') TO ('2025-09-01 00:00:00');

alter table public.logs_2025_08
    owner to gnduser;

create table if not exists public.logs_2025_09
    partition of public.logs
        FOR VALUES FROM ('2025-09-01 00:00:00') TO ('2025-10-01 00:00:00');

alter table public.logs_2025_09
    owner to gnduser;

create table if not exists public.logs_2025_10
    partition of public.logs
        FOR VALUES FROM ('2025-10-01 00:00:00') TO ('2025-11-01 00:00:00');

alter table public.logs_2025_10
    owner to gnduser;

create table if not exists public.logs_2025_11
    partition of public.logs
        FOR VALUES FROM ('2025-11-01 00:00:00') TO ('2025-12-01 00:00:00');

alter table public.logs_2025_11
    owner to gnduser;

create table if not exists public.logs_2025_12
    partition of public.logs
        FOR VALUES FROM ('2025-12-01 00:00:00') TO ('2026-01-01 00:00:00');

alter table public.logs_2025_12
    owner to gnduser;

create table if not exists public.logs_2026_01
    partition of public.logs
        FOR VALUES FROM ('2026-01-01 00:00:00') TO ('2026-02-01 00:00:00');

alter table public.logs_2026_01
    owner to gnduser;

create table if not exists public.logs_2026_02
    partition of public.logs
        FOR VALUES FROM ('2026-02-01 00:00:00') TO ('2026-03-01 00:00:00');

alter table public.logs_2026_02
    owner to gnduser;

create table if not exists public.logs_2026_03
    partition of public.logs
        FOR VALUES FROM ('2026-03-01 00:00:00') TO ('2026-04-01 00:00:00');

alter table public.logs_2026_03
    owner to gnduser;

create table if not exists public.logs_2026_04
    partition of public.logs
        FOR VALUES FROM ('2026-04-01 00:00:00') TO ('2026-05-01 00:00:00');

alter table public.logs_2026_04
    owner to gnduser;

create table if not exists public.logs_2026_05
    partition of public.logs
        FOR VALUES FROM ('2026-05-01 00:00:00') TO ('2026-06-01 00:00:00');

alter table public.logs_2026_05
    owner to gnduser;

create table if not exists public.wallets
(
    id          serial
        primary key,
    account_id  integer not null
        references public.accounts
            on delete cascade,
    address     varchar not null
        unique,
    public_key  varchar not null
        unique,
    private_key varchar not null,
    created_at  timestamp default now(),
    updated_at  timestamp default now(),
    status      varchar   default 'active'::character varying
);

alter table public.wallets
    owner to gnduser;

create table if not exists public.token_balances
(
    token_id integer not null
        references public.tokens,
    address  varchar not null
        references public.accounts (address),
    balance  numeric,
    primary key (token_id, address)
);

alter table public.token_balances
    owner to gnduser;

create table if not exists public.events
(
    id           bigserial
        primary key,
    type         varchar(50)                                        not null,
    contract     varchar(42)                                        not null,
    from_address varchar(42),
    to_address   varchar(42),
    amount       numeric(78),
    timestamp    timestamp with time zone default CURRENT_TIMESTAMP not null,
    tx_hash      varchar(66),
    error        text,
    metadata     jsonb,
    created_at   timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at   timestamp with time zone default CURRENT_TIMESTAMP not null
);

comment on table public.events is 'Таблица для хранения событий блокчейна';

comment on column public.events.type is 'Тип события (Transfer, Approval, ContractDeployment, Error)';

comment on column public.events.contract is 'Адрес контракта';

comment on column public.events.from_address is 'Адрес отправителя';

comment on column public.events.to_address is 'Адрес получателя';

comment on column public.events.amount is 'Количество токенов';

comment on column public.events.timestamp is 'Время создания события';

comment on column public.events.tx_hash is 'Хеш транзакции';

comment on column public.events.error is 'Текст ошибки';

comment on column public.events.metadata is 'Дополнительные метаданные в формате JSON';

comment on column public.events.created_at is 'Время создания записи';

comment on column public.events.updated_at is 'Время последнего обновления записи';

alter table public.events
    owner to gnduser;

create index if not exists idx_events_contract
    on public.events (contract);

create index if not exists idx_events_type
    on public.events (type);

create index if not exists idx_events_timestamp
    on public.events (timestamp);

create index if not exists idx_events_tx_hash
    on public.events (tx_hash);

create index if not exists idx_events_from_address
    on public.events (from_address);

create index if not exists idx_events_to_address
    on public.events (to_address);

create index if not exists idx_events_contract_type
    on public.events (contract, type);

create index if not exists idx_events_contract_timestamp
    on public.events (contract, timestamp);

create trigger update_events_updated_at
    before update
    on public.events
    for each row
execute procedure public.update_updated_at_column();

-- Индексы для оптимизации запросов
create index if not exists idx_blocks_timestamp
    on public.blocks (timestamp);

create index if not exists idx_contracts_created_at
    on public.contracts (created_at);

create index if not exists idx_validators_last_active
    on public.validators (last_active);

create index if not exists idx_token_balances_address
    on public.token_balances (address);

create index if not exists idx_token_balances_token_id
    on public.token_balances (token_id);

