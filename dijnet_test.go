package dijnet

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestThings(t *testing.T) {
	username := os.Getenv("DIJNET_USERNAME")
	password := os.Getenv("DIJNET_PASSWORD")
	srv := NewService()
	err := srv.Login(username, password)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	providers, err := srv.Providers()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	fmt.Println(providers)
	query := InvoicesQuery{
		From: time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC),
	}
	invoices, err := srv.Invoices(query)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	fmt.Printf("%+v", invoices)

	for _, i := range invoices {
		err = srv.DownloadInvoice(i,
			"invoices/"+strings.Replace(i.InvoiceID, "/", "_", -1)+".pdf",
			"invoices/"+strings.Replace(i.InvoiceID, "/", "_", -1)+".xml",
		)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}
