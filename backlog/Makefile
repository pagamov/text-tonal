# Variables
IMAGE_NAME = docker.io/library/postgres:latest
CONTAINER_NAME = database
DB_NAME = main
DB_USER = pagamov
DB_PASSWORD = multipass
HOST_PORT = 5432
DATA_DIR = ./data

# Check if Podman is installed
check-podman:
	@command -v podman >/dev/null 2>&1 || { \
		echo "Podman is not installed. Installing Podman..."; \
		$(MAKE) install-podman; \
	}

# Check Podman permissions
check-podman-permissions:
	@podman info >/dev/null 2>&1 || { \
		echo "You do not have permission to access Podman."; \
		echo "Please ensure you have the necessary permissions."; \
		exit 1; \
	}

# Install Podman (for Ubuntu and macOS)
install-podman:
	@if [ "$$(uname)" = "Linux" ]; then \
		echo "Installing Podman on Linux..."; \
		sudo apt-get update; \
		sudo apt-get install -y podman; \
	elif [ "$$(uname)" = "Darwin" ]; then \
		echo "Installing Podman on macOS..."; \
		brew install podman; \
	else \
		echo "Unsupported OS. Please install Podman manually."; \
		exit 1; \
	fi

# Check if PostgreSQL is installed
check-postgresql:
	@command -v psql >/dev/null 2>&1 || { \
		echo "PostgreSQL is not installed. Installing PostgreSQL..."; \
		$(MAKE) install-postgresql; \
	}

# Install PostgreSQL (for Ubuntu and macOS)
install-postgresql:
	@if [ "$$(uname)" = "Linux" ]; then \
		echo "Installing PostgreSQL on Linux..."; \
		sudo apt-get update; \
		sudo apt-get install -y postgresql postgresql-contrib; \
	elif [ "$$(uname)" = "Darwin" ]; then \
		echo "Installing PostgreSQL on macOS..."; \
		brew install postgresql; \
	else \
		echo "Unsupported OS. Please install PostgreSQL manually."; \
		exit 1; \
	fi

# Default target
.PHONY: all
all: check-podman check-podman-permissions check-postgresql start

# Start the PostgreSQL container
.PHONY: start
start: check-podman check-podman-permissions create-data-dir
	podman run  --name $(CONTAINER_NAME) -e POSTGRES_DB=$(DB_NAME) -e POSTGRES_USER=$(DB_USER) -e POSTGRES_PASSWORD=$(DB_PASSWORD) -p $(HOST_PORT):5432 -v $(DATA_DIR):/var/lib/postgresql/data -d $(IMAGE_NAME)
	

.PHONY: check
check:
	psql -h localhost -p 5432 -U $(DB_USER) -d $(DB_NAME)

# Create the data directory if it doesn't exist
.PHONY: create-data-dir
create-data-dir:
	@mkdir -p $(DATA_DIR)

# Stop the PostgreSQL container
.PHONY: stop
stop:
	podman stop $(CONTAINER_NAME)

# Remove the PostgreSQL container
.PHONY: rm
rm:
	podman rm $(CONTAINER_NAME)

# Remove the PostgreSQL image
.PHONY: rmi
rmi:
	podman rmi $(IMAGE_NAME)

# Clean up data directory
.PHONY: clean
clean:
	rm -rf $(DATA_DIR)/*

# Show logs from the PostgreSQL container
.PHONY: logs
logs:
	podman logs $(CONTAINER_NAME)

# Execute a command in the PostgreSQL container
.PHONY: exec
exec:
	podman exec -it $(CONTAINER_NAME) psql -U $(DB_USER) -d $(DB_NAME)

# Help message
.PHONY: help
help:
	@echo "Makefile for managing PostgreSQL Podman container"
	@echo "Usage:"
	@echo "  make start    - Start the PostgreSQL container"
	@echo "  make stop     - Stop the PostgreSQL container"
	@echo "  make rm       - Remove the PostgreSQL container"
	@echo "  make rmi      - Remove the PostgreSQL image"
	@echo "  make clean    - Clean up data directory"
	@echo "  make logs     - Show logs from the PostgreSQL container"
	@echo "  make exec     - Execute a command in the PostgreSQL container"
	@echo "  make help     - Show this help message"