create sequence transactions_id_seq;

alter sequence transactions_id_seq owner to gnduser;

create table if not exists contracts
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

alter table contracts
    owner to gnduser;

create table if not exists tokens
(
    id           serial
        primary key,
    contract_id  integer
        references contracts,
    standard     varchar,
    symbol       varchar
        constraint unique_token_symbol
            unique,
    name         varchar,
    decimals     integer,
    total_supply numeric
);

alter table tokens
    owner to gnduser;

create table if not exists accounts
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

alter table accounts
    owner to gnduser;

create index if not exists idx_accounts_address
    on accounts (address);

create table if not exists api_keys
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

alter table api_keys
    owner to gnduser;

create table if not exists oracles
(
    id         serial
        primary key,
    name       varchar,
    url        varchar,
    status     varchar,
    last_check timestamp
);

alter table oracles
    owner to gnduser;

create table if not exists metrics
(
    id        serial
        primary key,
    metric    varchar,
    value     numeric,
    timestamp timestamp
);

alter table metrics
    owner to gnduser;

create table if not exists validators
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

alter table validators
    owner to gnduser;

create table if not exists pos_validators
(
    id              serial
        primary key,
    validator_id    integer not null
        references validators
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

alter table pos_validators
    owner to gnduser;

create index if not exists idx_pos_validators_validator_id
    on pos_validators (validator_id);

create table if not exists poa_validators
(
    id                  serial
        primary key,
    validator_id        integer not null
        references validators
            on delete cascade,
    legal_name          varchar,
    registration_number varchar,
    jurisdiction        varchar,
    contact_info        jsonb,
    approval_date       timestamp,
    poa_metadata        jsonb
);

alter table poa_validators
    owner to gnduser;

create index if not exists idx_poa_validators_validator_id
    on poa_validators (validator_id);

create table if not exists blocks
(
    id           serial
        primary key,
    hash         varchar
        unique,
    prev_hash    varchar,
    timestamp    timestamp,
    validator_id integer
        references validators,
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
    nonce        varchar not null
);

alter table blocks
    owner to gnduser;

create table if not exists transactions
(
    id          integer   not null,
    block_id    integer
        references blocks,
    hash        varchar,
    sender      varchar,
    recipient   varchar,
    value       numeric,
    fee         numeric,
    nonce       bigint,
    type        varchar,
    contract_id integer
        references contracts,
    payload     jsonb,
    status      varchar,
    timestamp   timestamp not null,
    primary key (id, timestamp)
)
    partition by RANGE ("timestamp");

alter table transactions
    owner to gnduser;

create index if not exists idx_transactions_hash
    on transactions (hash);

create index if not exists idx_transactions_block_id
    on transactions (block_id);

create index if not exists idx_transactions_sender
    on transactions (sender);

create index if not exists idx_transactions_recipient
    on transactions (recipient);

create index if not exists idx_transactions_timestamp
    on transactions (timestamp);

create table if not exists transactions_2025_06
    partition of transactions
        FOR VALUES FROM ('2025-06-01 00:00:00') TO ('2025-07-01 00:00:00');

alter table transactions_2025_06
    owner to gnduser;

create table if not exists transactions_2025_07
    partition of transactions
        FOR VALUES FROM ('2025-07-01 00:00:00') TO ('2025-08-01 00:00:00');

alter table transactions_2025_07
    owner to gnduser;

create table if not exists transactions_2025_08
    partition of transactions
        FOR VALUES FROM ('2025-08-01 00:00:00') TO ('2025-09-01 00:00:00');

alter table transactions_2025_08
    owner to gnduser;

create table if not exists transactions_2025_09
    partition of transactions
        FOR VALUES FROM ('2025-09-01 00:00:00') TO ('2025-10-01 00:00:00');

alter table transactions_2025_09
    owner to gnduser;

create table if not exists transactions_2025_10
    partition of transactions
        FOR VALUES FROM ('2025-10-01 00:00:00') TO ('2025-11-01 00:00:00');

alter table transactions_2025_10
    owner to gnduser;

create table if not exists transactions_2025_11
    partition of transactions
        FOR VALUES FROM ('2025-11-01 00:00:00') TO ('2025-12-01 00:00:00');

alter table transactions_2025_11
    owner to gnduser;

create table if not exists transactions_2025_12
    partition of transactions
        FOR VALUES FROM ('2025-12-01 00:00:00') TO ('2026-01-01 00:00:00');

alter table transactions_2025_12
    owner to gnduser;

create table if not exists transactions_2026_01
    partition of transactions
        FOR VALUES FROM ('2026-01-01 00:00:00') TO ('2026-02-01 00:00:00');

alter table transactions_2026_01
    owner to gnduser;

create table if not exists transactions_2026_02
    partition of transactions
        FOR VALUES FROM ('2026-02-01 00:00:00') TO ('2026-03-01 00:00:00');

alter table transactions_2026_02
    owner to gnduser;

create table if not exists transactions_2026_03
    partition of transactions
        FOR VALUES FROM ('2026-03-01 00:00:00') TO ('2026-04-01 00:00:00');

alter table transactions_2026_03
    owner to gnduser;

create table if not exists transactions_2026_04
    partition of transactions
        FOR VALUES FROM ('2026-04-01 00:00:00') TO ('2026-05-01 00:00:00');

alter table transactions_2026_04
    owner to gnduser;

create table if not exists transactions_2026_05
    partition of transactions
        FOR VALUES FROM ('2026-05-01 00:00:00') TO ('2026-06-01 00:00:00');

alter table transactions_2026_05
    owner to gnduser;

create table if not exists logs
(
    id           integer   not null,
    tx_id        integer   not null,
    tx_timestamp timestamp not null,
    contract_id  integer
        references contracts,
    event        varchar,
    data         jsonb,
    timestamp    timestamp not null,
    primary key (id, timestamp),
    foreign key (tx_id, tx_timestamp) references transactions,
    constraint logs_tx_id_tx_timestamp_fkey1
        foreign key (tx_id, tx_timestamp) references transactions_2025_06,
    constraint logs_tx_id_tx_timestamp_fkey2
        foreign key (tx_id, tx_timestamp) references transactions_2025_07,
    constraint logs_tx_id_tx_timestamp_fkey3
        foreign key (tx_id, tx_timestamp) references transactions_2025_08,
    constraint logs_tx_id_tx_timestamp_fkey4
        foreign key (tx_id, tx_timestamp) references transactions_2025_09,
    constraint logs_tx_id_tx_timestamp_fkey5
        foreign key (tx_id, tx_timestamp) references transactions_2025_10,
    constraint logs_tx_id_tx_timestamp_fkey6
        foreign key (tx_id, tx_timestamp) references transactions_2025_11,
    constraint logs_tx_id_tx_timestamp_fkey7
        foreign key (tx_id, tx_timestamp) references transactions_2025_12,
    constraint logs_tx_id_tx_timestamp_fkey8
        foreign key (tx_id, tx_timestamp) references transactions_2026_01,
    constraint logs_tx_id_tx_timestamp_fkey9
        foreign key (tx_id, tx_timestamp) references transactions_2026_02,
    constraint logs_tx_id_tx_timestamp_fkey10
        foreign key (tx_id, tx_timestamp) references transactions_2026_03,
    constraint logs_tx_id_tx_timestamp_fkey11
        foreign key (tx_id, tx_timestamp) references transactions_2026_04,
    constraint logs_tx_id_tx_timestamp_fkey12
        foreign key (tx_id, tx_timestamp) references transactions_2026_05
)
    partition by RANGE ("timestamp");

alter table logs
    owner to gnduser;

create index if not exists idx_logs_tx_id
    on logs (tx_id);

create index if not exists idx_logs_contract_id
    on logs (contract_id);

create index if not exists idx_logs_timestamp
    on logs (timestamp);

create table if not exists logs_2025_06
    partition of logs
        (
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-06-01 00:00:00') TO ('2025-07-01 00:00:00');

alter table logs_2025_06
    owner to gnduser;

create table if not exists logs_2025_07
    partition of logs
        (
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-07-01 00:00:00') TO ('2025-08-01 00:00:00');

alter table logs_2025_07
    owner to gnduser;

create table if not exists logs_2025_08
    partition of logs
        (
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-08-01 00:00:00') TO ('2025-09-01 00:00:00');

alter table logs_2025_08
    owner to gnduser;

create table if not exists logs_2025_09
    partition of logs
        (
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-09-01 00:00:00') TO ('2025-10-01 00:00:00');

alter table logs_2025_09
    owner to gnduser;

create table if not exists logs_2025_10
    partition of logs
        (
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-10-01 00:00:00') TO ('2025-11-01 00:00:00');

alter table logs_2025_10
    owner to gnduser;

create table if not exists logs_2025_11
    partition of logs
        (
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-11-01 00:00:00') TO ('2025-12-01 00:00:00');

alter table logs_2025_11
    owner to gnduser;

create table if not exists logs_2025_12
    partition of logs
        (
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-12-01 00:00:00') TO ('2026-01-01 00:00:00');

alter table logs_2025_12
    owner to gnduser;

create table if not exists logs_2026_01
    partition of logs
        (
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2026-01-01 00:00:00') TO ('2026-02-01 00:00:00');

alter table logs_2026_01
    owner to gnduser;

create table if not exists logs_2026_02
    partition of logs
        (
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2026-02-01 00:00:00') TO ('2026-03-01 00:00:00');

alter table logs_2026_02
    owner to gnduser;

create table if not exists logs_2026_03
    partition of logs
        (
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2026-03-01 00:00:00') TO ('2026-04-01 00:00:00');

alter table logs_2026_03
    owner to gnduser;

create table if not exists logs_2026_04
    partition of logs
        (
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2026-04-01 00:00:00') TO ('2026-05-01 00:00:00');

alter table logs_2026_04
    owner to gnduser;

create table if not exists logs_2026_05
    partition of logs
        (
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2026-05-01 00:00:00') TO ('2026-06-01 00:00:00');

alter table logs_2026_05
    owner to gnduser;

create table if not exists wallets
(
    id          serial
        primary key,
    account_id  integer not null
        references accounts
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

alter table wallets
    owner to gnduser;

create table if not exists token_balances
(
    token_id integer not null
        references tokens,
    address  varchar not null
        references accounts (address),
    balance  numeric,
    primary key (token_id, address)
);

alter table token_balances
    owner to gnduser;

create table if not exists events
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

comment on table events is 'Таблица для хранения событий блокчейна';

comment on column events.type is 'Тип события (Transfer, Approval, ContractDeployment, Error)';

comment on column events.contract is 'Адрес контракта';

comment on column events.from_address is 'Адрес отправителя';

comment on column events.to_address is 'Адрес получателя';

comment on column events.amount is 'Количество токенов';

comment on column events.timestamp is 'Время создания события';

comment on column events.tx_hash is 'Хеш транзакции';

comment on column events.error is 'Текст ошибки';

comment on column events.metadata is 'Дополнительные метаданные в формате JSON';

comment on column events.created_at is 'Время создания записи';

comment on column events.updated_at is 'Время последнего обновления записи';

alter table events
    owner to gnduser;

create index if not exists idx_events_contract
    on events (contract);

create index if not exists idx_events_type
    on events (type);

create index if not exists idx_events_timestamp
    on events (timestamp);

create index if not exists idx_events_tx_hash
    on events (tx_hash);

create index if not exists idx_events_from_address
    on events (from_address);

create index if not exists idx_events_to_address
    on events (to_address);

create index if not exists idx_events_contract_type
    on events (contract, type);

create index if not exists idx_events_contract_timestamp
    on events (contract, timestamp);

create or replace view latest_events
            (id, type, contract, from_address, to_address, amount, timestamp, tx_hash, error, metadata, created_at,
             updated_at, event_rank)
as
SELECT id,
       type,
       contract,
       from_address,
       to_address,
       amount,
       "timestamp",
       tx_hash,
       error,
       metadata,
       created_at,
       updated_at,
       row_number() OVER (PARTITION BY contract, type ORDER BY "timestamp" DESC) AS event_rank
FROM events e;

alter table latest_events
    owner to gnduser;

create or replace function update_updated_at_column() returns trigger
    language plpgsql
as
$$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;

alter function update_updated_at_column() owner to gnduser;

create trigger update_events_updated_at
    before update
    on events
    for each row
execute procedure update_updated_at_column();

create or replace function get_latest_events(p_contract character varying, p_type character varying, p_limit integer)
    returns TABLE(id bigint, type character varying, contract character varying, from_address character varying, to_address character varying, amount numeric, "timestamp" timestamp with time zone, tx_hash character varying, error text, metadata jsonb)
    language plpgsql
as
$$
BEGIN
    RETURN QUERY
        SELECT
            e.id,
            e.type,
            e.contract,
            e.from_address,
            e.to_address,
            e.amount,
            e."timestamp",
            e.tx_hash,
            e.error,
            e.metadata
        FROM events e
        WHERE e.contract = p_contract
          AND e.type = p_type
        ORDER BY e."timestamp" DESC
        LIMIT p_limit;
END;
$$;

alter function get_latest_events(varchar, varchar, integer) owner to gnduser;

create or replace function get_event_stats(p_contract character varying, p_start_time timestamp with time zone, p_end_time timestamp with time zone)
    returns TABLE(event_type character varying, event_count bigint, total_amount numeric, first_event timestamp with time zone, last_event timestamp with time zone)
    language plpgsql
as
$$
BEGIN
    RETURN QUERY
        SELECT
            e.type as event_type,
            COUNT(*) as event_count,
            COALESCE(SUM(e.amount), 0) as total_amount,
            MIN(e."timestamp") as first_event,
            MAX(e."timestamp") as last_event
        FROM events e
        WHERE e.contract = p_contract
          AND e."timestamp" BETWEEN p_start_time AND p_end_time
        GROUP BY e.type;
END;
$$;

alter function get_event_stats(varchar, timestamp with time zone, timestamp with time zone) owner to gnduser;

