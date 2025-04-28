package common

const BaseKnnQuery = `{
			"size": %d,
			"query": {
				"knn": {
					"embedding": {
						"vector": %s,
						"k": %d
					}
				}
			}
		}`

// SearchResult opensearch的查询接口返回结构
type SearchResult struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Hits     Hits `json:"hits"`
}
type Hits struct {
	Total    Total       `json:"total"`
	MaxScore float64     `json:"max_score"`
	Hits     []HitDetail `json:"hits"`
}

type Total struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}

type HitDetail struct {
	Index  string                 `json:"_index"`
	ID     string                 `json:"_id"`
	Score  float64                `json:"_score"`
	Source map[string]interface{} `json:"_source"`
}

// BulkResponse 批量插入接口返回结构
type BulkResponse struct {
	Errors bool       `json:"errors"`
	Items  []BulkItem `json:"items"`
}
type BulkItem struct {
	Index struct {
		ID     string                 `json:"_id"`
		Status int                    `json:"status"`
		Error  map[string]interface{} `json:"error,omitempty"`
	} `json:"index"`
}
