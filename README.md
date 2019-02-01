# kafka_tutorial

<p align="left">
  <img src="https://img.shields.io/badge/Go%20version-1.11-brightgreen.svg" />
  <img src="https://img.shields.io/badge/License-MIT-blue.svg" />
</p>

### 构建 Kafka 集群
``` bash
$ docker-compose -f kafka/docker-compose.yml up -d
```

若构建下载过程等待连接超时，可尝试在 docker 的 `daemon.json` 中添加注册镜像：
``` json
{
    "registry-mirrors":["https://docker.mirrors.ustc.edu.cn"]
}
```

### hosts 域名配置

配置 /etc/hosts 文件，将域名 kfk1、kfk2、kfk3 与本地 IP 地址相关联；


### 待完成

* 消费组并发测试
* 编写 makefile