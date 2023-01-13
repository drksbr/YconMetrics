package WeLog

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
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
	Data    interface{} `json:"metadata"`
	Machine string      `json:"machineid"`
	Env     string      `json:"env"`
	Service string      `json:"servicename"`
}

type NetUsage struct {
	Input  uint64 `json:"input"`
	Output uint64 `json:"output"`
	Total  uint64 `json:"total"`
}

type DiskIOUsage struct {
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	Total      uint64 `json:"total"`
}

type DiskUsage struct {
	HddUsage  float64     `json:"hdd_usage"`
	SwapUsage float64     `json:"swap_usage"`
	IOUsage   DiskIOUsage `json:"io_usage"`
}

type HostInfoData struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Kernel   string `json:"kernel"`
	Platform string `json:"platform"`
	Uptime   string `json:"uptime"`
}

type Data struct {
	CPUusage    float64            `json:"cpu_usage"`
	MemoryUsage float64            `json:"memory_usage"`
	NetUsage    NetUsage           `json:"net_usage"`
	DiskUsage   DiskUsage          `json:"disk_usage"`
	HostTemp    map[string]float64 `json:"host_temp"`
	HostInfo    HostInfoData       `json:"host_info"`
}

func getSystemData(wait time.Duration) *Data {
	// Cria uma nova instância de Data
	data := &Data{}

	// Inicializa as variáveis de rede
	var input1, output1, total1 uint64 = 0, 0, 0
	var input2, output2, total2 uint64 = 0, 0, 0

	// Inicializa as variáveis de disco IO
	var ioread1, iowrite1, iototal1 uint64 = 0, 0, 0
	var ioread2, iowrite2, iototal2 uint64 = 0, 0, 0

	// Inicializa as variáveis de CPU
	var cputotal float64

	// Pega os dados do sistema
	m, _ := mem.VirtualMemory()
	c, _ := cpu.Percent(0, false)
	d, _ := disk.Usage("/")
	t, _ := host.SensorsTemperatures()
	netIOs, _ := net.IOCounters(true)

	// Preenche as variáveis de rede primeiro ciclo
	for _, netIO := range netIOs {
		input1 += netIO.BytesRecv
		output1 += netIO.BytesSent
		total1 += netIO.BytesRecv + netIO.BytesSent
	}

	// Preenche as variáveis de disco IO primeiro ciclo
	diskIOCounters, _ := disk.IOCounters()
	ioread1 = diskIOCounters["sda"].ReadBytes
	iowrite1 = diskIOCounters["sda"].WriteBytes
	iototal1 = diskIOCounters["sda"].ReadBytes + diskIOCounters["sda"].WriteBytes

	time.Sleep(wait * time.Second)

	// Preenche as variáveis de rede segundo ciclo
	netIOs2, _ := net.IOCounters(true)
	for _, netIO := range netIOs2 {
		input2 += netIO.BytesRecv
		output2 += netIO.BytesSent
		total2 += netIO.BytesRecv + netIO.BytesSent
	}

	diskIOCounters2, _ := disk.IOCounters()
	ioread2 = diskIOCounters2["sda"].ReadBytes
	iowrite2 = diskIOCounters2["sda"].WriteBytes
	iototal2 = diskIOCounters2["sda"].ReadBytes + diskIOCounters["sda"].WriteBytes

	// Preenche as variáveis de CPU
	for _, cpuPercent := range c {
		cputotal += cpuPercent
	}

	// Preenche as variáveis de disco
	data.DiskUsage.HddUsage = float64(d.UsedPercent)

	// calc swap usage in percent
	swap, _ := mem.SwapMemory()
	data.DiskUsage.SwapUsage = float64(swap.UsedPercent)

	// Preenche as variáveis de disco IO
	data.DiskUsage.IOUsage.ReadBytes = ioread2 - ioread1
	data.DiskUsage.IOUsage.WriteBytes = iowrite2 - iowrite1
	data.DiskUsage.IOUsage.Total = iototal2 - iototal1

	// Preenche as variáveis de memória, CPU e rede
	data.MemoryUsage = float64(m.UsedPercent)
	data.CPUusage = cputotal
	data.NetUsage.Input = (input2 - input1)
	data.NetUsage.Output = (output2 - output1)
	data.NetUsage.Total = (total2 - total1)

	// Preenche as variáveis de temperatura
	var tempMap = make(map[string]float64)
	for _, temp := range t {
		tempMap[temp.SensorKey] = temp.Temperature
	}
	data.HostTemp = tempMap

	// Preenche as variáveis de host
	hostInfo, _ := host.Info()
	duration := time.Duration(hostInfo.Uptime)

	data.HostInfo.Hostname = hostInfo.Hostname
	data.HostInfo.Kernel = hostInfo.KernelVersion
	data.HostInfo.Platform = hostInfo.Platform
	data.HostInfo.OS = hostInfo.OS
	data.HostInfo.Arch = hostInfo.KernelArch
	data.HostInfo.Uptime = convertToString(int(duration))

	return data
}

