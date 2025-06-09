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
    symbol       varchar,
    name         varchar,
    decimals     integer,
    total_supply numeric
);

alter table tokens
    owner to gnduser;

create table if not exists token_balances
(
    id       serial
        primary key,
    token_id integer
        references tokens,
    address  varchar,
    balance  numeric
);

alter table token_balances
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
    signature    varchar
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

