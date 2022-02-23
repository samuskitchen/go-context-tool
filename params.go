package context_tool

import "strings"

type Params struct {
	offset     int
	limit      int
	skipFields []string
}

func (p *Params) OffSet() int        { return p.offset }
func (p *Params) Limit() int         { return p.limit }
func (p *Params) SkipFields() string { return strings.Join(p.skipFields, "") }
