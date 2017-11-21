package isi

// Vocab user define vocab
type Vocab struct {
	RequestID    string         `json:"request_id,omitempty"`
	GlobalWeight int            `json:"global_weight,omitempty"`
	Words        []string       `json:"words,omitempty"`
	WordWeights  map[string]int `json:"word_weights,omitempty"`
}

type vocabRet struct {
	RequestID    string `json:"request_id,omitempty"`
	VocabularyID string `json:"vocabulary_id,omitempty"`
	ErrorCode    int    `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	Result       string `json:"result,omitempty"`
}

// TransBody body json
type TransBody struct {
	AppKey       string `json:"app_key,omitempty"`
	OssLink      string `json:"oss_link,omitempty"`
	VocabularyID string `json:"vocabulary_id,omitempty"`
}

type transID struct {
	ID string `json:"id,omitempty"`
}

type transRet struct {
	Status string `json:"status,omitempty"`
	Result []struct {
		Text string `json:"text,omitempty"`
	} `json:"result,omitempty"`
}
