package servicer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hoisie/mustache"
	"github.com/keighl/mandrill"
	opsee "github.com/opsee/basic/service"
	log "github.com/opsee/logrus"
	slacktmpl "github.com/opsee/notification-templates/dist/go/slack"
	"github.com/snorecone/closeio-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type MandrillMailer interface {
	MessagesSendTemplate(*mandrill.Message, string, interface{}) ([]*mandrill.Response, error)
}

var (
	opseeHost       string
	mailClient      MandrillMailer
	intercomKey     []byte
	closeioClient   *closeio.Closeio
	slackEndpoint   string
	slackTemplates  map[string]*mustache.Template
	slackDomain     string
	slackAdminToken string
	spanxClient     opsee.SpanxClient
)

func init() {
	slackTemplates = make(map[string]*mustache.Template)

	tmpl, err := mustache.ParseString(slacktmpl.NewSignup)
	if err != nil {
		panic(err)
	}

	slackTemplates["new-signup"] = tmpl
}

func Init(host string, mailer MandrillMailer, intercom, closeioKey, slackUrl, inviteSlackDomain, inviteSlackAdminToken, spanxHost string) error {
	opseeHost = host
	mailClient = mailer
	intercomKey = []byte(intercom)
	slackEndpoint = slackUrl
	slackDomain = inviteSlackDomain
	slackAdminToken = inviteSlackAdminToken

	if closeioKey != "" {
		closeioClient = closeio.New(closeioKey)
	}

	conn, err := grpc.Dial(spanxHost, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		return err
	}

	spanxClient = opsee.NewSpanxClient(conn)

	return nil
}

func mailTemplatedMessage(toEmail, toName, templateName string, mergeVars map[string]interface{}) ([]*mandrill.Response, error) {
	if mailClient == nil {
		return nil, nil
	}

	mergeVars["opsee_host"] = opseeHost

	message := &mandrill.Message{}
	message.AddRecipient(toEmail, toName, "to")
	message.Merge = true
	message.MergeLanguage = "handlebars"
	message.MergeVars = []*mandrill.RcptMergeVars{mandrill.MapToRecipientVars(toEmail, mergeVars)}
	return mailClient.MessagesSendTemplate(message, templateName, map[string]string{})
}

func createLead(lead *closeio.Lead) {
	if closeioClient != nil {
		resp, err := closeioClient.CreateLead(lead)
		if err != nil {
			log.Print(err.Error())
		} else {
			log.Printf("created closeio lead: %s", resp.Url)
		}
	}
}

func notifySlack(name string, vars map[string]interface{}) {
	log.Info("requested slack notification")

	template, ok := slackTemplates[name]
	if !ok {
		log.Errorf("not sending slack notification since template %s was not found", name)
		return
	}

	body := template.Render(vars)

	if slackEndpoint == "" {
		log.Warn("not sending slack notification since SLACK_ENDPOINT is not set")
		fmt.Println(body)
		return
	}

	resp, err := http.Post(slackEndpoint, "application/json", bytes.NewBufferString(body))
	if err != nil {
		log.WithError(err).Errorf("failed to send slack notification: %s", name)
		return
	}

	defer resp.Body.Close()
	log.WithField("status", resp.StatusCode).Info("sent slack request")
}

func inviteSlack(name, email string) {
	if slackDomain == "" || slackAdminToken == "" {
		log.Warn("not inviting user to the opsee support slack")
		return
	}

	log.Info("inviting user to the opsee support slack")

	v := url.Values{}
	v.Set("email", email)
	v.Set("token", slackAdminToken)
	v.Set("set_active", "true")
	v.Set("extra_message", fmt.Sprintf(`Thanks for signing up for Opsee, %s. If you use Slack, you can chat with our engineering and support teams any time in the opsee-support Slack team.`, name))

	resp, err := http.PostForm(fmt.Sprintf("https://%s/api/users.admin.invite", slackDomain), v)
	if err != nil {
		log.WithError(err).Errorf("failed to send slack invitation to: %s", email)
		return
	}

	defer resp.Body.Close()
	log.Infof("slack invitation sent to %s", email)
}
