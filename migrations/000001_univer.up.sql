CREATE TABLE if NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(64) unique,
    email TEXT NOT NULL,
    phone_number VARCHAR(20) unique,
    password TEXT NOT NULL,
    bio TEXT, 
    image_url text,
    refresh_token TEXT,
    role VARCHAR(10) NOT NULL, -- admin, user
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
insert into users (id, username, email, phone_number, password, bio, image_url, refresh_token, role)
values
('19d16003-586a-4190-92ee-ab0c45504023', 'Abdulaziz20', 'abdulazizxoshimov22@gmail.com', '998900556638', '$2a$14$3kk4Dfr4MumGCsQLbxLmuO0iICIyGor3GBSik6/DbeF47zZEDTusO', 'Bu dunyodagi ayriliq bari soxta', '', null, 'admin');

CREATE TABLE category (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL unique,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);


CREATE TABLE posts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL, 
    theme TEXT,
    path TEXT NOT NULL,
    views bigint DEFAULT 0,
    science VARCHAR(50), -- from which subject
    category_id UUID NOT NULL,  -- slayd, mustaqil ish va hk
    price FLOAT NOT NULL DEFAULT 0,
    price_status boolean NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    foreign key  (user_id) references users(id),
    foreign key (category_id) references category(id)
);

CREATE TABLE if not exists comments (
    id UUID PRIMARY KEY, 
    owner_id UUID  NOT NULL,
    post_id UUID NOT NULL, 
    message text NOT NULL,
    likes  bigint DEFAULT 0,
    dislikes bigint DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    foreign key (owner_id) references users(id),
    foreign key (post_id) references posts(id)
);

CREATE TABLE likes (
    comment_id UUID not NULL,
    owner_id  uuid not null, 
    post_id UUID not null,
    status boolean,
    foreign key (comment_id) references comments(id),
    foreign key (post_id)references posts(id),
    foreign key(owner_id) references users(id)
);

CREATE TABLE views (
    user_id uuid not null,
    post_id uuid not null
); 

CREATE TABLE if not exists comments (
    id UUID PRIMARY KEY, 
     owner_id UUID  NOT NULL,
    post_id UUID NOT NULL, 
     message text NOT NULL,
     likes  bigint DEFAULT 0,
     dislikes bigint DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    foreign key (owner_id) references users(id),
    foreign key (post_id) references posts(id)
);

