package plugin

import (
	"github.com/aiops/logsight-filebeat/plugin/logsight"
	"github.com/elastic/beats/v7/libbeat/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogMapper_doMap(t *testing.T) {
	type fields struct {
		timestampMapper StringMapper
		messageMapper   StringMapper
		levelMapper     StringMapper
		metadataMapper  StringMapper
	}
	logMapperFields := fields{
		timestampMapper: StringMapper{mapper: &ConstantStringMapper{constantString: "test"}},
		messageMapper:   StringMapper{mapper: &ConstantStringMapper{constantString: "test"}},
		levelMapper:     StringMapper{mapper: &ConstantStringMapper{constantString: "test"}},
		metadataMapper:  StringMapper{mapper: &ConstantStringMapper{constantString: "test"}},
	}
	type args struct {
		mapSource common.MapStr
	}
	logExpected := &logsight.Log{
		Timestamp: "test",
		Message:   "test",
		Level:     "test",
		Metadata:  "test",
	}
	testMap := common.MapStr{"key1": "value1"}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *logsight.Log
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
				timestampMapper: tt.fields.timestampMapper,
				messageMapper:   tt.fields.messageMapper,
				levelMapper:     tt.fields.levelMapper,
				metadataMapper:  tt.fields.metadataMapper,
			}
			got, err := lm.toLog(tt.args.mapSource)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "doMap(%v)", tt.args.mapSource)
		})
	}
}

func TestLogBatchMapper_doMap(t *testing.T) {
	type fields struct {
		applicationNameMapper StringMapper
		tagMapper             StringMapper
		logMapper             LogMapper
	}
	logBatchMapperFields := fields{
		applicationNameMapper: StringMapper{mapper: &ConstantStringMapper{constantString: "test"}},
		tagMapper:             StringMapper{mapper: &ConstantStringMapper{constantString: "test"}},
		logMapper: LogMapper{
			timestampMapper: StringMapper{mapper: &ConstantStringMapper{constantString: "test"}},
			messageMapper:   StringMapper{mapper: &ConstantStringMapper{constantString: "test"}},
			levelMapper:     StringMapper{mapper: &ConstantStringMapper{constantString: "test"}},
			metadataMapper:  StringMapper{mapper: &ConstantStringMapper{constantString: "test"}},
		},
	}
	type args struct {
		mapSources []common.MapStr
	}
	testMap := common.MapStr{"key1": "value1"}
	mapSources := []common.MapStr{testMap, testMap, testMap}
	logExpected := &logsight.Log{
		Timestamp: "test",
		Message:   "test",
		Level:     "test",
		Metadata:  "test",
	}
	logBatchExpected := []*logsight.LogBatch{
		{
			ApplicationName: "test",
			Tag:             "test",
			Logs:            []*logsight.Log{logExpected, logExpected, logExpected},
		},
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*logsight.LogBatch
		wantErr bool
	}{
		{
			name:    "pass",
			fields:  logBatchMapperFields,
			args:    args{mapSources: mapSources},
			want:    logBatchExpected,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lbm := &LogBatchMapper{
				applicationNameMapper: tt.fields.applicationNameMapper,
				tagMapper:             tt.fields.tagMapper,
				logMapper:             tt.fields.logMapper,
			}
			got, err := lbm.toLogBatch(tt.args.mapSources)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "doMap(%v)", tt.args.mapSources)
		})
	}
}
