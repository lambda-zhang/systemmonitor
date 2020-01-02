package systemmonitor

import (
	"io/ioutil"
	"strconv"
	"strings"
)

// ThermalStates 温度状态
type ThermalStates struct {
	Type string
	Temp int64
}

// Thermalinfo 温度信息
type Thermalinfo struct {
	Thermal map[string]*ThermalStates
}

func (tinfo *Thermalinfo) getThermal() error {
	if tinfo.Thermal == nil {
		tinfo.Thermal = make(map[string]*ThermalStates)
	}

	dir, err := ioutil.ReadDir("/sys/class/thermal")
	if err != nil {
		return err
	}

	for _, fi := range dir {
		idx := strings.Index(fi.Name(), "thermal_zone")
		dirname := fi.Name()
		dirnamelen := strings.Count(dirname, "") - 1
		if idx < 0 || dirnamelen < (strings.Count("thermal_zonex", "")-1) {
			continue
		}
		ttype, err2 := readFile2String("/sys/class/thermal/" + fi.Name() + "/type")
		if err2 != nil || strings.Count(ttype, "") < 1 {
			continue
		}

		name := ttype + dirname[dirnamelen-1:]
		tempVal, err3 := readFile2String("/sys/class/thermal/" + fi.Name() + "/temp")
		if err3 != nil || strings.Count(tempVal, "") < 1 {
			continue
		}

		if _, ok := tinfo.Thermal[name]; !ok {
			tinfo.Thermal[name] = &ThermalStates{}
			tinfo.Thermal[name].Type = name
		}
		tinfo.Thermal[name].Temp, _ = strconv.ParseInt(tempVal, 10, 32)
	}

	return nil
}
