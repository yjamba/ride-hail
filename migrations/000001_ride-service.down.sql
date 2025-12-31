begin;

-- Drop tables in reverse order of creation (respecting foreign key constraints)
drop table if exists ride_events cascade;
drop table if exists ride_event_type cascade;
drop table if exists rides cascade;
drop index if exists idx_coordinates_current;
drop index if exists idx_coordinates_entity;
drop table if exists coordinates cascade;
drop table if exists vehicle_type cascade;
drop table if exists ride_status cascade;
drop table if exists user_status cascade;
drop table if exists roles cascade;
drop index if exists idx_rides_status;
drop table if exists users cascade;

commit;
