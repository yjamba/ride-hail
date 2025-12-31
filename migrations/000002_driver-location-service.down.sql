begin;

-- Drop tables in reverse order of creation (respecting foreign key constraints)
drop table if exists location_history cascade;
drop table if exists driver_sessions cascade;
drop index if exists idx_drivers_status;
drop table if exists drivers cascade;
drop table if exists driver_status cascade;

commit;
