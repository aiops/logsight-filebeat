package api

import "sort"

type applicationCache struct {
	cache map[string]*Application // cache maps application name to application object pointer
}

func NewApplicationCache() *applicationCache {
	return &applicationCache{cache: make(map[string]*Application)}
}

func (ac *applicationCache) isEmpty() bool {
	return len(ac.cache) == 0
}

func (ac *applicationCache) clear() {
	ac.cache = make(map[string]*Application)
}

func (ac *applicationCache) contains(name string) bool {
	_, present := ac.cache[name]
	return present
}

func (ac *applicationCache) get(name string) *Application {
	application, _ := ac.cache[name]
	return application
}

func (ac *applicationCache) getAll() []*Application {
	if ac.isEmpty() {
		return nil
	} else {
		applications := make([]*Application, len(ac.cache))
		i := 0
		for _, app := range ac.cache {
			applications[i] = app
			i++
		}
		sort.Slice(applications, func(i, j int) bool {
			return *applications[i].Name < *applications[j].Name
		})
		return applications
	}
}

func (ac *applicationCache) add(application *Application) {
	if application != nil && application.Name != nil {
		ac.cache[*application.Name] = application
	}
}

func (ac *applicationCache) addAll(applications []*Application) {
	for _, app := range applications {
		if app != nil {
			ac.add(app)
		}
	}
}
