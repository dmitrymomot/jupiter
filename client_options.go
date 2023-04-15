package jupiter

import (
	"net/http"
	"strings"
)

// WithHTTPClient returns a ClientOption that configures the HTTP client used by the Jupiter client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.client = client
	}
}

// WithAPIURL returns a ClientOption that configures the API URL used by the Jupiter client.
func WithAPIURL(apiURL string) ClientOption {
	return func(c *Client) {
		c.apiURL = strings.TrimRight(apiURL, "/")
	}
}

// WithEndpointQuote returns a ClientOption that configures the quote endpoint used by the Jupiter client.
func WithEndpointQuote(endpointQuote string) ClientOption {
	return func(c *Client) {
		c.endpointQuote = endpointQuote
	}
}

// WithEndpointSwap returns a ClientOption that configures the swap endpoint used by the Jupiter client.
func WithEndpointSwap(endpointSwap string) ClientOption {
	return func(c *Client) {
		c.endpointSwap = endpointSwap
	}
}

// WithEndpointPrice returns a ClientOption that configures the price endpoint used by the Jupiter client.
func WithEndpointPrice(endpointPrice string) ClientOption {
	return func(c *Client) {
		c.endpointPrice = endpointPrice
	}
}

// WithEndpointRoutesMap returns a ClientOption that configures the routes map endpoint used by the Jupiter client.
func WithEndpointRoutesMap(endpointRoutesMap string) ClientOption {
	return func(c *Client) {
		c.endpointRoutesMap = endpointRoutesMap
	}
}
