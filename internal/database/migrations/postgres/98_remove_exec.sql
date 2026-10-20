UPDATE release_action_status
SET action_id = NULL
WHERE action_id IN (SELECT id FROM action WHERE type = 'EXEC');

DELETE
FROM action
WHERE type = 'EXEC';

DELETE
FROM filter_external
WHERE type = 'EXEC';

ALTER TABLE action
    DROP COLUMN exec_cmd,
    DROP COLUMN exec_args;

ALTER TABLE filter_external
    DROP COLUMN exec_cmd,
    DROP COLUMN exec_args,
    DROP COLUMN exec_expect_status;
