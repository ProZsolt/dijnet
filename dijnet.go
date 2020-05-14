package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
)

func download() {

	username := os.Getenv("DIJNET_USERNAME")
	password := os.Getenv("DIJNET_PASSWORD")
	c := colly.NewCollector()
	c.AllowURLRevisit = true

	c.Post("https://www.dijnet.hu/ekonto/login/login_check_password", map[string]string{
		"vfw_form": "login_check_password",
		"vfw_coll": "direct",
		"username": username,
		"password": password,
	})

	c.Visit("https://www.dijnet.hu/ekonto/control/szamla_search")

	c.OnHTML(".sortable tr", func(e *colly.HTMLElement) {
		fmt.Println(e.Attr("id"))
		id := e.Attr("id")
		rowid := strings.Split(id, "_")[1]

		fmt.Println(e.Text)

		c2 := c.Clone()
		c2.Visit("https://www.dijnet.hu/ekonto/control/szamla_select?vfw_coll=szamla_list&vfw_rowid=" + rowid + "&exp=K")
		c2.Visit("https://www.dijnet.hu/ekonto/control/szamla_letolt")

		c3 := c2.Clone()
		c3.OnResponse(func(r *colly.Response) {
			r.Save("./invoices/" + r.FileName())
		})
		c3.Visit("https://www.dijnet.hu/ekonto/control/szamla_pdf")
		c3.Visit("https://www.dijnet.hu/ekonto/control/szamla_xml")

		c2.Visit("https://www.dijnet.hu/ekonto/control/szamla_list")
	})

	c.Post("https://www.dijnet.hu/ekonto/control/szamla_search_submit", map[string]string{
		"vfw_form":     "szamla_search_submit",
		"vfw_coll":     "szamla_search_params",
		"szlaszolgnev": "",
		"regszolgid":   "",
		"datumtol":     "2020.03.15",
		"datumig":      "2020.04.12",
	})
}

func main() {
	download()
}
