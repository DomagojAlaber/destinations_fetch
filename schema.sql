CREATE TABLE destination (
  id serial PRIMARY KEY,
  name text UNIQUE,
  region text,
  lon double precision,
  lat double precision
);
