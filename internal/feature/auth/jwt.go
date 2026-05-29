package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Claims struct {
	ID uuid.UUID `json:"id"`

	jwt.RegisteredClaims
}

func GenerateJwtToken(id uuid.UUID, role string, expire time.Time) (string, error) {
	claim := &Claims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expire),
		},
	}

	secretKey := []byte(os.Getenv("JWT_SECRET"))
	tokens := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, err := tokens.SignedString(secretKey)

	slog.Warn("Token generated: ", role)
	if err != nil {
		slog.Error("error generate jwt token", err.Error())
		return "", err
	}
	return tokenString, nil
}

func GenerateJwtTokens(id uuid.UUID, role string) (string, string, error) {
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	accessClaims := &Claims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	// Refresh token (длинный)
	refreshClaims := &Claims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
		SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
func AddCookieTokens(id uuid.UUID, Role string, w http.ResponseWriter, domain string) error {
	expirationTimeAccess := time.Now().Add(4 * time.Hour)
	expirationTimeRefresh := time.Now().Add(14 * 24 * time.Hour)
	refreshToken, err := GenerateJwtToken(id, Role, expirationTimeRefresh)
	if err != nil {
		return err
	}
	accessToken, err := GenerateJwtToken(id, Role, expirationTimeAccess)
	if err != nil {
		return err
	}
	http.SetCookie(w, GenerateCookie("accessToken", expirationTimeAccess, false, accessToken, domain))
	http.SetCookie(w, GenerateCookie("refreshToken", expirationTimeRefresh, true, refreshToken, domain))

	return nil
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			slog.Error("Unexpected signing method", "method", token.Header["alg"])
			return nil, errors.New("unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		slog.Error("JWT parse error", "err", err)
		return nil, err
	}

	if !token.Valid {
		slog.Error("JWT token invalid")
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		slog.Error("Token claims type assertion failed")
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func NewAccessToken(tokenString string, threshold time.Duration, w http.ResponseWriter, domain string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	remainingTime := time.Until(claims.ExpiresAt.Time)
	if remainingTime > threshold {
		return tokenString, nil
	}

	newExpire := time.Now().Add(4 * time.Hour)
	newToken, err := GenerateJwtToken(claims.ID, "user", newExpire)
	if err != nil {
		return "", err
	}

	http.SetCookie(w, GenerateCookie("accessToken", newExpire, false, newToken, domain))
	return newToken, nil
}

func GenerateCookie(name string, expire time.Time, httpOnly bool, value string, domain string) *http.Cookie {
	cookie := &http.Cookie{
		Name:        name,
		Value:       value,
		Expires:     expire,
		Partitioned: true,
		Path:        "/",
		Secure:      true,
		HttpOnly:    httpOnly,
		SameSite:    http.SameSiteLaxMode,
	}
	if domain := os.Getenv("DOMAIN_PROD"); domain != "" {
		cookie.Domain = domain
	}
	return cookie
}
