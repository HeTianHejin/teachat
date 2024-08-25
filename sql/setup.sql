DROP database teachatwebdb;
CREATE database teachatwebdb;

drop table users;
drop table user_stars;
drop table user_default_teams;
drop table follows;
drop table friends;
drop table fans;
drop table sessions;
drop table objectives;
drop table objective_invited_teams;
drop table projects;
drop table project_invited_teams;
drop table draft_threads;
drop table threads;
drop table reads;
drop table draft_posts;
drop table posts;
drop table accept_messages;
drop table accept_objects;
drop table new_message_counts;
drop table acceptance;
drop table teams;
drop table team_members;
drop table team_roles;
drop table invitations;
drop table invitation_replies;
drop table families;
drop table communities;
drop table administrators;
drop table watchwords;
drop table monologues;
drop table goods;
drop table groups;
drop table handicrafts;
drop table inaugurations;
drop table parts;
drop table tool_lists;
drop table evidences;

CREATE TABLE evidences (
    id                 SERIAL PRIMARY KEY,
    uuid               VARCHAR(64) NOT NULL UNIQUE,
    handicraft_id      INTEGER NOT NULL,
    recorder           INTEGER NOT NULL,
    description        TEXT,
    images             VARCHAR(255),
    video              VARCHAR(255), 
    audio              VARCHAR(255), 
    created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tool_lists (
    id               BIGSERIAL PRIMARY KEY,
    uuid             VARCHAR(64) NOT NULL UNIQUE,
    part_id          BIGINT NOT NULL REFERENCES parts(id),
    goods_id         BIGINT NOT NULL REFERENCES goods(id),
    remark           TEXT,
    num              INTEGER NOT NULL, 
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE parts (
    id               SERIAL PRIMARY KEY,
    uuid             VARCHAR(64) NOT NULL UNIQUE,
    handicraft_id    INTEGER NOT NULL,
    name             VARCHAR(255) NOT NULL,
    nickname         VARCHAR(255),
    artist           INTEGER NOT NULL REFERENCES users(id),
    target_goods_id  INTEGER NOT NULL,
    tool_list_id     INTEGER NOT NULL,
    strength         INTEGER NOT NULL,
    intelligence     INTEGER NOT NULL,
    difficulty_level  INTEGER NOT NULL,
    recorder         INTEGER NOT NULL REFERENCES users(id),
    description      TEXT,
    evidence_id      INTEGER DEFAULT 0,
    status           INTEGER NOT NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inaugurations (
    id                  SERIAL PRIMARY KEY,
    uuid                VARCHAR(64) NOT NULL UNIQUE,
    handicraft_id       INTEGER NOT NULL,
    name                VARCHAR(255) NOT NULL,
    nickname            VARCHAR(255),
    artist              INTEGER NOT NULL REFERENCES users(id),
    recorder            INTEGER NOT NULL REFERENCES users(id),
    description         TEXT,
    evidence_id         INTEGER DEFAULT 0,
    status              INTEGER NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE handicrafts (
    id                        SERIAL PRIMARY KEY,
    uuid                      VARCHAR(64) NOT NULL,
    project_id                INTEGER REFERENCES projects(id),
    name                      VARCHAR(255) NOT NULL,
    nickname                  VARCHAR(255),
    client                    INTEGER NOT NULL REFERENCES users(id),
    target_goods_id           INTEGER,
    tool_list_id              INTEGER,        
    artist                    INTEGER NOT NULL REFERENCES users(id), 
    strength                  INTEGER,
    intelligence              INTEGER,
    difficulty_level           INTEGER,
    recorder                  INTEGER NOT NULL REFERENCES users(id),
    description               TEXT,
    evidence_id               INTEGER DEFAULT 0,
    status                    INTEGER,
    created_at                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE groups (
    id                      SERIAL PRIMARY KEY,
    uuid                    VARCHAR(64) NOT NULL,
    name                    VARCHAR(255) NOT NULL,
    mission                 TEXT,
    founder_id              INTEGER NOT NULL REFERENCES users(id),
    first_team_id            INTEGER REFERENCES teams(id),
    created_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    class                   VARCHAR(255),
    abbreviation            VARCHAR(255),
    logo                    VARCHAR(255),
    updated_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ministry_id             INTEGER default 0
);

CREATE TABLE goods (
    id                           SERIAL PRIMARY KEY,
    uuid                         VARCHAR(64) NOT NULL,
    user_id                      INTEGER NOT NULL REFERENCES users(id), 
    name                         VARCHAR(255) NOT NULL,
    nickname                     VARCHAR(255),
    designer                     VARCHAR(255),
    describe                     TEXT,
    price                        NUMERIC(10,2),
    applicability                VARCHAR(255),
    category                     VARCHAR(255),
    specification                 VARCHAR(255),
    brand_name                   VARCHAR(255),
    model                        VARCHAR(255),
    weight                       VARCHAR(255),
    dimensions                   VARCHAR(255),
    material                     VARCHAR(255),
    size                         VARCHAR(255),
    color                        VARCHAR(255),
    network_connection_type      VARCHAR(255),
    features                     TEXT,
    serial_number                VARCHAR(255),
    production_date              DATE,
    expiration_date              DATE,
    state                        VARCHAR(255),
    origin                       VARCHAR(255),
    manufacturer                 VARCHAR(255),
    manufacturer_link            VARCHAR(255),
    engine_type                  VARCHAR(255),
    purchase_link                VARCHAR(255),
    created_time                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time                 TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table users (
  id                 serial primary key,
  uuid               varchar(64) not null unique,
  name               varchar(255),
  email              varchar(255) not null unique,
  password           varchar(255) not null,
  created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  biography          text,
  role               varchar(64),
  gender             integer,
  avatar             varchar(255),
  updated_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE follows (
    id                   SERIAL PRIMARY KEY,
    uuid                 varchar(64) NOT NULL UNIQUE,
    user_id              INT NOT NULL REFERENCES users(id),
    followed_user_id     INT NOT NULL REFERENCES users(id),
    nickname             TEXT,
    note                 TEXT,
    relationship_level   INT, 
    is_disdain           boolean,
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE friends (
    id                   SERIAL PRIMARY KEY,
    uuid                 varchar(64) NOT NULL UNIQUE,
    user_id              INT NOT NULL REFERENCES users(id),
    friend_user_id       INT NOT NULL REFERENCES users(id),
    nickname             TEXT,
    note                 TEXT,
    relationship_level   INT, 
    is_rival             BOOLEAN,
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE fans (
    id                   SERIAL PRIMARY KEY,
    uuid                 varchar(64) NOT NULL UNIQUE,
    user_id              INT NOT NULL REFERENCES users(id),
    fan_user_id          INT NOT NULL REFERENCES users(id),
    nickname             TEXT,
    note                 TEXT,
    relationship_level   INT, 
    is_black_list        BOOLEAN,
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table user_stars (
  id             serial primary key,
  uuid           varchar(64) not null unique,
  user_id        integer references users(id),
  type           integer default 0,
  object_id      integer default 0,
  created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table user_default_teams (
  id             serial primary key,
  user_id        integer references users(id),
  team_id        integer references teams(id),
  created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table administrators (
  id         serial primary key
  user_id    integer references users(id),
);

create table sessions (
  id             serial primary key,
  uuid           varchar(64) not null unique,
  email          varchar(255),
  user_id        integer references users(id),
  created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  gender         integer
);

create table watchwords (
  id                   serial primary key,
  word                 varchar(255) not null,
  administrator_id     integer references administrators(id),
  created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP   
);

create table objectives (
  id              serial primary key,
  uuid            varchar(64) not null unique,
  title           varchar(64) not null,
  body            text,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  user_id         integer references users(id),
  class           integer,
  edit_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  star_count      integer default 0,
  cover           varchar(64)
);

create table projects (
  id              serial primary key,
  uuid            varchar(64) not null unique,
  title           varchar(64) not null,
  body            text,
  objective_id    integer references objectives(id),
  user_id         integer references users(id),
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  class           integer,
  edit_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  cover           varchar(64)
);

create table draft_threads (
  id            serial primary key,
  user_id       integer references users(id),
  project_id    integer references projects(id),
  title         varchar(64),
  body          text,
  class         integer default 0,
  created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  type          integer default 0,
  post_id       integer default 0
);

create table threads (
  id            serial primary key,
  uuid          varchar(64) not null unique,
  body          text,
  user_id       integer references users(id),
  created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  class         integer default 10,
  title         varchar(64),
  edit_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  project_id    integer references projects(id),
  hit_count     integer default 0,
  type          integer default 0,
  post_id       integer default 0
);

create table reads (
  id            serial primary key,
  user_id       integer,
  thread_id     integer references threads(id),
  read_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table draft_posts (
   id            serial primary key,
  body          text,
  user_id       integer references users(id),
  thread_id     integer references threads(id),
  created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  attitude      boolean,
  class         integer default 0
);

create table posts (
  id            serial primary key,
  uuid          varchar(64) not null unique,
  body          text,
  user_id       integer references users(id),
  thread_id     integer references threads(id),
  created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  edit_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  attitude      boolean,
  score         integer default 60
);

create table administrators (
  id              serial primary key,
  uuid            varchar(64) not null unique,
  user_id         integer references users(id),
  role            varchar(64) not null,
  password        varchar(255) not null,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  valid           boolean default false,
  invalidReason   text,
  invalid_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table teams (
  id              serial primary key,
  uuid            varchar(64) not null unique,
  name            varchar(255),
  mission         text,
  founder_id      integer references users(id),
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  class           integer,
  abbreviation    integer,
  logo            varchar(255),
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  group_id        integer default 0
);

create table team_members (
  id              serial primary key,
  uuid            varchar(64) not null unique,
  team_id         integer references teams(id),
  user_id         integer references users(id),
  role            varchar(255), 
  created_at      timestamp DEFAULT CURRENT_TIMESTAMP,
  class           integer default 1,
  updated_at      timestamp DEFAULT CURRENT_TIMESTAMP
);

create table project_invited_teams (
  id              serial primary key,
  project_id      integer references projects(id),
  team_id         integer references teams(id),
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table objective_invited_teams (
  id              serial primary key,
  objective_id    integer references objectives(id),
  team_id         integer references teams(id),
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table team_roles (
  id                     serial primary key,
  uuid                   varchar(64) not null unique,
  team_id                integer references teams(id),
  team_ceo_user_id       integer references users(id),
  target_team_member_id  integer references team_members(id),
  role                   varchar(64),
  word                   text,
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  check_team_member_id   integer references team_members(id),
  check_at               TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
); 

create table invitations (
  id                   serial primary key,
  uuid                 varchar(64) not null unique,
  team_id              integer references teams(id),
  invite_email         varchar(255),
  role                 varchar(50),
  invite_word          text,
  created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status               integer
);

CREATE TABLE invitation_replies (
  id                 SERIAL PRIMARY KEY,
  uuid               VARCHAR(50) NOT NULL,
  invitation_id      INT references invitations(id), 
  user_id            integer references users(id),
  reply_word         text NOT NULL,
  created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE accept_messages (
  id                 SERIAL PRIMARY KEY,
  from_user_id       integer references users(id),
  to_user_id         integer references users(id),
  title              varchar(64),
  content            text,
  accept_object_id   integer references accept_objects(id),
  class              integer default 0,
  created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE acceptances (
  id                  SERIAL PRIMARY KEY,
  accept_object_id    INTEGER references accept_objects(id),
  x_accept            BOOLEAN default false,
  x_user_id           INTEGER references users(id),
  x_accepted_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
  y_accept            BOOLEAN default false,
  y_user_id           INTEGER references users(id),
  y_accepted_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table accept_objects (
  id              SERIAL PRIMARY KEY,
  object_type     INTEGER default 0,
  object_id       INTEGER
);

CREATE TABLE new_message_counts (
  id              SERIAL PRIMARY KEY,
  user_id         INTEGER references users(id),
  count           INTEGER default 0
);

CREATE TABLE families (
    id                                    SERIAL PRIMARY KEY,
    uuid                                  VARCHAR(255),
    name                                  VARCHAR(255),
    introduction                          TEXT,
    husband_user_id                       INTEGER references users(id),
    wife_user_id                          INTEGER references users(id),
    child_user_id_set                     INTEGER[],
    husband_from_family_id_set            INTEGER[],
    wife_from_family_id_set               INTEGER[],
    married                               BOOLEAN default false,
    adopted_child_user_id_set             INTEGER[],
    class                                 INTEGER default 0,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    logo                                  VARCHAR(255)
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

CREATE TABLE monologues {
    id                                    SERIAL PRIMARY KEY,
    uuid                                  VARCHAR(255),
    title                                 VARCHAR(255),
    content                               TEXT,
    user_id                               INTEGER references users(id),
    note                                  VARCHAR(255)
    category                              INTEGER default 0,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
};




    
    