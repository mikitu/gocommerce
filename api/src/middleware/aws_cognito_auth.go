package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"math/big"
	"net/http"
	"strings"
	"time"

	apihttp "github.com/mikitu/gocommerce/api/src/http"
)

type (
	// BasicAuthConfig defines the config for BasicAuth middleware.
	CognitoAuthConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// Validator is a function to validate BasicAuth credentials.
		// Required.
		Validator CognitoAuthValidator

		// Realm is a string to define realm attribute of BasicAuth.
		// Default value "Restricted".
		Realm string

		AwsRegion string

		UserPoolId string
	}

	// CognitoAuthValidator defines a function to validate Aws Cognito credentials.
	CognitoAuthValidator func(string, string, echo.Context) (bool, error)
)

const (
	basic        = "basic"
	defaultRealm = "Restricted"
)

var (
	// DefaultBasicAuthConfig is the default BasicAuth middleware config.
	DefaultCognitoAuthConfig = CognitoAuthConfig{
		Skipper: middleware.DefaultSkipper,
		Realm:   defaultRealm,
	}
)

func CognitoAuth(fn CognitoAuthValidator) echo.MiddlewareFunc {
	c := DefaultCognitoAuthConfig
	c.Validator = fn
	return CognitoAuthWithConfig(c)
}

func CognitoAuthWithConfig(config CognitoAuthConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultCognitoAuthConfig.Skipper
	}
	if config.Realm == "" {
		config.Realm = defaultRealm
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fmt.Printf("%v+ , %+v\n", c.Path(), config.Skipper);

			if config.Skipper(c) {
				return next(c)
			}
			accessToken := c.Request().Header.Get("access_token")

			if accessToken == "" {
				return failWithError(c, errors.New("Please provide an access_token"))
			}

			// 1. Download and store the JSON Web Key (JWK) for your user pool.
			jwkURL := fmt.Sprintf("https://cognito-idp.%v.amazonaws.com/%v/.well-known/jwks.json", config.AwsRegion, config.UserPoolId)

			jwk := getJWK(jwkURL)

			//realm := defaultRealm
			token, err := validateToken(accessToken, config.AwsRegion, config.UserPoolId, jwk)
			if err == nil && token.Valid {
				if config.Realm != defaultRealm {
					//realm = strconv.Quote(config.Realm)
				}
				return next(c)
			}
			return failWithError(c, err)
		}
	}
}

func failWithError(c echo.Context, err error) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusUnauthorized)
	if err := json.NewEncoder(c.Response()).Encode(&apihttp.ResponseFormatter{Status: http.StatusUnauthorized, Data: nil, Errors: err.Error()}); err != nil {
		return err
	}
	c.Response().Flush()
	return err
}

func validateToken(tokenStr, region, userPoolID string, jwk map[string]JWKKey) (*jwt.Token, error) {

	// 2. Decode the token string into JWT format.
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {

		// cognito user pool : RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// 5. Get the kid from the JWT token header and retrieve the corresponding JSON Web Key that was stored
		if kid, ok := token.Header["kid"]; ok {
			if kidStr, ok := kid.(string); ok {
				key := jwk[kidStr]
				// 6. Verify the signature of the decoded JWT token.
				rsaPublicKey := convertKey(key.E, key.N)
				return rsaPublicKey, nil
			}
		}

		// rsa public key取得できず
		return "", nil
	})

	if err != nil {
		return token, err
	}

	claims := token.Claims.(jwt.MapClaims)

	iss, ok := claims["iss"]
	if !ok {
		return token, fmt.Errorf("token does not contain issuer")
	}
	issStr := iss.(string)
	if strings.Contains(issStr, "cognito-idp") {
		// 3. 4. 7.のチェックをまとめて
		err = validateAWSJwtClaims(claims, region, userPoolID)
		if err != nil {
			return token, err
		}
	}

	if token.Valid {
		return token, nil
	}
	return token, err
}

// validateAWSJwtClaims validates AWS Cognito User Pool JWT
func validateAWSJwtClaims(claims jwt.MapClaims, region, userPoolID string) error {
	var err error
	// 3. Check the iss claim. It should match your user pool.
	issShoudBe := fmt.Sprintf("https://cognito-idp.%v.amazonaws.com/%v", region, userPoolID)
	err = validateClaimItem("iss", []string{issShoudBe}, claims)
	if err != nil {
		return err
	}

	// 4. Check the token_use claim.
	validateTokenUse := func() error {
		if tokenUse, ok := claims["token_use"]; ok {
			if tokenUseStr, ok := tokenUse.(string); ok {
				if tokenUseStr == "id" || tokenUseStr == "access" {
					return nil
				}
			}
		}
		return errors.New("token_use should be id or access")
	}

	err = validateTokenUse()
	if err != nil {
		return err
	}

	// 7. Check the exp claim and make sure the token is not expired.
	err = validateExpired(claims)
	if err != nil {
		return err
	}

	return nil
}

func validateClaimItem(key string, keyShouldBe []string, claims jwt.MapClaims) error {
	if val, ok := claims[key]; ok {
		if valStr, ok := val.(string); ok {
			for _, shouldbe := range keyShouldBe {
				if valStr == shouldbe {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("%v does not match any of valid values: %v", key, keyShouldBe)
}

func validateExpired(claims jwt.MapClaims) error {
	if tokenExp, ok := claims["exp"]; ok {
		if exp, ok := tokenExp.(float64); ok {
			now := time.Now().Unix()
			fmt.Printf("current unixtime : %v\n", now)
			fmt.Printf("expire unixtime  : %v\n", int64(exp))
			if int64(exp) > now {
				return nil
			}
		}
		return errors.New("cannot parse token exp")
	}
	return errors.New("token is expired")
}

// https://gist.github.com/MathieuMailhos/361f24316d2de29e8d41e808e0071b13
func convertKey(rawE, rawN string) *rsa.PublicKey {
	decodedE, err := base64.RawURLEncoding.DecodeString(rawE)
	if err != nil {
		panic(err)
	}
	if len(decodedE) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(decodedE):], decodedE)
		decodedE = ndata
	}
	pubKey := &rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(decodedE[:])),
	}
	decodedN, err := base64.RawURLEncoding.DecodeString(rawN)
	if err != nil {
		panic(err)
	}
	pubKey.N.SetBytes(decodedN)
	return pubKey
}

// JWK is json data struct for JSON Web Key
type JWK struct {
	Keys []JWKKey
}

// JWKKey is json data struct for cognito jwk key
type JWKKey struct {
	Alg string
	E   string
	Kid string
	Kty string
	N   string
	Use string
}

func getJWK(jwkURL string) map[string]JWKKey {

	jwk := &JWK{}

	_ = getJSON(jwkURL, jwk)

	jwkMap := make(map[string]JWKKey, 0)
	for _, jwk := range jwk.Keys {
		jwkMap[jwk.Kid] = jwk
	}
	return jwkMap
}

func getJSON(url string, target interface{}) error {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}