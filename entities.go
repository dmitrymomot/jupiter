package jupiter

import (
	"strconv"
)

// MarketInfo is a market info object structure.
type MarketInfo struct {
	ID                 string  `json:"id"`
	Label              string  `json:"label"`
	InputMint          string  `json:"inputMint"`
	OutputMint         string  `json:"outputMint"`
	NotEnoughLiquidity bool    `json:"notEnoughLiquidity"`
	InAmount           string  `json:"inAmount"`
	OutAmount          string  `json:"outAmount"`
	MinInAmount        string  `json:"minInAmount,omitempty"`
	MinOutAmount       string  `json:"minOutAmount,omitempty"`
	PriceImpactPct     float64 `json:"priceImpactPct"`
	LpFee              *Fee    `json:"lpFee"`
	PlatformFee        *Fee    `json:"platformFee"`
}

// Fee is a fee object structure.
type Fee struct {
	Amount string  `json:"amount"`
	Mint   string  `json:"mint"`
	Pct    float64 `json:"pct"`
}

// Route is a route object structure.
type Route struct {
	InAmount             string       `json:"inAmount"`
	OutAmount            string       `json:"outAmount"`
	PriceImpactPct       float64      `json:"priceImpactPct"`
	MarketInfos          []MarketInfo `json:"marketInfos"`
	Amount               string       `json:"amount"`
	SlippageBps          int64        `json:"slippageBps"`          // minimum: 0, maximum: 10000
	OtherAmountThreshold string       `json:"otherAmountThreshold"` // The threshold for the swap based on the provided slippage: when swapMode is ExactIn the minimum out amount, when swapMode is ExactOut the maximum in amount
	SwapMode             string       `json:"swapMode"`
	Fees                 *struct {
		SignatureFee             int64   `json:"signatureFee"`             // This inidicate the total amount needed for signing transaction(s). Value in lamports.
		OpenOrdersDeposits       []int64 `json:"openOrdersDeposits"`       // This inidicate the total amount needed for deposit of serum order account(s). Value in lamports.
		AtaDeposits              []int64 `json:"ataDeposits"`              // This inidicate the total amount needed for deposit of associative token account(s). Value in lamports.
		TotalFeeAndDeposits      int64   `json:"totalFeeAndDeposits"`      // This indicate the total lamports needed for fees and deposits above.
		MinimumSolForTransaction int64   `json:"minimumSOLForTransaction"` // This inidicate the minimum lamports needed for transaction(s). Might be used to create wrapped SOL and will be returned when the wrapped SOL is closed. Also ensures rent exemption of the wallet.
	} `json:"fees,omitempty"`
}

// Price is a price object structure.
type Price struct {
	ID            string  `json:"id"`            // Address of the token
	MintSymbol    string  `json:"mintSymbol"`    // Symbol of the token
	VsToken       string  `json:"vsToken"`       // Address of the token to compare against
	VsTokenSymbol string  `json:"vsTokenSymbol"` // Symbol of the token to compare against
	Price         float64 `json:"price"`         // Price of the token in relation to the vsToken. Default to 1 unit of the token worth in USDC if vsToken is not specified.
}

// PriceMap is a price map objects structure.
type PriceMap map[string]Price

// QuoteParams are the parameters for a quote request.
type QuoteParams struct {
	InputMint  string `url:"inputMint"`  // required
	OutputMint string `url:"outputMint"` // required
	Amount     uint64 `url:"amount"`     // required

	SwapMode            string `url:"swapMode,omitempty"` // Swap mode, default is ExactIn; Available values : ExactIn, ExactOut.
	SlippageBps         uint64 `url:"slippageBps,omitempty"`
	FeeBps              uint64 `url:"feeBps,omitempty"`              // Fee BPS (only pass in if you want to charge a fee on this swap)
	OnlyDirectRoutes    bool   `url:"onlyDirectRoutes,omitempty"`    // Only return direct routes (no hoppings and split trade)
	AsLegacyTransaction bool   `url:"asLegacyTransaction,omitempty"` // Only return routes that can be done in a single legacy transaction. (Routes might be limited)
	UserPublicKey       string `url:"userPublicKey,omitempty"`       // Public key of the user (only pass in if you want deposit and fee being returned, might slow down query)
}

// QuoteResponse is the response from a quote request.
type QuoteResponse []Route

