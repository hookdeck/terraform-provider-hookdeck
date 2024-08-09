package sourceverification

import (
	"terraform-provider-hookdeck/internal/provider/sourceverification/generated"

	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func (m *sourceVerificationResourceModel) Refresh(verification *hookdeck.SourceVerification) {
}

func (m *sourceVerificationResourceModel) ToCreatePayload() *hookdeck.SourceUpdateRequest {
	return m.ToUpdatePayload()
}

func (m *sourceVerificationResourceModel) ToUpdatePayload() *hookdeck.SourceUpdateRequest {
	var verification *hookdeck.VerificationConfig

	for _, provider := range generated.Providers {
		if verification == nil {
			verification = provider.ToPayload(m.Verification)
		}
	}

	return &hookdeck.SourceUpdateRequest{
		Verification: hookdeck.Optional(*verification),
	}
}

func (m *sourceVerificationResourceModel) ToDeletePayload() *hookdeck.SourceUpdateRequest {
	return &hookdeck.SourceUpdateRequest{
		Verification: hookdeck.Null[hookdeck.VerificationConfig](),
	}
}
