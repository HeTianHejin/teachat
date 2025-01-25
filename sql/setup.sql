DROP database teachat;
CREATE database teachat;

drop table if exists users;
drop table if exists user_stars;
drop table if exists user_default_teams;
drop table if exists follows;
drop table if exists friends;
drop table if exists fans;
drop table if exists sessions;
drop table if exists objectives;
drop table if exists objective_invited_teams;
drop table if exists projects;
drop table if exists project_invited_teams;
drop table if exists draft_threads;
drop table if exists threads;
drop table if exists reads;
drop table if exists draft_posts;
drop table if exists posts;
drop table if exists accept_messages;
drop table if exists accept_objects;
drop table if exists new_message_counts;
drop table if exists acceptance;
drop table if exists teams;
drop table if exists team_members;
drop table if exists team_member_role_notices;
drop table if exists invitations;
drop table if exists invitation_replies;
drop table if exists families;
drop table if exists family_members;
drop table if exists family_member_role_notices;
drop table if exists family_member_sign_ins;
drop table if exists communities;
drop table if exists administrators;
drop table if exists watchwords;
drop table if exists monologues;
drop table if exists goods;
drop table if exists user_goods;
drop table if exists handicrafts;
drop table if exists inaugurations;
drop table if exists parts;
drop table if exists tools;
drop table if exists evidences;
drop table if exists last_queries;
drop table if exists place_addresses;
drop table if exists addresses;
drop table if exists places;
drop table if exists location_history;
drop table if exists user_place;
drop table if exists user_default_place;
drop table if exists user_address;
drop table if exists user_default_address;
drop table if exists project_place;
drop table if exists thread_approved;
drop table if exists thread_costs;
drop table if exists thread_time_slots;
drop table if exists thread_goods;
drop table if exists member_applications;
drop table if exists member_application_replies;
drop table if exists team_member_resignations;
drop table if exists footprints;


