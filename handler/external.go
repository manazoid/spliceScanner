package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// RequestService sending request to splice
func RequestService(c *http.Client, method, url, authorization, agent string) (*http.Response, error) {
	var req *http.Request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	h := &req.Header
	if authorization != "" {
		h.Set("Cookie", fmt.Sprintf("_splice_web_session=%s", authorization))
	}

	h.Set("Sec-Ch-Ua-Platform", platform)
	h.Set("Origin", spliceHost)
	h.Set("Host", httpsSpliceAPI)
	h.Set("Connection", "keep-alive")

	if agent != "" {
		h.Set("User-Agent", agent)
	} else {
		h.Set("User-Agent",
			`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.66 Safari/537.36 Edg/103.0.1264.44`)
		h.Set("Sec-Ch-Ua",
			`"Not;A Brand";v="99", "Microsoft Edge";v="103", "Chromium";v="103"`)
	}

	return c.Do(req.WithContext(ctx))
}

func ExtractBody(c *http.Client, method, url, auth string, body io.Reader) ([]byte, error) {
	lastRaw, err := RequestServer(method, c, url, auth, body)
	if err != nil {

		return nil, err
	}
	lastBody, err := io.ReadAll(lastRaw.Body)
	if err != nil {
		return nil, err
	}

	if lastRaw.StatusCode > 299 {
		return nil, errors.New(fmt.Sprintf("response %d: %s", lastRaw.StatusCode, lastBody))
	}

	return lastBody, nil
}

func ExtractServiceBody(c *http.Client, method, url, auth, agent string) ([]byte, int, error) {
	lastRaw, err := RequestService(c, method, url, auth, agent)
	if err != nil {
		return nil, 0, err
	}

	lastBody, err := io.ReadAll(lastRaw.Body)
	if err != nil {
		return nil, 0, err
	}

	return lastBody, lastRaw.StatusCode, nil
}

// RequestServer sending request to server
func RequestServer(method string, c *http.Client, url string, auth string, body io.Reader) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	h := &req.Header
	if auth != "" {
		h.Set("Cookie", fmt.Sprintf("%s=%s", authCookie, auth))
	}

	h.Set("Sec-Ch-Ua-Platform", platform)
	h.Set("Origin", spliceHost)
	h.Set("Host", httpsSpliceAPI)
	h.Set("Connection", "keep-alive")

	return c.Do(req.WithContext(ctx))
}

func LoginSplice(c *http.Client, url, agent string, values map[string]io.Reader) (*http.Response, error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var (
			fw  io.Writer
			err error
		)
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return nil, err
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return nil, err
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return nil, err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Create background context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, err
	}
	h := &req.Header
	// Don't forget to set the content type, this will contain the boundary.
	h.Set("Content-Type", w.FormDataContentType())
	h.Set("Sec-Ch-Ua-Platform", platform)
	h.Set("Origin", spliceHost)
	h.Set("Host", httpsSpliceAPI)
	h.Set("Connection", "keep-alive")

	if agent != "" {
		h.Set("User-Agent", agent)
	} else {
		h.Set("User-Agent",
			`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.66 Safari/537.36 Edg/103.0.1264.44`)
		h.Set("Sec-Ch-Ua",
			`"Not;A Brand";v="99", "Microsoft Edge";v="103", "Chromium";v="103"`)
	}

	// Submit the request
	return c.Do(req.WithContext(ctx))
}
