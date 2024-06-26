package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/FrostJ143/simplebank/internal/query"
	"github.com/FrostJ143/simplebank/internal/token"
	"github.com/gin-gonic/gin"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id"   binding:"required,min=1"`
	Amount        int64  `json:"amount"          binding:"required,gt=0"`
	Currency      string `json:"currency"        binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	fromAccount, valid := server.validAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(authPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from account doesn't belong to authenticated user")
		ctx.JSON(http.StatusForbidden, errResponse(err))
		return
	}

	if _, valid := server.validAccount(ctx, req.ToAccountID, req.Currency); !valid {
		return
	}

	arg := query.TransferTxParams{
		From_Account_ID: req.FromAccountID,
		To_Account_ID:   req.ToAccountID,
		Amount:          req.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validAccount(
	ctx *gin.Context,
	accountID int64,
	currency string,
) (*query.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errResponse(err))
			return account, false
		}

		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return account, false
	}

	if account.Currenncy != currency {
		err = fmt.Errorf(
			"account [%d] currency mismatch: %s and %s",
			account.ID,
			account.Currenncy,
			currency,
		)
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return account, false
	}

	return account, true
}
