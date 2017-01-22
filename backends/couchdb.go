package backends

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
