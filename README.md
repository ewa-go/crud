### Формат запросов адресной строки для методов GET и PUT:
|Знак|Значение|Описание|Пример|SQL|
|----|--------|--------|------|---|
|=|string<br/>date<br/>boolean<br/>int<br/>float|Знак равно. Указывайте любое значение в соответствии с форматом поля в таблице бд. Одинарные кавычки - обязательные символы при указании строковых значений и даты|`?name=Имя<br/>?count=10`|<nobr>`SELECT * FROM table1 WHERE name = 'Имя'`</nobr>|
|=|[like]string<br/>[is]string<br/>[is not]string|Знак равно. Указывайте любое значение в соответствии с форматом поля в таблице бд. Можно указать любой оператор PostgreSQL. Одинарные кавычки - обязательные символы при указании строковых значений и даты|`?is_connected=[is]true`|`SELECT * FROM table1 WHERE is_connected is true`|
|=|[value1,value2,...]|in - оператор SQL подобных языков|`?name=[Вася,Петя]`|`SELECT * FROM table1 WHERE name in('Вася','Петя')`|
|[!]=<br/>[<>]=|string<br/>date<br/>boolean<br/>int<br/>float|Отрицание. != (не равно <>). Одинарные кавычки - обязательные символы при указании строковых значений и даты|`?name[!]=Имя`<br/>`?count[!]=10`<br/>`?name[<>]=Имя`<br/>`?count[<>]=10`|`<nobr>SELECT * FROM table1 WHERE name != 'Имя'</nobr><br/>SELECT * FROM table1 WHERE count != 10`|
|[>]=<br/>[<]=<br/>[>-]=<br/>[<-]=|string<br/>date<br/>boolean<br/>int<br/>float|Знаки равенства. Одинарные кавычки - обязательные символы при указании строковых значений и даты|`?count[>]=10`<br/>`?count[<]=10`<br/>`?count[>-]=10`<br/>`?count[<-]=10`|`SELECT * FROM table1 WHERE count > 10`<br/>`SELECT * FROM table1 WHERE count < 10`<br/>`SELECT * FROM table1 WHERE count >= 10`<br/>`SELECT * FROM table1 WHERE count <= 10`|
|[~]=|posix|posix (регулярные выражения - ~). [Документация](https://postgrespro.ru/docs/postgrespro/9.5/functions-matching#functions-posix-regexp)|`?name[~]=Имя*`|`SELECT * FROM table1 WHERE name ~ 'Имя*'`|
|[~*]=|posix|posix (регулярные выражения - ~). [Документация](https://postgrespro.ru/docs/postgrespro/9.5/functions-matching#functions-posix-regexp). Не учитывать регистр|`?name[~*]=Имя*`|`SELECT * FROM table1 WHERE name ~ 'Имя*'`|
|[!~]=|posix|posix (регулярные выражения - ~). [Документация](https://postgrespro.ru/docs/postgrespro/9.5/functions-matching#functions-posix-regexp)|`?name[!~]=Имя*`|`SELECT * FROM table1 WHERE name !~ 'Имя*'`|
|[!~*]=|posix|posix (регулярные выражения - ~). [Документация](https://postgrespro.ru/docs/postgrespro/9.5/functions-matching#functions-posix-regexp). Не учитывать регистр|`?name[!~*]=Имя*`|`SELECT * FROM table1 WHERE name !~* 'Имя*'`|
|[+]=|regex|similar to (регулярные выражения). [Документация](https://postgrespro.ru/docs/postgrespro/9.5/functions-matching#functions-posix-regexp)|`?name[+]=^(и|м|я)`|`SELECT * FROM table1 WHERE name similar to '^(и|м|я)'`|
|[!+]=|regex|similar to (регулярные выражения). [Документация](https://postgrespro.ru/docs/postgrespro/9.5/functions-matching#functions-posix-regexp)|`?name[!+]=^(и|м|я)`|`SELECT * FROM table1 WHERE name not similar to '^(и|м|я)'`|
|[%]=|%25string%25|like - оператор SQL подобных языков. [URL encoding - Wikipedia](https://en.wikipedia.org/wiki/Percent-encoding#Percent-encoding_reserved_characters)|`?name[%]=Им%25`|`SELECT * FROM table1 WHERE name like'Им%'`|
|[!%]=|%25string%25|like - оператор SQL подобных языков. [URL encoding - Wikipedia](https://en.wikipedia.org/wiki/Percent-encoding#Percent-encoding_reserved_characters)|`?name[!%]=Им%25`|`SELECT * FROM table1 WHERE name not like'Им%'`|
|[:]=|[date:date]|between - оператор SQL подобных языков|`?created[:]=[2022-05-01T17:00:00.000:2022-05-27T17:00:00.000]`|`SELECT * FROM table1 WHERE created BETWEEN '2022-05-01T17:00:00.000' AND '2022-05-27T17:00:00.000'`|
|=|null|null - если вам необходимо проверить значение на null. Запрос в бд будет выглядеть вот так: is null|`?name=null`|`SELECT * FROM table1 WHERE name is null`|
|[!]=<br/>[<>]=|null|null - если вам необходимо проверить значение на null. Запрос в бд будет выглядеть вот так: is not null|`?name[!]=null`|`SELECT * FROM table1 WHERE name is not null`|
|*||Имя поля, поиск по всем полям таблицы. Примечание: если указаны поля в массиве fields, то поиск происходит только по ним.|`?*[%]=49%25`|`SELECT * FROM table1 WHERE col1='49%' or col2='49%' or ...`|

##### Примечание: Для того чтобы увидеть поля таблицы и их описание, необходимо указать в запросе GET заголовок(header) - Table-Info:full, либо через запятую укажите имена полей по отдельности - Table-Info:column_name,...

#### Поиск по всем полям - *. Пример: ```url?*[%]=49%```

### Фильтр для запросов GET
Если вам потребуется указать фильтр запроса, например ```ORDER BY```, ```LIMIT``` и прочее, то вам нужно указать необходимые поля в теле запроса в формате json.

|Поле|Тип|Пример|Описание|
|----|---|------|--------|
|fields|array string|["fields1","fields2"]|Указываются  имена полей таблицы бд|
|orders|array string|["fields1","fields2 desc"]|Указываются  имена полей таблицы бд, а также оператор сортировки (ASC, DESC)|
|limit|integer|10|Число возвращаемых строк. Полезно для пагинации|
|offset|integer|5|Число с которой следует начинать отсчет строк в запросе. Полезно для пагинации|

Пример в формате json:
```json
{
    "fields": [
        "agent_id",
        "display_name"
    ],
    "orders": [
        "agent_id desc"
    ],
    "limit": 10,
    "offset": 0
}
```

#### Примечание. Модули `java script` для работы с `http` запрещают отправлять, при методе `GET`, тело запроса, чтобы это обойти укажите запрос в параметрах адресной строки в виде:<br/>`?~={"fields":["id","hostname","description"],"orders": ["id"],"offset": 0,"limit": 30}`
