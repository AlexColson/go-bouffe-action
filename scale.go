package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	serial "go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

type ScaleReading struct {
	Weight float32
	Stable bool
}

// Function to read data from the serial port and send it to a channel
func RealScale(serialPort io.ReadCloser, dataChannel chan<- ScaleReading) {
	reader := bufio.NewReader(serialPort)
	for {
		rawline, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		// line will have a format like
		// ASNG/W+  0.00  kg
		// log.Println(line)
		line := strings.TrimSpace(rawline[:])
		stable := line[0] == 'S'
		words := strings.Fields(line)
		weight, err := strconv.ParseFloat(words[1], 64)
		if err != nil {
			log.Println("Unable to parse string scale reading:" + line)
			continue
		}
		reading := ScaleReading{float32(weight), stable}
		sendData(dataChannel, reading)
	}
}

func FakeScale(dataChannel chan<- ScaleReading) {
	for {
		weigth := rand.Float32()
		stable := rand.Int31n(10) > 3
		if !stable {
			weigth *= 0.3 + rand.Float32()
		}
		reading := ScaleReading{weigth, stable}
		sendData(dataChannel, reading)

		time.Sleep(1 * time.Second)
	}
}

// Send data in the channel if it's
func sendData(dataChannel chan<- ScaleReading, reading ScaleReading) {
	select {
	case dataChannel <- reading:
		// fmt.Println("Sent")
	default:
		// fmt.Println("Channel is full. Skipping value.")
	}
}

// Function to process data from the channel
func ReadScale(dataChannel <-chan ScaleReading) map[string]interface{} {

	data := <-dataChannel

	response := map[string]interface{}{
		"value":  data.Weight,
		"stable": data.Stable,
	}
	return response

}

// Function to initialize the serial port connection
func InitSerial(baudRate int, UsdScaleDeviceId string) (io.ReadCloser, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	selectedPort := ""
	if len(ports) == 0 {
		fmt.Println("ERREUR!: Aucun port serie detecte.")
		return nil, nil
	} else {
		for _, port := range ports {
			fmt.Printf("Found port: %s\n", port.Name)
			if port.IsUSB {
				fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
				fmt.Printf("   USB serial %s\n", port.SerialNumber)
			}
			if UsdScaleDeviceId == port.VID+":"+port.PID {
				selectedPort = port.Name
			}
		}
	}

	// Select the first found serial port
	if selectedPort != "" {
		log.Println("Utilitation du port " + selectedPort + " pour comminiquer avec la balance")
	} else {
		log.Println("ERREUR!:Impossible de trouver le port de la balance: '" + UsdScaleDeviceId + "'")
		log.Println("ERREUR!:Verifier que le cable est bien relie a la balance, ainsi que la configuration dans conf.toml.")
		fmt.Scanln()
		log.Panic()
	}
	options := serial.Mode{
		BaudRate: baudRate, // Set the baud rate
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	serialPort, err := serial.Open(selectedPort, &options)
	if err != nil {
		log.Fatal(err)
	}
	return serialPort, err
}
