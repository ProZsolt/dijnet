# Díjnet

(Unofficial) [Díjnet][1] API client for Go

Download it using

```go get github.com/ProZsolt/dijnet```

## Usage

```go
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
    // handle error
  }
  
  //List available providers
  providers, err := srv.Providers()
  if err != nil {
    // handle error
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
    // handle error
  }

  for _, invoice := range invoices {
    basePath := "invoices"
    PDF := filepath.Join(basePath, invoice.InvoiceID+".pdf")
    XML := filepath.Join(basePath, invoice.InvoiceID+".xml")
    err = srv.DownloadInvoice(invoice, PDF, XML)
    if err != nil {
      // handle error
    }
  }
}
```

## Limitation

Due to Díjnet don't let you navigate easily via URLs, you have to call Providers, Invoices, DownloadInvoice in this order otherwise it won't work.

[1]: https://www.dijnet.hu/