CREATE TABLE team_member_resignations (
    id                   SERIAL PRIMARY KEY,
    uuid                 VARCHAR(36),
    team_id              INTEGER,
    ceo_user_id          INTEGER,
    core_member_user_id  INTEGER,
    member_id            INTEGER,
    member_user_id       INTEGER,
    member_current_role  VARCHAR(36),
    title                VARCHAR(255),
    content              TEXT,
    status               SMALLINT,
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE footprints (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER,
    team_id               INTEGER,
    team_name             VARCHAR(255),
    team_type             SMALLINT,
    content               TEXT,
    content_id            INTEGER,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE member_applications (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(36),
    team_id               INTEGER,
    user_id               INTEGER,
    content               TEXT,
    status                SMALLINT NOT NULL DEFAULT 0,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE member_application_replies (
    id                                SERIAL PRIMARY KEY,
    uuid                              VARCHAR(36),
    member_application_id             INTEGER,
    team_id                           INTEGER,
    user_id                           INTEGER,
    reply_content                     VARCHAR(255),
    status                            SMALLINT NOT NULL DEFAULT 0,
    created_at                        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at                        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);




CREATE TABLE thread_goods (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER NOT NULL,
    thread_id             INTEGER NOT NULL,
    goods_id              INTEGER NOT NULL,
    project_id            INTEGER NOT NULL,
    type                  SMALLINT NOT NULL,
    number                INTEGER NOT NULL,
    created_time          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE thread_time_slots (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER NOT NULL,
    thread_id             INTEGER NOT NULL,
    time_slot             INTEGER NOT NULL,
    is_confirm             SMALLINT NOT NULL,
    created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    project_id            INTEGER NOT NULL
);
COMMENT ON COLUMN thread_time_slots.time_slot IS 'Duration in minutes';

CREATE TABLE thread_costs (
    id                    SERIAL PRIMARY KEY,
    user_id               INTEGER NOT NULL,
    thread_id             INTEGER NOT NULL,
    cost                  INTEGER NOT NULL,
    type                  SMALLINT NOT NULL,
    created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    project_id            INTEGER NOT NULL
);


CREATE TABLE thread_approved (
    id                    SERIAL PRIMARY KEY,
    project_id            INTEGER NOT NULL REFERENCES projects(id),
    thread_id             INTEGER NOT NULL REFERENCES threads(id),
    user_id               INTEGER NOT NULL REFERENCES users(id), 
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE project_place (
    id                   SERIAL PRIMARY KEY,
    project_id           INTEGER,
    place_id             INTEGER,
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE location_history (
    id                   SERIAL PRIMARY KEY,
    uuid                 VARCHAR(36),
    user_id              INTEGER,
    place_id             INTEGER,
    time                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    longitude            DOUBLE PRECISION NOT NULL,
    latitude             DOUBLE PRECISION NOT NULL,
    altitude             DOUBLE PRECISION,
    direction            DOUBLE PRECISION,
    speed                DOUBLE PRECISION,
    accuracy             DOUBLE PRECISION,
    adcode               INTEGER,
    provider             VARCHAR(255),
    addr                 VARCHAR(255)  
);


CREATE TABLE places (
    id                   SERIAL PRIMARY KEY,
    uuid                 VARCHAR(36) NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    nickname             VARCHAR(255),
    description          TEXT,
    icon                 VARCHAR(255) DEFAULT 'bootstrap-icons/bank.svg',
    occupant_user_id     INTEGER,
    owner_user_id        INTEGER,
    level                INTEGER DEFAULT 0,
    category             INTEGER DEFAULT 0,
    is_public            BOOLEAN DEFAULT true,
    is_government        BOOLEAN DEFAULT false,
    user_id              INTEGER,
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE addresses (
    id                   SERIAL PRIMARY KEY,
    uuid                 VARCHAR(36) NOT NULL,
    nation               VARCHAR(255),
    province             VARCHAR(255),
    city                 VARCHAR(255),
    district             VARCHAR(255),
    town                 VARCHAR(255),
    village              VARCHAR(255),
    street               VARCHAR(255),
    building             VARCHAR(255),
    unit                 VARCHAR(255),
    portal_number        VARCHAR(255),
    postal_code          VARCHAR(20) DEFAULT 0,
    category             INTEGER DEFAULT 0,
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE user_place (
    id                   SERIAL PRIMARY KEY,
    user_id              INTEGER REFERENCES users(id),
    place_id             INTEGER REFERENCES places(id),
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE user_default_place (
    id                   SERIAL PRIMARY KEY,
    user_id              INTEGER REFERENCES users(id),
    place_id             INTEGER REFERENCES places(id),
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE user_address (
    id                   SERIAL PRIMARY KEY,
    user_id              INTEGER REFERENCES users(id),
    address_id           INTEGER REFERENCES addresses(id),
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE user_default_address (
    id                   SERIAL PRIMARY KEY,
    user_id              INTEGER REFERENCES users(id),
    address_id           INTEGER REFERENCES addresses(id),
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);






CREATE TABLE place_addresses (
    place_id             INTEGER REFERENCES places(id),
    address_id           INTEGER REFERENCES addresses(id),
                         PRIMARY KEY (place_id, address_id)
);


CREATE TABLE last_queries (
    id                         SERIAL PRIMARY KEY,
    user_id                    INTEGER REFERENCES users(id),
    path                       VARCHAR(255),
    query                      VARCHAR(255),
    query_at                   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE evidences (
    id                         SERIAL PRIMARY KEY,
    uuid                       VARCHAR(64) NOT NULL UNIQUE,
    handicraft_id              INTEGER NOT NULL,
    recorder                   INTEGER NOT NULL,
    description                TEXT,
    images                     VARCHAR(255),
    video                      VARCHAR(255), 
    audio                      VARCHAR(255), 
    created_at                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tools (
    id                         BIGSERIAL PRIMARY KEY,
    uuid                       VARCHAR(64) NOT NULL UNIQUE,
    handicraft_id              INTEGER NOT NULL,
    part_id                    BIGINT,
    goods_id                   BIGINT,
    note                       TEXT,
    category                   INTEGER, 
    created_at                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE parts (
    id                         SERIAL PRIMARY KEY,
    uuid                       VARCHAR(64) NOT NULL UNIQUE,
    handicraft_id              INTEGER NOT NULL,
    name                       VARCHAR(255) NOT NULL,
    nickname                   VARCHAR(255),
    artist                     INTEGER NOT NULL REFERENCES users(id),
    target_goods_id            INTEGER,
    tool_list_id               INTEGER,
    magic_list_id              INTEGER,
    strength                   INTEGER,
    intelligence               INTEGER,
    difficulty_level            INTEGER,
    recorder                   INTEGER NOT NULL REFERENCES users(id),
    description                TEXT,
    evidence_id                INTEGER DEFAULT 0,
    status                     INTEGER NOT NULL,
    created_at                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inaugurations (
    id                         SERIAL PRIMARY KEY,
    uuid                       VARCHAR(64) NOT NULL UNIQUE,
    handicraft_id              INTEGER NOT NULL,
    name                       VARCHAR(255) NOT NULL,
    nickname                   VARCHAR(255),
    artist                     INTEGER NOT NULL REFERENCES users(id),
    recorder                   INTEGER NOT NULL REFERENCES users(id),
    description                TEXT,
    evidence_id                INTEGER DEFAULT 0,
    status                     INTEGER NOT NULL,
    created_at                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE handicrafts (
    id                        SERIAL PRIMARY KEY,
    uuid                      VARCHAR(64) NOT NULL,
    project_id                INTEGER NOT NULL,
    name                      VARCHAR(255) NOT NULL,
    nickname                  VARCHAR(255),
    client_team_id            INTEGER DEFAULT 0,
    goods_list_id             INTEGER DEFAULT 0,
    target_goods_id           INTEGER DEFAULT 0,
    tool_list_id              INTEGER DEFAULT 0,
    artist                    INTEGER NOT NULL, 
    strength                  INTEGER,
    intelligence              INTEGER,
    difficulty_level           INTEGER,
    recorder                  INTEGER NOT NULL,
    description               TEXT,
    category                  INTEGER DEFAULT 0,
    status                    INTEGER,
    created_at                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);



CREATE TABLE goods (
    id                           SERIAL PRIMARY KEY,
    uuid                         VARCHAR(64) NOT NULL,
    user_id                      INTEGER NOT NULL, 
    name                         VARCHAR(255) NOT NULL,
    nickname                     VARCHAR(255),
    designer                     VARCHAR(255),
    describe                     TEXT,
    price                        NUMERIC(10,2),
    applicability                VARCHAR(255),
    category                     integer,
    specification                 VARCHAR(255),
    brand_name                   VARCHAR(255),
    model                        VARCHAR(255),
    weight                       VARCHAR(255),
    dimensions                   VARCHAR(255),
    material                     VARCHAR(255),
    size                         VARCHAR(255),
    color                        VARCHAR(255),
    network_connection_type      VARCHAR(255),
    features                     integer,
    serial_number                VARCHAR(255),
    production_date              DATE DEFAULT CURRENT_DATE,
    expiration_date              DATE DEFAULT CURRENT_DATE + interval '365 day',
    state                        VARCHAR(255),
    origin                       VARCHAR(255),
    manufacturer                 VARCHAR(255),
    manufacturer_link            VARCHAR(255),
    engine_type                  VARCHAR(255),
    purchase_link                VARCHAR(255),
    created_time                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
create table user_goods (
    id                           SERIAL PRIMARY KEY,
    user_id                      INTEGER NOT NULL,
    goods_id                     INTEGER NOT NULL,
    created_at                   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


create table users (
  id                             serial primary key,
  uuid                           varchar(64) not null unique,
  name                           varchar(255),
  email                          varchar(255) not null unique,
  password                       varchar(255) not null,
  created_at                     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  biography                      text,
  role                           varchar(64),
  gender                         integer,
  avatar                         varchar(255),
  updated_at                     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE follows (
    id                           SERIAL PRIMARY KEY,
    uuid                         varchar(64) NOT NULL UNIQUE,
    user_id                      INT NOT NULL REFERENCES users(id),
    followed_user_id             INT NOT NULL REFERENCES users(id),
    nickname                     TEXT,
    note                         TEXT,
    relationship_level           INT, 
    is_disdain                   boolean,
    created_at                   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE friends (
    id                           SERIAL PRIMARY KEY,
    uuid                         varchar(64) NOT NULL UNIQUE,
    user_id                      INT NOT NULL REFERENCES users(id),
    friend_user_id               INT NOT NULL REFERENCES users(id),
    nickname                     TEXT,
    note                         TEXT,
    relationship_level           INT, 
    is_rival                     BOOLEAN,
    created_at                   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE fans (
    id                           SERIAL PRIMARY KEY,
    uuid                         varchar(64) NOT NULL UNIQUE,
    user_id                      INT NOT NULL REFERENCES users(id),
    fan_user_id                  INT NOT NULL REFERENCES users(id),
    nickname                     TEXT,
    note                         TEXT,
    relationship_level           INT, 
    is_black_list                BOOLEAN,
    created_at                   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table user_stars (
  id                             serial primary key,
  uuid                           varchar(64) not null unique,
  user_id                        integer references users(id),
  type                           integer default 0,
  object_id                      integer default 0,
  created_at                     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table user_default_teams (
  id                             serial primary key,
  user_id                        integer references users(id),
  team_id                        integer references teams(id),
  created_at                     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at                     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table administrators (
  id                     serial primary key
  user_id                integer references users(id),
);

create table sessions (
  id                     serial primary key,
  uuid                   varchar(64) not null unique,
  email                  varchar(255),
  user_id                integer references users(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  gender                 integer
);

create table watchwords (
  id                     serial primary key,
  word                   varchar(255) not null,
  administrator_id       integer references administrators(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP   
);

create table objectives (
  id                     serial primary key,
  uuid                   varchar(64) not null unique,
  title                  varchar(64) not null,
  body                   text,
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  user_id                integer references users(id),
  class                  integer,
  edit_at                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  star_count             integer default 0,
  cover                  varchar(64),
  team_id                integer not null default 2
);

create table projects (
  id                     serial primary key,
  uuid                   varchar(64) not null unique,
  title                  varchar(64) not null,
  body                   text,
  objective_id           integer references objectives(id),
  user_id                integer references users(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  class                  integer,
  edit_at                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  cover                  varchar(64),
  team_id                integer not null default 2
);

create table draft_threads (
  id                     serial primary key,
  user_id                integer references users(id),
  project_id             integer references projects(id),
  title                  varchar(64),
  body                   text,
  class                  integer default 0,
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  type                   integer default 0,
  post_id                integer default 0,
  team_id                integer not null default 2
);

create table threads (
  id                     serial primary key,
  uuid                   varchar(64) not null unique,
  body                   text,
  user_id                integer references users(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  class                  integer default 10,
  title                  varchar(64),
  edit_at                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  project_id             integer references projects(id),
  hit_count              integer default 0,
  type                   integer default 0,
  post_id                integer default 0,
  team_id                integer not null default 2
);

create table reads (
  id                     serial primary key,
  user_id                integer,
  thread_id              integer references threads(id),
  read_at                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table draft_posts (
   id                    serial primary key,
  body                   text,
  user_id                integer references users(id),
  thread_id              integer references threads(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  attitude               boolean,
  class                  integer default 0,
  team_id                integer not null default 2
);

create table posts (
  id                     serial primary key,
  uuid                   varchar(64) not null unique,
  body                   text,
  user_id                integer references users(id),
  thread_id              integer references threads(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  edit_at                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  attitude               boolean,
  score                  integer default 60,
  team_id                integer not null default 2
);

create table administrators (
  id                     serial primary key,
  uuid                   varchar(64) not null unique,
  user_id                integer references users(id),
  role                   varchar(64) not null,
  password               varchar(255) not null,
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  valid                  boolean default false,
  invalidReason          text,
  invalid_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table teams (
  id                     serial primary key,
  uuid                   varchar(64) not null unique,
  name                   varchar(255),
  mission                text,
  founder_id             integer references users(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  class                  integer,
  abbreviation           integer,
  logo                   varchar(255),
  updated_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  superior_team_id       integer default 0,
  subordinate_team_id    integer default 0
);

create table team_members (
  id                     serial primary key,
  uuid                   varchar(64) not null unique,
  team_id                integer references teams(id),
  user_id                integer references users(id),
  role                   varchar(255), 
  created_at             timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  class                  integer default 1,
  updated_at             timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);


create table team_member_role_notices (
  id                     serial primary key,
  uuid                   varchar(64) not null unique,
  team_id                integer references teams(id),
  ceo_id                 integer references users(id),
  member_id              integer,
  member_current_role    varchar(64),
  new_role               varchar(64),
  title                  varchar(64),
  content                text,
  status                 integer,
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table project_invited_teams (
  id                     serial primary key,
  project_id             integer references projects(id),
  team_id                integer references teams(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table objective_invited_teams (
  id                     serial primary key,
  objective_id           integer references objectives(id),
  team_id                integer references teams(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);



create table invitations (
  id                   serial primary key,
  uuid                 varchar(64) not null unique,
  team_id              integer references teams(id),
  invite_email         varchar(255),
  role                 varchar(50),
  invite_word          text,
  created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status               integer,
  author_user_id       integer
);

CREATE TABLE invitation_replies (
  id                   SERIAL PRIMARY KEY,
  uuid                 VARCHAR(50) NOT NULL,
  invitation_id        INT references invitations(id), 
  user_id              integer references users(id),
  reply_word           text NOT NULL,
  created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE accept_messages (
  id                   SERIAL PRIMARY KEY,
  from_user_id         integer references users(id),
  to_user_id           integer references users(id),
  title                varchar(64),
  content              text,
  accept_object_id     integer references accept_objects(id),
  class                integer default 0,
  created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE acceptances (
  id                   SERIAL PRIMARY KEY,
  accept_object_id     INTEGER references accept_objects(id),
  x_accept             BOOLEAN default false,
  x_user_id            INTEGER references users(id),
  x_accepted_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
  y_accept             BOOLEAN default false,
  y_user_id            INTEGER references users(id),
  y_accepted_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table accept_objects (
  id                   SERIAL PRIMARY KEY,
  object_type          INTEGER default 0,
  object_id            INTEGER
);

CREATE TABLE new_message_counts (
  id                   SERIAL PRIMARY KEY,
  user_id              INTEGER,
  count                INTEGER default 0
);

CREATE TABLE families (
    id                                    SERIAL PRIMARY KEY,
    uuid                                  VARCHAR(255),
    author_id                             INTEGER,
    name                                  VARCHAR(255),
    introduction                          TEXT,
    is_married                            BOOLEAN default true,
    has_child                             BOOLEAN,
    husband_from_family_id                INTEGER default 0,
    wife_from_family_id                   INTEGER default 0,
    status                                 INTEGER default 1,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    logo                                  VARCHAR(255),
    is_open                               BOOLEAN default true
);

CREATE TABLE family_members (
    id                                    SERIAL PRIMARY KEY,
    uuid                                  VARCHAR(255),
    family_id                             INTEGER,
    user_id                               INTEGER,
    role                                  INTEGER default 0,
    is_adult                              BOOLEAN default true,
    nick_name                             VARCHAR(255) default ':P',
    is_adopted                            BOOLEAN default false,
    age                                   INTEGER default 0,
    order_of_seniority                    INTEGER default 0,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE family_member_sign_ins (
    id                                    SERIAL PRIMARY KEY,
    uuid                                  VARCHAR(255),
    family_id                             INTEGER,
    user_id                               INTEGER,
    role                                  INTEGER default 0,
    is_adult                              BOOLEAN default true,
    title                                 VARCHAR(255),
    content                               TEXT,
    place_id                              INTEGER default 0,
    status                                INTEGER default 0,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_adopted                            BOOLEAN default false
);

CREATE TABLE user_default_families (
    id                                    SERIAL PRIMARY KEY,
    user_id                               INTEGER,
    family_id                             INTEGER,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);



CREATE TABLE communities (
    id                                    SERIAL PRIMARY KEY,
    uuid                                  VARCHAR(255),
    name                                  VARCHAR(255),
    introduction                          TEXT,
    family_id_set                         INTEGER[],
    influence_team_id                      INTEGER,
    edited_user_id_set                    INTEGER[],
    class                                 INTEGER default 0,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    flag                                   VARCHAR(255)
);

CREATE TABLE monologues (
    id                                    SERIAL PRIMARY KEY,
    uuid                                  VARCHAR(255),
    title                                 VARCHAR(255),
    content                               TEXT,
    user_id                               INTEGER references users(id),
    note                                  VARCHAR(255)
    category                              INTEGER default 0,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);




    
    