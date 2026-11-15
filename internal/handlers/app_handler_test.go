package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
	"example/mymodule/internal/servise"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newAppTestRouter(t *testing.T) (*gin.Engine, string, *gorm.DB) {
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
	if err := db.AutoMigrate(&model.User{}, &model.Music{}, &model.Like{}, &model.Playlist{}, &model.Repost{}); err != nil {
		t.Fatal(err)
	}

	sessionManager, err := NewSessionManager("test-session-secret-with-more-than-32-bytes")
	if err != nil {
		t.Fatal(err)
	}
	userRepository := repository.NewUserRepository(db)
	musicRepository := repository.NewMusicRepository(db)
	playlistRepository := repository.NewPlaylistRepository(db)
	deletionRepository := repository.NewDeletionRepository(db)
	userService := servise.NewUserService(userRepository)
	musicService := servise.NewMusicService(musicRepository)
	likeService := servise.NewLikeService(repository.NewLikeRepository(db))
	playlistService := servise.NewPlaylistService(playlistRepository)
	repostService := servise.NewRepostService(repository.NewRepostRepository(db))
	adminService := servise.NewAdminService(userRepository, musicRepository, playlistRepository, deletionRepository)

	router := gin.New()
	userHandler := NewUserHandler(userService, sessionManager)
	userHandler.RegisterPublicRoutes(router)
	uploadDir := t.TempDir()
	musicHandler := NewMusicHandler(musicService, deletionRepository, uploadDir)
	musicHandler.RegisterPublicRoutes(router)
	router.Static("/uploads", uploadDir)
	playlistHandler := NewPlaylistHandler(playlistService, deletionRepository)
	playlistHandler.RegisterPublicRoutes(router)
	NewRepostHandler(repostService).RegisterPublicRoutes(router)

	protectedAPI := router.Group("/api")
	protectedAPI.Use(sessionManager.RequireAuth())
	userHandler.RegisterProtectedRoutes(protectedAPI)
	musicHandler.RegisterProtectedRoutes(protectedAPI)
	NewLikeHandler(likeService).RegisterProtectedRoutes(protectedAPI)
	playlistHandler.RegisterProtectedRoutes(protectedAPI)
	NewRepostHandler(repostService).RegisterProtectedRoutes(protectedAPI)
	adminAPI := protectedAPI.Group("/admin")
	adminAPI.Use(userHandler.RequireAdmin())
	NewAdminHandler(adminService, uploadDir).RegisterRoutes(adminAPI)
	return router, uploadDir, db
}

