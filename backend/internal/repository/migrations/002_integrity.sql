ALTER TABLE content_categories DROP CONSTRAINT content_categories_content_id_fkey;
ALTER TABLE content_categories ADD CONSTRAINT content_categories_content_id_fkey FOREIGN KEY(content_id) REFERENCES contents(id) ON DELETE CASCADE;

DO $$
DECLARE fk record;
BEGIN
  FOR fk IN SELECT conrelid::regclass AS relation, conname FROM pg_constraint WHERE contype='f' AND connamespace=current_schema()::regnamespace
  LOOP
    EXECUTE format('ALTER TABLE %s ALTER CONSTRAINT %I DEFERRABLE INITIALLY IMMEDIATE',fk.relation,fk.conname);
  END LOOP;
END $$;

ALTER TABLE instance_tasks ADD CHECK(status IN ('pending','processing','completed','failed'));
CREATE INDEX idx_tasks_outbox ON instance_tasks(published_at) WHERE status IN ('pending','processing');
CREATE UNIQUE INDEX uq_active_attempt ON simulation_attempt(user_id) WHERE status='in_progress';
ALTER TABLE practice_questions ADD CONSTRAINT uq_practice_question UNIQUE(practice_id,question_id);
CREATE UNIQUE INDEX uq_instance_ssh_user ON instances(ssh_username);
