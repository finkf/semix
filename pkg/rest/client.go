package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"bitbucket.org/fflo/semix/pkg/semix"
)

// Client represents a connection to the rest service.
type Client struct {
	client *http.Client
	host   string
}

// NewClient create a new client that connects to the rest at
// a given host address.
func NewClient(host string) Client {
	return Client{
		client: new(http.Client),
		host:   host,
	}
}

// Search searches for concepts that match the given query string.
func (c Client) Search(q string) ([]*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/search?q=%s", url.QueryEscape(q))
	var cs []*semix.Concept
	err := c.get(url, &cs)
	return cs, err
}

// ParentsURL get the parent concepts searching by url.
func (c Client) ParentsURL(u string) ([]*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/info?url=%s", url.QueryEscape(u))
	var cs []*semix.Concept
	err := c.get(url, &cs)
	return cs, err
}

// ParentsID get the parent concepts searching by id.
func (c Client) ParentsID(id int) ([]*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/info?id=%d", id)
	var cs []*semix.Concept
	err := c.get(url, &cs)
	return cs, err
}

// InfoURL get the concept info searching by url.
func (c Client) InfoURL(u string) (ConceptInfo, error) {
	url := c.host + fmt.Sprintf("/info?url=%s", url.QueryEscape(u))
	var info ConceptInfo
	err := c.get(url, &info)
	return info, err
}

// InfoID get the concept info searching by id.
func (c Client) InfoID(id int) (ConceptInfo, error) {
	url := c.host + fmt.Sprintf("/info?id=%d", id)
	var info ConceptInfo
	err := c.get(url, &info)
	return info, err
}

// Get searches the index for the given query.
func (c Client) Get(q string) (Tokens, error) {
	url := c.host + fmt.Sprintf("/get?q=%s", url.QueryEscape(q))
	var ts Tokens
	err := c.get(url, &ts)
	return ts, err
}

// PutURL puts the given url into the index.
func (c Client) PutURL(u string) (Tokens, error) {
	url := c.host + fmt.Sprintf("/put?url=%s", url.QueryEscape(u))
	var ts Tokens
	err := c.get(url, &ts)
	return ts, err
}

// PutContent puts the given content into the index.
func (c Client) PutContent(r io.Reader, ct string) (Tokens, error) {
	var ts Tokens
	err := c.post(c.host+"/put", r, ct, ts)
	return ts, err
}

// Ctx returns the context of a given citation.
func (c Client) Ctx(u string, b, e, n int) (Context, error) {
	url := fmt.Sprintf("/ctx?s=%s&b=%d&e=%d&n=%d",
		url.QueryEscape(u), b, e, n)
	var ctx Context
	err := c.get(url, &ctx)
	return ctx, err
}

func (c Client) get(url string, out interface{}) error {
	res, err := c.client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(out)
}

func (c Client) post(url string, r io.Reader, ct string, out interface{}) error {
	res, err := c.client.Post(url, ct, r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(out)
}
