build:
	docker-compose up -d app

bash:
	docker-compose run --rm app bash

test:
	docker-compose run --rm app ./codeship/test.sh

clean:
	docker-compose kill
	docker-compose rm -f
