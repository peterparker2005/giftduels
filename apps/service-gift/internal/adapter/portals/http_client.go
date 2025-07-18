package portals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type NFTResult struct {
	FloorPrice string `json:"floor_price"`
	Price      string `json:"price"`
}

type NFTResponse struct {
	Results []NFTResult `json:"results"`
}

const (
	defaultBaseURL      = "https://portals-market.com"
	defaultHTTPTimeout  = 5 * time.Second
	defaultBackoffDelay = 300 * time.Millisecond
	maxRetries          = 5
)

type HTTPClient struct {
	baseURL     string
	authHeader  string
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	logger      *logger.Logger
}

func NewHTTPClient(authHeader string, logger *logger.Logger) *HTTPClient {
	//nolint:mnd // 3 req/s, burst = 1
	limiter := rate.NewLimiter(rate.Every(time.Second/3), 1)

	return &HTTPClient{
		baseURL: defaultBaseURL,
		// authHeader:  authHeader,
		authHeader:  "tma query_id=AAEdTxk2AwAAAB1PGTZXtyrZ&user=%7B%22id%22%3A7350079261%2C%22first_name%22%3A%22pp%22%2C%22last_name%22%3A%22%22%2C%22username%22%3A%22peterparkish%22%2C%22language_code%22%3A%22en%22%2C%22allows_write_to_pm%22%3Atrue%2C%22photo_url%22%3A%22https%3A%5C%2F%5C%2Ft.me%5C%2Fi%5C%2Fuserpic%5C%2F320%5C%2FjMwTE1p_IMe6se6v6t6X8uaS1ymy2hHPJ1Oqt3b13hES-84zfc1MJCUrxxLDLgap.svg%22%7D&auth_date=1752425809&signature=ZJxYc3GVBTG5a6xcx3JxpOqYtevxAju3PMQx40R42L9ZwKfkFdFGy0xk2Y5eyooby-dh-DYb3OQMDXBc0369AA&hash=451894c5ac74cc45a0c3eb75cf994f918534632239deda3946bf54b4dae78565",
		rateLimiter: limiter,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
			Transport: &limiterTransport{
				limiter: limiter,
				rt:      http.DefaultTransport,
			},
		},
		logger: logger,
	}
}

type limiterTransport struct {
	limiter *rate.Limiter
	rt      http.RoundTripper
}

func (t *limiterTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// блокируемся на token‑bucket перед каждым запросом
	if err := t.limiter.Wait(req.Context()); err != nil {
		return nil, err
	}
	return t.rt.RoundTrip(req)
}

func (c *HTTPClient) SearchNFTs(
	ctx context.Context,
	collection, model, symbol, backdrop string,
) (*NFTResponse, error) {
	// собираем URL
	u, _ := url.Parse(c.baseURL + "/api/nfts/search")
	q := u.Query()
	q.Set("offset", "0")
	q.Set("limit", "1")
	q.Set("status", "listed")
	q.Set("sort_by", "price asc")
	q.Set("filter_by_collections", collection)
	q.Set("filter_by_models", model)
	q.Set("filter_by_symbols", symbol)
	q.Set("filter_by_backdrops", backdrop)
	u.RawQuery = q.Encode()

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		c.logger.Info("SearchNFTs attempt",
			zap.Int("n", attempt),
			zap.String("url", u.String()),
		)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Authorization", c.authHeader)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			c.logger.Warn("http.Do failed", zap.Error(err))
			continue
		}
		defer resp.Body.Close()

		// если нас запоролили — ждём время из заголовков, а не просто backoff
		if resp.StatusCode == http.StatusTooManyRequests {
			waitSec := parseRetryAfter(resp.Header)
			c.logger.Warn(
				"429 получен — ждём перед новой попыткой",
				zap.Int("retry_after_s", waitSec),
			)
			select {
			case <-time.After(time.Duration(waitSec) * time.Second):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status: %s", resp.Status)
			c.logger.Error("unexpected status", zap.String("status", resp.Status))
			break
		}

		var r NFTResponse
		if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		if len(r.Results) == 0 {
			return nil, errors.New("no listings found")
		}
		return &r, nil
	}

	return nil, fmt.Errorf("all attempts failed, last error: %w", lastErr)
}

