-- Add new granular permission columns to board_permissions
ALTER TABLE board_permissions 
ADD COLUMN IF NOT EXISTS can_view_board BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS can_view_task_list BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS can_view_tasks BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS can_edit_tasks BOOLEAN DEFAULT false;

-- Migrate existing data: can_view -> can_view_board and can_view_tasks
UPDATE board_permissions 
SET 
    can_view_board = can_view,
    can_view_task_list = can_view,
    can_view_tasks = can_view,
    can_edit_tasks = can_create
WHERE can_view IS NOT NULL OR can_create IS NOT NULL;

-- Drop old columns
ALTER TABLE board_permissions 
DROP COLUMN IF EXISTS can_view,
DROP COLUMN IF EXISTS can_create;
