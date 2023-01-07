package yconmetrics

import "fmt"

type logger struct {
	url         string
	env         string
	machineId   string
	serviceName string
}

func (l *logger) topic(topic string) *valueLogger {
	return &valueLogger{
		topic:  topic,
		logger: l,
	}
}

type valueLogger struct {
	topic  string
	value  string
	logger *logger
}

func (vl *valueLogger) setvalue(value string) *printLogger {
	return &printLogger{
		topic:  vl.topic,
		value:  value,
		logger: vl.logger,
	}
}

type printLogger struct {
	topic  string
	value  string
	err    error
	logger *logger
}

func (pl *printLogger) print() {
	if pl.err != nil {
		fmt.Println("Erro ao enviar log para o servidor:", pl.err)
	} else {
		fmt.Println(pl.topic, pl.value)
	}
}

func (pl *printLogger) send() *printLogger {
	// aqui você pode enviar o log para o servidor usando o método POST para a url do logger
	// o conteúdo do log pode ser enviado como um JSON no corpo da requisição, por exemplo:
	// {
	//   "topic": pl.topic,
	//   "value": pl.value,
	//   "env": pl.logger.env,
	//   "machineid": pl.logger.machineId,
	//   "servicename: pl.logger.serviceName,
	// }
	pl.err = nil
	return pl
}

func Config(url, env, machineId, serviceName string) *logger {
	return &logger{
		url:         url,
		env:         env,
		machineId:   machineId,
		serviceName: serviceName,
	}
}
