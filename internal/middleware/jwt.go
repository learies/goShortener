package middleware

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/learies/goShortener/internal/config/contextutils"
)

// Claims represents the claims in a JWT token.
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

// CreateUserID generates a new UUID and returns it as a string.
func CreateUserID() string {
	userID := uuid.New().String()
	return userID
}

// JWTMiddleware is an HTTP middleware that handles JWT authentication.
// It checks for a JWT token in the request cookies and creates a new token if none is found.
// It sets the user ID in the request context and passes the request to the next handler.
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string
		var tokenString string

		// Чтение токена из куки
		cookie, err := r.Cookie("token")
		if err == nil {
			tokenString = cookie.Value
		}

		// Если токен не передан нужно создать userID и создать для него токен
		if tokenString == "" {
			userID = CreateUserID()

			// Время жизни токена 1 минута
			expirationTime := time.Now().Add(1 * time.Minute)

			claims := &Claims{
				UserID: userID,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(expirationTime),
				},
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

			tokenString, err := token.SignedString([]byte("qwerty"))
			if err != nil {
				http.Error(w, "Could not create token", http.StatusInternalServerError)
				return
			}

			// Устанавливаем токен в куки
			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    tokenString,
				Expires:  expirationTime,
				HttpOnly: true,
				Path:     "/",
			})
		} else {
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte("qwerty"), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			userID = claims.UserID
		}

		ctx := contextutils.WithUserID(r.Context(), uuid.MustParse(userID))
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
