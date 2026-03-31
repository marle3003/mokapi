#!/bin/bash

clear

EXE="/Users/maesi/GolandProjects/mokapi/mokapi-darwin-arm64"
#EXE="mokapi"

CMD="$EXE &"
echo $CMD | pv -qL 10 # simulate typing
eval $CMD
PID=$!

sleep 2 # to give docker time

# Your secret key (used to sign the JWT token)
SECRET_KEY="your_secret_key_here"

# Header (JWT header in JSON)
HEADER='{
  "alg": "HS256",
  "typ": "JWT"
}'

# Payload (Claims or your data, including scopes)
PAYLOAD='{
  "sub": "1234567890",
  "name": "John Doe",
  "iat": '$(date +%s)',
  "scopes": ["read", "write", "delete"]
}'

# Encode the header and payload to Base64Url (replace + and / with - and _)
encode_base64url() {
    echo -n "$1" | openssl base64 -e | tr '+/' '-_' | tr -d '='
}

# Create Base64Url encoded header and payload
HEADER_ENCODED=$(encode_base64url "$HEADER")
PAYLOAD_ENCODED=$(encode_base64url "$PAYLOAD")

# Create the message to be signed (Header and Payload combined)
MESSAGE="$HEADER_ENCODED.$PAYLOAD_ENCODED"

# Create the signature using HMAC SHA-256 (your secret key)
SIGNATURE=$(echo -n "$MESSAGE" | openssl dgst -sha256 -hmac "$SECRET_KEY" -binary | openssl base64 -e | tr '+/' '-_' | tr -d '=')

# Combine Header, Payload, and Signature to form the JWT
JWT="$MESSAGE.$SIGNATURE"

# Output the JWT
echo "Generated JWT Token: $JWT"

# Make the request
CMD="curl -X GET 'http://localhost/protected' \
     -H 'Authorization: Bearer $JWT'"

#echo $CMD | pv -qL 10
eval $CMD

# stop mokapi
kill $PID