CREATE TABLE IF NOT EXISTS task_recurrences (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    start_date DATE NOT NULL,
    every_n_days INTEGER,
    day_of_month INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_task_recurrences_type
        CHECK (type IN ('daily', 'monthly', 'specific_dates', 'even_days', 'odd_days')),

    CONSTRAINT chk_task_recurrences_every_n_days
        CHECK (
            (type = 'daily' AND every_n_days IS NOT NULL AND every_n_days > 0)
            OR
            (type <> 'daily' AND every_n_days IS NULL)
        ),

    CONSTRAINT chk_task_recurrences_day_of_month
        CHECK (
            (type = 'monthly' AND day_of_month IS NOT NULL AND day_of_month BETWEEN 1 AND 30)
            OR
            (type <> 'monthly' AND day_of_month IS NULL)
        )
);

COMMENT ON TABLE task_recurrences IS 'Настройки периодичности задачи';
COMMENT ON COLUMN task_recurrences.id IS 'Первичный ключ настройки периодичности';
COMMENT ON COLUMN task_recurrences.task_id IS 'Идентификатор задачи, к которой относится периодичность';
COMMENT ON COLUMN task_recurrences.type IS 'Тип периодичности: daily, monthly, specific_dates, even_days, odd_days';
COMMENT ON COLUMN task_recurrences.start_date IS 'Дата начала действия правила периодичности';
COMMENT ON COLUMN task_recurrences.every_n_days IS 'Интервал в днях для типа daily';
COMMENT ON COLUMN task_recurrences.day_of_month IS 'День месяца для типа monthly, допустимые значения от 1 до 30';
COMMENT ON COLUMN task_recurrences.is_active IS 'Признак активности правила периодичности';
COMMENT ON COLUMN task_recurrences.created_at IS 'Дата и время создания записи';
COMMENT ON COLUMN task_recurrences.updated_at IS 'Дата и время последнего обновления записи';

CREATE INDEX IF NOT EXISTS idx_task_recurrences_task_id
    ON task_recurrences (task_id);

CREATE INDEX IF NOT EXISTS idx_task_recurrences_type
    ON task_recurrences (type);

CREATE INDEX IF NOT EXISTS idx_task_recurrences_is_active
    ON task_recurrences (is_active);



CREATE TABLE IF NOT EXISTS task_recurrence_dates (
    id BIGSERIAL PRIMARY KEY,
    recurrence_id BIGINT NOT NULL REFERENCES task_recurrences(id) ON DELETE CASCADE,
    run_date DATE NOT NULL,

    CONSTRAINT uq_task_recurrence_dates_recurrence_id_run_date
        UNIQUE (recurrence_id, run_date)
);

COMMENT ON TABLE task_recurrence_dates IS 'Конкретные даты выполнения для периодичности типа specific_dates';
COMMENT ON COLUMN task_recurrence_dates.id IS 'Первичный ключ записи конкретной даты';
COMMENT ON COLUMN task_recurrence_dates.recurrence_id IS 'Идентификатор настройки периодичности';
COMMENT ON COLUMN task_recurrence_dates.run_date IS 'Конкретная дата, на которую должно быть создано выполнение задачи';

CREATE INDEX IF NOT EXISTS idx_task_recurrence_dates_recurrence_id
    ON task_recurrence_dates (recurrence_id);

CREATE INDEX IF NOT EXISTS idx_task_recurrence_dates_run_date
    ON task_recurrence_dates (run_date);



CREATE TABLE IF NOT EXISTS task_occurrences (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    recurrence_id BIGINT REFERENCES task_recurrences(id) ON DELETE SET NULL,
    scheduled_date DATE NOT NULL,
    status TEXT NOT NULL,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_task_occurrences_task_id_scheduled_date
        UNIQUE (task_id, scheduled_date),

    CONSTRAINT chk_task_occurrences_status
        CHECK (status IN ('new', 'in_progress', 'done'))
);

COMMENT ON TABLE task_occurrences IS 'Конкретные экземпляры выполнения задачи на определенную дату';
COMMENT ON COLUMN task_occurrences.id IS 'Первичный ключ экземпляра задачи';
COMMENT ON COLUMN task_occurrences.task_id IS 'Идентификатор задачи, для которой создан экземпляр';
COMMENT ON COLUMN task_occurrences.recurrence_id IS 'Идентификатор правила периодичности, по которому создан экземпляр';
COMMENT ON COLUMN task_occurrences.scheduled_date IS 'Дата, на которую запланирован экземпляр задачи';
COMMENT ON COLUMN task_occurrences.status IS 'Статус экземпляра задачи: new, in_progress, done';
COMMENT ON COLUMN task_occurrences.completed_at IS 'Дата и время завершения экземпляра задачи';
COMMENT ON COLUMN task_occurrences.created_at IS 'Дата и время создания записи';
COMMENT ON COLUMN task_occurrences.updated_at IS 'Дата и время последнего обновления записи';

CREATE INDEX IF NOT EXISTS idx_task_occurrences_task_id
    ON task_occurrences (task_id);

CREATE INDEX IF NOT EXISTS idx_task_occurrences_recurrence_id
    ON task_occurrences (recurrence_id);

CREATE INDEX IF NOT EXISTS idx_task_occurrences_scheduled_date
    ON task_occurrences (scheduled_date);

CREATE INDEX IF NOT EXISTS idx_task_occurrences_status
    ON task_occurrences (status);