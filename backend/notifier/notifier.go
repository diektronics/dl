package notifier

import (
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"

	"diektronics.com/carter/dl/protos/cfg"
	dlpb "diektronics.com/carter/dl/protos/dl"
)

type Client struct {
	addr      string
	port      int32
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
	links := make([]string, 0, len(down.Links))
	for _, l := range down.Links {
		links = append(links, fmt.Sprintf("%s: %s", l.Url, l.Status))
	}
	body := fmt.Sprintf("Name: %s\nStatus: %s\nErrors:\n\t%s\nPosthook: %s\nDestination: %s\nLinks:\n\t%s\n",
		down.Name, down.Status, strings.Join(down.Errors, "\n\t"), strings.Join(down.Posthook, ", "),
		down.Destination, strings.Join(links, "\n\t"))
	content := []byte(header + body)

	addrPort := net.JoinHostPort(n.addr, fmt.Sprint(n.port))

	auth := smtp.PlainAuth("", n.sender, n.password, n.addr)
	to := []string{n.recipient}
	if err := smtp.SendMail(addrPort, auth, n.sender, to, content); err != nil {
		log.Println("err: ", err)
	}
}
