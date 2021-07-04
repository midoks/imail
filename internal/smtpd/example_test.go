package smtpd_test

import (
	"errors"
	"net/smtp"
	"strings"

	"github.com/chrj/smtpd"
)

func ExampleServer() {
	var server *smtpd.Server

	// No-op server. Accepts and discards
	server = &smtpd.Server{}
	server.ListenAndServe("127.0.0.1:10025")

	// Relay server. Accepts only from single IP address and forwards using the Gmail smtp
	server = &smtpd.Server{

		HeloChecker: func(peer smtpd.Peer, name string) error {
			if !strings.HasPrefix(peer.Addr.String(), "42.42.42.42:") {
				return errors.New("Denied")
			}
			return nil
		},

		Handler: func(peer smtpd.Peer, env smtpd.Envelope) error {

			return smtp.SendMail(
				"smtp.gmail.com:587",
				smtp.PlainAuth(
					"",
					"username@gmail.com",
					"password",
					"smtp.gmail.com",
				),
				env.Sender,
				env.Recipients,
				env.Data,
			)

		},
	}

	server.ListenAndServe("127.0.0.1:10025")
}
