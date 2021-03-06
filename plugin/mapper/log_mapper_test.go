package mapper

import (
	"github.com/aiops/logsight-filebeat/plugin/api"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogMapper_doMap(t *testing.T) {
	type fields struct {
		timestampMapper *StringMapper
		messageMapper   *StringMapper
		levelMapper     *StringMapper
		tagsMapper      *MultipleKeyValueStringMapper
	}

	logMapperFieldsPass1 := fields{
		timestampMapper: &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "2022-04-01T20:10:57+02:00"}},
		messageMapper:   &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
		levelMapper:     &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "INFO"}},
		tagsMapper: &MultipleKeyValueStringMapper{
			Mapper: MultipleKeyValueMapper{map[string]string{}},
		},
	}
	logExpectedPass1 := &api.Log{
		Timestamp: "2022-04-01T20:10:57+02:00",
		Message:   "test",
		Level:     "INFO",
		Tags:      map[string]string{},
	}

	logMapperFieldsPass2 := fields{
		timestampMapper: &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "2022-04-01T20:10:57"}},
		messageMapper:   &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
		levelMapper:     &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "INFO"}},
		tagsMapper: &MultipleKeyValueStringMapper{
			Mapper: MultipleKeyValueMapper{map[string]string{}},
		},
	}
	logExpectedPass2 := &api.Log{
		Timestamp: "2022-04-01T20:10:57",
		Message:   "test",
		Level:     "INFO",
		Tags:      map[string]string{},
	}

	logMapperFieldsFailLevel1 := fields{
		timestampMapper: &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "2022-04-01T20:10:57+02:00"}},
		messageMapper:   &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
		levelMapper:     &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "BOGUS"}},
		tagsMapper: &MultipleKeyValueStringMapper{
			Mapper: MultipleKeyValueMapper{map[string]string{}},
		},
	}
	logMapperFieldsFailLevel2 := fields{
		timestampMapper: &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "2022-04-01T20:10:57+02:00"}},
		messageMapper:   &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
		levelMapper:     &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "INFOINFO"}},
		tagsMapper: &MultipleKeyValueStringMapper{
			Mapper: MultipleKeyValueMapper{map[string]string{}},
		},
	}
	logMapperFieldsFailTime := fields{
		timestampMapper: &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "2022-04-01"}},
		messageMapper:   &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
		levelMapper:     &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "INFO"}},
		tagsMapper: &MultipleKeyValueStringMapper{
			Mapper: MultipleKeyValueMapper{map[string]string{}},
		},
	}

	type args struct {
		event beat.Event
	}

	testEvent := beat.Event{Fields: common.MapStr{"key1": "value1"}}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *api.Log
		wantErr bool
	}{
		{
			name:    "pass1",
			fields:  logMapperFieldsPass1,
			args:    args{event: testEvent},
			want:    logExpectedPass1,
			wantErr: false,
		},
		{
			name:    "pass2",
			fields:  logMapperFieldsPass2,
			args:    args{event: testEvent},
			want:    logExpectedPass2,
			wantErr: false,
		},
		{
			name:    "fail invalid level 1",
			fields:  logMapperFieldsFailLevel1,
			args:    args{event: testEvent},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail invalid level 2",
			fields:  logMapperFieldsFailLevel2,
			args:    args{event: testEvent},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "fail invalid time 2",
			fields:  logMapperFieldsFailTime,
			args:    args{event: testEvent},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := &LogMapper{
				TimestampMapper: tt.fields.timestampMapper,
				MessageMapper:   tt.fields.messageMapper,
				LevelMapper:     tt.fields.levelMapper,
				TagsMapper: &MultipleKeyValueStringMapper{
					Mapper: MultipleKeyValueMapper{map[string]string{}},
				},
			}
			got, err := lm.ToLog(tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "DoMap(%v)", tt.args.event)
		})
	}
}
