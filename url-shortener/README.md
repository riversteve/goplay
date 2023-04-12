# url-shortener
URL shortener service

## Setup
```
git clone <repo url>

go run url-shortener/urlshortener.go
```

## Create short URL
`curl --fail-with-body -X POST http://localhost:8080/shorten -d "url=https://www.example.com"`

## Delete short URL
`curl --fail-with-body -X DELETE "http://localhost:8080/delete?shortUrl=test123"`
