package avurnav

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	BASE_PREMAR = "https://premar.antoine-augusti.fr"
)

// AVURNAVFetcher fetches AVURNAVs on the Préfet
// Maritime websites
type AVURNAVFetcher struct {
	service PremarInterface
}

// AVURNAVPayload is used to decode AVURNAVs from
// the Préfet Maritime websites
type AVURNAVPayload struct {
	Title      string  `json:"title"`
	ValidFrom  string  `json:"valid_from"`
	ValidUntil string  `json:"valid_until"`
	Latitude   float32 `json:"latitude"`
	Longitude  float32 `json:"longitude"`
	URL        string  `json:"url"`
	Number     string  `json:"number"`
}

// AVURNAV transforms a payload to a proper AVURNAV
func (p AVURNAVPayload) AVURNAV(premar PremarInterface) AVURNAV {
	from, to := p.ValidFrom, p.ValidUntil

	avurnav := AVURNAV{
		Number:       p.Number,
		Title:        strings.TrimSpace(p.Title),
		Latitude:     p.Latitude,
		Longitude:    p.Longitude,
		URL:          p.URL,
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

func (p AVURNAVPayload) parseInt(str string) int {
	s, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return s
}

// AVURNAV represents an AVURNAV
type AVURNAV struct {
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
	blocks := scrape.FindAllNested(root, scrape.ByClass("col-12"))
	divs := scrape.FindAllNested(blocks[1], scrape.ByTag(atom.Div))
	a.Content = scrape.Text(divs[3])
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
type AVURNAVPayloads []AVURNAVPayload

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
	relative, err := url.Parse("?region=" + strings.ToLower(f.service.Region()))
	if err != nil {
		return nil, nil, err
	}
	base, _ := url.Parse(BASE_PREMAR)

	theURL := base.ResolveReference(relative)

	req, err := f.service.Client().NewRequest("GET", theURL, nil)
	if err != nil {
		return nil, nil, err
	}

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
