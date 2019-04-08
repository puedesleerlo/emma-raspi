package main

import (
	"log"
	"github.com/sacOO7/gowebsocket"
	"os"
	"fmt"
	"time"
	"os/signal"
	"encoding/json"
	// "github.com/jacobsa/go-serial/serial"

	// "github.com/kraman/go-firmata"
	// gobot"gobot.io/x/gobot"
	// aio"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/platforms/firmata"
)

type Message struct {
    Sender    string `json:"sender,omitempty"`
    Recipient string `json:"recipient,omitempty"`
    Type     string `json:"type,omitempty"`
    Content   map[string]interface{} `json:"content,omitempty"`
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)


	socket := gowebsocket.New("ws://emmago.hopto.org/programs/ws")
	analogChan := ReadSerialFirmata()			
	
	SocketConfig(&socket)

	socket.Connect()
	go func() {
		for {
			select {
			case v:=<-analogChan:
					message := PrepareMessage(v)
					openMessage, _ := json.Marshal(&message)
  				socket.SendBinary(openMessage)
			}
		}
	}()
	

	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			// socket.Close()
			// port.Close()
			return
		}
	}
}

func SocketConfig(socket *gowebsocket.Socket) {
	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Fatal("Received connect error - ", err)
	}
  
	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server");
	}
  
	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		log.Println("Received message - " + message)
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
		return
	}
}

func PrepareMessage(v int) Message {
	content := map[string]interface{}{}
	variables := make([]map[string]interface{}, 0)
	x :=  map[string]interface{}{}
	x["name"] = "x"
	x["maxValue"] = 1023
	x["minValue"] = 0
	x["value"] = v
	variables = append(variables, x)
	content["variables"] = variables
	message := Message{Type:"sensor", Content: content,}
	return message
}

func ReadSerialFirmata() chan int {
	firmataAdaptor := firmata.NewAdaptor("COM4")
	firmataAdaptor.Connect()
	ticker := time.NewTicker(300 * time.Millisecond)
	analogread := make(chan int)
	value :=0
	rangeValue := 50
	maxValue := 50
	minValue := -50
	go func() {
		for {
			select {
			case <-ticker.C:
				red, _ := firmataAdaptor.AnalogRead("0")
				fmt.Println(red)
				maxValue = value + rangeValue
				minValue = value - rangeValue
				if !(red<=maxValue && red >= minValue) {
					analogread <- red
					fmt.Println(red)
				}	
				value = red
			}
		}
	}()

	return analogread
}