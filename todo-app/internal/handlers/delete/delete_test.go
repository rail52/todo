package delete

import (
	"log/slog"
	"net/http"
	"reflect"
	"testing"

	"github.com/rail52/myprojects/dbpb"
)

func TestDeleteTask(t *testing.T) {
	type args struct {
		log           *slog.Logger
		storage       dbpb.PostgresClient
		kafkaProducer KafkaProducer
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeleteTask(tt.args.log, tt.args.storage, tt.args.kafkaProducer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteTask() = %v, want %v", got, tt.want)
			}
		})
	}
}
