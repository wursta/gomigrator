-- migration: up
CREATE TABLE public.foo(
    id SERIAL    
);
-- migration: down
DROP TABLE public.foo;