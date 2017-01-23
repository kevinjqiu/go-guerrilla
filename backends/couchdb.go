package backends

import (
	"time"

	log "github.com/Sirupsen/logrus"
	couchdb "github.com/rhinoman/couchdb-go"
)

type Metadata struct {
	Sender        string              `json:"sender"`
	Recipients    []string            `json:"recipients"`
	ReceivedAt    string              `json:"receivedAt"`
	RemoteAddress string              `json:"remoteAddress"`
	Subject       string              `json:"subject"`
	Header        map[string][]string `json:"header"`
}

type EmailDocument struct {
	Metadata Metadata `json:"metadata"`
	Body     []byte   `json:"body"`
}

func init() {
	backends["couchdb"] = &AbstractBackend{
		extend: &CouchDBBackend{},
	}
}

type CouchDBConfig struct {
	Host     string `json:"couchdb_host"`
	User     string `json:"couchdb_user"`
	Password string `json:"couchdb_password"`
	DB       string `json:"couchdb_db"`
}

type CouchDBBackend struct {
	AbstractBackend
	config CouchDBConfig
}

func (b *CouchDBBackend) loadConfig(backendConfig BackendConfig) (err error) {
	configType := baseConfig(&CouchDBConfig{})
	bcfg, err := b.extractConfig(backendConfig, configType)
	if err != nil {
		return err
	}
	config := bcfg.(*CouchDBConfig)
	b.config = *config
	return nil
}

func (b *CouchDBBackend) saveMailWorker(saveMailChan chan *savePayload) {
	log.Info("Save Called")

	conn, err := couchdb.NewConnection("localhost", 5984, time.Duration(500*time.Millisecond))
	if err != nil {
		panic(err) // TODO: signal error response
	}

	db := conn.SelectDB("phantomail", &couchdb.BasicAuth{Username: "admin", Password: "password"})

	for {
		payload := <-saveMailChan
		if payload == nil {
			log.Info("No more saveMailChan payload")
			return
		}

		recipient := payload.recipient.User + "@" + payload.recipient.Host
		length := payload.mail.Data.Len()
		log.Info("length=", length)
		receivedAt := time.Now().UTC().Format(time.RFC3339)
		payload.mail.ParseHeaders()
		hash := MD5Hex(
			recipient,
			payload.mail.MailFrom.String(),
			payload.mail.Subject,
			receivedAt,
		)
		log.Info("hash=", hash)

		emailDoc := EmailDocument{
			Metadata{
				Sender:        payload.from.User + "@" + payload.from.Host,
				Recipients:    []string{recipient},
				ReceivedAt:    receivedAt,
				RemoteAddress: payload.mail.RemoteAddress,
				Subject:       payload.mail.Subject,
				Header:        payload.mail.Header,
			},
			payload.mail.Data.Bytes(),
		}

		rev, err := db.Save(emailDoc, hash, "")
		if err != nil {
			panic(err)
		}

		log.Info("Document saved: ", rev)
		payload.savedNotify <- &saveStatus{nil, hash}
	}
}
