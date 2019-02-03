.PHONY: vendor up down ps logs clean test kafka-up kafka-down kafka-clean kafka-test

all: clean vendor up test

vendor: ;@echo "auto generate vendor dir ...";
	cd src/produce/; go mod vendor;
	cd src/consume/; go mod vendor;

up: ;@echo "docker compose up ...";
	docker-compose -f src/docker-compose.yml up -d;

down: ;@echo "docker compose down ...";
	docker-compose -f src/docker-compose.yml down;

ps: ;@echo "docker compose ps ...";
	docker-compose -f src/docker-compose.yml ps;

logs: ;@echo "docker compose logs ...";
	docker-compose -f src/docker-compose.yml logs -f;

clean: ;@echo "docker image clean ...";
	docker image prune -f;
	docker rmi src_produce src_consume1 src_consume2;

test: ;@echo "apache benchmark test ...";
	ab -n100 -c10 -T application/json -p test/ab_post_test.json http://127.0.0.1:9000/api/v1/data;

kafka-up: ;@echo "docker compose up for kafka ...";
	docker-compose -f kafka/docker-compose.yml up -d;

kafka-down: ;@echo "docker compose down for kafka ...";
	docker-compose -f kafka/docker-compose.yml down;

kafka-clean: ;@echo "kafka dir clean ...";
	cd kafka/kfk1/data/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/kfk2/data/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/kfk3/data/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/zk1/data/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/zk1/log/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/zk2/data/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/zk2/log/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/zk3/data/; ls|grep -v .gitkeep|xargs rm -rf;
	cd kafka/zk3/log/; ls|grep -v .gitkeep|xargs rm -rf;

kafka-test: ;@echo "check kafka run status ...";
	kafkacat -L -b kfk1:19092;
