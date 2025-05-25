package utils

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "email_server/config"
)

type Claims struct {
    UserID   int64  `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(userID int64, username, role string) (string, error) {
    expirationTime := time.Now().Add(time.Duration(config.AppConfig.JWT.ExpiresIn) * time.Hour)
    
    claims := &Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "email-server",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(config.AppConfig.JWT.SecretKey))
}

func ParseToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(config.AppConfig.JWT.SecretKey), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("无效的token")
}

func RefreshToken(tokenString string) (string, error) {
    claims, err := ParseToken(tokenString)
    if err != nil {
        return "", err
    }
    
    // 如果token还有超过1小时的有效期，就不刷新
    if time.Until(claims.ExpiresAt.Time) > time.Hour {
        return tokenString, nil
    }
    
    return GenerateToken(claims.UserID, claims.Username, claims.Role)
}
