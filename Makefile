# DOCKER TASKS
# Build the container
build: ## Build the container
	docker build . -t stipendivm

run: ## Run container on port configured in `config.env`
	docker run -p 4567:4567 stipendivm

stop: ## Stop and remove a running container
	docker stop stipendivm; docker rm stipendivm
