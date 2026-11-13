package smtp

import (
	"context"
	"fmt"
	"mailer/internal/mailer/app"
	"net/smtp"
	"strings"
)

type sender struct{ address, from string }

func New(address, from string) app.Sender { return &sender{address, from} }

func (s *sender) Send(ctx context.Context, to, subject, body string) error {
	if strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("invalid subject")
	}
	done := make(chan error, 1)
	go func() {
		msg := []byte("From: " + s.from + "\r\nTo: " + to + "\r\nSubject: " + subject + "\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body)
		done <- smtp.SendMail(s.address, nil, s.fromAddress(), []string{to}, msg)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-done:
		return e
	}
}

func (s *sender) fromAddress() string {
	v := s.from
	if i := strings.Index(v, "<"); i >= 0 && strings.HasSuffix(v, ">") {
		return v[i+1 : len(v)-1]
	}
	return v
}
