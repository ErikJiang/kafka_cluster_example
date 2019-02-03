# :beetle: kafka_cluster_example

<p align="left">
  <img src="https://img.shields.io/badge/Go%20version-1.11-brightgreen.svg" />
  <img src="https://img.shields.io/badge/License-MIT-blue.svg" />
</p>

<p align="left">
  <img src="./wic.gif" />
</p>



> 该项目现处于施工状态

### 1. hosts 域名配置
使用 `ifconfig -a` 查看本地IP地址；

配置 /etc/hosts 文件，将域名 kfk1、kfk2、kfk3 与本地 IP 地址相关联；

``` ini
# 文件路径: /etc/hosts
# 假设IP: 192.168.0.66
192.168.0.66 kfk1 kfk2 kfk3
```

### 2. 构建 Kafka 集群
``` bash
$ docker-compose -f kafka/docker-compose.yml up -d
```

若构建下载过程等待连接超时，可尝试在 docker 的 `daemon.json` 中添加注册镜像：
``` json
{
    "registry-mirrors":["https://docker.mirrors.ustc.edu.cn"]
}
```

若构建完成，可使用 [kafkacat](https://github.com/edenhill/kafkacat) 检测服务是否正常运行：
``` bash
$ kafkacat -L -b kfk1:19092
```

### 3. 构建 Produce & Consume 服务

``` bash
# 生产者依赖添加
$ cd src/pruduce/
$ go mod vendor

# 消费者依赖添加
$ cd ../consume/
$ go mod vendor

# 返回根目录，构建服务
$ cd ../../
$ docker-compose -f src/docker-compose.yml up -d
```

若出现下载依赖 module 失败，或 build 缓慢，可尝试设置`GOPROXY`环境变量：
``` shell
$ export GOPROXY=https://goproxy.io
```

> todo 操作略繁琐，后期将尝试 makefile 来构建

### 3. 最终测试

这里采用 [ApacheBench](https://httpd.apache.org/docs/2.4/programs/ab.html) 来进行并发测试（并发数为10，总计100个请求）：
``` bash
$ ab -n100 -c10 -T application/json -p test/ab_post_test.json http://127.0.0.1:9000/api/v1/data
```

---

### 待完成

* 消费组并发测试
* 编写 makefile