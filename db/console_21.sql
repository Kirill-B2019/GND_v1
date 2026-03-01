-- Текущая консолидированная схема БД ноды GND_v1.
-- Эквивалент: базовая схема + db/002_schema_additions.sql + migrations 001, 004–009.
-- Используется как единый источник истины для развёртывания с нуля.
-- Миграции (001, 004–009) добавляют: events, signer_wallets, доработки wallets/api_keys/tokens и т.д.

create sequence logs_id_seq;

alter sequence logs_id_seq owner to gnduser;

create sequence transactions_id_seq;

alter sequence transactions_id_seq owner to gnduser;

create table contracts
(
    id           serial
        primary key,
    address      varchar
        unique,
    owner        varchar,
    code         bytea,
    abi          jsonb,
    created_at   timestamp,
    type         varchar,
    creator      varchar,
    bytecode     bytea,
    name         varchar,
    symbol       varchar,
    standard     varchar,
    description  text,
    version      varchar,
    status       varchar,
    block_id     integer,
    tx_id        integer,
    gas_limit    bigint,
    gas_used     bigint,
    value        varchar,
    data         bytea,
    updated_at   timestamp,
    is_verified  boolean default false,
    source_code  text,
    compiler     varchar,
    optimized    boolean default false,
    runs         integer,
    license      varchar,
    metadata     jsonb,
    params       jsonb,
    metadata_cid varchar
);

alter table contracts
    owner to gnduser;

create index idx_contracts_created_at
    on contracts (created_at);

create table tokens
(
    id                 serial
        primary key,
    contract_id        integer
        references contracts,
    standard           varchar,
    symbol             varchar
        constraint unique_token_symbol
            unique,
    name               varchar,
    decimals           integer,
    total_supply       numeric,
    status             varchar,
    updated_at         timestamp,
    is_verified        boolean default false,
    circulating_supply numeric,
    logo_url          varchar(512)
);

comment on column tokens.circulating_supply is 'Обращающееся предложение; заполняется из config coins.circulating_supply';
comment on column tokens.logo_url is 'URL или путь к логотипу токена (до 250x250 px, изображение)';

alter table tokens
    owner to gnduser;

create table accounts
(
    id          serial
        primary key,
    address     varchar
        unique,
    balance     numeric,
    nonce       bigint,
    stake       numeric,
    created_at  timestamp,
    type        varchar,
    status      varchar,
    block_id    integer,
    tx_id       integer,
    gas_limit   bigint,
    gas_used    bigint,
    value       varchar,
    data        bytea,
    updated_at  timestamp,
    is_verified boolean default false,
    source_code text,
    compiler    varchar,
    optimized   boolean default false,
    runs        integer,
    license     varchar,
    metadata    jsonb
);

alter table accounts
    owner to gnduser;

create index idx_accounts_address
    on accounts (address);

create table api_keys
(
    id          serial
        primary key,
    user_id     integer,
    key         varchar
        unique,
    permissions jsonb,
    created_at  timestamp,
    expires_at  timestamp,
    name        varchar(255),
    key_prefix  varchar(16),
    key_hash    varchar(64)
        unique,
    disabled    boolean default false not null
);

comment on column api_keys.name is 'Человекочитаемое имя ключа (например: Laravel Backend)';

comment on column api_keys.key_prefix is 'Префикс ключа для отображения в списке (например gnd_ab12)';

comment on column api_keys.key_hash is 'SHA-256 хеш ключа в hex; для проверки без хранения открытого ключа';

comment on column api_keys.disabled is 'Отозванный ключ не принимается';

alter table api_keys
    owner to gnduser;

create index idx_api_keys_key_hash
    on api_keys (key_hash)
    where (key_hash IS NOT NULL);

create index idx_api_keys_disabled
    on api_keys (disabled)
    where (disabled = false);

create table oracles
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

create table metrics
(
    id        serial
        primary key,
    metric    varchar,
    value     numeric,
    timestamp timestamp
);

alter table metrics
    owner to gnduser;

create table validators
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

create index idx_validators_last_active
    on validators (last_active);

create table pos_validators
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

create index idx_pos_validators_validator_id
    on pos_validators (validator_id);

create table poa_validators
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

create index idx_poa_validators_validator_id
    on poa_validators (validator_id);

