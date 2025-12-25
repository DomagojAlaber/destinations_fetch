-- name: GetDestination :one
select * from destination
where id = $1 limit 1;

-- name: GetDestinations :many
select * from destination
order by id desc;

-- name: UpsertDestination :one
insert into destination (name, region, lon, lat)
values ($1, $2, $3, $4)
ON CONFLICT (name) DO UPDATE
SET lon = EXCLUDED.lon,
    lat = EXCLUDED.lat
RETURNING id;
