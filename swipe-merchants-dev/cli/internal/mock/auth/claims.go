package auth

// Audience is the fixed `aud` claim value the mock issues into every JWT
// (scaffold §5.3, D-014). Tokens with any other audience are rejected by
// the auth middleware.
const Audience = "bml.swipe.merchants"

// ClaimClientID is the JWT claim name carrying the OAuth client id, in
// addition to the standard `sub` claim. The duplication mirrors D-014.
const ClaimClientID = "client_id"

// ClaimMerchantID is the JWT claim name carrying the merchant id the client
// belongs to.
const ClaimMerchantID = "merchant_id"

// ClaimScope is the JWT claim name for the space-separated scope string,
// matching OAuth2 conventions.
const ClaimScope = "scope"
