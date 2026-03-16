INSERT INTO teams (id, name, description) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Nairo', 'Equipo Nairo - Joseph, Mauri, Esteban, Sebas'),
    ('00000000-0000-0000-0000-000000000002', 'Botero', 'Equipo Botero'),
    ('00000000-0000-0000-0000-000000000003', 'Egan', 'Equipo Egan'),
    ('00000000-0000-0000-0000-000000000004', 'Raiders', 'Equipo Raiders')
ON CONFLICT (name) DO NOTHING;
