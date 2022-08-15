package gorm_db

import (
	"errors"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.WithField("error", err).Panic("Failed to connect database")
		panic("failed to connect database")
	}

	return db, nil
}

func Find(db *gorm.DB, id interface{}, doc interface{}) (bool, error) {
	// TODO just find
	return FindByField(db, "id", id, doc)
}

func FindByField(db *gorm.DB, fieldName string, fieldValue interface{}, doc interface{}) (bool, error) {
	result := db.First(doc, fmt.Sprintf("\"%v\" = ?", fieldName), fieldValue)
	if result.Error != nil {
		notFound := errors.Is(result.Error, gorm.ErrRecordNotFound)
		if !notFound {
			log.WithFields(log.Fields{"field_name": fieldName, "field_value": fieldValue, "type": ObjectTypeName(doc), "error": result.Error}).Error("orm.FindByField: failed")
		}
		return notFound, result.Error
	}

	return false, nil
}

func FindByFields(db *gorm.DB, fields map[string]interface{}, doc interface{}) (bool, error) {
	result := db. /*.Debug()*/ Where(fields).First(doc)
	if result.Error != nil {
		notFound := errors.Is(result.Error, gorm.ErrRecordNotFound)
		if !notFound {
			log.WithFields(log.Fields{"fields": fields, "type": ObjectTypeName(doc), "error": result.Error}).Error("orm.FindByFields: failed")
		}
		return notFound, result.Error
	}

	return false, nil
}

func FindAll(db *gorm.DB, docs interface{}) error {
	result := db.Find(docs)
	if result.Error != nil {
		log.WithFields(log.Fields{"type": reflect.TypeOf(docs).Name, "error": result.Error}).Error("orm.FindAll: failed")
		return result.Error
	}
	return nil
}

type Interval struct {
	From interface{}
	To   interface{}
}

func (i *Interval) IsNull() bool {
	return i.From == nil && i.To == nil
}

type Filter struct {
	PreconditionFields      map[string]interface{}
	IntervalFields          map[string]*Interval
	PreconditionFieldsIn    map[string][]interface{}
	PreconditionFieldsNotIn map[string][]interface{}

	Field         string
	SortField     string
	SortDirection string
	Offset        int
	Limit         int
	Interval
	In []string
}

func prepareInterval(db *gorm.DB, name string, interval *Interval) *gorm.DB {
	h := db

	if interval.From != nil && interval.To != nil {
		if interval.From == interval.To {
			h = h.Where(fmt.Sprintf("\"%v\" = ?", name), interval.From)
		} else {
			h = h.Where(fmt.Sprintf("\"%v\" >= ? AND \"%v\" <= ? ", name, name), interval.From, interval.To)
		}
	} else if interval.From != nil {
		h = h.Where(fmt.Sprintf("\"%v\" >= ? ", name), interval.From)
	} else if interval.To != nil {
		h = h.Where(fmt.Sprintf("\"%v\" <= ? ", name), interval.To)
	}
	return h
}

func prepareFilter(db *gorm.DB, filter *Filter) *gorm.DB {
	h := db

	if filter.PreconditionFields != nil {
		h = db.Where(filter.PreconditionFields)
	}

	if filter.PreconditionFieldsIn != nil {
		for field, values := range filter.PreconditionFieldsIn {
			h = h.Where(fmt.Sprintf("\"%v\" IN ? ", field), values)
		}
	}

	if filter.PreconditionFieldsNotIn != nil {
		for field, values := range filter.PreconditionFieldsNotIn {
			h = h.Where(fmt.Sprintf("\"%v\" NOT IN ? ", field), values)
		}
	}

	for name, interval := range filter.IntervalFields {
		h = prepareInterval(h, name, interval)
	}

	if filter.Field != "" {
		if filter.In != nil {
			h = h.Where(fmt.Sprintf("\"%v\" IN ? ", filter.Field), filter.In)
		} else {
			prepareInterval(h, filter.Field, &filter.Interval)
		}
	}

	return h
}

func FindWithFilter(db *gorm.DB, filter *Filter, docs interface{}) error {

	h := prepareFilter(db, filter)

	if filter.SortField != "" && (filter.SortDirection == "asc" || filter.SortDirection == "desc") {
		h = h.Order(fmt.Sprintf("\"%v\" %v", filter.SortField, filter.SortDirection))
	}

	if filter.Offset > 0 {
		h = h.Offset(filter.Offset)
	}

	if filter.Limit > 0 {
		h = h.Limit(filter.Limit)
	}

	// result := h.Debug().Find(docs)
	result := h.Find(docs)
	if result.Error != nil {
		log.WithFields(log.Fields{"filter": filter, "type": reflect.TypeOf(docs).Name, "error": result.Error}).Error("orm.FindAll: failed")
		return result.Error
	}
	return nil
}

func CountWithFilter(db *gorm.DB, filter *Filter, doc interface{}) int64 {

	m := db.Model(doc)
	h := prepareFilter(m, filter)

	var count int64
	h.Count(&count)
	return count
}

func SumWithFilter(db *gorm.DB, filter *Filter, fields map[string]string, doc interface{}, result interface{}) error {

	sums := ""
	for key, name := range fields {
		if sums != "" {
			sums += ", "
		}
		sums += fmt.Sprintf("sum(%v) as %v", key, name)
	}

	m := db.Model(doc).Select(sums)
	h := prepareFilter(m, filter)

	r := h.Take(result)
	return r.Error
}

