package backends

import (
	"github.com/boltdb/bolt"
	"github.com/flashmob/go-guerrilla/mail"
	"github.com/flashmob/go-guerrilla/response"
)

func init() {
	processors["boltdb"] = func() Decorator {
		return BoltDB()
	}
}

// BoltDBProcessorConfig allows the user to config BoltDB backend
type BoltDBProcessorConfig struct {
	BoltDBPath string `json:"boltdb_storage_path"`
}

// BoltDB is a decorator that stores email data in BoltDB
func BoltDB() Decorator {
	var config *BoltDBProcessorConfig

	Svc.AddInitializer(InitializeWith(func(backendConfig BackendConfig) error {
		configType := BaseConfig(&BoltDBProcessorConfig{})
		bcfg, err := Svc.ExtractConfig(backendConfig, configType)
		if err != nil {
			return err
		}
		config = bcfg.(*BoltDBProcessorConfig)
		return nil
	}))

	return func(p Processor) Processor {
		return ProcessWith(func(e *mail.Envelope, task SelectTask) (Result, error) {
			if task != TaskSaveMail {
				return p.Process(e, task)
			}

			if len(e.Hashes) == 0 {
				Log().Error("BoltDB needs a Hash() process before it")
				result := NewResult(response.Canned.FailBackendTransaction)
				return result, StorageError
			}

			hash := e.Hashes[0]
			e.QueuedId = hash

			db, err := bolt.Open(config.BoltDBPath, 0600, nil)
			if err != nil {
				Log().WithError(err).Warn("Error opening the db file: my.db")
				result := NewResult(response.Canned.FailBackendTransaction)
				return result, err
			}
			defer db.Close()

			err = db.Update(func(tx *bolt.Tx) error {
				var recipientBucket *bolt.Bucket
				for _, recipient := range e.RcptTo {
					recipientBucket, err = tx.CreateBucketIfNotExists([]byte(recipient.String()))
					if err != nil {
						Log().WithError(err).Warnf("Unable to create bucket for %s", recipientBucket)
						continue
					}
					err = recipientBucket.Put([]byte(hash), e.Data.Bytes())
					if err != nil {
						Log().WithError(err).Warn("Unable to save the email")
						continue
					}
				}
				return nil
			})
			return p.Process(e, task)
		})
	}
}
