package dijnet

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/htmlindex"
)

func NewService() Service {
	cookieJar, _ := cookiejar.New(nil)

	return Service{client: &http.Client{
		Jar: cookieJar,
	}}
}

type Service struct {
	client *http.Client
}

func isLoggedIn(body string) bool {
	if strings.Contains(string(body), "Bejelentkez&eacute;si n&eacute;v: <strong>") {
		return true
	}
	return false
}

func isRequestOrderRight(body string) bool {
	if strings.Contains(string(body), "K&eacute;rj&uuml;k, csak az oldalon tal&aacute;lhat&oacute; gombokat &eacute;s hivatkoz&aacute;sokat haszn&aacute;lja!") {
		return false
	}
	return true
}

func decodeHTMLBody(body io.Reader) io.Reader {
	e, _ := htmlindex.Get("iso-8859-2")
	return e.NewDecoder().Reader(body)
}

func (s Service) Login(username string, password string) error {
	payload := url.Values{}
	payload.Set("vfw_form", "login_check_password")
	payload.Set("vfw_coll", "direct")
	payload.Set("username", username)
	payload.Set("password", password)
	req, err := http.NewRequest(http.MethodPost,
		"https://www.dijnet.hu/ekonto/login/login_check_password",
		strings.NewReader(payload.Encode()),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if !isLoggedIn(string(body)) {
		return fmt.Errorf("wrong username or password")
	}
	return nil
}

func (s Service) Providers() ([]string, error) {
	var ret []string
	resp, err := s.client.Get("https://www.dijnet.hu/ekonto/control/szamla_search")
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ret, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}
	body, err := ioutil.ReadAll(decodeHTMLBody(resp.Body))
	if err != nil {
		return ret, err
	}
	r, _ := regexp.Compile("sopts.add\\('(.+)'\\)")
	m := r.FindAllStringSubmatch(string(body), -1)
	for _, e := range m {
		ret = append(ret, e[1])
	}
	return ret, err
}

type Invoice struct {
	ID              string
	Provider        string
	IssuerID        string
	InvoiceID       string
	DateOfIssue     time.Time
	Total           int
	PaymentDeadline time.Time
	Payable         int
	Status          string
}

type InvoicesQuery struct {
	provider string
	issuerID string
	from     time.Time
	to       time.Time
}

func cleanNumber(r rune) rune {
	if unicode.IsNumber(r) {
		return r
	}
	return -1
}

func (s Service) Invoices(query InvoicesQuery) ([]Invoice, error) {
	var ret []Invoice

	e, _ := htmlindex.Get("iso-8859-2")
	provider, err := e.NewEncoder().String(query.provider)
	if err != nil {
		return ret, err
	}

	dateLayout := "2006.01.02"
	var from, to string
	if !query.from.IsZero() {
		from = query.from.Format(dateLayout)
	}
	if !query.to.IsZero() {
		to = query.to.Format(dateLayout)
	}

	payload := url.Values{}
	payload.Set("vfw_form", "szamla_search_submit")
	payload.Set("vfw_coll", "szamla_search_params")
	payload.Set("szlaszolgnev", provider)
	payload.Set("regszolgid", query.issuerID)
	payload.Set("datumtol", from)
	payload.Set("datumig", to)
	req, err := http.NewRequest(http.MethodPost,
		"https://www.dijnet.hu/ekonto/control/szamla_search_submit",
		strings.NewReader(payload.Encode()),
	)
	if err != nil {
		return ret, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ret, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ret, err
	}

	doc.Find(".sortable tr").Each(func(_ int, s *goquery.Selection) {
		id, _ := s.Attr("id")
		invoice := Invoice{ID: strings.Split(id, "_")[1]}
		s.Find("td").Each(func(i int, s *goquery.Selection) {
			switch i {
			case 0:
				invoice.Provider = s.Text()
			case 1:
				invoice.IssuerID = s.Text()
			case 2:
				invoice.InvoiceID = s.Text()
			case 3:
				invoice.DateOfIssue, _ = time.Parse(dateLayout, s.Text())
			case 4:
				n := strings.Map(cleanNumber, s.Text())
				invoice.Total, _ = strconv.Atoi(n)
			case 5:
				invoice.PaymentDeadline, _ = time.Parse(dateLayout, s.Text())
			case 6:
				n := strings.Map(cleanNumber, s.Text())
				invoice.Payable, _ = strconv.Atoi(n)
			case 7:
				invoice.Status = s.Text()
			}
		})
		ret = append(ret, invoice)
	})
	return ret, nil
}

func (s Service) downloadFile(URL string, filename string) error {
	resp, err := s.client.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func (s Service) DownloadInvoice(i Invoice, pdf string, xml string) error {
	resp, err := s.client.Get(
		"https://www.dijnet.hu/ekonto/control/szamla_select?vfw_coll=szamla_list&vfw_rowid=" + i.ID + "&exp=K",
	)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	resp, err = s.client.Get(
		"https://www.dijnet.hu/ekonto/control/szamla_letolt",
	)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}

	if pdf != "" {
		err = s.downloadFile("https://www.dijnet.hu/ekonto/control/szamla_pdf", pdf)
		if err != nil {
			return err
		}
	}
	if xml != "" {
		err = s.downloadFile("https://www.dijnet.hu/ekonto/control/szamla_xml", xml)
		if err != nil {
			return err
		}
	}
	resp, err = s.client.Get(
		"https://www.dijnet.hu/ekonto/control/szamla_list",
	)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
}
