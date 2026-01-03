package main

import (
    "fmt"
    "strings"
)

// ========== Runtime Helpers ==========

// Exception 基础异常类型
type Exception struct {
	Message string
}

func (e *Exception) Init(message string) {
	e.Message = message
}

func (e *Exception) Getmessage() string {
	return e.Message
}

func (e *Exception) Error() string {
	return e.Message
}

// __parseInt 将字符串转换为整数
func __parseInt(s interface{}) int64 {
	switch v := s.(type) {
	case string:
		var result int64
		fmt.Sscanf(v, "%d", &result)
		return result
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}

// __parseFloat 将字符串转换为浮点数
func __parseFloat(s interface{}) float64 {
	switch v := s.(type) {
	case string:
		var result float64
		fmt.Sscanf(v, "%f", &result)
		return result
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	default:
		return 0
	}
}

// __typeof 获取值的类型
func __typeof(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch v.(type) {
	case string:
		return "STRING"
	case int, int64, int32:
		return "INT"
	case float64, float32:
		return "FLOAT"
	case bool:
		return "BOOL"
	case []interface{}:
		return "ARRAY"
	case map[string]interface{}:
		return "MAP"
	default:
		return "OBJECT"
	}
}

// __isset 检查 map 中是否存在键
func __isset(m interface{}, key interface{}) bool {
	if m == nil {
		return false
	}
	switch mv := m.(type) {
	case map[string]interface{}:
		keyStr := fmt.Sprint(key)
		_, ok := mv[keyStr]
		return ok
	default:
		return false
	}
}

// __len 获取集合的长度
func __len(v interface{}) int {
	if v == nil {
		return 0
	}
	switch vv := v.(type) {
	case string:
		return len(vv)
	case []interface{}:
		return len(vv)
	case map[string]interface{}:
		return len(vv)
	default:
		return 0
	}
}

// __createInstance 运行时创建实例
func __createInstance(className string) interface{} {
	// TODO: 实现运行时实例创建
	return nil
}

// __getMap 安全获取 map[string]interface{} 值
func __getMap(m interface{}, key interface{}) interface{} {
	if m == nil {
		return nil
	}
	keyStr := fmt.Sprint(key)
	if mv, ok := m.(map[string]interface{}); ok {
		return mv[keyStr]
	}
	return nil
}

// __setMap 安全设置 map[string]interface{} 值
func __setMap(m interface{}, key interface{}, value interface{}) {
	keyStr := fmt.Sprint(key)
	if mv, ok := m.(map[string]interface{}); ok {
		mv[keyStr] = value
	}
}

// __getIndex 安全获取数组索引
func __getIndex(arr interface{}, index int) interface{} {
	if arr == nil {
		return nil
	}
	if av, ok := arr.([]interface{}); ok {
		if index >= 0 && index < len(av) {
			return av[index]
		}
	}
	return nil
}

// __getIndexStr 安全获取数组索引并转换为字符串
func __getIndexStr(arr interface{}, index int) string {
	v := __getIndex(arr, index)
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

// ========== Reflection Helpers ==========

// ReflectionGetclassname 获取实例的类名
func ReflectionGetclassname(obj interface{}) string {
	if obj == nil {
		return ""
	}
	t := fmt.Sprintf("%T", obj)
	// 移除指针前缀 *
	if len(t) > 0 && t[0] == '*' {
		t = t[1:]
	}
	// 移除包名前缀 main.
	if idx := len("main."); len(t) > idx && t[:idx] == "main." {
		t = t[idx:]
	}
	return t
}

// ReflectionGetfieldvalue 获取字段值 (占位实现)
func ReflectionGetfieldvalue(obj interface{}, fieldName interface{}) interface{} {
	// TODO: 使用反射实现
	return nil
}

// ReflectionSetfieldvalue 设置字段值 (占位实现)
func ReflectionSetfieldvalue(obj interface{}, fieldName interface{}, value interface{}) {
	// TODO: 使用反射实现
}

// ReflectionGetclassannotation 获取类注解 (占位实现)
func ReflectionGetclassannotation(className string, annotationName string) interface{} {
	// TODO: 实现注解读取
	return nil
}

// ReflectionGetclassfields 获取类字段 (占位实现)
func ReflectionGetclassfields(className string) map[string]interface{} {
	// TODO: 实现字段读取
	return map[string]interface{}{}
}

// ReflectionHasfieldannotation 检查字段是否有注解 (占位实现)
func ReflectionHasfieldannotation(className string, fieldName string, annotationName string) bool {
	// TODO: 实现注解检查
	return false
}

// ReflectionGetfieldannotation 获取字段注解 (占位实现)
func ReflectionGetfieldannotation(className string, fieldName string, annotationName string) interface{} {
	// TODO: 实现注解读取
	return nil
}

// __toMap 安全转换为 map[string]interface{}
func __toMap(v interface{}) map[string]interface{} {
	if v == nil {
		return map[string]interface{}{}
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

// __toSlice 安全转换为 []interface{}
func __toSlice(v interface{}) []interface{} {
	if v == nil {
		return []interface{}{}
	}
	if s, ok := v.([]interface{}); ok {
		return s
	}
	return []interface{}{}
}

// __toString 安全转换为 string
func __toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// __toInt 安全转换为 int64
func __toInt(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case string:
		var result int64
		fmt.Sscanf(n, "%d", &result)
		return result
	default:
		return 0
	}
}

type Application struct {
}

func Main() {
fmt.Println("Hello from Application!")
    product := func() *Product { __t := &Product{}; __t.Init(1, "测试产品", 99.99); return __t }()
    fmt.Println(("Product: " + product.Getname()))
}

type User struct {
    Model
    Id int64
    Name string
    Email string
    Age int64
    Status string
    Password string
    Views int64
    Createdat string
}

func (u *User) Init() {
u.Model.Init()
    u.Status = "active"
}
func (u *User) Isadult() bool {
age := __parseInt(fmt.Sprint(u.Age))
    return (age >= 18)
}
func (u *User) Getdisplayname() string {
return (((u.Name + " <") + u.Email) + ">")
}

type Product struct {
    Id int64
    Name string
    Price float64
}

func (p *Product) Init(id int64, name string, price float64) {
p.Id = id
    p.Name = name
    p.Price = price
}
func (p *Product) Getid() int64 {
return p.Id
}
func (p *Product) Getname() string {
return p.Name
}
func (p *Product) Getprice() float64 {
return p.Price
}
func ProductSetprice() {
}

type Model struct {
    Exists bool
    Original interface{}
}

var ModelConnection interface{}

func (m *Model) Init() {
m.Exists = false
    m.Original = map[string]interface{}{}
}
func ModelSetconnection(conn interface{}) {
ModelConnection = conn
}
func ModelGetconnection() interface{} {
conn := ModelConnection
    if (conn == nil) {
    panic(func() *Databaseexception { __t := &Databaseexception{}; __t.Init("Model connection not set. Call Model::setConnection() first."); return __t }())
}
    return conn
}
func ModelQuery() interface{} {
classname := "Model"
    return func() *Modelbuilder { __t := &Modelbuilder{}; __t.Init(classname); return __t }()
}
func ModelCreate(attributes interface{}) interface{} {
classname := "Model"
    meta := ModelGetmeta(classname)
    instance := __createInstance(classname)
    fillable := __getMap(meta, "fillable")
    columns := __getMap(meta, "columns")
    for fieldname, _ := range __toMap(columns) {
    if __isset(attributes, fieldname) {
    if ModelIsfillable(fieldname, fillable) {
    ReflectionSetfieldvalue(instance, fieldname, __getMap(attributes, fieldname))
}
}
}
    instance.Save()
    return instance
}
func ModelIsfillable(fieldname string, fillable interface{}) bool {
if ((fillable == nil) || (__len(fillable) == 0)) {
    return true
}
    for i := 0; (i < __len(fillable)); i++ {
    if (__getIndex(fillable, i) == fieldname) {
    return true
}
}
    return false
}
func (m *Model) Save() bool {
classname := ReflectionGetclassname(m)
    meta := ModelGetmeta(classname)
    if m.Exists {
    return m.Performupdate(meta)
} else {
    return m.Performinsert(meta)
}
}
func (m *Model) Delete() bool {
if (!m.Exists) {
    return false
}
    classname := ReflectionGetclassname(m)
    meta := ModelGetmeta(classname)
    tablename := __getMap(meta, "table")
    primarykey := __getMap(meta, "primaryKey")
    pkvalue := ReflectionGetfieldvalue(m, primarykey)
    conn := ModelGetconnection()
    affected := conn.Table(tablename).Where(primarykey, pkvalue).Delete()
    if (affected > 0) {
    m.Exists = false
    return true
}
    return false
}
func (m *Model) Isdirty() bool {
return (__len(m.Getdirty()) > 0)
}
func (m *Model) Getdirty() interface{} {
classname := ReflectionGetclassname(m)
    meta := ModelGetmeta(classname)
    columns := __getMap(meta, "columns")
    dirty := map[string]interface{}{}
    for fieldname, _ := range __toMap(columns) {
    currentvalue := ReflectionGetfieldvalue(m, fieldname)
    var originalvalue interface{} = nil
    if __isset(m.Original, fieldname) {
    originalvalue = __toString(__getMap(m.Original, fieldname))
}
    if (fmt.Sprint(currentvalue) != fmt.Sprint(originalvalue)) {
    __setMap(dirty, fieldname, currentvalue)
}
}
    return dirty
}
func (m *Model) Toarray() interface{} {
classname := ReflectionGetclassname(m)
    meta := ModelGetmeta(classname)
    columns := __getMap(meta, "columns")
    hidden := __getMap(meta, "hidden")
    result := map[string]interface{}{}
    for fieldname, _ := range __toMap(columns) {
    ishidden := false
    for i := 0; (i < __len(hidden)); i++ {
    if (__getIndex(hidden, i) == fieldname) {
    ishidden = true
    break
}
}
    if ishidden {
    continue
}
    __setMap(result, fieldname, ReflectionGetfieldvalue(m, fieldname))
}
    return result
}
func (m *Model) Tojson() string {
data := m.Toarray()
    parts := []interface{}{}
    for key, value := range __toMap(data) {
    part := (("\"" + key) + "\": ")
    valuetype := __typeof(value)
    if (valuetype == "STRING") {
    part = (((part + "\"") + fmt.Sprint(value)) + "\"")
} else if (value == nil) {
    part = (part + "null")
} else {
    part = (part + fmt.Sprint(value))
}
    func() { parts = append(parts, part) }()
}
    result := "{"
    for i := 0; (i < __len(parts)); i++ {
    if (i > 0) {
    result = (result + ", ")
}
    result = (result + __getIndexStr(parts, i))
}
    return (result + "}")
}
func (m *Model) Performinsert(meta interface{}) bool {
tablename := __getMap(meta, "table")
    primarykey := __getMap(meta, "primaryKey")
    columns := __getMap(meta, "columns")
    data := map[string]interface{}{}
    for fieldname, colinfo := range __toMap(columns) {
    if (fieldname == primarykey) {
    continue
}
    value := ReflectionGetfieldvalue(m, fieldname)
    if (value != nil) {
    columnname := __getMap(colinfo, "column")
    __setMap(data, columnname, value)
}
}
    conn := ModelGetconnection()
    id := conn.Table(tablename).Insertgetid(data)
    if (id > 0) {
    ReflectionSetfieldvalue(m, primarykey, id)
    m.Exists = true
    m.Syncoriginal(meta)
    return true
}
    return false
}
func (m *Model) Performupdate(meta interface{}) bool {
dirty := m.Getdirty()
    if (__len(dirty) == 0) {
    return true
}
    tablename := __getMap(meta, "table")
    primarykey := __getMap(meta, "primaryKey")
    columns := __getMap(meta, "columns")
    pkvalue := ReflectionGetfieldvalue(m, primarykey)
    data := map[string]interface{}{}
    for fieldname, value := range __toMap(dirty) {
    if __isset(columns, fieldname) {
    colinfo := __getMap(columns, fieldname)
    columnname := __getMap(colinfo, "column")
    __setMap(data, columnname, value)
}
}
    conn := ModelGetconnection()
    affected := conn.Table(tablename).Where(primarykey, pkvalue).Update(data)
    if (affected >= 0) {
    m.Syncoriginal(meta)
    return true
}
    return false
}
func (m *Model) Syncoriginal(meta interface{}) {
columns := __getMap(meta, "columns")
    m.Original = map[string]interface{}{}
    for fieldname, _ := range __toMap(columns) {
    __setMap(m.Original, fieldname, ReflectionGetfieldvalue(m, fieldname))
}
}
func (m *Model) Hydrate(data interface{}, meta interface{}) {
columns := __getMap(meta, "columns")
    for fieldname, colinfo := range __toMap(columns) {
    columnname := __getMap(colinfo, "column")
    if __isset(data, columnname) {
    ReflectionSetfieldvalue(m, fieldname, __getMap(data, columnname))
}
}
    m.Exists = true
    m.Syncoriginal(meta)
}
func ModelGetmeta(classname string) interface{} {
tablename := (strings.ToLower(classname) + "s")
    tableannotation := ReflectionGetclassannotation(classname, "Table")
    if (tableannotation != nil) {
    if __isset(tableannotation, "name") {
    val := __getMap(tableannotation, "name")
    if (val != nil) {
    tablename = __toString(val)
}
}
}
    fields := ReflectionGetclassfields(classname)
    columns := map[string]interface{}{}
    primarykey := "id"
    hidden := []interface{}{}
    fillable := []interface{}{}
    createdat := ""
    updatedat := ""
    for fieldname, fieldinfo := range __toMap(fields) {
    if strings.HasPrefix(fieldname, "_") {
    continue
}
    if ReflectionHasfieldannotation(classname, fieldname, "Id") {
    primarykey = __toString(fieldname)
}
    columnname := fieldname
    colannotation := ReflectionGetfieldannotation(classname, fieldname, "Column")
    if (colannotation != nil) {
    if __isset(colannotation, "name") {
    colval := __getMap(colannotation, "name")
    if ((colval != nil) && (colval != "")) {
    columnname = __toString(colval)
}
}
}
    __setMap(columns, fieldname, map[string]interface{}{"column": columnname, "type": __getMap(fieldinfo, "type")})
    if ReflectionHasfieldannotation(classname, fieldname, "Hidden") {
    func() { hidden = append(hidden, fieldname) }()
}
    if ReflectionHasfieldannotation(classname, fieldname, "Fillable") {
    func() { fillable = append(fillable, fieldname) }()
}
    if ReflectionHasfieldannotation(classname, fieldname, "CreatedAt") {
    createdat = fieldname
}
    if ReflectionHasfieldannotation(classname, fieldname, "UpdatedAt") {
    updatedat = fieldname
}
}
    return map[string]interface{}{"table": tablename, "primaryKey": primarykey, "columns": columns, "hidden": hidden, "fillable": fillable, "createdAt": createdat, "updatedAt": updatedat}
}

type Databaseexception struct {
    Exception
}

func (d *Databaseexception) Init(message string) {
d.Exception.Init(message)
}
func (d *Databaseexception) Tostring() string {
return ("DatabaseException: " + d.Getmessage())
}

type Modelbuilder struct {
    Modelclass string
    Meta interface{}
    Querybuilder interface{}
}

func (m *Modelbuilder) Init(modelclass string) {
m.Modelclass = modelclass
    m.Meta = ModelGetmeta(modelclass)
    conn := ModelGetconnection()
    tablename := __getMap(m.Meta, "table")
    m.Querybuilder = conn.Table(tablename)
}
func (m *Modelbuilder) Where(column interface{}, operatororvalue interface{}, value interface{}) *Modelbuilder {
if (__typeof(column) == "FUNCTION") {
    m.Querybuilder.Where(column)
} else {
    m.Querybuilder.Where(column, operatororvalue, value)
}
    return m
}
func (m *Modelbuilder) Orwhere(column interface{}, operatororvalue interface{}, value interface{}) *Modelbuilder {
if (__typeof(column) == "FUNCTION") {
    m.Querybuilder.Orwhere(column)
} else {
    m.Querybuilder.Orwhere(column, operatororvalue, value)
}
    return m
}
func (m *Modelbuilder) Wherenot(column interface{}, operatororvalue interface{}, value interface{}) *Modelbuilder {
if (__typeof(column) == "FUNCTION") {
    m.Querybuilder.Wherenot(column)
} else {
    m.Querybuilder.Wherenot(column, operatororvalue, value)
}
    return m
}
func (m *Modelbuilder) Orwherenot(column interface{}, operatororvalue interface{}, value interface{}) *Modelbuilder {
if (__typeof(column) == "FUNCTION") {
    m.Querybuilder.Orwherenot(column)
} else {
    m.Querybuilder.Orwherenot(column, operatororvalue, value)
}
    return m
}
func (m *Modelbuilder) Wherein(column string, values interface{}) *Modelbuilder {
m.Querybuilder.Wherein(column, values)
    return m
}
func (m *Modelbuilder) Wherenotin(column string, values interface{}) *Modelbuilder {
m.Querybuilder.Wherenotin(column, values)
    return m
}
func (m *Modelbuilder) Orwherein(column string, values interface{}) *Modelbuilder {
m.Querybuilder.Orwherein(column, values)
    return m
}
func (m *Modelbuilder) Orwherenotin(column string, values interface{}) *Modelbuilder {
m.Querybuilder.Orwherenotin(column, values)
    return m
}
func (m *Modelbuilder) Wherenull(column string) *Modelbuilder {
m.Querybuilder.Wherenull(column)
    return m
}
func (m *Modelbuilder) Wherenotnull(column string) *Modelbuilder {
m.Querybuilder.Wherenotnull(column)
    return m
}
func (m *Modelbuilder) Orwherenull(column string) *Modelbuilder {
m.Querybuilder.Orwherenull(column)
    return m
}
func (m *Modelbuilder) Orwherenotnull(column string) *Modelbuilder {
m.Querybuilder.Orwherenotnull(column)
    return m
}
func (m *Modelbuilder) Wherebetween(column string, values interface{}) *Modelbuilder {
m.Querybuilder.Wherebetween(column, values)
    return m
}
func (m *Modelbuilder) Wherenotbetween(column string, values interface{}) *Modelbuilder {
m.Querybuilder.Wherenotbetween(column, values)
    return m
}
func (m *Modelbuilder) Orwherebetween(column string, values interface{}) *Modelbuilder {
m.Querybuilder.Orwherebetween(column, values)
    return m
}
func (m *Modelbuilder) Orwherenotbetween(column string, values interface{}) *Modelbuilder {
m.Querybuilder.Orwherenotbetween(column, values)
    return m
}
func (m *Modelbuilder) Wherelike(column string, value string) *Modelbuilder {
m.Querybuilder.Wherelike(column, value)
    return m
}
func (m *Modelbuilder) Wherenotlike(column string, value string) *Modelbuilder {
m.Querybuilder.Wherenotlike(column, value)
    return m
}
func (m *Modelbuilder) Orwherelike(column string, value string) *Modelbuilder {
m.Querybuilder.Orwherelike(column, value)
    return m
}
func (m *Modelbuilder) Orwherenotlike(column string, value string) *Modelbuilder {
m.Querybuilder.Orwherenotlike(column, value)
    return m
}
func (m *Modelbuilder) Wherecolumn(first string, operatororsecond interface{}, second string) *Modelbuilder {
m.Querybuilder.Wherecolumn(first, operatororsecond, second)
    return m
}
func (m *Modelbuilder) Whereany(columns interface{}, operator string, value interface{}) *Modelbuilder {
m.Querybuilder.Whereany(columns, operator, value)
    return m
}
func (m *Modelbuilder) Whereall(columns interface{}, operator string, value interface{}) *Modelbuilder {
m.Querybuilder.Whereall(columns, operator, value)
    return m
}
func (m *Modelbuilder) Wherenone(columns interface{}, operator string, value interface{}) *Modelbuilder {
m.Querybuilder.Wherenone(columns, operator, value)
    return m
}
func (m *Modelbuilder) Whereraw(sql string, bindings interface{}) *Modelbuilder {
m.Querybuilder.Whereraw(sql, bindings)
    return m
}
func (m *Modelbuilder) Orwhereraw(sql string, bindings interface{}) *Modelbuilder {
m.Querybuilder.Orwhereraw(sql, bindings)
    return m
}
func (m *Modelbuilder) Selectraw(expression string, bindings interface{}) *Modelbuilder {
m.Querybuilder.Selectraw(expression, bindings)
    return m
}
func (m *Modelbuilder) Orderbyraw(sql string, bindings interface{}) *Modelbuilder {
m.Querybuilder.Orderbyraw(sql, bindings)
    return m
}
func (m *Modelbuilder) Havingraw(sql string, bindings interface{}) *Modelbuilder {
m.Querybuilder.Havingraw(sql, bindings)
    return m
}
func (m *Modelbuilder) Select(columns interface{}) *Modelbuilder {
m.Querybuilder.Select(columns)
    return m
}
func (m *Modelbuilder) Distinct() *Modelbuilder {
m.Querybuilder.Distinct()
    return m
}
func (m *Modelbuilder) Orderby(column string, direction string) *Modelbuilder {
m.Querybuilder.Orderby(column, direction)
    return m
}
func (m *Modelbuilder) Orderbydesc(column string) *Modelbuilder {
m.Querybuilder.Orderbydesc(column)
    return m
}
func (m *Modelbuilder) Latest(column string) *Modelbuilder {
m.Querybuilder.Latest(column)
    return m
}
func (m *Modelbuilder) Oldest(column string) *Modelbuilder {
m.Querybuilder.Oldest(column)
    return m
}
func (m *Modelbuilder) Groupby(columns interface{}) *Modelbuilder {
m.Querybuilder.Groupby(columns)
    return m
}
func (m *Modelbuilder) Having(column string, operator interface{}, value interface{}) *Modelbuilder {
m.Querybuilder.Having(column, operator, value)
    return m
}
func (m *Modelbuilder) Limit(count int64) *Modelbuilder {
m.Querybuilder.Limit(count)
    return m
}
func (m *Modelbuilder) Take(count int64) *Modelbuilder {
return m.Limit(count)
}
func (m *Modelbuilder) Offset(count int64) *Modelbuilder {
m.Querybuilder.Offset(count)
    return m
}
func (m *Modelbuilder) Skip(count int64) *Modelbuilder {
return m.Offset(count)
}
func (m *Modelbuilder) Forpage(page int64, perpage int64) *Modelbuilder {
m.Querybuilder.Forpage(page, perpage)
    return m
}
func (m *Modelbuilder) When(condition bool, callback interface{}) *Modelbuilder {
if condition {
    callback(m)
}
    return m
}
func (m *Modelbuilder) Get() interface{} {
rows := m.Querybuilder.Get()
    return m.Hydratemany(rows)
}
func (m *Modelbuilder) First() interface{} {
m.Querybuilder.Limit(1)
    rows := m.Querybuilder.Get()
    if (__len(rows) == 0) {
    return nil
}
    return m.Hydrateone(__getIndex(rows, int(0)))
}
func (m *Modelbuilder) Firstorfail() interface{} {
result := m.First()
    if (result == nil) {
    panic(func() *Databaseexception { __t := &Databaseexception{}; __t.Init(("No query results for model: " + m.Modelclass)); return __t }())
}
    return result
}
func (m *Modelbuilder) Find(id interface{}) interface{} {
primarykey := __getMap(m.Meta, "primaryKey")
    return m.Where(primarykey, id).First()
}
func (m *Modelbuilder) Findorfail(id interface{}) interface{} {
result := m.Find(id)
    if (result == nil) {
    panic(func() *Databaseexception { __t := &Databaseexception{}; __t.Init(((("No query results for model: " + m.Modelclass) + " with id ") + fmt.Sprint(id))); return __t }())
}
    return result
}
func (m *Modelbuilder) Value(column string) interface{} {
m.Querybuilder.Limit(1)
    rows := m.Querybuilder.Get()
    if (__len(rows) == 0) {
    return nil
}
    row := __getIndex(rows, int(0))
    if __isset(row, column) {
    return __getMap(row, column)
}
    return nil
}
func (m *Modelbuilder) Pluck(column string, key string) interface{} {
rows := m.Querybuilder.Get()
    if (key == "") {
    result := []interface{}{}
    for i := 0; (i < __len(rows)); i++ {
    row := __getIndex(rows, i)
    if __isset(row, column) {
    func() { result = append(result, __getMap(row, column)) }()
}
}
    return result
} else {
    result := map[string]interface{}{}
    for i := 0; (i < __len(rows)); i++ {
    row := __getIndex(rows, i)
    if (__isset(row, column) && __isset(row, key)) {
    k := fmt.Sprint(__getMap(row, key))
    __setMap(result, k, __getMap(row, column))
}
}
    return result
}
}
func (m *Modelbuilder) Exists() bool {
return m.Querybuilder.Exists()
}
func (m *Modelbuilder) Count(column string) int64 {
return m.Querybuilder.Count(column)
}
func (m *Modelbuilder) Max(column string) interface{} {
return m.Querybuilder.Max(column)
}
func (m *Modelbuilder) Min(column string) interface{} {
return m.Querybuilder.Min(column)
}
func (m *Modelbuilder) Avg(column string) interface{} {
return m.Querybuilder.Avg(column)
}
func (m *Modelbuilder) Sum(column string) interface{} {
return m.Querybuilder.Sum(column)
}
func (m *Modelbuilder) Insert(values interface{}) bool {
return m.Querybuilder.Insert(values)
}
func (m *Modelbuilder) Insertgetid(values interface{}) int64 {
return m.Querybuilder.Insertgetid(values)
}
func (m *Modelbuilder) Update(attributes interface{}) int64 {
columns := __getMap(m.Meta, "columns")
    data := map[string]interface{}{}
    for fieldname, value := range __toMap(attributes) {
    if __isset(columns, fieldname) {
    colinfo := __getMap(columns, fieldname)
    columnname := __getMap(colinfo, "column")
    __setMap(data, columnname, value)
} else {
    __setMap(data, fieldname, value)
}
}
    return m.Querybuilder.Update(data)
}
func (m *Modelbuilder) Increment(column string, amount int64, extra interface{}) int64 {
return m.Querybuilder.Increment(column, amount, extra)
}
func (m *Modelbuilder) Decrement(column string, amount int64, extra interface{}) int64 {
return m.Querybuilder.Decrement(column, amount, extra)
}
func (m *Modelbuilder) Delete() int64 {
return m.Querybuilder.Delete()
}
func (m *Modelbuilder) Hydrateone(row interface{}) interface{} {
instance := __createInstance(m.Modelclass)
    instance.Hydrate(row, m.Meta)
    return instance
}
func (m *Modelbuilder) Hydratemany(rows interface{}) interface{} {
result := []interface{}{}
    for i := 0; (i < __len(rows)); i++ {
    func() { result = append(result, m.Hydrateone(__getIndex(rows, i))) }()
}
    return result
}

func main() {
    Main()
}
