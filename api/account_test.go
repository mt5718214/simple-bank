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
	"simplebank/util"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccount(t *testing.T) {
	account := randomAccount()

	testCase := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
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
			name:      "AccountNotFound",
			accountID: account.ID,
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

			server.router.ServeHTTP(w, req)
			// check res
			tc.checkResponse(t, w)
		})
	}
}

func TestCreateAccount(t *testing.T) {
	testCase := []struct {
		name          string
		input         createAccountReq
		buildStubs    func(store *mockdb.MockStore, arg createAccountReq)
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			input: createAccountReq{
				Owner:    util.RandomOwner(),
				Currency: util.RandomCurrency(),
			},
			buildStubs: func(store *mockdb.MockStore, arg createAccountReq) {
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
			name: "InvalidOwner",
			input: createAccountReq{
				Currency: util.RandomCurrency(),
			},
			buildStubs: func(store *mockdb.MockStore, arg createAccountReq) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
		{
			name: "InvalidCurrency",
			input: createAccountReq{
				Owner:    util.RandomOwner(),
				Currency: util.RandomString(4),
			},
			buildStubs: func(store *mockdb.MockStore, arg createAccountReq) {
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
			input: createAccountReq{
				Owner:    util.RandomOwner(),
				Currency: util.RandomCurrency(),
			},
			buildStubs: func(store *mockdb.MockStore, arg createAccountReq) {
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

			arg := tc.input
			jsonVal, err := json.Marshal(arg)
			require.NoError(t, err)

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store, arg)

			server := newTestServer(t, store)
			w := httptest.NewRecorder()

			url := "/accounts"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonVal))
			require.NoError(t, err)

			server.router.ServeHTTP(w, req)
			tc.checkResponse(t, w)
		})
	}
}

func TestListAccount(t *testing.T) {
	accounts := make([]db.Account, 5)
	for i := 0; i < 5; i++ {
		accounts = append(accounts, randomAccount())
	}
	testCase := []struct {
		name          string
		input         listAccountReq
		buildStubs    func(store *mockdb.MockStore, arg listAccountReq)
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			input: listAccountReq{
				PageID:   1,
				PageSize: 5,
			},
			buildStubs: func(store *mockdb.MockStore, arg listAccountReq) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{
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
			name: "InvalidPageId",
			input: listAccountReq{
				PageID:   0,
				PageSize: 5,
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
			buildStubs: func(store *mockdb.MockStore, arg listAccountReq) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{
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
			server.router.ServeHTTP(w, req)

			tc.checkResponse(t, w)
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	account := randomAccount()
	testCase := []struct {
		name          string
		input         updateAccountReq
		buildStubs    func(store *mockdb.MockStore, arg updateAccountReq)
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			input: updateAccountReq{
				ID:      account.ID,
				Balance: util.RandomBalance(),
			},
			buildStubs: func(store *mockdb.MockStore, arg updateAccountReq) {
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(db.UpdateAccountParams{
						ID:      account.ID,
						Balance: arg.Balance,
					})).
					Times(1).
					Return(db.Account{}, nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, w.Code)
			},
		},
		{
			name: "InvalidID",
			input: updateAccountReq{
				ID:      0,
				Balance: util.RandomBalance(),
			},
			buildStubs: func(store *mockdb.MockStore, arg updateAccountReq) {
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(db.UpdateAccountParams{
						ID:      account.ID,
						Balance: arg.Balance,
					})).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
		{
			name: "InvalBalance",
			input: updateAccountReq{
				ID:      account.ID,
				Balance: -1,
			},
			buildStubs: func(store *mockdb.MockStore, arg updateAccountReq) {
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(db.UpdateAccountParams{
						ID:      account.ID,
						Balance: arg.Balance,
					})).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
		{
			name: "InternalServerError",
			input: updateAccountReq{
				ID:      account.ID,
				Balance: util.RandomBalance(),
			},
			buildStubs: func(store *mockdb.MockStore, arg updateAccountReq) {
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(db.UpdateAccountParams{
						ID:      account.ID,
						Balance: arg.Balance,
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
			store := mockdb.NewMockStore(ctrl)

			arg := tc.input
			jsonVal, err := json.Marshal(arg)
			require.NoError(t, err)

			tc.buildStubs(store, arg)
			server := newTestServer(t, store)
			w := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", arg.ID)
			req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonVal))
			require.NoError(t, err)

			server.router.ServeHTTP(w, req)
			tc.checkResponse(t, w)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	account := randomAccount()

	testCase := []struct {
		name          string
		ID            int64
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			ID:   account.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, w.Code)
			},
		},
		{
			name: "InvalidID",
			ID:   0,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
		{
			name: "InternalServerError",
			ID:   account.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(sql.ErrConnDone)
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
			server := newTestServer(t, store)

			tc.buildStub(store)

			url := fmt.Sprintf("/accounts/%d", tc.ID)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			tc.checkResponse(t, w)
		})
	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
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
