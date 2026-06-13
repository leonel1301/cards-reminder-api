-- Migrate salary_day from users to self owner, then drop column from users.
-- Run only if you already added salary_day to users (migration 004 previous version).

UPDATE owners o
SET salary_day = u.salary_day, updated_at = now()
FROM users u
WHERE o.user_id = u.id
  AND o.is_self = true
  AND u.salary_day IS NOT NULL
  AND o.salary_day IS NULL;

ALTER TABLE users DROP COLUMN IF EXISTS salary_day;
