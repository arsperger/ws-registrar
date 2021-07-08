[![Go Report Card](https://goreportcard.com/badge/github.com/arsperger/ws-registrar)](https://goreportcard.com/report/github.com/arsperger/ws-registrar)

# Websockets (SIP) test server

A very simple Websocket server written in Go, meant for use with *WebRTC clients only*.

Main goal and motivation for this repo is having simple service to test load balancing and scaling scenarios with WebRTC clients. HTTP/1.1 upgrade request will negotiate websocket connection and switch to SIP subprotocol. Any subsequent requests from WebRTC clients will be responded with `200 OK` - it assumes that only REGISTER is sent, so the clients will be successfully registered and then stay idle.

Note: any request from websocket client will be responded with text based `SIP 200 OK` response

- Proxy protocol supported
- Has SSL/TLS support with self-signed certs.

To enable `proxy protocol` support set environment variable `PROXYP` to any value.
The default WS port is 8080, if you need to change it set new port to `PORT` env variable.

## SSL/TLS/WSS

The default encrypted port is `8443`, if you need to change it set new port to `SSLPORT` env variable.
Self signed ertificate is generated for `wss.local` domain
If you'd like to use your own, you can clone this repo, gen your own cert, and rebuild the bin/image.

### How to create an HTTPS certificate for localhost domains

This focuses on generating the certificates for loading local virtual hosts hosted on your computer, for development only.

### Certificate authority (CA)

Generate `RootCA.pem`, `RootCA.key` & `RootCA.crt`:

```sh
openssl req -x509 -nodes -new -sha256 -days 1024 -newkey rsa:2048 -keyout RootCA.key -out RootCA.pem -subj "/C=US/CN=Example-Root-CA"
openssl x509 -outform pem -in RootCA.pem -out RootCA.crt
```

Note that `Example-Root-CA` is an example, you can customize the name.

### Domain name certificate

Let's say you have two domains `fake1.local` and `fake2.local` that are hosted on your local machine
for development (using the `hosts` file to point them to `127.0.0.1`).

First, create a file `domains.ext` that lists all your local domains:

```sh
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = localhost
DNS.2 = fake1.local
DNS.3 = fake2.local
```

Generate `localhost.key`, `localhost.csr`, and `localhost.crt`:

```sh
openssl req -new -nodes -newkey rsa:2048 -keyout localhost.key -out localhost.csr -subj "/C=US/ST=YourState/L=YourCity/O=Example-Certificates/CN=localhost.local"
openssl x509 -req -sha256 -days 1024 -in localhost.csr -CA RootCA.pem -CAkey RootCA.key -CAcreateserial -extfile domains.ext -out localhost.crt
```

Note that the country / state / city / name in the first command  can be customized.

## REGISTER example

```sh
"REGISTER sips:ss2.biloxi.example.com SIP/2.0\r\n" +
"Via: SIP/2.0/TLS client.biloxi.example.com:5061;branch=z9hG4bKnashds7\r\n" +
"Max-Forwards: 70\r\n" +
"From: Bob <sips:bob@biloxi.example.com>;tag=a73kszlfl\r\n" +
"To: Bob <sips:bob@biloxi.example.com>\r\n" +
"Call-ID: 1j9FpLxk3uxtm8tn@biloxi.example.com\r\n" +
"CSeq: 1 REGISTER\r\n" +
"Contact: <sips:bob@client.biloxi.example.com>\r\n" +
"Content-Length: 0\r\n\r\n"
```

## Example Output

```bash
10.10.10.1:58198 | GET /
10.10.10.1:58198 | upgraded to websocket
10.10.10.1:58198 | txt | REGISTER sip:sip.sample.com SIP/2.0
Via: SIP/2.0/WSS 1dr8h0krkpml.invalid;branch=z9hG4bK5228023
Max-Forwards: 70
To: "Gob Bleuth" <sip:sipuser_1@sip.sample.com>
From: "Gob Bleuth" <sip:sipuser_1@sip.sample.com>;tag=n8dabtdci5
Call-ID: rodpg9i3i0dhgqpgmsfk5n
CSeq: 81 REGISTER
Contact: <sip:bqiqbcot@1dr8h0krkpml.invalid;transport=ws>;reg-id=1;+sip.instance="<urn:uuid:9e481b70-4eed-4799-b48f-a53f1d3875a7>";expires=300
Allow: ACK,CANCEL,INVITE,MESSAGE,BYE,OPTIONS,INFO,NOTIFY,REFER
Supported: path, gruu, outbound
User-Agent: SIP.js/0.7.8
Content-Length: 0


SIP/2.0 200 OK
Via: SIP/2.0/WSS 1dr8h0krkpml.invalid;branch=z9hG4bK5228023;received=10.10.10.1
From: <sip:sipuser_1@sip.sample.com>;tag=n8dabtdci5
To: <sip:sipuser_1@sip.sample.com>;tag=37GkEhwl6
Call-ID: rodpg9i3i0dhgqpgmsfk5n
CSeq: 81 REGISTER
Contact: <sip:bqiqbcot@1dr8h0krkpml.invalid;transport=ws>;reg-id=1;+sip.instance="<urn:uuid:9e481b70-4eed-4799-b48f-a53f1d3875a7>";expires=300
Content-Length: 0
```
