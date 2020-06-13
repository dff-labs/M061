package initialization

import (
	"time"

	"github.com/mzahmi/ventilator/control/adc"
	"github.com/mzahmi/ventilator/control/dac"
	"github.com/mzahmi/ventilator/control/ioexp"
	"github.com/mzahmi/ventilator/control/rpigpio"
	log "github.com/sirupsen/logrus"
)

//InitHardware ... this function should be called at the beginning of main.
//It will initialize all the hardware and check for errors
func HardwareInit() {
	log.Info("Beginning hardware initialization")

	//I2C init
	initI2C()
	//init ADC
	initADC()
	//init DAC
	initDAC()

	log.Info("End of hardware initialization")

}

//initI2C ... this function initializes I2C and checks for errors
//initI2C ... this function initializes I2C and checks for errors
func initI2C() {
	log.Info("Starting I2C init")
	err := ioexp.InitChip()
	if err != nil {
		log.Error(err)
		return
	}
	//Beep test
	log.Debug("Beep called")
	for ii := 0; ii < 3; ii++ {

		err = rpigpio.BeepOn()
		if err != nil {
			log.Error(err)
			return
		}
		time.Sleep(50 * time.Millisecond)
		err = rpigpio.BeepOff()
		if err != nil {
			log.Error(err)
			return
		}

		time.Sleep(50 * time.Millisecond)
	}

	//testing LEDs
	const blinkTime = 200
	log.Debug("Blinking LEDs")
	for ii := 0; ii < 2; ii++ {

		log.Debug("Yellow")
		err := ioexp.WritePin(ioexp.YellowLed, true)
		if err != nil {
			log.Error(err)
			return
		}
		time.Sleep(blinkTime * time.Millisecond)
		err = ioexp.WritePin(ioexp.YellowLed, false)
		if err != nil {
			log.Error(err)
			return
		}

		log.Debug("Red")
		err = ioexp.WritePin(ioexp.RedLed, true)
		if err != nil {
			log.Error(err)
			return
		}
		time.Sleep(blinkTime * time.Millisecond)
		err = ioexp.WritePin(ioexp.RedLed, false)
		if err != nil {
			log.Error(err)
			return
		}

		log.Debug("Green")
		err = ioexp.WritePin(ioexp.GreenLed, true)
		if err != nil {
			log.Error(err)
			return
		}
		time.Sleep(blinkTime * time.Millisecond)
		err = ioexp.WritePin(ioexp.GreenLed, false)
		if err != nil {
			log.Error(err)
			return
		}

		log.Debug("Blue")
		err = ioexp.WritePin(ioexp.BlueLed, true)
		if err != nil {
			log.Error(err)
			return
		}
		time.Sleep(blinkTime * time.Millisecond)
		err = ioexp.WritePin(ioexp.BlueLed, false)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

//initADC ... this function initializes ADC and checks for errors
func initADC() {

	log.Debug("readAdc called")

	_, err := adc.ReadADC(1)
	if err != nil {
		log.Error(err)
		return
	}

}

//initDAC ... this function initializes DAC and checks for errors
func initDAC() {

	log.Debug("dacsZero called")
	err := dac.DacsAllZeroOut()
	if err != nil {
		log.Error(err)
	}

}
