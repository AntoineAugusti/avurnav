package avurnav

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

// AVURNAVFetcher fetches AVURNAVs on the Préfet
// Maritime websites
type AVURNAVFetcher struct {
	service PremarInterface
}

// AVURNAVPayload
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

	return AVURNAV{
		ID:           p.parseInt(p.ID),
		Number:       p.Number,
		Title:        p.Title,
		Latitude:     p.parseFloat(p.Latitude),
		Longitude:    p.parseFloat(p.Longitude),
		City:         &p.City,
		URL:          premar.BaseURL().ResolveReference(relative).String(),
		ValidFrom:    &from,
		ValidTo:      &to,
		PreMarRegion: premar.Region(),
	}
}

func (p AVURNAVPayload) matchDates(str string) (from, to string) {
	re := regexp.MustCompile(`^En vigueur du : (\d{2}\/\d{2}\/\d{4}|Indéterminé) au (\d{2}\/\d{2}\/\d{4}|Indéterminé)$`)
	matches := re.FindStringSubmatch(str)
	return p.cleanDate(matches[0]), p.cleanDate(matches[1])
}

func (p AVURNAVPayload) cleanDate(str string) string {
	if str == "Indéterminé" {
		return ""
	}
	return str
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
	ID           int     `json:"id"`
	Number       string  `json:"number"`
	Title        string  `json:"title"`
	Latitude     float32 `json:"latitude"`
	Longitude    float32 `json:"longitude"`
	City         *string `json:"city"`
	URL          string  `json:"url"`
	ValidFrom    *string `json:"valid_from"`
	ValidTo      *string `json:"valid_to"`
	PreMarRegion string  `json:"premar_region"`
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

	url := f.service.BaseURL().ResolveReference(relative)

	req, err := f.service.Client().NewRequest("GET", url, nil)
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
