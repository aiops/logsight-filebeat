package api

type LogSender struct {
	LogApi *LogApi
}

func (as LogSender) Send(logs []*Log) error {
	if _, err := as.LogApi.SendLogs(logs); err != nil {
		return err
	}
	return nil
}

func (as LogSender) Close() {
	as.LogApi.HttpClient.CloseIdleConnections()
}
