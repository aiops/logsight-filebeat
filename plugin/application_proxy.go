package plugin

// ApplicationApiCacheProxy Proxy pattern to implement caching for ApplicationApi
// See https://refactoring.guru/design-patterns/proxy
type ApplicationApiCacheProxy struct {
	ApplicationApiInterface

	applicationAPI   *ApplicationApi
	applicationCache applicationCache
}

func (cap *ApplicationApiCacheProxy) ClearCache() {
	cap.applicationCache.clear()
}

func (cap *ApplicationApiCacheProxy) GetApplications() ([]*Application, error) {
	if cap.applicationCache.isEmpty() {
		applications, err := cap.applicationAPI.GetApplications()
		if err != nil {
			return nil, err
		}
		cap.applicationCache.addAll(applications)
		return applications, nil
	} else {
		return cap.applicationCache.getAll(), nil
	}
}

func (cap *ApplicationApiCacheProxy) CreateApplication(name string) (*Application, error) {
	if cap.applicationCache.contains(name) {
		return cap.applicationCache.get(name), nil
	} else {
		application, err := cap.applicationAPI.CreateApplication(name)
		if err != nil {
			return nil, err
		}
		cap.applicationCache.add(application)
		return application, nil
	}
}
