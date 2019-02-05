include .env

PROJECTNAME=$(shell basename "$(PWD)")

all: help

## vendor			Auto generate go vendor dir.
.PHONY: vendor
vendor:
	@echo "auto generate vendor dir ...";
	export GOPROXY=$(GOPROXY); cd src/produce/; go mod vendor;
	export GOPROXY=$(GOPROXY); cd src/consume/; go mod vendor;

## up			Docker compose up for src.
.PHONY: up
up:
	@echo "docker compose up ...";
	docker-compose -f src/docker-compose.yml up -d;

## down			Docker compose down for src.
.PHONY: down
down:
	@echo "docker compose down ...";
	docker-compose -f src/docker-compose.yml down;

## ps			Docker compose ps for src.
.PHONY: ps
ps:
	@echo "docker compose ps ...";
	docker-compose -f src/docker-compose.yml ps;

## logs			Docker compose logs for src.
.PHONY: logs
logs:
	@echo "docker compose logs ...";
	docker-compose -f src/docker-compose.yml logs -f;

## clean			Clean up docker images for src.
.PHONY: clean
clean:
	@echo "docker image clean ...";
	docker image prune -f;
	docker rmi src_produce src_consume1 src_consume2;

## test			Apache benchmark test for src.
.PHONY: test
test:
	@echo "apache benchmark test ...";
	ab -n100 -c10 -T application/json -p test/ab_post_test.json http://127.0.0.1:9000/api/v1/data;

## kafka-up		Docker compose up for kafka services.
.PHONY: kafka-up
kafka-up:
	@echo "docker compose up for kafka ...";
	docker-compose -f kafka/docker-compose.yml up -d;

## kafka-down		Docker compose down for kafka services.
.PHONY: kafka-down
kafka-down:
	@echo "docker compose down for kafka ...";
	docker-compose -f kafka/docker-compose.yml down;

## kafka-clean		Clean up log and data files for kafka services.
.PHONY: kafka-clean
kafka-clean:
	@echo "kafka dir clean ...";
	cd kafka/kfk1/data/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/kfk2/data/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/kfk3/data/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/zk1/data/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/zk1/log/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/zk2/data/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/zk2/log/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/zk3/data/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/zk3/log/; ls|grep -v .gitkeep|xargs rm -rf;

## kafka-test		Check running state of the kafka service.
.PHONY: kafka-test
kafka-test:
	@echo "check kafka run status ...";
	kafkacat -L -b $(KFKADDR);

## help			print this help message and exit.
.PHONY: help
help: Makefile
	@echo ""
	@echo "Choose a command run in "$(PROJECTNAME)":"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Valid target values are:"
	@echo ""
	@sed -n 's/^## //p' $<