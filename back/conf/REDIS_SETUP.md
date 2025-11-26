# Redis 持久化配置使用指南

## 📋 配置概览

已为项目配置了完整的 Redis 持久化方案，包括：

1. ✅ **RDB 持久化**（快照备份）
2. ✅ **AOF 持久化**（追加文件，推荐）
3. ✅ **Docker 数据卷**（数据持久化到宿主机）
4. ✅ **健康检查**（自动重启）

## 🚀 快速开始

### 1. 启动服务

```bash
cd back
docker-compose up -d
```

### 2. 验证 Redis 持久化

```bash
# 检查 Redis 是否运行
docker ps | grep redis

# 进入 Redis 容器检查配置
docker exec -it chainqa_redis redis-cli

# 在 Redis CLI 中执行：
CONFIG GET appendonly    # 应该返回 "yes"
CONFIG GET save          # 应该返回 RDB 配置
INFO persistence        # 查看持久化详细信息
```

### 3. 测试数据持久化

```bash
# 写入测试数据
docker exec -it chainqa_redis redis-cli SET test_key "test_value"

# 重启 Redis 容器
docker restart chainqa_redis

# 检查数据是否还在
docker exec -it chainqa_redis redis-cli GET test_key
# 应该返回 "test_value"
```

## 📁 文件说明

### 配置文件

- `conf/redis.conf`: Redis 服务器配置文件
  - 启用 AOF 持久化（`appendonly yes`）
  - 配置 RDB 快照（多个时间点）
  - 数据目录：`/data`

- `conf/config.ini`: 应用配置文件
  - Redis 连接配置（host, port, password 等）

### Docker 配置

- `docker-compose.yml`: 
  - Redis 服务定义
  - 数据卷挂载（`redis_data`）
  - 健康检查配置

## 🔧 配置说明

### Redis 持久化策略

#### AOF（追加文件）- 主要持久化方式
- **状态**: ✅ 已启用
- **同步策略**: `everysec`（每秒同步）
- **优点**: 数据安全，最多丢失1秒数据
- **文件**: `/data/appendonly.aof`

#### RDB（快照）- 辅助备份
- **状态**: ✅ 已启用
- **触发条件**:
  - 900秒内至少1个key变化
  - 300秒内至少10个key变化
  - 60秒内至少10000个key变化
- **文件**: `/data/dump.rdb`

### 数据卷

数据存储在 Docker 数据卷 `redis_data` 中，即使容器删除，数据也会保留。

查看数据卷位置：
```bash
docker volume inspect back_redis_data
```

## 💻 在代码中使用

### 初始化 Redis 客户端

```go
import (
    "context"
    "chainqa_offchain_demo/indexer"
    "chainqa_offchain_demo/setting"
)

func main() {
    // 加载配置
    setting.Init("./conf/config.ini")
    
    ctx := context.Background()
    
    // 初始化 Redis 客户端（会自动检查持久化配置）
    redisClient, err := indexer.InitRedisClient(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // 使用 Redis 客户端...
}
```

### 创建索引服务

```go
// 假设已有 chainClient
indexerService, err := indexer.NewIndexerService(chainClient, ctx)
if err != nil {
    log.Fatal(err)
}

// 启动区块监听
go indexerService.StartBlockListener()
```

## 📊 监控和维护

### 检查持久化状态

```bash
# 查看持久化信息
docker exec chainqa_redis redis-cli INFO persistence

# 查看最后保存时间
docker exec chainqa_redis redis-cli LASTSAVE

# 查看 AOF 文件大小
docker exec chainqa_redis ls -lh /data/appendonly.aof

# 查看 RDB 文件
docker exec chainqa_redis ls -lh /data/dump.rdb
```

### 手动触发保存

```bash
# 手动触发 RDB 保存
docker exec chainqa_redis redis-cli BGSAVE

# 查看保存进度
docker exec chainqa_redis redis-cli INFO persistence | grep rdb_bgsave_in_progress
```

### 备份数据

```bash
# 创建备份目录
mkdir -p ./backup/redis

# 备份 RDB 文件
docker cp chainqa_redis:/data/dump.rdb ./backup/redis/dump_$(date +%Y%m%d_%H%M%S).rdb

# 备份 AOF 文件
docker cp chainqa_redis:/data/appendonly.aof ./backup/redis/appendonly_$(date +%Y%m%d_%H%M%S).aof
```

## 🔒 安全建议

### 生产环境配置

1. **设置密码**（推荐）:
   ```ini
   # 在 config.ini 中
   [redis]
   password = your_strong_password
   ```

   然后在 `redis.conf` 中启用：
   ```conf
   requirepass your_strong_password
   ```

2. **限制网络访问**:
   - 如果不需要外部访问，可以移除 `docker-compose.yml` 中的端口映射
   - 只在 Docker 网络内部访问

3. **定期备份**:
   - 设置定时任务定期备份数据文件
   - 建议每天至少备份一次

## 🐛 故障排查

### 数据丢失问题

1. **检查持久化配置**:
   ```bash
   docker exec chainqa_redis redis-cli CONFIG GET appendonly
   docker exec chainqa_redis redis-cli CONFIG GET save
   ```

2. **检查磁盘空间**:
   ```bash
   docker exec chainqa_redis df -h /data
   ```

3. **查看日志**:
   ```bash
   docker logs chainqa_redis
   ```

### 性能问题

如果 Redis 性能下降：

1. **检查 AOF 文件大小**:
   ```bash
   docker exec chainqa_redis redis-cli INFO persistence
   ```

2. **手动触发 AOF 重写**:
   ```bash
   docker exec chainqa_redis redis-cli BGREWRITEAOF
   ```

3. **调整配置**:
   - 修改 `redis.conf` 中的 `auto-aof-rewrite-percentage` 和 `auto-aof-rewrite-min-size`

## 📝 注意事项

1. **数据目录权限**: 确保 Redis 容器有权限写入 `/data` 目录
2. **磁盘空间**: 定期检查磁盘空间，AOF 文件可能较大
3. **备份策略**: 建议同时保留 RDB 和 AOF 备份
4. **测试恢复**: 定期测试数据恢复流程，确保备份有效

## 🔗 相关文档

- [Redis 持久化官方文档](https://redis.io/docs/management/persistence/)
- [Docker Compose 文档](https://docs.docker.com/compose/)
- 项目内文档: `conf/README_REDIS.md`

