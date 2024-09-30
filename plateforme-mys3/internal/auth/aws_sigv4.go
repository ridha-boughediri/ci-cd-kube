// internal/auth/aws_sigv4.go
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"plateforme-mys3/config"
	"sort"
	"strings"
	"time"
)

// VerifyAWSSignature vérifie la signature AWS SigV4 d'une requête
func VerifyAWSSignature(r *http.Request, cfg config.Config) bool {
	// Étape 1 : Extraire l'en-tête Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("Authorization header missing")
		return false
	}

	// Étape 2 : Valider le préfixe de l'en-tête
	if !strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256 ") {
		log.Println("Authorization header does not start with AWS4-HMAC-SHA256")
		return false
	}

	// Étape 3 : Parser l'en-tête Authorization
	authParams := parseAuthorizationHeader(authHeader)
	if authParams == nil {
		log.Println("Failed to parse Authorization header")
		return false
	}

	// Étape 4 : Vérifier l'Access Key ID
	if len(authParams["Credential"]) == 0 {
		log.Println("Credential field missing in Authorization header")
		return false
	}
	accessKeyID := authParams["Credential"][0]
	if accessKeyID != cfg.AccessKeyID {
		log.Printf("Invalid Access Key ID: %s", accessKeyID)
		return false
	}

	// Étape 5 : Récupérer les SignedHeaders
	if len(authParams["SignedHeaders"]) == 0 {
		log.Println("SignedHeaders field missing in Authorization header")
		return false
	}
	signedHeaders := strings.Split(authParams["SignedHeaders"][0], ";")
	log.Printf("Signed Headers: %v", signedHeaders)

	// Étape 6 : Recalculer la signature
	canonicalRequest, err := buildCanonicalRequest(r, signedHeaders)
	if err != nil {
		log.Printf("Error building canonical request: %v", err)
		return false
	}
	log.Printf("Canonical Request: %s", canonicalRequest)

	stringToSign, err := buildStringToSign(r, canonicalRequest, cfg)
	if err != nil {
		log.Printf("Error building string to sign: %v", err)
		return false
	}
	log.Printf("String to Sign: %s", stringToSign)

	// Correction ici : Utiliser getTimestamp(r).Format("20060102") au lieu de getDate(r)
	t := getTimestamp(r)
	signingKey := getSignatureKey(cfg.SecretAccessKey, t.Format("20060102"), cfg.Region, "s3")
	expectedSignature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	log.Printf("Expected Signature: %s", expectedSignature)

	// Étape 7 : Comparer les signatures
	if len(authParams["Signature"]) == 0 {
		log.Println("Signature field missing in Authorization header")
		return false
	}
	providedSignature := authParams["Signature"][0]
	log.Printf("Provided Signature: %s", providedSignature)

	result := hmac.Equal([]byte(expectedSignature), []byte(providedSignature))
	log.Printf("Signature Valid: %v", result)
	return result
}

// parseAuthorizationHeader analyse l'en-tête Authorization et retourne un map des paramètres
func parseAuthorizationHeader(authHeader string) map[string][]string {
	params := make(map[string][]string)
	// Supprimer le préfixe "AWS4-HMAC-SHA256 "
	authorization := strings.TrimPrefix(authHeader, "AWS4-HMAC-SHA256 ")

	// Diviser les différents paramètres par ","
	parts := strings.Split(authorization, ",")
	for _, part := range parts {
		// Diviser chaque paramètre par "="
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		value := kv[1]
		params[key] = append(params[key], value)
	}
	return params
}

// buildCanonicalRequest construit la requête canonique selon les spécifications AWS SigV4
func buildCanonicalRequest(r *http.Request, signedHeaders []string) (string, error) {
	// Méthode HTTP
	method := r.Method

	// URI
	uri := getCanonicalURI(r.URL.Path)

	// Query String
	canonicalQueryString := getCanonicalQueryString(r.URL.RawQuery)

	// Headers
	canonicalHeaders, err := getCanonicalHeaders(r, signedHeaders)
	if err != nil {
		return "", err
	}

	// Liste des Signed Headers
	signedHeadersStr := strings.Join(signedHeaders, ";")

	// Hash du corps
	payloadHash := getPayloadHash(r)

	// Construire la requête canonique
	canonicalRequest := strings.Join([]string{
		method,
		uri,
		canonicalQueryString,
		canonicalHeaders,
		signedHeadersStr,
		payloadHash,
	}, "\n")

	return canonicalRequest, nil
}

