# Trivia Game

An app that a quizmaster can run that will load in a yaml file containing the rounds and questions for each round and allow multiple players/teams to connect and answer those questions. Designed for having fun with friends.

Quizmaster and players have different interfaces. Players can only see rounds and questions, while quizmasters have a private interface with controls to start the game, advance the questions/rounds, do scoring, etc.

At the beginning of a game, players will connect to the website and create a team with a team name. Whenever that player refreshes the page or leaves and comes back, the game should recognize their device and automatically know their team.

Rounds are played one at a time, and questions are revealed one at a time. Players connected on their devices should be able to see all the questions that have been revealed for the current round and enter/edit answers for any of them. The quizmaster will be able to have a page that they can share for everyone to see that only shows the current question. At the end of a round, once all questions are revealed, players can click a submit button to send their answers to the quizmaster. And then the quizmaster should be able to use that same page they're sharing to go over the answers to each question one at a time.

Once all teams/players have submitted, the quizmaster should have a scoring interface that shows what the answer was supposed to be, and then the submitted answers, so they can decide if the given answers were close enough and mark them right or wrong. Scores for each player/team should be tallied at the end of each round and visible to the quizmaster.

Many of the questions will just be text, but there should be an ability to have picture, video, or audio files for questions as well. Answers will always be plain text. Some questions are also multi part and can have multiple answers, sometimes the order of those answers matters, and sometimes not. There should also be the ability to have multiple choice questions.