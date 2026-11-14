package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"helpdesk-ticketing-system/backend/internal/model"
)

func callUserAPI(t *testing.T, router http.Handler, method, path, token, payload string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestAdminUserManagementEndpoints(t *testing.T) {
	repo := newMockUserRepository()
	admin := &model.User{ID: uuid.New(), Name: "Administrator", Email: "admin-users@example.com", Role: model.RoleAdmin}
	requester := &model.User{ID: uuid.New(), Name: "Requester", Email: "requester-users@example.com", Role: model.RoleRequester}
	agent := &model.User{ID: uuid.New(), Name: "Agent", Email: "agent-users@example.com", Role: model.RoleAgent}
	if err := repo.Create(context.Background(), admin); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(context.Background(), requester); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(context.Background(), agent); err != nil {
		t.Fatal(err)
	}
	router, auth := setupAuthTestRouter(repo, 24)
	adminToken, err := auth.GenerateToken(admin)
	if err != nil {
		t.Fatal(err)
	}
	requesterToken, err := auth.GenerateToken(requester)
	if err != nil {
		t.Fatal(err)
	}
	agentToken, err := auth.GenerateToken(agent)
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/users/" + uuid.NewString()
	for _, identity := range []struct{ name, token string }{{"requester", requesterToken}, {"agent", agentToken}} {
		for _, endpoint := range []struct{ method, path string }{{http.MethodPost, "/api/users"}, {http.MethodPatch, path}, {http.MethodDelete, path}} {
			w := callUserAPI(t, router, endpoint.method, endpoint.path, identity.token, `{"name":"Blocked User","email":"blocked@example.com","password":"secret1","role":"agent"}`)
			if w.Code != http.StatusForbidden {
				t.Errorf("%s %s status=%d body=%s", identity.name, endpoint.method, w.Code, w.Body.String())
			}
		}
	}
	created := callUserAPI(t, router, http.MethodPost, "/api/users", adminToken, `{"name":"New Agent","email":"new-agent@example.com","password":"secret1","role":"agent"}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	var envelope struct {
		Data model.UserResponse `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	id := envelope.Data.ID.String()
	if bytes.Contains(created.Body.Bytes(), []byte("password_hash")) {
		t.Fatal("password hash leaked")
	}
	dup := callUserAPI(t, router, http.MethodPost, "/api/users", adminToken, `{"name":"Duplicate","email":"new-agent@example.com","password":"secret1","role":"agent"}`)
	if dup.Code != http.StatusConflict {
		t.Errorf("duplicate status=%d body=%s", dup.Code, dup.Body.String())
	}
	updated := callUserAPI(t, router, http.MethodPatch, "/api/users/"+id, adminToken, `{"name":"Updated Agent","email":"updated-agent@example.com"}`)
	if updated.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", updated.Code, updated.Body.String())
	}
	deleted := callUserAPI(t, router, http.MethodDelete, "/api/users/"+id, adminToken, "")
	if deleted.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleted.Code, deleted.Body.String())
	}
	invalid := callUserAPI(t, router, http.MethodPatch, "/api/users/not-a-uuid", adminToken, `{"name":"Some User","email":"some@example.com"}`)
	if invalid.Code != http.StatusBadRequest {
		t.Errorf("invalid ID status=%d", invalid.Code)
	}
}
