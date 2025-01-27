package api

import (
    "bytes"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/http/cookiejar"
    "time"

    "github.com/jlengelbrecht/unifi-dns-sync/internal/models"
)

type UnifiClient struct {
    client  *http.Client
    baseURL string
    device  models.UnifiDevice
}

func NewUnifiClient(device models.UnifiDevice) (*UnifiClient, error) {
    jar, err := cookiejar.New(nil)
    if err != nil {
        return nil, err
    }

    client := &http.Client{
        Timeout: time.Second * 10,
        Jar:     jar,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        },
    }

    return &UnifiClient{
        client:  client,
        baseURL: fmt.Sprintf("https://%s", device.Address),
        device:  device,
    }, nil
}

func (c *UnifiClient) Login() error {
    loginData := map[string]string{
        "username": c.device.Credentials.Username,
        "password": c.device.Credentials.Password,
    }

    jsonData, err := json.Marshal(loginData)
    if err != nil {
        return err
    }

    resp, err := c.client.Post(c.baseURL+"/api/auth/login", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("login failed with status: %d", resp.StatusCode)
    }

    return nil
}

func (c *UnifiClient) GetDNSRecords() ([]models.DNSRecord, error) {
    resp, err := c.client.Get(c.baseURL + "/proxy/network/api/s/default/rest/dnsrecord")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to get DNS records: %d", resp.StatusCode)
    }

    var result struct {
        Data []models.DNSRecord `json:"data"`
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if err := json.Unmarshal(body, &result); err != nil {
        return nil, err
    }

    return result.Data, nil
}

func (c *UnifiClient) CreateDNSRecord(record models.DNSRecord) error {
    jsonData, err := json.Marshal(record)
    if err != nil {
        return err
    }

    resp, err := c.client.Post(
        c.baseURL+"/proxy/network/api/s/default/rest/dnsrecord",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to create DNS record: %d", resp.StatusCode)
    }

    return nil
}

func (c *UnifiClient) UpdateDNSRecord(record models.DNSRecord) error {
    jsonData, err := json.Marshal(record)
    if err != nil {
        return err
    }

    req, err := http.NewRequest(
        "PUT",
        fmt.Sprintf("%s/proxy/network/api/s/default/rest/dnsrecord/%s", c.baseURL, record.ID),
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    resp, err := c.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to update DNS record: %d", resp.StatusCode)
    }

    return nil
}

func (c *UnifiClient) DeleteDNSRecord(recordID string) error {
    req, err := http.NewRequest(
        "DELETE",
        fmt.Sprintf("%s/proxy/network/api/s/default/rest/dnsrecord/%s", c.baseURL, recordID),
        nil,
    )
    if err != nil {
        return err
    }

    resp, err := c.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to delete DNS record: %d", resp.StatusCode)
    }

    return nil
}
