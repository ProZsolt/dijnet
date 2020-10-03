package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ProZsolt/dijnet"
)

func main() {
	// Authentication
	srv := dijnet.NewService()
	err := srv.Login("username", "password")
	if err != nil {
		fmt.Println(err)
		return
	}

	//List available providers
	providers, err := srv.Providers()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, provider := range providers {
		fmt.Println(provider)
	}

	// Download the last 3 months of Invoices from NKM Főldgáz
	query := dijnet.InvoicesQuery{
		Provider: dijnet.NKMFoldgaz,
		From:     time.Now().AddDate(0, -3, 0),
	}
	invoices, err := srv.Invoices(query)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, invoice := range invoices {
		basePath := "invoices"
		PDF := filepath.Join(basePath, invoice.InvoiceID+".pdf")
		XML := filepath.Join(basePath, invoice.InvoiceID+".xml")
		err = srv.DownloadInvoice(invoice, PDF, XML)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
