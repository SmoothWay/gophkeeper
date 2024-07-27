.PHONY: all server auth client start stop

PID_DIR := .pids

all: start

start: | $(PID_DIR) server auth client

$(PID_DIR):
	@mkdir -p $(PID_DIR)

server:
	@echo "Starting server..."
	@go run ./cmd/server > server.log 2>&1 &
	@echo $$! > $(PID_DIR)/server.pid

auth:
	@echo "Starting auth service..."
	@go run ./cmd/auth > auth.log 2>&1 &
	@echo $$! > $(PID_DIR)/auth.pid

client:
	@echo "Starting client..."
	@go run ./cmd/client
	@echo $$! > $(PID_DIR)/client.pid

stop:
	@echo "Stopping server..."
	@kill `cat $(PID_DIR)/server.pid` || true
	@rm -f $(PID_DIR)/server.pid

	@echo "Stopping auth service..."
	@kill `cat $(PID_DIR)/auth.pid` || true
	@rm -f $(PID_DIR)/auth.pid

	@echo "Stopping client..."
	@kill `cat $(PID_DIR)/client.pid` || true
	@rm -f $(PID_DIR)/client.pid

clean:
	@rm -rf $(PID_DIR)
	@rm -f server.log auth.log client.log
