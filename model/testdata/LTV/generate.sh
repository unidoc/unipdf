openssl genrsa -des3 -out rootCA.key 4096
openssl req -x509 -new -nodes -key rootCA.key -sha256 -days 1024 -out rootCA.crt
openssl genrsa -out mydomain.com.key 2048

openssl req -new -sha256 \
    -key mydomain.com.key \
    -subj "/C=US/ST=CA/O=MyOrg, Inc./CN=mydomain.com" \
    -reqexts SAN \
    -config <(cat /etc/ssl/openssl.cnf \
        <(printf "\n[SAN]\nsubjectAltName=DNS:mydomain.com,DNS:www.mydomain.com")) \
    -out mydomain.com.csr

openssl req -in mydomain.com.csr -noout -text

openssl x509 -req -in mydomain.com.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out mydomain.com.crt -days 500 -sha256


openssl x509 -in mydomain.com.crt -text -noout

openssl pkcs12 -export -out mydomain.com.p12 -inkey mydomain.com.key -in mydomain.com.crt -certfile rootCA.crt