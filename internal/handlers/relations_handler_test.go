package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
	"example/mymodule/internal/servise"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newRelationsTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := db.AutoMigrate(&model.Like{}, &model.Playlist{}, &model.Repost{}); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	NewLikeHandler(servise.NewLikeService(repository.NewLikeRepository(db))).RegisterRoutes(router)
	NewPlaylistHandler(servise.NewPlaylistService(repository.NewPlaylistRepository(db))).RegisterRoutes(router)
	NewRepostHandler(servise.NewRepostService(repository.NewRepostRepository(db))).RegisterRoutes(router)
	return router
}

func sendJSON(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestLikeEndpoints(t *testing.T) {
	router := newRelationsTestRouter(t)

	created := sendJSON(router, http.MethodPost, "/api/likes", `{"id_user":1,"id_music":42}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("POST /api/likes status = %d, want %d: %s", created.Code, http.StatusCreated, created.Body)
	}

	response := sendJSON(router, http.MethodGet, "/api/likes?user_id=1", "")
	if response.Code != http.StatusOK {
		t.Fatalf("GET /api/likes status = %d, want %d: %s", response.Code, http.StatusOK, response.Body)
	}
	var likes []model.Like
	if err := json.Unmarshal(response.Body.Bytes(), &likes); err != nil {
		t.Fatal(err)
	}
	if len(likes) != 1 || len(likes[0].ID_music) != 1 || likes[0].ID_music[0] != 42 {
		t.Fatalf("unexpected likes response: %+v", likes)
	}
}

func TestPlaylistEndpointsOnlyListPublicPlaylists(t *testing.T) {
	router := newRelationsTestRouter(t)

	for _, body := range []string{
		`{"id_user":1,"name":"Public","private":false}`,
		`{"id_user":1,"name":"Private","private":true}`,
	} {
		response := sendJSON(router, http.MethodPost, "/api/playlists", body)
		if response.Code != http.StatusCreated {
			t.Fatalf("POST /api/playlists status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body)
		}
	}

	response := sendJSON(router, http.MethodGet, "/api/playlists?user_id=1", "")
	if response.Code != http.StatusOK {
		t.Fatalf("GET /api/playlists status = %d, want %d: %s", response.Code, http.StatusOK, response.Body)
	}
	var playlists []model.Playlist
	if err := json.Unmarshal(response.Body.Bytes(), &playlists); err != nil {
		t.Fatal(err)
	}
	if len(playlists) != 1 || playlists[0].Name != "Public" {
		t.Fatalf("unexpected public playlists response: %+v", playlists)
	}
}

func TestRepostEndpoints(t *testing.T) {
	router := newRelationsTestRouter(t)

	created := sendJSON(router, http.MethodPost, "/api/reposts", `{"id_user":1,"id_music":42}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("POST /api/reposts status = %d, want %d: %s", created.Code, http.StatusCreated, created.Body)
	}

	response := sendJSON(router, http.MethodGet, "/api/reposts?user_id=1", "")
	if response.Code != http.StatusOK {
		t.Fatalf("GET /api/reposts status = %d, want %d: %s", response.Code, http.StatusOK, response.Body)
	}
	var reposts []model.Repost
	if err := json.Unmarshal(response.Body.Bytes(), &reposts); err != nil {
		t.Fatal(err)
	}
	if len(reposts) != 1 || len(reposts[0].ID_music) != 1 || reposts[0].ID_music[0] != 42 {
		t.Fatalf("unexpected reposts response: %+v", reposts)
	}
}
