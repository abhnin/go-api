package storage

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"twreporter.org/go-api/models"
	"twreporter.org/go-api/utils"
	//log "github.com/Sirupsen/logrus"
)

// NewsStorage defines the methods we need to implement,
// in order to provide the news resource to twreporter main site.
type NewsStorage interface {
	/** Close DB Connection **/
	Close() error

	/** Posts methods **/
	GetMetaOfPosts(interface{}, int, int, string, []string) ([]models.PostMeta, error)
	GetTopics(interface{}, int, int, string, []string) ([]models.Topic, error)
}

// NewMongoStorage initializes the storage connected to Mongo database
func NewMongoStorage(db *mgo.Session) *MongoStorage {
	return &MongoStorage{db}
}

// MongoStorage implements `NewsStorage`
type MongoStorage struct {
	db *mgo.Session
}

// Close quits the DB connection gracefully
func (m *MongoStorage) Close() error {
	m.db.Close()
	return nil
}

// GetDocuments ...
func (m *MongoStorage) GetDocuments(qs interface{}, limit int, offset int, sort string, collection string, documents interface{}) error {
	var err error
	var q models.MongoQuery

	_qs, ok := qs.(string)

	if ok {
		err = models.GetQuery(_qs, &q)

		if err != nil {
			return m.NewStorageError(err, "GetDocuments", "storage.mongo_storage.get_documents.parse_query_error")
		}

		qs = q
	}

	err = m.db.DB(utils.Cfg.MongoDBSettings.DBName).C(collection).Find(qs).Limit(limit).Skip(offset).Sort(sort).All(documents)

	if err != nil {
		return m.NewStorageError(err, "GetDocuments", "storage.mongo_storage.get_documents_error")
	}

	return nil
}

// GetDocument ...
func (m *MongoStorage) GetDocument(id bson.ObjectId, collection string, doc interface{}) error {
	if id == "" {
		return m.NewStorageError(ErrMgoNotFound, "GetDocument", "storage.mongo_storage.get_document.id_not_provided")
	}

	err := m.db.DB(utils.Cfg.MongoDBSettings.DBName).C(collection).FindId(id).One(doc)

	if err != nil {
		return m.NewStorageError(err, "GetDocument", "storage.mongo_storage.get_document.error")
	}
	return nil
}