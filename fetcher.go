package avurnav

import (
	"net/http"
	"net/url"
)

// AVURNAVFetcher fetches AVURNAVs on the Préfet
// Maritime websites
type AVURNAVFetcher struct {
	service PremarInterface
}

// AVRUNAVItem
type AVRUNAVItem struct {
	ID        string `json:"id_centre"`
	Number    string `json:"numero"`
	Label     string `json:"label"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	City      string `json:"localite"`
	URL       string `json:"url"`
	Dates     string `json:"dates"`
}

// AVRUNAVs represents multiple AVRUNAV items
type AVRUNAVs map[string]AVRUNAVItem

// List lists AVURNAVs that are currently available
func (f *AVURNAVFetcher) List() (AVRUNAVs, *http.Response, error) {
	relative, err := url.Parse("avis-urgents-aux-navigateurs.html?frame=maps.php")
	if err != nil {
		return nil, nil, err
	}

	url := f.service.BaseURL().ResolveReference(relative)

	req, err := f.service.Client().NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	var res AVRUNAVs
	response, err := f.service.Client().Do(req, &res)
	if err != nil {
		return nil, response, err
	}

	return res, response, err
}
