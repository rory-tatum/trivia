// Incoming messages (client -> server)

export interface TeamRegisterMsg {
  event: "team_register";
  payload: { team_name: string };
}

export interface TeamRejoinMsg {
  event: "team_rejoin";
  payload: { team_id: string; device_token: string };
}

export interface DraftAnswerMsg {
  event: "draft_answer";
  payload: {
    team_name: string;
    round_index: number;
    question_index: number;
    answer: string;
  };
}

export interface SubmitAnswersMsg {
  event: "submit_answers";
  payload: {
    team_id: string;
    round_index: number;
    answers: Array<{ question_index: number; answer: string }>;
  };
}

export interface HostLoadQuizMsg {
  event: "host_load_quiz";
  payload: { file_path: string };
}

export interface HostStartRoundMsg {
  event: "host_start_round";
  payload: { round_index: number };
}

export interface HostRevealQuestionMsg {
  event: "host_reveal_question";
  payload: { round_index: number; question_index: number };
}

export interface HostMarkAnswerMsg {
  event: "host_mark_answer";
  payload: {
    team_id: string;
    round_index: number;
    question_index: number;
    verdict: string;
  };
}

export interface HostCeremonyShowQuestionMsg {
  event: "host_ceremony_show_question";
  payload: { question_index: number };
}

export interface HostCeremonyRevealAnswerMsg {
  event: "host_ceremony_reveal_answer";
  payload: { question_index: number };
}

export interface HostPublishScoresMsg {
  event: "host_publish_scores";
  payload: { round_index: number };
}

export interface HostEndGameMsg {
  event: "host_end_game";
  payload: Record<string, never>;
}

export type IncomingMessage =
  | TeamRegisterMsg
  | TeamRejoinMsg
  | DraftAnswerMsg
  | SubmitAnswersMsg
  | HostLoadQuizMsg
  | HostStartRoundMsg
  | HostRevealQuestionMsg
  | HostMarkAnswerMsg
  | HostCeremonyShowQuestionMsg
  | HostCeremonyRevealAnswerMsg
  | HostPublishScoresMsg
  | HostEndGameMsg;

// Outgoing messages (server -> client)

export interface SessionCreatedMsg {
  event: "session_created";
  payload: { session_id: string };
}

export interface TeamRegisteredMsg {
  event: "team_registered";
  payload: { team_id: string; team_name: string; device_token: string };
}

export interface RoundStartedMsg {
  event: "round_started";
  payload: { round_index: number; round_name: string };
}

export interface QuestionRevealedMsg {
  event: "question_revealed";
  payload: { round_index: number; question_index: number; text: string };
}

export interface SubmissionAckMsg {
  event: "submission_ack";
  payload: { team_id: string; round_index: number };
}

export interface ScoreUpdatedMsg {
  event: "score_updated";
  payload: { team_id: string; team_name: string; score: number };
}

export interface CeremonyQuestionShownMsg {
  event: "ceremony_question_shown";
  payload: { question_index: number; text: string };
}

export interface CeremonyAnswerRevealedMsg {
  event: "ceremony_answer_revealed";
  payload: { question_index: number; answer: string };
}

export interface ScoresPublishedMsg {
  event: "scores_published";
  payload: {
    round_index: number;
    scores: Array<{ team_id: string; team_name: string; score: number }>;
  };
}

export interface GameOverMsg {
  event: "game_over";
  payload: {
    scores: Array<{ team_id: string; team_name: string; score: number }>;
  };
}

export interface StateSnapshotMsg {
  event: "state_snapshot";
  payload: Record<string, unknown>;
}

export interface ErrorMsg {
  event: "error";
  payload: { message: string };
}

export type OutgoingMessage =
  | SessionCreatedMsg
  | TeamRegisteredMsg
  | RoundStartedMsg
  | QuestionRevealedMsg
  | SubmissionAckMsg
  | ScoreUpdatedMsg
  | CeremonyQuestionShownMsg
  | CeremonyAnswerRevealedMsg
  | ScoresPublishedMsg
  | GameOverMsg
  | StateSnapshotMsg
  | ErrorMsg;
