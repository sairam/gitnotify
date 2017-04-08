package gitnotify

import (
	"errors"
	"fmt"
	"github.com/stathat/go"
)

func statCount(key string) {
	stathat.PostEZCount(config.getStatHatPrefix()+key, config.StatHatKey, 1)
}

func statValue(key string, value interface{}) {
	val, err := interfaceStatValue(value)
	if err != nil {
		fmt.Println("Error writing key ", key, " value could not be assessed")
		return
	}
	stathat.PostEZValue(config.getStatHatPrefix()+key, config.StatHatKey, val)
}

func interfaceStatValue(value interface{}) (float64, error) {
	if v_flt, ok := value.(float64); ok {
		return v_flt, nil
	} else if v_int, ok := value.(int64); ok {
		return float64(v_int), nil
	} else {
		return 0.0, errors.New("Could not find type")
	}

}
