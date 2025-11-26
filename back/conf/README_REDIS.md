# Redis 持久化配置说明

## 配置文件

- `redis.conf`: Redis 服务器配置文件，已启用 AOF 和 RDB 持久化

## 持久化配置详情

### 1. RDB 持久化（快照）
- **触发条件**：
  - 900秒内至少1个key变化
  - 300秒内至少10个key变化
  - 60秒内至少10000个key变化
- **文件位置**: `/data/dump.rdb`
- **优点**: 文件小，恢复快
- **缺点**: 可能丢失最后一次快照后的数据

### 2. AOF 持久化（追加文件）⭐ 推荐
- **已启用**: `appendonly yes`
- **同步策略**: `everysec`（每秒同步一次，平衡性能和安全）
- **文件位置**: `/data/appendonly.aof`
- **优点**: 数据更安全，最多丢失1秒数据
- **缺点**: 文件较大，恢复较慢

## Docker 部署

### 启动服务
```bash
cd back
docker-compose up -d
```

### 检查 Redis 状态
```bash
# 进入 Redis 容器
docker exec -it chainqa_redis redis-cli

# 检查持久化配置
CONFIG GET appendonly
CONFIG GET save
CONFIG GET dir

# 检查数据
INFO persistence
```

### 查看持久化文件
```bash
# 查看数据卷
docker volume inspect back_redis_data

# 进入容器查看文件
docker exec -it chainqa_redis ls -lh /data
```

## 数据恢复

### 从 RDB 恢复
1. 停止 Redis
2. 将 `dump.rdb` 文件复制到 `/data` 目录
3. 启动 Redis（会自动加载 RDB 文件）

### 从 AOF 恢复
1. 停止 Redis
2. 将 `appendonly.aof` 文件复制到 `/data` 目录
3. 启动 Redis（会自动加载 AOF 文件）

## 备份建议

### 手动备份
```bash
# 备份 RDB 文件
docker exec chainqa_redis redis-cli BGSAVE
docker cp chainqa_redis:/data/dump.rdb ./backup/dump_$(date +%Y%m%d_%H%M%S).rdb

# 备份 AOF 文件
docker cp chainqa_redis:/data/appendonly.aof ./backup/appendonly_$(date +%Y%m%d_%H%M%S).aof
```

### 自动备份脚本
建议设置定时任务（cron）定期备份 Redis 数据文件。

## 性能优化

### 如果数据量很大
1. 调整 `auto-aof-rewrite-percentage` 和 `auto-aof-rewrite-min-size`
2. 考虑使用 Redis Cluster 进行分片
3. 定期清理过期数据

### 监控建议
- 监控 AOF 文件大小
- 监控 RDB 保存频率
- 监控 Redis 内存使用情况

## 故障排查

### 检查持久化是否工作
```bash
# 查看持久化信息
docker exec chainqa_redis redis-cli INFO persistence

# 查看最后保存时间
docker exec chainqa_redis redis-cli LASTSAVE
```

### 如果数据丢失
1. 检查 `/data` 目录权限
2. 检查磁盘空间
3. 查看 Redis 日志: `docker logs chainqa_redis`

