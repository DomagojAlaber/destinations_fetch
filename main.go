package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const endpoint = "https://query.wikidata.org/sparql"

const uri = `
SELECT ?item ?itemLabel ?type ?typeLabel ?coord WHERE {
  VALUES ?type { wd:Q15105893 wd:Q57058 }  # towns + municipalities
  ?item wdt:P31 ?type;
        wdt:P131 wd:Q58268;
        wdt:P625 ?coord.
  SERVICE wikibase:label { bd:serviceParam wikibase:language "hr,en". }
}
ORDER BY ?typeLabel ?itemLabel
`

func main() {

	u, err := url.Parse(endpoint)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse endpoint:", err)
		os.Exit(1)
	}

	q := u.Query()
	q.Set("format", "json")
	q.Set("query", uri)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "new request:", err)
		os.Exit(1)
	}
	// Wikidata recommends identifying your client; some endpoints may throttle anonymous clients.
	req.Header.Set("User-Agent", "destionations_fetch/1.0 (contact: you@example.com)")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error fetching request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "bad status: %s\n%s\n", resp.Status, string(b))
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading body:", err)
		os.Exit(1)
	}

	fmt.Println(string(body))
}
