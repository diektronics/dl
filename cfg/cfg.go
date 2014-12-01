package cfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Configuration struct {
	DbUser        string
	DbServer      string
	DbPassword    string
	DbDatabase    string
	MailAddr      string
	MailPort      string
	MailRecipient string
	MailSender    string
	MailPassword  string
	DownloadDir   string
	PlowdownPath  string
	HTTPPort      int
	BackendPort   int
}

func GetConfig(cfgFile string) (*Configuration, error) {
	cfg, err := os.Open(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("GetConfig: %v", err)
	}
	decoder := json.NewDecoder(cfg)
	c := &Configuration{}
	if err := decoder.Decode(c); err != nil {
		return nil, fmt.Errorf("GetConfig: %v", err)
	}

	if err := validate(c); err != nil {
		return nil, fmt.Errorf("GetConfig: Invalid configuration file: %v", err)
	}

	return c, nil
}

func validate(c *Configuration) error {
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
