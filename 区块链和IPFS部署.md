# 区块链和IPFS部署准备

## 部署区块链

- 请参考[本篇文档](https://docs.chainmaker.org.cn/v2.3.7/html/dev/%E9%95%BF%E5%AE%89%E9%93%BE%E7%AE%A1%E7%90%86%E5%8F%B0.html#id6)完成长安链部署和长安链管理平台部署
- **【重要】** 部署好后，请下载用户证书密钥，下载好后，放入 `/back/chain/chainConfig/config/user`下面。
- 【重要】随后，请修改 `/back/chain/chainConfig/sdk_config.yml`配置，如 `node_addr`等信息。相关信息可以从长安链的管理平台查询。

## 部署IPFS

- 新建一个文件夹，并进入该文件夹
- 编写 shell 脚本：请注意下面要修改的部分。

```bash
#!/bin/bash

# ====以下内容可能需要修改=========
# 设置跨域的命令
CORS_ORIGIN='["http://47.113.204.64:5001", "http://localhost:3000", "http://127.0.0.1:5001", "https://webui.ipfs.io"]'
# 【tip1】若为服务器，请修改第一个IP为服务器公网IP。否则删掉第一个！

CORS_METHODS='["PUT", "POST"]'


# 定义目录变量
base_dir="./ipfsDataAndExport"
# 【tip2】请以本脚本的位置为定位，按相对位置创建上面ipfsDataAndExport文件夹，下面文件夹不需要创建，会自动化创建


# ====以上内容可能需要修改=========

ipfs_data_dir="$base_dir/ipfs_data"
ipfs_export_dir="$base_dir/ipfs_export"

# 容器名
CONTAINER_NAME=ipfs_host

# 检查并删除现有的文件夹
if [ -d "$ipfs_data_dir" ]; then
  echo "删除旧ipfs_data"
  rm -rf "$ipfs_data_dir"
fi

if [ -d "$ipfs_export_dir" ]; then
  echo "删除旧ipfs_export"
  rm -rf "$ipfs_export_dir"
fi

# 等待容器启动
echo "等待删除程序，等待3s..."
sleep 3

# if [ $# -eq 0 ]; then
#   echo "未传入IPFS Docker连接网络名称参数。本脚本要求传入IPFS Docker连接网络名称参数"
#   echo "提示：因为需要保证区块链背书节点也能访问到ipfs，所以需要ipfs和区块链使用同一个网络，所以需要设置网络名称。建议先启动区块链。确认容器的网络配置是否正确!"
#   exit 1
# fi


#---------------------------------------------------------------





# 创建新的文件夹
echo "(1) 创建新ipfs_export和ipfs_data"
mkdir -p "$ipfs_data_dir"
mkdir -p "$ipfs_export_dir"

# 输出结果
if [ -d "$ipfs_data_dir" ] && [ -d "$ipfs_export_dir" ]; then
  echo -e "\033[0;32m 创建新ipfs_export和ipfs_data成功 \033[0m"
else
  echo "Failed to create folders."
  exit 1
fi



echo "(2) 启动ipfs_host容器..."
# export ipfs_host_NETWORK_NAME=$1
export ipfs_host_CONTAINER_NAME=$CONTAINER_NAME
docker-compose up -d

# 等待容器启动
echo "等待容器自行启动，等待10s..."
sleep 10

echo ">> 可以使用docker logs -f ipfs_host 自行查看容器是否启动成功"
echo ">> 若末尾出现Daemon is ready 说明启动成功"

# # 定义要查找的字符串
# STRING_TO_FIND="Daemon is ready"
# # 使用 docker logs -f 和 awk 来等待特定的日志消息
# docker logs -f ipfs_host | awk '/'"$STRING_TO_FIND"'/{print; exit}'

# # 检查 awk 的退出状态
# if [ $? -eq 0 ]; then
#   echo -e "\033[0;32m 容器启动成功 \033[0m"
#   # 在这里添加后续命令
# else
#   echo "错误：没有找到消息 '$STRING_TO_FIND' ，容器启动失败。"
#   exit 1
# fi


echo "(3) 开始设置跨域配置..."

# 进入容器内部并设置跨域
docker exec -it $CONTAINER_NAME sh -c "
    ipfs config --json API.HTTPHeaders.Access-Control-Allow-Origin '$CORS_ORIGIN' &&
    ipfs config --json API.HTTPHeaders.Access-Control-Allow-Methods '$CORS_METHODS'
  "

# 检查设置跨域命令是否成功
  if [ $? -eq 0 ]; then
  
    echo -e "\033[0;32m 跨域配置设置成功， \033[0m"
    echo "(4) 重启容器..."
    # 重启容器
    docker restart $CONTAINER_NAME
    echo -e "\033[0;32m 容器配置完成，可以访问 公网IP:5001/webui 或者 localhost:5001/webui 查看webUI界面 \033[0m"
  else
    echo "跨域配置设置失败。"
    exit 1
  fi
```

- 同一目录下创建 `docker-compose.yml` 文件，内容如下：

```yaml
version: '2'

services:
  ipfs_host:
    container_name: "${ipfs_host_CONTAINER_NAME}"
    image: ipfs/kubo:latest
    volumes:
      - ../ipfsDataAndExport/ipfs_export:/export
      - ../ipfsDataAndExport/ipfs_data:/data/ipfs
    #command: /bin/bash   
    ports:
      - 4001:4001/udp
      - 8080:8080
      - 5001:5001
      - 8080:8080
```

- 防火墙请放开相关端口
- 运行 shell 脚本

```bash
chmod +x ./create-ipfs-docker.sh
./create-ipfs-docker.sh
```

> 运行成功后，在浏览器中访问 `http://localhost:5001/webui` 或 `http://公网IP:5001/webui`，即可看到IPFS的webUI界面。
>
> 创建成功后，docker将启动一个名为 `ipfs_host`的容器，该容器将运行IPFS节点。