// parseRetryAfter: сначала смотрим Retry‑After, потом X-RateLimit-Reset, иначе базовый backoff + джиттер.
func parseRetryAfter(h http.Header) int {
	if ra := h.Get("Retry-After"); ra != "" {
		if sec, err := strconv.Atoi(ra); err == nil {
			return sec
		}
		if t, err := http.ParseTime(ra); err == nil {
			secs := int(time.Until(t).Seconds()) + 1
			if secs > 0 {
				return secs
			}
		}
	}
	//nolint:canonicalheader // we need to parse the header
	if reset := h.Get("X-RateLimit-Reset"); reset != "" {
		if sec, err := strconv.Atoi(reset); err == nil {
			return sec
		}
	}
	// fallback: defaultBackoffDelay + jitter [0..defaultBackoffDelay)
	baseMs := int(defaultBackoffDelay / time.Millisecond)
	//nolint:gosec // we need to generate a random number
	jitter := rand.IntN(baseMs)
	// переводим в секунды и добавляем 1s buffer
	return (baseMs+jitter)/1000 + 1
}

// URL: https://portals-market.com/api/nfts/search?offset=0&limit=20&filter_by_backdrops=Hunter+Green&filter_by_collections=Lol+Pop&filter_by_models=Angelina&filter_by_symbols=Feather&sort_by=price+asc&status=listed

// :method: GET
// :scheme: https
// :authority: portals-market.com
// :path: /api/nfts/search?offset=0&limit=20&filter_by_backdrops=Hunter+Green&filter_by_collections=Lol+Pop&filter_by_models=Angelina&filter_by_symbols=Feather&sort_by=price+asc&status=listed
// Accept: application/json, text/plain, */*
// Accept-Encoding: gzip, deflate, br
// Accept-Language: en-US,en;q=0.9
// Authorization: tma query_id=AAEdTxk2AwAAAB1PGTZXtyrZ&user=%7B%22id%22%3A7350079261%2C%22first_name%22%3A%22pp%22%2C%22last_name%22%3A%22%22%2C%22username%22%3A%22peterparkish%22%2C%22language_code%22%3A%22en%22%2C%22allows_write_to_pm%22%3Atrue%2C%22photo_url%22%3A%22https%3A%5C%2F%5C%2Ft.me%5C%2Fi%5C%2Fuserpic%5C%2F320%5C%2FjMwTE1p_IMe6se6v6t6X8uaS1ymy2hHPJ1Oqt3b13hES-84zfc1MJCUrxxLDLgap.svg%22%7D&auth_date=1752425809&signature=ZJxYc3GVBTG5a6xcx3JxpOqYtevxAju3PMQx40R42L9ZwKfkFdFGy0xk2Y5eyooby-dh-DYb3OQMDXBc0369AA&hash=451894c5ac74cc45a0c3eb75cf994f918534632239deda3946bf54b4dae78565
// Connection: keep-alive
// Cookie: _ym_d=1752354634; _ym_uid=1752354634839026226
// Host: portals-market.com
// Referer: https://portals-market.com/
// Sec-Fetch-Dest: empty
// Sec-Fetch-Mode: cors
// Sec-Fetch-Site: same-origin
// User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko)

// offset: 0
// limit: 20
// filter_by_backdrops: Hunter Green
// filter_by_collections: Lol Pop
// filter_by_models: Angelina
// filter_by_symbols: Feather
// sort_by: price asc
// status: listed

// results: [{
//   "id": "cfab2d3b-aae3-4f68-929b-bd8443b6b5a1",
//   "tg_id": "654220",
//   "collection_id": "a7253309-ed0e-4d5d-972c-2bc9c0194bea",
//   "external_collection_number": 46361,
//   "name": "Witch Hat",
//   "photo_url": "https://nft.fragment.com/gift/witchhat-46361.large.jpg",
//   "price": "4",
//   "attributes": [
//       {
//           "type": "model",
//           "value": "Mad Wizard",
//           "rarity_per_mille": 1.5
//       },
//       {
//           "type": "symbol",
//           "value": "Wasp",
//           "rarity_per_mille": 0.2
//       },
//       {
//           "type": "backdrop",
//           "value": "Mint Green",
//           "rarity_per_mille": 1
//       }
//   ],
//   "listed_at": "2025-07-13T18:27:42.747056Z",
//   "status": "listed",
//   "animation_url": "https://nft.fragment.com/gift/witchhat-46361.lottie.json",
//   "emoji_id": "5237969409271685566",
//   "has_animation": true,
//   "floor_price": "2.75",
//   "unlocks_at": "2025-02-20T22:13:37Z",
//   "is_owned": false
// }]
