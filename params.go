package context_tool

import "strings"

type Params struct {
	offset     int
	limit      int
	omitFields []string
}

func (p *Params) OffSet() int        { return p.offset }
func (p *Params) Limit() int         { return p.limit }
func (p *Params) OmitFields() string { return strings.Join(p.omitFields, "") }
