package cfg

import (
	"encoding/json"
	"fmt"
	"os"
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
}

func GetConfig(cfgFile string) (*Configuration, error) {
	cfg, err := os.Open(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("Open: %v", err)
	}
	decoder := json.NewDecoder(cfg)
	c := &Configuration{}
	if err := decoder.Decode(c); err != nil {
		return nil, fmt.Errorf("Decode: %v", err)
	}

	return c, nil
}
