package medic

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func promptUnseal(client *http.Client, cfg Config, opt Options) error {
	if opt.Quiet || opt.JSON {
		return nil
	}
	fmt.Println()
	fmt.Printf("%s Do you want to unseal now? [y/N]: ", cwrap("Node is sealed.", colYellow, opt))
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer != "y" && answer != "yes" {
		return nil
	}

	for {
		fmt.Print("Enter unseal key (blank to stop): ")
		byteKey, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return err
		}
		key := strings.TrimSpace(string(byteKey))
		if key == "" {
			break
		}
		sealed, err := unsealOnce(client, cfg, key)
		if err != nil {
			return err
		}
		if !sealed {
			fmt.Println(cwrap("Vault successfully unsealed!", colGreen, opt))
			break
		} else {
			fmt.Println(cwrap("Partial unseal, more keys required...", colYellow, opt))
		}
	}
	return nil
}

func unsealOnce(client *http.Client, cfg Config, key string) (bool, error) {
	type req struct {
		Key string `json:"key"`
	}
	type resp struct {
		Sealed bool `json:"sealed"`
	}

	url := strings.TrimRight(cfg.Addr, "/") + "/v1/sys/unseal"
	body, _ := json.Marshal(req{Key: key})
	httpReq := must(NewRequestJSON(http.MethodPost, url, body))
	withVaultHeaders(httpReq, cfg)

	res, err := client.Do(httpReq)
	if err != nil {
		return true, err
	}
	defer res.Body.Close()

	var out resp
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return true, err
	}
	return out.Sealed, nil
}
