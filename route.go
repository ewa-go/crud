package crud

import (
	"fmt"
	"github.com/ewa-go/ewa"
	"github.com/ewa-go/ewa/consts"
	"github.com/ewa-go/ewa/security"
	"strings"
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
}

var ErrQueryParam = "Укажите поля для уточнения изменения записи! Пример: ../path?name=Name"

type Handler func(*ewa.Context, CRUD, *security.Identity, *QueryParams, *Body) (int, error)

type StatusDict map[int]string

func (e StatusDict) Get(status int, def ...string) (s int, v any) {
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
func (r *CRUD) CustomHandler(c *ewa.Context, h func(c *ewa.Context, r *CRUD) error) error {
	return h(c, r)
}

// ReadHandler Обработчик получения записей
func (r *CRUD) ReadHandler(c *ewa.Context, before, after Handler) error {

	r.SetModelName(r.TableTypes.Get(c.Get(HeaderTableType)))
	// Аудит
	defer func() {
		if r.IAudit != nil {
			r.Exec(Read, c, r)
		}
	}()

	// Вернуть столбцы таблицы
	tableInfo := c.Get(HeaderTableInfo)
	if len(tableInfo) > 0 {
		var fields []string
		if tableInfo != "full" {
			fields = strings.Split(tableInfo, ",")
		}
		return r.Send(c, Read, 200, r.Columns(r.ModelName, fields...)) //c.JSON(r.JSON(200, r.Columns(r.ModelName, fields...)))
	}

	queryParams, err := r.NewQueryParams(c, true)
	if err != nil {
		return r.Send(c, Read, consts.StatusBadRequest, err) //c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	// Обработчик до обращения в бд
	if before != nil {
		if status, err := before(c, *r, c.Identity, queryParams, nil); err != nil {
			return r.Send(c, Read, status, err)
		}
	}

	// Если есть id возвращаем только одну запись
	if queryParams != nil && queryParams.ID != nil {
		record, err := r.GetRecord(r.ModelName, queryParams)
		if err != nil {
			return r.Send(c, Read, consts.StatusUnprocessableEntity, err) //c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
		}
		record.Excludes(r.Excludes...)

		// Обработчик после обращению в бд
		if after != nil {
			if status, err := after(c, *r, c.Identity, queryParams, nil); err != nil {
				return r.Send(c, Read, status, err) // c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
			}
		}

		return r.Send(c, Read, 200, record)
	}

	// Вернуть записи
	records, total, err := r.GetRecords(r.ModelName, queryParams)
	if err != nil {
		return r.Send(c, Read, consts.StatusUnprocessableEntity, err) //c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
	}
	// Заголовок Total
	c.Set(HeaderTotal, fmt.Sprintf("%d", total))
	records.Excludes(r.Excludes...)

	// Обработчик после обращению в бд
	if after != nil {
		if status, err := after(c, *r, c.Identity, queryParams, nil); err != nil {
			return r.Send(c, Read, status, err) //c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	return r.Send(c, Read, 200, records) //c.JSON(r.JSON(200, records))
}

// CreateHandler Обработчик для создания записей
func (r *CRUD) CreateHandler(c *ewa.Context, before, after Handler) error {

	r.SetModelName(r.TableTypes.Get(c.Get(HeaderTableType)))
	// Аудит
	if r.IAudit != nil {
		defer r.Exec(Created, c, r)
	}

	body := NewBody(r.FieldIdName)
	if c.Identity != nil {
		body.SetField("author", c.Identity.Username)
	}
	if err := body.Unmarshal(c.Body(), c.Get(HeaderXContentType) == "array"); err != nil {
		return r.Send(c, Created, consts.StatusBadRequest, err) //c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	queryParams, err := r.NewQueryParams(c, false)
	if err != nil {
		return r.Send(c, Created, consts.StatusBadRequest, err) //c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	/*if body.IsArray {
		var resp []any
		for i := range body.Array {

			// Обработчик до обращения в бд
			if before != nil {
				if status, err := before(c, *r, c.Identity, queryParams, body); err != nil {
					return r.Send(c, Created, status, err) //c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
				}
			}

			returning, err := r.SetRecord(r.ModelName, body.ToArrayMap(i), nil)
			if err != nil {
				resp = append(resp, r.Get(returning)) //.Created(returning, err))
				continue
			}
			resp = append(resp, r.Get(returning)) //r.Created(returning, nil))
		}

		// Обработчик после обращению в бд
		if after != nil {
			if status, err := after(c, *r, c.Identity, queryParams, body); err != nil {
				return r.Send(c, Created, status, err)
				//return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
			}
		}

		return c.JSON(200, resp)
	}*/

	// Обработчик до обращения в бд
	if before != nil {
		if status, err := before(c, *r, c.Identity, queryParams, body); err != nil {
			return r.Send(c, Created, status, err)
			//return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	result, err := r.SetRecord(r.ModelName, body, queryParams)
	if err != nil {
		return r.Send(c, Created, consts.StatusUnprocessableEntity, err) //c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
	}

	// Обработчик после обращению в бд
	if after != nil {
		if status, err := after(c, *r, c.Identity, queryParams, body); err != nil {
			return r.Send(c, Created, status, err) //c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	return r.Send(c, Created, 200, result) //c.JSON(200, r.Created(returning, nil))
}

// UpdateHandler Обновление записей
func (r *CRUD) UpdateHandler(c *ewa.Context, before, after Handler) error {

	r.SetModelName(r.TableTypes.Get(c.Get(HeaderTableType)))
	// Аудит
	if r.IAudit != nil {
		defer r.Exec(Updated, c, r)
	}

	// Получаем аргументы адресной строки
	queryParams, err := r.NewQueryParams(c, false)
	if err != nil {
		return r.Send(c, Updated, consts.StatusBadRequest, err) //c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	if queryParams != nil && queryParams.ID == nil && queryParams.Len() == 0 {
		return r.Send(c, Updated, consts.StatusBadRequest, ErrQueryParam) //c.SendString(r.String(consts.StatusBadRequest, "Укажите поля для уточнения изменения записи! Пример: ../path?name=Name"))
	}

	body := NewBody(r.FieldIdName)
	if c.Identity != nil {
		body.SetField("author", c.Identity.Username)
	}
	if err = body.Unmarshal(c.Body(), c.Get(HeaderXContentType) == "array"); err != nil {
		return r.Send(c, Updated, consts.StatusBadRequest, err) //c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	// Обработчик до обращения в бд
	if before != nil {
		if status, err := before(c, *r, c.Identity, queryParams, body); err != nil {
			return r.Send(c, Updated, status, err)
			//return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	// Пишем данные в бд
	result, err := r.UpdateRecord(r.ModelName, body, queryParams)
	if err != nil {
		return r.Send(c, Updated, consts.StatusUnprocessableEntity, err)
		//return c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
	}

	// Обработчик после обращению в бд
	if after != nil {
		if status, err := after(c, *r, c.Identity, queryParams, body); err != nil {
			return r.Send(c, Updated, status, err)
			//return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	return r.Send(c, Updated, 200, result) //c.JSON(r.JSON(200, r.Updated(body.GetField(r.FieldIdName), nil)))
}

// DeleteHandler Обработчик удаления записей
func (r *CRUD) DeleteHandler(c *ewa.Context, before, after Handler) (err error) {

	r.SetModelName(r.TableTypes.Get(c.Get(HeaderTableType)))
	// Аудит
	if r.IAudit != nil {
		defer r.Exec(Deleted, c, r)
	}

	// Получаем аргументы адресной строки
	queryParams, err := r.NewQueryParams(c, false)
	if err != nil {
		return r.Send(c, Deleted, consts.StatusBadRequest, err)
		//return c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	if queryParams != nil && queryParams.ID == nil && queryParams.Len() == 0 {
		return r.Send(c, Deleted, consts.StatusBadRequest, ErrQueryParam)
		//return c.SendString(r.String(consts.StatusBadRequest, "Укажите поля для уточнения удаления записи! Пример: ../path?name=Name"))
	}

	// Обработчик до обращения в бд
	if before != nil {
		if status, err := before(c, *r, c.Identity, queryParams, nil); err != nil {
			return r.Send(c, Deleted, status, err)
			//return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	// Удаление записи
	result, err := r.DeleteRecord(r.ModelName, queryParams)
	if err != nil {
		return r.Send(c, Deleted, consts.StatusUnprocessableEntity, err)
		//return c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
	}

	// Обработчик после обращению в бд
	if after != nil {
		if status, err := after(c, *r, c.Identity, queryParams, nil); err != nil {
			return r.Send(c, Deleted, status, err)
			//return c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	/*var id string
	if queryParams != nil && queryParams.ID != nil {
		id = queryParams.ID.Value
	}*/

	return r.Send(c, Deleted, 200, result) //c.JSON(r.JSON(200, r.Deleted(id, nil)))
}
