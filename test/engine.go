package test

import (
	"net/http"
	"time"

	"github.com/nyaruka/gocommon/urns"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/engine"
	"github.com/nyaruka/goflow/services/webhooks"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// NewEngine creates an engine instance for testing
func NewEngine() flows.Engine {
	return engine.NewBuilder().
		WithEmailServiceFactory(func(s flows.Session) (flows.EmailService, error) {
			return newEmailService(), nil
		}).
		WithWebhookServiceFactory(webhooks.NewServiceFactory(http.DefaultClient, map[string]string{"User-Agent": "goflow-testing"}, 10000)).
		WithClassificationServiceFactory(func(s flows.Session, c *flows.Classifier) (flows.ClassificationService, error) {
			return newClassificationService(c), nil
		}).
		WithAirtimeServiceFactory(func(flows.Session) (flows.AirtimeService, error) { return newAirtimeService("RWF"), nil }).
		Build()
}

// implementation of an email service for testing which just fakes sending the email
type emailService struct {
	classifier *flows.Classifier
}

func newEmailService() *emailService {
	return &emailService{}
}

func (s *emailService) Send(session flows.Session, addresses []string, subject, body string) error {
	return nil
}

// implementation of a classification service for testing which always returns the first intent
type classificationService struct {
	classifier *flows.Classifier
}

func newClassificationService(classifier *flows.Classifier) *classificationService {
	return &classificationService{classifier: classifier}
}

func (s *classificationService) Classify(session flows.Session, input string, logHTTP flows.HTTPLogCallback) (*flows.Classification, error) {
	classifierIntents := s.classifier.Intents()
	extractedIntents := make([]flows.ExtractedIntent, len(s.classifier.Intents()))
	confidence := decimal.RequireFromString("0.5")
	for i := range classifierIntents {
		extractedIntents[i] = flows.ExtractedIntent{classifierIntents[i], confidence}
		confidence = confidence.Div(decimal.RequireFromString("2"))
	}

	logHTTP(&flows.HTTPLog{
		URL:       "http://test.acme.ai?classifiy",
		Request:   "GET /?classifiy HTTP/1.1\r\nHost: test.acme.ai\r\nUser-Agent: Go-http-client/1.1\r\nAccept-Encoding: gzip\r\n\r\n",
		Response:  "HTTP/1.0 200 OK\r\nContent-Length: 14\r\n\r\n{\"intents\":[]}",
		Status:    "success",
		CreatedOn: time.Date(2019, 10, 16, 13, 59, 30, 123456789, time.UTC),
		ElapsedMS: 1000,
	})

	classification := &flows.Classification{
		Intents: extractedIntents,
		Entities: map[string][]flows.ExtractedEntity{
			"location": []flows.ExtractedEntity{
				flows.ExtractedEntity{"Quito", decimal.RequireFromString("1.0")},
			},
		},
	}

	return classification, nil
}

var _ flows.ClassificationService = (*classificationService)(nil)

// implementation of an airtime service for testing which uses a fixed currency
type airtimeService struct {
	fixedCurrency string
}

func newAirtimeService(currency string) *airtimeService {
	return &airtimeService{fixedCurrency: currency}
}

func (s *airtimeService) Transfer(session flows.Session, sender urns.URN, recipient urns.URN, amounts map[string]decimal.Decimal, logHTTP flows.HTTPLogCallback) (*flows.AirtimeTransfer, error) {
	logHTTP(&flows.HTTPLog{
		URL:       "http://send.airtime.com",
		Request:   "GET / HTTP/1.1\r\nHost: send.airtime.com\r\nUser-Agent: Go-http-client/1.1\r\nAccept-Encoding: gzip\r\n\r\n",
		Response:  "HTTP/1.0 200 OK\r\nContent-Length: 15\r\n\r\n{\"status\":\"ok\"}",
		Status:    "success",
		CreatedOn: time.Date(2019, 10, 16, 13, 59, 30, 123456789, time.UTC),
		ElapsedMS: 0,
	})

	amount, hasAmount := amounts[s.fixedCurrency]
	if !hasAmount {
		return nil, errors.Errorf("no amount configured for transfers in %s", s.fixedCurrency)
	}

	transfer := &flows.AirtimeTransfer{
		Sender:        sender,
		Recipient:     recipient,
		Currency:      s.fixedCurrency,
		DesiredAmount: amount,
		ActualAmount:  amount,
	}

	return transfer, nil
}

var _ flows.AirtimeService = (*airtimeService)(nil)
