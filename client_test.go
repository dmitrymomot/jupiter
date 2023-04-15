package jupiter_test

import (
	"testing"

	"github.com/dmitrymomot/jupiter"
	"github.com/dmitrymomot/jupiter/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	wSolMint = "So11111111111111111111111111111111111111112"
	usdcMint = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
)

func TestQuote(t *testing.T) {
	c := jupiter.NewClient()
	quotes, err := c.Quote(jupiter.QuoteParams{
		InputMint:        wSolMint,
		OutputMint:       usdcMint,
		Amount:           100000,
		OnlyDirectRoutes: true,
		SwapMode:         jupiter.SwapModeExactOut,
	})
	require.NoError(t, err)
	require.NotEmpty(t, quotes)
	require.GreaterOrEqual(t, len(quotes), 1)

	quote := quotes[0]
	// utils.PrettyPrint(quote)

	assert.Equal(t, wSolMint, quote.MarketInfos[0].InputMint)
	assert.Equal(t, usdcMint, quote.MarketInfos[0].OutputMint)
	assert.Equal(t, "100000", quote.Amount)
}

func TestSwap(t *testing.T) {
	c := jupiter.NewClient()
	var route jupiter.Route

	t.Run("get best route", func(t *testing.T) {
		quotes, err := c.Quote(jupiter.QuoteParams{
			InputMint:        wSolMint,
			OutputMint:       usdcMint,
			Amount:           100000,
			OnlyDirectRoutes: false,
		})
		require.NoError(t, err)
		require.NotEmpty(t, quotes)
		require.GreaterOrEqual(t, len(quotes), 1)

		route, err = quotes.GetBestRoute()
		require.NoError(t, err)
		if err == nil {
			require.NotEmpty(t, route)
		}
	})

	t.Run("create swap tx", func(t *testing.T) {
		swapTx, err := c.Swap(jupiter.SwapParams{
			UserPublicKey: "8HwPMNxtFDrvxXn1fJsAYB258TnA6Ydr1DWCtVYgRW4W",
			Route:         route,
			WrapUnwrapSol: utils.Pointer(true),
		})
		require.NoError(t, err)
		require.NotEmpty(t, swapTx)

		// t.Log(swapTx)
	})
}

func TestPrice(t *testing.T) {
	c := jupiter.NewClient()

	price, err := c.Price(jupiter.PriceParams{
		IDs:     "SOL",
		VsToken: usdcMint,
	})
	require.NoError(t, err)
	require.NotEmpty(t, price)
	assert.Equal(t, "So11111111111111111111111111111111111111112", price["SOL"].ID)
	assert.Equal(t, "SOL", price["SOL"].MintSymbol)
	assert.Equal(t, usdcMint, price["SOL"].VsToken)

	// utils.PrettyPrint(price)
}

func TestRoutesMap(t *testing.T) {
	c := jupiter.NewClient()

	routesMap, err := c.RoutesMap(true)
	require.NoError(t, err)
	require.NotEmpty(t, routesMap)
	assert.Greater(t, len(routesMap.GetRoutesForMint(usdcMint)), 0)
}

func TestExchangeRate(t *testing.T) {
	c := jupiter.NewClient()

	var amount uint64 = 100000
	exchangeRate, err := c.ExchangeRate(jupiter.ExchangeRateParams{
		InputMint:  wSolMint,
		OutputMint: usdcMint,
		Amount:     amount,
		SwapMode:   jupiter.SwapModeExactOut,
	})
	require.NoError(t, err)
	require.NotEmpty(t, exchangeRate)
	// utils.PrettyPrint(exchangeRate)

	assert.Equal(t, wSolMint, exchangeRate.InputMint)
	assert.Equal(t, usdcMint, exchangeRate.OutputMint)
	assert.EqualValues(t, amount, exchangeRate.OutAmount)
}

func TestBestSwap(t *testing.T) {
	c := jupiter.NewClient()

	var amount uint64 = 100000
	bestSwap, err := c.BestSwap(jupiter.BestSwapParams{
		UserPublicKey: "8HwPMNxtFDrvxXn1fJsAYB258TnA6Ydr1DWCtVYgRW4W",
		InputMint:     wSolMint,
		OutputMint:    usdcMint,
		Amount:        amount,
		SwapMode:      jupiter.SwapModeExactIn,
	})
	require.NoError(t, err)
	require.NotEmpty(t, bestSwap)
	utils.PrettyPrint(bestSwap)
}
