package context_tool

import (
	"strconv"
	"strings"

	"gorm.io/gorm"
)

const (
	defaultFiles = 10
	maxFields    = 100
	errMessage   = "query param `%s` not recognized, please check the documentation, or try using CamelCase notation"
)

type MapFunc map[string]func(tx *gorm.DB) *gorm.DB

type Skip interface {
	SkipFields() ([]string, []string)
}

type QueryParameter interface {
	QueryParam(name string) string
}

type ContextToolInterface interface {
	GetParams() Params

	//WithSkip receives as a parameter the implementation of:
	//	type Skip interface {
	//		//return (omitted parameters, preloads)
	//		OmitFields() ([]string, []string)
	//	}
	WithSkip(s Skip) ContextToolInterface

	AddCustomPreloadFunc(fns MapFunc)

	// SimpleGORM ideal for one-row queries, offset 0 and limit at 1
	SimpleGORM(conn *gorm.DB, preloads ...string) *gorm.DB

	// FormatGORM prepare offset and limit according to parameters
	FormatGORM(conn *gorm.DB, preloads ...string) *gorm.DB
}

type contextTool struct {
	Params
	fieldsForOmit    map[string]struct{}
	fieldsForPreload map[string]struct{}
	preloadFunctions MapFunc
}

// NewContextTool prepare the context
func NewContextTool(c QueryParameter) ContextToolInterface {
	var err error
	var skip string = c.QueryParam("skip")
	offset, err := strconv.Atoi(c.QueryParam("offset"))
	if err != nil {
		offset = 0
	}

	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil {
		limit = defaultFiles
	}

	if limit == 0 {
		limit = defaultFiles
	}

	if limit > maxFields {
		limit = maxFields
	}

	var omits []string
	if skip != "" {
		omits = strings.Split(skip, ",")
	}

	return &contextTool{
		Params: Params{
			skipFields: omits,
			offset:     offset,
			limit:      limit,
		},
		fieldsForOmit:    make(map[string]struct{}),
		fieldsForPreload: make(map[string]struct{}),
		preloadFunctions: make(MapFunc),
	}
}

func (c *contextTool) GetParams() Params {
	return c.Params
}

func (c *contextTool) WithSkip(skip Skip) ContextToolInterface {
	allowsOmits, allowPreloads := skip.SkipFields()
	for _, alp := range allowPreloads {
		c.fieldsForPreload[alp] = struct{}{}
	}

	for _, fil := range c.skipFields {
		if search(allowsOmits, fil) {
			c.fieldsForOmit[fil] = struct{}{}
			continue
		}

		if search(allowPreloads, fil) {
			delete(c.fieldsForPreload, fil)
			continue
		}
	}

	return c
}

// AddCustomPreloadFunc allows adding functions, which are executed before preloading, ideal for configuring omits
// selects or limits Keymap: Name of the field that will be preloaded
func (c *contextTool) AddCustomPreloadFunc(fns MapFunc) {
	for key, f := range fns {
		c.preloadFunctions[key] = f
	}
}

func (c *contextTool) formatGorm(conn *gorm.DB, simple bool, preloads []string) *gorm.DB {
	for _, p := range preloads {
		c.fieldsForOmit[p] = struct{}{}
	}
	var preload []string
	for val := range c.fieldsForOmit {
		preload = append(preload, val)
	}

	var tx *gorm.DB
	if simple {
		tx = conn.Limit(1).Omit(preload...)
	} else {
		tx = conn.Limit(c.limit).Offset(c.offset).Omit(preload...)
	}

	for p := range c.fieldsForPreload {
		function, ok := c.preloadFunctions[p]
		if ok {
			tx = tx.Preload(p, function)
			continue
		}

		tx = tx.Preload(p)
	}

	return tx
}

// SimpleGORM ideal for one-row queries, offset 0 and limit at 1
func (c *contextTool) SimpleGORM(conn *gorm.DB, preloads ...string) *gorm.DB {
	return c.formatGorm(conn, true, preloads)
}

// FormatGORM prepare offset and limit according to parameters
func (c *contextTool) FormatGORM(conn *gorm.DB, preloads ...string) *gorm.DB {
	return c.formatGorm(conn, false, preloads)
}

func search(collection []string, s string) bool {
	for _, item := range collection {
		if item == s {
			return true
		}
	}

	return false
}