create table blocks
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
    index        integer               not null
        constraint unique_block_index
            unique,
    miner        varchar               not null,
    gas_used     bigint                not null,
    gas_limit    bigint                not null,
    consensus    varchar               not null,
    nonce        varchar               not null,
    reward       numeric(78) default 0 not null,
    merkle_root  varchar,
    height       bigint,
    version      integer     default 1,
    size         bigint,
    difficulty   bigint,
    extra_data   bytea,
    created_at   timestamp,
    updated_at   timestamp,
    status       varchar,
    parent_id    bigint,
    is_orphaned  boolean     default false,
    is_finalized boolean     default false
);

alter table blocks
    owner to gnduser;

create index idx_blocks_timestamp
    on blocks (timestamp);

create table transactions
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
    signature   varchar,
    is_verified boolean default false,
    primary key (id, timestamp)
)
    partition by RANGE ("timestamp");

alter table transactions
    owner to gnduser;

create index idx_transactions_hash
    on transactions (hash);

create index idx_transactions_block_id
    on transactions (block_id);

create index idx_transactions_sender
    on transactions (sender);

create index idx_transactions_recipient
    on transactions (recipient);

create index idx_transactions_timestamp
    on transactions (timestamp);

create table transactions_2025_06
    partition of transactions
        (
            constraint transactions_block_id_fkey
                foreign key (block_id) references blocks,
            constraint transactions_contract_id_fkey
                foreign key (contract_id) references contracts
            )
        FOR VALUES FROM ('2025-06-01 00:00:00') TO ('2025-07-01 00:00:00');

alter table transactions_2025_06
    owner to gnduser;

create table transactions_2025_07
    partition of transactions
        (
            constraint transactions_block_id_fkey
                foreign key (block_id) references blocks,
            constraint transactions_contract_id_fkey
                foreign key (contract_id) references contracts
            )
        FOR VALUES FROM ('2025-07-01 00:00:00') TO ('2025-08-01 00:00:00');

alter table transactions_2025_07
    owner to gnduser;

create table transactions_2025_08
    partition of transactions
        (
            constraint transactions_block_id_fkey
                foreign key (block_id) references blocks,
            constraint transactions_contract_id_fkey
                foreign key (contract_id) references contracts
            )
        FOR VALUES FROM ('2025-08-01 00:00:00') TO ('2025-09-01 00:00:00');

alter table transactions_2025_08
    owner to gnduser;

create table transactions_2025_09
    partition of transactions
        (
            constraint transactions_block_id_fkey
                foreign key (block_id) references blocks,
            constraint transactions_contract_id_fkey
                foreign key (contract_id) references contracts
            )
        FOR VALUES FROM ('2025-09-01 00:00:00') TO ('2025-10-01 00:00:00');

alter table transactions_2025_09
    owner to gnduser;

create table transactions_2025_10
    partition of transactions
        (
            constraint transactions_block_id_fkey
                foreign key (block_id) references blocks,
            constraint transactions_contract_id_fkey
                foreign key (contract_id) references contracts
            )
        FOR VALUES FROM ('2025-10-01 00:00:00') TO ('2025-11-01 00:00:00');

alter table transactions_2025_10
    owner to gnduser;

create table transactions_2025_11
    partition of transactions
        (
            constraint transactions_block_id_fkey
                foreign key (block_id) references blocks,
            constraint transactions_contract_id_fkey
                foreign key (contract_id) references contracts
            )
        FOR VALUES FROM ('2025-11-01 00:00:00') TO ('2025-12-01 00:00:00');

alter table transactions_2025_11
    owner to gnduser;

create table transactions_2025_12
    partition of transactions
        (
            constraint transactions_block_id_fkey
                foreign key (block_id) references blocks,
            constraint transactions_contract_id_fkey
                foreign key (contract_id) references contracts
            )
        FOR VALUES FROM ('2025-12-01 00:00:00') TO ('2026-01-01 00:00:00');

alter table transactions_2025_12
    owner to gnduser;

create table transactions_2026_01
    partition of transactions
        (
            constraint transactions_block_id_fkey
                foreign key (block_id) references blocks,
            constraint transactions_contract_id_fkey
                foreign key (contract_id) references contracts
            )
        FOR VALUES FROM ('2026-01-01 00:00:00') TO ('2026-02-01 00:00:00');

alter table transactions_2026_01
    owner to gnduser;

create table transactions_2026_02
    partition of transactions
        (
            constraint transactions_block_id_fkey
                foreign key (block_id) references blocks,
            constraint transactions_contract_id_fkey
                foreign key (contract_id) references contracts
            )
        FOR VALUES FROM ('2026-02-01 00:00:00') TO ('2026-03-01 00:00:00');

