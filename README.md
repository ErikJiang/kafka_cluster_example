# :beetle: kafka_cluster_example

<p align="center">
  <img src="./gopher.png" />
</p>
<p align="center">
  <img src="https://img.shields.io/badge/Go%20version-1.11-brightgreen.svg" />
  <img src="https://img.shields.io/badge/License-MIT-blue.svg" />
</p>

---

### 项目依赖工具：
* [Make](https://www.gnu.org/software/make/)
* [Kafkacat](https://github.com/edenhill/kafkacat)
* [ApacheBench](https://httpd.apache.org/docs/2.4/programs/ab.html)

### 支持 make 构建：
```
$ make

Choose a command run in kafka_cluster_example:

Usage: make [target]

Valid target values are:

vendor                  Auto generate go vendor dir.
up                      Docker compose up for src.
down                    Docker compose down for src.
ps                      Docker compose ps for src.
logs                    Docker compose logs for src.
clean                   Clean up docker images for src.
test                    Apache benchmark test for src.
kafka-up                Docker compose up for kafka services.
kafka-down              Docker compose down for kafka services.
kafka-clean             Clean up log and data files for kafka services.
kafka-test              Check running state of the kafka service.
help                    print this help message and exit.
```

---

### 1. 配置 hosts 域名
使用 `ifconfig -a` 查看本地IP地址；

配置 /etc/hosts 文件，将域名 kfk1、kfk2、kfk3 映射到当前本地 IP 地址，例如；

``` sh
# 假设本地IP为: 192.168.0.166
192.168.0.166 kfk1 kfk2 kfk3
```

### 2. 构建 Kafka 集群

``` sh
# docker compose 构建方式：
$ docker-compose -f kafka/docker-compose.yml up -d

# 或使用 make 构建方式：
$ make kafka-up
```

如若在构建下载过程中，出现等待连接超时，可尝试在 docker 的 `daemon.json` 中添加注册镜像：
``` json
{
    "registry-mirrors":["https://docker.mirrors.ustc.edu.cn"]
}
```

若构建完成，可使用 kafkacat 检测服务是否正常运行：
``` sh
# 直接进行检测验证：
$ kafkacat -L -b kfk1:19092

# 或使用 make 方式验证：
$ make kafka-test
```

### 3. 构建 Produce & Consume 服务

为 produce 和 consume 生成 vendor 依赖：
``` sh
$ make vendor
```

构建 Produce & Consume Docker 服务
``` sh
# 使用 docker compose 直接构建方式：
$ docker-compose -f src/docker-compose.yml up -d

# 或者使用 make 构建方式：
$ make up
```

### 4. 最终测试

使用 ApacheBench 进行并发测试（并发数为10，总计100个请求）：
``` sh
# 直接使用 ab 命令进行测试：
$ ab -n100 -c10 -T application/json -p test/ab_post_test.json http://127.0.0.1:9000/api/v1/data

# 或者使用 make 方式测试：
$ make test
```

---