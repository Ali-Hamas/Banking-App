package session

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// SignQR signs the QR payload with HS256. Field names and value types match the Node
// jsonwebtoken output: { sid, atm, iat, exp } where iat/exp are Unix seconds.
func SignQR(secret string, p QrPayload) (string, error) {
	claims := jwt.MapClaims{
		"sid": p.Sid,
		"atm": p.Atm,
		"iat": p.Iat,
		"exp": p.Exp,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

// VerifyQR parses the token, enforces HS256, and returns the payload.
// Explicitly rejects unexpected algorithms (none, RS256, etc.) — banks care about this.
func VerifyQR(secret, token string) (QrPayload, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return QrPayload{}, err
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return QrPayload{}, errors.New("invalid claims")
	}
	out := QrPayload{}
	if v, ok := claims["sid"].(string); ok {
		out.Sid = v
	}
	if v, ok := claims["atm"].(string); ok {
		out.Atm = v
	}
	if v, ok := claims["iat"].(float64); ok {
		out.Iat = int64(v)
	}
	if v, ok := claims["exp"].(float64); ok {
		out.Exp = int64(v)
	}
	if out.Sid == "" {
		return QrPayload{}, errors.New("missing sid")
	}
	return out, nil
}
