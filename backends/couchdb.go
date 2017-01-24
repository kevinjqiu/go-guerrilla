package backends

import (
	"bytes"
	"time"

	log "github.com/Sirupsen/logrus"
	couchdb "github.com/rhinoman/couchdb-go"
)

// Metadata store the metadata of the email
type Metadata struct {
	Sender        string              `json:"sender"`
	Recipients    []string            `json:"recipients"`
	ReceivedAt    string              `json:"receivedAt"`
	RemoteAddress string              `json:"remoteAddress"`
	Subject       string              `json:"subject"`
	Header        map[string][]string `json:"header"`
	BodyLength    int                 `json:"bodyLength"`
}

// EmailDocument represents a received email
type EmailDocument struct {
	Metadata Metadata `json:"metadata"`
}

func init() {
	backends["couchdb"] = &AbstractBackend{
		extend: &CouchDBBackend{},
	}
}

// CouchDBConfig is a collection of configuration for the couchdb backend
type CouchDBConfig struct {
	Host     string `json:"couchdb_host"`
	Port     int    `json:"couchdb_port"`
	User     string `json:"couchdb_user"`
	Password string `json:"couchdb_password"`
	DB       string `json:"couchdb_db"`
}

// CouchDBBackend is the couchdb backend for go-guerrilla
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
	conn, err := couchdb.NewConnection(b.config.Host, b.config.Port, time.Duration(500*time.Millisecond))
	if err != nil {
		panic(err) // TODO: signal error response
	}

	db := conn.SelectDB(b.config.DB, &couchdb.BasicAuth{Username: b.config.User, Password: b.config.Password})

	for {
		payload := <-saveMailChan
		if payload == nil {
			log.Info("No more saveMailChan payload")
			return
		}

		recipient := payload.recipient.User + "@" + payload.recipient.Host
		length := payload.mail.Data.Len()
		receivedAt := time.Now().UTC().Format(time.RFC3339)

		payload.mail.ParseHeaders()
		hash := MD5Hex(
			recipient,
			payload.mail.MailFrom.String(),
			payload.mail.Subject,
			receivedAt,
		)

		emailDoc := EmailDocument{
			Metadata{
				Sender:        payload.from.User + "@" + payload.from.Host,
				Recipients:    []string{recipient},
				ReceivedAt:    receivedAt,
				RemoteAddress: payload.mail.RemoteAddress,
				Subject:       payload.mail.Subject,
				Header:        payload.mail.Header,
				BodyLength:    length,
			},
		}

		rev, err := db.Save(emailDoc, hash, "")
		if err != nil {
			panic(err)
		}

		// TODO: handle error
		db.SaveAttachment(hash, rev, "emailBody", "text/plain", bytes.NewReader(payload.mail.Data.Bytes()))

		log.Info("Document saved: ", rev)
		payload.savedNotify <- &saveStatus{nil, hash}
	}
}
