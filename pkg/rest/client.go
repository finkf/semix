package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"bitbucket.org/fflo/semix/pkg/index"
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

// Predicates searches for concepts that are connected via the given predicate.
func (c Client) Predicates(q string) ([]*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/predicates?q=%s", url.QueryEscape(q))
	var cs []*semix.Concept
	err := c.get(url, &cs)
	return cs, err
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
	url := c.host + fmt.Sprintf("/parents?url=%s", url.QueryEscape(u))
	var cs []*semix.Concept
	err := c.get(url, &cs)
	return cs, err
}

// ParentsID get the parent concepts searching by id.
func (c Client) ParentsID(id int) ([]*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/parents?id=%d", id)
	var cs []*semix.Concept
	err := c.get(url, &cs)
	return cs, err
}

// InfoURL gets the concept info searching by URL.
func (c Client) InfoURL(u string) (ConceptInfo, error) {
	url := c.host + fmt.Sprintf("/info?url=%s", url.QueryEscape(u))
	var info ConceptInfo
	err := c.get(url, &info)
	return info, err
}

// InfoID gets the concept info searching by ID.
func (c Client) InfoID(id int) (ConceptInfo, error) {
	url := c.host + fmt.Sprintf("/info?id=%d", id)
	var info ConceptInfo
	err := c.get(url, &info)
	return info, err
}

// ConceptURL gets the concept searching by URL.
func (c Client) ConceptURL(u string) (*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/concept?url=%s", url.QueryEscape(u))
	var con semix.Concept
	err := c.get(url, &con)
	return &con, err
}

// ConceptID gets the concept searching by ID.
func (c Client) ConceptID(id int) (*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/concept?url=%d", id)
	var con semix.Concept
	err := c.get(url, &con)
	return &con, err
}

// Get searches the index for the given query.
func (c Client) Get(q string, n, s int) ([]index.Entry, error) {
	data := struct {
		Q    string
		N, S int
	}{q, n, s}
	query, err := EncodeQuery(data)
	if err != nil {
		return nil, err
	}
	url := c.host + "/get" + query
	var es []index.Entry
	err = c.get(url, &es)
	return es, err
}

// PutURL puts the given url into the index.
func (c Client) PutURL(url string, ls []int, rs []Resolver) ([]index.Entry, error) {
	return c.doPut(PutData{
		URL:       url,
		Errors:    ls,
		Resolvers: rs,
	})
}

// PutLocalFile puts a local file into the index.
// This only works if the server has access to the same file system as the client.
// PutLocalFile calculates the absolute path for the given file.
func (c Client) PutLocalFile(path string, ls []int, rs []Resolver) ([]index.Entry, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return c.doPut(PutData{
		URL:       abs,
		Local:     true,
		Errors:    ls,
		Resolvers: rs,
	})
}

// PutContent puts the given content into the index.
func (c Client) PutContent(r io.Reader, url, ct string, ls []int, rs []Resolver) ([]index.Entry, error) {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return c.doPut(PutData{
		URL:         url,
		Errors:      ls,
		Resolvers:   rs,
		Content:     string(content),
		ContentType: ct,
	})
}

func (c Client) doPut(data PutData) ([]index.Entry, error) {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(data); err != nil {
		return nil, err
	}
	var es []index.Entry
	var err error
	err = c.post(c.host+"/put", b, "application/json", &es)
	return es, err
}

// Ctx returns the context of a given citation.
func (c Client) Ctx(u string, b, e, n int) (Context, error) {
	url := fmt.Sprintf("%s/ctx?url=%s&b=%d&e=%d&n=%d",
		c.host, url.QueryEscape(u), b, e, n)
	var ctx Context
	err := c.get(url, &ctx)
	return ctx, err
}

// DumpFile returns the dump file of the requested url.
func (c Client) DumpFile(u string) (DumpFileContent, error) {
	url := fmt.Sprintf("%s/dump?url=%s", c.host, url.QueryEscape(u))
	var data DumpFileContent
	err := c.get(url, &data)
	return data, err
}

func (c Client) get(url string, out interface{}) error {
	log.Printf("sending request [%s] %s", http.MethodGet, url)
	res, err := c.client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return fmt.Errorf("invalid status: %s", res.Status)
	}
	return decodeFromJSON(res.Body, out)
}

func (c Client) post(url string, r io.Reader, ct string, out interface{}) error {
	log.Printf("sending request [%s] %s", http.MethodPost, url)
	res, err := c.client.Post(url, ct, r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return fmt.Errorf("invalid status: %s", res.Status)
	}
	return decodeFromJSON(res.Body, out)
}

func decodeFromJSON(in io.Reader, out interface{}) error {
	// in = io.TeeReader(in, os.Stdout)
	return json.NewDecoder(in).Decode(out)
}
