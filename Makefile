APP_NAME=speedtest
PID_FILE=/tmp/$(APP_NAME).pid
LOG_FILE=/tmp/$(APP_NAME).log

.PHONY: build run stop restart status logs

build:
	go build -o $(APP_NAME) .

run: build
	@if [ -f $(PID_FILE) ]; then \
		PID=$$(cat $(PID_FILE)); \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "$(APP_NAME) já está rodando com PID $$PID"; \
			exit 1; \
		else \
			echo "Removendo PID file antigo"; \
			rm -f $(PID_FILE); \
		fi \
	fi

	@echo "Iniciando $(APP_NAME)..."
	@nohup ./$(APP_NAME) >> $(LOG_FILE) 2>&1 & echo $$! > $(PID_FILE)

	@sleep 1

	@PID=$$(cat $(PID_FILE)); \
	if ps -p $$PID > /dev/null 2>&1; then \
		echo "$(APP_NAME) iniciado com PID $$PID"; \
		echo "Logs: tail -f $(LOG_FILE)"; \
	else \
		echo "Falha ao iniciar $(APP_NAME)"; \
		rm -f $(PID_FILE); \
		exit 1; \
	fi

stop:
	@if [ ! -f $(PID_FILE) ]; then \
		echo "$(APP_NAME) não está rodando"; \
		exit 1; \
	fi

	@PID=$$(cat $(PID_FILE)); \
	if ps -p $$PID > /dev/null 2>&1; then \
		echo "Parando $(APP_NAME) PID $$PID"; \
		kill $$PID; \
		rm -f $(PID_FILE); \
		echo "Parado"; \
	else \
		echo "Processo não existe"; \
		rm -f $(PID_FILE); \
	fi

restart: stop
	@sleep 1
	@$(MAKE) run

status:
	@if [ -f $(PID_FILE) ]; then \
		PID=$$(cat $(PID_FILE)); \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "$(APP_NAME) está rodando com PID $$PID"; \
		else \
			echo "$(APP_NAME) não está rodando (PID file órfão)"; \
		fi \
	else \
		echo "$(APP_NAME) não está rodando"; \
	fi

logs:
	tail -f $(LOG_FILE)
