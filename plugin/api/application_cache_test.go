package api

import (
	"github.com/google/uuid"
	"reflect"
	"testing"
)

func TestNewApplicationCache(t *testing.T) {
	tests := []struct {
		name string
		want *applicationCache
	}{
		{
			name: "pass",
			want: &applicationCache{
				cache: make(map[string]*Application),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewApplicationCache(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewApplicationCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_applicationCache_add(t *testing.T) {
	type fields struct {
		cache map[string]*Application
	}
	type args struct {
		application *Application
	}
	appName := "Test"
	testApp := &Application{
		Id:   &uuid.UUID{},
		Name: &appName,
	}
	testAppEmpty := &Application{}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "pass",
			fields: fields{cache: map[string]*Application{}},
			args:   args{application: testApp},
		},
		{
			name:   "pass",
			fields: fields{cache: map[string]*Application{}},
			args:   args{application: testAppEmpty},
		},
		{
			name:   "pass",
			fields: fields{cache: map[string]*Application{}},
			args:   args{application: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &applicationCache{
				cache: tt.fields.cache,
			}
			ac.add(tt.args.application)
			if tt.args.application != nil && tt.args.application.Name != nil && !ac.contains(appName) {
				t.Errorf("cache must contain added non-nil application %v", testApp)
				return
			}
		})
	}
}

func Test_applicationCache_addAll(t *testing.T) {
	type fields struct {
		cache map[string]*Application
	}
	type args struct {
		applications []*Application
	}
	appName1 := "Test1"
	testApp1 := &Application{
		Id:   &uuid.UUID{},
		Name: &appName1,
	}
	appName2 := "Test2"
	testApp2 := &Application{
		Id:   &uuid.UUID{},
		Name: &appName2,
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantSize int
	}{
		{
			name:     "pass one",
			fields:   fields{cache: map[string]*Application{}},
			args:     args{applications: []*Application{testApp1}},
			wantSize: 1,
		},
		{
			name:     "pass two",
			fields:   fields{cache: map[string]*Application{}},
			args:     args{applications: []*Application{testApp1, testApp2}},
			wantSize: 2,
		},
		{
			name:     "pass same twice",
			fields:   fields{cache: map[string]*Application{}},
			args:     args{applications: []*Application{testApp1, testApp1}},
			wantSize: 1,
		},
		{
			name:     "pass app default",
			fields:   fields{cache: map[string]*Application{}},
			args:     args{applications: []*Application{{}}},
			wantSize: 0,
		},
		{
			name:     "pass empty",
			fields:   fields{cache: map[string]*Application{}},
			args:     args{applications: []*Application{}},
			wantSize: 0,
		},
		{
			name:     "pass app nil",
			fields:   fields{cache: map[string]*Application{}},
			args:     args{applications: []*Application{nil}},
			wantSize: 0,
		},
		{
			name:     "pass nil",
			fields:   fields{cache: map[string]*Application{}},
			args:     args{applications: nil},
			wantSize: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &applicationCache{
				cache: tt.fields.cache,
			}
			ac.addAll(tt.args.applications)
			if len(ac.getAll()) != tt.wantSize {
				t.Errorf("cache must contain %v added applications", tt.wantSize)
				return
			}
		})
	}
}

func Test_applicationCache_clear(t *testing.T) {
	type fields struct {
		cache map[string]*Application
	}
	appName := "Test"
	testApp := &Application{
		Id:   &uuid.UUID{},
		Name: &appName,
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "pass",
			fields: fields{cache: map[string]*Application{appName: testApp}},
		},
		{
			name:   "pass empty",
			fields: fields{cache: map[string]*Application{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &applicationCache{
				cache: tt.fields.cache,
			}
			ac.clear()
			if !ac.isEmpty() {
				t.Errorf("cache must be empty")
				return
			}
		})
	}
}

func Test_applicationCache_contains(t *testing.T) {
	type fields struct {
		cache map[string]*Application
	}
	type args struct {
		name string
	}
	appName := "Test"
	testApp := &Application{
		Id:   &uuid.UUID{},
		Name: &appName,
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "pass true",
			fields: fields{cache: map[string]*Application{appName: testApp}},
			args:   args{name: appName},
			want:   true,
		},
		{
			name:   "pass false",
			fields: fields{cache: map[string]*Application{appName: testApp}},
			args:   args{name: "bogus"},
			want:   false,
		},
		{
			name:   "pass empty false",
			fields: fields{cache: map[string]*Application{}},
			args:   args{name: appName},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &applicationCache{
				cache: tt.fields.cache,
			}
			if got := ac.contains(tt.args.name); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_applicationCache_get(t *testing.T) {
	type fields struct {
		cache map[string]*Application
	}
	type args struct {
		name string
	}
	appName := "Test"
	testApp := &Application{
		Id:   &uuid.UUID{},
		Name: &appName,
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Application
	}{
		{
			name:   "pass",
			fields: fields{cache: map[string]*Application{appName: testApp}},
			args:   args{name: appName},
			want:   testApp,
		},
		{
			name:   "pass nil",
			fields: fields{cache: map[string]*Application{appName: testApp}},
			args:   args{name: "bogus"},
			want:   nil,
		}, {
			name:   "pass empty nil",
			fields: fields{cache: map[string]*Application{}},
			args:   args{name: appName},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &applicationCache{
				cache: tt.fields.cache,
			}
			if got := ac.get(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_applicationCache_getAll(t *testing.T) {
	type fields struct {
		cache map[string]*Application
	}
	idUUID, _ := uuid.NewRandom()
	appName1 := "Test1"
	testApp1 := &Application{
		Id:   &idUUID,
		Name: &appName1,
	}
	appName2 := "Test2"
	testApp2 := &Application{
		Id:   &idUUID,
		Name: &appName2,
	}
	appName3 := "Test3"
	testApp3 := &Application{
		Id:   &idUUID,
		Name: &appName3,
	}
	cacheOne := map[string]*Application{appName1: testApp1}
	cacheTwo := map[string]*Application{appName2: testApp2, appName3: testApp3}
	appsOne := []*Application{testApp1}
	appsTwo := []*Application{testApp2, testApp3}

	tests := []struct {
		name   string
		fields fields
		want   []*Application
	}{
		{
			name:   "pass one",
			fields: fields{cache: cacheOne},
			want:   appsOne,
		},
		{
			name:   "pass two",
			fields: fields{cache: cacheTwo},
			want:   appsTwo,
		},
		{
			name:   "pass empty",
			fields: fields{cache: map[string]*Application{}},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &applicationCache{
				cache: tt.fields.cache,
			}
			if got := ac.getAll(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_applicationCache_isEmpty(t *testing.T) {
	type fields struct {
		cache map[string]*Application
	}
	appName := "Test"
	testApp := &Application{
		Id:   &uuid.UUID{},
		Name: &appName,
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "pass false",
			fields: fields{cache: map[string]*Application{appName: testApp}},
			want:   false,
		},
		{
			name:   "pass true",
			fields: fields{cache: map[string]*Application{}},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &applicationCache{
				cache: tt.fields.cache,
			}
			if got := ac.isEmpty(); got != tt.want {
				t.Errorf("isEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}