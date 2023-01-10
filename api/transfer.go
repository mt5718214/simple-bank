package api

import (
	"database/sql"
	"fmt"
	"net/http"
	db "simplebank/db/sqlc"

	"github.com/gin-gonic/gin"
)

type transferReq struct {
	From_account_id int64  `json:"from_account_id" binding:"required,min=1"`
	To_account_id   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount          int64  `json:"amount" binding:"required,gt=0"`
	Currency        string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(c *gin.Context) {
	var req transferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !server.validAccount(c, req.From_account_id, req.Currency) {
		return
	}

	if !server.validAccount(c, req.From_account_id, req.Currency) {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.From_account_id,
		ToAccountID:   req.To_account_id,
		Amount:        req.Amount,
	}
	result, err := server.store.TransferTx(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}

func (server *Server) validAccount(c *gin.Context, accountID int64, currency string) bool {
	account, err := server.store.GetAccount(c, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}

		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", accountID, account.Currency, currency)
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}

	return true
}
