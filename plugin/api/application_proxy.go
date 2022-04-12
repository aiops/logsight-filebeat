package api

// applicationApiCacheProxy Proxy pattern to implement caching for ApplicationApi
// See https://refactoring.guru/design-patterns/proxy
type applicationApiCacheProxy struct {
	ApplicationApiInterface

	applicationAPI   *ApplicationApi
	applicationCache *applicationCache
}

func NewApplicationApiCacheProxy(applicationAPI *ApplicationApi) *applicationApiCacheProxy {
	return &applicationApiCacheProxy{applicationAPI: applicationAPI, applicationCache: NewApplicationCache()}
}

func (cap *applicationApiCacheProxy) ClearCache() {
	cap.applicationCache.clear()
}

func (cap *applicationApiCacheProxy) GetApplications() ([]*Application, error) {
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

func (cap *applicationApiCacheProxy) GetApplicationByName(name string) (*Application, error) {
	if cap.applicationCache.contains(name) {
		return cap.applicationCache.get(name), nil
	} else {
		cap.ClearCache()
		if _, err := cap.GetApplications(); err != nil {
			return nil, err
		}
		if cap.applicationCache.contains(name) {
			return cap.applicationCache.get(name), nil
		} else {
			return nil, nil
		}
	}
}

func (cap *applicationApiCacheProxy) CreateApplication(req CreateApplicationRequest) (*Application, error) {
	if application, err := cap.GetApplicationByName(req.Name); err != nil {
		return nil, err
	} else if application != nil {
		return application, nil
	} else {
		application, err = cap.applicationAPI.CreateApplication(req)
		if err != nil {
			return nil, err
		}
		cap.applicationCache.add(application)
		return application, nil
	}
}
