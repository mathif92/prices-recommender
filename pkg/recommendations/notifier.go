package recommendations

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/sirupsen/logrus"
)

type Notifier struct {
	log    logrus.FieldLogger
	server string
	port   string
	user   string
	pass   string
	from   string
}

func NewNotifier(log logrus.FieldLogger, cfg Config) *Notifier {
	n := &Notifier{
		log:    log,
		server: cfg.SMTPServer,
		port:   cfg.SMTPPort,
		user:   cfg.SMTPUser,
		pass:   cfg.SMTPPass,
		from:   cfg.SMTPFrom,
	}
	if n.server == "" {
		n.log.Warn("SMTP_SERVER not configured, email notifications disabled")
	} else {
		n.log.Infof("SMTP configured: %s:%s (from: %s)", n.server, n.port, n.from)
	}
	return n
}

func (n *Notifier) SendPriceDropAlert(to string, drops []PriceDrop) error {
	if len(drops) == 0 {
		return nil
	}

	subject := fmt.Sprintf("Price Drop Alert: %d hotel%s just dropped!", len(drops), plural(len(drops)))

	type dropView struct {
		HotelName   string
		Location    string
		StartDate   string
		EndDate     string
		OldPrice    string
		NewPrice    string
		Currency    string
		DiscountPct string
	}

	var items []dropView
	for _, d := range drops {
		items = append(items, dropView{
			HotelName:   d.HotelName,
			Location:    d.Location,
			StartDate:   d.StartDate.Format("Jan 2, 2006"),
			EndDate:     d.EndDate.Format("Jan 2, 2006"),
			OldPrice:    fmt.Sprintf("%.2f", d.OldPrice),
			NewPrice:    fmt.Sprintf("%.2f", d.NewPrice),
			Currency:    d.Currency,
			DiscountPct: fmt.Sprintf("%.0f", d.DropRatio*100),
		})
	}

	data := struct {
		DropCount string
		Drops     []dropView
	}{
		DropCount: fmt.Sprintf("%d", len(drops)),
		Drops:     items,
	}

	var body bytes.Buffer
	if err := emailTemplate.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		n.from, to, subject, body.String())

	addr := fmt.Sprintf("%s:%s", n.server, n.port)
	auth := smtp.PlainAuth("", n.user, n.pass, n.server)

	return smtp.SendMail(addr, auth, n.from, []string{to}, []byte(msg))
}

func (n *Notifier) SendPriceDropAlertIfConfigured(to string, drops []PriceDrop) error {
	if n.server == "" {
		n.log.Warn("SMTP not configured, skipping email notification")
		return nil
	}
	if to == "" {
		n.log.Warn("no recipient email configured, skipping email notification")
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

var emailTemplate = template.Must(template.New("email").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
  body { font-family: Arial, sans-serif; background: #f4f4f7; margin: 0; padding: 0; }
  .container { max-width: 600px; margin: 0 auto; background: #ffffff; }
  .header { background: #2563eb; padding: 32px; text-align: center; }
  .header h1 { color: #ffffff; margin: 0; font-size: 24px; }
  .subtitle { color: rgba(255,255,255,0.85); font-size: 14px; margin-top: 8px; }
  .content { padding: 32px; }
  .drop-card {
    border: 1px solid #e5e7eb; border-radius: 8px; padding: 20px;
    margin-bottom: 16px; background: #f9fafb;
  }
  .hotel-name { font-size: 18px; font-weight: bold; color: #111827; margin: 0 0 4px; }
  .location { font-size: 14px; color: #6b7280; margin: 0 0 12px; }
  .dates { font-size: 14px; color: #374151; margin: 0 0 12px; }
  .price-row { display: flex; align-items: center; gap: 12px; margin-top: 8px; }
  .old-price {
    text-decoration: line-through; color: #9ca3af; font-size: 16px;
  }
  .new-price { color: #dc2626; font-size: 20px; font-weight: bold; }
  .discount {
    background: #dc2626; color: #ffffff; font-size: 12px; font-weight: bold;
    padding: 2px 8px; border-radius: 4px;
  }
  .footer {
    padding: 24px 32px; text-align: center; color: #9ca3af;
    font-size: 12px; border-top: 1px solid #e5e7eb;
  }
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <h1>Price Drop Alert</h1>
    <div class="subtitle">{{.DropCount}} hotel{{if ne .DropCount "1"}}s{{end}} just dropped in price!</div>
  </div>
  <div class="content">
    {{range .Drops}}
    <div class="drop-card">
      <p class="hotel-name">{{.HotelName}}</p>
      <p class="location">{{.Location}}</p>
      <p class="dates">Check-in: {{.StartDate}} &mdash; Check-out: {{.EndDate}}</p>
      <div class="price-row">
        <span class="old-price">{{.Currency}} {{.OldPrice}}</span>
        <span class="new-price">{{.Currency}} {{.NewPrice}}</span>
        <span class="discount">{{.DiscountPct}}% off</span>
      </div>
    </div>
    {{end}}
  </div>
  <div class="footer">
    Prices Recommender &mdash; This is an automated alert.
  </div>
</div>
</body>
</html>`))
