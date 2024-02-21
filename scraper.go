package main

import (
	"errors"
	"fmt" //formatted I/O
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly" //scraping framework
	log "github.com/sirupsen/logrus"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got / request\n")
	io.WriteString(w, "This is my website!\n")
}
func getHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /hello request\n")
	io.WriteString(w, "Hello, HTTP!\n")
}

//Results:div.s-result-list.s-search-results.sg-row
//Items:div.a-section.a-spacing-base
//Name:span.a-size-base-plus.a-color-base.a-text-normal
//Price:span.a-price-whole
//Stars:span.a-icon-alt

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", getRoot)
	mux.HandleFunc("GET /hello", getHello)
	mux.HandleFunc("GET /search/{search_word}", getSearch)

	server := &http.Server{Handler: mux}
	// err := http.ListenAndServe(":3030", nil)
	l, err := net.Listen("tcp4", "0.0.0.0:3030")
	if err != nil {
		log.WithError(err).Error("Failed to bind to the port")
		os.Exit(1)
	}
	err = server.Serve(l)
	if errors.Is(err, http.ErrServerClosed) {
		log.Error("server closed")
	} else if err != nil {
		log.WithError(err).Error("Failed to start the server")
		os.Exit(1)
	}
}

func getSearch(w http.ResponseWriter, r *http.Request) {

	//  var search_word string
	// fmt.Print("What do you want to search for today? ")
	// scanner := bufio.NewScanner(os.Stdin)
	// scanner.Scan()
	// search_word := scanner.Text()
	search_word := r.PathValue("search_word")
	search_word = strings.ReplaceAll(search_word, " ", "%20")
	log.Info("Search Word: %s", search_word)

	c := colly.NewCollector(colly.AllowedDomains("www.amazon.in"))

	c.OnRequest(func(r *colly.Request) {
		log.Info("Link of the page: %s", r.URL)
		// fmt.Println("Link of the page:", r.URL)
	})

	c.OnHTML("div.s-result-list.s-search-results.sg-row", func(h *colly.HTMLElement) {
		io.WriteString(w, "<html><body><table>\n<tr><th>Product Name</th><th>Ratings</th><th>Price</th></tr>\n")
		h.ForEach("div.a-section.a-spacing-base", func(_ int, h *colly.HTMLElement) {
			var name string
			name = h.ChildText("span.a-size-base-plus.a-color-base.a-text-normal")
			var stars string
			stars = h.ChildText("span.a-icon-alt")
			v := h.ChildText("span.a-price-whole")
			v = strings.Replace(v, ",", "", -1)
			price, err := strconv.ParseFloat(v, 32)
			if err != nil {
				log.WithError(err).Error("Error! Can't parse Price to Int")
				return
			}
			//  price = strconv.ParseInt(priceAsString)
			fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%.2f</td>\n", name, stars, price)
			// fmt.Println("ProductName: ", name)
			// fmt.Println("Ratings: ", stars)
			// fmt.Println("Price: ", price)
			log.WithFields(log.Fields{
				"ProductName": name,
				"Ratings":     stars,
				"Price":       price,
			}).Info("An entry scraped")
		})
		io.WriteString(w, "\n</table></body></html>")
	})
	URL := fmt.Sprintf("https://www.amazon.in/s?k=%s&ref=nb_sb_noss", search_word)
	w.Header().Set("Content-Type", "text/html")
	c.Visit(URL)
}
