package logsight

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testMap = common.MapStr{
		"key": common.MapStr{
			"key1": "value1",
		},
		"key3": "value2",
		"key4": 4,
	}
)

func TestConstantMapper_doMap(t *testing.T) {
	type fields struct {
		constantString string
	}
	type args struct {
		ignored common.MapStr
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "pass",
			fields:  fields{constantString: "app_name"},
			args:    args{ignored: testMap},
			want:    "app_name",
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &ConstantStringMapper{
				constantString: tt.fields.constantString,
			}
			got, err := cm.doMap(tt.args.ignored)
			if (err != nil) != tt.wantErr {
				t.Errorf("doMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("doMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyMapper_doMap(t *testing.T) {
	type fields struct {
		key string
	}
	type args struct {
		mapSource common.MapStr
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "pass simple key",
			fields:  fields{"key3"},
			args:    args{testMap},
			want:    "value2",
			wantErr: false,
		},
		{
			name:    "pass nested key",
			fields:  fields{"key.key1"},
			args:    args{testMap},
			want:    "value1",
			wantErr: false,
		},
		{
			name:    "pass value is not a string",
			fields:  fields{"key"},
			args:    args{testMap},
			want:    common.MapStr{"key1": "value1"},
			wantErr: false,
		},
		{
			name:    "fail key not found",
			fields:  fields{"key5"},
			args:    args{testMap},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			km := &KeyMapper{
				key: tt.fields.key,
			}
			got, err := km.doMap(tt.args.mapSource)
			if (err != nil) != tt.wantErr {
				t.Errorf("doMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// common.MapStr are not comparable by default. This is why this assert library is used to check equality
			assert.Equal(t, got, tt.want, fmt.Sprintf("doMap() got = %v, want %v", got, tt.want))
		})
	}
}

func TestKeyRegexMapper_doMap(t *testing.T) {
	type fields struct {
		mapper StringMapper
		expr   *regexp.Regexp
	}
	stringMapper := StringMapper{
		mapper: &KeyMapper{
			key: "key3",
		},
	}

	type args struct {
		mapSource common.MapStr
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "pass",
			fields:  fields{mapper: stringMapper, expr: regexp.MustCompile("va(.*)e")},
			args:    args{mapSource: testMap},
			want:    "lu",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			krm := &KeyRegexMapper{
				mapper: tt.fields.mapper,
				expr:   tt.fields.expr,
			}
			got, err := krm.doMap(tt.args.mapSource)
			if (err != nil) != tt.wantErr {
				t.Errorf("doMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("doMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyRegexMapper_applyRegex(t *testing.T) {
	type fields struct {
		mapper StringMapper
		expr   *regexp.Regexp
	}
	constMapper := StringMapper{
		mapper: &ConstantStringMapper{
			constantString: "key",
		},
	}

	type args struct {
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "pass path submatch",
			fields:  fields{constMapper, regexp.MustCompile(".*/(.*)/.*")},
			args:    args{"/path/app_name/here"},
			want:    "app_name",
			wantErr: false,
		},
		{
			name:    "fail no match",
			fields:  fields{constMapper, regexp.MustCompile(".*/(.*)/.*")},
			args:    args{"/path"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "fail empty string match",
			fields:  fields{constMapper, regexp.MustCompile(".*/(.*)/.*")},
			args:    args{"/path//here"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			krm := &KeyRegexMapper{
				mapper: tt.fields.mapper,
				expr:   tt.fields.expr,
			}
			got, err := krm.applyRegex(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("applyRegex() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringMapper_doStringMap(t *testing.T) {
	type fields struct {
		mapper Mapper
	}
	type args struct {
		mapSource common.MapStr
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "pass",
			fields:  fields{mapper: &ConstantStringMapper{constantString: "test"}},
			args:    args{mapSource: testMap},
			want:    "test",
			wantErr: false,
		},
		{
			name:    "pass empty string",
			fields:  fields{mapper: &ConstantStringMapper{constantString: ""}},
			args:    args{mapSource: testMap},
			want:    "",
			wantErr: false,
		},
		{
			name:    "fail int",
			fields:  fields{mapper: &KeyMapper{"key4"}},
			args:    args{mapSource: testMap},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &StringMapper{
				mapper: tt.fields.mapper,
			}
			got, err := sm.doStringMap(tt.args.mapSource)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "doStringMap(%v)", tt.args.mapSource)
		})
	}
}

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
	logExpected := &Log{
		timestamp: "test",
		message:   "test",
		level:     "test",
		metadata:  "test",
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Log
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
			got, err := lm.doMap(tt.args.mapSource)
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
	mapSources := []common.MapStr{testMap, testMap, testMap}
	logExpected := &Log{
		timestamp: "test",
		message:   "test",
		level:     "test",
		metadata:  "test",
	}
	logBatchExpected := []*LogBatch{
		{
			applicationName: "test",
			tag:             "test",
			logs:            []*Log{logExpected, logExpected, logExpected},
		},
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*LogBatch
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
			got, err := lbm.doMap(tt.args.mapSources)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "doMap(%v)", tt.args.mapSources)
		})
	}
}
