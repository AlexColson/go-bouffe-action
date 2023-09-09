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
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		// line will have a format like
		// ASNG/W+  0.00  kg
		// log.Println(line)
		stable := line[1] == 'S'
		words := strings.Split(line[4:], " ")
		weight, err := strconv.ParseFloat(words[2], 64)
		if err != nil {
			log.Fatal("Unable to parse string scale reading:" + line)
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
func InitSerial(baudRate int) (io.ReadCloser, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}

	if len(ports) == 0 {
		fmt.Println("No available serial ports found.")
		return nil, nil
	} else {
		for _, port := range ports {
			fmt.Printf("Found port: %s\n", port.Name)
			if port.IsUSB {
				fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
				fmt.Printf("   USB serial %s\n", port.SerialNumber)
			}
		}
	}

	// Select the first found serial port
	selectedPort := ports[0].Name
	log.Println("Using port " + selectedPort)
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
	return serialPort, nil
}
