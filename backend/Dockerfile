FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -o /server ./cmd/server && CGO_ENABLED=0 go build -trimpath -o /setup ./cmd/setup
FROM python:3.12-slim-bookworm
RUN groupadd --gid 10001 app && useradd --uid 10001 --gid app --create-home app
WORKDIR /app
COPY tools/requirements.txt /app/tools/requirements.txt
RUN pip install --no-cache-dir -r /app/tools/requirements.txt && mkdir -p /data/uploads && chown -R app:app /data /app
COPY --from=build /server /setup /app/
COPY tools/ /app/tools/
USER app
EXPOSE 8081
CMD ["/app/server"]
