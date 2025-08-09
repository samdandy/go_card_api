package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/schema"
	"github.com/samdandy/go_card_api/api"
	log "github.com/sirupsen/logrus"
)

func fetch_html(url_str string) (io.ReadCloser, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	defer client.CloseIdleConnections()

	resp, err := client.Get(url_str)
	if err != nil {
		return nil, fmt.Errorf("error fetching URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Status Code:", resp.StatusCode)
	fmt.Println("Content-Type:", resp.Header.Get("Content-Type"))
	fmt.Println("Content-Length:", resp.Header.Get("Content-Length"))

	return resp.Body, nil // Don't close here — caller will close
}

func parse_html(r io.Reader) (float64, []float64) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}

	var prices []float64
	var sum float64

	doc.Find(".s-item__price").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())

		// Remove $ and commas
		clean := strings.ReplaceAll(text, "$", "")
		clean = strings.ReplaceAll(clean, ",", "")

		// Some eBay prices are in format "$12.34 to $15.67"
		if strings.Contains(clean, " to ") {
			parts := strings.Split(clean, " to ")
			clean = parts[0] // take the first price
		}

		if f, err := strconv.ParseFloat(clean, 64); err == nil {
			prices = append(prices, f)
			sum += f
		}
	})

	if len(prices) > 0 {
		return sum / float64(len(prices)), prices
	}
	return 0, prices
}

func GetAvgPrice(w http.ResponseWriter, r *http.Request) {
	var params = api.CardAPIParams{}
	var decoder *schema.Decoder = schema.NewDecoder()
	var err error

	err = decoder.Decode(&params, r.URL.Query())

	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
	search := params.SearchCrit
	search = strings.ReplaceAll(search, " ", "+")
	html_data, err := fetch_html("https://www.ebay.com/sch/i.html?_nkw=" + search + "&LH_Complete=1&LH_Sold=1")
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
	defer html_data.Close()
	avg_price, prices := parse_html(html_data)

	var response = api.CardAvgPrice{
		AveragePrice: avg_price,
		StatusCode:   http.StatusOK,
		CardPrices:   prices,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

}
