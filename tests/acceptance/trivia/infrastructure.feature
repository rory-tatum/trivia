# Infrastructure -- Trivia Game
# Feature ID: trivia
# Release: 1 (Text-Only, Full Game Loop)
#
# Infrastructure scenarios validate the deployment configuration, CI/CD pipeline
# quality gates, and container startup behavior.
# These scenarios exercise real infrastructure behaviors: Docker build, env var
# validation, and fail-fast startup checks.
#
# All scenarios @skip until walking skeleton passes.

Feature: Infrastructure -- deployment and CI/CD validation

  # -----------------------------------------------------------------------
  # Container startup -- fail-fast behavior
  # DEC-020: HOST_TOKEN must be present at startup
  # -----------------------------------------------------------------------

  @skip
  Scenario: Server starts successfully when required environment variables are set
    Given HOST_TOKEN is set to "valid-token"
    And QUIZ_DIR is set to an accessible directory
    When the server process starts
    Then the server binds to port 8080 without error
    And the server is ready to accept connections

  @skip
  Scenario: Server refuses to start without HOST_TOKEN
    Given HOST_TOKEN is not set in the environment
    When the server process starts
    Then the server process exits immediately
    And the error output contains "HOST_TOKEN environment variable is required"
    And no port is bound

  @skip
  Scenario: Server refuses to start when QUIZ_DIR is not accessible
    Given HOST_TOKEN is set to "valid-token"
    And QUIZ_DIR is set to "/nonexistent/path"
    When the server process starts
    Then the server process exits immediately
    And the error output identifies the inaccessible path

  # -----------------------------------------------------------------------
  # Host token authentication
  # DEC-020: URL query token validated by auth guard middleware
  # -----------------------------------------------------------------------

  @skip
  Scenario: Requests to the quizmaster panel without a token receive a 403 response
    Given the server is running with HOST_TOKEN "test-secret-token"
    When an HTTP request is made to the quizmaster panel without a token parameter
    Then the response status is 403
    And the game state is not affected

  @skip
  Scenario: Requests to the quizmaster panel with an incorrect token receive a 403 response
    Given the server is running with HOST_TOKEN "test-secret-token"
    When an HTTP request is made to the quizmaster panel with token "wrong-token"
    Then the response status is 403

  @skip
  Scenario: Requests to the quizmaster panel with the correct token are accepted
    Given the server is running with HOST_TOKEN "test-secret-token"
    When an HTTP request is made to the quizmaster panel with token "test-secret-token"
    Then the response status is 200

  @skip
  Scenario: The player interface is accessible without any token
    Given the server is running
    When an HTTP request is made to the player interface URL
    Then the response status is 200

  @skip
  Scenario: The display interface is accessible without any token
    Given the server is running
    When an HTTP request is made to the display interface URL
    Then the response status is 200

  # -----------------------------------------------------------------------
  # Media file serving
  # DEC-007: Media files served from read-only bind mount
  # -----------------------------------------------------------------------

  @skip
  Scenario: Media files in the quiz directory are served under the media path
    Given the quiz directory contains "eiffel.jpg"
    When an HTTP request is made to the media path for "eiffel.jpg"
    Then the response status is 200
    And the response content type is "image/jpeg"

  @skip
  Scenario: Requesting a media file that does not exist returns a 404 response
    Given "no-such-file.jpg" does not exist in the quiz directory
    When an HTTP request is made to the media path for "no-such-file.jpg"
    Then the response status is 404

  # -----------------------------------------------------------------------
  # Frontend asset serving
  # DEC-022: React SPA served from embedded assets (go:embed)
  # -----------------------------------------------------------------------

  Scenario: The React application is served from the root path
    Given the server is running
    When an HTTP request is made to the root path
    Then the response contains the React application HTML
    And the response status is 200

  # -----------------------------------------------------------------------
  # Docker image build
  # Multi-stage build: node -> golang -> distroless
  # -----------------------------------------------------------------------

  @infrastructure
  Scenario: The Docker image builds without errors
    Given the project source code is present
    When the Docker image build is run
    Then the build completes without error
    And the resulting image uses the distroless runtime base

  @skip
  @infrastructure
  Scenario: The built Docker container starts and serves the application
    Given the Docker image has been built
    And HOST_TOKEN is set in the environment
    When the container is started with docker-compose
    Then the container becomes healthy within 30 seconds
    And the player interface responds to HTTP requests on the configured port

  # -----------------------------------------------------------------------
  # Architecture boundary enforcement
  # DEC-031: go-arch-lint as hard CI gate
  # -----------------------------------------------------------------------

  @skip
  @infrastructure
  Scenario: Package dependency rules pass go-arch-lint validation
    Given the project Go source code is present
    When go-arch-lint check is run against the project
    Then no architecture violations are reported
    And specifically the handler package has no reference to QuestionFull
    And specifically the hub package has no reference to QuestionFull

  # -----------------------------------------------------------------------
  # CI pipeline quality gates
  # -----------------------------------------------------------------------

  @skip
  @infrastructure
  Scenario: TypeScript type checking passes with zero errors
    Given the frontend source code is present
    When TypeScript type checking is run with strict mode
    Then zero type errors are reported

  @skip
  @infrastructure
  Scenario: Go tests pass with the race detector enabled
    Given the project Go source code is present
    When go test is run with the race detector flag
    Then all tests pass
    And no race conditions are detected
