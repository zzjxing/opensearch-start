package opensearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"opensearch-start/resource/common"
	"strings"
	"sync"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"

	"opensearch-start/config"
)

const logPrefix = "opensearch"

var once sync.Once

func Init() {
	once.Do(func() {
		client, err := opensearch.NewClient(opensearch.Config{
			Addresses: []string{config.OpenSearchADDR},
			Username:  config.OpenSearchUser,
			Password:  config.OpenSearchPassword,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // 禁用证书验证
				},
			},
		})
		if err != nil {
			panic(err)
		}
		osClient = &OpenSearchClient{client}
	})
}

var osClient *OpenSearchClient

func Client() *OpenSearchClient {
	return osClient
}

type OpenSearchClient struct {
	client *opensearch.Client
}

func (osClient *OpenSearchClient) Ping(ctx context.Context) error {
	resp, err := osClient.client.Cluster.Health(
		osClient.client.Cluster.Health.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to OpenSearch: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return parseError(resp)
}

func (osClient *OpenSearchClient) CreateIndex(ctx context.Context, index string, body string) error {
	bodyReader := bytes.NewReader([]byte(body))
	resp, err := osClient.client.Indices.Create(
		index,
		osClient.client.Indices.Create.WithBody(bodyReader),
		osClient.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %s", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return parseError(resp)
}

// DeleteIndex 删除指定的索引
func (osClient *OpenSearchClient) DeleteIndex(ctx context.Context, index string) error {
	resp, err := osClient.client.Indices.Delete(
		[]string{index}, // 索引名称
		osClient.client.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to delete index: %s", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return parseError(resp)
}

func (osClient *OpenSearchClient) InsertDocument(ctx context.Context, index string, doc Document) error {
	body, err := doc.Bytes()
	if err != nil {
		return err
	}

	resp, err := osClient.client.Index(
		index,
		bytes.NewReader(body),
		osClient.client.Index.WithDocumentID(doc.GetID()),
		osClient.client.Index.WithContext(ctx), // 加上 context
	)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return parseError(resp)
}

func (osClient *OpenSearchClient) BulkInsertDocument(ctx context.Context, index string, docs []Document) ([]string, error) {
	var buf bytes.Buffer
	var failedIds []string
	for _, doc := range docs {
		body, err := doc.Bytes()
		if err != nil {
			log.Printf("%s: failed to marshal document, err: %v, body: %v", logPrefix, err, body)
			continue
		}
		// 指定每条数据的index和id
		meta := fmt.Sprintf(`{ "index" : { "_index" : "%s", "_id" : "%s" } }%s`, index, doc.GetID(), "\n")
		_, err = buf.WriteString(meta)
		if err != nil {
			failedIds = append(failedIds, doc.GetID())
			log.Printf("%s: failed to write document: %s", logPrefix, err)
			continue
		}
		body = append(body, byte('\n'))
		_, err = buf.Write(body)
		if err != nil {
			failedIds = append(failedIds, doc.GetID())
			log.Printf("%s: failed to write document: %s", logPrefix, err)
			continue
		}
	}

	resp, err := osClient.client.Bulk(
		bytes.NewReader(buf.Bytes()),
		osClient.client.Bulk.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk insert document: %w", err)
	}

	// 解析批量响应结果
	br, err := decodeResponse[common.BulkResponse](resp.Body)
	if err != nil {
		return failedIds, fmt.Errorf("failed to decode bulk response: %w", err)
	}

	if !br.Errors {
		return nil, nil // 全部成功
	}
	for _, item := range br.Items {
		if item.Index.Status >= 400 {
			log.Printf("%s: failed to index document id %s, error: %+v", logPrefix, item.Index.ID, item.Index.Error)
			failedIds = append(failedIds, item.Index.ID)
		}
	}
	return failedIds, fmt.Errorf("failed to index document ids: %s", strings.Join(failedIds, ","))
}

func (osClient *OpenSearchClient) SearchByKNN(ctx context.Context, index string, embedding []float64, k int) (*common.SearchResult, error) {
	emStr, _ := json.Marshal(embedding)
	query := fmt.Sprintf(common.BaseKnnQuery, k, emStr, k)
	bodyReader := bytes.NewReader([]byte(query))

	// 调用 OpenSearch 的 Search API 执行查询
	resp, err := osClient.client.Search(
		osClient.client.Search.WithContext(ctx),     // 使用上下文控制请求
		osClient.client.Search.WithIndex(index),     // 查询指定的索引
		osClient.client.Search.WithBody(bodyReader), // 查询体
	)

	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %s", err)
	}
	defer resp.Body.Close()

	// 解析搜索响应
	searchResp, err := decodeResponse[common.SearchResult](resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode knn-search response: %s", err)
	}
	return searchResp, parseError(resp)
}

// GetAllDocuments 获取Index下的全部文档，ps：debug
func (osClient *OpenSearchClient) GetAllDocuments(index string) (*common.SearchResult, error) {
	// 查询体
	query := `{
		"size": 10000,  
		"query": {
			"match_all": {} 
		}
	}`

	bodyReader := bytes.NewReader([]byte(query))

	// 调用 OpenSearch 的 Search API 执行查询
	resp, err := osClient.client.Search(
		osClient.client.Search.WithContext(context.Background()), // 使用上下文控制请求
		osClient.client.Search.WithIndex(index),                  // 查询指定的索引
		osClient.client.Search.WithBody(bodyReader),              // 查询体
	)

	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %s", err)
	}
	defer resp.Body.Close()

	// 解析搜索响应
	searchResp, err := decodeResponse[common.SearchResult](resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode knn-search response: %s", err)
	}
	return searchResp, parseError(resp)
}

// parseError 解析opensearch返回结果中的错误信息
// 解析错误信息应在最后调用，调用后resp.Body不可再读取
func parseError(resp *opensearchapi.Response) error {
	if resp.Body == nil {
		return fmt.Errorf("%s: response body is empty", logPrefix)
	}

	if !resp.IsError() {
		return nil
	}

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("%s: (err status: %s), and failed to read response body: %v", logPrefix, resp.Status(), readErr)
	}
	return fmt.Errorf("%s (err status: %s): %s", logPrefix, resp.Status(), string(respBody))
}

// decodeResponse 解析opensearch返回的结果
func decodeResponse[T any](body io.ReadCloser) (*T, error) {
	defer body.Close()
	var t T
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}
