package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"example.com/csiproject/backend/model"
)

type Client struct {
	Hostname string
	Port     string
	Username string
	Password string
}

type GetVolumeResponse struct {
	Volume model.Volume
}
type CreateVolumeResponse struct {
	Volume model.Volume
}
type GetAllVolumesResponse struct {
	Volumes []model.Volume
}

func NewClient(hostname, port string) *Client {
	c := Client{
		Hostname: hostname,
		Port:     port,
	}
	return &c

}

func (c Client) GetAllVolumes(reqContext context.Context) (*GetAllVolumesResponse, error) {

	url := fmt.Sprintf("http://%s:%s/volumes", c.Hostname, c.Port)
	fmt.Printf("url %s\n", url)
	mm, err := Get[[]model.Volume](reqContext, url)
	if err != nil {
		return nil, err
	}
	resp := GetAllVolumesResponse{
		Volumes: mm,
	}
	return &resp, nil
}
func (c Client) GetVolume(reqContext context.Context, id string) (*GetVolumeResponse, error) {

	url := fmt.Sprintf("http://%s:%s/volumes/%s", c.Hostname, c.Port, id)
	fmt.Printf("url %s\n", url)
	m, err := Get[model.Volume](reqContext, url)
	if err != nil {
		return nil, err
	}
	resp := GetVolumeResponse{
		Volume: m,
	}
	return &resp, nil
}

func Get[T any](ctx context.Context, url string) (T, error) {
	var m T
	r, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return m, err
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return m, err
	}
	switch res.StatusCode {
	case 200:
		fmt.Println("work!")
	case 404:
		fmt.Println("not found!")
		return m, fmt.Errorf("GET %s not found returned", url)
	default:
		fmt.Printf("bad status code from GET %d\n", res.StatusCode)
		return m, fmt.Errorf("GET %s bad statuscode %d returned", url, res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return m, err
	}
	return parseJSON[T](body)
}
func parseJSON[T any](s []byte) (T, error) {
	var r T
	if err := json.Unmarshal(s, &r); err != nil {
		return r, err
	}
	return r, nil
}

func toJSON(T any) ([]byte, error) {
	return json.Marshal(T)
}

func (c Client) CreateVolume(reqContext context.Context, newVolume model.Volume) (*CreateVolumeResponse, error) {

	url := fmt.Sprintf("http://%s:%s/volumes", c.Hostname, c.Port)
	fmt.Printf("url %s\n", url)
	newVolume, err := Post[model.Volume](reqContext, url, newVolume)
	if err != nil {
		return nil, err
	}
	resp := CreateVolumeResponse{
		Volume: newVolume,
	}
	return &resp, nil
}
func Post[T any](ctx context.Context, url string, data any) (T, error) {
	var m T
	b, err := toJSON(data)
	if err != nil {
		return m, err
	}
	byteReader := bytes.NewReader(b)
	r, err := http.NewRequestWithContext(ctx, "POST", url, byteReader)
	if err != nil {
		return m, err
	}
	// Important to set
	r.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return m, err
	}
	switch res.StatusCode {
	case 200, 201:
		fmt.Println("POST work!")
	case 404:
		fmt.Println("not found!")
		return m, fmt.Errorf("POST %s not found returned", url)
	default:
		fmt.Printf("bad status code from POST %d\n", res.StatusCode)
		return m, fmt.Errorf("POST %s bad statuscode %d returned", url, res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return m, err
	}
	return parseJSON[T](body)
}

func (c Client) DeleteVolume(reqContext context.Context, id string) error {

	url := fmt.Sprintf("http://%s:%s/volumes/%s", c.Hostname, c.Port, id)
	fmt.Printf("url %s\n", url)
	r, err := http.NewRequestWithContext(reqContext, "DELETE", url, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	switch res.StatusCode {
	case 200:
		fmt.Println("work!")
	case 404:
		fmt.Println("not found!")
		return fmt.Errorf("DELETE %s not found returned", url)
	default:
		fmt.Printf("bad status code from DELETE %d\n", res.StatusCode)
		return fmt.Errorf("DELETE %s bad statuscode %d returned", url, res.StatusCode)
	}
	/**
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	*/
	return nil
}