// getCanonicalURI formate l'URI selon les spécifications AWS SigV4
func getCanonicalURI(path string) string {
	if path == "" {
		return "/"
	}
	// Encodage URI
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		segments[i] = escapePath(seg)
	}
	return strings.Join(segments, "/")
}

// escapePath encode les segments du chemin selon les spécifications AWS SigV4
func escapePath(seg string) string {
	return strings.ReplaceAll(strings.ReplaceAll(seg, " ", "%20"), "/", "%2F")
}

// getCanonicalQueryString formate la query string selon les spécifications AWS SigV4
func getCanonicalQueryString(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	params := strings.Split(rawQuery, "&")
	sort.Strings(params)
	return strings.Join(params, "&")
}

// getCanonicalHeaders formate les en-têtes selon les spécifications AWS SigV4
func getCanonicalHeaders(r *http.Request, signedHeaders []string) (string, error) {
	var headers []string
	for _, header := range signedHeaders {
		// Les en-têtes doivent être en minuscules
		headerLower := strings.ToLower(header)
		values := r.Header[http.CanonicalHeaderKey(headerLower)]
		if len(values) == 0 {
			log.Printf("Signed header %s not present in request", headerLower)
			continue // Selon les spécifications, ignorer les en-têtes manquants
		}
		// Supprimer les espaces supplémentaires
		cleanValue := strings.Join(values, ",")
		cleanValue = strings.TrimSpace(cleanValue)
		cleanValue = strings.ReplaceAll(cleanValue, "\t", " ")
		cleanValue = strings.Join(strings.Fields(cleanValue), " ") // Remplacer les espaces multiples par un seul espace
		headers = append(headers, headerLower+":"+cleanValue+"\n")
	}
	return strings.Join(headers, ""), nil
}

// getPayloadHash retourne le hash SHA256 du corps de la requête
func getPayloadHash(r *http.Request) string {
	if r.Body == nil {
		return sha256Hex("")
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		return sha256Hex("")
	}
	// Remettre le corps dans le lecteur pour une utilisation ultérieure
	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
	return sha256Hex(string(bodyBytes))
}

// sha256Hex retourne le hash SHA256 hexadécimal d'une chaîne donnée
func sha256Hex(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// buildStringToSign construit la chaîne à signer selon les spécifications AWS SigV4
func buildStringToSign(r *http.Request, canonicalRequest string, cfg config.Config) (string, error) {
	// Algorithme
	algorithm := "AWS4-HMAC-SHA256"

	// Horodatage (format ISO8601)
	t := getTimestamp(r)
	credentialScope := getCredentialScope(t, cfg)

	// Hash de la requête canonique
	hashedCanonicalRequest := sha256Hex(canonicalRequest)

	// Construire la chaîne à signer
	stringToSign := strings.Join([]string{
		algorithm,
		t.Format("20060102T150405Z"),
		credentialScope,
		hashedCanonicalRequest,
	}, "\n")

	return stringToSign, nil
}

// getTimestamp extrait le timestamp de la requête ou utilise l'heure actuelle
func getTimestamp(r *http.Request) time.Time {
	// Chercher l'en-tête x-amz-date
	dateHeader := r.Header.Get("x-amz-date")
	if dateHeader != "" {
		t, err := time.Parse("20060102T150405Z", dateHeader)
		if err == nil {
			return t
		}
	}
	// Sinon, utiliser l'heure actuelle
	return time.Now().UTC()
}

// getCredentialScope construit le scope de credential
func getCredentialScope(t time.Time, cfg config.Config) string {
	date := t.Format("20060102")
	return strings.Join([]string{
		date,
		cfg.Region,
		"s3",
		"aws4_request",
	}, "/")
}

// getSignatureKey génère la clé de signature basée sur les informations fournies
func getSignatureKey(secret, date, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "aws4_request")
	return kSigning
}

// hmacSHA256 calcule le HMAC-SHA256 d'un message avec une clé donnée
func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}
