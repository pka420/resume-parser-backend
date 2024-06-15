BEGIN;

CREATE TYPE user_type AS ENUM (
    'admin',
    'user'
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    email VARCHAR(50) NOT NULL,
    address VARCHAR(200) NOT NULL,
    type user_type NOT NULL,
    password_hash VARCHAR(200) NOT NULL,
    profile_headline VARCHAR(200) NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE profile (
    id SERIAL PRIMARY KEY,
    applicant INT REFERENCES users(id),
    resume_file_address VARCHAR(200) NOT NULL,
    skills VARCHAR(200),
    education VARCHAR(200),
    experience VARCHAR(200),
    name VARCHAR(50),
    email VARCHAR(50),
    phone VARCHAR(50),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    title VARCHAR(50) NOT NULL,
    description VARCHAR(200) NOT NULL,
    posted_on TIMESTAMP NOT NULL,
    total_applications INT NOT NULL,
    applicants INTEGER[],
    company_name VARCHAR(50) NOT NULL,
    posted_by INT REFERENCES users(id)
);

CREATE OR REPLACE FUNCTION set_created_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.created_at := CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_users_created_at
BEFORE INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION set_created_at();

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER set_profile_created_at
BEFORE INSERT ON profile
FOR EACH ROW
EXECUTE FUNCTION set_created_at();

CREATE TRIGGER update_profile_updated_at
BEFORE UPDATE ON profile
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();


COMMIT;
