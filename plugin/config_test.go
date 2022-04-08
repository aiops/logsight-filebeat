package plugin

import (
	"github.com/aiops/logsight-filebeat/plugin/mapper"
	"reflect"
	"regexp"
	"testing"
)

func Test_application_toMapper(t *testing.T) {
	type fields struct {
		Name         string
		Map          string
		RegexMatcher string
		AutoCreate   bool
	}
	name := "default"
	key := "test.test"
	exprValid := "^.*([T|t]est).*$"
	exprValidCompiled := regexp.MustCompile(exprValid)
	exprInvalid := "^.*($[T|t]est.*$"

	wantRegexpMapper := mapper.KeyRegexMapper{
		Mapper: mapper.StringMapper{Mapper: mapper.KeyMapper{Key: "test.test"}},
		Expr:   exprValidCompiled,
	}
	wantKeyMapper := mapper.KeyMapper{Key: key}
	wantNameMapper := mapper.ConstantStringMapper{ConstantString: name}

	tests := []struct {
		name    string
		fields  fields
		want    mapper.Mapper
		wantErr bool
	}{
		{
			name: "pass regex",
			fields: fields{
				Name:         "",
				Map:          key,
				RegexMatcher: exprValid,
			},
			want:    &wantRegexpMapper,
			wantErr: false,
		},
		{
			name: "fail invalid regex",
			fields: fields{
				Name:         "",
				Map:          key,
				RegexMatcher: exprInvalid,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "pass key",
			fields: fields{
				Name:         "",
				Map:          key,
				RegexMatcher: "",
			},
			want:    &wantKeyMapper,
			wantErr: false,
		},
		{
			name: "pass name",
			fields: fields{
				Name:         name,
				Map:          "",
				RegexMatcher: "",
			},
			want:    &wantNameMapper,
			wantErr: false,
		},
		{
			name: "fail",
			fields: fields{
				Name:         "",
				Map:          "",
				RegexMatcher: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "pass test hierarchy regex",
			fields: fields{
				Name:         name,
				Map:          key,
				RegexMatcher: exprValid,
			},
			want:    &wantRegexpMapper,
			wantErr: false,
		},
		{
			name: "pass test hierarchy key",
			fields: fields{
				Name:         name,
				Map:          key,
				RegexMatcher: "",
			},
			want:    &wantKeyMapper,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &mapperConf{
				Name:         tt.fields.Name,
				Key:          tt.fields.Map,
				RegexMatcher: tt.fields.RegexMatcher,
			}
			got, err := a.toMapper()
			if (err != nil) != tt.wantErr {
				t.Errorf("toMapper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toMapper() got = %v, want %v", got, tt.want)
			}
		})
	}
}
