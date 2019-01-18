package systemmonitor

import (
	"io/ioutil"
	"strconv"
	"strings"
)

type ThermalStates struct {
	Type string
	Temp int64
}

type Thermalinfo struct {
	Thermal map[string]*ThermalStates
}

func (this *Thermalinfo) GetThermal() error {
	if this.Thermal == nil {
		this.Thermal = make(map[string]*ThermalStates)
	}

	dir, err := ioutil.ReadDir("/sys/class/thermal")
	if err != nil {
		return err
	}

	for _, fi := range dir {
		idx := strings.Index(fi.Name(), "thermal_zone")
		dirname := fi.Name()
		dirname_len := strings.Count(dirname, "") - 1
		if idx < 0 || dirname_len < (strings.Count("thermal_zonex", "")-1) {
			continue
		}
		ttype, err2 := readFile2String("/sys/class/thermal/" + fi.Name() + "/type")
		if err2 != nil || strings.Count(ttype, "") < 1 {
			continue
		}

		name := ttype + dirname[dirname_len-1:]
		temp_val, err3 := readFile2String("/sys/class/thermal/" + fi.Name() + "/temp")
		if err3 != nil || strings.Count(temp_val, "") < 1 {
			continue
		}

		if _, ok := this.Thermal[name]; !ok {
			this.Thermal[name] = &ThermalStates{}
			this.Thermal[name].Type = name
		}
		this.Thermal[name].Temp, _ = strconv.ParseInt(temp_val, 10, 32)
	}

	return nil
}
