package produce_consume

import (
	"context"
	"fmt"
	"github.com/gocomponents/core/proto"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestConsume_getIndexName(t *testing.T) {
	indexName, err := getIndexName(time.Now().Format("2006-01-02 15:04:05"))
	if nil != err {
		t.Error(err)
	}
	fmt.Println(indexName)
}

func TestProduce(t *testing.T) {
	go Consume()
	for i := 0; i < 1000; i++ {
		log := proto.Log{
			App:        "test",
			Module:     "consume",
			Level:      proto.Log_Info,
			TraceId:    "123",
			Message:    "456",
			Exception:  "456",
			UserIp:     "192.168.11.11",
			ExecTime:   12,
			CreateTime: time.Now().Add(time.Duration(i) * time.Millisecond).Format("2006-01-02 15:04:05"),
		}
		Produce(&log)
	}

	<-make(chan int)
}

//TODO:grpc client
func TestProduce2(t *testing.T) {
	conn, err := grpc.Dial("47.244.216.246:18080", grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
	}

	defer conn.Close()
	client := proto.NewLogStashClient(conn)

	for i := 0; i < 10000; i++ {
		log := proto.Log{
			App:        "test",
			Module:     "consume",
			Level:      proto.Log_Info,
			TraceId:    "123",
			Message:    "456",
			Exception:  nil,
			UserIp:     "192.168.11.11",
			ExecTime:   12,
			CreateTime: time.Now().Add(time.Duration(i) * time.Millisecond).Format("2006-01-02 15:04:05"),
		}

		timer := time.Now()
		_,err:=client.Send(context.Background(), &log)
		fmt.Println(err)
		fmt.Println(time.Since(timer))

	}
}
