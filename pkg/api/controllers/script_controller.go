package controllers

import (
	"net/http"
	"strings"

	"github.com/formancehq/go-libs/api"
	"github.com/formancehq/go-libs/logging"
	"github.com/gin-gonic/gin"
	"github.com/numary/ledger/pkg/api/apierrors"
	"github.com/numary/ledger/pkg/core"
	"github.com/numary/ledger/pkg/ledger"
)

type ScriptResponse struct {
	api.ErrorResponse
	Transaction *core.ExpandedTransaction `json:"transaction,omitempty"`
}

type ScriptController struct{}

func NewScriptController() ScriptController {
	return ScriptController{}
}

func (ctl *ScriptController) PostScript(c *gin.Context) {
	l, _ := c.Get("ledger")

	var script core.ScriptData
	if err := c.ShouldBindJSON(&script); err != nil {
		panic(err)
	}

	value, ok := c.GetQuery("preview")
	preview := ok && (strings.ToUpper(value) == "YES" || strings.ToUpper(value) == "TRUE" || value == "1")

	res := ScriptResponse{}
	execRes, err := l.(*ledger.Ledger).ExecuteScript(c.Request.Context(), preview, script)
	if err != nil {
		var (
			code    = apierrors.ErrInternal
			message string
		)
		switch e := err.(type) {
		case *ledger.ScriptError:
			code = e.Code
			message = e.Message
		case *ledger.ConflictError:
			code = apierrors.ErrConflict
			message = e.Error()
		default:
			logging.GetLogger(c.Request.Context()).Errorf(
				"internal errors executing script: %s", err)
		}
		res.ErrorResponse = api.ErrorResponse{
			ErrorCode:              code,
			ErrorMessage:           message,
			ErrorCodeDeprecated:    code,
			ErrorMessageDeprecated: message,
		}
		if message != "" {
			res.Details = apierrors.EncodeLink(message)
		}
	}
	res.Transaction = &execRes

	c.JSON(http.StatusOK, res)
}
