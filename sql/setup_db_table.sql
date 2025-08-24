DROP database teachat;
CREATE database teachat;

drop table if exists users;
drop table if exists user_stars;
drop table if exists user_default_teams;
drop table if exists sessions;
drop table if exists objectives;
drop table if exists objective_invited_teams;
drop table if exists projects;
drop table if exists project_invited_teams;
drop table if exists draft_posts;
drop table if exists posts;
drop table if exists draft_threads;
drop table if exists threads;
drop table if exists reads;
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
drop table if exists family_member_sign_in_replies;
drop table if exists family_member_sign_outs;
drop table if exists administrators;
drop table if exists goods;
drop table if exists goods_teams;
drop table if exists goods_users;
drop table if exists goods_families;
drop table if exists handicrafts;
drop table if exists inaugurations;
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
drop table if exists member_applications;
drop table if exists member_application_replies;
drop table if exists team_member_resignations;
drop table if exists footprints;
drop table if exists project_appointments;
drop table if exists environments;
drop table if exists see_seek_risks;
drop table if exists see_seek_hazards;
drop table if exists see_seek_environments;
drop table if exists see_seeks;
drop table if exists safety_protections;
drop table if exists safety_measures;
drop table if exists hazards;
drop table if exists risks;






CREATE TABLE places (
    id                   SERIAL PRIMARY KEY,
    uuid                 VARCHAR(36) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
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
    updated_at           TIMESTAMP
);

CREATE TABLE addresses (
    id                   SERIAL PRIMARY KEY,
    uuid                 VARCHAR(36) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
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
    updated_at           TIMESTAMP
);


create table users (
  id                             serial primary key,
  uuid                           varchar(64) not null unique DEFAULT gen_random_uuid(),
  name                           varchar(255),
  email                          varchar(255) not null unique,
  password                       varchar(255) not null,
  created_at                     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  biography                      text,
  role                           varchar(64),
  gender                         integer,
  avatar                         varchar(255),
  updated_at                     TIMESTAMP
);

CREATE TABLE families (
    id                                    SERIAL PRIMARY KEY,
    uuid                                  VARCHAR(255) DEFAULT gen_random_uuid(),
    author_id                             INTEGER,
    name                                  VARCHAR(255),
    introduction                          TEXT,
    is_married                            BOOLEAN default true,
    has_child                             BOOLEAN,
    husband_from_family_id                INTEGER default 0,
    wife_from_family_id                   INTEGER default 0,
    status                                 INTEGER default 1,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                            TIMESTAMP,
    logo                                  VARCHAR(255),
    is_open                               BOOLEAN default true
);

