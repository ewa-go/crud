package crud

import (
	"fmt"
	"github.com/egovorukhin/ewa-crud/utils"
	"strconv"
	"strings"
)

package api

import (
"fmt"
"github.com/ewa-go/ewa"
"github.com/ewa-go/ewa/consts"
"github.com/ewa-go/ewa/security"
"strconv"
"strings"
model "tbi/models"
"tbi/models/settings"
"tbi/src/db"
"tbi/src/utils"
)

type Excludes []string

func (ex Excludes) Is(v string) bool {
	for _, e := range ex {
		if e == v {
			return true
		}
	}
	return false
}

const (
	filterName = "~"

	HeaderXContentType = "X-Content-Type"
)

var ErrDatabase = "Ошибка данных"

type Handler func(id interface{}, data utils.Body)

func ReadHandler(c *ewa.Context, fieldIdName, tableName string, excludes Excludes) error {

	author, _ := Author(c.Identity)
	audit := settings.NewAuditRead(tableName, author, c.Path())
	defer audit.Insert()

	tableInfo := c.Get("Table-Info")
	if len(tableInfo) > 0 {
		var fields []string
		if tableInfo != "full" {
			fields = strings.Split(tableInfo, ",")
		}
		columns, err := db.DB(tableName).Columns(fields...)
		if err != nil {
			return c.SendString(audit.String(consts.StatusBadRequest, err.Error()))
		}
		return c.JSON(audit.JSON(200, columns))
	}

	// Получаем фильтр
	filter, err := GetFilter(c)
	if err != nil {
		return c.SendString(audit.String(consts.StatusBadRequest, err.Error()))
	}

	// Если есть id возвращаем только одну запись
	id := c.Params(fieldIdName)
	if len(id) > 0 {
		record, err := filter.GetRecord(tableName, fmt.Sprintf("%s=?", fieldIdName), id)
		if err != nil {
			return c.SendString(audit.String(consts.StatusUnprocessableEntity, ErrDatabase))
		}
		record.Excludes(excludes...)
		return c.JSON(audit.JSON(200, record))
	}

	// Формирование запроса
	query := GetQuery(c, tableName, filter.Fields)

	// Вернуть записи
	records, total, err := filter.GetRecords(tableName, query)
	if err != nil {
		return c.SendString(audit.String(consts.StatusUnprocessableEntity, ErrDatabase))
	}
	// Заголовок Total
	c.Set("Total", fmt.Sprintf("%d", total))
	records.Excludes(excludes...)
	return c.JSON(audit.JSON(200, records))
}

// CreateHandler Обработчик для создания записей
func CreateHandler(c *ewa.Context, tableName, fieldIdName string, isMapTable ...bool) error {

	author, _ := Author(c.Identity)
	audit := settings.NewAuditCreate(tableName, author, c.Path())
	defer audit.Insert()

	body := utils.NewBody(fieldIdName, author)
	if err := body.Unmarshal(c.Body(), c.Get(HeaderXContentType) == "array"); err != nil {
		return c.SendString(audit.String(consts.StatusBadRequest, err.Error()))
	}
	if body.IsArray {
		var resp []model.CrudResponse
		for _, data := range body.Array {
			if err := Create(tableName, data); err != nil {
				resp = append(resp, model.NewCreated(body.Id(), err.Error(), false))
				continue
			}
			resp = append(resp, model.NewCreated(body.Id(), "OK", true))
		}
		return c.JSON(200, resp)
	}
	if err := Create(tableName, body.Data); err != nil {
		return c.SendString(audit.String(consts.StatusUnprocessableEntity, err.Error()))
	}
	return c.JSON(200, model.NewCreated(body.Id(), "OK", true))
}

func Create(tableName string, data interface{}) (err error) {
	if err = db.DB(tableName).Create(data); err != nil {
		if db.IsErrDuplicateKey(err) {
			return fmt.Errorf("запись уже существует")
		}
	}
	return err
}

