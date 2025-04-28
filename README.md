# OpenSearch 操作封装项目

本项目基于 [opensearch-go](https://github.com/opensearch-project/opensearch-go) 对 OpenSearch 的常用操作进行了封装，提供了简单易用的 API，包括索引创建、文档插入、KNN 搜索、批量操作等。

---

## 功能简介

- **索引管理**：创建、删除索引，支持 KNN 配置。
- **文档操作**：支持单条和批量文档插入，获取全部文档。
- **KNN 查询**：基于向量的 KNN 搜索，快速检索相似数据。
- **环境变量配置**：通过环境变量简化 OpenSearch 客户端的初始化。

---

## 项目结构

```plaintext
opensearch-start/
├── config/                   # 配置模块
│   └── config.go             # 环境变量初始化
├── main/                     # 主程序入口
│   └── main.go               # 示例代码，运行主程序
├── resource/                 # 核心功能模块
│   ├── common/               # 通用工具类
│   │   └── common.go         # 包含查询模板和公共结构体
│   ├── opensearch/           # OpenSearch 操作封装
│   │   ├── document.go       # 文档操作
│   │   ├── opensearch.go     # OpenSearch 客户端初始化及核心功能
│   │   └── resource.go       # 资源管理
├── go.mod                    # Go modules 配置文件
├── .gitignore                # Git 忽略文件
```

---

## 环境准备

### 1. 安装 OpenSearch

#### 使用 Docker 安装（推荐）

运行以下命令快速搭建 OpenSearch 环境：

```bash
docker run -d --name opensearch \
  -e "discovery.type=single-node" \
  -e "plugins.security.disabled=false" \
  -e "OPENSEARCH_SECURITY_ADMIN_PASSWORD=admin" \
  -e "OPENSEARCH_SECURITY_ADMIN_USERNAME=admin" \
  -p 9200:9200 \
  -p 9600:9600 \
  opensearchproject/opensearch:latest
```

- **`discovery.type=single-node`**：单节点模式。
- **`plugins.security.disabled=false`**：启用安全插件。
- **`OPENSEARCH_SECURITY_ADMIN_USERNAME` 和 `OPENSEARCH_SECURITY_ADMIN_PASSWORD`**：设置管理员用户名和密码。

验证安装是否成功：
```bash
curl -u admin:admin http://localhost:9200
```

---

### 2. 配置环境变量

在运行程序之前，需配置以下环境变量：

- `OPENSEARCH_HOSTS`：OpenSearch 服务地址，例如 `http://localhost:9200`。
- `OPENSEARCH_USERNAME`：OpenSearch 用户名，例如 `admin`。
- `OPENSEARCH_PASSWORD`：OpenSearch 密码，例如 `admin`。

#### 示例配置（`~/.bash_profile` 或 `~/.zshrc`）

```bash
export OPENSEARCH_HOSTS="http://localhost:9200"
export OPENSEARCH_USERNAME="admin"
export OPENSEARCH_PASSWORD="admin"
```

加载配置：
```bash
source ~/.bash_profile
```

---

## 快速开始

### 1. 安装依赖

确保已安装 Go 环境（建议 Go 1.18+），然后运行以下命令安装依赖：

```bash
go mod tidy
```

### 2. 运行程序

通过以下命令运行 `main.main`：

```bash
go run main/main.go
```

---

## 示例功能

以下是项目中主要功能的示例说明。

### 初始化环境

`config.Init()` 和 `resource.Init()` 用于初始化环境变量和 OpenSearch 客户端。

```go
func init() {
	config.Init()
	resource.Init()
}
```

### 创建索引

使用 `CreateIndex` 方法创建索引，并配置副本数、分片数及向量维度：

```go
ctx := context.Background()
client := opensearch.Client()
indexSetting := opensearch.GetVecDocConfig(1, 1, 3) // 副本数:1, 分片数:1, 向量维度:3
indexName := "test_index"

if err := client.CreateIndex(ctx, indexName, indexSetting); err != nil {
	fmt.Println(err)
} else {
	fmt.Printf("Index created: %s\n", indexName)
}
```

### 插入单条文档

使用 `InsertDocument` 方法插入单条文档：

```go
doc := opensearch.NewDocument("1001", []float64{100.0, 100.0, 100.0})

if err := client.InsertDocument(ctx, indexName, doc); err != nil {
	fmt.Println(err)
} else {
	fmt.Println("Document inserted")
}
```

### 批量插入文档

使用 `BulkInsertDocument` 方法批量插入文档：

```go
docs := []opensearch.Document{
	opensearch.NewDocument("1002", []float64{101.0, 101.0, 101.0}),
	opensearch.NewDocument("1003", []float64{102.0, 102.0, 102.0}),
}

if failedIds, err := client.BulkInsertDocument(ctx, indexName, docs); err != nil {
	fmt.Printf("Bulk insert failed for IDs: %v\n", failedIds)
} else {
	fmt.Println("Bulk insert successful")
}
```

### KNN 查询

使用 `SearchByKNN` 方法进行基于向量的 KNN 搜索：

```go
queryVec := []float64{100.0, 100.0, 100.0}
k := 3 // 搜索前3个最相似的向量

if res, err := client.SearchByKNN(ctx, indexName, queryVec, k); err != nil {
	fmt.Println(err)
} else {
	fmt.Printf("KNN search results: %+v\n", res)
}
```

### 删除索引

使用 `DeleteIndex` 方法删除索引：

```go
if err := client.DeleteIndex(ctx, indexName); err != nil {
	fmt.Println(err)
} else {
	fmt.Printf("Index deleted: %s\n", indexName)
}
```

---

## 常见问题

### 无法连接 OpenSearch
- 确认 OpenSearch 服务是否已启动。
- 检查 `OPENSEARCH_HOSTS` 环境变量是否配置正确。

### 插入文档失败
- 确认文档的数据结构是否与索引定义的 schema 匹配。
- 检查向量维度是否正确。

---

## 贡献

欢迎提交 Issue 和 Pull Request 来改进本项目！

---
*本文件由 Copilot 生成。*