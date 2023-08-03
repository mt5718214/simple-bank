package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/util"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCreateTransfer(t *testing.T) {
	USD := util.USD
	account1 := randomAccount()
	account2 := randomAccount()
	account1.Currency = USD
	account2.Currency = USD

	testCase := []struct {
		name          string
		input         transferReq
		buildStubs    func(store *mockdb.MockStore, arg transferReq)
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			input: transferReq{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        10,
				Currency:      USD,
			},
			buildStubs: func(store *mockdb.MockStore, arg transferReq) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(arg.FromAccountID)).AnyTimes().Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(arg.ToAccountID)).AnyTimes().Return(account2, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
					FromAccountID: arg.FromAccountID,
					ToAccountID:   arg.ToAccountID,
					Amount:        arg.Amount,
				})).Times(1).Return(db.TransferTxResult{}, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name: "AccountNotFound",
			input: transferReq{
				FromAccountID: util.RandomInt(1, 1000),
				ToAccountID:   account2.ID,
				Amount:        10,
				Currency:      USD,
			},
			buildStubs: func(store *mockdb.MockStore, arg transferReq) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(arg.FromAccountID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, w.Code)
			},
		},
		{
			name: "CurrencyMismatch",
			input: transferReq{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        10,
				Currency:      util.EUR,
			},
			buildStubs: func(store *mockdb.MockStore, arg transferReq) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(arg.FromAccountID)).
					Times(1).
					Return(account1, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
		{
			name: "InternalServerError",
			input: transferReq{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        10,
				Currency:      USD,
			},
			buildStubs: func(store *mockdb.MockStore, arg transferReq) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(arg.FromAccountID)).AnyTimes().Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(arg.ToAccountID)).AnyTimes().Return(account2, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
					FromAccountID: arg.FromAccountID,
					ToAccountID:   arg.ToAccountID,
					Amount:        arg.Amount,
				})).Times(1).Return(db.TransferTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, w.Code)
			},
		},
	}

	for i := range testCase {
		tc := testCase[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)

			arg := tc.input
			jsonValue, err := json.Marshal(arg)
			require.NoError(t, err)

			server := newTestServer(t, store)
			w := httptest.NewRecorder()

			tc.buildStubs(store, arg)

			url := "/transfers"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonValue))
			require.NoError(t, err)

			server.router.ServeHTTP(w, req)
			tc.checkResponse(t, w)
		})
	}
}
