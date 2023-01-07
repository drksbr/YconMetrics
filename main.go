package yconmetrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

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
	logData := map[string]string{
		"topic":       pl.topic,
		"value":       pl.value,
		"env":         pl.logger.env,
		"machineid":   pl.logger.machineId,
		"servicename": pl.logger.serviceName,
	}

	logDataBytes, err := json.Marshal(logData)
	if err != nil {
		pl.err = err
		return pl
	}

	resp, err := http.Post(pl.logger.url, "application/json", bytes.NewBuffer(logDataBytes))
	if err != nil {
		pl.err = err
		return pl
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		pl.err = fmt.Errorf("erro ao enviar log: código de status %d", resp.StatusCode)
	}
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
