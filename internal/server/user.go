package server

import (
    jwt "github.com/golang-jwt/jwt/v5"
    "time"
    "os"
    "errors"
    "log"
    "encoding/json"
)


const privKeyPath =  "/keys/app.rsa"     // `$ openssl genrsa -out app.rsa 2048`
const pubKeyPath  =  "/keys/app.rsa.pub" // `$ openssl rsa -in app.rsa -pubout > app.rsa.pub`

const AuthTokenValidTime = time.Minute * 15

const Issuer string = "ResumeParser"

func PassToHash(password string) (string, error) {
    return password, nil
}

type ResumeClaims struct {
    TokenType string `json:"tokenType"`
    jwt.RegisteredClaims
}

func customParser(token string) (*jwt.Token, error) {
    current_dir, err := os.Getwd()
    if err != nil {
        return nil, err
    }
    verifyBytes, err := os.ReadFile(current_dir + pubKeyPath)
    if err != nil {
        return nil, err
    }
    verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
    if err != nil {
        return nil, err
    }

    parsedToken, err := jwt.Parse(
        token,
        func(token *jwt.Token) (interface{}, error) {
            return verifyKey, nil
        },
    )
    if err != nil || !parsedToken.Valid{
        return nil, errors.New("Invalid token")
    }
    return parsedToken, err
}

func getTokenExpirationTime(token string) (time.Time, error) {
    parsedToken, err := customParser(token)
    jsonString, err := json.Marshal(parsedToken.Claims)
    if err != nil {
        return time.Now().UTC(), err
    }
    claims := ResumeClaims{}
    if json.Unmarshal(jsonString, &claims) != nil {
        return time.Now().UTC(), errors.New("Invalid token")
    }
    if claims.ExpiresAt.Before(time.Now().UTC()) {
        return time.Now().UTC(), errors.New("Token has expired")
    }
    return claims.ExpiresAt.UTC(), nil
}

func DecodeAuthToken(token string) (string, error) {
    if (len(token) < 8) {
        return "", errors.New("token length too small")
    }
    token = token[7:]
    parsedToken, err := customParser(token)
    if err != nil {
        return "", err
    }
    jsonString, err := json.Marshal(parsedToken.Claims)
    if err != nil {
        return "", err
    }
    claims := ResumeClaims{}
    if json.Unmarshal(jsonString, &claims) != nil {
        return "", errors.New("Invalid token")
    }
    if claims.TokenType != "auth" {
        return "", errors.New("Invalid token type")
    }
    if claims.ExpiresAt.Before(time.Now().UTC()) {
        return "", errors.New("Token has expired")
    }
    // maybe check if sub actually exists in db?
    sub, err := parsedToken.Claims.GetSubject()
    if err != nil {
        return "", err
    }
    return sub, nil
}

func CreateTokens(email string) (string, error) {
    authClaims := ResumeClaims{
        "auth",
        jwt.RegisteredClaims{
            Issuer:    Issuer,
            IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
            ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(1 * time.Hour)),
            NotBefore: jwt.NewNumericDate(time.Now().UTC()),
            Subject:   email,
            ID:        "1",
            Audience:  jwt.ClaimStrings{"https://reflecto.trend"},
        },
    }

    current_dir, err := os.Getwd()
    if err != nil {
        log.Println("error getting current dir", err)
        return "", err
    }

    privBytes, err := os.ReadFile(current_dir + privKeyPath)
    if err != nil {
        log.Println("error reading privKey", err)
        return "", err
    }
    privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
    if err != nil {
        log.Println("error parsing privKey", err)
        return "", err
    }

    authToken, err := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), authClaims).SignedString(privKey)
    if err != nil {
        log.Println("Error signing token:", err)
        return "", err
    }

    return authToken, nil
}




