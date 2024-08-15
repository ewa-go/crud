package crud

import (
	"fmt"
	"github.com/ewa-go/ewa"
	"github.com/ewa-go/ewa/consts"
	"github.com/ewa-go/ewa/security"
	"strings"
	"time"
)

type CRUD struct {
	FieldIdName string
	ModelName   string
	Excludes    []string
	TableTypes  TableTypes
	StatusDict  StatusDict
	Variables   map[string]any

	IAudit
	IHandlers
	IResponse

	//BeforeHandler Handler
	//AfterHandler  Handler
}

type Handler func(*ewa.Context, CRUD, Identity, *QueryParams, *Body) (int, error)

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

func New(h IHandlers) *CRUD {

	var errDict = StatusDict{
		422: "Ошибка возвращаемых данных",
		412: "Ошибка",
		404: "Не найден",
		400: "Ошибка входных данных",
	}

	return &CRUD{
		IHandlers:  h,
		IAudit:     new(audit),
		IResponse:  new(Response),
		StatusDict: errDict,
	}
}

// SetErrorDict Справочник ошибок
func (r *CRUD) SetErrorDict(errorDict map[int]string) *CRUD {
	r.StatusDict = errorDict
	return r
}

// SetIAudit Установка интерфейса для аудита
func (r *CRUD) SetIAudit(audit IAudit) *CRUD {
	r.IAudit = audit
	return r
}

// SetIResponse Установка интерфейса для ответов
func (r *CRUD) SetIResponse(resp IResponse) *CRUD {
	r.IResponse = resp
	return r
}

// SetVariable Установка интерфейса для аудита
func (r *CRUD) SetVariable(key string, value any) *CRUD {
	if r.Variables == nil {
		r.Variables = make(map[string]any)
	}
	r.Variables[key] = value
	return r
}

// Is Проверка переменной на существование
func (r *CRUD) Is(key string) (ok bool) {
	_, ok = r.Variables[key]
	return ok
}

// SetTableType Установка заголовков
func (r *CRUD) SetTableType(key, value string, isDefault ...bool) *CRUD {
	r.TableTypes = r.TableTypes.Add(key, value, isDefault...)
	return r
}

// SetTableTypeTable Установка заголовков Table-Type:table
func (r *CRUD) SetTableTypeTable(value string, isDefault ...bool) *CRUD {
	r.TableTypes = r.TableTypes.Add("table", value, isDefault...)
	return r
}

// SetTableTypeView Установка заголовков Table-Type:view
func (r *CRUD) SetTableTypeView(value string, isDefault ...bool) *CRUD {
	r.TableTypes = r.TableTypes.Add("view", value, isDefault...)
	return r
}

// SetFieldIdName Установка имени идентификационного поля ../name/{id}
func (r *CRUD) SetFieldIdName(fieldIdName string) *CRUD {
	r.FieldIdName = fieldIdName
	return r
}

// SetModelName Установка имени модели - таблицы
func (r *CRUD) SetModelName(modelName string) *CRUD {
	r.ModelName = modelName
	return r
}

// SetExcludes Установка исключения полей из данных
func (r *CRUD) SetExcludes(excludes ...string) *CRUD {
	r.Excludes = append(r.Excludes, excludes...)
	return r
}

// SetBeforeHandler Установка обработчика до действий в бд
/*func (r *CRUD) SetBeforeHandler(h Handler) *CRUD {
	before = h
	return r
}*/

// SetAfterHandler Установка обработчика после действий в бд
/*func (r *CRUD) SetAfterHandler(h Handler) *CRUD {
	after = h
	return r
}*/

// NewIdentity Извлечение идентификации
func NewIdentity(identity *security.Identity) (i Identity) {
	if identity != nil {
		i.Author = identity.Username
		i.Datetime = identity.Datetime
		i.fields = identity.Variables
	}
	return
}

// NewQueryParams Извлечение параметров адресной строки
func (r *CRUD) NewQueryParams(c *ewa.Context, isFilter bool) (*QueryParams, error) {

	var queryParams QueryParams
	if isFilter {
		// Получаем фильтр
		body := c.Body()
		filterParam := c.QueryParam(filterParamName)
		if len(filterParam) > 0 {
			body = []byte(filterParam)
		}
		// Получаем фильтр
		filter, err := NewFilter(body)
		if err != nil {
			return nil, err
		}

		// Применение фильтра для запроса
		queryParams.Filter = &filter
	}

	paramId := c.Params(r.FieldIdName)
	if len(paramId) > 0 {
		queryParams.ID = QueryFormat(r.FieldIdName, paramId)
		//queryParams.Set(r.FieldIdName, )
	}
	c.QueryParams(func(key, value string) {
		if key == filterParamName {
			return
		}
		q := QueryFormat(key, value)
		queryParams.Set(q.Key, q)
	})

	return &queryParams, nil
}

// CustomHandler Установка обработчика маршрута
func (r *CRUD) CustomHandler(c *ewa.Context, h func(c *ewa.Context, r CRUD) error) error {
	return h(c, *r)
}

