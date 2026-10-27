DROP TABLE IF EXISTS simulation_answer_trace;
DROP TABLE IF EXISTS knowledge_mastery;

-- Keep test scores and answers while removing retired inference and Q-matrix data.
UPDATE simulation_attempt
SET result_payload = (result_payload::jsonb
  - 'knowledge_mastery'
  - 'q_matrix'
  - 'q_matrix_by_question'
  - 'q_matrix_by_tag')::text
WHERE result_payload IS NOT NULL
  AND result_payload::jsonb ?| ARRAY[
    'knowledge_mastery', 'q_matrix', 'q_matrix_by_question', 'q_matrix_by_tag'
  ];
