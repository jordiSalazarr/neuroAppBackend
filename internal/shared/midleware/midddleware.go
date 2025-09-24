package midleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"neuro.app.jordi/internal/shared/config"
	jwtService "neuro.app.jordi/internal/shared/jwt"
	logging "neuro.app.jordi/internal/shared/logger"
)

type ctxKey string

const (
	CtxUserID ctxKey = "user_id"
)

type AccessLogConfig struct {
	MaxBodyBytes int     // p.ej. 64 * 1024
	Env          string  // "dev" | "local" | "prod" (por si quieres condicionar)
	SampleRatio  float64 // 1.0 = siempre; 0.1 = 10% (muestreo simple)
}

func NewAccessLogConfig() AccessLogConfig {
	return AccessLogConfig{
		MaxBodyBytes: 64 * 1024,
		Env:          config.GetCurrentEnvironment(),
		SampleRatio:  1.0,
	}
}
func AccessLog(logger logging.Logger, cfg AccessLogConfig) gin.HandlerFunc {
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 64 * 1024
	}
	if cfg.SampleRatio <= 0 {
		cfg.SampleRatio = 1.0
	}

	return func(c *gin.Context) {
		start := time.Now()

		method := c.Request.Method
		ct := c.ContentType()
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		// ----- Leer y restaurar body (POST/PUT/PATCH no-multipart) -----
		var body string
		var truncated bool
		if (method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch) &&
			ct != "" && !strings.HasPrefix(strings.ToLower(ct), "multipart/") {

			limited := io.LimitReader(c.Request.Body, int64(cfg.MaxBodyBytes+1))
			buf, _ := io.ReadAll(limited)
			if len(buf) > cfg.MaxBodyBytes {
				truncated = true
				buf = buf[:cfg.MaxBodyBytes]
			}
			if !utf8.Valid(buf) {
				body = "<non-utf8>"
			} else {
				body = string(buf)
			}
			body = redactSensitive(body, ct)

			// Restaurar para el handler
			c.Request.Body = io.NopCloser(bytes.NewBuffer(buf))
		}

		// Procesa la request
		c.Next()

		lat := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		query := c.Request.URL.RawQuery

		// Path params
		pp := map[string]any{}
		for _, p := range c.Params {
			pp[p.Key] = p.Value
		}

		// Muestreo sencillo
		if cfg.SampleRatio < 1.0 {
			h := fnv32(method + route + query + c.ClientIP())
			if float64(h%1000)/1000.0 >= cfg.SampleRatio {
				return
			}
		}

		fields := map[string]any{
			"event":      "http_request",
			"method":     method,
			"route":      route,
			"path":       c.Request.URL.Path,
			"status":     status,
			"latency_ms": lat,
			"ip":         c.ClientIP(),
			"ua":         c.Request.UserAgent(),
		}
		if query != "" {
			fields["query"] = query
		}
		if len(pp) > 0 {
			fields["path_params"] = pp
		}
		if body != "" {
			fields["body"] = body
			fields["body_truncated"] = truncated
			fields["content_type"] = ct
		}

		switch {
		case status >= 500:
			logger.Warn(c.Request.Context(), "http_request_error", fields)
		case status >= 400:
			logger.Warn(c.Request.Context(), "http_request_warn", fields)
		default:
			logger.Info(c.Request.Context(), "http_request", fields)
		}
	}
}

// ---- utils ----

func fnv32(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

// Redacción básica para JSON y x-www-form-urlencoded
func redactSensitive(body, ct string) string {
	lct := strings.ToLower(ct)
	if strings.Contains(lct, "json") {
		return redactJSONLike(body, []string{
			"password", "pass", "pwd", "authorization", "auth",
			"token", "refresh_token", "email", "dni", "id_number",
		})
	}
	if strings.Contains(lct, "x-www-form-urlencoded") {
		pairs := strings.Split(body, "&")
		for i, p := range pairs {
			kv := strings.SplitN(p, "=", 2)
			if len(kv) == 2 && isSensitiveKey(kv[0]) {
				pairs[i] = kv[0] + "=<redacted>"
			}
		}
		return strings.Join(pairs, "&")
	}
	// evita logs gigantes si no es JSON ni form
	if len(body) > 4*1024 {
		return body[:4*1024] + "...<truncated>"
	}
	return body
}

func isSensitiveKey(k string) bool {
	k = strings.ToLower(strings.TrimSpace(k))
	sens := []string{
		"password", "pass", "pwd", "authorization", "auth",
		"token", "refresh_token", "email", "dni", "id_number",
	}
	for _, s := range sens {
		if k == s {
			return true
		}
	}
	return false
}

// muy simple: en JSON plano sustituye "key":"value" por "key":"<redacted>"
func redactJSONLike(s string, keys []string) string {
	out := s
	for _, k := range keys {
		// casos: "k":"...", 'k':'...', "k":..., (número/boolean)
		out = redactKey(out, k)
	}
	if len(out) > 64*1024 {
		return out[:64*1024] + "...<truncated>"
	}
	return out
}

func redactKey(s, key string) string {
	// Redacción naive para no introducir dependencias (regex simple)
	// Sustituye valores entre comillas o sin comillas tras "key":
	pats := []string{
		`"` + key + `"\s*:\s*"(.*?)"`,
		`"` + key + `"\s*:\s*'(.*?)'`,
		`"` + key + `"\s*:\s*([^\s,}]+)`,
	}
	repls := []string{
		`"` + key + `":"<redacted>"`,
		`"` + key + `':'<redacted>'`,
		`"` + key + `":"<redacted>"`,
	}
	for i := range pats {
		re := regexpMustCompile(pats[i])
		s = re.ReplaceAllString(s, repls[i])
	}
	return s
}

// pequeñísimo wrapper de regexp para evitar pánico en typo
type safeRegexp struct{ *regexp.Regexp }

func regexpMustCompile(p string) *safeRegexp {
	r, _ := regexp.Compile(p)
	return &safeRegexp{r}
}
func ExtractJWTFromRequest(jwtSvc *jwtService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := strings.TrimSpace(c.GetHeader("Authorization"))
		lower := strings.ToLower(auth)
		if auth == "" || !strings.HasPrefix(lower, "bearer ") {
			c.Header("WWW-Authenticate", `Bearer realm="api", error="invalid_request"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token_requerido"})
			return
		}
		tokenStr := strings.TrimSpace(auth[len("Bearer "):])

		claims, err := jwtSvc.ValidateToken(tokenStr)
		if err != nil || claims == nil {
			c.Header("WWW-Authenticate", `Bearer realm="api", error="invalid_token"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token_invalido"})
			return
		}

		userID := claims.Id
		// gin.Context
		c.Set("user_id", userID)
		// request.Context (para logger u otros paquetes)
		ctx := context.WithValue(c.Request.Context(), CtxUserID, userID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
func GetUserIdFromRequest(c *gin.Context) (string, bool) {
	id, exists := c.Get("id")
	if !exists {
		return "", false
	}
	userId, ok := id.(string)
	if !ok {
		return "", false
	}
	return userId, true
}
