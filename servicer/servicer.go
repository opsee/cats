package servicer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/hoisie/mustache"
	"github.com/keighl/mandrill"
	"github.com/opsee/basic/schema"
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

// 0111b -- TODO(dan) AllPerms(permset) (uint64, error) in basic
const AllUserPerms = uint64(0x7)

type Signup struct {
	Id         int               `json:"id"`
	Email      string            `json:"email"`
	Name       string            `json:"name"`
	Claimed    bool              `json:"claimed"`
	Activated  bool              `json:"activated"`
	Referrer   string            `json:"referrer"`
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at" db:"updated_at"`
	CustomerId string            `json:"-" db:"customer_id"`
	Perms      *schema.UserFlags `json:"-" db:"perms"`
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

type Config struct {
	Host        string
	MandrillKey string
	IntercomKey string
	CloseIOKey  string
	SlackUrl    string
}

func Init(config Config) error {
	opseeHost = config.Host
	intercomKey = []byte(config.IntercomKey)
	slackEndpoint = config.SlackUrl

	if config.MandrillKey != "" {
		mailClient = mandrill.ClientWithKey(config.MandrillKey)
	}

	if config.CloseIOKey != "" {
		closeioClient = closeio.New(config.CloseIOKey)
	}

	conn, err := grpc.Dial("spanx.in.opsee.com:8443", grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
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
