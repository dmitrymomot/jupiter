package jupiter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/dmitrymomot/jupiter/utils"
)

const (
	// ContentTypeJSON is the content type for JSON.
	ContentTypeJSON = "application/json"
)

type (
	// Client is a Jupiter client that can be used to make requests to the Jupiter API.
	Client struct {
		client *http.Client

		apiURL            string
		endpointQuote     string
		endpointSwap      string
		endpointPrice     string
		endpointRoutesMap string
	}

	// ClientOption is a function that can be used to configure a Jupiter client.
	ClientOption func(*Client)

	// Response is a generic response structure.
	Response struct {
		Data        json.RawMessage `json:"data"`
		TimeTaken   float64         `json:"timeTaken"`
		ContextSlot int64           `json:"contextSlot"`
	}
)

// NewClient returns a new Jupiter client.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},

		apiURL:            "https://quote-api.jup.ag/v4",
		endpointQuote:     "/quote",
		endpointSwap:      "/swap",
		endpointPrice:     "/price",
		endpointRoutesMap: "/indexed-route-map",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// get makes a GET request to the specified endpoint with the given parameters.
// It returns the response as is without parsing or any error encountered.
// The caller is responsible for closing the response body.
func (c *Client) get(endpoint string, params interface{}) (*http.Response, error) {
	uv, err := utils.StructToUrlValues(params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to url values: %w", err)
	}

	parsedURL, err := url.Parse(c.apiURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if len(uv) > 0 {
		parsedURL.RawQuery = uv.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}
	req.Header.Set("Accept", ContentTypeJSON)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request: %w", err)
	}

	return resp, nil
}

// postRaw makes a POST request to the specified URL with the given parameters.
// It returns the response as is without parsing or any error encountered.
// The caller is responsible for closing the response body.
func (c *Client) post(endpoint string, params interface{}) (*http.Response, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal POST params: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.apiURL+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Set("Content-Type", ContentTypeJSON)
	req.Header.Set("Accept", ContentTypeJSON)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make POST request: %w", err)
	}

	return resp, nil
}

// parseResponse parses the response body into the given response structure.
func (c *Client) parseResponse(resp *http.Response) (json.RawMessage, error) {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data, nil
}

// Quote returns a quote for a given input mint, output mint and amount
func (c *Client) Quote(params QuoteParams) (QuoteResponse, error) {
	resp, err := c.get(c.endpointQuote, params)
	if err != nil {
		return nil, fmt.Errorf("failed to make quote request: %w", err)
	}

	data, err := c.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse quote response: %w", err)
	}

	var quotes QuoteResponse
	if err := json.Unmarshal(data, &quotes); err != nil {
		return nil, fmt.Errorf("failed to parse quote response: %w", err)
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quotes returned")
	}

	return quotes, nil
}

// Swap returns swap base64 serialized transaction for a route.
// The caller is responsible for signing the transactions.
func (c *Client) Swap(params SwapParams) (string, error) {
	resp, err := c.post(c.endpointSwap, params)
	if err != nil {
		return "", fmt.Errorf("failed to make swap request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response SwapResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.SwapTransaction, nil
}

// Price returns simple price for a given input mint, output mint and amount.
func (c *Client) Price(params PriceParams) (PriceMap, error) {
	resp, err := c.get(c.endpointPrice, params)
	if err != nil {
		return nil, fmt.Errorf("failed to make price request: %w", err)
	}

	data, err := c.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price response: %w", err)
	}

	var price PriceMap
	if err := json.Unmarshal(data, &price); err != nil {
		return nil, fmt.Errorf("failed to parse price response: %w", err)
	}

	return price, nil
}

// RoutesMap returns a hash map, input mint as key and an array of valid output mint as values,
// token mints are indexed to reduce the file size.
func (c *Client) RoutesMap(onlyDirectRoutes bool) (IndexedRoutesMap, error) {
	resp, err := c.get(c.endpointRoutesMap, url.Values{
		"onlyDirectRoutes": []string{strconv.FormatBool(onlyDirectRoutes)},
	})
	if err != nil {
		return IndexedRoutesMap{}, fmt.Errorf("failed to make routes map request: %w", err)
	}

	var routesMap IndexedRoutesMap
	if err := json.NewDecoder(resp.Body).Decode(&routesMap); err != nil {
		return IndexedRoutesMap{}, fmt.Errorf("failed to parse routes map response: %w", err)
	}

	return routesMap, nil
}

// BestSwap returns the ebase64 encoded transaction for the best swap route
// for a given input mint, output mint and amount.
// Default swap mode: ExactOut, so the amount is the amount of output token.
// Default wrap unwrap sol: true
func (c *Client) BestSwap(params BestSwapParams) (string, error) {
	if params.SwapMode == "" {
		params.SwapMode = SwapModeExactIn
	}
	routes, err := c.Quote(QuoteParams{
		InputMint:        params.InputMint,
		OutputMint:       params.OutputMint,
		Amount:           params.Amount,
		FeeBps:           params.FeeAmount,
		SwapMode:         params.SwapMode,
		OnlyDirectRoutes: false,
	})
	if err != nil {
		return "", err
	}

	route, err := routes.GetBestRoute()
	if err != nil {
		return "", err
	}

	swap, err := c.Swap(SwapParams{
		Route:               route,
		UserPublicKey:       params.UserPublicKey,
		DestinationWallet:   params.DestinationPublicKey,
		FeeAccount:          params.FeeAccount,
		WrapUnwrapSol:       utils.Pointer(true),
		AsLegacyTransaction: utils.Pointer(true),
	})
	if err != nil {
		return "", err
	}

	return swap, nil
}

// ExchangeRate returns the exchange rate for a given input mint, output mint and amount.
// Default swap mode: ExactOut, so the amount is the amount of output token.
func (c *Client) ExchangeRate(params ExchangeRateParams) (Rate, error) {
	result := Rate{
		InputMint:  params.InputMint,
		OutputMint: params.OutputMint,
	}
	routes, err := c.Quote(QuoteParams{
		InputMint:        params.InputMint,
		OutputMint:       params.OutputMint,
		Amount:           params.Amount,
		SwapMode:         params.SwapMode,
		OnlyDirectRoutes: false,
	})
	if err != nil {
		return result, err
	}

	route, err := routes.GetBestRoute()
	if err != nil {
		return result, err
	}

	inAmount, err := strconv.ParseInt(route.InAmount, 10, 64)
	if err != nil {
		return result, fmt.Errorf("failed to parse in amount: %w", err)
	}
	outAmount, err := strconv.ParseInt(route.OutAmount, 10, 64)
	if err != nil {
		return result, fmt.Errorf("failed to parse out amount: %w", err)
	}

	result.InAmount = uint64(inAmount)
	result.OutAmount = uint64(outAmount)

	return result, nil
}
