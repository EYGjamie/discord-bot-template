-- Create boards table
CREATE TABLE IF NOT EXISTS boards (
    id SERIAL PRIMARY KEY,
    guild_id VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    color VARCHAR(7) DEFAULT '#6aa6ff',
    position INTEGER DEFAULT 0,
    created_by VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_boards_guild_id ON boards(guild_id);

-- Create board_permissions table
CREATE TABLE IF NOT EXISTS board_permissions (
    id SERIAL PRIMARY KEY,
    board_id INTEGER NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    role_id VARCHAR(20),
    user_id VARCHAR(20),
    can_view BOOLEAN DEFAULT false,
    can_create BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT board_permissions_role_or_user CHECK (
        (role_id IS NOT NULL AND user_id IS NULL) OR
        (role_id IS NULL AND user_id IS NOT NULL)
    )
);

CREATE INDEX idx_board_permissions_board_id ON board_permissions(board_id);
CREATE INDEX idx_board_permissions_role_id ON board_permissions(role_id);
CREATE INDEX idx_board_permissions_user_id ON board_permissions(user_id);

-- Create task_groups table
CREATE TABLE IF NOT EXISTS task_groups (
    id SERIAL PRIMARY KEY,
    guild_id VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    color VARCHAR(7) DEFAULT '#39d98a',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_groups_guild_id ON task_groups(guild_id);

-- Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    board_id INTEGER NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    group_id INTEGER REFERENCES task_groups(id) ON DELETE SET NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT DEFAULT '',
    status VARCHAR(20) DEFAULT 'todo',
    position INTEGER DEFAULT 0,
    assignee_id VARCHAR(20),
    due_date TIMESTAMP,
    tags TEXT DEFAULT '[]',
    created_by VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_board_id ON tasks(board_id);
CREATE INDEX idx_tasks_group_id ON tasks(group_id);
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX idx_tasks_status ON tasks(status);

-- Create task_group_permissions table
CREATE TABLE IF NOT EXISTS task_group_permissions (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES task_groups(id) ON DELETE CASCADE,
    role_id VARCHAR(20),
    user_id VARCHAR(20),
    permission VARCHAR(20) DEFAULT 'none',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT task_group_permissions_role_or_user CHECK (
        (role_id IS NOT NULL AND user_id IS NULL) OR
        (role_id IS NULL AND user_id IS NOT NULL)
    )
);

CREATE INDEX idx_task_group_permissions_group_id ON task_group_permissions(group_id);
CREATE INDEX idx_task_group_permissions_role_id ON task_group_permissions(role_id);
CREATE INDEX idx_task_group_permissions_user_id ON task_group_permissions(user_id);
