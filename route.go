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
	Variables   map[string]any

	IHandlers
	IResponse
	IQueryParam
}

var ErrQueryParam = "Укажите поля для уточнения изменения записи! Пример: ../path?name=Name"

type Handler func(*ewa.Context, *CRUD, *security.Identity, *QueryParams, *Body) (int, error)

func New(h IHandlers) *CRUD {
	return &CRUD{
		IHandlers:   h,
		IResponse:   new(Response),
		IQueryParam: new(PostgresFormat),
	}
}

// SetIQueryParam Установка интерфейса для форматирования query параметров
func (r *CRUD) SetIQueryParam(queryParam IQueryParam) *CRUD {
	r.IQueryParam = queryParam
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
		qf, err := r.QueryFormat(r.FieldIdName, paramId)
		if err != nil {
			return nil, err
		}
		queryParams.ID = qf
	}
	c.QueryParams(func(key, value string) {
		if key == filterParamName {
			return
		}
		qf, err := r.QueryFormat(key, value)
		if err != nil {
			return
		}
		queryParams.Set(qf.Key, qf)
	})

	return &queryParams, nil
}

// CustomHandler Установка обработчика маршрута
func (r *CRUD) CustomHandler(c *ewa.Context, h func(c *ewa.Context, r *CRUD) error) error {
	return h(c, r)
}

