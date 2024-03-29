package crud

import (
	"fmt"
	"github.com/ewa-go/ewa"
	"github.com/ewa-go/ewa/consts"
	"github.com/ewa-go/ewa/security"
	"strings"
	"time"
)

type Route struct {
	Route       *ewa.Route
	Audit       IAudit
	CRUD        ICRUD
	FieldIdName string
	ModelName   string
	IsMapTable  bool
	Excludes    []string
	Headers     HeaderValues
	StatusDict  StatusDict

	Handler       ewa.Handler
	BeforeHandler Handler
	AfterHandler  Handler
}

type IAudit interface {
	String(statusCode int, data string) (int, string)
	JSON(statusCode int, v any) (int, any)
	Insert(tableName, action string, identity Identity, path string)
}

type ICRUD interface {
	Columns(tableName string, fields ...string) (Maps, error)
	Create(tableName string, data interface{}) (i uint, err error)
	GetRecord(tableName string, f Filter, query string, args ...interface{}) (Map, error)
	GetRecords(tableName string, f Filter, query string, args ...interface{}) (Maps, int64, error)
	Update(tableName string, v interface{}, query interface{}, args ...interface{}) error
	Delete(tableName string, v interface{}, query interface{}, args ...interface{}) error
}

type Handler func(*ewa.Context, Route, Identity, *Body) (int, error)

type Identity struct {
	Author   string
	Datetime time.Time
	fields   map[string]interface{}
}

// Get Получить значение по ключу
func (i Identity) Get(key string) interface{} {
	if value, ok := i.fields[key]; ok {
		return value
	}
	return nil
}

// Is Проверка на существование ключа
func (i Identity) Is(key string) bool {
	_, ok := i.fields[key]
	return ok
}

type StatusDict map[int]string

func (e StatusDict) Get(status int, def ...string) (s int, v string) {
	if val, ok := e[status]; ok {
		v = val
	} else {
		if def != nil {
			v = def[0]
		}
	}
	return status, v
}

func NewRoute(r *ewa.Route) Route {
	errDict := StatusDict{
		422: "Ошибка возвращаемых данных",
		412: "Ошибка",
		404: "Не найден",
		400: "Ошибка входных",
	}
	return Route{
		Route:      r,
		StatusDict: errDict,
	}
}

// SetErrorDict Справочник ошибок
func (r Route) SetErrorDict(errorDict map[int]string) Route {
	r.StatusDict = errorDict
	return r
}

// SetAudit Установка интерфейса для аудита
func (r Route) SetAudit(audit IAudit) Route {
	r.Audit = audit
	return r
}

// SetCRUD Установка интерфейса для CRUD функций
func (r Route) SetCRUD(crud ICRUD) Route {
	r.CRUD = crud
	return r
}

// SetHeader Установка заголовков
func (r Route) SetHeader(key, value string, isDefault ...bool) Route {
	r.Headers.Add(key, value, isDefault...)
	return r
}

// SetModelName Установка имени модели - таблицы
func (r Route) SetModelName(fieldIdName string) Route {
	r.FieldIdName = fieldIdName
	return r
}

func (r Route) SetFieldIdName(modelName string) Route {
	r.ModelName = modelName
	return r
}

// MapTable Установка флага для связной таблицы
func (r Route) MapTable() Route {
	r.IsMapTable = true
	return r
}

// SetExcludes Установка исключения полей из данных
func (r Route) SetExcludes(excludes ...string) Route {
	r.Excludes = append(r.Excludes, excludes...)
	return r
}

// SetHandler Установка обработчика маршрута
func (r Route) SetHandler(h func(*ewa.Context, Route) error) Route {
	r.Route.Handler = func(c *ewa.Context) error {
		return h(c, r)
	}
	return r
}

// SetBeforeHandler Установка обработчика до действий в бд
func (r Route) SetBeforeHandler(h Handler) Route {
	r.BeforeHandler = h
	return r
}

// SetAfterHandler Установка обработчика после действий в бд
func (r Route) SetAfterHandler(h Handler) Route {
	r.AfterHandler = h
	return r
}

// Identity Извлечение идентификации
func (r Route) Identity(identity *security.Identity) (i Identity) {
	if identity != nil {
		i.Author = identity.Username
		i.Datetime = identity.Datetime
		i.fields = identity.Variables
	}
	return
}

