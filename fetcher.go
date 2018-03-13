package avurnav

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

const (
	// AVURNAVCategoryID is the category ID for AVURNAVs on the Préfet Maritime
	// websites
	AVURNAVCategoryID = 12
)

// AVURNAVFetcher fetches AVURNAVs on the Préfet
// Maritime websites
type AVURNAVFetcher struct {
	service PremarInterface
}

// AVURNAVPayload is used to decode AVURNAVs from
// the Préfet Maritime websites
type AVURNAVPayload struct {
	ID        string `json:"id_centre"`
	Number    string `json:"numero"`
	Title     string `json:"label"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	City      string `json:"localite"`
	URL       string `json:"url"`
	Dates     string `json:"dates"`
}

// AVURNAV transforms a payload to a proper AVURNAV
func (p AVURNAVPayload) AVURNAV(premar PremarInterface) AVURNAV {
	relative, err := url.Parse(p.URL)
	if err != nil {
		panic(err)
	}

	from, to := p.matchDates(p.Dates)

	avurnav := AVURNAV{
		ID:           p.parseInt(p.ID),
		Number:       p.Number,
		Title:        strings.TrimSpace(p.Title),
		Latitude:     p.parseFloat(p.Latitude),
		Longitude:    p.parseFloat(p.Longitude),
		URL:          premar.BaseURL().ResolveReference(relative).String(),
		PreMarRegion: premar.Region(),
	}

	if from != "" {
		avurnav.ValidFrom = &from
	}
	if to != "" {
		avurnav.ValidUntil = &to
	}

	return avurnav
}

func (p AVURNAVPayload) matchDates(str string) (from, to string) {
	re := regexp.MustCompile(`^En vigueur du : (\d{2}\/\d{2}\/\d{4}|Indéterminé) au (\d{2}\/\d{2}\/\d{4}|Indéterminé)$`)
	matches := re.FindStringSubmatch(str)
	return p.cleanDate(matches[1]), p.cleanDate(matches[2])
}

func (p AVURNAVPayload) cleanDate(str string) string {
	if str == "Indéterminé" {
		return ""
	}
	t, _ := time.Parse("02/01/2006", str)
	return t.Format("2006-01-02")
}

func (p AVURNAVPayload) parseFloat(str string) float32 {
	s, err := strconv.ParseFloat(str, 32)
	if err != nil {
		panic(err)
	}
	return float32(s)
}

func (p AVURNAVPayload) parseInt(str string) int {
	s, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return s
}

// AVURNAV represents an AVURNAV
type AVURNAV struct {
	// ID is the a technical unique number
	ID int `json:"id"`
	// Number is the number of the AVURNAV. This is the main public identifier
	Number string `json:"number"`
	// Title is the title of the AVURNAV
	Title string `json:"title"`
	// Content is the content of the AVURNAV
	Content string `json:"content"`
	// Latitude gives an indication about the localisation of the AVURNAV.
	// It's not super reliable for now because AVURNAVs can spawn multiple
	// geographical regions but for now Préfet Maritimes only give a single point.
	Latitude float32 `json:"latitude"`
	// Longitude gives an indication about the localisation of the AVURNAV.
	// It's not super reliable for now because AVURNAVs can spawn multiple
	// geographical regions but for now Préfet Maritimes only give a single point.
	Longitude float32 `json:"longitude"`
	// URL gives a full URL to a Préfet Maritime website concerning this specific AVURNAV
	URL string `json:"url"`
	// ValidFrom tells when the AVURNAV will be in force. Format: YYYY-MM-DD
	ValidFrom *string `json:"valid_from"`
	// ValidUntil tells when the AVURNAV will not be valid anymore. Format: YYYY-MM-DD
	ValidUntil *string `json:"valid_until"`
	// PreMarRegion gives the region under the authority of this Préfet Maritime
	PreMarRegion string `json:"premar_region"`
}

// ParseContent fills the content section of an AVURNAV and returns a new one
func (a AVURNAV) ParseContent(reader io.Reader) AVURNAV {
	root, err := html.Parse(reader)
	if err != nil {
		panic(err)
	}
	blocks := scrape.FindAll(root, scrape.ByClass("block-subcontenu"))
	content := scrape.Text(blocks[2])
	a.Content = content
	return a
}

// JSON gets the JSON representation of an AVURNAV
func (a AVURNAV) JSON() string {
	res, err := a.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return string(res)
}

// MarshalBinary marshals the object
func (a AVURNAV) MarshalBinary() ([]byte, error) {
	return json.Marshal(a)
}

// UnmarshalBinary unmarshals the object
func (a *AVURNAV) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &a)
}

// AVURNAVs represents multiple AVURNAV
type AVURNAVs []AVURNAV

// AVURNAVPayloads represents multiple AVRUNAV payloads
type AVURNAVPayloads map[string]AVURNAVPayload

// AVURNAVs transforms payloads to AVURNAVs
func (p AVURNAVPayloads) AVURNAVs(premar PremarInterface) AVURNAVs {
	var avurnavs AVURNAVs
	for _, v := range p {
		avurnavs = append(avurnavs, v.AVURNAV(premar))
	}
	return avurnavs
}

// List lists AVURNAVs that are currently available
func (f *AVURNAVFetcher) List() (AVURNAVs, *http.Response, error) {
	relative, err := url.Parse("avis-urgents-aux-navigateurs.html?frame=maps.php")
	if err != nil {
		return nil, nil, err
	}

	theURL := f.service.BaseURL().ResolveReference(relative)

	body := url.Values{}
	body.Add("id_categorie", strconv.Itoa(AVURNAVCategoryID))

	req, err := f.service.Client().NewRequest("POST", theURL, body)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var payloads AVURNAVPayloads
	response, err := f.service.Client().Do(req, &payloads)
	if err != nil {
		return nil, response, err
	}

	return payloads.AVURNAVs(f.service), response, err
}

// Get fetches the content of an AVURNAV from the web and returns it
func (f *AVURNAVFetcher) Get(a AVURNAV) (AVURNAV, *http.Response, error) {
	url, err := url.Parse(a.URL)
	if err != nil {
		return AVURNAV{}, nil, err
	}

	req, err := f.service.Client().NewRequest("GET", url, nil)
	if err != nil {
		return AVURNAV{}, nil, err
	}

	var buf bytes.Buffer
	response, err := f.service.Client().Do(req, &buf)
	if err != nil {
		return AVURNAV{}, response, err
	}

	return a.ParseContent(&buf), response, err
}
