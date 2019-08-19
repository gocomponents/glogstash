package produce_consume

import (
	"context"
	"fmt"
	"github.com/gocomponents/core/proto"
	"github.com/gocomponents/core/util"
	"github.com/olivere/elastic/v7"
	"glogstash/config"
	"time"
)

var logCh = make(chan *proto.Log, 500)

var client *elastic.Client

func init() {
	var err error
	esConfig := config.GetElasticConfig()
	client, err = elastic.NewClient(elastic.SetURL(esConfig), elastic.SetSniff(false))
	if nil != err {
		panic(err)
	}
	err = createIndex()
	if nil != err {
		fmt.Println("create index Error", err)
	}
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for {
			select {
			case <-ticker.C:
				{
					err = createIndex()
					if nil != err {
						fmt.Println("create index Error", err)
					}
				}
			}
		}
	}()
}

const logMapping = `
{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 3
	},
	"mappings":{
			"properties":{
				"app":{
					"type":"keyword"
				},
				"module":{
					"type":"keyword"
				},
				"level":{
					"type":"integer"
				},
				"trace_id":{
					"type":"keyword"
				},
				"message":{
					"type":"text"
				},
				"exception":{
					"type":"text"
				},
				"user_ip":{
					"type":"keyword"
				},
				"exec_time":{
					"type":"integer"
				},
				"create_time":{
					"type":"keyword"
				}
			}
	}
}`

//TODO:定时创建新的index
func createIndex() error {
	indexName, err := getIndexName(time.Now().Format("2006-01-02 15:04:05"))
	if nil != err {
		return err
	}
	exists, err := client.IndexExists(indexName).Do(context.Background())
	if nil != err {
		return err
	}
	if !exists {
		_, err = client.CreateIndex(indexName).BodyString(logMapping).Do(context.Background())
		if nil != err {
			return err
		}
	}
	return nil
}

func Produce(log *proto.Log) {
	logCh <- log
}

func Consume() {
	for {
		log, ok := <-logCh
		if ok {
			go func(log *proto.Log) {
				indexName, err := getIndexName(log.CreateTime)
				if nil != err {
					fmt.Println("ES GetIndexName Error", err)
					return
				}
				if nil != err {
					fmt.Println("ES BodyJsonMarshal Error", err)
					return
				}
				_, err = client.Index().
					Index(indexName).
					Id(util.GetGUID()).
					BodyJson(log).
					Do(context.Background())
				if nil != err {
					fmt.Println("ES Index Error", err)
				}
			}(log)
		}
	}
}

func getIndexName(createTime string) (string, error) {
	time, err := time.Parse("2006-01-02 15:04:05", createTime)
	if err != nil {
		return "", err
	}
	return time.Format("20060102"), nil
}
