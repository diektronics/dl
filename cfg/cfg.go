package cfg

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	cfgpb "diektronics.com/carter/dl/protos/cfg"
	"github.com/golang/protobuf/proto"
)

func GetConfig(cfgFile string) (*cfgpb.Config, error) {
	content, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("GetConfig Readfile: %v", err)
	}
	c := &cfgpb.Config{}
	if err := proto.UnmarshalText(string(content), c); err != nil {
		return nil, fmt.Errorf("GetConfig Unmarshal: %v", err)
	}

	if err := validate(c); err != nil {
		return nil, fmt.Errorf("GetConfig: Invalid configuration file: %v", err)
	}

	return c, nil
}

func validate(c *cfgpb.Config) error {
	allErrors := []string{}

	cv := reflect.ValueOf(*c)
	ct := reflect.TypeOf(*c)
	for i := 0; i < cv.NumField(); i++ {
		if cv.Field(i).Kind() == reflect.Struct {
			s := cv.Field(i)
			st := cv.Field(i).Type()
			for j := 0; j < s.NumField(); j++ {
				if s.Field(j).Kind() == reflect.String && len(s.Field(j).String()) == 0 {
					allErrors = append(allErrors, fmt.Sprintf("%s.%s cannot be empty", ct.Field(i).Name, st.Field(i).Name))
				}
			}
		}
	}

	if stat, err := os.Stat(c.Download.PlowdownPath); os.IsNotExist(err) {
		allErrors = append(allErrors, fmt.Sprintf("%q does not exist", c.Download.PlowdownPath))
	} else if stat.Mode().Perm()&0111 == 0 {
		allErrors = append(allErrors, fmt.Sprintf("%q is not executable", c.Download.PlowdownPath))
	}

	if len(allErrors) != 0 {
		return errors.New(strings.Join(allErrors, ", "))
	}

	return nil
}
