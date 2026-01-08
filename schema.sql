CREATE TABLE destination (
  id serial PRIMARY KEY,
  name text UNIQUE,
  description text,
  region text,
  lon double precision,
  lat double precision
);
