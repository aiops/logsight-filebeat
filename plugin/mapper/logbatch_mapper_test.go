package mapper

import (
	"github.com/aiops/logsight-filebeat/plugin/api"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogBatchMapper_ToLogBatch(t *testing.T) {
	type fields struct {
		ApplicationNameMapper *StringMapper
		TagMapper             *StringMapper
		LogMapper             *LogMapper
	}
	logBatchMapperFieldsGood := fields{
		ApplicationNameMapper: &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
		TagMapper:             &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
		LogMapper: &LogMapper{
			TimestampMapper: &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "2022-04-01T20:10:57+02:00"}},
			MessageMapper:   &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
			LevelMapper:     &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "INFO"}},
			MetadataMapper:  &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
		},
	}
	logBatchMapperFieldsBadLevel := fields{
		ApplicationNameMapper: &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
		TagMapper:             &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
		LogMapper: &LogMapper{
			TimestampMapper: &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "2022-04-01T20:10:57+02:00"}},
			MessageMapper:   &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
			LevelMapper:     &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "BOGUS"}},
			MetadataMapper:  &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
		},
	}

	type args struct {
		events []publisher.Event
	}
	testMap := common.MapStr{"key1": "value1"}
	testEvent := publisher.Event{
		Content: beat.Event{Fields: testMap},
		Flags:   0,
		Cache:   publisher.EventCache{},
	}
	events := []publisher.Event{testEvent, testEvent, testEvent}

	log := &api.Log{
		Timestamp: "2022-04-01T20:10:57+02:00",
		Message:   "test",
		Level:     "INFO",
		Metadata:  "test",
	}

	logBatchExpected := &api.LogBatch{
		ApplicationName: "test",
		Tag:             "test",
		Logs:            []*api.Log{log, log, log},
	}
	mappedLogBatchExpected := []*MappedLogBatch{
		{
			LogBatch: logBatchExpected,
			Events:   []publisher.Event{testEvent, testEvent, testEvent},
		},
	}
	failedEventsExpected := &FailedEvents{
		Events: []*publisher.Event{&testEvent, &testEvent, &testEvent},
		Errs:   []error{api.InvalidLevelError, api.InvalidLevelError, api.InvalidLevelError},
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*MappedLogBatch
		want1  *FailedEvents
	}{
		{
			name:   "pass",
			fields: logBatchMapperFieldsGood,
			args:   args{events: events},
			want:   mappedLogBatchExpected,
			want1:  nil,
		},
		{
			name:   "fail bad level",
			fields: logBatchMapperFieldsBadLevel,
			args:   args{events: events},
			want:   []*MappedLogBatch{},
			want1:  failedEventsExpected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lbm := &LogBatchMapper{
				ApplicationNameMapper: tt.fields.ApplicationNameMapper,
				TagMapper:             tt.fields.TagMapper,
				LogMapper:             tt.fields.LogMapper,
			}
			got, got1 := lbm.ToLogBatch(tt.args.events)
			assert.Equalf(t, tt.want, got, "ToLogBatch(%v)", tt.args.events)
			assert.Equalf(t, tt.want1, got1, "ToLogBatch(%v)", tt.args.events)
		})
	}
}
