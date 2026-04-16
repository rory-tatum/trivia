# Stage 1: Build the React frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
# vite.config.ts writes output to ../internal/static/dist
# but we redirect outDir here by building in place and copying result
RUN npm run build

# Stage 2: Build the Go binary
FROM golang:1.23-alpine AS go-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy all Go source code
COPY . .

# Copy the compiled frontend assets into the embed path
# vite outputs to internal/static/dist (relative to project root)
COPY --from=frontend-builder /app/frontend/../internal/static/dist /app/internal/static/dist

# Build statically linked binary (required for scratch/distroless)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /server ./cmd/server

# Stage 3: Minimal runtime image
FROM gcr.io/distroless/static:nonroot

COPY --from=go-builder /server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