func convertToString(seconds int) string {
	var timeString string
	// Calculate years
	years := seconds / 31536000
	if years > 0 {
		timeString += fmt.Sprintf("%da ", years)
	}

	// Calculate months
	seconds -= years * 31536000
	months := seconds / 2592000
	if months > 0 {
		timeString += fmt.Sprintf("%dm ", months)
	}

	// Calculate days
	seconds -= months * 2592000
	days := seconds / 86400
	if days > 0 {
		timeString += fmt.Sprintf("%dd ", days)
	}

	// Calculate hours
	seconds -= days * 86400
	hours := seconds / 3600
	if hours > 0 {
		timeString += fmt.Sprintf("%dh ", hours)
	}

	// Calculate minutes
	seconds -= hours * 3600
	minutes := seconds / 60
	if minutes > 0 {
		timeString += fmt.Sprintf("%dm ", minutes)
	}

	// Calculate seconds
	seconds -= minutes * 60
	if seconds > 0 {
		timeString += fmt.Sprintf("%ds ", seconds)
	}

	return timeString
}

// getMachineHash returns a reduced SHA-256 hash of the machine ID and hostname.
func getMachineHash() (string, error) {
	// Get the hostname
	hostInfo, err := host.Info()
	if err != nil {
		return "", err
	}

	// Get the hostname and machine ID
	hostname := hostInfo.Hostname
	machineID := hostInfo.HostID

	// Define a salt string
	salt := "yconapp-2022"

	// Concatenate the machine ID and hostname
	data := []byte(machineID + hostname + salt)

	// Create a SHA-256 hash of the data
	hash := sha256.New()
	io.WriteString(hash, string(data))
	hashString := fmt.Sprintf("%x", hash.Sum(nil))

	// Truncate the hash to 32 characters
	reducedHash := hashString[:32]

	return reducedHash + " - " + hostname, nil
}

// New cria uma nova instância da biblioteca de log
func New(apiURL, environment, serviceName string) *WeLog {

	machineID, err := getMachineHash()
	if err != nil {
		log.Default().Println("erro ao obter ID da máquina: ", err)
	}

	return &WeLog{
		APIURL:      apiURL,
		MachineID:   machineID,
		Environment: environment,
		ServiceName: serviceName,
	}
}

// Topic envia um log para a API
func (l *WeLog) Topic(topic string, data interface{}) error {
	// Converte os dados para JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("erro ao converter dados para JSON: %v", err)
	}

	// Converte a string para base64
	dataBase64 := base64.StdEncoding.EncodeToString([]byte(jsonData))

	// Cria um novo objeto LogData com os dados do log
	logData := &LogData{
		Topic:   topic,
		Data:    dataBase64,
		Machine: l.MachineID,
		Env:     l.Environment,
		Service: l.ServiceName,
	}

	// Converte o objeto LogData para JSON
	jsonLogData, err := json.Marshal(logData)
	if err != nil {
		return fmt.Errorf("erro ao converter dados para JSON: %v", err)
	}

	// Envia a solicitação HTTP POST para a API
	_, err = http.Post(l.APIURL, "application/json", bytes.NewBuffer(jsonLogData))
	if err != nil {
		return fmt.Errorf("erro ao enviar solicitação HTTP POST: %v", err)
	}

	return nil
}

func (l *WeLog) Resources(i time.Duration) error {
	res := getSystemData(i)
	l.Topic("resources", res)
	return nil
}

func (l *WeLog) ResourcesDaemon(i time.Duration) {
	for {
		l.Resources(i)
	}
}
