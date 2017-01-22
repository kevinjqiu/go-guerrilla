package backends

import (
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
	log.Info("loadConfig")
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

}