alter table transactions_2026_02
    owner to gnduser;

create table transactions_2026_03
    partition of transactions
        (
            constraint transactions_block_id_fkey
                foreign key (block_id) references blocks,
            constraint transactions_contract_id_fkey
                foreign key (contract_id) references contracts
            )
        FOR VALUES FROM ('2026-03-01 00:00:00') TO ('2026-04-01 00:00:00');

alter table transactions_2026_03
    owner to gnduser;

create table transactions_2026_04
    partition of transactions
        (
            constraint transactions_block_id_fkey
                foreign key (block_id) references blocks,
            constraint transactions_contract_id_fkey
                foreign key (contract_id) references contracts
            )
        FOR VALUES FROM ('2026-04-01 00:00:00') TO ('2026-05-01 00:00:00');

alter table transactions_2026_04
    owner to gnduser;

create table transactions_2026_05
    partition of transactions
        (
            constraint transactions_block_id_fkey
                foreign key (block_id) references blocks,
            constraint transactions_contract_id_fkey
                foreign key (contract_id) references contracts
            )
        FOR VALUES FROM ('2026-05-01 00:00:00') TO ('2026-06-01 00:00:00');

alter table transactions_2026_05
    owner to gnduser;

create table logs
(
    id           integer default nextval('logs_id_seq'::regclass) not null,
    tx_id        integer                                          not null,
    tx_timestamp timestamp                                        not null,
    contract_id  integer
        references contracts,
    event        varchar,
    data         jsonb,
    timestamp    timestamp                                        not null,
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

create index idx_logs_tx_id
    on logs (tx_id);

create index idx_logs_contract_id
    on logs (contract_id);

create index idx_logs_timestamp
    on logs (timestamp);

create table logs_2025_06
    partition of logs
        (
            constraint logs_contract_id_fkey
                foreign key (contract_id) references contracts,
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-06-01 00:00:00') TO ('2025-07-01 00:00:00');

alter table logs_2025_06
    owner to gnduser;

create table logs_2025_07
    partition of logs
        (
            constraint logs_contract_id_fkey
                foreign key (contract_id) references contracts,
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-07-01 00:00:00') TO ('2025-08-01 00:00:00');

alter table logs_2025_07
    owner to gnduser;

create table logs_2025_08
    partition of logs
        (
            constraint logs_contract_id_fkey
                foreign key (contract_id) references contracts,
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-08-01 00:00:00') TO ('2025-09-01 00:00:00');

alter table logs_2025_08
    owner to gnduser;

create table logs_2025_09
    partition of logs
        (
            constraint logs_contract_id_fkey
                foreign key (contract_id) references contracts,
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-09-01 00:00:00') TO ('2025-10-01 00:00:00');

alter table logs_2025_09
    owner to gnduser;

create table logs_2025_10
    partition of logs
        (
            constraint logs_contract_id_fkey
                foreign key (contract_id) references contracts,
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-10-01 00:00:00') TO ('2025-11-01 00:00:00');

alter table logs_2025_10
    owner to gnduser;

create table logs_2025_11
    partition of logs
        (
            constraint logs_contract_id_fkey
                foreign key (contract_id) references contracts,
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-11-01 00:00:00') TO ('2025-12-01 00:00:00');

alter table logs_2025_11
    owner to gnduser;

create table logs_2025_12
    partition of logs
        (
            constraint logs_contract_id_fkey
                foreign key (contract_id) references contracts,
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2025-12-01 00:00:00') TO ('2026-01-01 00:00:00');

alter table logs_2025_12
    owner to gnduser;

create table logs_2026_01
    partition of logs
        (
            constraint logs_contract_id_fkey
                foreign key (contract_id) references contracts,
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2026-01-01 00:00:00') TO ('2026-02-01 00:00:00');

alter table logs_2026_01
    owner to gnduser;

create table logs_2026_02
    partition of logs
        (
            constraint logs_contract_id_fkey
                foreign key (contract_id) references contracts,
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2026-02-01 00:00:00') TO ('2026-03-01 00:00:00');

alter table logs_2026_02
    owner to gnduser;

create table logs_2026_03
    partition of logs
        (
            constraint logs_contract_id_fkey
                foreign key (contract_id) references contracts,
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2026-03-01 00:00:00') TO ('2026-04-01 00:00:00');

alter table logs_2026_03
    owner to gnduser;

create table logs_2026_04
    partition of logs
        (
            constraint logs_contract_id_fkey
                foreign key (contract_id) references contracts,
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2026-04-01 00:00:00') TO ('2026-05-01 00:00:00');

alter table logs_2026_04
    owner to gnduser;

create table logs_2026_05
    partition of logs
        (
            constraint logs_contract_id_fkey
                foreign key (contract_id) references contracts,
            constraint logs_tx_id_tx_timestamp_fkey
                foreign key (tx_id, tx_timestamp) references transactions
            )
        FOR VALUES FROM ('2026-05-01 00:00:00') TO ('2026-06-01 00:00:00');

alter table logs_2026_05
    owner to gnduser;

create table token_balances
(
    token_id integer not null
        references tokens,
    address  varchar not null
        references accounts (address),
    balance  numeric,
    symbol   varchar,
    primary key (token_id, address)
);

alter table token_balances
    owner to gnduser;

create index idx_token_balances_address
    on token_balances (address);

create index idx_token_balances_token_id
    on token_balances (token_id);

create unique index idx_token_balances_address_symbol
    on token_balances (address, symbol)
    where (symbol IS NOT NULL);

create table events
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
    updated_at   timestamp with time zone default CURRENT_TIMESTAMP not null,
    block_id     integer,
    tx_id        integer,
    address      varchar(42),
    topics       jsonb,
    data         jsonb,
    index        integer,
    removed      boolean                  default false,
    status       varchar,
    processed_at timestamp with time zone
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

create index idx_events_contract
    on events (contract);

create index idx_events_type
    on events (type);

create index idx_events_timestamp
    on events (timestamp);

create index idx_events_tx_hash
    on events (tx_hash);

create index idx_events_from_address
    on events (from_address);

create index idx_events_to_address
    on events (to_address);

create index idx_events_contract_type
    on events (contract, type);

create index idx_events_contract_timestamp
    on events (contract, timestamp);

create table states
(
    id           serial
        primary key,
    block_id     integer not null,
    address      varchar not null,
    balance      numeric,
    nonce        bigint default 0,
    storage_root varchar,
    code_hash    varchar,
    created_at   timestamp,
    updated_at   timestamp,
    metadata     bytea
);

comment on table states is 'Состояние по адресам и блокам (core.BlockchainState)';

alter table states
    owner to gnduser;

create index idx_states_block_id
    on states (block_id);

create index idx_states_address
    on states (address);

create unique index idx_states_address_block
    on states (address, block_id);

create table signer_wallets
(
    id             uuid                     default gen_random_uuid() not null
        primary key,
    account_id     integer                                            not null
        unique
        references accounts
            on delete cascade,
    public_key     bytea                                              not null,
    encrypted_priv bytea                                              not null,
    disabled       boolean                  default false             not null,
    created_at     timestamp with time zone default now()             not null,
    updated_at     timestamp with time zone default now()             not null
);

alter table signer_wallets
    owner to gnduser;

create table wallets
(
    id               serial
        primary key,
    account_id       integer                 not null
        references accounts
            on delete cascade,
    address          varchar                 not null
        unique,
    public_key       varchar                 not null
        unique,
    private_key      varchar,
    created_at       timestamp default now(),
    updated_at       timestamp default now(),
    status           varchar   default 'active'::character varying,
    signer_wallet_id uuid
                                             references signer_wallets
                                                 on delete set null,
    name             varchar(255),
    role             varchar(64),
    disabled         boolean   default false not null
);

comment on column wallets.name is 'Человекочитаемое имя (например: Validator, Treasury)';

comment on column wallets.role is 'Системная роль: validator, treasury, fee_collector или NULL';

comment on column wallets.disabled is 'Мягкое удаление: кошелёк скрыт из списка админки';

alter table wallets
    owner to gnduser;

create index idx_wallets_signer_wallet_id
    on wallets (signer_wallet_id)
    where (signer_wallet_id IS NOT NULL);

create index idx_wallets_role
    on wallets (role)
    where (role IS NOT NULL);

create index idx_wallets_disabled
    on wallets (disabled)
    where (disabled = false);

create index idx_signer_wallets_account_id
    on signer_wallets (account_id);

create function update_updated_at_column() returns trigger
    language plpgsql
as
$$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;

alter function update_updated_at_column() owner to gnduser;

-- Триггер автообновления updated_at для events (соответствует 002_schema_additions / 001)
create trigger update_events_updated_at
    before update on events
    for each row
    execute procedure update_updated_at_column();