// ReadHandler Обработчик получения записей
func (r Route) ReadHandler(c *ewa.Context) error {

	r.ModelName = r.GetValueFromHeader(c, HeaderTableType)

	identity := r.Identity(c.Identity)
	defer r.Audit.Insert(Read, r.ModelName, identity, c.Path())

	tableInfo := c.Get(HeaderTableInfo)
	if len(tableInfo) > 0 {
		var fields []string
		if tableInfo != "full" {
			fields = strings.Split(tableInfo, ",")
		}
		columns, err := r.CRUD.Columns(r.ModelName, fields...)
		if err != nil {
			return c.SendString(r.Audit.String(consts.StatusBadRequest, err.Error()))
		}
		return c.JSON(r.Audit.JSON(200, columns))
	}

	// Получаем фильтр
	filter, err := r.GetFilter(c)
	if err != nil {
		return c.SendString(r.Audit.String(consts.StatusBadRequest, err.Error()))
	}

	// Обработчик до обращения в бд
	if r.BeforeHandler != nil {
		if status, err := r.BeforeHandler(c, r, identity, nil); err != nil {
			return c.SendString(r.Audit.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	// Если есть id возвращаем только одну запись
	id := c.Params(r.FieldIdName)
	if len(id) > 0 {
		record, err := r.GetRecord(filter, fmt.Sprintf("%s=?", r.FieldIdName), id)
		if err != nil {
			return c.SendString(r.Audit.String(r.StatusDict.Get(consts.StatusUnprocessableEntity)))
		}
		record.Excludes(r.Excludes...)

		// Обработчик после обращению в бд
		if r.AfterHandler != nil {
			if status, err := r.AfterHandler(c, r, identity, nil); err != nil {
				return c.SendString(r.Audit.String(r.StatusDict.Get(status, err.Error())))
			}
		}

		return c.JSON(r.Audit.JSON(200, record))
	}

	// Вернуть записи
	records, total, err := r.GetRecords(filter, r.GetQuery(c, filter.Fields))
	if err != nil {
		return c.SendString(r.Audit.String(r.StatusDict.Get(consts.StatusUnprocessableEntity)))
	}
	// Заголовок Total
	c.Set("Total", fmt.Sprintf("%d", total))
	records.Excludes(r.Excludes...)

	// Обработчик после обращению в бд
	if r.AfterHandler != nil {
		if status, err := r.AfterHandler(c, r, identity, nil); err != nil {
			return c.SendString(r.Audit.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	return c.JSON(r.Audit.JSON(200, records))
}

// CreateHandler Обработчик для создания записей
func (r Route) CreateHandler(c *ewa.Context) error {

	identity := r.Identity(c.Identity)
	defer r.Audit.Insert(Created, r.ModelName, identity, c.Path())

	body := NewBody(r.FieldIdName, NewFields("author", identity.Author)...)
	if err := body.Unmarshal(c.Body(), c.Get(HeaderXContentType) == "array"); err != nil {
		return c.SendString(r.Audit.String(consts.StatusBadRequest, err.Error()))
	}

	// Обработчик до обращения в бд
	if r.BeforeHandler != nil {
		if status, err := r.BeforeHandler(c, r, identity, nil); err != nil {
			return c.SendString(r.Audit.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	if body.IsArray {
		var resp []Response
		for i := range body.Array {
			id, err := r.CRUD.Create(r.ModelName, body.ToArrayMap(i))
			if err != nil {
				_, e := r.StatusDict.Get(422)
				resp = append(resp, NewCreated(id, e, false))
				continue
			}
			resp = append(resp, NewCreated(id, "OK", true))
		}

		// Обработчик после обращению в бд
		if r.AfterHandler != nil {
			if status, err := r.AfterHandler(c, r, identity, nil); err != nil {
				return c.SendString(r.Audit.String(r.StatusDict.Get(status, err.Error())))
			}
		}

		return c.JSON(200, resp)
	}
	id, err := r.CRUD.Create(r.ModelName, body.ToMap())
	if err != nil {
		return c.SendString(r.Audit.String(r.StatusDict.Get(consts.StatusUnprocessableEntity)))
	}

	// Обработчик после обращению в бд
	if r.AfterHandler != nil {
		if status, err := r.AfterHandler(c, r, identity, nil); err != nil {
			return c.SendString(r.Audit.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	return c.JSON(200, NewCreated(id, "OK", true))
}

// UpdateHandler Обновление записей
func (r Route) UpdateHandler(c *ewa.Context) error {

	identity := r.Identity(c.Identity)
	defer r.Audit.Insert(Updated, r.ModelName, identity, c.Path())

	// Получаем аргументы адресной строки
	var (
		values  []string
		idParam = c.Params(r.FieldIdName)
	)
	if len(idParam) > 0 {
		values = append(values, QueryFormat(r.FieldIdName, idParam).String())
	}
	// Получаем аргументы адресной строки
	c.QueryParams(func(key, value string) {
		values = append(values, QueryFormat(key, value).String())
	})

	if len(values) == 0 {
		return c.SendString(r.Audit.String(consts.StatusBadRequest, "Укажите поля для уточнения изменения записи! Пример: ../path?name=Name"))
	}

	body := NewBody(r.FieldIdName, NewFields("author", identity.Author)...)
	if err := body.Unmarshal(c.Body(), c.Get(HeaderXContentType) == "array"); err != nil {
		return c.SendString(r.Audit.String(consts.StatusBadRequest, err.Error()))
	}

	// Обработчик до обращения в бд
	if r.BeforeHandler != nil {
		if status, err := r.BeforeHandler(c, r, identity, body); err != nil {
			return c.SendString(r.Audit.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	// Пишем данные в бд
	if err := r.CRUD.Update(r.ModelName, body.ToMap(), strings.Join(values, " and ")); err != nil {
		return c.SendString(r.Audit.String(r.StatusDict.Get(consts.StatusUnprocessableEntity)))
	}

	// Обработчик после обращению в бд
	if r.AfterHandler != nil {
		if status, err := r.AfterHandler(c, r, identity, nil); err != nil {
			return c.SendString(r.Audit.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	return c.JSON(r.Audit.JSON(200, NewUpdated(body.GetField(r.FieldIdName), "OK", true)))
}

// DeleteHandler Обработчик удаления записей
func (r Route) DeleteHandler(c *ewa.Context) (err error) {

	identity := r.Identity(c.Identity)
	defer r.Audit.Insert(Deleted, r.ModelName, identity, c.Path())

	var (
		values  []string
		paramId = c.Params(r.FieldIdName)
		id      int
	)
	if len(paramId) > 0 {
		values = append(values, QueryFormat(r.FieldIdName, paramId).String())
	}

	// Получаем аргументы адресной строки
	c.QueryParams(func(key, value string) {
		values = append(values, QueryFormat(key, value).String())
	})

	if len(values) == 0 {
		return c.SendString(r.Audit.String(consts.StatusBadRequest, "Обязательно укажите условие для удаления записей! Пример: ?name=Name"))
	}

	// Обработчик до обращения в бд
	if r.BeforeHandler != nil {
		if status, err := r.BeforeHandler(c, r, identity, nil); err != nil {
			return c.SendString(r.Audit.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	// Удаление записи
	if err = r.CRUD.Delete(r.ModelName, nil, strings.Join(values, " and ")); err != nil {
		return c.SendString(r.Audit.String(r.StatusDict.Get(consts.StatusUnprocessableEntity)))
	}

	// Обработчик после обращению в бд
	if r.AfterHandler != nil {
		if status, err := r.AfterHandler(c, r, identity, nil); err != nil {
			return c.SendString(r.Audit.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	return c.JSON(r.Audit.JSON(200, NewDeleted(id, "OK", true)))
}

// GetFilter Инициализация фильтра
func (r Route) GetFilter(c *ewa.Context) (Filter, error) {
	body := c.Body()
	filterParam := c.QueryParam(filterName)
	if len(filterParam) > 0 {
		body = []byte(filterParam)
	}
	// Получаем фильтр
	return NewFilter(body)
}

// GetQuery Формирование запроса
func (r Route) GetQuery(c *ewa.Context, fields []string) string {
	var (
		values      []string
		valueFields []string
		queryParams = make(map[string]QueryParam)
		query       string
	)
	// Применение фильтра для запроса
	c.QueryParams(func(key, value string) {
		if key == filterName {
			return
		}
		q := QueryFormat(key, value)
		queryParams[q.Key] = q
	})

	for key, value := range queryParams {
		if key == "*" {
			continue
		}
		values = append(values, value.String())
	}

	if value, ok := queryParams["*"]; ok {
		// Параметр адресной строки *=
		if len(fields) > 0 {
			for _, field := range fields {
				if _, ok = queryParams[field]; !ok {
					value.Key = field
					valueFields = append(valueFields, value.String())
				}
			}
		} else {
			columnName := "column_name"
			columns, _ := r.CRUD.Columns(columnName)
			for _, column := range columns {
				field := fmt.Sprintf("%v", column[columnName])
				if _, ok = queryParams[field]; !ok {
					value.Key = field
					valueFields = append(valueFields, value.String())
				}
			}
		}
	}

	if len(valueFields) > 0 {
		query = "(" + strings.Join(valueFields, " or ") + ")"
	}
	if len(values) > 0 {
		v := strings.Join(values, " and ")
		if len(query) > 0 {
			query += " and " + v
		} else {
			query = v
		}
	}
	return query
}
