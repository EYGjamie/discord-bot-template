-- Add can_edit_board permission to board_permissions table
ALTER TABLE board_permissions
ADD COLUMN can_edit_board BOOLEAN DEFAULT FALSE;

-- Migrate existing data: if user can edit tasks, they should be able to edit the board too
UPDATE board_permissions
SET can_edit_board = can_edit_tasks
WHERE can_edit_tasks = TRUE;
