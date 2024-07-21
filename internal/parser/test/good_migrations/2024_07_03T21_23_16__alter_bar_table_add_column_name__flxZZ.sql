-- migration: up
ALTER TABLE public.bar ADD COLUMN name varchar(255);
-- migration: down
ALTER TABLE public.bar DROP COLUMN name;