func sendJSON(router http.Handler, method, path, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func registerTestUser(t *testing.T, router http.Handler, login, username string) (*http.Cookie, model.User) {
	t.Helper()
	response := sendJSON(router, http.MethodPost, "/api/users",
		`{"username":"`+username+`","login":"`+login+`","password":"correct-horse-battery-staple"}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("POST /api/users status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body)
	}
	cookies := response.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != sessionCookieName {
		t.Fatal("registration did not set a session cookie")
	}
	var user model.User
	if err := json.Unmarshal(response.Body.Bytes(), &user); err != nil {
		t.Fatal(err)
	}
	return cookies[0], user
}

func sendAudio(router http.Handler, path string, cookie *http.Cookie) *httptest.ResponseRecorder {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("name", "Shared track")
	_ = writer.WriteField("artist", "Test artist")
	_ = writer.WriteField("genre", "Indie")
	audio, err := writer.CreateFormFile("audio", "track.mp3")
	if err != nil {
		panic(err)
	}
	_, _ = io.WriteString(audio, "test audio bytes")
	if err := writer.Close(); err != nil {
		panic(err)
	}
	request := httptest.NewRequest(http.MethodPost, path, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestRegistrationAndLoginUseHashedPasswords(t *testing.T) {
	router, _, _ := newAppTestRouter(t)
	cookie, user := registerTestUser(t, router, "listener", "Listener")
	if user.Userpass != "" {
		t.Fatal("password hash was included in the public user response")
	}

	session := sendJSON(router, http.MethodGet, "/api/session", "", cookie)
	if session.Code != http.StatusOK {
		t.Fatalf("GET /api/session status = %d, want %d: %s", session.Code, http.StatusOK, session.Body)
	}

	invalidLogin := sendJSON(router, http.MethodPost, "/api/sessions",
		`{"login":"listener","password":"wrong-password"}`)
	if invalidLogin.Code != http.StatusUnauthorized {
		t.Fatalf("invalid login status = %d, want %d", invalidLogin.Code, http.StatusUnauthorized)
	}
	login := sendJSON(router, http.MethodPost, "/api/sessions",
		`{"login":" LISTENER ","password":"correct-horse-battery-staple"}`)
	if login.Code != http.StatusOK {
		t.Fatalf("valid login status = %d, want %d: %s", login.Code, http.StatusOK, login.Body)
	}
}

func TestSharedMusicPrivatePlaylistsLikesAndReposts(t *testing.T) {
	router, uploadDir, _ := newAppTestRouter(t)
	ownerCookie, owner := registerTestUser(t, router, "owner", "Owner")
	guestCookie, guest := registerTestUser(t, router, "guest", "Guest")

	unauthenticatedUpload := sendAudio(router, "/api/music", nil)
	if unauthenticatedUpload.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous upload status = %d, want %d", unauthenticatedUpload.Code, http.StatusUnauthorized)
	}
	createdMusic := sendAudio(router, "/api/music", ownerCookie)
	if createdMusic.Code != http.StatusCreated {
		t.Fatalf("POST /api/music status = %d, want %d: %s", createdMusic.Code, http.StatusCreated, createdMusic.Body)
	}
	var music model.Music
	if err := json.Unmarshal(createdMusic.Body.Bytes(), &music); err != nil {
		t.Fatal(err)
	}
	if music.Username != int(owner.ID) || !strings.HasPrefix(music.PathToMusic, "/uploads/") {
		t.Fatalf("track owner or audio path was not saved: %+v", music)
	}
	storedAudio := filepath.Join(uploadDir, filepath.Base(music.PathToMusic))
	if _, err := os.Stat(storedAudio); err != nil {
		t.Fatalf("uploaded audio was not stored: %v", err)
	}
	audioResponse := sendJSON(router, http.MethodGet, music.PathToMusic, "")
	if audioResponse.Code != http.StatusOK {
		t.Fatalf("uploaded audio status = %d, want %d", audioResponse.Code, http.StatusOK)
	}
	publicCatalog := sendJSON(router, http.MethodGet, "/api/music", "")
	var catalog []model.Music
	if err := json.Unmarshal(publicCatalog.Body.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog) != 1 || catalog[0].ID != music.ID {
		t.Fatalf("new track was not shared in the public catalog: %+v", catalog)
	}

	publicPlaylist := sendJSON(router, http.MethodPost, "/api/playlists",
		`{"name":"Everyone","private":false}`, ownerCookie)
	privatePlaylist := sendJSON(router, http.MethodPost, "/api/playlists",
		`{"name":"Only mine","private":true}`, ownerCookie)
	var private model.Playlist
	if publicPlaylist.Code != http.StatusCreated || privatePlaylist.Code != http.StatusCreated {
		t.Fatalf("playlist creation failed: public=%d private=%d", publicPlaylist.Code, privatePlaylist.Code)
	}
	if err := json.Unmarshal(privatePlaylist.Body.Bytes(), &private); err != nil {
		t.Fatal(err)
	}
	anonymousPlaylists := sendJSON(router, http.MethodGet, "/api/playlists", "")
	var publicLists []model.Playlist
	if err := json.Unmarshal(anonymousPlaylists.Body.Bytes(), &publicLists); err != nil {
		t.Fatal(err)
	}
	if len(publicLists) != 1 || publicLists[0].Name != "Everyone" {
		t.Fatalf("private playlist was exposed publicly: %+v", publicLists)
	}
	ownerListsResponse := sendJSON(router, http.MethodGet, "/api/my/playlists", "", ownerCookie)
	guestListsResponse := sendJSON(router, http.MethodGet, "/api/my/playlists", "", guestCookie)
	var ownerLists []model.Playlist
	var guestLists []model.Playlist
	if err := json.Unmarshal(ownerListsResponse.Body.Bytes(), &ownerLists); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(guestListsResponse.Body.Bytes(), &guestLists); err != nil {
		t.Fatal(err)
	}
	if len(ownerLists) != 2 || len(guestLists) != 0 {
		t.Fatalf("private playlist ownership was not enforced: owner=%+v guest=%+v", ownerLists, guestLists)
	}

	addTrack := sendJSON(router, http.MethodPost,
		"/api/playlists/"+strconv.FormatUint(uint64(private.ID), 10)+"/tracks",
		`{"music_id":`+strconv.FormatUint(uint64(music.ID), 10)+`}`, ownerCookie)
	if addTrack.Code != http.StatusOK {
		t.Fatalf("owner could not add a track to a private playlist: %d %s", addTrack.Code, addTrack.Body)
	}
	foreignAccess := sendJSON(router, http.MethodPost,
		"/api/playlists/"+strconv.FormatUint(uint64(private.ID), 10)+"/tracks",
		`{"music_id":`+strconv.FormatUint(uint64(music.ID), 10)+`}`, guestCookie)
	if foreignAccess.Code != http.StatusNotFound {
		t.Fatalf("non-owner private playlist access status = %d, want %d", foreignAccess.Code, http.StatusNotFound)
	}

	like := sendJSON(router, http.MethodPost, "/api/likes", `{"id_music":`+strconv.FormatUint(uint64(music.ID), 10)+`}`, guestCookie)
	if like.Code != http.StatusCreated {
		t.Fatalf("POST /api/likes status = %d: %s", like.Code, like.Body)
	}
	userLikes := sendJSON(router, http.MethodGet, "/api/likes", "", guestCookie)
	var likes []model.Like
	if err := json.Unmarshal(userLikes.Body.Bytes(), &likes); err != nil {
		t.Fatal(err)
	}
	if len(likes) != 1 || len(likes[0].ID_music) != 1 || likes[0].ID_music[0] != int(music.ID) {
		t.Fatalf("like was not stored for the authenticated user: %+v", likes)
	}
	unlike := sendJSON(router, http.MethodDelete, "/api/likes/"+strconv.FormatUint(uint64(music.ID), 10), "", guestCookie)
	if unlike.Code != http.StatusNoContent {
		t.Fatalf("DELETE /api/likes status = %d, want %d: %s", unlike.Code, http.StatusNoContent, unlike.Body)
	}
	repost := sendJSON(router, http.MethodPost, "/api/reposts", `{"id_music":`+strconv.FormatUint(uint64(music.ID), 10)+`}`, guestCookie)
	if repost.Code != http.StatusCreated {
		t.Fatalf("POST /api/reposts status = %d: %s", repost.Code, repost.Body)
	}
	profileReposts := sendJSON(router, http.MethodGet, "/api/users/"+strconv.FormatUint(uint64(guest.ID), 10)+"/reposts", "")
	var reposts []model.Repost
	if err := json.Unmarshal(profileReposts.Body.Bytes(), &reposts); err != nil {
		t.Fatal(err)
	}
	if len(reposts) != 1 || len(reposts[0].ID_music) != 1 || reposts[0].ID_music[0] != int(music.ID) {
		t.Fatalf("repost did not appear on the user's public profile: %+v", reposts)
	}
}

func TestDeletionPermissionsAndAdminCleanup(t *testing.T) {
	router, uploadDir, db := newAppTestRouter(t)
	adminCookie, admin := registerTestUser(t, router, "admin", "Admin")
	userCookie, user := registerTestUser(t, router, "listener", "Listener")

	forbiddenAdminList := sendJSON(router, http.MethodGet, "/api/admin/users", "", userCookie)
	if forbiddenAdminList.Code != http.StatusForbidden {
		t.Fatalf("non-admin GET /api/admin/users status = %d, want %d", forbiddenAdminList.Code, http.StatusForbidden)
	}

	if err := db.Model(&model.User{}).Where("id = ?", admin.ID).Update("is_admin", true).Error; err != nil {
		t.Fatal(err)
	}
	adminList := sendJSON(router, http.MethodGet, "/api/admin/users", "", adminCookie)
	if adminList.Code != http.StatusOK {
		t.Fatalf("admin GET /api/admin/users status = %d, want %d: %s", adminList.Code, http.StatusOK, adminList.Body)
	}
	selfDelete := sendJSON(router, http.MethodDelete, "/api/admin/users/"+strconv.FormatUint(uint64(admin.ID), 10), "", adminCookie)
	if selfDelete.Code != http.StatusConflict {
		t.Fatalf("admin self-delete status = %d, want %d", selfDelete.Code, http.StatusConflict)
	}

	adminPlaylist := model.Playlist{ID_user: int(admin.ID), Name: "Admin playlist"}
	if err := db.Create(&adminPlaylist).Error; err != nil {
		t.Fatal(err)
	}
	forbiddenPlaylistDelete := sendJSON(router, http.MethodDelete, "/api/playlists/"+strconv.FormatUint(uint64(adminPlaylist.ID), 10), "", userCookie)
	if forbiddenPlaylistDelete.Code != http.StatusNotFound {
		t.Fatalf("non-owner DELETE /api/playlists status = %d, want %d", forbiddenPlaylistDelete.Code, http.StatusNotFound)
	}
	adminPlaylistDelete := sendJSON(router, http.MethodDelete, "/api/playlists/"+strconv.FormatUint(uint64(adminPlaylist.ID), 10), "", adminCookie)
	if adminPlaylistDelete.Code != http.StatusNoContent {
		t.Fatalf("admin DELETE /api/playlists status = %d, want %d: %s", adminPlaylistDelete.Code, http.StatusNoContent, adminPlaylistDelete.Body)
	}

	userPlaylist := model.Playlist{ID_user: int(user.ID), Name: "Own playlist"}
	if err := db.Create(&userPlaylist).Error; err != nil {
		t.Fatal(err)
	}
	userPlaylistDelete := sendJSON(router, http.MethodDelete, "/api/playlists/"+strconv.FormatUint(uint64(userPlaylist.ID), 10), "", userCookie)
	if userPlaylistDelete.Code != http.StatusNoContent {
		t.Fatalf("owner DELETE /api/playlists status = %d, want %d", userPlaylistDelete.Code, http.StatusNoContent)
	}

	adminMusicResponse := sendAudio(router, "/api/music", adminCookie)
	if adminMusicResponse.Code != http.StatusCreated {
		t.Fatalf("admin POST /api/music status = %d: %s", adminMusicResponse.Code, adminMusicResponse.Body)
	}
	var adminMusic model.Music
	if err := json.Unmarshal(adminMusicResponse.Body.Bytes(), &adminMusic); err != nil {
		t.Fatal(err)
	}
	forbiddenDelete := sendJSON(router, http.MethodDelete, "/api/music/"+strconv.FormatUint(uint64(adminMusic.ID), 10), "", userCookie)
	if forbiddenDelete.Code != http.StatusNotFound {
		t.Fatalf("non-owner DELETE /api/music status = %d, want %d", forbiddenDelete.Code, http.StatusNotFound)
	}

	createdMusic := sendAudio(router, "/api/music", userCookie)
	if createdMusic.Code != http.StatusCreated {
		t.Fatalf("POST /api/music status = %d: %s", createdMusic.Code, createdMusic.Body)
	}
	var music model.Music
	if err := json.Unmarshal(createdMusic.Body.Bytes(), &music); err != nil {
		t.Fatal(err)
	}
	storedAudio := filepath.Join(uploadDir, filepath.Base(music.PathToMusic))

	if err := db.Create(&model.Like{ID_user: int(admin.ID), ID_music: []int{int(music.ID), 999}}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Like{ID_user: int(user.ID), ID_music: []int{int(music.ID), 998}}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Playlist{ID_user: int(admin.ID), Name: "Mixed", ID_music: []int{int(music.ID), 999}}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Repost{ID_user: int(admin.ID), ID_music: []int{int(music.ID), 999}}).Error; err != nil {
		t.Fatal(err)
	}
	unlike := sendJSON(router, http.MethodDelete, "/api/likes/"+strconv.FormatUint(uint64(music.ID), 10), "", userCookie)
	if unlike.Code != http.StatusNoContent {
		t.Fatalf("DELETE /api/likes status = %d, want %d: %s", unlike.Code, http.StatusNoContent, unlike.Body)
	}
	var remainingUserLike model.Like
	if err := db.Where("id_user = ?", user.ID).First(&remainingUserLike).Error; err != nil {
		t.Fatal(err)
	}
	if len(remainingUserLike.ID_music) != 1 || remainingUserLike.ID_music[0] != 998 {
		t.Fatalf("removing one liked track removed other liked tracks: %+v", remainingUserLike.ID_music)
	}

	ownerDelete := sendJSON(router, http.MethodDelete, "/api/music/"+strconv.FormatUint(uint64(music.ID), 10), "", userCookie)
	if ownerDelete.Code != http.StatusNoContent {
		t.Fatalf("owner DELETE /api/music status = %d, want %d: %s", ownerDelete.Code, http.StatusNoContent, ownerDelete.Body)
	}
	if _, err := os.Stat(storedAudio); !os.IsNotExist(err) {
		t.Fatalf("deleted track audio still exists or could not be checked: %v", err)
	}

	var like model.Like
	if err := db.Where("id_user = ?", admin.ID).First(&like).Error; err != nil {
		t.Fatal(err)
	}
	if len(like.ID_music) != 1 || like.ID_music[0] != 999 {
		t.Fatalf("music deletion did not preserve unrelated liked track IDs: %+v", like.ID_music)
	}
	var playlist model.Playlist
	if err := db.Where("id_user = ?", admin.ID).First(&playlist).Error; err != nil {
		t.Fatal(err)
	}
	if len(playlist.ID_music) != 1 || playlist.ID_music[0] != 999 {
		t.Fatalf("music deletion did not preserve unrelated playlist track IDs: %+v", playlist.ID_music)
	}
	var repost model.Repost
	if err := db.Where("id_user = ?", admin.ID).First(&repost).Error; err != nil {
		t.Fatal(err)
	}
	if len(repost.ID_music) != 1 || repost.ID_music[0] != 999 {
		t.Fatalf("music deletion did not preserve unrelated repost track IDs: %+v", repost.ID_music)
	}

	adminDeleteTrackResponse := sendAudio(router, "/api/music", userCookie)
	if adminDeleteTrackResponse.Code != http.StatusCreated {
		t.Fatalf("second POST /api/music status = %d: %s", adminDeleteTrackResponse.Code, adminDeleteTrackResponse.Body)
	}
	var adminDeleteTrack model.Music
	if err := json.Unmarshal(adminDeleteTrackResponse.Body.Bytes(), &adminDeleteTrack); err != nil {
		t.Fatal(err)
	}
	adminDelete := sendJSON(router, http.MethodDelete, "/api/music/"+strconv.FormatUint(uint64(adminDeleteTrack.ID), 10), "", adminCookie)
	if adminDelete.Code != http.StatusNoContent {
		t.Fatalf("admin DELETE /api/music status = %d, want %d: %s", adminDelete.Code, http.StatusNoContent, adminDelete.Body)
	}

	createdMusic = sendAudio(router, "/api/music", userCookie)
	if createdMusic.Code != http.StatusCreated {
		t.Fatalf("third POST /api/music status = %d: %s", createdMusic.Code, createdMusic.Body)
	}
	var userTrack model.Music
	if err := json.Unmarshal(createdMusic.Body.Bytes(), &userTrack); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Playlist{ID_user: int(user.ID), Name: "User playlist", ID_music: []int{int(userTrack.ID)}}).Error; err != nil {
		t.Fatal(err)
	}

	deleteUser := sendJSON(router, http.MethodDelete, "/api/admin/users/"+strconv.FormatUint(uint64(user.ID), 10), "", adminCookie)
	if deleteUser.Code != http.StatusNoContent {
		t.Fatalf("admin DELETE /api/admin/users status = %d, want %d: %s", deleteUser.Code, http.StatusNoContent, deleteUser.Body)
	}
	if err := db.First(&model.User{}, user.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("deleted user still exists or lookup failed: %v", err)
	}
	if err := db.Where("username = ?", user.ID).First(&model.Music{}).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("deleted user's tracks remain or lookup failed: %v", err)
	}
}
