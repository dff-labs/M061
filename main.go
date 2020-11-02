package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/mzahmi/ventilator/control/cli"
	"github.com/mzahmi/ventilator/control/initialization"
	"github.com/mzahmi/ventilator/control/modeselect"
	"github.com/mzahmi/ventilator/control/rpigpio"
	"github.com/mzahmi/ventilator/control/sensors"
	"github.com/mzahmi/ventilator/control/valves"
	"github.com/mzahmi/ventilator/logger"
	"github.com/mzahmi/ventilator/params"

	// "github.com/mzahmi/ventilator/control/alarms"
	log "github.com/sirupsen/logrus"
)

var UI = params.DefaultParams
var wg sync.WaitGroup
var mux sync.Mutex

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	// //Creates log file called Events.log
	// f, err := os.OpenFile("file.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	fmt.Println(errng)
	// }
	// defer f.Close()

	// // declare new loggers with different prefixes
	// logger := log.New(f, "Event ", log.LstdFlags) // event logger
	// logErr := log.New(f, "Error ", log.LstdFlags) // errors logger
	// logArm := log.New(f, "Alarm ", log.LstdFlags) // alarms logger

	// Initialize Logger
	logStruct := logger.LoggerInit()
	defer logger.LoggerClose()

	// Let's create some command-line arguments to avoid the default hard-coded entries
	redisAddressDefault := "dupl1.local:6379"

	redisAddressPtr := flag.String("redis-addr", redisAddressDefault, "a string formatted as hostname:port")
	redisPasswordPtr := flag.String("redis-password", "", "a string")
	redisDBPtr := flag.Int("redis-db", 0, "an int")
	// If you opt to use github.com/go-redis/v8 then, then you can enable username for auth
	// redisUsernamePtr := flag.String("redis-username", "", "a string")
	flag.Parse() // Flags can now be used in the program

	//initialize the hardware
	initialization.HardwareInit()
	logStruct.Event("Hardware Initialized")
	//logger.Println("Hardware Initialized")

	//establish connection with redis client
	log.Info(fmt.Sprintf("%s: %s (DB=%d)..", "Trying to connect to redis server: ", *redisAddressPtr, *redisAddressPtr))
	redis_client := redis.NewClient(&redis.Options{
		// Addr:     "dupi1.local:6379",
		Addr:     *redisAddressPtr,
		Password: *redisPasswordPtr,
		// Username: *redisUsernamePtr,
		DB: *redisDBPtr,
	})

	_, err := redis_client.Ping().Result()
	for {
		if err == nil {
			break
		}
		time.Sleep(1000 * time.Millisecond)
		_, err = redis_client.Ping().Result()
		log.Info("\r- Trying to connect to redis server: ", *redisAddressPtr)
	}

	check(err, logStruct)
	logStruct.Event("Redis Client Initialized")
	// logger.Println("Client Initialized")

	// set the critical records in redis to zero or NA
	redis_client.Set("status", "NA", 0).Err()
	redis_client.Set("pressure", 0, 0).Err()
	redis_client.Set("volume", 0, 0).Err()
	redis_client.Set("flow", 0, 0).Err()

	//initialize the user input parameters
	params.InitParams(redis_client)
	logStruct.Event("Parameters Initialized")
	//logger.Println("Parameters Initialized")

	//initialize a SensorsReading struct to store all of the sensor readings
	s := sensors.SensorsReading{
		PressureInput:  0,
		PressureOutput: 0,
		FlowInput:      0,
	}

	// Reads sensors and populate the graph
	// limit the reading frequency to a predefined value
	rate := float64(200)                                                   // Hz rate
	timePerLoopIteration := time.Duration(1000000/rate) * time.Microsecond //(1 / rate) us

	go func() {
		calCounter := 0
		var InitPin []float32
		var calP float32
		for {
			t1 := time.Now()
			// reads all of the sensors on the system
			for ; calCounter < 6; calCounter++ {
				P, _, _ := sensors.ReadAllSensors()
				InitPin = append(InitPin, P)
				calP = average(InitPin)
			}

			Pin, Pout, Fin := sensors.ReadAllSensors()
			//locks the populating of the sensors struct
			mux.Lock()
			s = sensors.SensorsReading{
				PressureInput:  (Pin - calP) * 1020,
				PressureOutput: Pout * 1020,
				FlowInput:      Fin * 100,
			}
			mux.Unlock()
			runtime.Gosched()
			//sends the pressure reading from Pin to GUI
			redis_client.Set("pressure", (Pin-calP)*1020/2, 0).Err()
			redis_client.Set("flow", (Fin)*100, 0).Err()
			//fmt.Println(Pin*1020)
			//calculates the delay based on a specified rate
			loopTime := time.Since(t1)
			if loopTime < timePerLoopIteration {
				diff := (timePerLoopIteration - loopTime)
				//fmt.Println("Sleeping for:", diff)
				time.Sleep(diff)
			}
			//t3 := time.Now()
			//fmt.Println("Tdiff=", t3.Sub(t1))
		}
	}()

	go func() {
		for {
			valves.MV.Open()
			valves.MExp.Open()
			valves.InProp.IncrementValve(1.0)
			time.Sleep(time.Millisecond * 2000)			
			for ii := 0; ii < 1; ii++ {
				err = rpigpio.BeepOn()
				check(err, logStruct)
				time.Sleep(100 * time.Millisecond)
				err = rpigpio.BeepOff()
				check(err, logStruct)
				time.Sleep(100 * time.Millisecond)
			}
			valves.MV.Close()
			valves.MExp.Close()
			valves.InProp.IncrementValve(0.0)
			valves.InProp.Close()
			time.Sleep(time.Millisecond * 2000)
		}
	}()

	// Provides CLI interface
	go cli.Run(&s, redis_client, &mux)

	//checks for sys interupt
	SetupCloseHandler()

	//Airway pressure alarm check
	// go alarms.AirwayPressureAlarms(&s, &mux, UI.UpperLimitPIP, UI.LowerLimitPIP, &logStruct, redis_client)

	go func() {
		for {
			mux.Lock()
			airpress := s.PressureInput
			mux.Unlock()
			runtime.Gosched()
			if (airpress) >= 60 {
				//msg := "Airway Pressure high"
				redis_client.Set("alarm_status", "critical", 0).Err()
				redis_client.Set("alarm_title", "Airway Pressure high", 0).Err()
				redis_client.Set("alarm_text", "Airway Pressure exceeded limits check for obstruction", 0).Err()
				for ii := 0; ii < 5; ii++ {
					err = rpigpio.BeepOn()
					check(err, logStruct)
					time.Sleep(100 * time.Millisecond)
					err = rpigpio.BeepOff()
					check(err, logStruct)
					time.Sleep(100 * time.Millisecond)
				}
			} else {
				redis_client.Set("alarm_status", "none", 0).Err()
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	for {
		status, err := redis_client.Get("status").Result()
		check(err, logStruct)
		if status == "start" {
			// logger.Printf("Ventilation status changed to %s\n", status)
			logStruct.Event(fmt.Sprintf("Ventilation status changed to %s\n", status))
			UI = params.ReadParams(redis_client)
			go modeselect.ModeSelection(&UI, &s, redis_client, &mux, &logStruct)
			redis_client.Set("status", "ventilating", 0).Err()
			// logger.Printf("Ventilation status changed to %s\n", status)
			logStruct.Event(fmt.Sprintf("Ventilation status changed to %s\n", status))
		} else if status == "stop" {
			// logger.Println("Stopping system")
			logStruct.Event("Stopping system")
			valves.CloseAllValves(&valves.MV, &valves.MExp, &valves.InProp)
			// redis_client.Set("status", "waiting", 0).Err()
		} else if status == "exit" {
			// logger.Println("Exiting system")
			logStruct.Event("Exiting system")
		}
	}
}

// prints out the checked error err
func check(err error, logStruct logger.Logging) {
	if err != nil {
		logStruct.Err(err)
	}
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("\r- Ctrl+C pressed in Terminal")
		valves.CloseAllValves(&valves.InProp, &valves.MExp, &valves.MV)
		os.Exit(0)
	}()
}

func average(xs []float32) float32 {
	var total float32
	for _, v := range xs {
		total += v
	}
	return total / float32(len(xs))
}
