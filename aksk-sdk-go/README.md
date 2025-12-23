# aksk-sdk-go

A production-ready AK/SK authentication SDK for Go.

## Install

```bash
go get github.com/night1008/aksk-sdk-go
```


## Usage
```go
cli := client.New("ak", "sk", "https://api.example.com")

req, _ := http.NewRequest("POST",
  "https://api.example.com/v1/test?a=1&b=2", nil)

resp, err := cli.Do(req, []byte(`{"name":"demo"}`))
```

---

## 版本发布规范

```bash
git tag v1.0.0
git push origin v1.0.0
```