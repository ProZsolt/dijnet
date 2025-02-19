package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProZsolt/dijnet"
)

func main() {
	user := os.Getenv("DIJNET_USERNAME")
	password := os.Getenv("DIJNET_PASSWORD")
	invoicePath := os.Getenv("DIJNET_INVOICE_PATH")

	srv := dijnet.NewService()
	err := srv.Login(user, password)
	if err != nil {
		log.Fatalln("Login error: ", err)
	}

	_, token, err := srv.Providers()
	if err != nil {
		log.Fatalln("Unable to get providers: ", err)
	}

	log.Println("Get invoices")
	invoices, err := srv.Invoices(dijnet.InvoicesQuery{Token: token})
	if err != nil {
		log.Fatalln("Unable to get invoices: ", err)
	}
	if len(invoices) == 0 {
		log.Println("Unable to find any invoice")
	}

	for i, invoice := range invoices {
		fmt.Printf("Downloading invoice %d/%d\n", i+1, len(invoices))
		basePath := "invoices"
		if invoicePath != "" {
			basePath = invoicePath
		}
		providerPath := filepath.Join(basePath, invoice.Provider, invoice.IssuerID)
		err = os.MkdirAll(providerPath, os.ModePerm)
		if err != nil {
			log.Fatalf("Unable to create directory %s: %v", providerPath, err)
		}

		invoiceFilename := invoice.DateOfIssue.Format("2006-01-02") + "_" + strings.Replace(invoice.InvoiceID, "/", "_", -1)
		PDF := filepath.Join(providerPath, invoiceFilename+".pdf")
		XML := filepath.Join(providerPath, invoiceFilename+".xml")
		err = srv.DownloadInvoice(invoice, PDF, XML)
		if err != nil {
			fmt.Println("DownloadInvoice error: ", err)
			return
		}
	}
}
