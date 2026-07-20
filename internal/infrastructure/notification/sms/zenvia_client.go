package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type Provider struct {
	client *http.Client
}

func NewProvider(client *http.Client) Provider {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	return Provider{client: client}
}

type Payload struct {
	From           string
	To             string
	Msg            string
	Schedule       *string
	CallbackOption string
	id             string
	AggregateId    string
	FlashSms       bool
}

func (p *Provider) Send(ctx context.Context, phone, message string) error {

	url := os.Getenv("ZENVIA_API_URL")
	token := os.Getenv("ZENVIA_API_TOKEN")
	from := os.Getenv("ZENVIA_FROM")

	data := Payload{
		From:           from,
		To:             phone,
		Msg:            message,
		Schedule:       nil,
		CallbackOption: "NONE",
		id:             uuid.New().String(),
		AggregateId:    "001",
		FlashSms:       false,
	}

	jsonData, err := json.Marshal(map[string]interface{}{"sendSmsRequest": data})
	if err != nil {
		fmt.Println("Erro ao converter para JSON:", err)
		return err
	}

	// Cria a requisição POST
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Erro ao criar requisição:", err)
		return err
	}

	// Define os headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+token)
	req.Header.Set("Accept", "application/json")

	// Envia a requisição
	resp, err := p.client.Do(req)
	if err != nil {
		fmt.Println("Numero:", phone)
		fmt.Println("Erro ao enviar requisição:", err)
		return err
	}
	defer resp.Body.Close()

	// Lê a resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler resposta:", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Numero:", phone)
		fmt.Println("Erro na resposta:", string(body))
		return fmt.Errorf("erro na resposta: %s", string(body))
	}

	return nil

}
