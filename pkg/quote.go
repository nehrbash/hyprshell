package core

import (
	"context"
	"fmt"
	"hyprshell/pkg/quote"
	"log"

	"github.com/godbus/dbus/v5"
)

const QuoteInterfaceName = "com.hypr.QuoteService"

type QuoteService struct {
	quoteSignal chan string
	Author      string
	Quote       string
}

func (q *QuoteService) Update() {
	var err error
	q.Quote, q.Author, err = quote.GetRandomQuote()
	if err != nil {
		log.Print(err)
	}
	q.quoteSignal <- fmt.Sprintf(`{ "quote": "%s", "author": "%s" }`, q.Quote, q.Author)
}

func (q *QuoteService) GetQuote(msg string) (err *dbus.Error) {
	q.Update()
	return nil
}

func DbusQuote(ctx context.Context) {
	cron := GetGlobalScheduler()
	service := &QuoteService{quoteSignal: make(chan string, 2)}
	cron.Every("5m").Do(func() {
		service.Update()
	})

	conn := GetDbusConnection()
	err := conn.Export(service, ServiceObjectPath, QuoteInterfaceName)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-service.quoteSignal:
			err = conn.Emit(ServiceObjectPath, QuoteInterfaceName+".quote", data)
			if err != nil {
				log.Print(err)
			}
		}
	}
}