func UpdateHandler(c *ewa.Context, fieldIdName, tableName string, data *utils.Body, isMapTable ...bool) error {

	author, _ := Author(c.Identity)
	audit := settings.NewAuditUpdate(tableName, author, c.Path())
	defer audit.Insert()

	// Получаем аргументы адресной строки
	var (
		values  []string
		idParam = c.Params(fieldIdName)
	)
	if len(idParam) > 0 {
		id, err := strconv.Atoi(idParam)
		if err != nil {
			return c.SendString(audit.String(consts.StatusBadRequest, "Параметр адресной строки должен быть числом. Пример: ../path/1"))
		}
		values = append(values, fmt.Sprintf("%s=%d", fieldIdName, id))
	}
	c.QueryParams(func(key, value string) {
		values = append(values, utils.QueryFormat(key, value).String())
	})

	if len(values) == 0 {
		return c.SendString(audit.String(consts.StatusBadRequest, "Укажите поля для уточнения изменения записи! Пример: ../path?name=Name"))
	}

	body := utils.NewBody(fieldIdName, author)
	if err := body.Unmarshal(c.Body(), c.Get(HeaderXContentType) == "array"); err != nil {
		return c.SendString(audit.String(consts.StatusBadRequest, err.Error()))
	}
	for key, value := range data.ToMap() {
		body.Set(key, value)
	}
	if len(isMapTable) > 0 && !isMapTable[0] {
		body = body.Datetime()
	} else {
		body = body.Modified()
	}

	// Пишем данные в бд
	if err := db.DB(tableName).Update(body.SetAuthor(author).ToMap(), strings.Join(values, " and ")); err != nil {
		return c.SendString(audit.String(consts.StatusUnprocessableEntity, ErrDatabase))
	}

	// Запуск обработчиков
	/*for _, handler := range handlers {
		go handler(body.Id(), body)
	}*/

	return c.JSON(audit.JSON(200, model.NewUpdated(body.Id(), "OK", true)))
}

func DeleteHandler(c *ewa.Context, fieldIdName, tableName string) (err error) {

	author, _ := Author(c.Identity)
	audit := settings.NewAuditDelete(tableName, author, c.Path())
	defer audit.Insert()

	var (
		values  []string
		paramId = c.Params(fieldIdName)
		id      int
	)
	if len(paramId) > 0 {
		// Если есть id возвращаем только одну запись
		id, err = strconv.Atoi(c.Params(fieldIdName))
		if err != nil {
			return c.SendString(audit.String(consts.StatusBadRequest, "Параметр адресной строки должен быть числом. Пример: ../path/1"))
		}
		values = append(values, fmt.Sprintf("%s=%d", fieldIdName, id))
	}

	// Получаем аргументы адресной строки
	c.QueryParams(func(key, value string) {
		values = append(values, utils.QueryFormat(key, value).String())
	})

	if len(values) == 0 {
		return c.SendString(audit.String(consts.StatusBadRequest, "Обязательно укажите условие для удаления записей! Пример: ?name=Name"))
	}

	// Удаление записи
	if err = db.DB(tableName).Delete(nil, strings.Join(values, " and ")); err != nil {
		return c.SendString(audit.String(consts.StatusUnprocessableEntity, ErrDatabase))
	}

	// Запуск обработчиков
	/*for _, handler := range handlers {
		go handler(id, nil)
	}*/

	return c.JSON(audit.JSON(200, model.NewDeleted(id, "OK", true)))
}

func Author(identity *security.Identity) (author string, isAdmin bool) {
	if identity != nil {
		author = identity.Username
		if admin := identity.GetVariable("admin"); admin != nil {
			isAdmin = admin.(bool)
		}
	}
	return
}

func Get(route *ewa.Route, fieldIdName string, h utils.HeaderValues, handler ewa.Handler, excludes ...string) {
	modelName := ""
	if h != nil {
		modelName = h.Default()
	}
	route.SetSecurity(security.ApiKeyAuth, security.BasicAuth).
		SetSummary("Get data").
		SetParameters(DefaultParameters(fieldIdName)...).
		InitParametersByModel(modelName).
		SetResponse(consts.StatusOK, modelName, nil, "OK").
		SetEmptyParam("Get data").SetResponseArray(200, modelName, nil, "Return array data")
	Response(route)
	if handler != nil {
		route.Handler = handler
	} else {
		route.Handler = func(c *ewa.Context) error {
			return ReadHandler(c, fieldIdName, utils.GetValueFromHeader(c, "Table-Type", h), excludes)
		}
	}
}

