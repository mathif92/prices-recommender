package recommendations

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Notifier struct {
	server   string
	port     string
	user     string
	pass     string
	from     string
}

func NewNotifier(cfg Config) *Notifier {
	return &Notifier{
		server: cfg.SMTPServer,
		port:   cfg.SMTPPort,
		user:   cfg.SMTPUser,
		pass:   cfg.SMTPPass,
		from:   cfg.SMTPFrom,
	}
}

func (n *Notifier) SendPriceDropAlert(to string, drops []PriceDrop) error {
	if len(drops) == 0 {
		return nil
	}

	subject := fmt.Sprintf("Price Drop Alert: %d hotel%s just dropped!", len(drops), plural(len(drops)))

	var body strings.Builder
	body.WriteString("The following hotels have dropped in price:\n\n")
	for _, d := range drops {
		body.WriteString(fmt.Sprintf("  %s (%s)\n", d.HotelName, d.Location))
		body.WriteString(fmt.Sprintf("    Dates: %s - %s\n", d.StartDate.Format("Jan 2, 2006"), d.EndDate.Format("Jan 2, 2006")))
		body.WriteString(fmt.Sprintf("    Was: %.2f %s\n", d.OldPrice, d.Currency))
		body.WriteString(fmt.Sprintf("    Now: %.2f %s (%.0f%% off!)\n", d.NewPrice, d.Currency, d.DropRatio*100))
		body.WriteString("\n")
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		n.from, to, subject, body.String())

	addr := fmt.Sprintf("%s:%s", n.server, n.port)
	auth := smtp.PlainAuth("", n.user, n.pass, n.server)

	return smtp.SendMail(addr, auth, n.from, []string{to}, []byte(msg))
}

func (n *Notifier) SendPriceDropAlertIfConfigured(to string, drops []PriceDrop) error {
	if n.server == "" || to == "" {
		return nil
	}
	return n.SendPriceDropAlert(to, drops)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
