package facebox

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Similar checks the image in the io.Reader for similar faces.
// Deprecated: use Similars to support multiple faces.
func (c *Client) Similar(image io.Reader) ([]Similar, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "image.dat")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(fw, image)
	if err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	u, err := url.Parse(c.addr + "/facebox/similar")
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		return nil, errors.New("box address must be absolute")
	}
	req, err := http.NewRequest("POST", u.String(), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	return c.parseSimilarResponse(resp.Body)
}

// SimilarURL checks the image at the specified URL for similar faces.
// Deprecated: use SimilarsURL to support multiple faces.
func (c *Client) SimilarURL(imageURL *url.URL) ([]Similar, error) {
	u, err := url.Parse(c.addr + "/facebox/similar")
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		return nil, errors.New("box address must be absolute")
	}
	if !imageURL.IsAbs() {
		return nil, errors.New("url must be absolute")
	}
	form := url.Values{}
	form.Set("url", imageURL.String())
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	return c.parseSimilarResponse(resp.Body)
}

// SimilarID returns similar faces based on the ID provided.
func (c *Client) SimilarID(id string) ([]Similar, error) {
	u, err := url.Parse(c.addr + "/facebox/similar")
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		return nil, errors.New("box address must be absolute")
	}
	if id == "" {
		return nil, errors.New("id can not be empty")
	}
	q := u.Query()
	q.Set("id", id)
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	return c.parseSimilarResponse(resp.Body)
}

// SimilarBase64 checks the Base64 encoded image for similar faces.
// Deprecated: use SimilarsBase64 to support multiple faces.
func (c *Client) SimilarBase64(data string) ([]Similar, error) {
	u, err := url.Parse(c.addr + "/facebox/similar")
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		return nil, errors.New("box address must be absolute")
	}
	form := url.Values{}
	form.Set("base64", data)
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	return c.parseSimilarResponse(resp.Body)
}

func (c *Client) parseSimilarResponse(r io.Reader) ([]Similar, error) {
	var similarResponse struct {
		Success bool
		Error   string
		Similar []Similar
	}
	if err := json.NewDecoder(r).Decode(&similarResponse); err != nil {
		return nil, errors.Wrap(err, "decoding response")
	}
	if !similarResponse.Success {
		return nil, ErrFacebox(similarResponse.Error)
	}
	return similarResponse.Similar, nil
}

// SimilarFace describes a face with similatiries.
type SimilarFace struct {
	Rect         Rect      `json:"rect"`
	SimilarFaces []Similar `json:"similar_faces"`
}

// Similars checks the image in the io.Reader for similar faces.
// Will look for a maximum of limit similar faces for each face.
func (c *Client) Similars(image io.Reader, limit int) ([]SimilarFace, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "image.dat")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(fw, image)
	if err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	u, err := url.Parse(c.addr + "/facebox/similars")
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		return nil, errors.New("box address must be absolute")
	}
	if limit < 1 {
		limit = 5
	}
	q := url.Values{}
	q.Add("limit", strconv.Itoa(limit))
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("POST", u.String(), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	return c.parseSimilarsResponse(resp.Body)
}

// SimilarsURL checks the image at the specified URL for similar faces.
// Will look for a maximum of limit similar faces for each face.
func (c *Client) SimilarsURL(imageURL *url.URL, limit int) ([]SimilarFace, error) {
	u, err := url.Parse(c.addr + "/facebox/similars")
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		return nil, errors.New("box address must be absolute")
	}
	if !imageURL.IsAbs() {
		return nil, errors.New("url must be absolute")
	}
	form := url.Values{}
	form.Set("url", imageURL.String())
	if limit < 1 {
		limit = 5
	}
	form.Set("limit", strconv.Itoa(limit))
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	return c.parseSimilarsResponse(resp.Body)
}

// SimilarsBase64 checks the Base64 encoded image for similar faces.
// Will look for a maximum of limit similar faces for each face.
func (c *Client) SimilarsBase64(data string, limit int) ([]SimilarFace, error) {
	u, err := url.Parse(c.addr + "/facebox/similars")
	if err != nil {
		return nil, err
	}
	if !u.IsAbs() {
		return nil, errors.New("box address must be absolute")
	}
	form := url.Values{}
	form.Set("base64", data)
	if limit < 1 {
		limit = 5
	}
	form.Set("limit", strconv.Itoa(limit))
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	return c.parseSimilarsResponse(resp.Body)
}

func (c *Client) parseSimilarsResponse(r io.Reader) ([]SimilarFace, error) {
	var similarsResponse struct {
		Success bool
		Error   string
		Faces   []SimilarFace
	}
	if err := json.NewDecoder(r).Decode(&similarsResponse); err != nil {
		return nil, errors.Wrap(err, "decoding response")
	}
	if !similarsResponse.Success {
		return nil, ErrFacebox(similarsResponse.Error)
	}
	return similarsResponse.Faces, nil
}
