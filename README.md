# kafka_tutorial
kafka tutorial

### 构建 Kafka 集群
``` bash
 $ docker-compose up -d
```

若构建下载过程等待连接超时，可尝试在 docker 的 `daemon.json` 中添加注册镜像：
``` json
{
    "registry-mirrors":["https://docker.mirrors.ustc.edu.cn"]
}
```
