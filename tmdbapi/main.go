package main

import (
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	http.HandleFunc("/movie/", func(w http.ResponseWriter, r *http.Request) {
		// id := strings.TrimPrefix(r.URL.Path, "/movie/")

		// cl := NewClient()
		// m, err := cl.FindMovie(id)
		// if err != nil {
		// 	log.Print(err)
		// }
		var err error

		m := Movie{
			ID:          "670292",
			Title:       "The Creator",
			ReleaseYear: 2023,
			PosterURL:   "https://www.themoviedb.org/t/p/w300_and_h450_bestv2/vBZ0qvaRxqEhZwl6LWmruJqWE8Z.jpg",
			BackdropURL: "https://www.themoviedb.org/t/p/original/kjQBrc00fB2RjHZB3PGR4w9ibpz.jpg",
			Overview:    "Amid a future war between the human race and the forces of artificial intelligence, a hardened ex-special forces agent grieving the disappearance of his wife, is recruited to hunt down and kill the Creator, the elusive architect of advanced AI who has developed a mysterious weapon with the power to end the war—and mankind itself.",
			Directors:   []string{"Gareth Edwards", "Claudio Sampaio"},
		}

		type Source struct {
			Title      string
			Resolution string
			Languages  []string
			Seeders    int
			Size       int
		}

		data := struct {
			Content Movie
			Sources []Source
		}{
			Content: m,
			Sources: []Source{
				{
					Title:      "The Creator (2023) 1080p WebRIP.mkv",
					Resolution: "4K",
					Languages:  []string{"us", "br"},
					Seeders:    18,
				},
				{
					Title:      "The Creator (2023) 1080p WebRIP.mkv",
					Resolution: "1080p",
					Languages:  []string{"us"},
					Seeders:    8,
				},
				{
					Title:      "The Creator (2023) 1080p WebRIP.mkv",
					Resolution: "1080p",
					Languages:  []string{"us", "ua"},
					Seeders:    5,
				},
				{
					Title:      "The Creator (2023) 1080p WebRIP.mkv",
					Resolution: "720p",
					Languages:  []string{"br", "fr"},
					Seeders:    2,
				},
			},
		}
		// fmt.Printf("%#v\n", m)

		err = template.Must(template.ParseFiles("content.tmpl.html")).Execute(w, data)
		if err != nil {
			log.Print(err)
		}
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}

var DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

type Client struct {
	HTTP    *http.Client
	BaseURL string
}

func NewClient() Client {
	return Client{
		HTTP:    http.DefaultClient,
		BaseURL: "https://www.themoviedb.org",
	}
}

type Movie struct {
	ID          string
	Title       string
	ReleaseYear int
	PosterURL   string
	BackdropURL string
	Overview    string
	Directors   []string
}

func (cl Client) FindMovie(id string) (mov Movie, err error) {
	req, err := http.NewRequest("GET", cl.BaseURL+"/movie/"+id, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", DefaultUserAgent)

	resp, err := cl.HTTP.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	headerEl := doc.Find("#original_header").First()

	mov.ID = id
	mov.Title = headerEl.Find("h2 a").First().Text()

	img := headerEl.Find("img.poster").First()
	mov.PosterURL = cl.BaseURL + img.AttrOr("data-src", "")

	mov.Overview = strings.TrimSpace(headerEl.Find(".overview").Text())

	headerEl.Find(".profile").Each(func(i int, s *goquery.Selection) {
		character := s.Find(".character").First().Text()
		if strings.Contains(character, "Director") {
			mov.Directors = append(mov.Directors, s.Find("a").First().Text())
		}
	})

	sYear := headerEl.Find(".release_date").Text()
	sYear = strings.Trim(sYear, "()")
	mov.ReleaseYear, err = strconv.Atoi(sYear)
	if err != nil {
		return
	}

	style := doc.Find("#main style").First().Text()

	re := regexp.MustCompile(`div\.header\.large\.first \{(.*|\n)+\}`)
	headerStyle := re.FindString(style)

	re = regexp.MustCompile(`background-image: url\('(.*)'\)`)
	m := re.FindStringSubmatch(headerStyle)
	if len(m) >= 2 {
		url := m[1]
		re = regexp.MustCompile(`/t/p/.*?/`)
		mov.BackdropURL = re.ReplaceAllString(url, "/t/p/original/")
		mov.BackdropURL = cl.BaseURL + mov.BackdropURL
	}

	return
}
