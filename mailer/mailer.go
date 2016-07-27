package mailer

import (
	"github.com/keighl/mandrill"
	log "github.com/opsee/logrus"
)

var (
	Client  *mandrill.Client
	BaseURL string
)

func init() {
	if BaseURL == "" {
		log.Info("no base url configured for mandrill messages, defaulting to https://app.opsee.com")
		BaseURL = "https://app.opsee.com"
	}
}

func Send(toEmail, toName, templateName string, mergeVars map[string]interface{}) ([]*mandrill.Response, error) {
	if Client == nil {
		log.Info("no mandrill client configured, not sending message")
		return nil, nil
	}

	mergeVars["opsee_host"] = BaseURL

	message := &mandrill.Message{}
	message.AddRecipient(toEmail, toName, "to")
	message.Merge = true
	message.MergeLanguage = "handlebars"
	message.MergeVars = []*mandrill.RcptMergeVars{mandrill.MapToRecipientVars(toEmail, mergeVars)}
	return Client.MessagesSendTemplate(message, templateName, map[string]string{})
}