create table teams (
  id                     serial primary key,
  uuid                   varchar(64) not null unique DEFAULT gen_random_uuid(),
  name                   varchar(255),
  mission                text,
  founder_id             integer references users(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  class                  integer,
  abbreviation           integer,
  logo                   varchar(255),
  updated_at             TIMESTAMP,
  superior_team_id       integer default 0,
  subordinate_team_id    integer default 0
);

CREATE TABLE user_place (
    id                   SERIAL PRIMARY KEY,
    user_id              INTEGER,
    place_id             INTEGER,
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_default_place (
    id                   SERIAL PRIMARY KEY,
    user_id              INTEGER,
    place_id             INTEGER,
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


CREATE TABLE family_members (
    id                                    SERIAL PRIMARY KEY,
    uuid                                  VARCHAR(255) DEFAULT gen_random_uuid(),
    family_id                             INTEGER,
    user_id                               INTEGER,
    role                                  INTEGER default 0,
    is_adult                              BOOLEAN default true,
    nick_name                             VARCHAR(255) default ':P',
    is_adopted                            BOOLEAN default false,
    age                                   INTEGER default 0,
    order_of_seniority                    INTEGER default 0,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                            TIMESTAMP
);

CREATE TABLE family_member_sign_ins (
    id                                    SERIAL PRIMARY KEY,
    uuid                                  VARCHAR(255) DEFAULT gen_random_uuid(),
    family_id                             INTEGER,
    user_id                               INTEGER,
    role                                  INTEGER default 0,
    is_adult                              BOOLEAN default true,
    title                                 VARCHAR(255),
    content                               TEXT,
    place_id                              INTEGER default 0,
    status                                INTEGER default 0,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                            TIMESTAMP,
    is_adopted                            BOOLEAN default false
);

CREATE TABLE family_member_sign_in_replies (
    id                                    SERIAL PRIMARY KEY,
    uuid                                  VARCHAR(255) DEFAULT gen_random_uuid(),
    sign_in_id                            INTEGER,
    user_id                               INTEGER,
    is_confirm                            BOOLEAN default false,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE family_member_sign_outs (
    id                                    SERIAL PRIMARY KEY,
    uuid                                  VARCHAR(255) DEFAULT gen_random_uuid(),
    family_id                             INTEGER,
    user_id                               INTEGER,
    role                                  INTEGER default 0,
    is_adult                              BOOLEAN default true,
    title                                 VARCHAR(255),
    content                               TEXT,
    place_id                              INTEGER default 0,
    status                                INTEGER default 0,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                            TIMESTAMP,
    is_adopted                            BOOLEAN default false,
    author_user_id                        INTEGER
);

CREATE TABLE user_default_families (
    id                                    SERIAL PRIMARY KEY,
    user_id                               INTEGER,
    family_id                             INTEGER,
    created_at                            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);



create table team_members (
  id                     serial primary key,
  uuid                   varchar(64) not null unique DEFAULT gen_random_uuid(),
  team_id                integer references teams(id),
  user_id                integer references users(id),
  role                   varchar(255), 
  created_at             timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  class                  integer default 1,
  updated_at             timestamp
);

create table user_default_teams (
  id                             serial primary key,
  user_id                        integer references users(id),
  team_id                        integer references teams(id),
  created_at                     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at                     TIMESTAMP
);

create table team_member_role_notices (
  id                     serial primary key,
  uuid                   varchar(64) not null unique DEFAULT gen_random_uuid(),
  team_id                integer references teams(id),
  ceo_id                 integer references users(id),
  member_id              integer,
  member_current_role    varchar(64),
  new_role               varchar(64),
  title                  varchar(64),
  content                text,
  status                 integer,
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at             TIMESTAMP
);

create table project_invited_teams (
  id                     serial primary key,
  project_id             integer references projects(id),
  team_id                integer references teams(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at             TIMESTAMP
);

create table objective_invited_teams (
  id                     serial primary key,
  objective_id           integer references objectives(id),
  team_id                integer references teams(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at             TIMESTAMP
);

CREATE TABLE last_queries (
    id                         SERIAL PRIMARY KEY,
    user_id                    INTEGER REFERENCES users(id),
    path                       VARCHAR(255),
    query                      VARCHAR(255),
    query_at                   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table sessions (
  id                     serial primary key,
  uuid                   varchar(64) not null unique DEFAULT gen_random_uuid(),
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
  uuid                   varchar(64) not null unique DEFAULT gen_random_uuid(),
  title                  varchar(64) not null,
  body                   text,
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  user_id                integer NOT NULL,
  class                  integer NOT NULL,
  edit_at                TIMESTAMP,
  family_id              integer default 0,
  cover                  varchar(64),
  team_id                integer not null default 2,
  is_private             boolean default false
);


create table projects (
  id                     serial primary key,
  uuid                   varchar(64) not null unique DEFAULT gen_random_uuid(),
  title                  varchar(64) not null,
  body                   text,
  objective_id           integer,
  user_id                integer,
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  class                  integer,
  edit_at                TIMESTAMP,
  cover                  varchar(64) default 'default-pr-cover',
  team_id                integer not null default 2,
  is_private             boolean default false,
  family_id              integer not null default 0
);

create table project_approved (
  id                     serial primary key,
  user_id                integer not null,
  project_id             integer not null,
  objective_id           integer not null,
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table draft_posts (
   id                    serial primary key,
  body                   text,
  user_id                integer,
  thread_id              integer,
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  attitude               boolean,
  class                  integer default 0,
  team_id                integer not null default 2,
  is_private             boolean default false,
  family_id              integer default 0
);

create table posts (
  id                     serial primary key,
  uuid                   varchar(64) not null unique DEFAULT gen_random_uuid(),
  body                   text,
  user_id                integer references user(id),
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  edit_at                TIMESTAMP,
  attitude               boolean,
  family_id              integer default 0,
  team_id                integer not null default 2,
  is_private             boolean default false,
  class                  integer default 1
);

create table draft_threads (
  id                     serial primary key,
  user_id                integer not null,
  project_id             integer not null,
  title                  varchar(64),
  body                   text,
  class                  integer default 0,
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  type                   integer default 0,
  post_id                integer default 0,
  team_id                integer not null default 2,
  is_private             boolean default false,
  family_id              integer default 0,
  category               integer default 0
);

create table threads (
  id                     serial primary key,
  uuid                   varchar(64) not null unique DEFAULT gen_random_uuid(),
  body                   text,
  user_id                integer references users(id),
  class                  integer default 10,
  title                  varchar(64),
  project_id             integer references projects(id),
  family_id              integer default 0,
  team_id                integer not null default 2,
  type                   integer default 0,
  post_id                INTEGER REFERENCES post(id),
  is_private             boolean default false,
  category               integer default 0,
  created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  edit_at                TIMESTAMP
);

ALTER TABLE post
ADD COLUMN thread_id INTEGER NOT NULL REFERENCES thread(id);

create table reads (
  id                     serial primary key,
  user_id                integer,
  thread_id              integer references threads(id),
  read_at                TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table invitations (
  id                   serial primary key,
  uuid                 varchar(64) not null unique DEFAULT gen_random_uuid(),
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
  uuid                 VARCHAR(50) NOT NULL DEFAULT gen_random_uuid(),
  invitation_id        INTEGER, 
  user_id              integer,
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
  updated_at           TIMESTAMP
);

CREATE TABLE acceptances (
  id                   SERIAL PRIMARY KEY,
  accept_object_id     INTEGER,
  x_accept             BOOLEAN default false,
  x_user_id            INTEGER,
  x_accepted_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
  y_accept             BOOLEAN default false,
  y_user_id            INTEGER default 0,
  y_accepted_at        TIMESTAMP
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

create table goods (
    id                    serial primary key,
    uuid                  varchar(64) not null unique DEFAULT gen_random_uuid(),
    recorder_user_id      integer,
    name                  varchar(255),
    nickname              varchar(255),
    designer              varchar(255),
    describe              text,
    price                 float,
    applicability         varchar(255),
    category              integer,
    specification         varchar(255),
    brand_name            varchar(255),
    model                 varchar(255),
    weight                float,
    dimensions            varchar(255),
    material              varchar(255),
    size                  varchar(255),
    color                 varchar(255),
    network_connection_type varchar(255),
    features              integer,
    serial_number         varchar(255),
    state                 varchar(255),
    origin                varchar(255),
    manufacturer          varchar(255),
    manufacturer_url      varchar(255),
    engine_type           varchar(255),
    purchase_url          varchar(255),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE goods_teams (
    id            serial primary key,
    team_id       integer,
    goods_id      integer,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE goods_users (
    id            serial primary key,
    user_id       integer,
    goods_id      integer,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE goods_families (
    id            serial primary key,
    user_id       integer,
    goods_id      integer,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_id              INTEGER
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

CREATE TABLE team_member_resignations (
    id                   SERIAL PRIMARY KEY,
    uuid                 VARCHAR(36) DEFAULT gen_random_uuid(),
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
    updated_at           TIMESTAMP
);

CREATE TABLE member_applications (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(36) DEFAULT gen_random_uuid(),
    team_id               INTEGER,
    user_id               INTEGER,
    content               TEXT,
    status                SMALLINT NOT NULL DEFAULT 0,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE member_application_replies (
    id                                SERIAL PRIMARY KEY,
    uuid                              VARCHAR(36) DEFAULT gen_random_uuid(),
    member_application_id             INTEGER,
    team_id                           INTEGER,
    user_id                           INTEGER,
    reply_content                     VARCHAR(255),
    status                            SMALLINT NOT NULL DEFAULT 0,
    created_at                        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at                        TIMESTAMP
);

CREATE TABLE project_appointments (
    id                    SERIAL PRIMARY KEY,
    project_id            INTEGER NOT NULL,
    note                  VARCHAR(255),
    start_time            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_time              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    place_id              INTEGER NOT NULL DEFAULT 0,
    payer_user_id         INTEGER,
    payer_team_id         INTEGER,
    payer_family_id       INTEGER,
    payee_user_id         INTEGER,
    payee_team_id         INTEGER,
    payee_family_id       INTEGER,
    verifier_user_id      INTEGER,
    verifier_family_id    INTEGER,
    verifier_team_id      INTEGER,
    status                SMALLINT NOT NULL DEFAULT 0,
    confirmed_at          TIMESTAMP,
    rejected_at           TIMESTAMP,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE environments (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    name                  VARCHAR(255),
    summary               TEXT,
    temperature           INTEGER,
    humidity              INTEGER,
    pm25                  INTEGER,
    noise                 INTEGER,
    light                 INTEGER,
    wind                  INTEGER,
    flow                  INTEGER,
    rain                  INTEGER,
    pressure              INTEGER,
    smoke                 INTEGER,
    dust                  INTEGER,
    odor                  INTEGER,
    visibility            INTEGER,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE hazards (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    user_id               INTEGER REFERENCES users(id),
    name                  VARCHAR(255) NOT NULL,
    nickname              VARCHAR(255),
    keywords              VARCHAR(255),
    description           TEXT,
    source                VARCHAR(255),
    severity              INTEGER NOT NULL DEFAULT 1,
    category              INTEGER NOT NULL DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE risks (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    user_id               INTEGER REFERENCES users(id),
    name                  VARCHAR(255) NOT NULL,
    nickname              VARCHAR(255),
    keywords              VARCHAR(255),
    description           TEXT,
    source                VARCHAR(255),
    severity              INTEGER NOT NULL DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE safety_measures (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    hazard_id             INTEGER REFERENCES hazards(id),
    user_id               INTEGER REFERENCES users(id),
    title                 VARCHAR(255) NOT NULL,
    description           TEXT,
    priority              INTEGER NOT NULL DEFAULT 3,
    status                INTEGER NOT NULL DEFAULT 1,
    planned_date          TIMESTAMP,
    completed_date        TIMESTAMP,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE safety_protections (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    risk_id               INTEGER REFERENCES risks(id),
    user_id               INTEGER REFERENCES users(id),
    title                 VARCHAR(255) NOT NULL,
    description           TEXT,
    type                  INTEGER NOT NULL DEFAULT 1,
    priority              INTEGER NOT NULL DEFAULT 3,
    status                INTEGER NOT NULL DEFAULT 1,
    equipment             VARCHAR(255),
    planned_date          TIMESTAMP,
    completed_date        TIMESTAMP,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE see_seeks (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    name                  VARCHAR(255) NOT NULL,
    nickname              VARCHAR(255),
    description           TEXT,
    place_id              INTEGER,
    project_id            INTEGER,
    start_time            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time              TIMESTAMP DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    payer_user_id         INTEGER,
    payer_team_id         INTEGER,
    payer_family_id       INTEGER,
    payee_user_id         INTEGER,
    payee_team_id         INTEGER,
    payee_family_id       INTEGER,
    verifier_user_id      INTEGER,
    verifier_team_id      INTEGER,
    verifier_family_id    INTEGER,
    category              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    step                  INTEGER DEFAULT 1,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE see_seek_environments (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    environment_id        INTEGER REFERENCES environments(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE see_seek_hazards (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    hazard_id             INTEGER REFERENCES hazards(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE see_seek_risks (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    risk_id               INTEGER REFERENCES risks(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);
CREATE INDEX idx_safety_measures_hazard_id ON safety_measures(hazard_id);
CREATE INDEX idx_safety_measures_status ON safety_measures(status);
CREATE INDEX idx_safety_protections_risk_id ON safety_protections(risk_id);
CREATE INDEX idx_safety_protections_status ON safety_protections(status);
CREATE INDEX idx_hazards_severity ON hazards(severity);
CREATE INDEX idx_hazards_category ON hazards(category);
CREATE INDEX idx_see_seeks_project_id ON see_seeks(project_id);
CREATE INDEX idx_see_seeks_status ON see_seeks(status);

CREATE TABLE see_seek_looks (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    classify              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    outline               TEXT,
    is_deform             BOOLEAN DEFAULT false,
    skin                  TEXT,
    is_graze              BOOLEAN DEFAULT false,
    color                 VARCHAR(255),
    is_change             BOOLEAN DEFAULT false,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE see_seek_listens (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    classify              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    sound                 TEXT,
    is_abnormal           BOOLEAN DEFAULT false,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE see_seek_smells (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    classify              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    odour                 TEXT,
    is_foul_odour         BOOLEAN DEFAULT false,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE TABLE see_seek_touches (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    classify              INTEGER DEFAULT 0,
    status                INTEGER DEFAULT 0,
    temperature           VARCHAR(255),
    is_fever              BOOLEAN DEFAULT false,
    stretch               VARCHAR(255),
    is_stiff              BOOLEAN DEFAULT false,
    shake                 VARCHAR(255),
    is_shake              BOOLEAN DEFAULT false,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE INDEX idx_see_seek_environments_see_seek_id ON see_seek_environments(see_seek_id);
CREATE INDEX idx_see_seek_hazards_see_seek_id ON see_seek_hazards(see_seek_id);
CREATE INDEX idx_see_seek_risks_see_seek_id ON see_seek_risks(see_seek_id);

CREATE TABLE see_seek_examination_reports (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    see_seek_id           INTEGER REFERENCES see_seeks(id),
    classify              INTEGER DEFAULT 1,
    status                INTEGER DEFAULT 0,
    name                  VARCHAR(255),
    nickname              VARCHAR(255),
    description           TEXT,
    sample_type           VARCHAR(255),
    sample_order          VARCHAR(255),
    instrument_goods_id   INTEGER,
    report_title          VARCHAR(255),
    report_content        TEXT,
    master_user_id        INTEGER,
    reviewer_user_id      INTEGER,
    report_date           TIMESTAMP,
    attachment            TEXT,
    tags                  VARCHAR(255),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP
);

CREATE INDEX idx_see_seek_looks_see_seek_id ON see_seek_looks(see_seek_id);
CREATE INDEX idx_see_seek_listens_see_seek_id ON see_seek_listens(see_seek_id);
CREATE INDEX idx_see_seek_smells_see_seek_id ON see_seek_smells(see_seek_id);
CREATE INDEX idx_see_seek_touches_see_seek_id ON see_seek_touches(see_seek_id);
CREATE TABLE see_seek_examination_items (
    id                              SERIAL PRIMARY KEY,
    uuid                            VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    classify                        INTEGER DEFAULT 1,
    see_seek_examination_report_id  INTEGER REFERENCES see_seek_examination_reports(id),
    item_code                       VARCHAR(50),
    item_name                       VARCHAR(255) NOT NULL,
    result                          TEXT,
    result_unit                     VARCHAR(50),
    reference_min                   DECIMAL(10,4),
    reference_max                   DECIMAL(10,4),
    remark                          TEXT,
    abnormal_flag                   BOOLEAN DEFAULT false,
    method                          VARCHAR(255),
    operator                        VARCHAR(255),
    status                          INTEGER DEFAULT 0,
    created_at                      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                      TIMESTAMP
);

CREATE INDEX idx_see_seek_examination_reports_see_seek_id ON see_seek_examination_reports(see_seek_id);
CREATE INDEX idx_see_seek_examination_items_report_id ON see_seek_examination_items(see_seek_examination_report_id);