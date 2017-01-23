package backends

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
)

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

	for {
		payload := <-saveMailChan
		if payload == nil {
			log.Info("No more saveMailChan payload")
			return
		}

		//recipient := payload.recipient.User + "@" + b.config.PrimaryHost
		recipient := payload.recipient.User + "@phantomail.com"
		length := payload.mail.Data.Len()
		log.Info("length=", length)
		receivedAt := fmt.Sprintf("%d", time.Now().UnixNano())
		payload.mail.ParseHeaders()
		hash := MD5Hex(
			recipient,
			payload.mail.MailFrom.String(),
			payload.mail.Subject,
			receivedAt,
		)
		log.Info("hash=", hash)
	}
}