func Post(route *ewa.Route, modelName string, handler ewa.Handler, isMapTable ...bool) {
	route.SetSecurity(security.ApiKeyAuth, security.BasicAuth).
		SetSummary("Create data").
		SetParameters(ewa.NewBodyParam(true, modelName, false, "Must have request body")).
		SetResponse(consts.StatusOK, model.ModelCrudResponse, nil, "OK").
		SetConsumes(consts.MIMEApplicationJSON)
	Response(route)
	if handler != nil {
		route.Handler = handler
	} else {
		route.Handler = func(c *ewa.Context) error {
			return CreateHandler(c, modelName, "id", isMapTable...)
		}
	}
}

func Put(route *ewa.Route, fieldIdName, modelName string, handler ewa.Handler, isMapTable ...bool) {
	route.SetSecurity(security.ApiKeyAuth, security.BasicAuth).
		SetSummary("Update data").
		SetParameters(
			ewa.NewPathParam(fmt.Sprintf("/{%s}", fieldIdName), "ID data"),
			ewa.NewBodyParam(true, modelName, false, "Must have request body"),
		).
		InitParametersByModel(modelName).
		SetResponse(consts.StatusOK, model.ModelCrudResponse, nil, "OK").
		SetConsumes(consts.MIMEApplicationJSON).
		SetEmptyParam("Update data").SetResponse(200, model.ModelCrudResponse, nil, "Response update data")
	Response(route)
	if handler != nil {
		route.Handler = handler
	} else {
		route.Handler = func(c *ewa.Context) error {
			return UpdateHandler(c, fieldIdName, modelName, nil, isMapTable...)
		}
	}
}

func Delete(route *ewa.Route, fieldIdName, modelName string, handler ewa.Handler) {
	route.SetSecurity(security.ApiKeyAuth, security.BasicAuth).
		SetSummary("Delete data").
		SetParameters(ewa.NewPathParam("/{id}", "ID data")).
		InitParametersByModel(modelName).
		SetResponse(consts.StatusOK, model.ModelCrudResponse, nil, "OK").
		SetEmptyParam("Delete data").SetResponse(200, model.ModelCrudResponse, nil, "Response delete data")
	Response(route)
	if handler != nil {
		route.Handler = handler
	} else {
		route.Handler = func(c *ewa.Context) error {
			return DeleteHandler(c, fieldIdName, modelName)
		}
	}
}

func Response(route *ewa.Route) {
	route.SetResponse(consts.StatusBadRequest, "", nil, "Return parse parameter/body error").
		SetResponse(consts.StatusForbidden, "", nil, "Not Permission").
		SetResponse(consts.StatusUnprocessableEntity, "", nil, "Database error")
}

func DefaultParameters(fieldIdName string) []*ewa.Parameter {
	return []*ewa.Parameter{
		ewa.NewPathParam(fmt.Sprintf("/{%s}", fieldIdName), "ID data"),
		ewa.NewQueryParam("~", false, `Filter. Example: ../path?~={"fields":[],"orders":[],limit:-1,offset:-1}`),
		ewa.NewHeaderParam("Table-Info", false, "Get information table. Default: full. Example: column_name,data_type"),
		ewa.NewHeaderParam("Table-Type", false, "Get data type: table, view. Default: table. Example: Table-Typ=view"),
	}
}

func GetFilter(c *ewa.Context) (model.Filter, error) {
	body := c.Body()
	filterParam := c.QueryParam(filterName)
	if len(filterParam) > 0 {
		body = []byte(filterParam)
	}
	// Получаем фильтр
	return model.NewFilter(body)
}

func GetQuery(c *ewa.Context, tableName string, fields []string) string {
	var (
		values      []string
		valueFields []string
		queryParams = make(map[string]utils.QueryParam)
		query       string
	)
	// Применение фильтра для запроса
	c.QueryParams(func(key, value string) {
		if key == filterName {
			return
		}
		q := utils.QueryFormat(key, value)
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
			columns, _ := db.DB(tableName).Columns(columnName)
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

func GetQuery1(c *ewa.Context, tableName string, fields []string) (string, []interface{}) {
	var (
		values      []string
		valueFields []string
		queryParams = make(map[string]utils.QueryParam)
		query       string
	)
	// Применение фильтра для запроса
	c.QueryParams(func(key, value string) {
		if key == filterName {
			return
		}
		q := utils.QueryFormat(key, value)
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
			columns, _ := db.DB(tableName).Columns(columnName)
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
	return query, nil
}
