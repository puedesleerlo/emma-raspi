package main

import (
	"log"
	"github.com/sacOO7/gowebsocket"
	"os"
	"fmt"
	"time"
	"os/signal"
	// "encoding/json"
	// "github.com/jacobsa/go-serial/serial"

	// "github.com/kraman/go-firmata"
	// gobot"gobot.io/x/gobot"
	// aio"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/platforms/firmata"
)

import "flag"

var serialport = flag.String("serialport", "/dev/ttyUSB0", "please choose a serial port")
type Message struct {
    Sender    string `json:"sender,omitempty"`
    Recipient string `json:"recipient,omitempty"`
    Type     string `json:"type,omitempty"`
    Content   map[string]interface{} `json:"content,omitempty"`
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)


	// socket := gowebsocket.New("ws://emmago.hopto.org/programs/ws")
	firmataAdaptor := firmata.NewAdaptor(*serialport)
	firmataAdaptor.Connect()
	analogChan := ReadSerialFirmata(firmataAdaptor)			
	encoderRotationChan := ReadRotaryEncoder(firmataAdaptor)
	// SocketConfig(&socket)

	// socket.Connect()
	go func() {
		tick := time.NewTicker(10*time.Second)
		for {
			select {
				case <-analogChan:
					// message := PrepareMessage(v)
					// openMessage, _ := json.Marshal(&message)
					// fmt.Println("mensaje enviado")
					// socket.SendBinary(openMessage)
					break;
				case rotation:= <-encoderRotationChan:
					fmt.Println("el encoder es", rotation)
					// message := PrepareMessage(rotation)
					// openMessage, _ := json.Marshal(&message)
					// fmt.Println("mensaje enviado")
					// socket.SendBinary(openMessage)
					break;
				case <-tick.C:
					// ping := Message{Type:"ping"}
					// pingMsg, _ := json.Marshal(&ping)
					// socket.SendBinary(pingMsg)
					break;
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
		// log.Println("Received message - " + message)
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
		return
	}
}

func PrepareMessage(v int) Message {
	content := map[string]interface{}{}
	variables := map[int]interface{}{}
	variables[0] = v
	content["variables"] = variables
	message := Message{Type:"sensor", Content: content,}
	return message
}

func ReadSerialFirmata(firmataAdaptor *firmata.Adaptor) chan int {
	
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
				fmt.Println(*serialport)
				red, _ := firmataAdaptor.AnalogRead("0")
				fmt.Println(red)
				maxValue = value + rangeValue
				minValue = value - rangeValue
				if !(red<=maxValue && red >= minValue) {
					analogread <- red
				}	
				value = red
			}
		}
	}()

	return analogread
}

func ReadRotaryEncoder(firmata *firmata.Adaptor) chan int {
	ticker := time.NewTicker(30 * time.Millisecond)
	encoderRotationChan := make(chan int)
	encoderPosCount := 0
	pinA := "2"
	pinB := "3"
	pinALast,_ := firmata.DigitalRead(pinA); 
	go func() {
		for {
			select {
			case <-ticker.C:
				aVal,_ := firmata.DigitalRead(pinA);
				if (aVal != pinALast){ // Means the knob is rotating
					// if the knob is rotating, we need to determine direction
					// We do that by reading pin B.
				if readB,_ := firmata.DigitalRead(pinB); readB != aVal {  // Means pin A Changed first - We're Rotating Clockwise
					encoderPosCount ++;
				} else {// Otherwise B changed first and we're moving CCW
					encoderPosCount--;
				}
					if encoderPosCount > 360 {
						encoderPosCount = 360
					} else if encoderPosCount < 0 {
						encoderPosCount = 0
					}
					
					encoderRotationChan<-encoderPosCount
				} 
				pinALast = aVal;
				 
			}
		}
	}()

	return encoderRotationChan
}