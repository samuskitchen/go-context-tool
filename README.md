## GO-CONTEXT-TOOL
It is a helper that helps us perform dynamic queries using GORM, in terms of the limit of records consulted and the offset, we can also indicate which fields can be omitted, which will not be consulted in the database.

* Works with: `not supported by other web frameworks`
    * GORM go orm
    * ECHO go web framework

Ejemplo
```go

    func (h *handler) FindAll(c echo.Context) error {
        data, err := h.service.FindAll(ctxman.NewContextTool(c)) ///context-tool prepare and retrieve parameters in the context
        if err != nil {
            log.Error(err)
            return c.String(http.StatusBadRequest, "something happened")
        }
		
        return c.JSON(http.StatusOK, data)
    }

    // Implementation of logical layer or services
    type Service struct {
	    repo repositries.Repository
    }   

    // Implementation of the Skip interface of context_tool which helps to indicate 
	// which attributes are skippable and which have to be preloaded by GORM
    func (s *Service) OmitFields() ([]string, []string) { 
        // First slice indicates omittable attributes and the second slice pre-loaded ones
        // Gorm will preload all the fields that appear in the slide       
	    return []string{"Name"}, []string{"Books"}
    }
	
    func (s *Service) FindAll(ctx ctxman.Ctxx) ([]*models.Data, error) {
        // We prepare the context by passing it the implementation of the Skip interface
        return s.repo.FindAll(ctx.WithSkip(s))
    }
	
    func (s *editorialService) FindByID(ctx ctxman.Ctxx, ID uint) (*models.Data, error) {
        // We prepare the context by passing it the implementation of the Skip interface
	    return s.r.FindByCode(ctx.WithSkip(s), ID)
    }

    //Repository Implementation
    type Repository struct {
	    grom_conn *gorm.DB
    }
	
    func (r *Repository) FindAll(ctx ctxman.Ctxx) ([]*models.Data, error) {
        datos := []*models.Data{}
        tx := ctx.FormatGORM(r.grom_conn) // Configure gorm connection
        if err := tx.Find(&datos).Error; err != nil {
            return nil, err
        }
		
        return datos, nil
    }
	
    func (r *Repository) FindByID(ctx ctxman.Ctxx, code uint) (*models.Data, error) {
        data := models.Data{}
        // SimpleGORM ignores limit and offset
        tx := ctx.SimpleGORM(r.grom_conn) // Configure gorm connection
        if err := tx.Find(&data, "id=?", code).Error; err != nil {
            return nil, err
        }
		
        return &data, nil
    }
```
When consulting a URL we can pass the fields in the query params:
- `limit`: indicates the number of records
- `offset`: indicates where the registers will be read from, or offset
- `skip`:  The fields that you want to skip in the query, must be separated by commas and as defined in the skip interface, probably CamelCase

Example:
- GET: `http://localhost:8080/data?skip=Description,Address,Monto&offset=10&limit=10`