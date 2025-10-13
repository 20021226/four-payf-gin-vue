package xingyi

import (
    "encoding/json"
    "io"
    "net/http"
    "net/url"
)

// OCRProvider defines an interface to recognize captcha from base64 image
type OCRProvider interface {
    // Recognize takes base64 image string and returns recognized text
    Recognize(base64Image string) (string, error)
}

// NoopOCR returns empty string; replace in wiring with real impl
type NoopOCR struct{}

func (NoopOCR) Recognize(base64Image string) (string, error) { return "", nil }

// RemoteOCR calls an external OCR HTTP service like the Python implementation
// Python reference:
// url = "http://192.168.0.208:9580/ocr"
// data = {"image": encoded_string, "probability": False, "png_fix": False}
// response = requests.post(url, data=data); response.json()['data']
// This Go implementation mirrors the same behavior.
type RemoteOCR struct {
    URL string
}

type ocrResponse struct {
    Data string `json:"data"`
}

func (r RemoteOCR) Recognize(base64Image string) (string, error) {
    if r.URL == "" {
        return "", nil
    }
    form := url.Values{}
    form.Set("image", base64Image)
    form.Set("probability", "false")
    form.Set("png_fix", "false")

    resp, err := http.PostForm(r.URL, form)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    var out ocrResponse
    if err := json.Unmarshal(body, &out); err != nil {
        return "", err
    }
    return out.Data, nil
}