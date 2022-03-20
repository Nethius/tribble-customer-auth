CREATE TABLE users
(
    id       integer GENERATED BY DEFAULT AS IDENTITY,
    email    text        NOT NULL UNIQUE,
    password varchar(60) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE devices
(
    user_id integer,
    imei    varchar(15) NOT NULL,
    PRIMARY KEY (user_id, imei),
    FOREIGN KEY (user_id)
        REFERENCES users (id)
);
