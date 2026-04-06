// Event type constants — align with backend hub/events.go

// Incoming (client -> server)
export const TEAM_REGISTER = "team_register" as const;
export const TEAM_REJOIN = "team_rejoin" as const;
export const DRAFT_ANSWER = "draft_answer" as const;
export const SUBMIT_ANSWERS = "submit_answers" as const;
export const HOST_LOAD_QUIZ = "host_load_quiz" as const;
export const HOST_START_ROUND = "host_start_round" as const;
export const HOST_REVEAL_QUESTION = "host_reveal_question" as const;
export const HOST_MARK_ANSWER = "host_mark_answer" as const;
export const HOST_CEREMONY_SHOW_QUESTION = "host_ceremony_show_question" as const;
export const HOST_CEREMONY_REVEAL_ANSWER = "host_ceremony_reveal_answer" as const;
export const HOST_PUBLISH_SCORES = "host_publish_scores" as const;
export const HOST_END_GAME = "host_end_game" as const;

// Outgoing (server -> client)
export const SESSION_CREATED = "session_created" as const;
export const TEAM_REGISTERED = "team_registered" as const;
export const ROUND_STARTED = "round_started" as const;
export const QUESTION_REVEALED = "question_revealed" as const;
export const SUBMISSION_ACK = "submission_ack" as const;
export const SCORE_UPDATED = "score_updated" as const;
export const CEREMONY_QUESTION_SHOWN = "ceremony_question_shown" as const;
export const CEREMONY_ANSWER_REVEALED = "ceremony_answer_revealed" as const;
export const SCORES_PUBLISHED = "scores_published" as const;
export const GAME_OVER = "game_over" as const;
export const STATE_SNAPSHOT = "state_snapshot" as const;
export const ERROR = "error" as const;

// Internal client events
export const RECONNECT_FAILED = "reconnect_failed" as const;
