package WeLog

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

// WeLog é a estrutura da nossa biblioteca de log
type WeLog struct {
	APIURL      string
	MachineID   string
	Environment string
	ServiceName string
}

// LogData é a estrutura dos dados do log que serão enviados para a API
type LogData struct {
	Topic   string      `json:"topic"`
	Data    interface{} `json:"data"`
	Machine string      `json:"machine"`
	Env     string      `json:"env"`
	Service string      `json:"service"`
}

// New cria uma nova instância da biblioteca de log
func New(apiURL, machineID, environment, serviceName string) *WeLog {
	return &WeLog{
		APIURL:      apiURL,
		MachineID:   machineID,
		Environment: environment,
		ServiceName: serviceName,
	}
}

// Topic envia um log para a API
func (l *WeLog) Topic(topic string, data string) error {
	// Converte os dados para base64
	dataBase64 := base64.StdEncoding.EncodeToString([]byte(data))

	// Cria um novo objeto LogData com os dados do log
	logData := &LogData{
		Topic:   topic,
		Data:    dataBase64,
		Machine: l.MachineID,
		Env:     l.Environment,
		Service: l.ServiceName,
	}

	// Converte o objeto LogData para JSON
	jsonData, err := json.Marshal(logData)
	if err != nil {
		return fmt.Errorf("erro ao converter dados para JSON: %v", err)
	}

	// Envia a solicitação HTTP POST para a API
	_, err = http.Post(l.APIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("erro ao enviar solicitação HTTP POST: %v", err)
	}

	return nil
}