// ReadHandler Обработчик получения записей
func (r *CRUD) ReadHandler(c *ewa.Context, before, after Handler) error {

	r.SetModelName(r.TableTypes.Get(c.Get(HeaderTableType)))

	identity := NewIdentity(c.Identity)
	defer r.Insert(Read, r.ModelName, identity, c.Path())

	tableInfo := c.Get(HeaderTableInfo)
	if len(tableInfo) > 0 {
		var fields []string
		if tableInfo != "full" {
			fields = strings.Split(tableInfo, ",")
		}
		return c.JSON(r.JSON(200, r.Columns(r.ModelName, fields...)))
	}

	queryParams, err := r.NewQueryParams(c, true)
	if err != nil {
		return c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	// Обработчик до обращения в бд
	if before != nil {
		if status, err := before(c, *r, identity, queryParams, nil); err != nil {
			return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	// Если есть id возвращаем только одну запись
	if queryParams != nil && queryParams.ID != nil {
		record, err := r.GetRecord(r.ModelName, queryParams)
		if err != nil {
			return c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
		}
		record.Excludes(r.Excludes...)

		// Обработчик после обращению в бд
		if after != nil {
			if status, err := after(c, *r, identity, queryParams, nil); err != nil {
				return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
			}
		}

		return c.JSON(r.JSON(200, record))
	}

	// Вернуть записи
	records, total, err := r.GetRecords(r.ModelName, queryParams)
	if err != nil {
		return c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
	}
	// Заголовок Total
	c.Set(HeaderTotal, fmt.Sprintf("%d", total))
	records.Excludes(r.Excludes...)

	// Обработчик после обращению в бд
	if after != nil {
		if status, err := after(c, *r, identity, queryParams, nil); err != nil {
			return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	return c.JSON(r.JSON(200, records))
}

// CreateHandler Обработчик для создания записей
func (r *CRUD) CreateHandler(c *ewa.Context, before, after Handler) error {

	r.SetModelName(r.TableTypes.Get(c.Get(HeaderTableType)))

	identity := NewIdentity(c.Identity)
	defer r.Insert(Created, r.ModelName, identity, c.Path())

	body := NewBody(r.FieldIdName, NewFields("author", identity.Author)...)
	if err := body.Unmarshal(c.Body(), c.Get(HeaderXContentType) == "array"); err != nil {
		return c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	queryParams, err := r.NewQueryParams(c, false)
	if err != nil {
		return c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	if body.IsArray {
		var resp []any
		for i := range body.Array {

			// Обработчик до обращения в бд
			if before != nil {
				if status, err := before(c, *r, identity, queryParams, body); err != nil {
					return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
				}
			}

			id, err := r.SetRecord(r.ModelName, body.ToArrayMap(i), nil)
			if err != nil {
				resp = append(resp, r.Created(id, err))
				continue
			}
			resp = append(resp, r.Created(id, nil))
		}

		// Обработчик после обращению в бд
		if after != nil {
			if status, err := after(c, *r, identity, queryParams, body); err != nil {
				return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
			}
		}

		return c.JSON(200, resp)
	}

	// Обработчик до обращения в бд
	if before != nil {
		if status, err := before(c, *r, identity, queryParams, body); err != nil {
			return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	id, err := r.SetRecord(r.ModelName, body.ToMap(), nil)
	if err != nil {
		return c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
	}

	// Обработчик после обращению в бд
	if after != nil {
		if status, err := after(c, *r, identity, queryParams, body); err != nil {
			return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	return c.JSON(200, r.Created(id, nil))
}

// UpdateHandler Обновление записей
func (r *CRUD) UpdateHandler(c *ewa.Context, before, after Handler) error {

	r.SetModelName(r.TableTypes.Get(c.Get(HeaderTableType)))

	identity := NewIdentity(c.Identity)
	defer r.Insert(Updated, r.ModelName, identity, c.Path())

	// Получаем аргументы адресной строки
	queryParams, err := r.NewQueryParams(c, false)
	if err != nil {
		return c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	if queryParams != nil && queryParams.ID == nil && queryParams.Len() == 0 {
		return c.SendString(r.String(consts.StatusBadRequest, "Укажите поля для уточнения изменения записи! Пример: ../path?name=Name"))
	}

	body := NewBody(r.FieldIdName, NewFields("author", identity.Author)...)
	if err := body.Unmarshal(c.Body(), c.Get(HeaderXContentType) == "array"); err != nil {
		return c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	// Обработчик до обращения в бд
	if before != nil {
		if status, err := before(c, *r, identity, queryParams, body); err != nil {
			return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	// Пишем данные в бд
	if err := r.UpdateRecord(r.ModelName, body.ToMap(), queryParams); err != nil {
		return c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
	}

	// Обработчик после обращению в бд
	if after != nil {
		if status, err := after(c, *r, identity, queryParams, body); err != nil {
			return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	return c.JSON(r.JSON(200, r.Updated(body.GetField(r.FieldIdName), nil)))
}

// DeleteHandler Обработчик удаления записей
func (r *CRUD) DeleteHandler(c *ewa.Context, before, after Handler) (err error) {

	r.SetModelName(r.TableTypes.Get(c.Get(HeaderTableType)))

	identity := NewIdentity(c.Identity)
	defer r.Insert(Deleted, r.ModelName, identity, c.Path())

	// Получаем аргументы адресной строки
	queryParams, err := r.NewQueryParams(c, false)
	if err != nil {
		return c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	if queryParams != nil && queryParams.ID == nil && queryParams.Len() == 0 {
		return c.SendString(r.String(consts.StatusBadRequest, "Укажите поля для уточнения удаления записи! Пример: ../path?name=Name"))
	}

	// Обработчик до обращения в бд
	if before != nil {
		if status, err := before(c, *r, identity, queryParams, nil); err != nil {
			return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	// Удаление записи
	if err = r.DeleteRecord(r.ModelName, queryParams); err != nil {
		return c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
	}

	// Обработчик после обращению в бд
	if after != nil {
		if status, err := after(c, *r, identity, queryParams, nil); err != nil {
			return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	var id string
	if queryParams != nil && queryParams.ID != nil {
		id = queryParams.ID.Value
	}

	return c.JSON(r.JSON(200, r.Deleted(id, nil)))
}
