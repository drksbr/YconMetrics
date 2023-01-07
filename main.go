package yconmetrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Logger struct {
	url         string
	env         string
	machineId   string
	serviceName string
}

func (l *Logger) topic(topic string) *ValueLogger {
	return &ValueLogger{
		topic:  topic,
		logger: l,
	}
}

type ValueLogger struct {
	topic  string
	value  string
	logger *Logger
}

func (vl *ValueLogger) setvalue(value string) *PrintLogger {
	return &PrintLogger{
		topic:  vl.topic,
		value:  value,
		logger: vl.logger,
	}
}

type PrintLogger struct {
	topic  string
	value  string
	err    error
	logger *Logger
}

func (pl *PrintLogger) print() {
	if pl.err != nil {
		fmt.Println("Erro ao enviar log para o servidor:", pl.err)
	} else {
		fmt.Println(pl.topic, pl.value)
	}
}

func (pl *PrintLogger) send() *PrintLogger {
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

func Config(url, env, machineId, serviceName string) *Logger {
	return &Logger{
		url:         url,
		env:         env,
		machineId:   machineId,
		serviceName: serviceName,
	}
}
