-- migration: up
ALTER TABLE public.foo ADD COLUMN name varchar(255);
-- migration: down
ALTER TABLE public.foo DROP COLUMN name;