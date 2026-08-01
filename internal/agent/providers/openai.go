package providers

type ProviderOpenAI struct {
	baseUrl string
	apiKey  string
}

func NewProviderOpenAI(baseURL string, apiKey string) Provider {
	return &ProviderOpenAI{baseURL, apiKey}
}

func (p *ProviderOpenAI) StreamChat(modelId string) {
	// TODO : call openai
}
