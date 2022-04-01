package mapper

import (
	"github.com/aiops/logsight-filebeat/plugin/api"
	"github.com/elastic/beats/v7/libbeat/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogMapper_doMap(t *testing.T) {
	type fields struct {
		timestampMapper *StringMapper
		messageMapper   *StringMapper
		levelMapper     *StringMapper
		metadataMapper  *StringMapper
	}
	logMapperFields := fields{
		timestampMapper: &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "2022-04-01T20:10:57+02:00"}},
		messageMapper:   &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
		levelMapper:     &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "INFO"}},
		metadataMapper:  &StringMapper{Mapper: &ConstantStringMapper{ConstantString: "test"}},
	}
	type args struct {
		mapSource common.MapStr
	}
	logExpected := &api.Log{
		Timestamp: "2022-04-01T20:10:57+02:00",
		Message:   "test",
		Level:     "INFO",
		Metadata:  "test",
	}
	testMap := common.MapStr{"key1": "value1"}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *api.Log
		wantErr bool
	}{
		{
			name:    "pass",
			fields:  logMapperFields,
			args:    args{mapSource: testMap},
			want:    logExpected,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := &LogMapper{
				TimestampMapper: tt.fields.timestampMapper,
				MessageMapper:   tt.fields.messageMapper,
				LevelMapper:     tt.fields.levelMapper,
				MetadataMapper:  tt.fields.metadataMapper,
			}
			got, err := lm.ToLog(tt.args.mapSource)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "DoMap(%v)", tt.args.mapSource)
		})
	}
}