// GetBestRoute returns the best route from a quote response.
func (q QuoteResponse) GetBestRoute() (Route, error) {
	if len(q) == 0 {
		return Route{}, ErrNoRoute
	}
	if len(q) == 1 {
		return q[0], nil
	}

	bestRoute := q[0]
	for _, route := range q {
		if route.PriceImpactPct < bestRoute.PriceImpactPct {
			bestRoute = route
		}
	}
	return bestRoute, nil
}

// SwapParams are the parameters for a swap request.
type SwapParams struct {
	Route                         Route  `json:"route"`                   // required
	UserPublicKey                 string `json:"userPublicKey,omitempty"` // required
	WrapUnwrapSol                 *bool  `json:"wrapUnwrapSOL,omitempty"`
	FeeAccount                    string `json:"feeAccount,omitempty"`                    // Fee token account for the platform fee (only pass in if you set a feeBps), the mint is outputMint for the default swapMode.ExactOut and inputMint for swapMode.ExactIn.
	AsLegacyTransaction           *bool  `json:"asLegacyTransaction,omitempty"`           // Request a legacy transaction rather than the default versioned transaction, needs to be paired with a quote using asLegacyTransaction otherwise the transaction might be too large.
	ComputeUnitPriceMicroLamports *int64 `json:"computeUnitPriceMicroLamports,omitempty"` // Compute unit price to prioritize the transaction, the additional fee will be compute unit consumed * computeUnitPriceMicroLamports.
	DestinationWallet             string `json:"destinationWallet,omitempty"`             // Public key of the wallet that will receive the output of the swap, this assumes the associated token account exists, currently adds a token transfer.
}

// SwapResponse is the response from a swap request.
type SwapResponse struct {
	SwapTransaction string `json:"swapTransaction"` // base64 encoded transaction string
}

// PriceParams are the parameters for a price request.
type PriceParams struct {
	IDs      string  `url:"ids"`                // required; Symbol or address of a token, (e.g. SOL or EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v). Use `,` to query multiple tokens, e.g. (sol,btc,mer,...)
	VsToken  string  `url:"vsToken,omitempty"`  // optional; Default to USDC. Symbol or address of a token, (e.g. SOL or EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v).
	VsAmount float64 `url:"vsAmount,omitempty"` // optional; Unit amount of specified input token. Default to 1.
}

// IndexedRoutesMap is a map of routes indexed by the route ID.
type IndexedRoutesMap struct {
	MintKeys        []string         `json:"mintKeys"`        // All the mints that are indexed to match in indexedRouteMap.
	IndexedRouteMap map[string][]int `json:"indexedRouteMap"` // All the possible route and their corresponding output mints.
}

// GetRoutesForMint returns the routes for a given mint.
func (r *IndexedRoutesMap) GetRoutesForMint(mint string) []string {
	// Find index of mint in mintKeys.
	var mintKeys []int
	for key, val := range r.MintKeys {
		if val == mint {
			mintKeys = r.IndexedRouteMap[strconv.Itoa(key)]
		}
	}

	// Find the mint in mintKeys.
	result := make([]string, 0, len(mintKeys))
	for _, key := range mintKeys {
		result = append(result, r.MintKeys[key])
	}

	return result
}

// BestSwapParams contains the parameters for the best swap route.
type BestSwapParams struct {
	UserPublicKey        string // user base58 encoded public key
	DestinationPublicKey string // destination base58 encoded public key (optional)
	FeeAmount            uint64 // fee amount in token basis points (optional)
	FeeAccount           string // fee token account for the platform fee (only pass in if you set a FeeAmount).
	InputMint            string // input mint
	OutputMint           string // output mint
	Amount               uint64 // amount of output token
	SwapMode             string // swap mode, default: ExactIn (Available: ExactIn, ExactOut)
}

// ExchangeRateParams contains the parameters for the exchange rate request.
type ExchangeRateParams struct {
	InputMint  string // input token mint
	OutputMint string // output token mint
	Amount     uint64 // amount of token, depending on the swap mode
	SwapMode   string // swap mode, default: ExactOut (Available: ExactIn, ExactOut)
}

// ExchangeRate returns the exchange rate for a given input mint, output mint and amount.
type Rate struct {
	InputMint  string `json:"inputMint"`  // input token mint
	OutputMint string `json:"outputMint"` // output token mint
	InAmount   uint64 `json:"inAmount"`   // amount of input token
	OutAmount  uint64 `json:"outAmount"`  // amount of output token
}