func FindAllByFields(db *gorm.DB, fields map[string]interface{}, docs interface{}) error {
	result := db.Where(fields).Find(docs)
	if result.Error != nil {
		log.WithFields(log.Fields{"fields": fields, "type": reflect.TypeOf(docs).Name, "error": result.Error}).Error("orm.FindAllByFields: failed")
		return result.Error
	}
	return nil
}

func FindNotIn(db *gorm.DB, fields map[string]interface{}, docs interface{}) error {
	result := db.Not(fields).Find(docs)
	if result.Error != nil {
		log.WithFields(log.Fields{"fields": fields, "type": reflect.TypeOf(docs).Name, "error": result.Error}).Error("orm.FindNotIn: failed")
		return result.Error
	}
	return nil
}

func FindSelectNotIn(db *gorm.DB, fields map[string]interface{}, docModel interface{}, docs interface{}) error {
	result := db.Model(docModel).Not(fields).Find(docs)
	if result.Error != nil {
		log.WithFields(log.Fields{"fields": fields, "type": reflect.TypeOf(docs).Name, "error": result.Error}).Error("orm.FindSelectNotIn: failed")
		return result.Error
	}
	return nil
}

func RemoveById(db *gorm.DB, id interface{}, doc interface{}) error {
	result := db.Where("id = ?", id).Delete(doc)
	if result.Error != nil {
		log.WithFields(log.Fields{"id": id, "type": ObjectTypeName(doc), "error": result.Error}).Error("orm.RemoveById: failed")
	}
	return result.Error
}

func RemoveByField(db *gorm.DB, field string, value interface{}, doc interface{}) error {
	result := db.Where(fmt.Sprintf("\"%v\" = ?", field), value).Delete(doc)
	if result.Error != nil {
		log.WithFields(log.Fields{"field": field, "value": value, "type": ObjectTypeName(doc), "error": result.Error}).Error("orm.RemoveByField: failed")
	}
	return result.Error
}

func Create(db *gorm.DB, doc interface{}) error {
	result := db.Create(doc)
	if result.Error != nil {
		log.WithFields(log.Fields{"type": ObjectTypeName(doc), "error": result.Error}).Error("orm.Create: failed")
	}
	return result.Error
}

func UpdateFields(db *gorm.DB, fields []string, doc interface{}) error {
	result := db.Model(doc).Select(fields).Updates(doc)
	if result.Error != nil {
		log.WithFields(log.Fields{"type": ObjectTypeName(doc), "error": result.Error}).Error("orm.UpdateFields: failed")
	}
	return result.Error
}

func UpdateField(db *gorm.DB, field string, doc interface{}) error {
	result := db.Model(doc).Select(field).Updates(doc)
	if result.Error != nil {
		log.WithFields(log.Fields{"type": ObjectTypeName(doc), "error": result.Error}).Error("orm.UpdateField: failed")
	}
	return result.Error
}

func collectFieldNames(t reflect.Type, names *[]string) {

	// Return if not struct or pointer to struct.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	// Iterate through fields collecting names in map.
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)

		// Recurse into anonymous fields.
		if sf.Anonymous {
			if sf.Name != "BaseObject" {
				collectFieldNames(sf.Type, names)
			}
		} else {
			*names = append(*names, sf.Name)
		}
	}
}

func Update(db *gorm.DB, doc interface{}) error {

	v := reflect.ValueOf(doc)
	if reflect.ValueOf(doc).Kind() == reflect.Ptr {
		v = v.Elem()
	}

	fields := make([]string, 0)
	collectFieldNames(v.Type(), &fields)

	return UpdateFields(db, fields, doc)
}

type TransactionHandler func(tx *gorm.DB) error

func Transaction(db *gorm.DB, handler TransactionHandler) error {
	return db.Transaction(handler)
}

func ObjectTypeName(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if reflect.ValueOf(obj).Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func UpdateFieldMulti(db *gorm.DB, fields map[string]interface{}, doc interface{}, field string, value interface{}) error {
	result := db.Model(doc).Where(fields).Update(field, value)
	if result.Error != nil {
		log.WithFields(log.Fields{"fields": fields, "type": ObjectTypeName(doc), "error": result.Error}).Error("orm.UpdateFieldMulti: failed")
	}
	return result.Error
}

func DeleteAllByFields(db *gorm.DB, fields map[string]interface{}, docs interface{}) error {
	result := db.Where(fields).Delete(docs)
	if result.Error != nil {
		log.WithFields(log.Fields{"fields": fields, "type": reflect.TypeOf(docs).Name, "error": result.Error}).Error("orm.DeleteAllByFields: failed")
		return result.Error
	}
	return nil
}

func FindAllInterval(db *gorm.DB, name string, interval *Interval, docs interface{}) error {
	h := prepareInterval(db, name, interval)
	result := h.Find(docs)
	if result.Error != nil {
		log.WithFields(log.Fields{"field": name, "type": reflect.TypeOf(docs).Name, "error": result.Error}).Error("orm.FindAllInterval: failed")
		return result.Error
	}
	return nil
}

func DeleteAll(db *gorm.DB, docs interface{}) error {
	result := db.Where("1 = 1").Delete(docs)
	if result.Error != nil {
		log.WithFields(log.Fields{"type": reflect.TypeOf(docs).Name, "error": result.Error}).Error("orm.DeleteAll: failed")
		return result.Error
	}
	return nil
}
