```bash
go run $(ls *.go | grep -v '_test.go')

curl -H "Authorization: Bearer token" 127.0.0.1:3000/api/v1/health
```