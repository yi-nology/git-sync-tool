package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/yi-nology/git-sync-tool/biz/config"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(rate.Limit(config.WebhookRateLimit/60.0), config.WebhookRateLimit)

func WebhookAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 1. IP Whitelist Check (Optional)
		if len(config.WebhookIPWhitelist) > 0 {
			clientIP := c.ClientIP()
			allowed := false
			for _, ip := range config.WebhookIPWhitelist {
				if ip == clientIP {
					allowed = true
					break
				}
			}
			if !allowed {
				c.AbortWithStatusJSON(http.StatusForbidden, map[string]string{"error": "IP not allowed"})
				return
			}
		}

		// 2. Rate Limiting
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, map[string]string{"error": "Rate limit exceeded"})
			return
		}

		// 3. Signature Verification
		signature := string(c.GetHeader("X-Hub-Signature-256"))
		if signature == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]string{"error": "Missing signature"})
			return
		}

		// Signature format: sha256=<hex_digest>
		parts := strings.SplitN(signature, "=", 2)
		if len(parts) != 2 || parts[0] != "sha256" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]string{"error": "Invalid signature format"})
			return
		}

		body := c.GetRequest().Body()
		mac := hmac.New(sha256.New, []byte(config.WebhookSecret))
		mac.Write(body)
		expectedMAC := mac.Sum(nil)
		expectedSignature := hex.EncodeToString(expectedMAC)

		if !hmac.Equal([]byte(parts[1]), []byte(expectedSignature)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]string{"error": "Invalid signature"})
			return
		}

		c.Next(ctx)
	}
}
