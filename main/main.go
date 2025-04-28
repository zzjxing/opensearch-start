package main

import (
	"context"
	"fmt"
	"opensearch-start/config"
	"opensearch-start/resource"
	"opensearch-start/resource/opensearch"
)

func init() {
	config.Init()
	resource.Init()
}

func main() {
	ctx := context.Background()
	client := opensearch.Client()
	indexSetting := opensearch.GetVecDocConfig(1, 1, 3)
	indexName := "test_index"

	// 创建index
	if err := client.CreateIndex(ctx, indexName, indexSetting); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("create index\n\n")
	}
	// 单条插入document
	corpus := opensearch.NewDocument("1001", []float64{100.0, 100.0, 100.0})
	if err := client.InsertDocument(ctx, indexName, corpus); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("insert one document\n\n")
	}

	// 获取index下全部document
	if res, err := client.GetAllDocuments(indexName); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res)
	}

	// 批量插入
	vecDocs := []opensearch.Document{
		opensearch.NewDocument("1002", []float64{100.0, 101.0, 101.0}),
		opensearch.NewDocument("1003", []float64{101.0, 101.0, 101.0}),
		opensearch.NewDocument("1004", []float64{101.0, 100.0, 100.0}),
		opensearch.NewDocument("1005", []float64{10.0, 10.0, 10.0}),
		opensearch.NewDocument("1006", []float64{1000.0, 1000.0, 1000.0}),
		opensearch.NewDocument("1007", []float64{1.0, 1.0, 1.0}),
		opensearch.NewDocument("1008", []float64{10011.0, 10011.0, 10011.0, 10011.0}),
		opensearch.NewDocument("1009", []float64{1.0, 2.0, 3.0}),
		opensearch.NewDocument("1010", []float64{10.0, 10.0, 100.0}),
	}
	if failedIds, err := client.BulkInsertDocument(ctx, indexName, vecDocs); err != nil {
		if len(failedIds) != 0 {
			fmt.Printf("BulkInsertFailed:%v\n", failedIds)
		}
	} else {
		fmt.Printf("BulkInsert all success\n")
	}

	// 获取index下全部document
	if res, err := client.GetAllDocuments(indexName); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res)
	}

	// knn查询
	if res, err := client.SearchByKNN(ctx, indexName, []float64{100, 100, 100}, 3); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("SearchByKNN:%+v\n", res)
	}

	// 删除index
	if err := client.DeleteIndex(ctx, indexName); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("delete index")
	}
}
