build:
	docker compose up -d app

bash:
	docker compose run --rm app bash

test:
	docker compose run --rm app ./scripts/test.sh

clean:
	docker compose kill
	docker compose rm -f
