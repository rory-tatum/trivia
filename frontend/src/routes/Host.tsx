import { useEffect, useRef, useState } from "react";
import { WsClient } from "../ws/client";
import { CONNECTED, AUTH_FAILED, RECONNECT_FAILED } from "../ws/events";
import type { OutgoingMessage, ScoringQuestion } from "../ws/messages";

function getHostWsUrl(): string {
  const params = new URLSearchParams(window.location.search);
  const token = params.get("token") ?? "";
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/ws?token=${token}`;
}

type ConnStatus =
  | "connecting"
  | "connected"
  | "auth_failed"
  | "reconnecting"
  | "reconnect_failed";

type Phase =
  | "lobby"
  | "quiz_loaded"
  | "round_active"
  | "round_ended"
  | "scoring"
  | "round_scores"
  | "ceremony"
  | "game_over";

interface QuizMeta {
  title: string;
  round_count: number;
  player_url: string;
  display_url: string;
  confirmation: string;
}

interface Team {
  id: string;
  name: string;
}

export default function Host(): JSX.Element {
  const clientRef = useRef<WsClient | null>(null);

  const [connStatus, setConnStatus] = useState<ConnStatus>("connecting");
  const [phase, setPhase] = useState<Phase>("lobby");
  const [error, setError] = useState<string | null>(null);

  // Persistent data
  const [teams, setTeams] = useState<Team[]>([]);
  const [quizMeta, setQuizMeta] = useState<QuizMeta | null>(null);
  const [submissionCount, setSubmissionCount] = useState(0);

  // Round data
  const [currentRound, setCurrentRound] = useState(0);
  const [roundQuestionCount, setRoundQuestionCount] = useState(0);
  const [revealedCount, setRevealedCount] = useState(0);
  const [nextQuestionIndex, setNextQuestionIndex] = useState(0);

  // Scoring data
  const [scoringData, setScoringData] = useState<ScoringQuestion[]>([]);
  const [verdicts, setVerdicts] = useState<Record<string, "correct" | "incorrect">>({});
  const [runningTotals, setRunningTotals] = useState<Record<string, number>>({});

  // Round scores / game over data
  const [publishedScores, setPublishedScores] = useState<Record<string, number>>({});
  const [finalScores, setFinalScores] = useState<Record<string, number>>({});

  // Ceremony state (host-local, server only broadcasts to display/play)
  const [ceremonyIndex, setCeremonyIndex] = useState(0);
  const [ceremonyAnswerRevealed, setCeremonyAnswerRevealed] = useState(false);

  // Load quiz form
  const [filePath, setFilePath] = useState("");

  const send = (msg: Parameters<WsClient["send"]>[0]) => {
    clientRef.current?.send(msg);
  };

  useEffect(() => {
    const client = new WsClient(getHostWsUrl());
    clientRef.current = client;

    client.on(CONNECTED, () => setConnStatus("connected"));
    client.on(AUTH_FAILED, () => setConnStatus("auth_failed"));
    client.on(RECONNECT_FAILED, () => setConnStatus("reconnect_failed"));

    client.onMessage((msg: OutgoingMessage) => {
      setError(null);

      switch (msg.event) {
        case "quiz_loaded":
          setQuizMeta({
            title: msg.payload.title,
            round_count: msg.payload.round_count,
            player_url: msg.payload.player_url,
            display_url: msg.payload.display_url,
            confirmation: msg.payload.confirmation,
          });
          setCurrentRound(0);
          setSubmissionCount(0);
          setPhase("quiz_loaded");
          break;

        case "team_joined":
          setTeams((prev) => [
            ...prev,
            { id: msg.payload.team_id, name: msg.payload.team_name },
          ]);
          break;

        case "submission_received":
          setSubmissionCount((prev) => prev + 1);
          break;

        case "round_started":
          setRevealedCount(0);
          setNextQuestionIndex(0);
          setRoundQuestionCount(msg.payload.question_count);
          setCurrentRound(msg.payload.round_index);
          setSubmissionCount(0);
          setPhase("round_active");
          break;

        case "question_revealed":
          setRevealedCount(msg.payload.revealed_count);
          setRoundQuestionCount(msg.payload.total_questions);
          setNextQuestionIndex(msg.payload.question.index + 1);
          break;

        case "scoring_data":
          setScoringData(msg.payload.questions);
          setVerdicts({});
          setRunningTotals({});
          setPhase("scoring");
          break;

        case "score_updated":
          setRunningTotals((prev) => ({
            ...prev,
            [msg.payload.team_id]: msg.payload.running_total,
          }));
          break;

        case "round_scores_published":
          setPublishedScores(msg.payload.scores);
          setPhase("round_scores");
          break;

        case "game_over":
          setFinalScores(msg.payload.final_scores);
          setPhase("game_over");
          break;

        case "error":
          setError(msg.payload.message);
          break;
      }
    });

    client.connect();

    return () => {
      client.disconnect();
    };
  }, []);

  // --- Connection status gate ---
  if (connStatus === "auth_failed") {
    return (
      <div>
        <h1>Quizmaster Panel</h1>
        <p>
          <strong>Invalid host token.</strong> Check the token in your URL and
          reload the page.
        </p>
      </div>
    );
  }

  if (connStatus === "reconnect_failed") {
    return (
      <div>
        <h1>Quizmaster Panel</h1>
        <p>Connection lost and could not be re-established.</p>
        <button onClick={() => window.location.reload()}>Reload</button>
      </div>
    );
  }

  if (connStatus === "connecting") {
    return (
      <div>
        <h1>Quizmaster Panel</h1>
        <p>Connecting...</p>
      </div>
    );
  }

  // --- Phase renders ---

  const renderLobby = () => (
    <div>
      <h2>Load Quiz</h2>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (!filePath.trim()) return;
          send({ event: "host_load_quiz", payload: { file_path: filePath.trim() } });
        }}
      >
        <label>
          Quiz file path:{" "}
          <input
            type="text"
            value={filePath}
            onChange={(e) => setFilePath(e.target.value)}
            placeholder="/path/to/quiz.yaml"
            size={40}
          />
        </label>{" "}
        <button type="submit">Load Quiz</button>
      </form>
      {error && <p><strong>Error:</strong> {error}</p>}
      {teams.length > 0 && (
        <p>Teams in lobby: {teams.map((t) => t.name).join(", ")}</p>
      )}
    </div>
  );

  const renderQuizLoaded = () => {
    if (!quizMeta) return null;
    return (
      <div>
        <h2>Quiz Ready</h2>
        <p>{quizMeta.confirmation}</p>
        <p>Players join at: <strong>{quizMeta.player_url}</strong></p>
        <p>Display screen: <strong>{quizMeta.display_url}</strong></p>
        {teams.length > 0 && (
          <p>Teams: {teams.map((t) => t.name).join(", ")}</p>
        )}
        {error && <p><strong>Error:</strong> {error}</p>}
        <button
          onClick={() => {
            send({ event: "host_start_round", payload: { round_index: 0 } });
          }}
        >
          Start Round 1
        </button>
      </div>
    );
  };

  const renderRoundActive = () => (
    <div>
      <h2>Round {currentRound + 1}</h2>
      <p>
        Questions revealed: {revealedCount} / {roundQuestionCount}
      </p>
      <p>Submissions received: {submissionCount}</p>
      {error && <p><strong>Error:</strong> {error}</p>}
      {revealedCount < roundQuestionCount && (
        <button
          onClick={() => {
            send({
              event: "host_reveal_question",
              payload: { round_index: currentRound, question_index: nextQuestionIndex },
            });
          }}
        >
          Reveal Question {nextQuestionIndex + 1}
        </button>
      )}
      {revealedCount > 0 && revealedCount === roundQuestionCount && (
        <button
          onClick={() => {
            send({ event: "host_end_round", payload: { round_index: currentRound } });
            setPhase("round_ended");
          }}
        >
          End Round
        </button>
      )}
    </div>
  );

  const renderRoundEnded = () => (
    <div>
      <h2>Round {currentRound + 1} — Closed</h2>
      <p>Submissions received: {submissionCount} of {teams.length} teams</p>
      {error && <p><strong>Error:</strong> {error}</p>}
      <button
        onClick={() => {
          send({ event: "host_begin_scoring", payload: { round_index: currentRound } });
        }}
      >
        Begin Scoring
      </button>
    </div>
  );

  const renderScoring = () => {
    const verdictKey = (teamId: string, qi: number) => `${teamId}-${qi}`;
    const allMarked =
      scoringData.length > 0 &&
      scoringData.every((q) =>
        q.submissions.every((s) => verdictKey(s.team_id, q.question_index) in verdicts)
      );

    return (
      <div>
        <h2>Scoring — Round {currentRound + 1}</h2>
        {error && <p><strong>Error:</strong> {error}</p>}
        {scoringData.map((q) => (
          <div key={q.question_index} style={{ marginBottom: "1em", borderBottom: "1px solid #ccc", paddingBottom: "0.5em" }}>
            <p>
              <strong>Q{q.question_index + 1}:</strong> {q.text}
            </p>
            <p>
              <em>Correct answer: {q.correct_answer}</em>
            </p>
            <table>
              <thead>
                <tr>
                  <th>Team</th>
                  <th>Answer</th>
                  <th>Verdict</th>
                  <th>Running Total</th>
                </tr>
              </thead>
              <tbody>
                {q.submissions.map((s) => {
                  const key = verdictKey(s.team_id, q.question_index);
                  const verdict = verdicts[key];
                  return (
                    <tr key={s.team_id}>
                      <td>{s.team_name}</td>
                      <td>{s.answer || <em>no answer</em>}</td>
                      <td>
                        {verdict ? (
                          <span>{verdict === "correct" ? "✓ Correct" : "✗ Wrong"}</span>
                        ) : (
                          <>
                            <button
                              onClick={() => {
                                setVerdicts((prev) => ({ ...prev, [key]: "correct" }));
                                send({
                                  event: "host_mark_answer",
                                  payload: {
                                    team_id: s.team_id,
                                    round_index: currentRound,
                                    question_index: q.question_index,
                                    verdict: "correct",
                                  },
                                });
                              }}
                            >
                              Correct
                            </button>{" "}
                            <button
                              onClick={() => {
                                setVerdicts((prev) => ({ ...prev, [key]: "incorrect" }));
                                send({
                                  event: "host_mark_answer",
                                  payload: {
                                    team_id: s.team_id,
                                    round_index: currentRound,
                                    question_index: q.question_index,
                                    verdict: "incorrect",
                                  },
                                });
                              }}
                            >
                              Wrong
                            </button>
                          </>
                        )}
                      </td>
                      <td>{runningTotals[s.team_id] ?? 0}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        ))}
        <button
          onClick={() => {
            send({ event: "host_publish_scores", payload: { round_index: currentRound } });
          }}
          disabled={!allMarked}
        >
          {allMarked ? "Publish Scores" : "Mark all answers to publish"}
        </button>
      </div>
    );
  };

  const renderRoundScores = () => {
    const isLastRound = !quizMeta || currentRound >= quizMeta.round_count - 1;
    const scoreEntries = teams
      .map((t) => ({ name: t.name, score: publishedScores[t.id] ?? 0 }))
      .sort((a, b) => b.score - a.score);

    return (
      <div>
        <h2>Round {currentRound + 1} — Scores</h2>
        {error && <p><strong>Error:</strong> {error}</p>}
        <table>
          <thead>
            <tr>
              <th>Team</th>
              <th>Round Score</th>
            </tr>
          </thead>
          <tbody>
            {scoreEntries.map((e) => (
              <tr key={e.name}>
                <td>{e.name}</td>
                <td>{e.score}</td>
              </tr>
            ))}
          </tbody>
        </table>
        <p>
          <button
            onClick={() => {
              setCeremonyIndex(0);
              setCeremonyAnswerRevealed(false);
              send({ event: "host_ceremony_show_question", payload: { question_index: 0 } });
              setPhase("ceremony");
            }}
          >
            Start Answer Ceremony
          </button>{" "}
          {!isLastRound && (
            <button
              onClick={() => {
                const nextRound = currentRound + 1;
                send({ event: "host_start_round", payload: { round_index: nextRound } });
              }}
            >
              Start Round {currentRound + 2}
            </button>
          )}{" "}
          {isLastRound && (
            <button
              onClick={() => {
                send({ event: "host_end_game", payload: {} });
              }}
            >
              End Game
            </button>
          )}
        </p>
      </div>
    );
  };

  const renderCeremony = () => {
    const isLastQuestion = ceremonyIndex >= roundQuestionCount - 1;

    return (
      <div>
        <h2>Answer Ceremony — Round {currentRound + 1}</h2>
        <p>
          Question {ceremonyIndex + 1} of {roundQuestionCount} on display
        </p>
        {error && <p><strong>Error:</strong> {error}</p>}
        {!ceremonyAnswerRevealed && (
          <button
            onClick={() => {
              send({ event: "host_ceremony_reveal_answer", payload: { question_index: ceremonyIndex } });
              setCeremonyAnswerRevealed(true);
            }}
          >
            Reveal Answer
          </button>
        )}{" "}
        {ceremonyAnswerRevealed && !isLastQuestion && (
          <button
            onClick={() => {
              const next = ceremonyIndex + 1;
              setCeremonyIndex(next);
              setCeremonyAnswerRevealed(false);
              send({ event: "host_ceremony_show_question", payload: { question_index: next } });
            }}
          >
            Next Question
          </button>
        )}
        {ceremonyAnswerRevealed && isLastQuestion && (
          <>
            {quizMeta && currentRound < quizMeta.round_count - 1 ? (
              <button
                onClick={() => {
                  const nextRound = currentRound + 1;
                  send({ event: "host_start_round", payload: { round_index: nextRound } });
                }}
              >
                Start Round {currentRound + 2}
              </button>
            ) : (
              <button
                onClick={() => {
                  send({ event: "host_end_game", payload: {} });
                }}
              >
                End Game
              </button>
            )}
          </>
        )}
      </div>
    );
  };

  const renderGameOver = () => {
    const scoreEntries = teams
      .map((t) => ({ name: t.name, score: finalScores[t.id] ?? 0 }))
      .sort((a, b) => b.score - a.score);

    return (
      <div>
        <h2>Game Over — Final Leaderboard</h2>
        <table>
          <thead>
            <tr>
              <th>Place</th>
              <th>Team</th>
              <th>Total Score</th>
            </tr>
          </thead>
          <tbody>
            {scoreEntries.map((e, i) => (
              <tr key={e.name}>
                <td>{i + 1}</td>
                <td>{e.name}</td>
                <td>{e.score}</td>
              </tr>
            ))}
          </tbody>
        </table>
        <p>Thanks for playing!</p>
      </div>
    );
  };

  const phaseContent = () => {
    switch (phase) {
      case "lobby":        return renderLobby();
      case "quiz_loaded":  return renderQuizLoaded();
      case "round_active": return renderRoundActive();
      case "round_ended":  return renderRoundEnded();
      case "scoring":      return renderScoring();
      case "round_scores": return renderRoundScores();
      case "ceremony":     return renderCeremony();
      case "game_over":    return renderGameOver();
    }
  };

  return (
    <div>
      <h1>Quizmaster Panel</h1>
      <p>
        Status:{" "}
        {connStatus === "connected" ? "Connected" : "Reconnecting..."}
        {quizMeta && ` — ${quizMeta.title}`}
      </p>
      {phaseContent()}
    </div>
  );
}
