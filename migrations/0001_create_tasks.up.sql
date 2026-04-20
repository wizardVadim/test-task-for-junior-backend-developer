CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE tasks IS 'Основная сущность задачи';
COMMENT ON COLUMN tasks.id IS 'Первичный ключ задачи';
COMMENT ON COLUMN tasks.title IS 'Название задачи';
COMMENT ON COLUMN tasks.description IS 'Описание задачи';
COMMENT ON COLUMN tasks.status IS 'Текущий агрегированный статус задачи: new, in_progress, done';
COMMENT ON COLUMN tasks.created_at IS 'Дата и время создания задачи';
COMMENT ON COLUMN tasks.updated_at IS 'Дата и время последнего обновления задачи';

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
