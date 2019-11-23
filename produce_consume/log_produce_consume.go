package produce_consume

import (
	"context"
	"github.com/gocomponents/core/proto"
	"github.com/gocomponents/core/util"
	"github.com/gocomponents/glogstash/config"
	"github.com/golang/glog"
	"github.com/olivere/elastic/v7"
	"sync"
	"time"
)

var logCh = make(chan *proto.Log, 500)

var client *elastic.Client

var currentIndexName string

var currentIndexNameMutex sync.Mutex

func init() {
	var err error
	esConfig :=config.GetElasticConfig()
	client, err = elastic.NewClient(elastic.SetURL(esConfig), elastic.SetSniff(false))
	if nil != err {
		panic(err)
	}
	currentIndexName,err=getIndexName(time.Now().Format("2006-01-02 15:04:05"))
	if err!=nil {
		glog.Warningf("get currentIndexName error,%s",err.Error())
	}
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
func createIndex(indexName string) error {
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
				defer func() {
					if err:=recover();err!=nil {
						glog.Errorf("error:%v,log:%v",err,log)
					}
				}()

				indexName, err := getIndexName(log.CreateTime)
				if nil != err {
					panic(err)
				}

				if currentIndexName!=indexName {
					currentIndexNameMutex.Lock()
					defer currentIndexNameMutex.Unlock()
					err=createIndex(indexName)
					if nil!=err {
						panic(err)
					}
					currentIndexName=indexName
					glog.Infof("createIndex(%s) success",indexName)
				}

				_, err = client.Index().
					Index(indexName).
					Id(util.GetGUID()).
					BodyJson(log).
					Do(context.Background())
				if nil != err {
					panic(err)
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
