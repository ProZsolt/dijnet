package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
)

func main() {

	username := os.Getenv("DIJNET_USERNAME")
	password := os.Getenv("DIJNET_PASSWORD")
	c := colly.NewCollector()
	c.AllowURLRevisit = true

	err := c.Post("https://www.dijnet.hu/ekonto/login/login_check_password", map[string]string{
		"vfw_form": "login_check_password",
		"vfw_coll": "direct",
		"username": username,
		"password": password,
	})
	if err != nil {
		log.Fatal(err)
	}

	c.Visit("https://www.dijnet.hu/ekonto/control/szamla_search")

	c.OnHTML(".szamla_table tr td:nth-of-type(1)", func(e *colly.HTMLElement) {
		onclick := e.Attr("onclick")
		link := strings.Split(onclick, "'")[1]

		fmt.Println(e.Text)

		c2 := c.Clone()
		c2.Visit(e.Request.AbsoluteURL(link))
		c2.Visit("https://www.dijnet.hu/ekonto/control/szamla_letolt")

		c3 := c2.Clone()
		c3.OnResponse(func(r *colly.Response) {
			r.Save(r.FileName())
		})
		c3.Visit("https://www.dijnet.hu/ekonto/control/szamla_pdf")
		c3.Visit("https://www.dijnet.hu/ekonto/control/szamla_xml")

		c2.Visit("https://www.dijnet.hu/ekonto/control/szamla_list")
	})

	c.Post("https://www.dijnet.hu/ekonto/control/szamla_search_submit", map[string]string{
		"vfw_form": "szamla_search_submit",
		"vfw_coll": "szamla_search_params",
	})
}
