-- migration: up
CREATE TABLE public.bar(
    id SERIAL    
);
-- migration: down
DROP TABLE public.bar;