// ReadHandler Обработчик получения записей
func (r *CRUD) ReadHandler(c *ewa.Context, before, after Handler) error {

	if r.TableTypes != nil {
		r.SetModelName(r.TableTypes.Get(c.Get(HeaderTableType)))
	}
	// Аудит
	defer r.Audit(Read, c, r)

	// Вернуть столбцы таблицы
	tableInfo := strings.ToLower(c.Get(HeaderTableInfo))
	if len(tableInfo) > 0 {
		var fields []string
		if tableInfo != "full" {
			fields = strings.Split(tableInfo, ",")
		}
		return r.Send(c, Read, 200, r.Columns(r, fields...))
	}

	queryParams, err := r.NewQueryParams(c, true)
	if err != nil {
		return r.Send(c, Read, consts.StatusBadRequest, err) //c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	// Обработчик до обращения в бд
	if before != nil {
		if status, err := before(c, r, c.Identity, queryParams, nil); err != nil {
			return r.Send(c, Read, status, err)
		}
	}

	// Если есть id возвращаем только одну запись
	if queryParams != nil && queryParams.ID != nil {
		record, err := r.GetRecord(r, queryParams)
		if err != nil {
			return r.Send(c, Read, consts.StatusUnprocessableEntity, err) //c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
		}
		record.Excludes(r.Excludes...)

		// Обработчик после обращению в бд
		if after != nil {
			if status, err := after(c, r, c.Identity, queryParams, nil); err != nil {
				return r.Send(c, Read, status, err)
			}
		}

		return r.Send(c, Read, 200, record)
	}

	// Вернуть записи
	records, total, err := r.GetRecords(r, queryParams)
	if err != nil {
		return r.Send(c, Read, consts.StatusUnprocessableEntity, err) //c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
	}
	// Заголовок Total
	c.Set(HeaderTotal, fmt.Sprintf("%d", total))
	records.Excludes(r.Excludes...)

	// Обработчик после обращению в бд
	if after != nil {
		if status, err := after(c, r, c.Identity, queryParams, nil); err != nil {
			return r.Send(c, Read, status, err) //c.SendString(r.String(r.StatusDict.Get(status, err.Error())))
		}
	}

	return r.Send(c, Read, 200, records)
}

// CreateHandler Обработчик для создания записей
func (r *CRUD) CreateHandler(c *ewa.Context, before, after Handler) error {

	if r.TableTypes != nil {
		r.SetModelName(r.TableTypes.Get(c.Get(HeaderTableType)))
	}
	// Аудит
	defer r.Audit(Created, c, r)

	body := NewBody(r.FieldIdName)
	if c.Identity != nil {
		body.SetField("author", c.Identity.Username)
	}
	if err := body.Unmarshal(c.Body(), c.Get(HeaderXContentType) == "array"); err != nil {
		return r.Send(c, Created, consts.StatusBadRequest, err)
	}

	queryParams, err := r.NewQueryParams(c, false)
	if err != nil {
		return r.Send(c, Created, consts.StatusBadRequest, err) //c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	// Обработчик до обращения в бд
	if before != nil {
		if status, err := before(c, r, c.Identity, queryParams, body); err != nil {
			return r.Send(c, Created, status, err)
		}
	}

	result, err := r.SetRecord(r, body, queryParams)
	if err != nil {
		return r.Send(c, Created, consts.StatusUnprocessableEntity, err) //c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
	}

	// Обработчик после обращению в бд
	if after != nil {
		if status, err := after(c, r, c.Identity, queryParams, body); err != nil {
			return r.Send(c, Created, status, err)
		}
	}

	return r.Send(c, Created, 200, result)
}

// UpdateHandler Обновление записей
func (r *CRUD) UpdateHandler(c *ewa.Context, before, after Handler) error {

	if r.TableTypes != nil {
		r.SetModelName(r.TableTypes.Get(c.Get(HeaderTableType)))
	}
	// Аудит
	defer r.Audit(Updated, c, r)

	// Получаем аргументы адресной строки
	queryParams, err := r.NewQueryParams(c, false)
	if err != nil {
		return r.Send(c, Updated, consts.StatusBadRequest, err) //c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	if queryParams != nil && queryParams.ID == nil && queryParams.Len() == 0 {
		return r.Send(c, Updated, consts.StatusBadRequest, ErrQueryParam)
	}

	body := NewBody(r.FieldIdName)
	if c.Identity != nil {
		body.SetField("author", c.Identity.Username)
	}
	if err = body.Unmarshal(c.Body(), c.Get(HeaderXContentType) == "array"); err != nil {
		return r.Send(c, Updated, consts.StatusBadRequest, err)
	}

	// Обработчик до обращения в бд
	if before != nil {
		if status, err := before(c, r, c.Identity, queryParams, body); err != nil {
			return r.Send(c, Updated, status, err)
		}
	}

	// Пишем данные в бд
	result, err := r.UpdateRecord(r, body, queryParams)
	if err != nil {
		return r.Send(c, Updated, consts.StatusUnprocessableEntity, err)
		//return c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
	}

	// Обработчик после обращению в бд
	if after != nil {
		if status, err := after(c, r, c.Identity, queryParams, body); err != nil {
			return r.Send(c, Updated, status, err)
		}
	}

	return r.Send(c, Updated, 200, result)
}

// DeleteHandler Обработчик удаления записей
func (r *CRUD) DeleteHandler(c *ewa.Context, before, after Handler) (err error) {

	if r.TableTypes != nil {
		r.SetModelName(r.TableTypes.Get(c.Get(HeaderTableType)))
	}
	// Аудит
	defer r.Audit(Deleted, c, r)

	// Получаем аргументы адресной строки
	queryParams, err := r.NewQueryParams(c, false)
	if err != nil {
		return r.Send(c, Deleted, consts.StatusBadRequest, err)
		//return c.SendString(r.String(consts.StatusBadRequest, err.Error()))
	}

	if queryParams != nil && queryParams.ID == nil && queryParams.Len() == 0 {
		return r.Send(c, Deleted, consts.StatusBadRequest, ErrQueryParam)
	}

	// Обработчик до обращения в бд
	if before != nil {
		if status, err := before(c, r, c.Identity, queryParams, nil); err != nil {
			return r.Send(c, Deleted, status, err)
		}
	}

	// Удаление записи
	result, err := r.DeleteRecord(r, queryParams)
	if err != nil {
		return r.Send(c, Deleted, consts.StatusUnprocessableEntity, err)
		//return c.SendString(r.String(consts.StatusUnprocessableEntity, err.Error()))
	}

	// Обработчик после обращению в бд
	if after != nil {
		if status, err := after(c, r, c.Identity, queryParams, nil); err != nil {
			return r.Send(c, Deleted, status, err)
		}
	}

	return r.Send(c, Deleted, 200, result)
}
