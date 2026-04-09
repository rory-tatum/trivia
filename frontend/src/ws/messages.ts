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

export interface HostEndRoundMsg {
  event: "host_end_round";
  payload: { round_index: number };
}

export interface HostBeginScoringMsg {
  event: "host_begin_scoring";
  payload: { round_index: number };
}

export interface HostMarkAnswerMsg {
  event: "host_mark_answer";
  payload: {
    team_id: string;
    round_index: number;
    question_index: number;
    verdict: "correct" | "incorrect";
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
  | HostEndRoundMsg
  | HostBeginScoringMsg
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
  payload: { team_id: string; device_token: string };
}

export interface TeamJoinedMsg {
  event: "team_joined";
  payload: { team_id: string; team_name: string };
}

export interface QuizLoadedMsg {
  event: "quiz_loaded";
  payload: {
    title: string;
    round_count: number;
    question_count: number;
    player_url: string;
    display_url: string;
    confirmation: string;
    session_id: string;
  };
}

export interface RoundStartedMsg {
  event: "round_started";
  payload: { round_index: number; question_count: number };
}

export interface QuestionRevealedMsg {
  event: "question_revealed";
  payload: {
    question: { text: string; index: number };
    revealed_count: number;
    total_questions: number;
  };
}

export interface SubmissionAckMsg {
  event: "submission_ack";
  payload: { team_id: string; round_index: number; locked: boolean };
}

export interface SubmissionReceivedMsg {
  event: "submission_received";
  payload: { team_id: string; team_name: string; round_index: number };
}

export interface ScoringOpenedMsg {
  event: "scoring_opened";
  payload: { round_index: number };
}

export interface TeamSubmission {
  team_id: string;
  team_name: string;
  answer: string;
}

export interface ScoringQuestion {
  question_index: number;
  text: string;
  correct_answer: string;
  submissions: TeamSubmission[];
}

export interface ScoringDataMsg {
  event: "scoring_data";
  payload: { round_index: number; questions: ScoringQuestion[] };
}

export interface ScoreUpdatedMsg {
  event: "score_updated";
  payload: { team_id: string; round_index: number; running_total: number };
}

export interface CeremonyQuestionShownMsg {
  event: "ceremony_question_shown";
  payload: { question_index: number; question: { text: string; index: number } };
}

export interface CeremonyAnswerRevealedMsg {
  event: "ceremony_answer_revealed";
  payload: { question_index: number; answer: string };
}

export interface RoundScoresPublishedMsg {
  event: "round_scores_published";
  payload: { round_index: number; scores: Record<string, number> };
}

export interface GameOverMsg {
  event: "game_over";
  payload: { final_scores: Record<string, number> };
}

export interface StateSnapshotMsg {
  event: "state_snapshot";
  payload: Record<string, unknown>;
}

export interface ErrorMsg {
  event: "error";
  payload: { code: string; message: string };
}

export type OutgoingMessage =
  | SessionCreatedMsg
  | TeamRegisteredMsg
  | TeamJoinedMsg
  | QuizLoadedMsg
  | RoundStartedMsg
  | QuestionRevealedMsg
  | SubmissionAckMsg
  | SubmissionReceivedMsg
  | ScoringOpenedMsg
  | ScoringDataMsg
  | ScoreUpdatedMsg
  | CeremonyQuestionShownMsg
  | CeremonyAnswerRevealedMsg
  | RoundScoresPublishedMsg
  | GameOverMsg
  | StateSnapshotMsg
  | ErrorMsg;
