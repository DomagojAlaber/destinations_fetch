package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DomagojAlaber/destinations_fetch/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

type WikiDataResponse struct {
	Results struct {
		Bindings []struct {
			ItemLabel struct {
				Value string `json:"value"`
			} `json:"itemLabel"`
			Coord struct {
				Value string `json:"value"`
			} `json:"coord"`
		} `json:"bindings"`
	} `json:"results"`
}

type Place struct {
	Name  string
	Coord string
}

func main() {
	var r WikiDataResponse

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

	if err := json.Unmarshal(body, &r); err != nil {
		log.Fatal(err)
	}

	places := make([]Place, 0, len(r.Results.Bindings))
	for _, b := range r.Results.Bindings {
		places = append(places, Place{
			Name:  b.ItemLabel.Value,
			Coord: b.Coord.Value,
		})
	}

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "")
	if err != nil {
		fmt.Print(err)
	}
	defer conn.Close(ctx)

	queries := db.New(conn)

	for _, b := range places {
		lon, lat, err := parseWTKPoint(b.Coord)
		if err != nil {
			log.Fatal(err)
		}
		queries.UpsertDestination(ctx, db.UpsertDestinationParams{
			Name:   pgtype.Text{String: b.Name},
			Region: pgtype.Text{String: "Istria"},
			Lon:    pgtype.Float8{Float64: lon},
			Lat:    pgtype.Float8{Float64: lat},
		})
		fmt.Printf("%s -> lon=%f lat=%f\n", b.Name, lon, lat)
	}
}

func parseWTKPoint(s string) (lon, lat float64, err error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "Point(")
	s = strings.TrimSuffix(s, ")")

	parts := strings.Fields(s)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected WKT point: %q", s)
	}

	lon, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, err
	}

	lat, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, err
	}

	return lon, lat, err
}
