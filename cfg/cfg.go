package cfg

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	cfgpb "diektronics.com/carter/dl/protos/cfg"
	"github.com/golang/protobuf/proto"
)

func GetConfig(cfgFile string) (*cfgpb.Configuration, error) {
	f, err := os.Open(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("getConfig: %v", err)
	}
	content := make([]byte, 1000)
	count, err := f.Read(content)
	if err != nil {
		return nil, fmt.Errorf("getConfig: %v", err)
	}
	if count == 1000 {
		return nil, fmt.Errorf("getConfig: read buffer is too small %v", count)
	}
	c := &cfgpb.Configuration{}
	if err := proto.UnmarshalText(string(content[:count]), c); err != nil {
		return nil, fmt.Errorf("getConfig: %v", err)
	}

	if err := validate(c); err != nil {
		return nil, fmt.Errorf("getConfig: Invalid configuration file: %v", err)
	}

	return c, nil
}

func validate(c *cfgpb.Configuration) error {
	allErrors := []string{}

	cv := reflect.ValueOf(*c)
	ct := reflect.TypeOf(*c)
	for i := 0; i < cv.NumField(); i++ {
		if cv.Field(i).Kind() == reflect.String && len(cv.Field(i).String()) == 0 {
			allErrors = append(allErrors, fmt.Sprintf("%q cannot be empty", ct.Field(i).Name))
		}
	}

	if stat, err := os.Stat(c.PlowdownPath); os.IsNotExist(err) {
		allErrors = append(allErrors, fmt.Sprintf("%q does not exist", c.PlowdownPath))
	} else if stat.Mode().Perm()&0111 == 0 {
		allErrors = append(allErrors, fmt.Sprintf("%q is not executable", c.PlowdownPath))
	}

	if len(allErrors) != 0 {
		return errors.New(strings.Join(allErrors, ", "))
	}

	return nil
}
