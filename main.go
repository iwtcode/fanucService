package fanucService

import (
	"encoding/json"
	"fmt"
	"log"

	fanuc "github.com/iwtcode/fanucAdapter"
)

func main() {
	// 1. Настройка конфигурации
	cfg := &fanuc.Config{
		IP:          "192.168.56.1",
		Port:        8193,
		TimeoutMs:   5000,
		ModelSeries: "0i",
		LogPath:     "./focas.log",
	}

	fmt.Printf("Подключение к %s:%d...\n", cfg.IP, cfg.Port)

	// 2. Создание клиента
	client, err := fanuc.New(cfg)
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer client.Close()

	fmt.Println("Успешное подключение!")

	// 3. Тест GetSystemInfo
	sysInfo := client.GetSystemInfo()
	logAsJSON("System Info", sysInfo)

	// 4. Тест GetMachineState
	state, err := client.GetMachineState()
	if err != nil {
		log.Printf("Ошибка получения MachineState: %v", err)
	} else {
		logAsJSON("Machine State", state)
	}

	// 5. Тест GetAlarms
	alarms, err := client.GetAlarms()
	if err != nil {
		log.Printf("Ошибка получения Alarms: %v", err)
	} else {
		logAsJSON("Alarms", alarms)
	}

	// 6. Тест GetAxisData
	axes, err := client.GetAxisData()
	if err != nil {
		log.Printf("Ошибка получения AxisData: %v", err)
	} else {
		logAsJSON("Axis Data", axes)
	}

	fmt.Println("Тест завершен.")
}

// Вспомогательная функция для вывода в формате JSON
func logAsJSON(name string, data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("Ошибка маршалинга JSON для %s: %v", name, err)
		return
	}
	fmt.Printf("--- %s ---\n%s\n\n", name, string(jsonData))
}
