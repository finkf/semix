package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rest"
	"bitbucket.org/fflo/semix/pkg/say"
	"bitbucket.org/fflo/semix/pkg/semix"
	"github.com/pkg/errors"
)

// Option is a functional configuration option for the client.
type Option func(*Client)

// WithResolvers sets the resolvers for the client to use.
func WithResolvers(rs ...rest.Resolver) Option {
	return func(c *Client) {
		c.rs = rs
	}
}

// WithErrorLimits sets the errorlimits for the client to use.
func WithErrorLimits(ks ...int) Option {
	return func(c *Client) {
		c.ks = ks
		sort.Ints(c.ks)
	}
}

// WithSkip sets the query skip value.
func WithSkip(s int) Option {
	return func(c *Client) {
		c.skip = s
	}
}

// WithMax sets the query max value.
func WithMax(m int) Option {
	return func(c *Client) {
		c.max = m
	}
}

// Client represents a connection to the rest service.
type Client struct {
	client    *http.Client
	host      string
	rs        []rest.Resolver
	ks        []int
	skip, max int
}

// New create a new client that connects to the rest at
// a given host address.
func New(host string, opts ...Option) *Client {
	c := &Client{
		client: new(http.Client),
		host:   host,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Predicates searches for concepts that are connected via the given predicate.
func (c *Client) Predicates(q string) ([]*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/predicates?q=%s", url.QueryEscape(q))
	var cs []*semix.Concept
	err := c.get(url, &cs)
	return cs, err
}

// Search searches for concepts that match the given query string.
func (c *Client) Search(q string) ([]*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/search?q=%s", url.QueryEscape(q))
	var cs []*semix.Concept
	err := c.get(url, &cs)
	return cs, err
}

// ParentsURL get the parent concepts searching by url.
func (c *Client) ParentsURL(u string) ([]*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/parents?url=%s", url.QueryEscape(u))
	var cs []*semix.Concept
	err := c.get(url, &cs)
	return cs, err
}

// ParentsID get the parent concepts searching by id.
func (c *Client) ParentsID(id int) ([]*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/parents?id=%d", id)
	var cs []*semix.Concept
	err := c.get(url, &cs)
	return cs, err
}

// InfoURL gets the concept info searching by URL.
func (c *Client) InfoURL(u string) (rest.ConceptInfo, error) {
	url := c.host + fmt.Sprintf("/info?url=%s", url.QueryEscape(u))
	var info rest.ConceptInfo
	err := c.get(url, &info)
	return info, err
}

// InfoID gets the concept info searching by ID.
func (c *Client) InfoID(id int) (rest.ConceptInfo, error) {
	url := c.host + fmt.Sprintf("/info?id=%d", id)
	var info rest.ConceptInfo
	err := c.get(url, &info)
	return info, err
}

// ConceptURL gets the concept searching by URL.
func (c *Client) ConceptURL(u string) (*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/concept?url=%s", url.QueryEscape(u))
	var con semix.Concept
	err := c.get(url, &con)
	return &con, err
}

// ConceptID gets the concept searching by ID.
func (c *Client) ConceptID(id int) (*semix.Concept, error) {
	url := c.host + fmt.Sprintf("/concept?url=%d", id)
	var con semix.Concept
	err := c.get(url, &con)
	return &con, err
}

// Get searches the index for the given query.
func (c *Client) Get(q string) ([]index.Entry, error) {
	data := struct {
		Q    string
		N, S int
	}{q, c.max, c.skip}
	query, err := rest.EncodeQuery(data)
	if err != nil {
		return nil, err
	}
	url := c.host + "/get" + query
	var es []index.Entry
	err = c.get(url, &es)
	return es, err
}

// PutURL puts the given url into the index.
func (c *Client) PutURL(url string) ([]index.Entry, error) {
	return c.doPut(rest.PutData{
		URL:       url,
		Errors:    c.ks,
		Resolvers: c.rs,
	})
}

// PutLocalFile puts a local file into the index.
// This only works if the server has access to the same file system as the client.
// PutLocalFile calculates the absolute path for the given file.
func (c *Client) PutLocalFile(path string) ([]index.Entry, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return c.doPut(rest.PutData{
		URL:       abs,
		Local:     true,
		Errors:    c.ks,
		Resolvers: c.rs,
	})
}

// PutString puts the given string into the index.
func (c *Client) PutString(content, ct string) ([]index.Entry, error) {
	return c.doPut(rest.PutData{
		Errors:      c.ks,
		Resolvers:   c.rs,
		Content:     content,
		ContentType: ct,
	})
}

// PutContent puts the given content into the index.
func (c *Client) PutContent(r io.Reader, url, ct string) ([]index.Entry, error) {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return c.doPut(rest.PutData{
		URL:         url,
		Errors:      c.ks,
		Resolvers:   c.rs,
		Content:     string(content),
		ContentType: ct,
	})
}

func (c *Client) doPut(data rest.PutData) ([]index.Entry, error) {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(data); err != nil {
		return nil, errors.Wrapf(err, "cannot encode JSON")
	}
	var es []index.Entry
	err := c.post(c.host+"/put", b, "application/json", &es)
	return es, errors.Wrapf(err, "cannot put data")
}

// Ctx returns the context of a given citation.
func (c *Client) Ctx(u string, b, e, n int) (rest.Context, error) {
	url := fmt.Sprintf("%s/ctx?url=%s&b=%d&e=%d&n=%d",
		c.host, url.QueryEscape(u), b, e, n)
	var ctx rest.Context
	err := c.get(url, &ctx)
	return ctx, err
}

// Flush flushes the index.
func (c *Client) Flush() error {
	url := fmt.Sprintf("%s/flush", c.host)
	return c.get(url, nil)
}

// DumpFile returns the dump file of the requested url.
func (c *Client) DumpFile(u string) (rest.DumpFileContent, error) {
	url := fmt.Sprintf("%s/dump?url=%s", c.host, url.QueryEscape(u))
	var data rest.DumpFileContent
	err := c.get(url, &data)
	return data, errors.Wrapf(err, "cannot dump file: %s", u)
}

func (c *Client) get(url string, out interface{}) error {
	say.Debug("sending request [%s] %s", http.MethodGet, url)
	res, err := c.client.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return fmt.Errorf("invalid status: %s", res.Status)
	}
	return decodeFromJSON(res.Body, out)
}

func (c *Client) post(url string, r io.Reader, ct string, out interface{}) error {
	say.Debug("sending request [%s] %s", http.MethodPost, url)
	res, err := c.client.Post(url, ct, r)
	if err != nil {
		return errors.Wrapf(err, "cannot send post request")
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return fmt.Errorf("invalid status: %s", res.Status)
	}
	return decodeFromJSON(res.Body, out)
}

func decodeFromJSON(in io.Reader, out interface{}) error {
	// in = io.TeeReader(in, os.Stdout)
	return json.NewDecoder(in).Decode(out)
}
