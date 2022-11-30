package fnutils

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

func LoadConfig(filename string) string {
	m := make(map[interface{}]interface{})
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, m)
	str := fmt.Sprintf("%v", string(yamlFile))
	return str
}
