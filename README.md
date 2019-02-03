# :beetle: kafka_cluster_example

<p align="left">
  <img src="https://img.shields.io/badge/Go%20version-1.11-brightgreen.svg" />
  <img src="https://img.shields.io/badge/License-MIT-blue.svg" />
</p>

![Valar Morghulis](./ValarMorghulis.gif)

---

### 本项目依赖工具：
* [Make](https://www.gnu.org/software/make/)
* [Kafkacat](https://github.com/edenhill/kafkacat)
* [ApacheBench](https://httpd.apache.org/docs/2.4/programs/ab.html)

### 1. 配置 hosts 域名
使用 `ifconfig -a` 查看本地IP地址；

配置 /etc/hosts 文件，将域名 kfk1、kfk2、kfk3 映射到当前本地 IP 地址，例如；

``` bash
# 假设本地IP为: 192.168.0.166
192.168.0.166 kfk1 kfk2 kfk3
```

### 2. 构建 Kafka 集群

``` bash
$ docker-compose -f kafka/docker-compose.yml up -d
```
或者使用：
``` bash
$ make kafka-up
```

如若在构建下载过程中，出现等待连接超时，可尝试在 docker 的 `daemon.json` 中添加注册镜像：
``` json
{
    "registry-mirrors":["https://docker.mirrors.ustc.edu.cn"]
}
```

若构建完成，可使用 kafkacat 检测服务是否正常运行：
``` bash
$ kafkacat -L -b kfk1:19092
```
或使用：
``` bash
$ make kafka-test
```

### 3. 构建 Produce & Consume 服务

为 produce 和 consume 生成 vendor 依赖：
``` bash
$ make vendor
```
(todo 内置makefile)若出现下载依赖 module 失败，或 build 缓慢，可尝试设置`GOPROXY`环境变量：
``` shell
$ export GOPROXY=https://goproxy.io
```

构建 Produce & Consume Docker 服务
``` bash
$ docker-compose -f src/docker-compose.yml up -d
```
或者使用：
``` bash
$ make up
```

### 4. 最终测试

使用 ApacheBench 进行并发测试（并发数为10，总计100个请求）：
``` bash
$ ab -n100 -c10 -T application/json -p test/ab_post_test.json http://127.0.0.1:9000/api/v1/data
```
或者使用：
``` bash
$ make test
```

---