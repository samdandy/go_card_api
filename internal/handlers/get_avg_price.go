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
	db_tools "github.com/samdandy/go_card_api/internal/tools"
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

func parse_html(r io.Reader) (float64, []api.Card) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}
	var cards []api.Card
	var sum float64
	var prices []float64
	var titles []string
	var image_urls []string
	doc.Find(".s-item__title").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Text())
		titles = append(titles, title)
	})
	doc.Find(".s-item img").Each(func(i int, s *goquery.Selection) {
		image_url, _ := s.Attr("src")
		image_urls = append(image_urls, image_url)
	})
	doc.Find(".s-item__price").Each(func(i int, s *goquery.Selection) {
		fmt.Println("Found price:", s.Text())
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
		var averagePrice float64 = sum / float64(len(prices))
		for i := range prices {
			cards = append(cards, api.Card{
				ListingTitle: titles[i],
				Price:        prices[i],
				ImageURL:     image_urls[i],
			})
		}
		return averagePrice, cards
	}
	return 0, cards
}

func CleanseCards(cards []api.Card) []api.Card {

	var cleanedCards []api.Card

	for _, card := range cards {
		if !ValidCardTitle(card.ListingTitle) {
			fmt.Println("Skipping invalid card:", card.ListingTitle)
			continue
		}
		if card.Price <= 0 {
			fmt.Println("Skipping card with non-positive price:", card.ListingTitle, "Price:", card.Price)
			continue
		}
		cleanedCards = append(cleanedCards, card)
	}
	return cleanedCards
}

func ValidCardTitle(title string) bool {
	unwantedPhrases := []string{
		"shop on ebay",
		"see more like this",
		"sponsored",
		"advertisement",
	}
	titleLower := strings.ToLower(title)
	for _, phrase := range unwantedPhrases {
		if strings.Contains(titleLower, phrase) {
			fmt.Println("Invalid card title:", title)
			return false
		}
	}
	return true
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
	html_data, err := fetch_html("https://www.ebay.com/sch/i.html?_nkw=" + search + "&_sacat=64482&LH_Complete=1&LH_Sold=1")
	fmt.Println("Fetching URL:", "https://www.ebay.com/sch/i.html?_nkw="+search+"&_sacat=64482&LH_Complete=1&LH_Sold=1")
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
	defer html_data.Close()
	avg_price, cards := parse_html(html_data)
	cards = CleanseCards(cards)
	var response = api.CardSearchResponse{
		AveragePrice: avg_price,
		StatusCode:   http.StatusOK,
		Cards:        cards,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	go db_tools.DB.WriteCardSearchLog(search, int64(len(cards)))

}
