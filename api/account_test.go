package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/token"
	"simplebank/util"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccount(t *testing.T) {
	user, _ := randomUser()
	account := randomAccount(user.Username)

	testCase := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, w.Code)
				requireBodyMatchAccount(t, w.Body, account)
			},
		},
		{
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, "unauthorized_user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name:      "AccountNotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, w.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, w.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
	}

	for i := range testCase {
		tc := testCase[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// build stubs
			tc.buildStubs(store)

			// test server and send request
			server := newTestServer(t, store)
			w := httptest.NewRecorder()
			url := fmt.Sprintf("/accounts/%d", tc.accountID)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, server.tokenMaker, req)
			server.router.ServeHTTP(w, req)
			// check res
			tc.checkResponse(t, w)
		})
	}
}

func TestCreateAccount(t *testing.T) {
	user, _ := randomUser()
	account := randomAccount(user.Username)

	testCase := []struct {
		name          string
		input         gin.H
		setupAuth     func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request)
		buildStubs    func(store *mockdb.MockStore, arg db.CreateAccountParams)
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			input: gin.H{
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, arg db.CreateAccountParams) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(db.CreateAccountParams{
						Owner:    arg.Owner,
						Currency: arg.Currency,
						Balance:  0,
					})).
					Times(1).
					Return(db.Account{}, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name: "NoAuthorization",
			input: gin.H{
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
			},
			buildStubs: func(store *mockdb.MockStore, arg db.CreateAccountParams) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name: "InvalidCurrency",
			input: gin.H{
				"currency": util.RandomString(4),
			},
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, arg db.CreateAccountParams) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
		{
			name: "InternalServerError",
			input: gin.H{
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, arg db.CreateAccountParams) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(db.CreateAccountParams{
						Owner:    arg.Owner,
						Currency: arg.Currency,
						Balance:  0,
					})).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
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

			jsonVal, err := json.Marshal(tc.input)
			require.NoError(t, err)

			arg := db.CreateAccountParams{
				Owner:    account.Owner,
				Currency: account.Currency,
			}
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store, arg)

			server := newTestServer(t, store)
			w := httptest.NewRecorder()

			url := "/accounts"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonVal))
			require.NoError(t, err)

			tc.setupAuth(t, server.tokenMaker, req)
			server.router.ServeHTTP(w, req)
			tc.checkResponse(t, w)
		})
	}
}

func TestListAccount(t *testing.T) {
	user, _ := randomUser()
	accounts := make([]db.Account, 5)
	for i := 0; i < 5; i++ {
		accounts = append(accounts, randomAccount(user.Username))
	}
	testCase := []struct {
		name          string
		input         listAccountReq
		setupAuth     func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request)
		buildStubs    func(store *mockdb.MockStore, arg listAccountReq)
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			input: listAccountReq{
				PageID:   1,
				PageSize: 5,
			},
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, arg listAccountReq) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{
						Owner:  user.Username,
						Limit:  arg.PageSize,
						Offset: (arg.PageID - 1) * arg.PageSize,
					})).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, w.Code)

				var res []db.Account
				err := json.Unmarshal(w.Body.Bytes(), &res)
				require.NoError(t, err)
				require.Equal(t, accounts, res)
			},
		},
		{
			name: "NoAuthorization",
			input: listAccountReq{
				PageID:   1,
				PageSize: 5,
			},
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
			},
			buildStubs: func(store *mockdb.MockStore, arg listAccountReq) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name: "InvalidPageId",
			input: listAccountReq{
				PageID:   0,
				PageSize: 5,
			},
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, arg listAccountReq) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
		{
			name: "InvalidPageSize",
			input: listAccountReq{
				PageID:   1,
				PageSize: 0,
			},
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, arg listAccountReq) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
		{
			name: "InternalServerError",
			input: listAccountReq{
				PageID:   1,
				PageSize: 5,
			},
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, arg listAccountReq) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{
						Owner:  user.Username,
						Limit:  arg.PageSize,
						Offset: (arg.PageID - 1) * arg.PageSize,
					})).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
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

			arg := tc.input
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store, arg)

			url := fmt.Sprintf("/accounts?page_id=%d&page_size=%d", arg.PageID, arg.PageSize)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server := newTestServer(t, store)
			w := httptest.NewRecorder()
			tc.setupAuth(t, server.tokenMaker, req)
			server.router.ServeHTTP(w, req)

			tc.checkResponse(t, w)
		})
	}
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomBalance(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
