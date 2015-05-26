package notifier

import (
	"fmt"
	"log"
	"net/smtp"

	"diektronics.com/carter/dl/protos/cfg"
	dlpb "diektronics.com/carter/dl/protos/dl"
)

type Client struct {
	addr      string
	port      string
	recipient string
	sender    string
	password  string
}

func New(c *cfg.Configuration) *Client {
	return &Client{
		addr:      c.MailAddr,
		port:      c.MailPort,
		recipient: c.MailRecipient,
		sender:    c.MailSender,
		password:  c.MailPassword,
	}
}

func (n Client) Notify(down *dlpb.Down) {
	header := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s %v\n\n",
		n.sender, n.recipient, down.Name, down.Status.String())
	content := []byte(header + down.String())

	addrPort := n.addr
	if n.port != "" {
		addrPort += ":" + n.port
	}

	auth := smtp.PlainAuth("", n.sender, n.password, n.addr)
	to := []string{n.recipient}
	if err := smtp.SendMail(addrPort, auth, n.sender, to, content); err != nil {
		log.Println("err: ", err)
	}
}
