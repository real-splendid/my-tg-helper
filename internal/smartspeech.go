package internal

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type OAuthResponse struct {
	AccessToken string `json:"access_token"`
}

// GetSberToken obtains OAuth token from Sber API
func GetSberToken(authToken string) (string, error) {
	endpoint := "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
	data := url.Values{}
	data.Set("scope", "SALUTE_SPEECH_PERS")

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+authToken)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("bad response status %d: %s", resp.StatusCode, string(body))
	}

	var oauthResp OAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&oauthResp); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return oauthResp.AccessToken, nil
}

// SynthesizeText converts text to speech using Sber SmartSpeech API
func SynthesizeText(text string, accessToken string) ([]byte, error) {
	endpoint := "https://smartspeech.sber.ru/rest/v1/text:synthesize"
	
	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(text))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/text")
	req.Header.Set("Accept", "audio/x-wav")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad response status %d: %s", resp.StatusCode, string(body))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return audioData, nil
}
