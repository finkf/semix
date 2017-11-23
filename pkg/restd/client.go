package restd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"bitbucket.org/fflo/semix/pkg/semix"
)

// Client represents a connection to the restd service.
type Client struct {
	client *http.Client
	host   string
}

// NewClient create a new client that connects to the restd at
// a given host address.
func NewClient(host string) Client {
	return Client{
		client: new(http.Client),
		host:   host,
	}
}

// Search searches for concepts that match the given query string.
func (c Client) Search(q string) ([]semix.Concept, error) {
	url := c.host + fmt.Sprintf("/search?q=%s", url.QueryEscape(q))
	var cs []semix.Concept
	if err := c.get(url, &cs); err != nil {
		return nil, err
	}
	return cs, nil
}

func (c Client) get(url string, res interface{}) error {
	r, err := c.client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(res)
}
