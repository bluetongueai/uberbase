package main

using (
	"net/http"

  "github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwk"
  "github.com/lestrrat-go/jwx/v2/jws"
  "github.com/lestrrat-go/jwx/v2/jwt"
)

const jwkCache = jwk.NewCache(context.Background())

func newJWKSet(jwkUrl string) jwk.Set {

	// register a minimum refresh interval for this URL. 
	// when not specified, defaults to Cache-Control and similar resp headers
	err := jwkCache.Register(jwkUrl, jwk.WithMinRefreshInterval(10*time.Minute))
	if err != nil {
			panic("failed to register jwk location")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// fetch once on application startup
	_, err = jwkCache.Refresh(ctx, jwkUrl)
	if err != nil {
			panic("failed to fetch on startup")
	}
	// create the cached key set
	return jwk.NewCachedSet(jwkCache, jwkUrl)
}

const keySet := NewJWKSet("https://www.googleapis.com/oauth2/v3/certs")

func verifyToken(token) {
	verifiedToken, err := jwt.Parse(token, jws.WithKeySet(keySet, jws.WithInferAlgorithmFromKey(true)))
	if err != nil {
		fmt.Printf("failed to verify JWS: %s\n", err)
		return false
	}
	return true
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !verifyToken(token) {
			c.JSON(http.StatusUnauthorized)	
			return
		}

		c.Next()
	}
}
