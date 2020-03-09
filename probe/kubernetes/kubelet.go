package kubernetes

import (
	"crypto/tls"
	"fmt"
	"github.com/weaveworks/scope/vendor/k8s.io/client-go/transport"
	restclient "k8s.io/client-go/rest"
	"net/http"

	"github.com/ugorji/go/codec"
)

// Intentionally not using the full kubernetes library DS
// to make parsing faster and more tolerant to schema changes
type podList struct {
	Items []struct {
		Metadata struct {
			UID string `json:"uid"`
		} `json:"metadata"`
	} `json:"items"`
}

// kubeletClientConfig contains the TLS configuration to get access to the kubelet endpoint with TLS enabled
type kubeletClientConfig struct {
	restclient.TLSClientConfig
}

// GetLocalPodUIDs obtains the UID of the pods run locally (it's just exported for testing)
var GetLocalPodUIDs = func(kubeletHost string) (map[string]struct{}, error) {
	url := fmt.Sprintf("https://%s/pods/", kubeletHost)
	config := kubeletClientConfig{}
	client := kubeletClient(config)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var localPods podList
	if err := codec.NewDecoder(resp.Body, &codec.JsonHandle{}).Decode(&localPods); err != nil {
		return nil, err
	}
	result := make(map[string]struct{}, len(localPods.Items))
	for _, pod := range localPods.Items {
		result[pod.Metadata.UID] = struct{}{}
	}
	return result, nil
}

func kubeletClient(config kubeletClientConfig) *http.Client {
	cfg := &transport.Config{
		TLS: transport.TLSConfig{
			CAFile:   config.CAFile,
			CAData:   config.CAData,
			CertFile: config.CertFile,
			CertData: config.CertData,
			KeyFile:  config.KeyFile,
			KeyData:  config.KeyData,
		},
	}
	transport := http.Transport{TLSClientConfig: cfg}
	return &http.Client{}
}
